package adapters

import (
	"bufio"
	"context"
	"devops_console/internal/domain/entities"
	"devops_console/internal/ports"
	"fmt"
	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type taskState struct {
	execution *entities.TaskExecution
	cancel    context.CancelFunc
}

type ExecutionError struct {
	Message string
	Code    string
}

func (e ExecutionError) Error() string {
	return e.Message
}

func NewExecutionError(code, message string) ExecutionError {
	return ExecutionError{
		Message: message,
		Code:    code,
	}
}

type TaskProgressPayload struct {
	Status entities.TaskStatus `json:"status"`
	Error  string              `json:"error,omitempty"`
}

type K8sTaskExecutor struct {
	clientset      *kubernetes.Clientset
	namespace      string
	eventStream    ports.TaskEventStream
	taskExecutions map[string]*entities.TaskExecution
	tasks          sync.Map // Usar sync.Map en lugar de map con mutex
}

func NewK8sTaskExecutor(namespace string, eventStream ports.TaskEventStream) (*K8sTaskExecutor, error) {
	var config *rest.Config
	var err error

	// Check if kubeconfig file exists
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if _, err := os.Stat(kubeconfig); err == nil {
		// Use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	} else {
		// Use the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sTaskExecutor{
		clientset:      clientset,
		namespace:      namespace,
		eventStream:    eventStream,
		taskExecutions: make(map[string]*entities.TaskExecution),
	}, nil
}

func (e *K8sTaskExecutor) ExecuteTask(ctx context.Context, task *entities.DevOpsTask) (string, error) {
	timeout, ok := task.Worker.GetDetails()["JobTimeout"].(time.Duration)
	if !ok {
		timeout = 30 * time.Second // Default value
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	executionID := uuid.New().String()
	taskExecution := &entities.TaskExecution{
		ID:           executionID,
		DevOpsTaskID: task.ID,
		Status:       entities.TaskRunning,
		StartedAt:    time.Now(),
	}

	e.tasks.Store(executionID, &taskState{
		execution: taskExecution,
		cancel:    cancel,
	})

	go func() {
		defer cancel()
		e.runTask(ctx, task, taskExecution)
	}()

	return executionID, nil
}

func (e *K8sTaskExecutor) runTask(ctx context.Context, task *entities.DevOpsTask, taskExecution *entities.TaskExecution) {
	jobName := fmt.Sprintf("task-%s", taskExecution.ID)

	// Crear el objeto Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    task.Name,
							Image:   task.Worker.GetDetails()["Image"].(string),
							Command: task.Worker.GetDetails()["Command"].([]string),
							Env:     getEnvVars(task.Worker.GetDetails()),
						},
					},
				},
			},
		},
	}

	defer e.cleanup(context.Background(), jobName, e.namespace)

	// Crear el Job en Kubernetes
	jobsClient := e.clientset.BatchV1().Jobs(e.namespace)
	_, err := jobsClient.Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskFailed, fmt.Sprintf("Failed to create job: %v", err))
		return
	}

	// Esperar a que el Pod esté en ejecución
	podName, err := e.waitForPodRunning(ctx, jobName)
	if err != nil {
		e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskFailed, fmt.Sprintf("Failed to wait for pod running: %v", err))
		return
	}

	// Publish event that the pod is running
	e.publishEvent(taskExecution.ID, entities.EventTypeTaskStarted, "Pod is started")

	// Hacer streaming de los logs
	if err := e.streamPodLogs(ctx, podName, taskExecution); err != nil {
		e.publishEvent(taskExecution.ID, entities.EventTypeTaskError, fmt.Sprintf("Error streaming logs: %v", err))
	}

	defer e.eventStream.Close(taskExecution.ID)
	// Esperar a que el Job complete
	if err := e.waitForJobCompletion(ctx, jobName); err != nil {
		e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskFailed, fmt.Sprintf("Failed to wait for job completion: %v", err))
		return
	}

	// Actualizar el estado de la ejecución de la tarea a Succeeded
	e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskSucceeded, "")
}

func (e *K8sTaskExecutor) cleanup(ctx context.Context, jobName string, namespace string) error {
	propagationPolicy := metav1.DeletePropagationBackground
	return e.clientset.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
}

func (e *K8sTaskExecutor) waitForPodRunning(ctx context.Context, jobName string) (string, error) {
	podsClient := e.clientset.CoreV1().Pods(e.namespace)
	var podName string

	err := wait.PollUntilContextTimeout(ctx, time.Second, 5*time.Minute, true, func(context.Context) (bool, error) {
		podList, err := podsClient.List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", jobName),
		})
		if err != nil {
			return false, err
		}
		if len(podList.Items) == 0 {
			return false, nil
		}
		pod := podList.Items[0]
		podName = pod.Name
		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, NewExecutionError("POD_TERMINATED", "Pod terminated before running")
		default:
			return false, nil
		}
	})
	return podName, err
}

func (e *K8sTaskExecutor) SubscribeToTaskEvents(taskExecutionID string) (<-chan entities.TaskEvent, error) {
	return e.eventStream.Subscribe(taskExecutionID)
}

func (e *K8sTaskExecutor) streamPodLogs(ctx context.Context, podName string, taskExecution *entities.TaskExecution) error {
	podsClient := e.clientset.CoreV1().Pods(e.namespace)
	req := podsClient.GetLogs(podName, &corev1.PodLogOptions{
		Follow: true,
	})

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer podLogs.Close()

	scanner := bufio.NewScanner(podLogs)
	for scanner.Scan() {
		line := scanner.Text()
		event := entities.TaskEvent{
			ExecutionID: taskExecution.ID,
			Payload:     line,
			Timestamp:   time.Now(),
			EventType:   entities.EventTypeTaskOutput,
		}
		e.eventStream.Publish(event)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (e *K8sTaskExecutor) waitForJobCompletion(ctx context.Context, jobName string) error {
	jobsClient := e.clientset.BatchV1().Jobs(e.namespace)

	return wait.PollUntilContextTimeout(ctx, time.Second, 10*time.Minute, true, func(context.Context) (bool, error) {
		job, err := jobsClient.Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if job.Status.Succeeded > 0 {
			return true, nil
		}
		if job.Status.Failed > 0 {
			return false, NewExecutionError("JOB_FAILED", "Job execution failed")
		}
		return false, nil
	})
}

func (e *K8sTaskExecutor) updateTaskExecutionStatus(executionID string, status entities.TaskStatus, errMsg string) {
	if state, ok := e.tasks.Load(executionID); ok {
		taskState := state.(*taskState)
		taskState.execution.Status = status
		taskState.execution.FinishedAt = time.Now()
		if errMsg != "" {
			taskState.execution.Error = errMsg
		}
	}
	typeEvent := entities.EventTypeTaskProgress
	if status == entities.TaskSucceeded {
		typeEvent = entities.EventTypeTaskCompleted
	} else if status == entities.TaskFailed {
		typeEvent = entities.EventTypeTaskFailed
	} else if status == entities.TaskError {
		typeEvent = entities.EventTypeTaskError
	}

	e.publishEvent(executionID, typeEvent, TaskProgressPayload{
		Status: status,
		Error:  errMsg,
	})
}

func (e *K8sTaskExecutor) GetTaskStatus(ctx context.Context, taskExecutionID string) (entities.TaskStatus, error) {
	if state, ok := e.tasks.Load(taskExecutionID); ok {
		return state.(*taskState).execution.Status, nil
	}
	return entities.TaskError, NewExecutionError("TASK_NOT_FOUND", "Task execution ID not found")
}

func (e *K8sTaskExecutor) CancelTask(ctx context.Context, executionID string) error {
	if state, ok := e.tasks.Load(executionID); ok {
		taskState := state.(*taskState)
		taskState.cancel()
		return e.cleanup(ctx, fmt.Sprintf("task-%s", executionID), e.namespace)
	}
	return NewExecutionError("TASK_NOT_FOUND", "Task execution ID not found")
}

func (e *K8sTaskExecutor) monitorJobProgress(ctx context.Context, jobName string, executionID string) {
	job, err := e.clientset.BatchV1().Jobs(e.namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		e.publishEvent(executionID, entities.EventTypeTaskError, err.Error())
		return
	}

	pods, err := e.clientset.CoreV1().Pods(e.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil || len(pods.Items) == 0 {
		e.publishEvent(executionID, entities.EventTypeTaskError, "No pods found for job")
		return
	}
	pod := pods.Items[0]

	for {
		select {
		case <-ctx.Done():
			return
		default:
			logs, err := e.clientset.CoreV1().Pods(e.namespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).Stream(ctx)
			if err == nil {
				buffer := make([]byte, 2048)
				if n, err := logs.Read(buffer); err == nil {
					e.publishEvent(executionID, entities.EventTypeTaskOutput, string(buffer[:n]))
				}
				logs.Close()
			}

			if job.Status.Succeeded > 0 {
				e.publishEvent(executionID, entities.EventTypeTaskCompleted, "Job completed successfully")
				return
			} else if job.Status.Failed > 0 {
				e.publishEvent(executionID, entities.EventTypeTaskFailed, "Job failed")
				return
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func (e *K8sTaskExecutor) publishEvent(executionID string, eventType entities.TaskEventType, payload interface{}) {
	event := entities.TaskEvent{
		ID:          generateEventID(),
		ExecutionID: executionID,
		Timestamp:   time.Now(),
		EventType:   eventType,
		Payload:     payload,
	}
	e.eventStream.Publish(event)
}

func generateEventID() string {
	return uuid.New().String()
}

func getEnvVars(parameters map[string]interface{}) []corev1.EnvVar {
	if env, ok := parameters["EnvVars"].([]corev1.EnvVar); ok {
		return env
	}
	return []corev1.EnvVar{} // Return an empty slice if "Env" is not defined
}

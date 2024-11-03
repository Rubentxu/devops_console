package executor

import (
	"bufio"
	"context"
	"devops_console/internal/domain/entities"
	"devops_console/internal/ports"
	"fmt"
	"github.com/docker/docker/api/types/container"
	containerImage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"io"
	"os"
	"time"
)

type DockerTaskExecutor struct {
	client         *client.Client
	eventStream    ports.TaskEventStream
	taskExecutions map[string]*entities.TaskExecution
}

func NewDockerTaskExecutor(eventStream ports.TaskEventStream) (*DockerTaskExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerTaskExecutor{
		client:         cli,
		eventStream:    eventStream,
		taskExecutions: make(map[string]*entities.TaskExecution),
	}, nil
}

func (e *DockerTaskExecutor) ExecuteTask(ctx context.Context, task *entities.DevOpsTask) (string, error) {
	timeout, ok := task.Config.Parameters["JobTimeout"].(time.Duration)
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

	e.taskExecutions[executionID] = taskExecution

	go func() {
		defer cancel()
		e.runTask(ctx, task, taskExecution)
	}()

	return executionID, nil
}

func (e *DockerTaskExecutor) runTask(ctx context.Context, task *entities.DevOpsTask, taskExecution *entities.TaskExecution) {
	image := task.Config.Parameters["Image"].(string)

	// Pull the image if it does not exist
	_, _, err := e.client.ImageInspectWithRaw(ctx, image)
	if err != nil {
		if client.IsErrNotFound(err) {
			e.publishEvent(taskExecution.ID, entities.EventTypeTaskProgress, fmt.Sprintf("Pulling image: %s", image))
			out, err := e.client.ImagePull(ctx, image, containerImage.PullOptions{})
			if err != nil {
				e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskFailed, fmt.Sprintf("Failed to pull image: %v", err))
				return
			}
			defer out.Close()
			io.Copy(os.Stdout, out)
		} else {
			e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskFailed, fmt.Sprintf("Failed to inspect image: %v", err))
			return
		}
	}

	config := &container.Config{
		Image: image,
		Cmd:   task.Config.Parameters["Command"].([]string),
		Env:   getDockerEnvVars(task.Config.Parameters),
	}

	resp, err := e.client.ContainerCreate(ctx, config, nil, nil, nil, "")
	if err != nil {
		e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskFailed, fmt.Sprintf("Failed to create container: %v", err))
		return
	}

	containerID := resp.ID
	defer e.cleanup(ctx, containerID)

	if err := e.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskFailed, fmt.Sprintf("Failed to start container: %v", err))
		return
	}

	e.publishEvent(taskExecution.ID, entities.EventTypeTaskStarted, "Container is started")

	if err := e.streamContainerLogs(ctx, containerID, taskExecution); err != nil {
		e.publishEvent(taskExecution.ID, entities.EventTypeTaskError, fmt.Sprintf("Error streaming logs: %v", err))
	}

	statusCh, errCh := e.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskFailed, fmt.Sprintf("Container wait error: %v", err))
			return
		}
	case <-statusCh:
		e.updateTaskExecutionStatus(taskExecution.ID, entities.TaskSucceeded, "")
	}
}

func (e *DockerTaskExecutor) updateTaskExecutionStatus(executionID string, status entities.TaskStatus, errMsg string) {
	if taskExecution, ok := e.taskExecutions[executionID]; ok {
		taskExecution.Status = status
		taskExecution.FinishedAt = time.Now()
		if errMsg != "" {
			taskExecution.Error = errMsg
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

func (e *DockerTaskExecutor) streamContainerLogs(ctx context.Context, containerID string, taskExecution *entities.TaskExecution) error {
	out, err := e.client.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return err
	}
	defer out.Close()

	scanner := bufio.NewScanner(out)
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

func (e *DockerTaskExecutor) cleanup(ctx context.Context, containerID string) error {
	return e.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

func (e *DockerTaskExecutor) GetTaskStatus(ctx context.Context, taskExecutionID string) (entities.TaskStatus, error) {
	if state, ok := e.taskExecutions[taskExecutionID]; ok {
		return state.Status, nil
	}
	return entities.TaskError, fmt.Errorf("task execution ID not found")
}

func (e *DockerTaskExecutor) CancelTask(ctx context.Context, executionID string) error {
	if state, ok := e.taskExecutions[executionID]; ok {
		return e.cleanup(ctx, state.ID)
	}
	return fmt.Errorf("task execution ID not found")
}

func (e *DockerTaskExecutor) SubscribeToTaskEvents(taskExecutionID string) (<-chan entities.TaskEvent, error) {
	return e.eventStream.Subscribe(taskExecutionID)
}

func (e *DockerTaskExecutor) publishEvent(executionID string, eventType entities.TaskEventType, payload interface{}) {
	event := entities.TaskEvent{
		ID:          uuid.New().String(),
		ExecutionID: executionID,
		Timestamp:   time.Now(),
		EventType:   eventType,
		Payload:     payload,
	}
	e.eventStream.Publish(event)
}

func getDockerEnvVars(parameters map[string]interface{}) []string {
	if env, ok := parameters["Env"].([]string); ok {
		return env
	}
	return []string{} // Return an empty slice if "Env" is not defined
}

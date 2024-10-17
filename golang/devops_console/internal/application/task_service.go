package application

import (
	"context"
	"devops_console/internal/domain/task"
	"devops_console/internal/infrastructure/workers"
	"devops_console/internal/infrastructure/workers/factories"
	"fmt"
	"strings"
	"time"
)

type TaskService struct {
    taskRepository task.TaskRepository
    workerConfigs  map[factories.WorkerType]workers.WorkerConfig
    workerFactory  factories.WorkerFactory
    taskQueue      []string
    taskStatistics TaskStatistics
    pausedTasks    map[string]time.Time
}

type TaskStatistics struct {
    TotalTasks      int
    SuccessfulTasks int
    FailedTasks     int
    TotalDuration   time.Duration
}

func NewTaskService(taskRepository task.TaskRepository, workerConfigs map[factories.WorkerType]workers.WorkerConfig) *TaskService {
    return &TaskService{
        taskRepository: taskRepository,
        workerConfigs:  workerConfigs,
        workerFactory:  factories.NewWorkerFactory(),
        taskQueue:      make([]string, 0, 100),
        pausedTasks:    make(map[string]time.Time),
    }
}

func (s *TaskService) GetWorker(workerType factories.WorkerType) (factories.JobLauncher, factories.JobMonitor, factories.LogStreamer, workers.WorkerConfig, error) {
    config, ok := s.workerConfigs[workerType]
    if !ok {
        return nil, nil, nil, workers.WorkerConfig{}, fmt.Errorf("no configuration found for worker type: %v", workerType)
    }

    jobLauncher := s.workerFactory.GetJobLauncher(workerType)
    jobMonitor := s.workerFactory.GetJobMonitor(workerType)
    logStreamer := s.workerFactory.GetLogStreamer(workerType)

    return jobLauncher, jobMonitor, logStreamer, config, nil
}

func (s *TaskService) UpdateWorkerConfigs(newConfigs map[factories.WorkerType]workers.WorkerConfig) {
    for k, v := range newConfigs {
        s.workerConfigs[k] = v
    }
}

func (s *TaskService) CreateTask(taskCreate task.TaskCreate) (*task.Task, error) {
    return s.taskRepository.Create(taskCreate)
}

func (s *TaskService) GetAllTasks() ([]*task.Task, error) {
    return s.taskRepository.GetAll()
}

func (s *TaskService) GetTaskByID(taskID string) (*task.Task, error) {
    return s.taskRepository.GetByID(taskID)
}

func (s *TaskService) UpdateTask(taskID string, taskUpdate task.TaskUpdate) (*task.Task, error) {
    return s.taskRepository.Update(taskID, taskUpdate)
}

func (s *TaskService) DeleteTask(taskID string) error {
    return s.taskRepository.Delete(taskID)
}

func (s *TaskService) AddTaskExecution(taskID string, taskExecuted task.TaskExecuted) (*task.Task, error) {
    t, err := s.GetTaskByID(taskID)
    if err != nil {
        return nil, err
    }
    t.TasksExecuted = append(t.TasksExecuted, taskExecuted)
    return s.UpdateTask(taskID, task.TaskUpdate{TasksExecuted: &t.TasksExecuted})
}

func (s *TaskService) AddTaskToQueue(taskID string) {
    s.taskQueue = append(s.taskQueue, taskID)
}

func (s *TaskService) GetNextTaskFromQueue() (string, bool) {
    if len(s.taskQueue) == 0 {
        return "", false
    }
    taskID := s.taskQueue[0]
    s.taskQueue = s.taskQueue[1:]
    return taskID, true
}

func (s *TaskService) UpdateTaskStatistics(success bool, duration time.Duration) {
    s.taskStatistics.TotalTasks++
    if success {
        s.taskStatistics.SuccessfulTasks++
    } else {
        s.taskStatistics.FailedTasks++
    }
    s.taskStatistics.TotalDuration += duration
}

func (s *TaskService) GetTaskStatistics() map[string]interface{} {
    avgDuration := time.Duration(0)
    successRate := 0.0
    if s.taskStatistics.TotalTasks > 0 {
        avgDuration = s.taskStatistics.TotalDuration / time.Duration(s.taskStatistics.TotalTasks)
        successRate = float64(s.taskStatistics.SuccessfulTasks) / float64(s.taskStatistics.TotalTasks)
    }
    return map[string]interface{}{
        "average_duration":  avgDuration,
        "success_rate":      successRate,
        "total_tasks":       s.taskStatistics.TotalTasks,
        "successful_tasks":  s.taskStatistics.SuccessfulTasks,
        "failed_tasks":      s.taskStatistics.FailedTasks,
    }
}

func (s *TaskService) ExecuteTaskWithTimeout(ctx context.Context, taskID string, timeoutSeconds int) (map[string]interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
    defer cancel()

    resultChan := make(chan map[string]interface{}, 1)
    errChan := make(chan error, 1)

    go func() {
        result, err := s.ExecuteTask(ctx, taskID)
        if err != nil {
            errChan <- err
        } else {
            resultChan <- result
        }
    }()

    select {
    case result := <-resultChan:
        s.UpdateTaskStatistics(result["type"] != "error", time.Duration(timeoutSeconds)*time.Second)
        return result, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        return map[string]interface{}{
            "type":    "error",
            "message": "La tarea excedió el tiempo límite",
        }, ctx.Err()
    }
}

func (s *TaskService) PauseTask(taskID string) (map[string]interface{}, error) {
    s.pausedTasks[taskID] = time.Now()
    return map[string]interface{}{
        "type":    "status",
        "message": fmt.Sprintf("Tarea %s pausada", taskID),
    }, nil
}

func (s *TaskService) ResumeTask(taskID string) (map[string]interface{}, error) {
    if _, exists := s.pausedTasks[taskID]; exists {
        delete(s.pausedTasks, taskID)
        return map[string]interface{}{
            "type":    "status",
            "message": fmt.Sprintf("Tarea %s reanudada", taskID),
        }, nil
    }
    return nil, fmt.Errorf("tarea %s no está pausada", taskID)
}

func (s *TaskService) ExecuteTask(ctx context.Context, taskID string) (map[string]interface{}, error) {
    t, err := s.GetTaskByID(taskID)
    if err != nil {
        return nil, err
    }

    workerType := s.DetermineWorkerType(t)
    jobLauncher, _, _, config, err := s.GetWorker(workerType)
    if err != nil {
        return nil, err
    }

    result, err := jobLauncher.LaunchJob(ctx, t.Name, config.GetLaunchConfig())
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
    "status": "executing",
            "websocket_url": "/ws/task/" + taskID,
    }, nil
}

func (s *TaskService) GetTaskStatus(ctx context.Context, taskID string) (string, error) {
    t, err := s.GetTaskByID(taskID)
    if err != nil {
        return "", err
    }

    workerType := s.DetermineWorkerType(t)
    _, jobMonitor, _, _, err := s.GetWorker(workerType)
    if err != nil {
        return "", err
    }

    return jobMonitor.GetJobStatus(ctx, t.Name)
}

func (s *TaskService) MonitorTask(ctx context.Context, taskID string) (<-chan string, error) {
    t, err := s.GetTaskByID(taskID)
    if err != nil {
        return nil, err
    }

    workerType := s.DetermineWorkerType(t)
    _, jobMonitor, _, _, err := s.GetWorker(workerType)
    if err != nil {
        return nil, err
    }

    return jobMonitor.MonitorJob(ctx, t.Name)
}

func (s *TaskService) StreamTaskLogs(ctx context.Context, taskID string) (<-chan string, error) {
    t, err := s.GetTaskByID(taskID)
    if err != nil {
        return nil, err
    }

    workerType := s.DetermineWorkerType(t)
    _, _, logStreamer, _, err := s.GetWorker(workerType)
    if err != nil {
        return nil, err
    }

    return logStreamer.StreamLogs(ctx, t.Name)
}

func (s *TaskService) ExecuteAndMonitorTask(ctx context.Context, taskID string) (<-chan map[string]interface{}, error) {
    t, err := s.GetTaskByID(taskID)
    if err != nil {
        return nil, err
    }

    workerType := s.DetermineWorkerType(t)
    jobLauncher, jobMonitor, logStreamer, config, err := s.GetWorker(workerType)
    if err != nil {
        return nil, err
    }

    resultChan := make(chan map[string]interface{})

    go func() {
        defer close(resultChan)

        launchResult, err := jobLauncher.LaunchJob(ctx, t.Name, config.GetLaunchConfig())
        if err != nil {
            resultChan <- map[string]interface{}{"type": "error", "message": err.Error()}
            return
        }
        resultChan <- map[string]interface{}{"type": "launch", "message": launchResult}

        statusChan, err := jobMonitor.MonitorJob(ctx, t.Name)
        if err != nil {
            resultChan <- map[string]interface{}{"type": "error", "message": err.Error()}
            return
        }

        logChan, err := logStreamer.StreamLogs(ctx, t.Name)
        if err != nil {
            resultChan <- map[string]interface{}{"type": "error", "message": err.Error()}
            return
        }

        for {
            select {
            case status, ok := <-statusChan:
                if !ok {
                    return
                }
                resultChan <- map[string]interface{}{"type": "status", "status": status}
                if status == string(task.TaskStatusCompleted) || status == string(task.TaskStatusFailed) {
                    return
                }
            case log, ok := <-logChan:
                if !ok {
                    return
                }
                resultChan <- map[string]interface{}{"type": "log", "message": log}
            case <-ctx.Done():
                resultChan <- map[string]interface{}{"type": "error", "message": "Task execution cancelled"}
                return
            }
        }
    }()

    return resultChan, nil
}

func (s *TaskService) DetermineWorkerType(t *task.Task) factories.WorkerType {
    if t.WorkerType != nil {
        return *t.WorkerType
    }

    technologyMap := map[string]factories.WorkerType{
        "kubernetes": factories.WorkerTypeKubernetes,
        "openshift":  factories.WorkerTypeOpenShift,
        "docker":     factories.WorkerTypeDocker,
        "podman":     factories.WorkerTypePodman,
    }

    if t.Technology != "" {
        for key, workerType := range technologyMap {
            if strings.Contains(strings.ToLower(t.Technology), key) {
                return workerType
            }
        }
    }

    for _, tag := range t.Tags {
        for key, workerType := range technologyMap {
            if strings.Contains(strings.ToLower(tag), key) {
                return workerType
            }
        }
    }

    if t.TaskType != "" {
        taskType := strings.ToLower(t.TaskType)
        if strings.Contains(taskType, "deployment") {
            return factories.WorkerTypeKubernetes
        } else if strings.Contains(taskType, "build") {
            return factories.WorkerTypeDocker
        }
    }

    return factories.WorkerTypeKubernetes
}

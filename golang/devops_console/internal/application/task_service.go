package application

import (
	"context"
	task "devops_console/internal/domain/task"
	worker "devops_console/internal/domain/worker"
	"fmt"
	"strings"
	"time"
)

type TaskService struct {
	taskRepository task.TaskRepository
	workerFactory  worker.WorkerFactoryInterface
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

func NewTaskService(taskRepository task.TaskRepository, factory worker.WorkerFactoryInterface) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
		workerFactory:  factory,
		taskQueue:      make([]string, 0, 100),
		pausedTasks:    make(map[string]time.Time),
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
		"average_duration": avgDuration,
		"success_rate":     successRate,
		"total_tasks":      s.taskStatistics.TotalTasks,
		"successful_tasks": s.taskStatistics.SuccessfulTasks,
		"failed_tasks":     s.taskStatistics.FailedTasks,
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
	worker, err := s.workerFactory.GetWorker(workerType)
	if err != nil {
		return nil, err
	}

	jobID, err := worker.LaunchJob(ctx, taskID, t.WorkerConfig)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":        "executing",
		"websocket_url": "/ws/task/" + taskID,
		"job_id":        jobID,
	}, nil
}

func (s *TaskService) GetTaskStatus(ctx context.Context, taskID string) (string, error) {
	t, err := s.GetTaskByID(taskID)
	if err != nil {
		return "", err
	}

	workerType := s.DetermineWorkerType(t)
	worker, err := s.workerFactory.GetWorker(workerType)
	if err != nil {
		return "", err
	}

	return worker.GetJobStatus(ctx, taskID)
}

func (s *TaskService) MonitorTask(ctx context.Context, taskID string) (<-chan string, error) {
	t, err := s.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	workerType := s.DetermineWorkerType(t)
	worker, err := s.workerFactory.GetWorker(workerType)
	if err != nil {
		return nil, err
	}

	statusChan, errChan := worker.MonitorJob(ctx, taskID)

	resultChan := make(chan string)
	go func() {
		defer close(resultChan)
		for {
			select {
			case status := <-statusChan:
				resultChan <- status
			case err := <-errChan:
				if err != nil {
					resultChan <- fmt.Sprintf("Error: %v", err)
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultChan, nil
}

func (s *TaskService) StreamTaskLogs(ctx context.Context, taskID string) (<-chan string, error) {
	t, err := s.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	workerType := s.DetermineWorkerType(t)
	worker, err := s.workerFactory.GetWorker(workerType)
	if err != nil {
		return nil, err
	}

	logChan, errChan := worker.StreamLogs(ctx, taskID)

	resultChan := make(chan string)
	go func() {
		defer close(resultChan)
		for {
			select {
			case log := <-logChan:
				resultChan <- log
			case err := <-errChan:
				if err != nil {
					resultChan <- fmt.Sprintf("Error: %v", err)
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultChan, nil
}

func (s *TaskService) ExecuteAndMonitorTask(ctx context.Context, taskID string) (<-chan map[string]interface{}, error) {
	t, err := s.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	workerType := s.DetermineWorkerType(t)
	worker, err := s.workerFactory.GetWorker(workerType)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan map[string]interface{})

	go func() {
		defer close(resultChan)

		jobID, err := worker.LaunchJob(ctx, taskID, t.WorkerConfig)
		if err != nil {
			resultChan <- map[string]interface{}{"type": "error", "message": err.Error()}
			return
		}
		resultChan <- map[string]interface{}{"type": "launch", "job_id": jobID}

		statusChan, statusErrChan := worker.MonitorJob(ctx, taskID)
		logChan, logErrChan := worker.StreamLogs(ctx, taskID)

		for {
			select {
			case status := <-statusChan:
				resultChan <- map[string]interface{}{"type": "status", "status": status}
				if status == string(task.TaskStatusCompleted) || status == string(task.TaskStatusFailed) {
					return
				}
			case log := <-logChan:
				resultChan <- map[string]interface{}{"type": "log", "message": log}
			case err := <-statusErrChan:
				if err != nil {
					resultChan <- map[string]interface{}{"type": "error", "message": err.Error()}
				}
				return
			case err := <-logErrChan:
				if err != nil {
					resultChan <- map[string]interface{}{"type": "error", "message": err.Error()}
				}
				return
			case <-ctx.Done():
				resultChan <- map[string]interface{}{"type": "error", "message": "Task execution cancelled"}
				return
			}
		}
	}()

	return resultChan, nil
}

func (s *TaskService) DetermineWorkerType(t *task.Task) worker.WorkerType {
	if t.WorkerType != "" {
		return worker.WorkerType(t.WorkerType)
	}

	technologyMap := map[string]worker.WorkerType{
		"kubernetes": worker.WorkerType("kubernetes"),
		"openshift":  worker.WorkerType("openshift"),
		"docker":     worker.WorkerType("docker"),
		"podman":     worker.WorkerType("podman"),
	}

	if t.Technology != "" {
		for key, workerType := range technologyMap {
			if strings.Contains(strings.ToLower(t.Technology), key) {
				return workerType
			}
		}
	}

	if t.Tags != nil {
		for _, tag := range *t.Tags {
			for key, workerType := range technologyMap {
				if strings.Contains(strings.ToLower(tag), key) {
					return workerType
				}
			}
		}
	}

	if t.TaskType != "" {
		taskType := strings.ToLower(t.TaskType)
		if strings.Contains(taskType, "deployment") {
			return worker.WorkerType("kubernetes")
		} else if strings.Contains(taskType, "build") {
			return worker.WorkerType("docker")
		}
	}

	return worker.WorkerType("kubernetes")
}

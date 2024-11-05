package application

import (
	"context"
	"devops_console/internal/domain/entities"
	"devops_console/internal/ports"
	"errors"
	"github.com/google/uuid"
	"time"
)

type TaskService interface {
	CreateTask(task entities.DevOpsTask) (entities.DevOpsTask, error)
	UpdateTask(taskID string, updates ports.TaskUpdate) (entities.DevOpsTask, error)
	DeleteTask(taskID string) error
	GetTask(taskID string) (entities.DevOpsTask, error)
	GetTasks(filters ports.TaskFilters) ([]entities.DevOpsTask, error)
	ExecuteTask(taskID string) (string, error)
	GetTaskStatus(executionID string) (entities.TaskStatus, error)
	CancelTask(executionID string) error
	SubscribeToTaskEvents(executionID string) (<-chan entities.TaskEvent, error)
}

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidTask  = errors.New("invalid task")
)

type IDGenerator func() string

func defaultIDGenerator() string {
	return uuid.New().String()
}

type TaskServiceImpl struct {
	repository ports.TaskRepository
	executors  map[string]ports.TaskExecutor
	GenerateID IDGenerator
}

func NewTaskServiceImpl(taskRepo ports.TaskRepository) *TaskServiceImpl {
	return &TaskServiceImpl{
		repository: taskRepo,
		executors:  make(map[string]ports.TaskExecutor),
		GenerateID: defaultIDGenerator,
	}
}

func (s *TaskServiceImpl) RegisterExecutor(workerType string, executor ports.TaskExecutor) {
	s.executors[workerType] = executor
}

// Implementación de la interfaz TaskService
func (s *TaskServiceImpl) CreateTask(task entities.DevOpsTask) (entities.DevOpsTask, error) {
	if task.ID == "" {
		task.ID = s.GenerateID()
	}
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	err := s.repository.Create(context.Background(), &task)
	if err != nil {
		return entities.DevOpsTask{}, err
	}
	return task, nil
}

func (s *TaskServiceImpl) UpdateTask(taskID string, updates ports.TaskUpdate) (entities.DevOpsTask, error) {
	ctx := context.Background()
	task, err := s.repository.GetByID(ctx, taskID)
	if err != nil {
		return entities.DevOpsTask{}, err
	}

	// Actualizar los campos según el TaskUpdate
	if updates.Name != "" {
		task.Name = updates.Name
	}
	if updates.Description != "" {
		task.Description = updates.Description
	}
	if updates.Config.Parameters != nil {
		task.Config = updates.Config
	}
	if updates.TaskType != "" {
		task.TaskType = updates.TaskType
	}
	// if updates.Approvals != nil {
	//     task.Approvals = *updates.Approvals
	// }

	task.UpdatedAt = time.Now()

	err = s.repository.Update(ctx, &task)
	if err != nil {
		return entities.DevOpsTask{}, err
	}
	return task, nil
}

func (s *TaskServiceImpl) DeleteTask(taskID string, workspace entities.Workspace) error {
	ctx := context.Background()
	return s.repository.Delete(ctx, taskID)
}

func (s *TaskServiceImpl) GetTask(taskID string, workspace entities.Workspace) (entities.DevOpsTask, error) {
	ctx := context.Background()
	return s.repository.GetByID(ctx, taskID)
}

func (s *TaskServiceImpl) GetTasks(filters ports.TaskFilters) ([]entities.DevOpsTask, error) {
	ctx := context.Background()
	return s.repository.GetAll(ctx, filters)
}

func (s *TaskServiceImpl) ExecuteTask(taskID string) (string, error) {
	ctx := context.Background()
	task, err := s.repository.GetByID(ctx, taskID)
	if err != nil {
		return "", err
	}

	worker := task.Worker
	executor, ok := s.executors[worker.GetType()]
	if !ok {
		return "", errors.New("unsupported worker type")
	}

	executionID, err := executor.ExecuteTask(ctx, &task)
	if err != nil {
		return "", err
	}

	// Update the task with the new execution
	taskExecution := entities.TaskExecution{
		ID:           executionID,
		DevOpsTaskID: task.ID,
		Status:       entities.TaskRunning,
		StartedAt:    time.Now(),
	}
	task.Executions = append(task.Executions, &taskExecution)
	task.UpdatedAt = time.Now()
	err = s.repository.Update(ctx, &task)
	if err != nil {
		return "", err
	}

	return executionID, nil
}

func (s *TaskServiceImpl) CancelTask(executionID string) error {
	ctx := context.Background()
	task, err := s.repository.GetByExecutionID(ctx, executionID)
	if err != nil {
		return err
	}

	worker := task.Worker
	executor, ok := s.executors[worker.GetType()]
	if !ok {
		return errors.New("unsupported worker type")
	}

	return executor.CancelTask(ctx, executionID)
}

func (s *TaskServiceImpl) SubscribeToTaskEvents(executionID string) (<-chan entities.TaskEvent, error) {
	ctx := context.Background()
	task, err := s.repository.GetByExecutionID(ctx, executionID)
	if err != nil {
		return nil, err
	}

	worker := task.Worker
	executor, ok := s.executors[worker.GetType()]
	if !ok {
		return nil, errors.New("unsupported worker type")
	}

	return executor.SubscribeToTaskEvents(executionID)
}

func (s *TaskServiceImpl) GetTaskStatus(executionID string) (entities.TaskStatus, error) {
	ctx := context.Background()
	task, err := s.repository.GetByExecutionID(ctx, executionID)
	if err != nil {
		return "", err
	}

	for _, execution := range task.Executions {
		if execution.ID == executionID {
			return execution.Status, nil
		}
	}

	return "", errors.New("execution not found")
}

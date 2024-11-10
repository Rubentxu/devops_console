// internal/infrastructure/repositories/memory_task_repository.go
package adapters

import (
	"context"
	"devops_console/internal/domain/entities/orchestrator"
	"devops_console/internal/ports/orchestrator"
	"fmt"
	"sync"
)

// Implementación sencilla de TaskRepository en memoria para el integration-tests
type InMemoryTaskRepository struct {
	tasks map[string]entities.DevOpsTask
	mu    sync.Mutex
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		tasks: make(map[string]entities.DevOpsTask),
	}
}

func (r *InMemoryTaskRepository) Create(ctx context.Context, task *entities.DevOpsTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = *task
	return nil
}

func (r *InMemoryTaskRepository) Update(ctx context.Context, task *entities.DevOpsTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = *task
	return nil
}

func (r *InMemoryTaskRepository) Delete(ctx context.Context, taskID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, taskID)
	return nil
}

func (r *InMemoryTaskRepository) GetByID(ctx context.Context, taskID string) (entities.DevOpsTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.tasks[taskID]
	if !ok {
		return entities.DevOpsTask{}, fmt.Errorf("task not found")
	}
	return task, nil
}

func (r *InMemoryTaskRepository) GetAll(ctx context.Context, filters ports.TaskFilters) ([]entities.DevOpsTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tasks := make([]entities.DevOpsTask, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *InMemoryTaskRepository) GetByExecutionID(ctx context.Context, executionID string) (entities.DevOpsTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, task := range r.tasks {
		for _, execution := range task.Executions {
			if execution.ID == executionID {
				return task, nil
			}
		}
	}
	return entities.DevOpsTask{}, fmt.Errorf("task not found")
}

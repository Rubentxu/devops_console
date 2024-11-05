// internal/ports/task_repository.go
package ports

import (
	"context"
	"devops_console/internal/domain/entities"
	"devops_console/internal/domain/workspace"
)

type TaskFilters struct {
	workspaceID string
	taskType    string
	subjects    []entities.Subject
}

type TaskUpdate struct {
	Name        string
	Description string
	Config      entities.TaskConfig
	workspace.Workspace
	entities.TaskType
	Approvals []entities.Approval
	Triggers  []entities.Trigger
}

// TaskRepository define las operaciones de persistencia para las tareas.
type TaskRepository interface {
	Create(ctx context.Context, task *entities.DevOpsTask) error
	GetByID(ctx context.Context, taskID string) (entities.DevOpsTask, error)
	Update(ctx context.Context, task *entities.DevOpsTask) error
	Delete(ctx context.Context, taskID string) error
	GetAll(ctx context.Context, filters TaskFilters) ([]entities.DevOpsTask, error)
	GetByExecutionID(ctx context.Context, executionID string) (entities.DevOpsTask, error)
}

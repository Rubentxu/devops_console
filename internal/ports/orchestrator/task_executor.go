// internal/ports/executor.go
package ports

import (
	"context"
	"devops_console/internal/domain/entities/orchestrator"
)

type TaskExecutor interface {
	// ExecuteTask: Ejecuta una tarea y devuelve un executionID que el cliente puede usar
	// para consultar el estado o suscribirse a eventos.
	ExecuteTask(ctx context.Context, task *entities.DevOpsTask) (string, error)
	GetTaskStatus(ctx context.Context, taskExecutionID string) (entities.TaskStatus, error)
	CancelTask(ctx context.Context, taskExecutionID string) error
	// SubscribeToTaskEvents: Permite al cliente suscribirse a los eventos de la tarea, incluyendo logs y cambios de estado.
	SubscribeToTaskEvents(taskExecutionID string) (<-chan entities.TaskEvent, error)
}

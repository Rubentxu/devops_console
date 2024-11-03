// internal/ports/scheduler.go
package ports

import (
	"context"
	"devops_console/internal/domain/entities"
)

// Scheduler define la interfaz para planificar y gestionar la ejecución de tareas y pipelines.
type TaskScheduler interface {
	ScheduleTask(ctx context.Context, task *entities.DevOpsTask) error
	CancelScheduledTask(ctx context.Context, taskID string) error
}

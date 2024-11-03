// internal/domain/entities/event_stream.go
package ports

import "devops_console/internal/domain/entities"

// TaskEventStream representa un flujo de eventos al que los consumidores pueden suscribirse.
type TaskEventStream interface {
	Subscribe(taskExecutionID string) (<-chan entities.TaskEvent, error)
	Publish(event entities.TaskEvent) error
	Close(taskExecutionID string)
}

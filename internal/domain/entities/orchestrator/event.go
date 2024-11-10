// internal/domain/entities/event.go
package entities

import "time"

type TaskEventType string

const (
	EventTypeTaskStarted     TaskEventType = "TaskStarted"
	EventTypeTaskProgress    TaskEventType = "TaskProgress"
	EventTypeTaskCompleted   TaskEventType = "TaskCompleted"
	EventTypeTaskFailed      TaskEventType = "TaskFailed"
	EventTypeTaskCanceled    TaskEventType = "TaskCanceled"
	EventTypeTaskOutput      TaskEventType = "TaskOutput"
	EventTypeTaskError       TaskEventType = "TaskError"
	EventTypePodName         TaskEventType = "POD_NAME"        // Nuevo tipo de evento
	EventTypeWorkerConnected TaskEventType = "WorkerConnected" // Nuevo tipo de evento
	// Otros tipos de eventos según sea necesario
)

type TaskEvent struct {
	ID          string
	ExecutionID string
	Timestamp   time.Time
	EventType   TaskEventType
	Payload     interface{}
}

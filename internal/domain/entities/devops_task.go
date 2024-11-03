// internal/domain/entities/task.go
package entities

import "time"

type TaskStatus string

const (
	TaskPending   TaskStatus = "PENDING"
	TaskRunning   TaskStatus = "RUNNING"
	TaskSucceeded TaskStatus = "SUCCEEDED"
	TaskFailed    TaskStatus = "FAILED"
	TaskCanceled  TaskStatus = "CANCELED"
	TaskError     TaskStatus = "ERROR"
)

type TaskType string

const (
	TaskTypeScheduled TaskType = "SCHEDULED"
	TaskTypeTriggered TaskType = "TRIGGERED"
	TaskTypeApproval  TaskType = "APPROVAL"
	TaskTypeManual    TaskType = "MANUAL"
)

type DevOpsTask struct {
	ID          string
	Name        string
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Config      TaskConfig
	Executions  []*TaskExecution
	Workspace   Workspace
	TaskType    TaskType
	Approvals   []*Approval
	Trigger     *Trigger
}

type TaskConfig struct {
	Parameters map[string]interface{}
	Workspace  string
}

type TaskExecution struct {
	ID             string
	DevOpsTaskID   string
	Status         TaskStatus
	StartedAt      time.Time
	FinishedAt     time.Time
	TaskExecutorID string
	Output         *Artifact
	Error          string
}

type Approval struct {
	ID         string
	UserID     string
	ApprovedAt time.Time
	Approved   bool
}

type Trigger interface {
	Evaluate() bool
}

type ScheduledTrigger struct {
	Expression string
}

func (t *ScheduledTrigger) Evaluate() bool {
	// Implementar la evaluación de la expresión
	return false
}

type Workspace struct {
	ID   string
	Name string
}

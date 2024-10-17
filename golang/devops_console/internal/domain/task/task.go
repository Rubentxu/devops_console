package domain

import (
	domain "devops_console/internal/domain/worker"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending           TaskStatus = "Pending"
	TaskStatusInProgress        TaskStatus = "InProgress"
	TaskStatusCompleted         TaskStatus = "Completed"
	TaskStatusFailed            TaskStatus = "Failed"
	TaskStatusScheduled         TaskStatus = "Scheduled"
	TaskStatusPendingValidation TaskStatus = "PendingValidation"
)

type Form struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Fields map[string]string `json:"fields"`
}

type Approval struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Approved     bool       `json:"approved"`
	ApprovalDate *time.Time `json:"approval_date,omitempty"`
}

type TaskExecuted struct {
	ID          string     `json:"id"`
	RunAt       time.Time  `json:"run_at"`
	WorkspaceID string     `json:"workspace_id"`
	Done        bool       `json:"done"`
	Status      TaskStatus `json:"status"`
}

type Task struct {
	ID            string              `json:"id"`
	CreatedAt     time.Time           `json:"created_at"`
	WorkspaceID   string              `json:"workspace_id"`
	Name          string              `json:"name"`
	TaskType      string              `json:"task_type"`
	Technology    string              `json:"technology"`
	WorkerType    string              `json:"worker_type,omitempty"`
	Description   string              `json:"description,omitempty"`
	ExtendedInfo  string              `json:"extended_info,omitempty"`
	Tags          *[]string           `json:"tags"`
	Forms         *[]Form             `json:"forms"`
	Approvals     *[]Approval         `json:"approvals"`
	Metadata      map[string]string   `json:"metadata"`
	TasksExecuted []TaskExecuted      `json:"tasks_executed"`
	WorkerConfig  domain.WorkerConfig `json:"worker_config"`
}

type TaskCreate struct {
	WorkspaceID  string              `json:"workspace_id"`
	Name         string              `json:"name"`
	TaskType     string              `json:"task_type"`
	Technology   string              `json:"technology"`
	Description  string              `json:"description,omitempty"`
	ExtendedInfo string              `json:"extended_info,omitempty"`
	Tags         *[]string           `json:"tags"`
	Forms        *[]Form             `json:"forms"`
	Approvals    *[]Approval         `json:"approvals"`
	Metadata     map[string]string   `json:"metadata"`
	WorkerConfig domain.WorkerConfig `json:"worker_config"`
}

type TaskUpdate struct {
	ID            string              `json:"id,omitempty"`
	Name          string              `json:"name,omitempty"`
	TaskType      string              `json:"task_type,omitempty"`
	Technology    string              `json:"technology,omitempty"`
	Description   string              `json:"description,omitempty"`
	ExtendedInfo  string              `json:"extended_info,omitempty"`
	Tags          *[]string           `json:"tags,omitempty"`
	Forms         *[]Form             `json:"forms,omitempty"`
	Approvals     *[]Approval         `json:"approvals,omitempty"`
	Metadata      map[string]string   `json:"metadata,omitempty"`
	TasksExecuted *[]TaskExecuted     `json:"tasks_executed,omitempty"`
	WorkerConfig  domain.WorkerConfig `json:"worker_config,omitempty"`
}

type TaskRepository interface {
	Create(task TaskCreate) (*Task, error)
	GetAll() ([]*Task, error)
	GetByID(taskID string) (*Task, error)
	Update(taskID string, taskUpdate TaskUpdate) (*Task, error)
	Delete(taskID string) error
}

package job

import (
	"time"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "Pending"
	JobStatusRunning   JobStatus = "Running"
	JobStatusCompleted JobStatus = "Completed"
	JobStatusFailed    JobStatus = "Failed"
)

type Job struct {
	ID        string                 `json:"id"`
	WorkerID  string                 `json:"worker_id"`
	Name      string                 `json:"name"`
	Status    JobStatus              `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	StartedAt *time.Time             `json:"started_at,omitempty"`
	FinishedAt *time.Time            `json:"finished_at,omitempty"`
	Result    *string                `json:"result,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type JobCreate struct {
	WorkerID string                 `json:"worker_id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

type JobUpdate struct {
	Status     *JobStatus              `json:"status,omitempty"`
	StartedAt  *time.Time              `json:"started_at,omitempty"`
	FinishedAt *time.Time              `json:"finished_at,omitempty"`
	Result     *string                 `json:"result,omitempty"`
	Metadata   map[string]interface{}  `json:"metadata,omitempty"`
}

type JobRepository interface {
	Create(job JobCreate) (*Job, error)
	GetAll() ([]*Job, error)
	GetByID(jobID string) (*Job, error)
	Update(jobID string, jobUpdate JobUpdate) (*Job, error)
	Delete(jobID string) error
	GetJobsByWorkerID(workerID string) ([]*Job, error)
}

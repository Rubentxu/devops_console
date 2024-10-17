package domain

import "context"

type WorkerType string

const (
	WorkerTypeJobKubernetes     WorkerType = "JobKubernetes"
	WorkerTypeCronJobKubernetes WorkerType = "CronJobKubernetes"
	WorkerTypeDocker            WorkerType = "Docker"
	WorkerTypePodman            WorkerType = "Podman"
)

type WorkerConfig map[string]interface{}

type Worker interface {
	LaunchJob(ctx context.Context, name string, config WorkerConfig) (string, error)
	GetJobStatus(ctx context.Context, name string) (string, error)
	MonitorJob(ctx context.Context, name string) (<-chan string, <-chan error)
	StreamLogs(ctx context.Context, name string) (<-chan string, <-chan error)
}

type WorkerFactoryInterface interface {
	GetWorker(workerType WorkerType) (Worker, error)
	RegisterWorker(workerType WorkerType, worker Worker)
}

type WorkerCreate struct {
	Name   string       `json:"name"`
	Type   WorkerType   `json:"type"`
	Config WorkerConfig `json:"config"`
}

type WorkerUpdate struct {
	Name   *string      `json:"name,omitempty"`
	Type   *string      `json:"type,omitempty"`
	Config WorkerConfig `json:"config,omitempty"`
}

type WorkerRepository interface {
	Create(worker WorkerCreate) (*Worker, error)
	GetAll() ([]*Worker, error)
	GetByID(workerID string) (*Worker, error)
	Update(workerID string, workerUpdate WorkerUpdate) (*Worker, error)
	Delete(workerID string) error
}

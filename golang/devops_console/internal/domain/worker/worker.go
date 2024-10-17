package worker

import (
	"github.com/google/uuid"
)

type WorkerType string

const (
    WorkerTypeKubernetes WorkerType = "Kubernetes"
    WorkerTypeOpenShift  WorkerType = "OpenShift"
    WorkerTypeDocker     WorkerType = "Docker"
    WorkerTypePodman     WorkerType = "Podman"
)

type Worker struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Type   WorkerType                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type WorkerCreate struct {
	Name   string                 `json:"name"`
	Type   WorkerType             `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type WorkerUpdate struct {
	Name   *string                `json:"name,omitempty"`
	Type   *string                `json:"type,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

func NewWorker(create WorkerCreate) *Worker {
    return &Worker{
        ID:     uuid.New().String(),
        Name:   create.Name,
        Type:   create.Type,
        Config: create.Config,
    }
}

type WorkerRepository interface {
    Create(worker WorkerCreate) (*Worker, error)
    GetAll() ([]*Worker, error)
    GetByID(workerID string) (*Worker, error)
    Update(workerID string, workerUpdate WorkerUpdate) (*Worker, error)
    Delete(workerID string) error
}

package infra

import (
	domain "devops_console/internal/domain/worker"
	"fmt"
	"sync"
)

type WorkerFactory struct {
	workers map[domain.WorkerType]domain.Worker
	mu      sync.RWMutex
}

func NewWorkerFactory() *WorkerFactory {
	return &WorkerFactory{
		workers: make(map[domain.WorkerType]domain.Worker),
	}
}

func (f *WorkerFactory) RegisterWorker(workerType domain.WorkerType, worker domain.Worker) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.workers[workerType] = worker
}

func (f *WorkerFactory) GetWorker(workerType domain.WorkerType) (domain.Worker, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	worker, exists := f.workers[workerType]
	if !exists {
		return nil, fmt.Errorf("worker type %s not registered", workerType)
	}
	return worker, nil
}

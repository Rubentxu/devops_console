package factories

import (
    "devops_console/internal/domain/worker"
    "devops_console/internal/infrastructure/workers"
    "errors"
)

type WorkerFactory struct {
    kubeconfig string
    namespace  string
}

func NewWorkerFactory(kubeconfig, namespace string) *WorkerFactory {
    return &WorkerFactory{
        kubeconfig: kubeconfig,
        namespace:  namespace,
    }
}

func (f *WorkerFactory) GetJobLauncher(workerType worker.WorkerType) (workers.JobLauncher, error) {
    switch workerType {
    case worker.WorkerTypeOpenShift:
        return workers.NewOpenShiftWorker(f.kubeconfig, f.namespace)
    // Añadir casos para otros tipos de workers
    default:
        return nil, errors.New("unsupported worker type")
    }
}

func (f *WorkerFactory) GetJobMonitor(workerType worker.WorkerType) (workers.JobMonitor, error) {
    switch workerType {
    case worker.WorkerTypeOpenShift:
        return workers.NewOpenShiftWorker(f.kubeconfig, f.namespace)
    // Añadir casos para otros tipos de workers
    default:
        return nil, errors.New("unsupported worker type")
    }
}

func (f *WorkerFactory) GetLogStreamer(workerType worker.WorkerType) (workers.LogStreamer, error) {
    switch workerType {
    case worker.WorkerTypeOpenShift:
        return workers.NewOpenShiftWorker(f.kubeconfig, f.namespace)
    // Añadir casos para otros tipos de workers
    default:
        return nil, errors.New("unsupported worker type")
    }
}
package application

import (
    "context"
    "devops_console/internal/domain/worker"
    "devops_console/internal/infrastructure/workers/factories"
)

type JobService struct {
    workerFactory *factories.WorkerFactory
}

func NewJobService(kubeconfig, namespace string) *JobService {
    return &JobService{
        workerFactory: factories.NewWorkerFactory(kubeconfig, namespace),
    }
}

func (s *JobService) LaunchJob(ctx context.Context, name string, workerType worker.WorkerType, config map[string]interface{}) (string, error) {
    launcher, err := s.workerFactory.GetJobLauncher(workerType)
    if err != nil {
        return "", err
    }

    return launcher.LaunchJob(ctx, name, config)
}

func (s *JobService) MonitorJob(ctx context.Context, name string, workerType worker.WorkerType) (<-chan string, error) {
    monitor, err := s.workerFactory.GetJobMonitor(workerType)
    if err != nil {
        return nil, err
    }

    return monitor.MonitorJob(ctx, name)
}

func (s *JobService) StreamJobLogs(ctx context.Context, name string, workerType worker.WorkerType) (<-chan string, error) {
    streamer, err := s.workerFactory.GetLogStreamer(workerType)
    if err != nil {
        return nil, err
    }

    return streamer.StreamLogs(ctx, name)
}
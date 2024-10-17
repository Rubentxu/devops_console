package workers

import (
    "context"
)

type JobLauncher interface {
    LaunchJob(ctx context.Context, name string, config map[string]interface{}) (string, error)
}

type JobMonitor interface {
    GetJobStatus(ctx context.Context, name string) (string, error)
    MonitorJob(ctx context.Context, name string) (<-chan string, error)
}

type LogStreamer interface {
    StreamLogs(ctx context.Context, name string) (<-chan string, error)
}

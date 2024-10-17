package application

import (
    "context"
    "devops_console/internal/domain/task"
    "devops_console/internal/infrastructure/workers"
    "devops_console/internal/infrastructure/workers/factories"
    "reflect"
    "testing"
    "time"
)

// Mock para TaskRepository
type mockTaskRepository struct {
    tasks map[string]*task.Task
}

func (m *mockTaskRepository) Create(t task.TaskCreate) (*task.Task, error) {
    newTask := &task.Task{
        ID:   "mock-id",
        Name: t.Name,
        // ... otros campos ...
    }
    m.tasks[newTask.ID] = newTask
    return newTask, nil
}

func (m *mockTaskRepository) GetAll() ([]*task.Task, error) {
    tasks := make([]*task.Task, 0, len(m.tasks))
    for _, t := range m.tasks {
        tasks = append(tasks, t)
    }
    return tasks, nil
}

func (m *mockTaskRepository) GetByID(id string) (*task.Task, error) {
    return m.tasks[id], nil
}

// ... implementar otros métodos del repositorio ...

// Mock para JobLauncher
type mockJobLauncher struct{}

func (m *mockJobLauncher) LaunchJob(ctx context.Context, name string, config map[string]interface{}) (string, error) {
    return "Job launched", nil
}

// Mock para JobMonitor
type mockJobMonitor struct{}

func (m *mockJobMonitor) GetJobStatus(ctx context.Context, name string) (string, error) {
    return string(task.TaskStatusRunning), nil
}

func (m *mockJobMonitor) MonitorJob(ctx context.Context, name string) (<-chan string, error) {
    ch := make(chan string, 1)
    go func() {
        ch <- string(task.TaskStatusRunning)
        close(ch)
    }()
    return ch, nil
}

// Mock para LogStreamer
type mockLogStreamer struct{}

func (m *mockLogStreamer) StreamLogs(ctx context.Context, name string) (<-chan string, error) {
    ch := make(chan string, 1)
    go func() {
        ch <- "Log message"
        close(ch)
    }()
    return ch, nil
}

// Mock para WorkerFactory
type mockWorkerFactory struct{}

func (m *mockWorkerFactory) GetJobLauncher(workerType factories.WorkerType) factories.JobLauncher {
    return &mockJobLauncher{}
}

func (m *mockWorkerFactory) GetJobMonitor(workerType factories.WorkerType) factories.JobMonitor {
    return &mockJobMonitor{}
}

func (m *mockWorkerFactory) GetLogStreamer(workerType factories.WorkerType) factories.LogStreamer {
    return &mockLogStreamer{}
}

// Tests
func TestCreateTask(t *testing.T) {
    mockRepo := &mockTaskRepository{tasks: make(map[string]*task.Task)}
    service := NewTaskService(mockRepo, make(map[factories.WorkerType]workers.WorkerConfig))

    taskCreate := task.TaskCreate{
        Name: "Test Task",
        // ... otros campos ...
    }

    createdTask, err := service.CreateTask(taskCreate)
    if err != nil {
        t.Fatalf("Error creating task: %v", err)
    }

    if createdTask.Name != taskCreate.Name {
        t.Errorf("Expected task name %s, got %s", taskCreate.Name, createdTask.Name)
    }
}

func TestGetAllTasks(t *testing.T) {
    mockRepo := &mockTaskRepository{tasks: make(map[string]*task.Task)}
    service := NewTaskService(mockRepo, make(map[factories.WorkerType]workers.WorkerConfig))

    // Crear algunas tareas de prueba
    mockRepo.tasks["1"] = &task.Task{ID: "1", Name: "Task 1"}
    mockRepo.tasks["2"] = &task.Task{ID: "2", Name: "Task 2"}

    tasks, err := service.GetAllTasks()
    if err != nil {
        t.Fatalf("Error getting all tasks: %v", err)
    }

    if len(tasks) != 2 {
        t.Errorf("Expected 2 tasks, got %d", len(tasks))
    }
}

func TestExecuteTask(t *testing.T) {
    mockRepo := &mockTaskRepository{tasks: make(map[string]*task.Task)}
    mockWorkerConfigs := map[factories.WorkerType]workers.WorkerConfig{
        factories.WorkerTypeKubernetes: {},
    }
    service := NewTaskService(mockRepo, mockWorkerConfigs)
    service.workerFactory = &mockWorkerFactory{}

    mockRepo.tasks["1"] = &task.Task{ID: "1", Name: "Test Task", Technology: "kubernetes"}

    ctx := context.Background()
    result, err := service.ExecuteTask(ctx, "1")
    if err != nil {
        t.Fatalf("Error executing task: %v", err)
    }

    expectedResult := map[string]interface{}{
        "type":    "success",
        "message": "Job launched",
    }

    if !reflect.DeepEqual(result, expectedResult) {
        t.Errorf("Expected result %v, got %v", expectedResult, result)
    }
}

func TestExecuteTaskWithTimeout(t *testing.T) {
    mockRepo := &mockTaskRepository{tasks: make(map[string]*task.Task)}
    mockWorkerConfigs := map[factories.WorkerType]workers.WorkerConfig{
        factories.WorkerTypeKubernetes: {},
    }
    service := NewTaskService(mockRepo, mockWorkerConfigs)
    service.workerFactory = &mockWorkerFactory{}

    mockRepo.tasks["1"] = &task.Task{ID: "1", Name: "Test Task", Technology: "kubernetes"}

    ctx := context.Background()
    result, err := service.ExecuteTaskWithTimeout(ctx, "1", 5)
    if err != nil {
        t.Fatalf("Error executing task with timeout: %v", err)
    }

    expectedResult := map[string]interface{}{
        "type":    "success",
        "message": "Job launched",
    }

    if !reflect.DeepEqual(result, expectedResult) {
        t.Errorf("Expected result %v, got %v", expectedResult, result)
    }
}

func TestDetermineWorkerType(t *testing.T) {
    service := NewTaskService(nil, nil)

    testCases := []struct {
        task           *task.Task
        expectedWorker factories.WorkerType
    }{
        {
            task:           &task.Task{Technology: "kubernetes"},
            expectedWorker: factories.WorkerTypeKubernetes,
        },
        {
            task:           &task.Task{Technology: "docker"},
            expectedWorker: factories.WorkerTypeDocker,
        },
        {
            task:           &task.Task{Tags: []string{"openshift"}},
            expectedWorker: factories.WorkerTypeOpenShift,
        },
        {
            task:           &task.Task{TaskType: "deployment"},
            expectedWorker: factories.WorkerTypeKubernetes,
        },
        {
            task:           &task.Task{TaskType: "build"},
            expectedWorker: factories.WorkerTypeDocker,
        },
    }

    for _, tc := range testCases {
        result := service.DetermineWorkerType(tc.task)
        if result != tc.expectedWorker {
            t.Errorf("For task %+v, expected worker type %v, but got %v", tc.task, tc.expectedWorker, result)
        }
    }
}
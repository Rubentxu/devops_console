package interfaces

import (
    "context"
    "devops_console/internal/application"
    "devops_console/internal/infrastructure/config"
    "devops_console/internal/infrastructure/repositories"
    "devops_console/internal/infrastructure/workers"
     "devops_console/internal/domain/task"
)

type App struct {
    ctx            context.Context
    taskService    *application.TaskService
    tenantService  *application.TenantService
    workerService  *application.WorkerService
    jobService     *application.JobService
    workspaceService *application.WorkspaceService
}

func NewApp() *App {
    return &App{}
}

func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx

    // Inicializar repositorios
    taskRepo := repositories.NewInMemoryTaskRepository()
    // tenantRepo := repositories.NewInMemoryTenantRepository()
    // workerRepo := repositories.NewInMemoryWorkerRepository()
    // jobRepo := repositories.NewInMemoryJobRepository()
    // workspaceRepo := repositories.NewInMemoryWorkspaceRepository()

    // Cargar configuraciones de workers
    workerConfigs := workers.LoadWorkerConfigs("worker_config.json")

    // Inicializar servicios
    a.taskService = application.NewTaskService(taskRepo, workerConfigs)
    // a.tenantService = application.NewTenantService(tenantRepo)
    // a.workerService = application.NewWorkerService(workerRepo)
    // a.jobService = application.NewJobService(jobRepo)
    // a.workspaceService = application.NewWorkspaceService(workspaceRepo)

    // Cargar datos de desarrollo si es necesario
    if config.IsDevelopmentMode() {
        a.loadDevData()
    }
}

func (a *App) loadDevData() {
    // Implementar la carga de datos de desarrollo aquí
}

// Métodos expuestos al frontend

func (a *App) CreateTask(task task.TaskCreate) (*task.Task, error) {
    return a.taskService.CreateTask(task)
}

func (a *App) GetAllTasks() ([]*task.Task, error) {
    return a.taskService.GetAllTasks()
}

func (a *App) ExecuteTask(taskID string) (map[string]interface{}, error) {
    return a.taskService.ExecuteTask(a.ctx, taskID)
}

// Implementar métodos similares para otros servicios (Tenant, Worker, Job, Workspace)
package interfaces

import (
	"context"
	"devops_console/internal/application"
	domain "devops_console/internal/domain/task"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
    ctx         context.Context
    taskService *application.TaskService
}

func NewApp() *App {
    return &App{}
}

func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
    // Inicializa taskService y otros servicios aquí
}

// CreateTask crea una nueva tarea
func (a *App) CreateTask(taskCreate domain.TaskCreate) (*domain.Task, error) {
    task, err := a.taskService.CreateTask(taskCreate)
    if err == nil {
        runtime.EventsEmit(a.ctx, "task:created", task)
    }
    return task, err
}

// GetAllTasks obtiene todas las tareas
func (a *App) GetAllTasks() ([]*domain.Task, error) {
    return a.taskService.GetAllTasks()
}

// GetTaskByID obtiene una tarea por su ID
func (a *App) GetTaskByID(taskID string) (*domain.Task, error) {
    return a.taskService.GetTaskByID(taskID)
}

// UpdateTask actualiza una tarea existente
func (a *App) UpdateTask(taskID string, taskUpdate domain.TaskUpdate) (*domain.Task, error) {
    task, err := a.taskService.UpdateTask(taskID, taskUpdate)
    if err == nil {
        runtime.EventsEmit(a.ctx, "task:updated", task)
    }
    return task, err
}

// DeleteTask elimina una tarea
func (a *App) DeleteTask(taskID string) error {
    err := a.taskService.DeleteTask(taskID)
    if err == nil {
        runtime.EventsEmit(a.ctx, "task:deleted", taskID)
    }
    return err
}

// AddTaskExecution añade una ejecución a una tarea
func (a *App) AddTaskExecution(taskID string, taskExecuted domain.TaskExecuted) (*domain.Task, error) {
    task, err := a.taskService.AddTaskExecution(taskID, taskExecuted)
    if err == nil {
        runtime.EventsEmit(a.ctx, "task:execution:added", map[string]interface{}{
            "taskID": taskID,
            "execution": taskExecuted,
        })
    }
    return task, err
}

// ExecuteTask ejecuta una tarea
func (a *App) ExecuteTask(taskID string) error {
    go func() {
        result, err := a.taskService.ExecuteTask(a.ctx, taskID)
        if err != nil {
            runtime.EventsEmit(a.ctx, "task:execution:error", map[string]string{
                "taskID": taskID,
                "error":  err.Error(),
            })
        } else {
            runtime.EventsEmit(a.ctx, "task:execution:result", map[string]interface{}{
                "taskID": taskID,
                "result": result,
            })
        }
    }()
    return nil
}

// ExecuteTaskWithTimeout ejecuta una tarea con un tiempo límite
func (a *App) ExecuteTaskWithTimeout(taskID string, timeoutSeconds int) error {
    go func() {
        result, err := a.taskService.ExecuteTaskWithTimeout(a.ctx, taskID, timeoutSeconds)
        if err != nil {
            runtime.EventsEmit(a.ctx, "task:execution:error", map[string]string{
                "taskID": taskID,
                "error":  err.Error(),
            })
        } else {
            runtime.EventsEmit(a.ctx, "task:execution:result", map[string]interface{}{
                "taskID": taskID,
                "result": result,
            })
        }
    }()
    return nil
}

// PauseTask pausa una tarea
func (a *App) PauseTask(taskID string) error {
    result, err := a.taskService.PauseTask(taskID)
    if err == nil {
        runtime.EventsEmit(a.ctx, "task:paused", map[string]interface{}{
            "taskID": taskID,
            "result": result,
        })
    }
    return err
}

// ResumeTask reanuda una tarea pausada
func (a *App) ResumeTask(taskID string) error {
    result, err := a.taskService.ResumeTask(taskID)
    if err == nil {
        runtime.EventsEmit(a.ctx, "task:resumed", map[string]interface{}{
            "taskID": taskID,
            "result": result,
        })
    }
    return err
}

// GetTaskStatus obtiene el estado actual de una tarea
func (a *App) GetTaskStatus(taskID string) (string, error) {
    return a.taskService.GetTaskStatus(a.ctx, taskID)
}

// GetTaskStatistics obtiene estadísticas generales de las tareas
func (a *App) GetTaskStatistics() map[string]interface{} {
    return a.taskService.GetTaskStatistics()
}

// MonitorTask inicia el monitoreo de una tarea
func (a *App) MonitorTask(taskID string) error {
    go func() {
        statusChan, err := a.taskService.MonitorTask(a.ctx, taskID)
        if err != nil {
            runtime.EventsEmit(a.ctx, "task:monitor:error", map[string]string{
                "taskID": taskID,
                "error":  err.Error(),
            })
            return
        }

        for status := range statusChan {
            runtime.EventsEmit(a.ctx, "task:status:update", map[string]string{
                "taskID": taskID,
                "status": status,
            })
        }
    }()
    return nil
}

// StreamTaskLogs inicia la transmisión de logs de una tarea
func (a *App) StreamTaskLogs(taskID string) error {
    go func() {
        logChan, err := a.taskService.StreamTaskLogs(a.ctx, taskID)
        if err != nil {
            runtime.EventsEmit(a.ctx, "task:logs:error", map[string]string{
                "taskID": taskID,
                "error":  err.Error(),
            })
            return
        }

        for log := range logChan {
            runtime.EventsEmit(a.ctx, "task:log", map[string]string{
                "taskID": taskID,
                "log":    log,
            })
        }
    }()
    return nil
}

// ExecuteAndMonitorTask ejecuta y monitorea una tarea
func (a *App) ExecuteAndMonitorTask(taskID string) error {
    go func() {
        resultChan, err := a.taskService.ExecuteAndMonitorTask(a.ctx, taskID)
        if err != nil {
            runtime.EventsEmit(a.ctx, "task:error", map[string]string{"taskID": taskID, "error": err.Error()})
            return
        }

        for result := range resultChan {
            runtime.EventsEmit(a.ctx, "task:update", map[string]interface{}{
                "taskID": taskID,
                "data":   result,
            })
        }

        runtime.EventsEmit(a.ctx, "task:complete", map[string]string{"taskID": taskID})
    }()

    return nil
}
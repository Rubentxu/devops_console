package repositories

import (
    "devops_console/internal/domain/task"
    "sync"
)

type InMemoryTaskRepository struct {
    tasks map[string]*task.Task
    mutex sync.RWMutex
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
    return &InMemoryTaskRepository{
        tasks: make(map[string]*task.Task),
    }
}

func (r *InMemoryTaskRepository) Create(taskCreate task.TaskCreate) (*task.Task, error) {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    newTask := task.NewTask(taskCreate)
    r.tasks[newTask.ID] = newTask
    return newTask, nil
}

func (r *InMemoryTaskRepository) GetAll() ([]*task.Task, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()

    tasks := make([]*task.Task, 0, len(r.tasks))
    for _, t := range r.tasks {
        tasks = append(tasks, t)
    }
    return tasks, nil
}

// Implementar otros métodos del repositorio
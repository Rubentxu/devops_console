package infra

import (
	domain "devops_console/internal/domain/task"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"sync"
	"time"
)

type JSONTaskRepository struct {
	filePath string
	tasks    map[string]*domain.Task
	mutex    sync.RWMutex
}

func NewJSONTaskRepository(filePath string) (*JSONTaskRepository, error) {
	repo := &JSONTaskRepository{
		filePath: filePath,
		tasks:    make(map[string]*domain.Task),
	}
	if err := repo.loadFromFile(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JSONTaskRepository) loadFromFile() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file yet, it's okay for the first run
		}
		return err
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	if len(byteValue) > 0 {
		if err := json.Unmarshal(byteValue, &r.tasks); err != nil {
			return err
		}
	}
	return nil
}

func (r *JSONTaskRepository) saveToFile() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	byteValue, err := json.MarshalIndent(r.tasks, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(r.filePath, byteValue, 0644); err != nil {
		return err
	}
	return nil
}

func (r *JSONTaskRepository) Create(taskCreate domain.TaskCreate) (*domain.Task, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	newID := uuid.New().String() // Generar un nuevo ID
	newTask := &domain.Task{
		ID:           newID,
		CreatedAt:    time.Now(), // Asegúrate de que CreatedAt es parte de domain.Task
		WorkspaceID:  taskCreate.WorkspaceID,
		Name:         taskCreate.Name,
		TaskType:     taskCreate.TaskType,
		Technology:   taskCreate.Technology,
		Description:  taskCreate.Description,
		ExtendedInfo: taskCreate.ExtendedInfo,
		Tags:         taskCreate.Tags,
		Forms:        taskCreate.Forms,
		Approvals:    taskCreate.Approvals,
		Metadata:     taskCreate.Metadata,
		WorkerConfig: taskCreate.WorkerConfig,
	}
	r.tasks[newID] = newTask

	if err := r.saveToFile(); err != nil {
		return nil, err
	}

	return newTask, nil
}

func (r *JSONTaskRepository) GetAll() ([]*domain.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tasks := make([]*domain.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *JSONTaskRepository) GetByID(id string) (*domain.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", id)
	}
	return task, nil
}

func (r *JSONTaskRepository) Update(id string, taskUpdate domain.TaskUpdate) (*domain.Task, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", id)
	}

	task.Name = taskUpdate.Name
	task.TaskType = taskUpdate.TaskType
	task.Technology = taskUpdate.Technology
	task.Description = taskUpdate.Description
	task.ExtendedInfo = taskUpdate.ExtendedInfo
	task.Tags = taskUpdate.Tags
	task.Forms = taskUpdate.Forms
	task.Approvals = taskUpdate.Approvals
	task.Metadata = taskUpdate.Metadata

	if err := r.saveToFile(); err != nil {
		return nil, err
	}

	return task, nil
}

func (r *JSONTaskRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return fmt.Errorf("task with ID %s not found", id)
	}
	delete(r.tasks, id)

	if err := r.saveToFile(); err != nil {
		return err
	}

	return nil
}

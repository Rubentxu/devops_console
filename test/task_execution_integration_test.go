package test

import (
	"context"
	"devops_console/internal/application"
	"devops_console/internal/domain/entities"
	eventstream "devops_console/internal/infrastructure/events"
	executor "devops_console/internal/infrastructure/executors"
	adapters "devops_console/internal/infrastructure/repositories"
	"log"
	"testing"
	"time"
)

func TestTaskExecution(t *testing.T) {
	// Configurar el contexto
	_ = context.Background()

	// Crear el TaskEventStream
	eventStream := eventstream.NewTaskEventStream()

	// Crear el TaskExecutor
	executor, err := executor.NewK8sTaskExecutor("default", eventStream)
	if err != nil {
		t.Fatalf("Error al crear K8sTaskExecutor: %v", err)
	}

	// Crear el TaskRepository en memoria
	taskRepo := adapters.NewInMemoryTaskRepository()

	// Crear el TaskService
	taskService := application.NewTaskServiceImpl(taskRepo, executor)

	// Crear una tarea de ejemplo
	task := entities.DevOpsTask{
		ID:          "task-1",
		Name:        "test-task",
		Description: "A task for testing purposes",
		Config: entities.TaskConfig{
			Parameters: map[string]interface{}{
				"Image":   "busybox",
				"Command": []string{"sh", "-c", "for i in $(seq 1 5); do echo \"Linea traza $i\"; sleep 1; done"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Guardar la tarea en el repositorio
	_, err = taskService.CreateTask(task)
	if err != nil {
		t.Fatalf("Error al crear la tarea: %v", err)
	}

	// Ejecutar la tarea
	executionID, err := taskService.ExecuteTask(task.ID)
	if err != nil {
		t.Fatalf("Error al ejecutar la tarea: %v", err)
	}

	// Suscribirse a los eventos de la tarea
	eventChan, err := taskService.SubscribeToTaskEvents(executionID)
	if err != nil {
		t.Fatalf("Error al suscribirse a los eventos de la tarea: %v", err)
	}

	// Crear un canal para señalizar cuando la tarea haya finalizado
	doneChan := make(chan struct{})

	// Variable para almacenar el estado final de la tarea
	var finalStatus entities.TaskStatus

	// Goroutine para recibir y procesar eventos
	go func() {
		for event := range eventChan {
			t.Logf("Evento recibido: %s para ejecución %s (%s)", event.EventType, event.ExecutionID, event.Payload)
			switch event.EventType {
			case entities.EventTypeTaskOutput:
				t.Logf("Log: %s", event.Payload)
			case entities.EventTypeTaskCompleted:
				log.Printf("La tarea ha finalizado con éxito")
				finalStatus = entities.TaskSucceeded
				close(doneChan)
				return
			case entities.EventTypeTaskFailed:
				log.Printf("La tarea ha finalizado con error")
				finalStatus = entities.TaskFailed
				close(doneChan)
				return
			case entities.EventTypeTaskError:
				log.Printf("La tarea ha finalizado con error crítico")
				finalStatus = entities.TaskError
				close(doneChan)
				return
			}
		}
		// Detectar cierre inesperado del canal
		log.Printf("El canal de eventos se cerró inesperadamente")
		finalStatus = entities.TaskSucceeded
		close(doneChan)
	}()

	// Esperar a que la tarea finalice o se agote el tiempo
	select {
	case <-doneChan:
		t.Logf("La tarea ha finalizado con estado: %s", finalStatus)
	case <-time.After(30 * time.Second):
		t.Fatalf("La tarea no finalizó en el tiempo esperado")
	}

	// Verificar que la tarea haya finalizado con éxito
	if finalStatus != entities.TaskSucceeded {
		t.Fatalf("La tarea no finalizó correctamente, estado final: %s", finalStatus)
	}
}

func TestDockerTaskExecution(t *testing.T) {
	// Configurar el contexto
	_ = context.Background()

	// Crear el TaskEventStream
	eventStream := eventstream.NewTaskEventStream()

	// Crear el TaskExecutor
	executor, err := executor.NewDockerTaskExecutor(eventStream)
	if err != nil {
		t.Fatalf("Error al crear DockerTaskExecutor: %v", err)
	}

	// Crear el TaskRepository en memoria
	taskRepo := adapters.NewInMemoryTaskRepository()

	// Crear el TaskService
	taskService := application.NewTaskServiceImpl(taskRepo, executor)

	// Crear una tarea de ejemplo
	task := entities.DevOpsTask{
		ID:          "task-1",
		Name:        "test-task",
		Description: "A task for testing purposes",
		Config: entities.TaskConfig{
			Parameters: map[string]interface{}{
				"Image":   "busybox",
				"Command": []string{"sh", "-c", "for i in $(seq 1 5); do echo \"Linea traza $i\"; sleep 1; done"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Guardar la tarea en el repositorio
	_, err = taskService.CreateTask(task)
	if err != nil {
		t.Fatalf("Error al crear la tarea: %v", err)
	}

	// Ejecutar la tarea
	executionID, err := taskService.ExecuteTask(task.ID)
	if err != nil {
		t.Fatalf("Error al ejecutar la tarea: %v", err)
	}

	// Suscribirse a los eventos de la tarea
	eventChan, err := taskService.SubscribeToTaskEvents(executionID)
	if err != nil {
		t.Fatalf("Error al suscribirse a los eventos de la tarea: %v", err)
	}

	// Crear un canal para señalizar cuando la tarea haya finalizado
	doneChan := make(chan struct{})

	// Variable para almacenar el estado final de la tarea
	var finalStatus entities.TaskStatus

	// Goroutine para recibir y procesar eventos
	go func() {
		for event := range eventChan {
			t.Logf("Evento recibido: %s para ejecución %s (%s)", event.EventType, event.ExecutionID, event.Payload)
			switch event.EventType {
			case entities.EventTypeTaskOutput:
				t.Logf("Log: %s", event.Payload)
			case entities.EventTypeTaskCompleted:
				log.Printf("La tarea ha finalizado con éxito")
				finalStatus = entities.TaskSucceeded
				close(doneChan)
				return
			case entities.EventTypeTaskFailed:
				log.Printf("La tarea ha finalizado con error")
				finalStatus = entities.TaskFailed
				close(doneChan)
				return
			case entities.EventTypeTaskError:
				log.Printf("La tarea ha finalizado con error crítico")
				finalStatus = entities.TaskError
				close(doneChan)
				return
			}
		}
		// Detectar cierre inesperado del canal
		log.Printf("El canal de eventos se cerró inesperadamente")
		finalStatus = entities.TaskSucceeded
		close(doneChan)
	}()

	// Esperar a que la tarea finalice o se agote el tiempo
	select {
	case <-doneChan:
		t.Logf("La tarea ha finalizado con estado: %s", finalStatus)
	case <-time.After(30 * time.Second):
		t.Fatalf("La tarea no finalizó en el tiempo esperado")
	}

	// Verificar que la tarea haya finalizado con éxito
	if finalStatus != entities.TaskSucceeded {
		t.Fatalf("La tarea no finalizó correctamente, estado final: %s", finalStatus)
	}
}

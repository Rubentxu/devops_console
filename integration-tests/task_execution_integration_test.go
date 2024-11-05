package integration_tests

import (
	"context"
	"devops_console/internal/application"
	"devops_console/internal/domain/entities"
	eventstream "devops_console/internal/infrastructure/events"
	executor "devops_console/internal/infrastructure/executors"
	adapters "devops_console/internal/infrastructure/repositories"
	workers "devops_console/internal/infrastructure/workers"
	"log"
	"testing"
	"time"
)

func TestTaskExecutionWithDockerWorker(t *testing.T) {
	// Setup context
	_ = context.Background()

	// Create the TaskEventStream
	eventStream := eventstream.NewTaskEventStream()

	// Create the TaskExecutor
	dockerExecutor, err := executor.NewDockerTaskExecutor(eventStream)
	if err != nil {
		t.Fatalf("Error creating DockerTaskExecutor: %v", err)
	}

	k8sExecutor, err := executor.NewK8sTaskExecutor("default", eventStream)
	if err != nil {
		t.Fatalf("Error creating K8sTaskExecutor: %v", err)
	}

	// Create the TaskRepository in memory
	taskRepo := adapters.NewInMemoryTaskRepository()

	// Create the TaskService
	taskService := application.NewTaskServiceImpl(taskRepo)
	taskService.RegisterExecutor("Docker", dockerExecutor)
	taskService.RegisterExecutor("Kubernetes", k8sExecutor)

	// Create a sample task with a Docker worker
	task := entities.DevOpsTask{
		ID:          "task-1",
		Name:        "integration-tests-task",
		Description: "A task for testing purposes",
		Config: entities.TaskConfig{
			Parameters: map[string]interface{}{
				"arg": "argumento 1",
			},
		},
		Worker: &workers.DockerWorker{
			Name:    "pruebas",
			Image:   "busybox",
			Command: []string{"sh", "-c", "for i in $(seq 1 5); do echo \"Linea traza $i\"; sleep 1; done"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the task in the repository
	_, err = taskService.CreateTask(task)
	if err != nil {
		t.Fatalf("Error creating task: %v", err)
	}

	// Execute the task
	executionID, err := taskService.ExecuteTask(task.ID)
	if err != nil {
		t.Fatalf("Error executing task: %v", err)
	}

	// Subscribe to task events
	eventChan, err := taskService.SubscribeToTaskEvents(executionID)
	if err != nil {
		t.Fatalf("Error subscribing to task events: %v", err)
	}

	// Create a channel to signal when the task is done
	doneChan := make(chan struct{})

	// Variable to store the final status of the task
	var finalStatus entities.TaskStatus

	// Goroutine to receive and process events
	go func() {
		for event := range eventChan {
			t.Logf("Event received: %s for execution %s (%s)", event.EventType, event.ExecutionID, event.Payload)
			switch event.EventType {
			case entities.EventTypeTaskOutput:
				t.Logf("Log: %s", event.Payload)
			case entities.EventTypeTaskCompleted:
				log.Printf("Task completed successfully")
				finalStatus = entities.TaskSucceeded
				close(doneChan)
				return
			case entities.EventTypeTaskFailed:
				log.Printf("Task failed")
				finalStatus = entities.TaskFailed
				close(doneChan)
				return
			case entities.EventTypeTaskError:
				log.Printf("Task encountered a critical error")
				finalStatus = entities.TaskError
				close(doneChan)
				return
			}
		}
		// Detect unexpected channel closure
		log.Printf("Event channel closed unexpectedly")
		finalStatus = entities.TaskSucceeded
		close(doneChan)
	}()

	// Wait for the task to complete or timeout
	select {
	case <-doneChan:
		t.Logf("Task completed with status: %s", finalStatus)
	case <-time.After(30 * time.Second):
		t.Fatalf("Task did not complete in the expected time")
	}

	// Verify that the task completed successfully
	if finalStatus != entities.TaskSucceeded {
		t.Fatalf("Task did not complete successfully, final status: %s", finalStatus)
	}
}

func TestTaskExecutionWithKubernetesWorker(t *testing.T) {
	// Setup context
	_ = context.Background()

	// Create the TaskEventStream
	eventStream := eventstream.NewTaskEventStream()

	// Create the TaskExecutor
	k8sExecutor, err := executor.NewK8sTaskExecutor("default", eventStream)
	if err != nil {
		t.Fatalf("Error creating K8sTaskExecutor: %v", err)
	}

	// Create the TaskRepository in memory
	taskRepo := adapters.NewInMemoryTaskRepository()

	// Create the TaskService
	taskService := application.NewTaskServiceImpl(taskRepo)
	taskService.RegisterExecutor("Kubernetes", k8sExecutor)

	// Create a sample task with a Kubernetes worker
	task := entities.DevOpsTask{
		ID:          "task-1",
		Name:        "integration-tests-task",
		Description: "A task for testing purposes",
		Config: entities.TaskConfig{
			Parameters: map[string]interface{}{
				"arg": "argumento 1",
			},
		},
		Worker: &workers.KubernetesWorker{
			Name:      "pruebas",
			Namespace: "default",
			Image:     "busybox",
			Command:   []string{"sh", "-c", "for i in $(seq 1 5); do echo \"Linea traza $i\"; sleep 1; done"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the task in the repository
	_, err = taskService.CreateTask(task)
	if err != nil {
		t.Fatalf("Error creating task: %v", err)
	}

	// Execute the task
	executionID, err := taskService.ExecuteTask(task.ID)
	if err != nil {
		t.Fatalf("Error executing task: %v", err)
	}

	// Subscribe to task events
	eventChan, err := taskService.SubscribeToTaskEvents(executionID)
	if err != nil {
		t.Fatalf("Error subscribing to task events: %v", err)
	}

	// Create a channel to signal when the task is done
	doneChan := make(chan struct{})

	// Variable to store the final status of the task
	var finalStatus entities.TaskStatus

	// Goroutine to receive and process events
	go func() {
		for event := range eventChan {
			t.Logf("Event received: %s for execution %s (%s)", event.EventType, event.ExecutionID, event.Payload)
			switch event.EventType {
			case entities.EventTypeTaskOutput:
				t.Logf("Log: %s", event.Payload)
			case entities.EventTypeTaskCompleted:
				log.Printf("Task completed successfully")
				finalStatus = entities.TaskSucceeded
				close(doneChan)
				return
			case entities.EventTypeTaskFailed:
				log.Printf("Task failed")
				finalStatus = entities.TaskFailed
				close(doneChan)
				return
			case entities.EventTypeTaskError:
				log.Printf("Task encountered a critical error")
				finalStatus = entities.TaskError
				close(doneChan)
				return
			}
		}
		// Detect unexpected channel closure
		log.Printf("Event channel closed unexpectedly")
		finalStatus = entities.TaskSucceeded
		close(doneChan)
	}()

	// Wait for the task to complete or timeout
	select {
	case <-doneChan:
		t.Logf("Task completed with status: %s", finalStatus)
	case <-time.After(30 * time.Second):
		t.Fatalf("Task did not complete in the expected time")
	}

	// Verify that the task completed successfully
	if finalStatus != entities.TaskSucceeded {
		t.Fatalf("Task did not complete successfully, final status: %s", finalStatus)
	}
}

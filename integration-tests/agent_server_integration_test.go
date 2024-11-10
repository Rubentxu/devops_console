package integration_tests

import (
	"context"
	"devops_console/internal/domain/entities/orchestrator"
	agentClient "devops_console/internal/infrastructure/agent"
	pb "devops_console/internal/infrastructure/agent/proto/agent/v1"
	eventstream "devops_console/internal/infrastructure/orchestrator/events"
	agent "devops_console/internal/infrastructure/orchestrator/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

const bufSize = 1024 * 1024

func TestAgentServerCommandExecutionWithAgent(t *testing.T) {
	ctx := context.Background()

	// Create an in-memory listener for testing
	lis := bufconn.Listen(bufSize)
	log.Println("In-memory listener created")

	// Create the event stream
	eventStream := eventstream.NewTaskEventStream()
	log.Println("Event stream created")

	// Create and start the server
	agentServer := agent.NewAgentServer(eventStream)
	grpcServer := grpc.NewServer()
	pb.RegisterAgentServiceServer(grpcServer, agentServer)

	log.Println("gRPC server created and agent service registered")

	// Start the server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}()
	log.Println("gRPC server started")

	// Helper function to create client connections
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	// Create client connection
	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	log.Println("Client connection created")

	// Create and start the agent
	agentID := "test-agent-1"
	agent := agentClient.NewAgent(agentID, "bufnet", insecure.NewCredentials())
	go func() {
		if err := agent.Start(ctx); err != nil {
			t.Fatalf("Failed to start agent: %v", err)
		} else {
			log.Println("Agent started")
		}
	}()

	// Test commands
	testCommands := []struct {
		id      string
		command string
		args    []string
	}{
		{"cmd1", "echo", []string{"Hello World"}},
		{"cmd2", "ls", []string{"-la"}},
		{"cmd3", "pwd", []string{}},
	}
	log.Println("Test commands defined")

	var wg sync.WaitGroup

	// Map to store event channels by command
	eventChannels := make(map[string]<-chan entities.TaskEvent)

	// Subscribe to events for each command
	for _, cmd := range testCommands {
		eventChan, err := eventStream.Subscribe(cmd.id)
		if err != nil {
			t.Fatalf("Failed to subscribe to events for command %s: %v", cmd.id, err)
		}
		eventChannels[cmd.id] = eventChan
		log.Printf("Subscribed to events for command %s", cmd.id)
	}

	// Execute commands
	for _, cmd := range testCommands {
		wg.Add(1)
		go func(cmdID, command string, args []string) {
			defer wg.Done()
			log.Printf("Executing command %s: %s %v", cmdID, command, args)

			// Send command to the agent
			agentServer.TaskQueue <- &pb.Command{
				CommandId: cmdID,
				Command:   command,
				Args:      args,
			}
			log.Printf("Command sent to agent: %s", cmdID)

			// Wait for events from the agent
			for event := range eventChannels[cmdID] {
				t.Logf("Event received: %s for command %s (%s)", event.EventType, event.ExecutionID, event.Payload)
				if event.EventType == entities.EventTypeTaskCompleted {
					break
				}
			}
		}(cmd.id, cmd.command, cmd.args)
	}

	// Wait for all commands to complete
	wg.Wait()
	log.Println("All commands executed")

	// Cleanup
	for _, cmd := range testCommands {
		eventStream.Close(cmd.id)
	}
	grpcServer.Stop()
	log.Println("gRPC server stopped and cleanup done")
}

func TestAgentServerCommandExecution(t *testing.T) {
	ctx := context.Background()

	// Create an in-memory listener for testing
	lis := bufconn.Listen(bufSize)

	// Create the event stream
	eventStream := eventstream.NewTaskEventStream()

	// Create and start the server
	agentServer := agent.NewAgentServer(eventStream)
	grpcServer := grpc.NewServer()
	pb.RegisterAgentServiceServer(grpcServer, agentServer)

	// Start the server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}()

	// Helper function to create client connections
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	// Create client connection
	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewAgentServiceClient(conn)

	// Simulate agent connection
	agentID := "test-agent-1"
	_, err = client.Connect(ctx, &pb.ConnectRequest{
		AgentId: agentID,
		SystemInfo: &pb.SystemInfo{
			Hostname: "test-host",
			Os:       "linux",
			Arch:     "amd64",
		},
	})
	if err != nil {
		t.Fatalf("Failed to connect agent: %v", err)
	}

	// Test commands
	testCommands := []struct {
		id      string
		command string
		args    []string
	}{
		{"cmd1", "echo", []string{"Hello World"}},
		{"cmd2", "ls", []string{"-la"}},
		{"cmd3", "pwd", []string{}},
	}

	var wg sync.WaitGroup

	// Map to store event channels by command
	eventChannels := make(map[string]<-chan entities.TaskEvent)

	// Subscribe to events for each command
	for _, cmd := range testCommands {
		eventChan, err := eventStream.Subscribe(cmd.id)
		if err != nil {
			t.Fatalf("Failed to subscribe to events for command %s: %v", cmd.id, err)
		}
		eventChannels[cmd.id] = eventChan
	}

	// Execute commands
	for _, cmd := range testCommands {
		wg.Add(1)
		go func(cmdID, command string, args []string) {
			defer wg.Done()

			// Send command to the agent
			agentServer.TaskQueue <- &pb.Command{
				CommandId: cmdID,
				Command:   command,
				Args:      args,
			}

			// Simulate execution and sending events from the agent
			events := []struct {
				eventType pb.EventType
				payload   string
			}{
				{pb.EventType_STARTED, "Command started"},
				{pb.EventType_OUTPUT, "Command output"},
				{pb.EventType_COMPLETED, "Command completed"},
			}

			for _, evt := range events {
				_, err := client.SendEvent(ctx, &pb.ExecutionEvent{
					CommandId: cmdID,
					Type:      evt.eventType,
					Payload:   evt.payload,
					Timestamp: time.Now().UnixNano(),
				})
				if err != nil {
					t.Errorf("Failed to send event for command %s: %v", cmdID, err)
				}
				time.Sleep(100 * time.Millisecond) // Simulate delay between events
			}
		}(cmd.id, cmd.command, cmd.args)
	}

	// Verify events for each command
	for cmdID, eventChan := range eventChannels {
		wg.Add(1)
		go func(cmdID string, ch <-chan entities.TaskEvent) {
			defer wg.Done()

			expectedEvents := []entities.TaskEventType{
				entities.EventTypeTaskStarted,
				entities.EventTypeTaskOutput,
				entities.EventTypeTaskCompleted,
			}

			eventIndex := 0
			for event := range ch {
				if event.ExecutionID != cmdID {
					t.Errorf("Received event for wrong command. Expected %s, got %s", cmdID, event.ExecutionID)
					continue
				}

				if event.EventType != expectedEvents[eventIndex] {
					t.Errorf("Wrong event order for command %s. Expected %v, got %v",
						cmdID, expectedEvents[eventIndex], event.EventType)
				}

				eventIndex++
				if eventIndex == len(expectedEvents) {
					break
				}
			}

			if eventIndex != len(expectedEvents) {
				t.Errorf("Did not receive all expected events for command %s. Got %d of %d",
					cmdID, eventIndex, len(expectedEvents))
			}
		}(cmdID, eventChan)
	}

	// Wait for all commands and verifications to complete
	wg.Wait()

	// Cleanup
	for _, cmd := range testCommands {
		eventStream.Close(cmd.id)
	}
	grpcServer.Stop()
}

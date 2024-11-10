package agent

import (
	"context"
	pb "devops_console/internal/infrastructure/agent/proto/agent/v1"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

const bufSize = 1024 * 1024

type Agent struct {
	id            string
	masterAddr    string
	credentials   credentials.TransportCredentials
	executor      *CommandExecutor
	metrics       *MetricsCollector
	reconnectWait time.Duration
}

func NewAgent(id, masterAddr string, creds credentials.TransportCredentials) *Agent {
	return &Agent{
		id:            id,
		masterAddr:    masterAddr,
		credentials:   creds,
		executor:      NewCommandExecutor(),
		metrics:       NewMetricsCollector(),
		reconnectWait: 5 * time.Second,
	}
}

func (a *Agent) Start(ctx context.Context) error {
	for {
		log.Printf("Agent.Start(). Connecting to master at %s", a.masterAddr)
		if err := a.connect(ctx); err != nil {
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled: %v", ctx.Err())
				return ctx.Err()
			case <-time.After(a.reconnectWait):
				log.Printf("Error connecting to master: %v. Retrying...", err)
				continue
			}
		}
	}
}

//func (a *Agent) connect(ctx context.Context) error {
//	log.Printf("Agent.connect(): Connecting to master at %s", a.masterAddr)
//
//	// Create an in-memory listener for testing
//	lis := bufconn.Listen(bufSize)
//
//	conn, err := grpc.NewClient(
//		a.masterAddr,
//		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
//			return lis.Dial()
//		}),
//		grpc.WithTransportCredentials(a.credentials),
//	)
//	if err != nil {
//		return fmt.Errorf("failed to connect: %v", err)
//	}
//
//	client := pb.NewAgentServiceClient(conn)
//	err = a.runEventLoop(ctx, client)
//	conn.Close() // Close the connection after the event loop ends
//	return err
//}

func (a *Agent) connect(ctx context.Context) error {
	log.Printf("Agent.connect(): Connecting to master at %s", a.masterAddr)

	// Create an in-memory listener for testing
	lis := bufconn.Listen(bufSize)

	// Create a dialer function that uses the bufconn listener
	bufDialer := func(ctx context.Context, address string) (net.Conn, error) {
		return lis.Dial()
	}

	// Establish the client connection using grpc.Dial
	conn, err := grpc.NewClient(
		"bufnet", // The address here is arbitrary but must match between client and dialer
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Use insecure credentials for testing
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewAgentServiceClient(conn)
	err = a.runEventLoop(ctx, client)
	return err
}

func (a *Agent) runEventLoop(ctx context.Context, client pb.AgentServiceClient) error {
	log.Println("Starting event loop")
	cmdStream, err := client.Connect(ctx, &pb.ConnectRequest{
		AgentId:      a.id,
		AgentVersion: "1.0.0",
		SystemInfo:   a.collectSystemInfo(),
	})
	if err != nil {
		return err
	}

	go a.startMetricsReporting(ctx, client)

	go func() {
		for event := range a.executor.eventChan {
			log.Printf("Sending event: %v", event)
			if _, err := client.SendEvent(ctx, event); err != nil {
				log.Printf("Error sending event: %v", err)
			}
		}
	}()

	for {
		cmd, err := cmdStream.Recv()
		if err != nil {
			log.Printf("Error receiving command: %v", err)
			return err
		}
		log.Printf("Received command: %v", cmd)
		go a.executor.Execute(ctx, cmd)
	}
}

func (a *Agent) collectSystemInfo() *pb.SystemInfo {
	return &pb.SystemInfo{
		Hostname: "hostname",
		Os:       "linux",
		Arch:     "amd64",
	}
}

func (a *Agent) startMetricsReporting(ctx context.Context, client pb.AgentServiceClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := a.metrics.Collect()
			if _, err := client.SendMetrics(ctx, metrics); err != nil {
				fmt.Printf("Error sending metrics: %v\n", err)
			}
		}
	}
}

type CommandExecutor struct {
	eventChan chan *pb.ExecutionEvent
}

func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{
		eventChan: make(chan *pb.ExecutionEvent, 100),
	}
}

func (e *CommandExecutor) Execute(ctx context.Context, cmd *pb.Command) {
	startTime := time.Now()
	e.eventChan <- &pb.ExecutionEvent{
		CommandId: cmd.CommandId,
		Type:      pb.EventType_STARTED,
		Payload:   "Command started",
		Timestamp: startTime.UnixNano(),
	}

	command := exec.CommandContext(ctx, cmd.Command, cmd.Args...)
	if len(cmd.Environment) > 0 {
		envSlice := make([]string, 0, len(cmd.Environment))
		for k, v := range cmd.Environment {
			envSlice = append(envSlice, k+"="+v)
		}
		command.Env = append(command.Env, envSlice...)
	}

	output, err := command.CombinedOutput()
	log.Printf("Log Command output: %s", output)
	if err != nil {
		e.eventChan <- &pb.ExecutionEvent{
			CommandId: cmd.CommandId,
			Type:      pb.EventType_ERROR,
			Payload:   err.Error(),
			Timestamp: time.Now().UnixNano(),
		}
	} else {
		e.eventChan <- &pb.ExecutionEvent{
			CommandId: cmd.CommandId,
			Type:      pb.EventType_OUTPUT,
			Payload:   strings.TrimSpace(string(output)),
			Timestamp: time.Now().UnixNano(),
		}
	}

	e.eventChan <- &pb.ExecutionEvent{
		CommandId: cmd.CommandId,
		Type:      pb.EventType_COMPLETED,
		Payload:   "Command completed",
		Timestamp: time.Now().UnixNano(),
	}
}

type MetricsCollector struct{}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (m *MetricsCollector) Collect() *pb.MetricsUpdate {
	// Simulate metrics collection
	return &pb.MetricsUpdate{
		AgentId:   "agent-1",
		Timestamp: time.Now().UnixNano(),
		System: &pb.SystemMetrics{
			CpuUsage:    10.5,
			MemoryUsage: 2048,
			DiskUsage:   50000,
		},
	}
}

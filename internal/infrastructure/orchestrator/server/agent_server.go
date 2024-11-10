package server

import (
	"context"
	entities "devops_console/internal/domain/entities/orchestrator"
	pb "devops_console/internal/infrastructure/agent/proto/agent/v1"
	ports "devops_console/internal/ports/orchestrator"
	"fmt"
	"log"
	"sync"
	"time"
)

// AgentServer es el servidor que maneja las conexiones de los agentes.
type AgentServer struct {
	pb.UnimplementedAgentServiceServer
	agents      map[string]*ConnectedAgent
	eventStream ports.TaskEventStream
	TaskQueue   chan *pb.Command
	mu          sync.RWMutex
}

// ConnectedAgent representa un agente conectado.
type ConnectedAgent struct {
	ID        string
	Info      *pb.SystemInfo
	Stream    pb.AgentService_ConnectServer
	Connected bool
}

// NewAgentServer crea una nueva instancia de AgentServer.
func NewAgentServer(eventStream ports.TaskEventStream) *AgentServer {
	return &AgentServer{
		agents:      make(map[string]*ConnectedAgent),
		eventStream: eventStream,
		TaskQueue:   make(chan *pb.Command, 100),
	}
}

// Connect maneja la conexión de un agente.
func (s *AgentServer) Connect(req *pb.ConnectRequest, stream pb.AgentService_ConnectServer) error {
	s.mu.Lock()
	agent := &ConnectedAgent{
		ID:        req.AgentId,
		Info:      req.SystemInfo,
		Stream:    stream,
		Connected: true,
	}
	s.agents[agent.ID] = agent
	s.mu.Unlock()

	log.Printf("Agent %s connected with system info: %+v", agent.ID, agent.Info)

	defer func() {
		s.mu.Lock()
		delete(s.agents, agent.ID)
		s.mu.Unlock()
		log.Printf("Agent %s disconnected", agent.ID)
	}()

	// Enviar evento de conexión
	s.eventStream.Publish(entities.TaskEvent{
		ID:        agent.ID,
		EventType: entities.EventTypeWorkerConnected,
		Payload:   "Worker connected successfully",
	})

	// Esperar comandos y enviarlos al agente
	for cmd := range s.TaskQueue {
		if err := stream.Send(cmd); err != nil {
			log.Printf("Error sending command to agent %s: %v", agent.ID, err)
			return err
		}
		log.Printf("Command sent to agent %s: %+v", agent.ID, cmd)
	}

	return nil
}

// SendEvent maneja el envío de eventos desde el agente.
func (s *AgentServer) SendEvent(ctx context.Context, event *pb.ExecutionEvent) (*pb.EventAck, error) {
	log.Printf("Received event from agent: %s, type: %s, payload: %s", event.CommandId, event.Type, event.Payload)

	// Convertir y publicar en EventStream
	s.eventStream.Publish(entities.TaskEvent{
		ID:        event.CommandId,
		EventType: mapEventType(event.Type),
		Payload:   event.Payload,
		Timestamp: time.Unix(0, event.Timestamp),
	})
	log.Printf("Event received from agent %s: %+v", event.CommandId, event)
	return &pb.EventAck{}, nil
}

// SendMetrics maneja el envío de métricas desde el agente.
func (s *AgentServer) SendMetrics(ctx context.Context, metrics *pb.MetricsUpdate) (*pb.MetricsAck, error) {
	// Procesar métricas recibidas
	s.eventStream.Publish(entities.TaskEvent{
		ID:        metrics.AgentId,
		EventType: mapEventType(pb.EventType_METRICS),
		Payload: fmt.Sprintf("CPU: %.2f%%, Memory: %.2f%%",
			metrics.System.CpuUsage,
			metrics.System.MemoryUsage),
		Timestamp: time.Unix(0, metrics.Timestamp),
	})
	log.Printf("Metrics received from agent %s: CPU: %.2f%%, Memory: %.2f%%", metrics.AgentId, metrics.System.CpuUsage, metrics.System.MemoryUsage)
	return &pb.MetricsAck{}, nil
}

// mapEventType convierte el tipo de evento de protobuf a un tipo de evento interno.
func mapEventType(pbEventType pb.EventType) entities.TaskEventType {
	switch pbEventType {
	case pb.EventType_STARTED:
		return entities.EventTypeTaskStarted
	case pb.EventType_OUTPUT:
		return entities.EventTypeTaskOutput
	case pb.EventType_ERROR:
		return entities.EventTypeTaskError
	case pb.EventType_COMPLETED:
		return entities.EventTypeTaskCompleted
	case pb.EventType_FAILED:
		return entities.EventTypeTaskFailed
	default:
		return entities.EventTypeTaskProgress
	}
}

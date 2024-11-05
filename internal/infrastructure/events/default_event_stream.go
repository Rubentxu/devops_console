package adapters

import (
	"devops_console/internal/domain/entities"
	"log"
	"sync"
)

type TaskEventStreamImpl struct {
	subscribers map[string][]chan entities.TaskEvent
	mu          sync.RWMutex
}

func NewTaskEventStream() *TaskEventStreamImpl {
	return &TaskEventStreamImpl{
		subscribers: make(map[string][]chan entities.TaskEvent),
	}
}

func (es *TaskEventStreamImpl) Subscribe(taskExecutionID string) (<-chan entities.TaskEvent, error) {
	ch := make(chan entities.TaskEvent, 100)
	es.mu.Lock()
	es.subscribers[taskExecutionID] = append(es.subscribers[taskExecutionID], ch)
	es.mu.Unlock()
	return ch, nil
}

func (es *TaskEventStreamImpl) Publish(event entities.TaskEvent) error {
	es.mu.RLock()
	subs, ok := es.subscribers[event.ExecutionID]
	es.mu.RUnlock()
	if ok {
		for _, ch := range subs {
			ch <- event
			if event.EventType == entities.EventTypeTaskCompleted || event.EventType == entities.EventTypeTaskFailed || event.EventType == entities.EventTypeTaskError {
				log.Printf("Cerrando canal para ejecución: %s", event.ExecutionID)
				close(ch)
			}
		}
		if event.EventType == entities.EventTypeTaskCompleted || event.EventType == entities.EventTypeTaskFailed || event.EventType == entities.EventTypeTaskError {
			es.mu.Lock()
			delete(es.subscribers, event.ExecutionID)
			es.mu.Unlock()
		}
	}
	return nil
}

func (s *TaskEventStreamImpl) Close(taskExecutionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if subs, ok := s.subscribers[taskExecutionID]; ok {
		for _, ch := range subs {
			close(ch)
		}
		delete(s.subscribers, taskExecutionID)
	}
}

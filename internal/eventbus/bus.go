package eventbus

import (
	"context"
	"sync"
	"time"
)

type EventType string

const (
	EventIncidentCreated      EventType = "incident.created"
	EventIncidentUpdated      EventType = "incident.updated"
	EventRCATriggered         EventType = "rca.triggered"
	EventRCACompleted         EventType = "rca.completed"
	EventRemediationCreated   EventType = "remediation.created"
	EventRemediationStarted   EventType = "remediation.started"
	EventRemediationCompleted EventType = "remediation.completed"
	EventChaosStarted         EventType = "chaos.started"
	EventChaosCompleted       EventType = "chaos.completed"
)

type Event struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

type EventBus interface {
	Publish(event Event)
	Subscribe(types ...EventType) (<-chan Event, func())
	Close()
}

type bus struct {
	mu          sync.RWMutex
	subscribers map[int]subEntry
	nextID      int
}

type subEntry struct {
	ch     chan Event
	types  map[EventType]bool
}

func New() EventBus {
	return &bus{
		subscribers: make(map[int]subEntry),
	}
}

func (b *bus) Publish(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		if sub.types == nil || sub.types[event.Type] {
			select {
			case sub.ch <- event:
			default:
			}
		}
	}
}

func (b *bus) Subscribe(types ...EventType) (<-chan Event, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	ch := make(chan Event, 64)
	typeSet := make(map[EventType]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	b.subscribers[id] = subEntry{ch: ch, types: typeSet}

	unsub := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if sub, ok := b.subscribers[id]; ok {
			close(sub.ch)
			delete(b.subscribers, id)
		}
	}

	return ch, unsub
}

func (b *bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, sub := range b.subscribers {
		close(sub.ch)
	}
	b.subscribers = make(map[int]subEntry)
}

var _ EventBus = (*bus)(nil)

var _ = context.Background

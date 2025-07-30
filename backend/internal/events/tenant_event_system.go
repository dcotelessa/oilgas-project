// backend/internal/events/tenant_event_system.go
package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type EventType string

const (
	InventoryCreated EventType = "inventory.created"
	InventoryUpdated EventType = "inventory.updated"
	InventoryDeleted EventType = "inventory.deleted"
	CustomerCreated  EventType = "customer.created"
	CustomerUpdated  EventType = "customer.updated"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type EventHandler func(ctx context.Context, event *Event) error

type EventBus struct {
	handlers map[EventType][]EventHandler
	mutex    sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(ctx context.Context, event *Event) error {
	eb.mutex.RLock()
	handlers := eb.handlers[event.Type]
	eb.mutex.RUnlock()
	
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("event handler failed: %w", err)
		}
	}
	
	return nil
}

func EventBusMiddleware(eventBus *EventBus) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("event_bus", eventBus)
		c.Next()
	}
}

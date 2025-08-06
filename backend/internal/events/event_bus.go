// backend/internal/events/event_bus.go
package events

import (
    "context"
    "encoding/json"
    "log"
    "reflect"
    "sync"
    "time"
)

// Event represents a domain event
type Event interface {
    EventType() string
    EventID() string
    TenantID() string
    Timestamp() time.Time
    Data() interface{}
}

// BaseEvent provides common event functionality
type BaseEvent struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"`
    Tenant    string    `json:"tenant_id"`
    CreatedAt time.Time `json:"timestamp"`
    Payload   interface{} `json:"data"`
}

func (e BaseEvent) EventType() string { return e.Type }
func (e BaseEvent) EventID() string { return e.ID }
func (e BaseEvent) TenantID() string { return e.Tenant }
func (e BaseEvent) Timestamp() time.Time { return e.CreatedAt }
func (e BaseEvent) Data() interface{} { return e.Payload }

// EventHandler processes events
type EventHandler func(ctx context.Context, event Event) error

// EventBus coordinates event publishing and subscription
type EventBus struct {
    handlers map[string][]EventHandler
    mutex    sync.RWMutex
    store    EventStore // For audit trail
}

func NewEventBus(store EventStore) *EventBus {
    return &EventBus{
        handlers: make(map[string][]EventHandler),
        store:    store,
    }
}

// Subscribe registers an event handler for a specific event type
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
    eb.mutex.Lock()
    defer eb.mutex.Unlock()
    
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish sends an event to all registered handlers
func (eb *EventBus) Publish(ctx context.Context, event Event) error {
    // Store event for audit trail
    if err := eb.store.Store(ctx, event); err != nil {
        log.Printf("Failed to store event %s: %v", event.EventID(), err)
        // Don't fail the operation, but log the issue
    }
    
    eb.mutex.RLock()
    handlers := eb.handlers[event.EventType()]
    eb.mutex.RUnlock()
    
    // Process handlers asynchronously to avoid blocking
    for _, handler := range handlers {
        go func(h EventHandler) {
            if err := h(ctx, event); err != nil {
                log.Printf("Event handler failed for %s: %v", event.EventType(), err)
            }
        }(handler)
    }
    
    return nil
}

// EventStore persists events for audit trails
type EventStore interface {
    Store(ctx context.Context, event Event) error
    GetEventsByTenant(ctx context.Context, tenantID string, filters EventFilters) ([]Event, error)
    GetEventsByEntity(ctx context.Context, entityType, entityID string) ([]Event, error)
}

// Database-backed event store
type DatabaseEventStore struct {
    authDB *sqlx.DB // Events stored in central auth DB for cross-tenant visibility
}

func NewDatabaseEventStore(authDB *sqlx.DB) *DatabaseEventStore {
    return &DatabaseEventStore{authDB: authDB}
}

func (es *DatabaseEventStore) Store(ctx context.Context, event Event) error {
    eventData, err := json.Marshal(event.Data())
    if err != nil {
        return fmt.Errorf("failed to marshal event data: %w", err)
    }
    
    query := `
        INSERT INTO events (id, event_type, tenant_id, data, created_at)
        VALUES ($1, $2, $3, $4, $5)`
    
    _, err = es.authDB.ExecContext(ctx, query,
        event.EventID(),
        event.EventType(),
        event.TenantID(),
        eventData,
        event.Timestamp(),
    )
    
    return err
}

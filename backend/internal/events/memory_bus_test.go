// backend/internal/events/memory_bus_test.go
package events_test

import (
	"context"
	"testing"
	"time"

	"oilgas-backend/internal/events"
)

func TestMemoryBus(t *testing.T) {
	bus := events.NewMemoryBus()
	
	received := false
	bus.Subscribe("test.event", func(ctx context.Context, event *events.Event) error {
		received = true
		return nil
	})
	
	event := &events.Event{
		ID:        "test-1",
		Type:      "test.event",
		TenantID:  "tenant1",
		Data:      map[string]interface{}{"test": true},
		Timestamp: time.Now(),
	}
	
	err := bus.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !received {
		t.Error("Event handler was not called")
	}
}


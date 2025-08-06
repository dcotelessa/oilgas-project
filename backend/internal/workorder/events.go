// backend/internal/workorder/events.go
package workorder

import (
    "time"
    "github.com/google/uuid"
    "oilgas-backend/internal/shared/events"
)

// Work Order Domain Events
type WorkOrderCreatedEvent struct {
    events.BaseEvent
    WorkOrderID     int    `json:"work_order_id"`
    CustomerID      int    `json:"customer_id"`
    ServiceType     string `json:"service_type"`
    CreatedByUserID int    `json:"created_by_user_id"`
}

func NewWorkOrderCreatedEvent(tenantID string, workOrderID, customerID, createdBy int, serviceType string) *WorkOrderCreatedEvent {
    return &WorkOrderCreatedEvent{
        BaseEvent: events.BaseEvent{
            ID:        uuid.New().String(),
            Type:      "workorder.created",
            Tenant:    tenantID,
            CreatedAt: time.Now(),
        },
        WorkOrderID:     workOrderID,
        CustomerID:      customerID,
        ServiceType:     serviceType,
        CreatedByUserID: createdBy,
    }
}

type WorkOrderStatusChangedEvent struct {
    events.BaseEvent
    WorkOrderID int    `json:"work_order_id"`
    OldStatus   string `json:"old_status"`
    NewStatus   string `json:"new_status"`
    ChangedBy   int    `json:"changed_by_user_id"`
    Notes       string `json:"notes"`
}

type WorkOrderItemCompletedEvent struct {
    events.BaseEvent
    WorkOrderID     int    `json:"work_order_id"`
    ItemID          int    `json:"item_id"`
    InventoryItemID *int   `json:"inventory_item_id"`
    ServiceNotes    string `json:"service_notes"`
    CompletedBy     int    `json:"completed_by_user_id"`
}

type InvoiceGeneratedEvent struct {
    events.BaseEvent
    WorkOrderID int     `json:"work_order_id"`
    InvoiceID   int     `json:"invoice_id"`
    Amount      float64 `json:"amount"`
    GeneratedBy int     `json:"generated_by_user_id"`
}

// Event Handlers for Cross-Domain Coordination
func RegisterWorkOrderEventHandlers(eventBus *events.EventBus, inventoryService InventoryService) {
    // When work order item is completed, update inventory status
    eventBus.Subscribe("workorder.item.completed", func(ctx context.Context, event events.Event) error {
        if itemEvent, ok := event.(*WorkOrderItemCompletedEvent); ok {
            if itemEvent.InventoryItemID != nil {
                return inventoryService.UpdateItemStatus(ctx, event.TenantID(), 
                    *itemEvent.InventoryItemID, "SERVICED")
            }
        }
        return nil
    })
    
    // When invoice is generated, notify customer service
    eventBus.Subscribe("invoice.generated", func(ctx context.Context, event events.Event) error {
        if invoiceEvent, ok := event.(*InvoiceGeneratedEvent); ok {
            // Could send email, update customer billing status, etc.
            log.Printf("Invoice %d generated for work order %d", 
                invoiceEvent.InvoiceID, invoiceEvent.WorkOrderID)
        }
        return nil
    })
}

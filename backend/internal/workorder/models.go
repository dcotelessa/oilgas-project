// backend/internal/workorder/models.go
package workorder

import (
    "time"
    "encoding/json"
)

// WorkOrder represents a service work order
type WorkOrder struct {
    ID               int                    `json:"id" db:"id"`
    TenantID         string                 `json:"tenant_id" db:"tenant_id"`
    CustomerID       int                    `json:"customer_id" db:"customer_id"`
    
    // Work Order Details
    WorkOrderNumber  string                 `json:"work_order_number" db:"work_order_number"`
    ServiceType      ServiceType            `json:"service_type" db:"service_type"`
    Status           WorkOrderStatus        `json:"status" db:"status"`
    Priority         Priority               `json:"priority" db:"priority"`
    
    // Service Details
    Description      string                 `json:"description" db:"description"`
    Instructions     *string                `json:"instructions" db:"instructions"`
    EstimatedHours   *float64               `json:"estimated_hours" db:"estimated_hours"`
    ActualHours      *float64               `json:"actual_hours" db:"actual_hours"`
    
    // Pricing & Billing
    HourlyRate       *float64               `json:"hourly_rate" db:"hourly_rate"`
    MaterialsCost    *float64               `json:"materials_cost" db:"materials_cost"`
    TotalAmount      *float64               `json:"total_amount" db:"total_amount"`
    
    // Assignment & Tracking
    AssignedToUserID *int                   `json:"assigned_to_user_id" db:"assigned_to_user_id"`
    CreatedByUserID  int                    `json:"created_by_user_id" db:"created_by_user_id"`
    
    // Dates & Timeline
    ScheduledDate    *time.Time             `json:"scheduled_date" db:"scheduled_date"`
    StartedAt        *time.Time             `json:"started_at" db:"started_at"`
    CompletedAt      *time.Time             `json:"completed_at" db:"completed_at"`
    DueDate          *time.Time             `json:"due_date" db:"due_date"`
    
    // Metadata
    IsActive         bool                   `json:"is_active" db:"is_active"`
    CreatedAt        time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
    
    // Relationships (loaded separately)
    Items            []WorkOrderItem        `json:"items,omitempty"`
    History          []WorkOrderHistory     `json:"history,omitempty"`
    Approvals        []WorkOrderApproval    `json:"approvals,omitempty"`
}

// ServiceType defines the type of service being performed
type ServiceType string

const (
    ServiceInspection    ServiceType = "INSPECTION"
    ServiceMaintenance   ServiceType = "MAINTENANCE" 
    ServiceRepair        ServiceType = "REPAIR"
    ServiceCleaning      ServiceType = "CLEANING"
    ServiceTesting       ServiceType = "TESTING"
    ServiceCustom        ServiceType = "CUSTOM"
)

// WorkOrderStatus defines the current state of the work order
type WorkOrderStatus string

const (
    StatusDraft       WorkOrderStatus = "DRAFT"
    StatusPending     WorkOrderStatus = "PENDING"        // Awaiting approval
    StatusApproved    WorkOrderStatus = "APPROVED"       // Ready to start
    StatusInProgress  WorkOrderStatus = "IN_PROGRESS"    // Work started
    StatusCompleted   WorkOrderStatus = "COMPLETED"      // Work finished
    StatusInvoiced    WorkOrderStatus = "INVOICED"       // Invoice generated
    StatusPaid        WorkOrderStatus = "PAID"           // Payment received
    StatusCancelled   WorkOrderStatus = "CANCELLED"
    StatusOnHold      WorkOrderStatus = "ON_HOLD"
)

// Priority levels for work orders
type Priority string

const (
    PriorityLow      Priority = "LOW"
    PriorityMedium   Priority = "MEDIUM"  
    PriorityHigh     Priority = "HIGH"
    PriorityUrgent   Priority = "URGENT"
)

// WorkOrderItem represents inventory items included in the work order
type WorkOrderItem struct {
    ID               int     `json:"id" db:"id"`
    WorkOrderID      int     `json:"work_order_id" db:"work_order_id"`
    InventoryItemID  *int    `json:"inventory_item_id" db:"inventory_item_id"`
    
    // Item Details
    Description      string  `json:"description" db:"description"`
    Quantity         int     `json:"quantity" db:"quantity"`
    UnitPrice        *float64 `json:"unit_price" db:"unit_price"`
    TotalPrice       *float64 `json:"total_price" db:"total_price"`
    
    // Service Details
    ServiceNotes     *string `json:"service_notes" db:"service_notes"`
    IsCompleted      bool    `json:"is_completed" db:"is_completed"`
    CompletedAt      *time.Time `json:"completed_at" db:"completed_at"`
    
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// WorkOrderHistory tracks status changes and updates
type WorkOrderHistory struct {
    ID               int                    `json:"id" db:"id"`
    WorkOrderID      int                    `json:"work_order_id" db:"work_order_id"`
    ChangedByUserID  int                    `json:"changed_by_user_id" db:"changed_by_user_id"`
    
    Action           string                 `json:"action" db:"action"`
    OldValue         *string                `json:"old_value" db:"old_value"`
    NewValue         *string                `json:"new_value" db:"new_value"`
    Notes            *string                `json:"notes" db:"notes"`
    
    CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// WorkOrderApproval handles multi-level approval workflows
type WorkOrderApproval struct {
    ID               int                    `json:"id" db:"id"`
    WorkOrderID      int                    `json:"work_order_id" db:"work_order_id"`
    ApproverUserID   int                    `json:"approver_user_id" db:"approver_user_id"`
    
    ApprovalLevel    int                    `json:"approval_level" db:"approval_level"`
    Status           ApprovalStatus         `json:"status" db:"status"`
    Comments         *string                `json:"comments" db:"comments"`
    
    RequestedAt      time.Time              `json:"requested_at" db:"requested_at"`
    RespondedAt      *time.Time             `json:"responded_at" db:"responded_at"`
}

type ApprovalStatus string

const (
    ApprovalPending  ApprovalStatus = "PENDING"
    ApprovalApproved ApprovalStatus = "APPROVED"
    ApprovalRejected ApprovalStatus = "REJECTED"
)

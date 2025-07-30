// backend/internal/models/workorder.go
package models

import "time"

// ReceivedItem represents work orders and incoming materials
type ReceivedItem struct {
	ID             int        `json:"id" db:"id"`
	WorkOrder      *string    `json:"work_order" db:"work_order"`
	CustomerID     *int       `json:"customer_id" db:"customer_id"`
	Customer       *string    `json:"customer" db:"customer"`
	Joints         *int       `json:"joints" db:"joints"`
	Size           *string    `json:"size" db:"size"`
	Weight         *float64   `json:"weight" db:"weight"`
	Grade          *string    `json:"grade" db:"grade"`
	Connection     *string    `json:"connection" db:"connection"`
	Well           *string    `json:"well" db:"well"`
	Lease          *string    `json:"lease" db:"lease"`
	OrderedBy      *string    `json:"ordered_by" db:"ordered_by"`
	Notes          *string    `json:"notes" db:"notes"`
	DateReceived   *time.Time `json:"date_received" db:"date_received"`
	InProduction   bool       `json:"in_production" db:"in_production"`
	Complete       bool       `json:"complete" db:"complete"`
	TenantID       string     `json:"tenant_id" db:"tenant_id"`
	ImportedAt     *time.Time `json:"imported_at,omitempty" db:"imported_at"`
	Deleted        bool       `json:"deleted" db:"deleted"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// WorkOrder represents aggregated work order information
type WorkOrder struct {
	WorkOrder    *string `json:"work_order"`
	Customer     *string `json:"customer"`
	CustomerID   *int    `json:"customer_id"`
	ItemCount    int     `json:"item_count"`
	TotalJoints  int     `json:"total_joints"`
	TotalWeight  float64 `json:"total_weight"`
	StartDate    *string `json:"start_date"`
	LatestDate   *string `json:"latest_date"`
	Locations    *string `json:"locations"`
	Sizes        *string `json:"sizes"`
	Grades       *string `json:"grades"`
	Status       string  `json:"status"`
	TenantID     string  `json:"tenant_id"`
}

// WorkOrderDetails provides comprehensive work order information
type WorkOrderDetails struct {
	Summary          WorkOrder       `json:"summary"`
	InventoryItems   []InventoryItem `json:"inventory_items"`
	ReceivedItems    []ReceivedItem  `json:"received_items"`
	TotalValue       float64         `json:"total_value"`
	RelatedOrders    []string        `json:"related_work_orders,omitempty"`
}

// CreateReceivedRequest for API endpoints
type CreateReceivedRequest struct {
	WorkOrder    *string `json:"work_order"`
	CustomerID   *int    `json:"customer_id"`
	Joints       *int    `json:"joints"`
	Size         *string `json:"size"`
	Weight       *float64 `json:"weight"`
	Grade        *string `json:"grade"`
	Connection   *string `json:"connection"`
	Well         *string `json:"well"`
	Lease        *string `json:"lease"`
	OrderedBy    *string `json:"ordered_by"`
	Notes        *string `json:"notes"`
	InProduction bool    `json:"in_production"`
}

type UpdateReceivedRequest struct {
	WorkOrder    *string `json:"work_order"`
	CustomerID   *int    `json:"customer_id"`
	Joints       *int    `json:"joints"`
	Size         *string `json:"size"`
	Weight       *float64 `json:"weight"`
	Grade        *string `json:"grade"`
	Connection   *string `json:"connection"`
	Well         *string `json:"well"`
	Lease        *string `json:"lease"`
	OrderedBy    *string `json:"ordered_by"`
	Notes        *string `json:"notes"`
	InProduction *bool   `json:"in_production"`
	Complete     *bool   `json:"complete"`
}

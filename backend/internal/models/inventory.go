// backend/internal/models/inventory.go
package models

import "time"

// InventoryItem represents pipes, casings, equipment with tenant ownership
type InventoryItem struct {
	ID           int        `json:"id" db:"id"`
	WorkOrder    *string    `json:"work_order" db:"work_order"`
	RNumber      *string    `json:"r_number" db:"r_number"`
	CustomerID   *int       `json:"customer_id" db:"customer_id"`
	Customer     *string    `json:"customer" db:"customer"`
	Joints       *int       `json:"joints" db:"joints"`
	Rack         *string    `json:"rack" db:"rack"`
	Size         *string    `json:"size" db:"size"`
	Weight       *float64   `json:"weight" db:"weight"`
	Grade        *string    `json:"grade" db:"grade"`
	Connection   *string    `json:"connection" db:"connection"`
	CTD          *string    `json:"ctd" db:"ctd"`
	WString      *string    `json:"w_string" db:"w_string"`
	Color        *string    `json:"color" db:"color"`
	DateIn       *time.Time `json:"date_in" db:"date_in"`
	DateOut      *time.Time `json:"date_out" db:"date_out"`
	WellIn       *string    `json:"well_in" db:"well_in"`
	LeaseIn      *string    `json:"lease_in" db:"lease_in"`
	WellOut      *string    `json:"well_out" db:"well_out"`
	LeaseOut     *string    `json:"lease_out" db:"lease_out"`
	Location     *string    `json:"location" db:"location"`
	Notes        *string    `json:"notes" db:"notes"`
	TenantID     string     `json:"tenant_id" db:"tenant_id"`
	ImportedAt   *time.Time `json:"imported_at,omitempty" db:"imported_at"`
	Deleted      bool       `json:"deleted" db:"deleted"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// InventorySummary provides aggregated inventory metrics
type InventorySummary struct {
	TotalItems       int                `json:"total_items"`
	TotalJoints      int                `json:"total_joints"`
	TotalWeight      float64            `json:"total_weight"`
	AvailableItems   int                `json:"available_items"`
	UniqueCustomers  int                `json:"unique_customers"`
	UniqueWorkOrders int                `json:"unique_work_orders"`
	UniqueSizes      int                `json:"unique_sizes"`
	UniqueGrades     int                `json:"unique_grades"`
	LocationCounts   map[string]int     `json:"location_counts"`
	GradeCounts      map[string]int     `json:"grade_counts"`
	SizeCounts       map[string]int     `json:"size_counts"`
}

// CreateInventoryRequest for API endpoints
type CreateInventoryRequest struct {
	WorkOrder  *string  `json:"work_order"`
	RNumber    *string  `json:"r_number"`
	CustomerID *int     `json:"customer_id"`
	Joints     *int     `json:"joints"`
	Rack       *string  `json:"rack"`
	Size       *string  `json:"size"`
	Weight     *float64 `json:"weight"`
	Grade      *string  `json:"grade"`
	Connection *string  `json:"connection"`
	CTD        *string  `json:"ctd"`
	WString    *string  `json:"w_string"`
	Color      *string  `json:"color"`
	WellIn     *string  `json:"well_in"`
	LeaseIn    *string  `json:"lease_in"`
	Location   *string  `json:"location"`
	Notes      *string  `json:"notes"`
}

type UpdateInventoryRequest struct {
	WorkOrder  *string  `json:"work_order"`
	RNumber    *string  `json:"r_number"`
	CustomerID *int     `json:"customer_id"`
	Joints     *int     `json:"joints"`
	Rack       *string  `json:"rack"`
	Size       *string  `json:"size"`
	Weight     *float64 `json:"weight"`
	Grade      *string  `json:"grade"`
	Connection *string  `json:"connection"`
	CTD        *string  `json:"ctd"`
	WString    *string  `json:"w_string"`
	Color      *string  `json:"color"`
	WellOut    *string  `json:"well_out"`
	LeaseOut   *string  `json:"lease_out"`
	Location   *string  `json:"location"`
	Notes      *string  `json:"notes"`
}

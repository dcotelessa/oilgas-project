// backend/internal/models/inventory_models.go
package models

import (
	"fmt"
	"strings"
	"time"
)

// InventoryItem represents an inventory item
type InventoryItem struct {
	ID         int        `json:"id" db:"id"`
	Username   string     `json:"username" db:"username"`
	WorkOrder  string     `json:"work_order" db:"work_order"`
	RNumber    int        `json:"r_number" db:"r_number"`
	CustomerID int        `json:"customer_id" db:"customer_id"`
	Customer   string     `json:"customer" db:"customer"`
	Joints     int        `json:"joints" db:"joints"`
	Rack       string     `json:"rack" db:"rack"`
	Size       string     `json:"size" db:"size"`
	Weight     string     `json:"weight" db:"weight"`
	Grade      string     `json:"grade" db:"grade"`
	Connection string     `json:"connection" db:"connection"`
	CTD        bool       `json:"ctd" db:"ctd"`
	WString    bool       `json:"w_string" db:"w_string"`
	SWGCC      string     `json:"swgcc" db:"swgcc"`
	Color      string     `json:"color" db:"color"`
	CustomerPO string     `json:"customer_po" db:"customer_po"`
	Fletcher   string     `json:"fletcher" db:"fletcher"`
	DateIn     *time.Time `json:"date_in" db:"date_in"`
	DateOut    *time.Time `json:"date_out" db:"date_out"`
	WellIn     string     `json:"well_in" db:"well_in"`
	LeaseIn    string     `json:"lease_in" db:"lease_in"`
	WellOut    string     `json:"well_out" db:"well_out"`
	LeaseOut   string     `json:"lease_out" db:"lease_out"`
	Trucking   string     `json:"trucking" db:"trucking"`
	Trailer    string     `json:"trailer" db:"trailer"`
	Location   string     `json:"location" db:"location"`
	Notes      string     `json:"notes" db:"notes"`
	PCode      string     `json:"pcode" db:"pcode"`
	CN         int        `json:"cn" db:"cn"`
	OrderedBy  string     `json:"ordered_by" db:"ordered_by"`
	Deleted    bool       `json:"deleted" db:"deleted"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// InventorySummary represents dashboard summary data
type InventorySummary struct {
	TotalItems        int                    `json:"total_items"`
	TotalJoints       int                    `json:"total_joints"`
	ItemsByGrade      map[string]int         `json:"items_by_grade"`
	ItemsByCustomer   map[string]int         `json:"items_by_customer"`
	ItemsByLocation   map[string]int         `json:"items_by_location"`
	GradeDistribution map[string]int         `json:"grade_distribution"`
	RecentActivity    []InventoryItem        `json:"recent_activity"`
	TopCustomers      []Customer             `json:"top_customers"`
	LowStock          []InventoryItem        `json:"low_stock"`
	PendingInspection int                    `json:"pending_inspection"`
	InProduction      int                    `json:"in_production"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// IsActive returns true if item is not deleted and has positive joints
func (i *InventoryItem) IsActive() bool {
	return !i.Deleted && i.Joints > 0
}

// IsShipped returns true if item has been shipped (negative joints or date_out set)
func (i *InventoryItem) IsShipped() bool {
	return i.Joints < 0 || i.DateOut != nil
}

// GetStatus returns the current status of the inventory item
func (i *InventoryItem) GetStatus() string {
	if i.Deleted {
		return "deleted"
	}
	if i.IsShipped() {
		return "shipped"
	}
	if i.DateIn != nil {
		return "in_inventory"
	}
	return "pending"
}

// GetDaysInInventory returns how many days the item has been in inventory
func (i *InventoryItem) GetDaysInInventory() int {
	if i.DateIn == nil {
		return 0
	}
	
	endDate := time.Now()
	if i.DateOut != nil {
		endDate = *i.DateOut
	}
	
	return int(endDate.Sub(*i.DateIn).Hours() / 24)
}

// GetDescription returns a human-readable description
func (i *InventoryItem) GetDescription() string {
	return fmt.Sprintf("%d joints of %s %s %s", i.Joints, i.Size, i.Weight, i.Grade)
}

// GetLocationInfo returns location and rack information
func (i *InventoryItem) GetLocationInfo() string {
	if i.Location != "" && i.Rack != "" {
		return fmt.Sprintf("%s - %s", i.Location, i.Rack)
	}
	if i.Location != "" {
		return i.Location
	}
	if i.Rack != "" {
		return i.Rack
	}
	return "No location assigned"
}

// HasConnectionType checks if item has specific connection type
func (i *InventoryItem) HasConnectionType(connectionType string) bool {
	return strings.EqualFold(i.Connection, connectionType)
}

// IsGrade checks if item matches specific grade
func (i *InventoryItem) IsGrade(grade string) bool {
	return strings.EqualFold(i.Grade, grade)
}

// IsSize checks if item matches specific size
func (i *InventoryItem) IsSize(size string) bool {
	return strings.EqualFold(i.Size, size)
}

// GetColorNumber returns the CN (Color Number) value
func (i *InventoryItem) GetColorNumber() int {
	return i.CN
}

// NeedsInspection returns true if item might need inspection
func (i *InventoryItem) NeedsInspection() bool {
	// Items without fletcher assignment might need inspection
	return i.Fletcher == "" && i.DateIn != nil && !i.IsShipped()
}

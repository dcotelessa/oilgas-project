// backend/pkg/validation/received_validation.go
package validation

import (
	"fmt"
	"strings"
	"time"

	"oilgas-backend/internal/models"
)

// ReceivedValidation validates received item creation and updates
type ReceivedValidation struct {
	// Required fields
	WorkOrder  string `json:"work_order" binding:"required"`
	CustomerID int    `json:"customer_id" binding:"required"`
	Joints     int    `json:"joints" binding:"required"`
	Size       string `json:"size" binding:"required"`
	Weight     string `json:"weight" binding:"required"`
	Grade      string `json:"grade" binding:"required"`
	Connection string `json:"connection" binding:"required"`

	// Optional basic fields
	CTD         bool   `json:"ctd,omitempty"`
	WString     bool   `json:"w_string,omitempty"`
	Rack        string `json:"rack,omitempty"`
	
	// Location fields
	Well        string `json:"well,omitempty"`
	Lease       string `json:"lease,omitempty"`
	
	// Business fields
	CustomerPO  string `json:"customer_po,omitempty"`
	OrderedBy   string `json:"ordered_by,omitempty"`
	EnteredBy   string `json:"entered_by,omitempty"`
	
	// Logistics
	Trucking    string `json:"trucking,omitempty"`
	Trailer     string `json:"trailer,omitempty"`
	
	// Notes and services
	Notes       string `json:"notes,omitempty"`
	Services    string `json:"services,omitempty"`
	Background  string `json:"background,omitempty"`
}

// Validate validates received item data
func (rv *ReceivedValidation) Validate() error {
	var errors []string

	// Validate required fields
	if rv.WorkOrder == "" {
		errors = append(errors, "work_order is required")
	} else if len(rv.WorkOrder) > 50 {
		errors = append(errors, "work_order must be 50 characters or less")
	}

	if rv.CustomerID <= 0 {
		errors = append(errors, "customer_id must be positive")
	}

	if rv.Joints <= 0 {
		errors = append(errors, "joints must be positive")
	} else if err := ValidateJointsCount(rv.Joints); err != nil {
		errors = append(errors, fmt.Sprintf("joints: %v", err))
	}

	// Validate oil & gas specifications
	if err := ValidateSize(rv.Size); err != nil {
		errors = append(errors, fmt.Sprintf("size: %v", err))
	}

	if err := ValidateWeight(rv.Weight); err != nil {
		errors = append(errors, fmt.Sprintf("weight: %v", err))
	}

	if err := ValidateGrade(rv.Grade); err != nil {
		errors = append(errors, fmt.Sprintf("grade: %v", err))
	}

	if err := ValidateConnection(rv.Connection); err != nil {
		errors = append(errors, fmt.Sprintf("connection: %v", err))
	}

	// Validate optional fields length
	if len(rv.CustomerPO) > 50 {
		errors = append(errors, "customer_po must be 50 characters or less")
	}

	if len(rv.OrderedBy) > 100 {
		errors = append(errors, "ordered_by must be 100 characters or less")
	}

	if len(rv.EnteredBy) > 100 {
		errors = append(errors, "entered_by must be 100 characters or less")
	}

	if len(rv.Well) > 100 {
		errors = append(errors, "well must be 100 characters or less")
	}

	if len(rv.Lease) > 100 {
		errors = append(errors, "lease must be 100 characters or less")
	}

	if len(rv.Rack) > 50 {
		errors = append(errors, "rack must be 50 characters or less")
	}

	if len(rv.Trucking) > 100 {
		errors = append(errors, "trucking must be 100 characters or less")
	}

	if len(rv.Trailer) > 50 {
		errors = append(errors, "trailer must be 50 characters or less")
	}

	if len(rv.Notes) > 1000 {
		errors = append(errors, "notes must be 1000 characters or less")
	}

	if len(rv.Services) > 500 {
		errors = append(errors, "services must be 500 characters or less")
	}

	if len(rv.Background) > 500 {
		errors = append(errors, "background must be 500 characters or less")
	}

	// Return combined errors
	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ToReceivedModel converts validation to received item model
func (rv *ReceivedValidation) ToReceivedModel() *models.ReceivedItem {
	now := time.Now()
	
	return &models.ReceivedItem{
		WorkOrder:    rv.WorkOrder,
		CustomerID:   rv.CustomerID,
		Joints:       rv.Joints,
		Size:         NormalizeSize(rv.Size),
		Weight:       rv.Weight,
		Grade:        NormalizeGrade(rv.Grade),
		Connection:   NormalizeConnection(rv.Connection),
		CTD:          rv.CTD,
		WString:      rv.WString,
		Rack:         rv.Rack,
		Well:         rv.Well,
		Lease:        rv.Lease,
		CustomerPO:   rv.CustomerPO,
		OrderedBy:    rv.OrderedBy,
		EnteredBy:    rv.EnteredBy,
		Trucking:     rv.Trucking,
		Trailer:      rv.Trailer,
		Notes:        rv.Notes,
		Services:     rv.Services,
		Background:   rv.Background,
		DateReceived: &now,
		CreatedAt:    now,
		Complete:     false,
		Deleted:      false,
	}
}

// FromReceivedModel converts received item model to validation
func FromReceivedModel(item *models.ReceivedItem) *ReceivedValidation {
	return &ReceivedValidation{
		WorkOrder:   item.WorkOrder,
		CustomerID:  item.CustomerID,
		Joints:      item.Joints,
		Size:        item.Size,
		Weight:      item.Weight,
		Grade:       item.Grade,
		Connection:  item.Connection,
		CTD:         item.CTD,
		WString:     item.WString,
		Rack:        item.Rack,
		Well:        item.Well,
		Lease:       item.Lease,
		CustomerPO:  item.CustomerPO,
		OrderedBy:   item.OrderedBy,
		EnteredBy:   item.EnteredBy,
		Trucking:    item.Trucking,
		Trailer:     item.Trailer,
		Notes:       item.Notes,
		Services:    item.Services,
		Background:  item.Background,
	}
}

// NormalizeReceivedData normalizes received item data for consistency
func (rv *ReceivedValidation) NormalizeReceivedData() {
	rv.WorkOrder = strings.TrimSpace(rv.WorkOrder)
	rv.Size = NormalizeSize(rv.Size)
	rv.Weight = strings.TrimSpace(rv.Weight)
	rv.Grade = NormalizeGrade(rv.Grade)
	rv.Connection = NormalizeConnection(rv.Connection)
	rv.Rack = strings.TrimSpace(rv.Rack)
	rv.Well = strings.TrimSpace(rv.Well)
	rv.Lease = strings.TrimSpace(rv.Lease)
	rv.CustomerPO = strings.TrimSpace(rv.CustomerPO)
	rv.OrderedBy = strings.TrimSpace(rv.OrderedBy)
	rv.EnteredBy = strings.TrimSpace(rv.EnteredBy)
	rv.Trucking = strings.TrimSpace(rv.Trucking)
	rv.Trailer = strings.TrimSpace(rv.Trailer)
	rv.Notes = strings.TrimSpace(rv.Notes)
	rv.Services = strings.TrimSpace(rv.Services)
	rv.Background = strings.TrimSpace(rv.Background)
}

// ValidateWorkOrderFormat validates work order format
func ValidateWorkOrderFormat(workOrder string) error {
	if workOrder == "" {
		return fmt.Errorf("work order is required")
	}

	// Basic format validation
	if len(workOrder) < 3 {
		return fmt.Errorf("work order must be at least 3 characters")
	}

	if len(workOrder) > 50 {
		return fmt.Errorf("work order must be 50 characters or less")
	}

	// Optional: Add specific format rules here
	// e.g., must start with letters, contain numbers, etc.

	return nil
}

// BatchReceivedValidation for bulk operations
type BatchReceivedValidation struct {
	Items []ReceivedValidation `json:"items"`
}

// Validate validates all items in batch
func (brv *BatchReceivedValidation) Validate() map[int]error {
	errors := make(map[int]error)

	for i, item := range brv.Items {
		if err := item.Validate(); err != nil {
			errors[i] = err
		}
	}

	return errors
}

// ToReceivedModels converts batch to slice of models
func (brv *BatchReceivedValidation) ToReceivedModels() []*models.ReceivedItem {
	var items []*models.ReceivedItem

	for _, validation := range brv.Items {
		items = append(items, validation.ToReceivedModel())
	}

	return items
}

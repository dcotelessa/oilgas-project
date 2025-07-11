// backend/internal/models/received_go
package models

import (
	"fmt"
	"time"
)

// ReceivedItem represents items received into inventory
type ReceivedItem struct {
	ID                int        `json:"id" db:"id"`
	WorkOrder         string     `json:"work_order" db:"work_order"`
	CustomerID        int        `json:"customer_id" db:"customer_id"`
	Customer          string     `json:"customer" db:"customer"`
	Joints            int        `json:"joints" db:"joints"`
	Rack              string     `json:"rack" db:"rack"`
	SizeID            int        `json:"size_id" db:"size_id"`
	Size              string     `json:"size" db:"size"`
	Weight            string     `json:"weight" db:"weight"`
	Grade             string     `json:"grade" db:"grade"`
	Connection        string     `json:"connection" db:"connection"`
	CTD               bool       `json:"ctd" db:"ctd"`
	WString           bool       `json:"w_string" db:"w_string"`
	Well              string     `json:"well" db:"well"`
	Lease             string     `json:"lease" db:"lease"`
	OrderedBy         string     `json:"ordered_by" db:"ordered_by"`
	Notes             string     `json:"notes" db:"notes"`
	CustomerPO        string     `json:"customer_po" db:"customer_po"`
	DateReceived      *time.Time `json:"date_received" db:"date_received"`
	Background        string     `json:"background" db:"background"`
	Norm              string     `json:"norm" db:"norm"`
	Services          string     `json:"services" db:"services"`
	BillToID          string     `json:"bill_to_id" db:"bill_to_id"`
	EnteredBy         string     `json:"entered_by" db:"entered_by"`
	WhenEntered       *time.Time `json:"when_entered" db:"when_entered"`
	Trucking          string     `json:"trucking" db:"trucking"`
	Trailer           string     `json:"trailer" db:"trailer"`
	InProduction      *time.Time `json:"in_production" db:"in_production"`
	InspectedDate     *time.Time `json:"inspected_date" db:"inspected_date"`
	ThreadingDate     *time.Time `json:"threading_date" db:"threading_date"`
	StraightenRequired bool      `json:"straighten_required" db:"straighten_required"`
	ExcessMaterial    bool       `json:"excess_material" db:"excess_material"`
	Complete          bool       `json:"complete" db:"complete"`
	InspectedBy       string     `json:"inspected_by" db:"inspected_by"`
	UpdatedBy         string     `json:"updated_by" db:"updated_by"`
	WhenUpdated       *time.Time `json:"when_updated" db:"when_updated"`
	Deleted           bool       `json:"deleted" db:"deleted"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// GetCurrentState returns the current workflow state
func (r *ReceivedItem) GetCurrentState() WorkflowState {
	if r.Complete {
		return StateCompleted
	}
	if r.InspectedDate != nil {
		return StateInspection
	}
	if r.InProduction != nil {
		return StateProduction
	}
	return StateReceived
}

// GetDaysInCurrentState returns how many days in current state
func (r *ReceivedItem) GetDaysInCurrentState() int {
	var stateStartTime *time.Time
	
	switch r.GetCurrentState() {
	case StateReceived:
		stateStartTime = r.DateReceived
	case StateProduction:
		stateStartTime = r.InProduction
	case StateInspection:
		stateStartTime = r.InspectedDate
	case StateCompleted:
		if r.WhenUpdated != nil {
			stateStartTime = r.WhenUpdated
		} else {
			stateStartTime = r.InspectedDate
		}
	}
	
	if stateStartTime == nil {
		return 0
	}
	
	return int(time.Since(*stateStartTime).Hours() / 24)
}

// IsReadyForProduction returns true if item can be moved to production
func (r *ReceivedItem) IsReadyForProduction() bool {
	return r.GetCurrentState() == StateReceived && !r.Deleted
}

// IsReadyForInspection returns true if item can be inspected
func (r *ReceivedItem) IsReadyForInspection() bool {
	return r.GetCurrentState() == StateProduction && !r.Deleted
}

// IsOverdue returns true if item has been in current state too long
func (r *ReceivedItem) IsOverdue() bool {
	days := r.GetDaysInCurrentState()
	
	switch r.GetCurrentState() {
	case StateReceived:
		return days > 3 // Should move to production within 3 days
	case StateProduction:
		return days > 7 // Should complete production within 7 days
	case StateInspection:
		return days > 2 // Should complete inspection within 2 days
	}
	
	return false
}

// GetDescription returns a human-readable description
func (r *ReceivedItem) GetDescription() string {
	return fmt.Sprintf("WO-%s: %d joints of %s %s %s", 
		r.WorkOrder, r.Joints, r.Size, r.Weight, r.Grade)
}

// GetWellInfo returns well and lease information
func (r *ReceivedItem) GetWellInfo() string {
	if r.Well != "" && r.Lease != "" {
		return fmt.Sprintf("%s/%s", r.Well, r.Lease)
	}
	if r.Well != "" {
		return r.Well
	}
	if r.Lease != "" {
		return r.Lease
	}
	return ""
}

// RequiresSpecialHandling returns true if item needs special processing
func (r *ReceivedItem) RequiresSpecialHandling() bool {
	return r.StraightenRequired || r.ExcessMaterial
}

// GetProcessingNotes returns notes about special processing requirements
func (r *ReceivedItem) GetProcessingNotes() []string {
	var notes []string
	
	if r.StraightenRequired {
		notes = append(notes, "Straightening required")
	}
	if r.ExcessMaterial {
		notes = append(notes, "Excess material present")
	}
	if r.CTD {
		notes = append(notes, "CTD (Cut to Depth)")
	}
	if r.WString {
		notes = append(notes, "Work String")
	}
	
	return notes
}

// CanAdvanceToNextState returns true if item can move to next workflow state
func (r *ReceivedItem) CanAdvanceToNextState() bool {
	if r.Deleted || r.Complete {
		return false
	}
	
	switch r.GetCurrentState() {
	case StateReceived:
		return true // Can always move to production
	case StateProduction:
		return true // Can move to inspection
	case StateInspection:
		return true // Can mark as complete
	}
	
	return false
}


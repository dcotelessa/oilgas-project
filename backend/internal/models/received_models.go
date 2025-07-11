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
	DateOut           *time.Time `json:"date_out" db:"date_out"`
	Deleted           bool       `json:"deleted" db:"deleted"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// GetCurrentState returns the current workflow state
func (r *ReceivedItem) GetCurrentState() string {
	if r.Complete {
		return StateCompleted
	}
	if r.DateOut != nil {
		return StateShipped
	}
	if r.InspectedDate != nil {
		return StateInspection
	}
	if r.InProduction != nil {
		return StateProduction
	}
	if r.DateReceived != nil {
		return StateReceived
	}
	return "unknown"
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
	case StateInventory:
		// TODO: Still need inventory table lookup or additional field
		stateStartTime = r.InspectedDate
	case StateShipped:
		stateStartTime = r.DateOut
	case StateCompleted:
		if r.WhenUpdated != nil {
			stateStartTime = r.WhenUpdated
		} else {
			stateStartTime = r.DateOut
		}
	}
	
	if stateStartTime == nil {
		return 0
	}
	
	return int(time.Since(*stateStartTime).Hours() / 24)
}

// GetWorkflowStatus creates a WorkflowStatus object for this received item
func (r *ReceivedItem) GetWorkflowStatus() *WorkflowStatus {
	return &WorkflowStatus{
		WorkOrder:     r.WorkOrder,
		Customer:      r.Customer,
		Joints:        r.Joints,
		Grade:         r.Grade,
		Size:          r.Size,
		CurrentState:  r.GetCurrentState(),
		DateReceived:  r.DateReceived,
		InProduction:  r.InProduction,
		InspectedDate: r.InspectedDate,
		Complete:      r.Complete,
		DaysInState:   r.GetDaysInCurrentState(),
	}
}

// CanTransitionTo checks if item can transition to target state
func (r *ReceivedItem) CanTransitionTo(targetState string) bool {
	if r.Deleted {
		return false
	}
	
	// Use WorkflowStatus validation logic
	status := r.GetWorkflowStatus()
	return status.IsValidTransition(targetState)
}

// GetValidTransitions returns possible next states for this item
func (r *ReceivedItem) GetValidTransitions() []string {
	if r.Deleted {
		return []string{}
	}
	
	status := r.GetWorkflowStatus()
	return status.GetNextStates()
}

// IsReadyForProduction returns true if item can be moved to production
func (r *ReceivedItem) IsReadyForProduction() bool {
	return r.CanTransitionTo(StateProduction) && !r.Deleted
}

// IsReadyForInspection returns true if item can be inspected
func (r *ReceivedItem) IsReadyForInspection() bool {
	return r.CanTransitionTo(StateInspection) && !r.Deleted
}

// IsReadyForInventory returns true if item can be moved to inventory
func (r *ReceivedItem) IsReadyForInventory() bool {
	return r.CanTransitionTo(StateInventory) && !r.Deleted
}

// IsReadyForShipping returns true if item can be shipped
func (r *ReceivedItem) IsReadyForShipping() bool {
	return r.CanTransitionTo(StateShipped) && !r.Deleted
}

// IsReadyForCompletion returns true if item can be marked complete
func (r *ReceivedItem) IsReadyForCompletion() bool {
	return r.CanTransitionTo(StateCompleted) && !r.Deleted
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
	case StateInventory:
		return days > 14 // Should ship within 14 days
	case StateShipped:
		return days > 5 // Should complete within 5 days of shipping
	case StateCompleted:
		return false // Completed items are never overdue
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
	
	// Use WorkflowStatus logic instead of hard-coded rules
	status := r.GetWorkflowStatus()
	nextStates := status.GetNextStates()
	
	return len(nextStates) > 0
}

// GetNextAvailableStates returns all possible next states
func (r *ReceivedItem) GetNextAvailableStates() []string {
	if r.Deleted {
		return []string{}
	}
	
	status := r.GetWorkflowStatus()
	return status.GetNextStates()
}

// ValidateForTransition validates if item is ready for specific transition
func (r *ReceivedItem) ValidateForTransition(targetState string) error {
	if r.Deleted {
		return fmt.Errorf("cannot transition deleted item")
	}
	
	if !r.CanTransitionTo(targetState) {
		currentState := r.GetCurrentState()
		return fmt.Errorf("invalid transition from %s to %s", currentState, targetState)
	}
	
	// Business rule validations
	switch targetState {
	case StateProduction:
		if r.Joints <= 0 {
			return fmt.Errorf("cannot move to production: no joints specified")
		}
		if r.Size == "" {
			return fmt.Errorf("cannot move to production: pipe size not specified")
		}
		if r.Grade == "" {
			return fmt.Errorf("cannot move to production: pipe grade not specified")
		}
	case StateInspection:
		if r.InProduction == nil {
			return fmt.Errorf("cannot move to inspection: item not in production")
		}
	case StateInventory:
		if r.InspectedDate == nil {
			return fmt.Errorf("cannot move to inventory: item not inspected")
		}
	case StateShipped:
		// Would need to check inventory status
		// For now, just check inspection is complete
		if r.InspectedDate == nil {
			return fmt.Errorf("cannot ship: item not inspected")
		}
	}
	
	return nil
}

// GetWorkflowSummary returns a summary of the item's workflow progress
func (r *ReceivedItem) GetWorkflowSummary() map[string]interface{} {
	return map[string]interface{}{
		"current_state":        r.GetCurrentState(),
		"days_in_state":       r.GetDaysInCurrentState(),
		"is_overdue":          r.IsOverdue(),
		"can_advance":         r.CanAdvanceToNextState(),
		"next_states":         r.GetNextAvailableStates(),
		"requires_special":    r.RequiresSpecialHandling(),
		"processing_notes":    r.GetProcessingNotes(),
	}
}

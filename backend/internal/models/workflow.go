package models

import (
	"fmt"
	"time"
)

// Workflow state constants as strings (matches workflow_status.go)
const (
	StateReceived    = "received"
	StateProduction  = "in_production" 
	StateInspection  = "inspection"
	StateInventory   = "inventory"
	StateShipped     = "shipped"
	StateCompleted   = "completed"
)

// ColorNumber represents the CN (Color Number) classification
type ColorNumber int

const (
	CNPremium   ColorNumber = 1 // WHT - White (Premium Quality)
	CNStandard  ColorNumber = 2 // BLU - Blue (Standard Quality)
	CNEconomy   ColorNumber = 3 // GRN - Green (Economy Quality)
	CNRejected  ColorNumber = 4 // RED - Red (Rejected/Problem)
	CNGrade5    ColorNumber = 5 // Grade 5 Quality
	CNGrade6    ColorNumber = 6 // Grade 6 Quality
)

// Job represents a complete work order with workflow state
type Job struct {
	ID           int                    `json:"id" db:"id"`
	WorkOrder    string                 `json:"work_order" db:"wkorder"`
	RNumber      *int                   `json:"r_number,omitempty" db:"rnumber"`
	CustomerID   int                    `json:"customer_id" db:"custid"`
	Customer     string                 `json:"customer" db:"customer"`
	CustomerPO   string                 `json:"customer_po,omitempty" db:"customerpo"`
	
	// Pipe specifications
	Size         string                 `json:"size" db:"size"`
	Weight       string                 `json:"weight" db:"weight"`
	Grade        string                 `json:"grade" db:"grade"`
	Connection   string                 `json:"connection" db:"connection"`
	Joints       int                    `json:"joints" db:"joints"`
	CTD          bool                   `json:"ctd" db:"ctd"`
	WString      bool                   `json:"wstring" db:"wstring"`
	SWGCC        string                 `json:"swgcc" db:"swgcc"`
	
	// Location and logistics
	Well         string                 `json:"well,omitempty" db:"well"`
	Lease        string                 `json:"lease,omitempty" db:"lease"`
	Rack         string                 `json:"rack,omitempty" db:"rack"`
	Location     string                 `json:"location,omitempty" db:"location"`
	Trucking     string                 `json:"trucking,omitempty" db:"trucking"`
	Trailer      string                 `json:"trailer,omitempty" db:"trailer"`
	
	// Workflow dates - used to determine state
	DateReceived *time.Time             `json:"date_received,omitempty" db:"daterecvd"`
	InProduction *time.Time             `json:"in_production,omitempty" db:"inproduction"`
	Inspected    *time.Time             `json:"inspected,omitempty" db:"inspected"`
	DateIn       *time.Time             `json:"date_in,omitempty" db:"datein"`
	DateOut      *time.Time             `json:"date_out,omitempty" db:"dateout"`
	
	// Workflow flags
	Complete     bool                   `json:"complete" db:"complete"`
	Deleted      bool                   `json:"deleted" db:"deleted"`
	
	// People
	OrderedBy    string                 `json:"ordered_by,omitempty" db:"orderedby"`
	EnteredBy    string                 `json:"entered_by,omitempty" db:"enteredby"`
	InspectedBy  string                 `json:"inspected_by,omitempty" db:"inspectedby"`
	
	// Notes and services
	Notes        string                 `json:"notes,omitempty" db:"notes"`
	Services     string                 `json:"services,omitempty" db:"services"`
	Background   string                 `json:"background,omitempty" db:"background"`
	
	// Computed fields
	CurrentState string		    `json:"current_state"`
	ColorDetails map[ColorNumber]int    `json:"color_details,omitempty"` // CN -> joints count
	
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time             `json:"updated_at,omitempty" db:"updated_at"`
}

// PipeSize represents size/weight/connection combinations
type PipeSize struct {
	ID         int    `json:"id" db:"sizeid"`
	CustomerID int    `json:"customer_id" db:"custid"`
	Size       string `json:"size" db:"size"`
	Weight     string `json:"weight" db:"weight"`
	Connection string `json:"connection" db:"connection"`
}

// CustomerStats for dashboard
type CustomerStats struct {
	CustomerID   int    `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	ActiveJobs   int    `json:"active_jobs"`
	TotalJoints  int    `json:"total_joints"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// GetCurrentState determines workflow state based on dates and flags
func (j *Job) GetCurrentState() string {
	// Completed/Shipped (joints < 0 indicates shipped)
	if j.DateOut != nil && j.Joints < 0 {
		return StateCompleted
	}
	
	// Inventory (has inventory date and positive joints)
	if j.DateIn != nil && j.Joints > 0 {
		return StateInventory
	}
	
	// Inspection (inspected but not yet in inventory)
	if j.Inspected != nil && j.Complete {
		return StateInspection
	}
	
	// Production (in production but not inspected)
	if j.InProduction != nil {
		return StateProduction
	}
	
	// Default to receiving
	return StateReceived
}

// IsActive returns true if job is not completed/shipped
func (j *Job) IsActive() bool {
	return !j.Deleted && j.GetCurrentState() != StateCompleted
}

// CanTransitionTo checks if job can transition to target state using workflow_status.go logic
func (j *Job) CanTransitionTo(targetState string) bool {
	if j.Deleted {
		return false
	}
	
	// Use the validation from workflow_status.go (we'll need to import or duplicate the logic)
	// For now, use basic validation - this should use WorkflowStatus.IsValidTransition
	currentState := j.GetCurrentState()
	
	validTransitions := map[string][]string{
		StateReceived:    {StateProduction},
		StateProduction:  {StateInspection},
		StateInspection:  {StateInventory, StateCompleted},
		StateInventory:   {StateShipped, StateCompleted},
		StateShipped:     {StateCompleted},
		StateCompleted:   {}, // Terminal state
	}
	
	allowedStates, exists := validTransitions[currentState]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedStates {
		if allowed == targetState {
			return true
		}
	}
	
	return false
}

// GetValidTransitions returns the valid transitions from current state
func (j *Job) GetValidTransitions() []string {
	if j.Deleted {
		return []string{}
	}
	
	currentState := j.GetCurrentState()
	
	transitions := map[string][]string{
		StateReceived:    {StateProduction},
		StateProduction:  {StateInspection},
		StateInspection:  {StateInventory, StateCompleted},
		StateInventory:   {StateShipped, StateCompleted},
		StateShipped:     {StateCompleted},
		StateCompleted:   {}, // Terminal state
	}
	
	if validStates, exists := transitions[currentState]; exists {
		return validStates
	}
	
	return []string{}
}

// GetColorName returns the color name for a CN
func (cn ColorNumber) GetColorName() string {
	switch cn {
	case CNPremium:
		return "WHT"
	case CNStandard:
		return "BLU"
	case CNEconomy:
		return "GRN"
	case CNRejected:
		return "RED"
	case CNGrade5:
		return "Grade 5"
	case CNGrade6:
		return "Grade 6"
	default:
		return "Unknown"
	}
}

// GetQualityDescription returns quality description for CN
func (cn ColorNumber) GetQualityDescription() string {
	switch cn {
	case CNPremium:
		return "Premium Quality"
	case CNStandard:
		return "Standard Quality"
	case CNEconomy:
		return "Economy Quality"
	case CNRejected:
		return "Rejected/Problem"
	case CNGrade5:
		return "Grade 5 Quality"
	case CNGrade6:
		return "Grade 6 Quality"
	default:
		return "Unknown Quality"
	}
}

// GetAllWorkflowStates returns all valid workflow states
func GetAllWorkflowStates() []string {
	return []string{
		StateReceived,
		StateProduction,
		StateInspection,
		StateInventory,
		StateShipped,
		StateCompleted,
	}
}

// IsValidWorkflowState checks if a state string is valid
func IsValidWorkflowState(state string) bool {
	validStates := GetAllWorkflowStates()
	for _, validState := range validStates {
		if validState == state {
			return true
		}
	}
	return false
}

// ValidateWorkflowState returns error if state is invalid
func ValidateWorkflowState(state string) error {
	if !IsValidWorkflowState(state) {
		return fmt.Errorf("invalid workflow state: %s. Valid states: %v", state, GetAllWorkflowStates())
	}
	return nil
}

// GetWorkflowStateName returns display name for state
func GetWorkflowStateName(state string) string {
	switch state {
	case StateReceived:
		return "Received"
	case StateProduction:
		return "In Production"
	case StateInspection:
		return "Inspection"
	case StateInventory:
		return "Inventory"
	case StateShipped:
		return "Shipped"
	case StateCompleted:
		return "Completed"
	default:
		return "Unknown"
	}
}

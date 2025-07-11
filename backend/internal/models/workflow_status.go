// backend/internal/models/workflow_status.go
package models

import "time"

// workflowTransitions defines all valid state transitions
// Single source of truth - no duplication
var workflowTransitions = map[string][]string{
	StateReceived:    {StateProduction},
	StateProduction:  {StateInspection},
	StateInspection:  {StateInventory, StateCompleted},
	StateInventory:   {StateShipped, StateCompleted},
	StateShipped:     {StateCompleted},
	StateCompleted:   {}, // Terminal state
}

// WorkflowStatus represents the current status of a work order in the workflow
type WorkflowStatus struct {
	WorkOrder     string     `json:"work_order"`
	Customer      string     `json:"customer"`
	Joints        int        `json:"joints"`
	Grade         string     `json:"grade"`
	Size          string     `json:"size"`
	CurrentState  string     `json:"current_state"`
	DaysInState   int        `json:"days_in_state"`
	
	// Workflow timestamps
	DateReceived  *time.Time `json:"date_received,omitempty"`
	InProduction  *time.Time `json:"in_production,omitempty"`
	InspectedDate *time.Time `json:"inspected_date,omitempty"`
	InventoryDate *time.Time `json:"inventory_date,omitempty"`
	Complete      bool       `json:"complete"`
}

// WorkflowStateTransition represents a state change request
type WorkflowStateTransition struct {
	WorkOrder    string `json:"work_order" validate:"required"`
	FromState    string `json:"from_state"`
	ToState      string `json:"to_state" validate:"required"`
	Username     string `json:"username" validate:"required"`
	Notes        string `json:"notes,omitempty"`
	TransitionAt time.Time `json:"transition_at"`
}

// StateChange represents a workflow state transition  
type StateChange struct {
	FromState string	`json:"from_state"`
	ToState   string	`json:"to_state"`
	ChangedBy string        `json:"changed_by"`
	ChangedAt time.Time     `json:"changed_at"`
	Notes     string        `json:"notes,omitempty"`
	Duration  string        `json:"duration,omitempty"` // Time spent in FromState
}

// IsValidTransition checks if a state transition is allowed
func (w *WorkflowStatus) IsValidTransition(toState string) bool {
	allowedStates, exists := workflowTransitions[w.CurrentState]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedStates {
		if allowed == toState {
			return true
		}
	}
	
	return false
}

// GetNextStates returns the possible next states from current state
func (w *WorkflowStatus) GetNextStates() []string {
	if states, exists := workflowTransitions[w.CurrentState]; exists {
		return states
	}
	return []string{}
}

// GetAllValidTransitions returns the complete transition map (for debugging/docs)
func GetAllValidTransitions() map[string][]string {
	// Return copy to prevent external modification
	result := make(map[string][]string)
	for state, transitions := range workflowTransitions {
		result[state] = append([]string{}, transitions...)
	}
	return result
}

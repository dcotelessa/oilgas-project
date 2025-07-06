// backend/internal/services/workflow_state_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
)

type WorkflowStateService interface {
	// State transition methods
	AdvanceToProduction(ctx context.Context, workOrder string, username string) error
	AdvanceToInspection(ctx context.Context, workOrder string, inspectedBy string) error
	AdvanceToInventory(ctx context.Context, workOrder string) error
	MarkAsComplete(ctx context.Context, workOrder string) error

	// Status and history methods
	GetCurrentState(ctx context.Context, workOrder string) (*models.WorkflowState, error)
	GetWorkflowStatus(ctx context.Context, workOrder string) (*models.WorkflowStatus, error)
	GetStateHistory(ctx context.Context, workOrder string) ([]models.StateChange, error)
	
	// Bulk operations
	GetItemsByState(ctx context.Context, state models.WorkflowState) ([]string, error)
	GetJobsByState(ctx context.Context, state string, limit, offset int) ([]models.ReceivedItem, int, error)
	
	// Validation and business logic
	ValidateTransition(ctx context.Context, workOrder string, targetState models.WorkflowState) error
	CanAdvanceToNextState(ctx context.Context, workOrder string) (bool, []models.WorkflowState, error)
	
	// Analytics and monitoring
	GetWorkflowMetrics(ctx context.Context) (*WorkflowMetrics, error)
	GetBottlenecks(ctx context.Context) ([]WorkflowBottleneck, error)
	GetOverdueItems(ctx context.Context, state models.WorkflowState) ([]models.ReceivedItem, error)
}

type workflowStateService struct {
	workflowRepo repository.WorkflowStateRepository
	receivedRepo repository.ReceivedRepository
	cache        *cache.Cache
}

func NewWorkflowStateService(
	workflowRepo repository.WorkflowStateRepository,
	receivedRepo repository.ReceivedRepository,
	cache *cache.Cache,
) WorkflowStateService {
	return &workflowStateService{
		workflowRepo: workflowRepo,
		receivedRepo: receivedRepo,
		cache:        cache,
	}
}

// State transition methods

func (s *workflowStateService) AdvanceToProduction(ctx context.Context, workOrder string, username string) error {
	// Validate transition
	if err := s.ValidateTransition(ctx, workOrder, models.StateProduction); err != nil {
		return fmt.Errorf("cannot advance to production: %w", err)
	}

	// Execute transition
	if err := s.workflowRepo.AdvanceToProduction(ctx, workOrder, username); err != nil {
		return fmt.Errorf("failed to advance to production: %w", err)
	}

	// Invalidate caches
	s.invalidateWorkflowCaches(workOrder)

	return nil
}

func (s *workflowStateService) AdvanceToInspection(ctx context.Context, workOrder string, inspectedBy string) error {
	// Validate transition
	if err := s.ValidateTransition(ctx, workOrder, models.StateInspection); err != nil {
		return fmt.Errorf("cannot advance to inspection: %w", err)
	}

	// Execute transition
	if err := s.workflowRepo.AdvanceToInspection(ctx, workOrder, inspectedBy); err != nil {
		return fmt.Errorf("failed to advance to inspection: %w", err)
	}

	// Invalidate caches
	s.invalidateWorkflowCaches(workOrder)

	return nil
}

func (s *workflowStateService) AdvanceToInventory(ctx context.Context, workOrder string) error {
	// Validate transition
	if err := s.ValidateTransition(ctx, workOrder, models.StateInventory); err != nil {
		return fmt.Errorf("cannot advance to inventory: %w", err)
	}

	// Execute transition
	if err := s.workflowRepo.AdvanceToInventory(ctx, workOrder); err != nil {
		return fmt.Errorf("failed to advance to inventory: %w", err)
	}

	// Invalidate caches
	s.invalidateWorkflowCaches(workOrder)

	return nil
}

func (s *workflowStateService) MarkAsComplete(ctx context.Context, workOrder string) error {
	// Validate transition
	if err := s.ValidateTransition(ctx, workOrder, models.StateComplete); err != nil {
		return fmt.Errorf("cannot mark as complete: %w", err)
	}

	// Execute transition
	if err := s.workflowRepo.MarkAsComplete(ctx, workOrder); err != nil {
		return fmt.Errorf("failed to mark as complete: %w", err)
	}

	// Invalidate caches
	s.invalidateWorkflowCaches(workOrder)

	return nil
}

// Status and history methods

func (s *workflowStateService) GetCurrentState(ctx context.Context, workOrder string) (*models.WorkflowState, error) {
	cacheKey := fmt.Sprintf("workflow:state:%s", workOrder)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if state, ok := cached.(*models.WorkflowState); ok {
			return state, nil
		}
	}

	// Get status from repository
	status, err := s.workflowRepo.GetWorkflowStatus(ctx, workOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	// Cache for 2 minutes
	s.cache.SetWithTTL(cacheKey, &status.CurrentState, 2*time.Minute)

	return &status.CurrentState, nil
}

func (s *workflowStateService) GetWorkflowStatus(ctx context.Context, workOrder string) (*models.WorkflowStatus, error) {
	cacheKey := fmt.Sprintf("workflow:status:%s", workOrder)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if status, ok := cached.(*models.WorkflowStatus); ok {
			return status, nil
		}
	}

	// Get from repository
	status, err := s.workflowRepo.GetWorkflowStatus(ctx, workOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	// Enhance with next possible states
	canAdvance, nextStates, err := s.CanAdvanceToNextState(ctx, workOrder)
	if err != nil {
		// Log error but don't fail the request
		canAdvance = false
		nextStates = []models.WorkflowState{}
	}

	status.CanAdvance = canAdvance
	status.NextStates = nextStates

	// Cache for 1 minute
	s.cache.SetWithTTL(cacheKey, status, time.Minute)

	return status, nil
}

func (s *workflowStateService) GetStateHistory(ctx context.Context, workOrder string) ([]models.StateChange, error) {
	cacheKey := fmt.Sprintf("workflow:history:%s", workOrder)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if history, ok := cached.([]models.StateChange); ok {
			return history, nil
		}
	}

	// Get status which includes history
	status, err := s.workflowRepo.GetWorkflowStatus(ctx, workOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	// Cache for 5 minutes (history doesn't change often)
	s.cache.SetWithTTL(cacheKey, status.StateHistory, 5*time.Minute)

	return status.StateHistory, nil
}

// Bulk operations

func (s *workflowStateService) GetItemsByState(ctx context.Context, state models.WorkflowState) ([]string, error) {
	cacheKey := fmt.Sprintf("workflow:items_by_state:%s", state)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if items, ok := cached.([]string); ok {
			return items, nil
		}
	}

	// Get items in this state (using jobs by state with large limit)
	jobs, _, err := s.workflowRepo.GetJobsByState(ctx, string(state), 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by state: %w", err)
	}

	// Extract work orders
	workOrders := make([]string, len(jobs))
	for i, job := range jobs {
		workOrders[i] = job.WorkOrder
	}

	// Cache for 3 minutes
	s.cache.SetWithTTL(cacheKey, workOrders, 3*time.Minute)

	return workOrders, nil
}

func (s *workflowStateService) GetJobsByState(ctx context.Context, state string, limit, offset int) ([]models.ReceivedItem, int, error) {
	cacheKey := fmt.Sprintf("workflow:jobs_by_state:%s:%d:%d", state, limit, offset)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(CachedJobsResult); ok {
			return result.Jobs, result.Total, nil
		}
	}

	// Get from repository
	jobs, total, err := s.workflowRepo.GetJobsByState(ctx, state, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get jobs by state: %w", err)
	}

	// Set current state for each job
	for i := range jobs {
		jobs[i].CurrentState = jobs[i].GetCurrentState()
	}

	// Cache result
	result := CachedJobsResult{Jobs: jobs, Total: total}
	s.cache.SetWithTTL(cacheKey, result, 2*time.Minute)

	return jobs, total, nil
}

// Validation and business logic

func (s *workflowStateService) ValidateTransition(ctx context.Context, workOrder string, targetState models.WorkflowState) error {
	// Get current state
	currentState, err := s.GetCurrentState(ctx, workOrder)
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	// Check if transition is valid
	if !currentState.CanAdvanceTo(targetState) {
		return fmt.Errorf("invalid transition from %s to %s", *currentState, targetState)
	}

	// Additional business logic validations
	switch targetState {
	case models.StateProduction:
		// Must have all required fields for production
		if err := s.validateForProduction(ctx, workOrder); err != nil {
			return fmt.Errorf("production validation failed: %w", err)
		}
	case models.StateInspection:
		// Must have completed production
		if *currentState != models.StateProduction {
			return fmt.Errorf("must complete production before inspection")
		}
	case models.StateInventory:
		// Must have passed inspection
		if *currentState != models.StateInspection {
			return fmt.Errorf("must complete inspection before moving to inventory")
		}
	case models.StateComplete:
		// Must be in inventory
		if *currentState != models.StateInventory {
			return fmt.Errorf("must be in inventory before marking complete")
		}
	}

	return nil
}

func (s *workflowStateService) CanAdvanceToNextState(ctx context.Context, workOrder string) (bool, []models.WorkflowState, error) {
	// Get current state
	currentState, err := s.GetCurrentState(ctx, workOrder)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get current state: %w", err)
	}

	// Define valid transitions
	validTransitions := map[models.WorkflowState][]models.WorkflowState{
		models.StateReceived:   {models.StateProduction},
		models.StateProduction: {models.StateInspection},
		models.StateInspection: {models.StateInventory},
		models.StateInventory:  {models.StateComplete},
		models.StateComplete:   {}, // Terminal state
	}

	nextStates, exists := validTransitions[*currentState]
	if !exists {
		return false, nil, nil
	}

	canAdvance := len(nextStates) > 0

	// Additional checks for specific transitions
	if canAdvance && len(nextStates) > 0 {
		// Check if any validations would prevent advancement
		for _, nextState := range nextStates {
			if err := s.ValidateTransition(ctx, workOrder, nextState); err != nil {
				// If validation fails, can't advance
				canAdvance = false
				break
			}
		}
	}

	return canAdvance, nextStates, nil
}

// Analytics and monitoring

func (s *workflowStateService) GetWorkflowMetrics(ctx context.Context) (*WorkflowMetrics, error) {
	cacheKey := "workflow:metrics"
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if metrics, ok := cached.(*WorkflowMetrics); ok {
			return metrics, nil
		}
	}

	// Build metrics from state counts
	metrics := &WorkflowMetrics{
		StateDistribution: make(map[models.WorkflowState]int),
		LastUpdated:       time.Now(),
	}

	// Count items in each state
	states := []models.WorkflowState{
		models.StateReceived,
		models.StateProduction,
		models.StateInspection,
		models.StateInventory,
		models.StateComplete,
	}

	totalItems := 0
	for _, state := range states {
		items, err := s.GetItemsByState(ctx, state)
		if err != nil {
			// Log error but continue
			metrics.StateDistribution[state] = 0
		} else {
			count := len(items)
			metrics.StateDistribution[state] = count
			totalItems += count
		}
	}

	metrics.TotalItems = totalItems

	// Calculate average processing time (simplified)
	if totalItems > 0 {
		metrics.AverageProcessingDays = s.calculateAverageProcessingTime(ctx)
	}

	// Cache for 10 minutes
	s.cache.SetWithTTL(cacheKey, metrics, 10*time.Minute)

	return metrics, nil
}

func (s *workflowStateService) GetBottlenecks(ctx context.Context) ([]WorkflowBottleneck, error) {
	cacheKey := "workflow:bottlenecks"
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if bottlenecks, ok := cached.([]WorkflowBottleneck); ok {
			return bottlenecks, nil
		}
	}

	var bottlenecks []WorkflowBottleneck

	// Check each state for potential bottlenecks
	states := []models.WorkflowState{
		models.StateReceived,
		models.StateProduction,
		models.StateInspection,
	}

	for _, state := range states {
		items, err := s.GetItemsByState(ctx, state)
		if err != nil {
			continue
		}

		// Consider it a bottleneck if too many items in one state
		if len(items) > 20 { // Configurable threshold
			bottlenecks = append(bottlenecks, WorkflowBottleneck{
				State:     state,
				ItemCount: len(items),
				Severity:  s.calculateBottleneckSeverity(len(items)),
				Message:   fmt.Sprintf("%d items stuck in %s state", len(items), state),
			})
		}
	}

	// Cache for 5 minutes
	s.cache.SetWithTTL(cacheKey, bottlenecks, 5*time.Minute)

	return bottlenecks, nil
}

func (s *workflowStateService) GetOverdueItems(ctx context.Context, state models.WorkflowState) ([]models.ReceivedItem, error) {
	// Get all items in the state
	jobs, _, err := s.GetJobsByState(ctx, string(state), 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs in state %s: %w", state, err)
	}

	// Filter overdue items
	var overdueItems []models.ReceivedItem
	for _, job := range jobs {
		if job.IsOverdue() {
			overdueItems = append(overdueItems, job)
		}
	}

	return overdueItems, nil
}

// Helper methods

func (s *workflowStateService) validateForProduction(ctx context.Context, workOrder string) error {
	// Get the received item details
	// This would typically check that all required fields are present
	// For now, just return nil (implement based on your business rules)
	return nil
}

func (s *workflowStateService) calculateAverageProcessingTime(ctx context.Context) float64 {
	// Simplified calculation - in practice, you'd query historical data
	return 7.5 // Average days
}

func (s *workflowStateService) calculateBottleneckSeverity(itemCount int) string {
	if itemCount > 50 {
		return "critical"
	} else if itemCount > 30 {
		return "high"
	} else if itemCount > 20 {
		return "medium"
	}
	return "low"
}

func (s *workflowStateService) invalidateWorkflowCaches(workOrder string) {
	// Clear specific workflow caches
	s.cache.Delete(fmt.Sprintf("workflow:state:%s", workOrder))
	s.cache.Delete(fmt.Sprintf("workflow:status:%s", workOrder))
	s.cache.Delete(fmt.Sprintf("workflow:history:%s", workOrder))
	
	// Clear general workflow caches
	s.cache.DeletePattern("workflow:items_by_state:")
	s.cache.DeletePattern("workflow:jobs_by_state:")
	s.cache.Delete("workflow:metrics")
	s.cache.Delete("workflow:bottlenecks")
}

// Helper types

type WorkflowMetrics struct {
	TotalItems              int                                `json:"total_items"`
	StateDistribution       map[models.WorkflowState]int       `json:"state_distribution"`
	AverageProcessingDays   float64                            `json:"average_processing_days"`
	LastUpdated             time.Time                          `json:"last_updated"`
}

type WorkflowBottleneck struct {
	State     models.WorkflowState `json:"state"`
	ItemCount int                  `json:"item_count"`
	Severity  string               `json:"severity"`
	Message   string               `json:"message"`
}

type CachedJobsResult struct {
	Jobs  []models.ReceivedItem `json:"jobs"`
	Total int                   `json:"total"`
}

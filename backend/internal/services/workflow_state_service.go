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
	AdvanceToShipping(ctx context.Context, workOrder string) error
	MarkAsComplete(ctx context.Context, workOrder string) error

	// Status and history methods
	GetCurrentState(ctx context.Context, workOrder string) (*string, error)
	GetWorkflowStatus(ctx context.Context, workOrder string) (*models.WorkflowStatus, error)
	GetStateHistory(ctx context.Context, workOrder string) ([]models.StateChange, error)
	
	// Bulk operations
	GetItemsByState(ctx context.Context, state string) ([]string, error)
	GetJobsByState(ctx context.Context, state string, limit, offset int) ([]models.ReceivedItem, int, error)
	
	// Validation and business logic
	ValidateTransition(ctx context.Context, workOrder string, targetState string) error
	CanAdvanceToNextState(ctx context.Context, workOrder string) (bool, []string, error)
	TransitionTo(ctx context.Context, workOrder string, state string, reason string) error // Add this method
	
	// Analytics and monitoring
	GetWorkflowMetrics(ctx context.Context) (*WorkflowMetrics, error)
	GetBottlenecks(ctx context.Context) ([]WorkflowBottleneck, error)
	GetOverdueItems(ctx context.Context, state string) ([]models.ReceivedItem, error)
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
	// Validate transition using string
	if err := s.ValidateTransition(ctx, workOrder, "in_production"); err != nil {
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
	// Validate transition using string
	if err := s.ValidateTransition(ctx, workOrder, "inspection"); err != nil {
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
	// Validate transition using string
	if err := s.ValidateTransition(ctx, workOrder, "inventory"); err != nil {
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

func (s *workflowStateService) AdvanceToShipping(ctx context.Context, workOrder string) error {
	// Validate transition
	if err := s.ValidateTransition(ctx, workOrder, "shipped"); err != nil {
		return fmt.Errorf("cannot advance to shipping: %w", err)
	}

	// Execute transition
	if err := s.workflowRepo.AdvanceToShipping(ctx, workOrder); err != nil {
		return fmt.Errorf("failed to advance to shipping: %w", err)
	}

	// Invalidate caches
	s.invalidateWorkflowCaches(workOrder)

	return nil
}

func (s *workflowStateService) MarkAsComplete(ctx context.Context, workOrder string) error {
	// Validate transition using string (completed instead of StateComplete)
	if err := s.ValidateTransition(ctx, workOrder, "completed"); err != nil {
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

func (s *workflowStateService) GetCurrentState(ctx context.Context, workOrder string) (models.WorkflowState, error) {
	cacheKey := fmt.Sprintf("workflow:state:%s", workOrder)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if state, ok := cached.(models.WorkflowState); ok {
			return state, nil
		}
	}

	state, err := s.workflowRepo.GetCurrentState(ctx, workOrder)
	if err != nil {
		return "", fmt.Errorf("failed to get workflow state: %w", err)
	}

	s.cache.SetWithTTL(cacheKey, state, 2*time.Minute)
	return state, nil
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

	// Return empty history for now
	history := []models.StateChange{}

	// Cache for 5 minutes
	s.cache.SetWithTTL(cacheKey, history, 5*time.Minute)

	return history, nil
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

	// Get items in this state
	jobs, _, err := s.workflowRepo.GetJobsByState(ctx, state, 1000, 0)
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

func (s *workflowStateService) GetJobsByState(ctx context.Context, state models.WorkflowState, limit, offset int) ([]models.ReceivedItem, int, error) {
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

	// Cache result
	result := CachedJobsResult{Jobs: jobs, Total: total}
	s.cache.SetWithTTL(cacheKey, result, 2*time.Minute)

	return jobs, total, nil
}

// Validation and business logic

func (s *workflowStateService) ValidateTransition(ctx context.Context, workOrder string, targetState string) error {
	// Use existing WorkflowStatus instead of duplicate logic
	status, err := s.GetWorkflowStatus(ctx, workOrder)
	if err != nil {
		return fmt.Errorf("failed to get workflow status: %w", err)
	}
	
	// Use WorkflowStatus.IsValidTransition (already exists and works)
	if !status.IsValidTransition(targetState) {
		return fmt.Errorf("invalid transition from %s to %s", status.CurrentState, targetState)
	}
	
	// Keep existing business logic validations
	switch targetState {
	case "in_production":
		if err := s.validateForProduction(ctx, workOrder); err != nil {
			return fmt.Errorf("production validation failed: %w", err)
		}
	case "inspection":
		if status.CurrentState != "in_production" {
			return fmt.Errorf("must complete production before inspection")
		}
	case "inventory":
		if status.CurrentState != "inspection" {
			return fmt.Errorf("must complete inspection before moving to inventory")
		}
	case "completed":
		if status.CurrentState != "inventory" && status.CurrentState != "inspection" {
			return fmt.Errorf("must be in inventory or inspection before marking complete")
		}
	}

	return nil
}

func (s *workflowStateService) CanAdvanceToNextState(ctx context.Context, workOrder string) (bool, []string, error) {
	// Use existing WorkflowStatus instead of duplicate logic
	status, err := s.GetWorkflowStatus(ctx, workOrder)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get workflow status: %w", err)
	}
	
	// Use WorkflowStatus.GetNextStates (already exists and works)
	nextStates := status.GetNextStates()
	canAdvance := len(nextStates) > 0
	
	// Additional validation checks
	if canAdvance {
		for _, nextState := range nextStates {
			if err := s.ValidateTransition(ctx, workOrder, nextState); err != nil {
				canAdvance = false
				break
			}
		}
	}

	return canAdvance, nextStates, nil
}

func (s *workflowStateService) TransitionTo(ctx context.Context, workOrder string, state string, reason string) error {
	// Validate state
	if err := models.ValidateWorkflowState(state); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}
	
	// Validate transition
	if err := s.ValidateTransition(ctx, workOrder, state); err != nil {
		return fmt.Errorf("invalid transition: %w", err)
	}
	
	// Execute via repository
	if err := s.workflowRepo.TransitionTo(ctx, workOrder, state, reason); err != nil {
		return fmt.Errorf("failed to transition: %w", err)
	}
	
	// Invalidate caches
	s.invalidateWorkflowCaches(workOrder)
	
	return nil
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
		StateDistribution: make(map[string]int),
		LastUpdated:       time.Now(),
	}

	// Count items in each state using WorkflowState constants
	states := []models.WorkflowState{
		models.StateReceived,
		models.StateProduction,
		models.StateInspection,
		models.StateInventory,
		models.StateCompleted,
	}

	totalItems := 0
	for _, state := range states {
		items, err := s.GetItemsByState(ctx, state)
		if err != nil {
			metrics.StateDistribution[string(state)] = 0
		} else {
			count := len(items)
			metrics.StateDistribution[string(state)] = count
			totalItems += count
		}
	}

	metrics.TotalItems = totalItems

	// Calculate average processing time
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

	// Check each state for potential bottlenecks using WorkflowState constants
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
		if len(items) > 20 {
			bottlenecks = append(bottlenecks, WorkflowBottleneck{
				State:     string(state), // Convert to string for JSON
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
	jobs, _, err := s.GetJobsByState(ctx, state, 1000, 0)
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
	item, err := s.receivedRepo.GetByWorkOrder(ctx, workOrder)
	if err != nil {
		return fmt.Errorf("work order not found: %w", err)
	}

	if item.Joints <= 0 {
		return fmt.Errorf("cannot move to production: no joints specified")
	}
	if item.Size == "" {
		return fmt.Errorf("cannot move to production: pipe size not specified")
	}
	if item.Grade == "" {
		return fmt.Errorf("cannot move to production: pipe grade not specified")
	}
	if item.InProduction != nil {
		return fmt.Errorf("work order already in production")
	}

	return nil
}

func (s *workflowStateService) calculateAverageProcessingTime(ctx context.Context) float64 {
	// Simplified calculation
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
	TotalItems              int                 `json:"total_items"`
	StateDistribution       map[string]int      `json:"state_distribution"`
	AverageProcessingDays   float64             `json:"average_processing_days"`
	LastUpdated             time.Time           `json:"last_updated"`
}

type WorkflowBottleneck struct {
	State     string  `json:"state"`
	ItemCount int     `json:"item_count"`
	Severity  string  `json:"severity"`
	Message   string  `json:"message"`
}

type CachedJobsResult struct {
	Jobs  []models.ReceivedItem `json:"jobs"`
	Total int                   `json:"total"`
}

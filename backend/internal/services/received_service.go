// backend/internal/services/received_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/pkg/validation"
)

type ReceivedService interface {
	GetAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.ReceivedItem, int, error)
	GetByID(ctx context.Context, id int) (*models.ReceivedItem, error)
	Create(ctx context.Context, req *validation.ReceivedValidation) (*models.ReceivedItem, error)
	Update(ctx context.Context, id int, req *validation.ReceivedValidation) (*models.ReceivedItem, error)
	Delete(ctx context.Context, id int) error
	GetByWorkOrder(ctx context.Context, workOrder string) (*models.ReceivedItem, error)
	UpdateStatus(ctx context.Context, id int, status string, notes string) error
	GetPendingInspection(ctx context.Context) ([]models.ReceivedItem, error)
	GetByCustomer(ctx context.Context, customerID int, limit, offset int) ([]models.ReceivedItem, int, error)
	GetOverdueItems(ctx context.Context) ([]models.ReceivedItem, error)
	ValidateWorkOrder(ctx context.Context, workOrder string) error
}

type receivedService struct {
	receivedRepo  repository.ReceivedRepository
	customerRepo  repository.CustomerRepository
	cache         *cache.Cache
}

func NewReceivedService(
	receivedRepo repository.ReceivedRepository,
	customerRepo repository.CustomerRepository,
	cache *cache.Cache,
) ReceivedService {
	return &receivedService{
		receivedRepo: receivedRepo,
		customerRepo: customerRepo,
		cache:        cache,
	}
}

func (s *receivedService) GetAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.ReceivedItem, int, error) {
	// Build cache key from filters
	cacheKey := fmt.Sprintf("received:filtered:%v:%d:%d", filters, limit, offset)
	
	// Try cache first for short-term caching (1 minute)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(CachedReceivedResult); ok {
			return result.Items, result.Total, nil
		}
	}

	// Convert filters to repository filters
	repoFilters := s.convertToRepoFilters(filters)
	
	// Get from repository
	items, pagination, err := s.receivedRepo.GetFiltered(ctx, repoFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get received items: %w", err)
	}

	total := 0
	if pagination != nil {
		total = pagination.Total
	}

	// Cache result briefly
	result := CachedReceivedResult{Items: items, Total: total}
	s.cache.SetWithTTL(cacheKey, result, time.Minute)

	return items, total, nil
}

func (s *receivedService) GetByID(ctx context.Context, id int) (*models.ReceivedItem, error) {
	cacheKey := fmt.Sprintf("received:id:%d", id)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if item, ok := cached.(*models.ReceivedItem); ok {
			return item, nil
		}
	}

	item, err := s.receivedRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get received item %d: %w", id, err)
	}

	// Cache for 5 minutes
	s.cache.SetWithTTL(cacheKey, item, 5*time.Minute)

	return item, nil
}

func (s *receivedService) Create(ctx context.Context, req *validation.ReceivedValidation) (*models.ReceivedItem, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	customer, err := s.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer %d not found: %w", req.CustomerID, err)
	}

	item := req.ToReceivedModel()
	item.Customer = customer.Customer
	item.DateReceived = timePtr(time.Now())
	item.CreatedAt = time.Now()

	if err := s.receivedRepo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create received item: %w", err)
	}

	s.invalidateReceivedCaches()

	return item, nil
}

func (s *receivedService) Update(ctx context.Context, id int, req *validation.ReceivedValidation) (*models.ReceivedItem, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	existing, err := s.receivedRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("received item %d not found: %w", id, err)
	}

	// Update fields from validation
	existing.CustomerID = req.CustomerID
	existing.Joints = req.Joints
	existing.Size = validation.NormalizeSize(req.Size)
	existing.Weight = req.Weight
	existing.Grade = validation.NormalizeGrade(req.Grade)
	existing.Connection = validation.NormalizeConnection(req.Connection)
	existing.Well = req.Well
	existing.Lease = req.Lease
	existing.Notes = req.Notes
	existing.WhenUpdated = timePtr(time.Now()) // Use WhenUpdated instead of UpdatedAt

	if err := s.receivedRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update received item: %w", err)
	}

	s.cache.Delete(fmt.Sprintf("received:id:%d", id))
	s.invalidateReceivedCaches()

	return existing, nil
}

func (s *receivedService) Delete(ctx context.Context, id int) error {
	// Check if item can be deleted
	canDelete, reason, err := s.receivedRepo.CanDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if item can be deleted: %w", err)
	}

	if !canDelete {
		return fmt.Errorf("cannot delete received item: %s", reason)
	}

	// Delete from repository
	if err := s.receivedRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete received item: %w", err)
	}

	// Invalidate caches
	s.cache.Delete(fmt.Sprintf("received:id:%d", id))
	s.invalidateReceivedCaches()

	return nil
}

func (s *receivedService) GetByWorkOrder(ctx context.Context, workOrder string) (*models.ReceivedItem, error) {
	cacheKey := fmt.Sprintf("received:work_order:%s", workOrder)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if item, ok := cached.(*models.ReceivedItem); ok {
			return item, nil
		}
	}

	filters := repository.ReceivedFilters{
		WorkOrder: &workOrder,
		Page:      1,
		PerPage:   1,
	}

	items, _, err := s.receivedRepo.GetFiltered(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get received item by work order: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("work order %s not found", workOrder)
	}

	item := &items[0]
	s.cache.SetWithTTL(cacheKey, item, 5*time.Minute)

	return item, nil
}

func (s *receivedService) UpdateStatus(ctx context.Context, id int, status string, notes string) error {
	// Validate status
	validStatuses := []string{"received", "in_production", "inspected", "completed"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Update in repository
	if err := s.receivedRepo.UpdateStatus(ctx, id, status, notes); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Invalidate caches
	s.cache.Delete(fmt.Sprintf("received:id:%d", id))
	s.invalidateReceivedCaches()

	return nil
}

func (s *receivedService) GetPendingInspection(ctx context.Context) ([]models.ReceivedItem, error) {
	cacheKey := "received:pending_inspection"
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if items, ok := cached.([]models.ReceivedItem); ok {
			return items, nil
		}
	}

	// Fix: Status field is string, not *string based on the errors
	filters := repository.ReceivedFilters{
		Status: "in_production",
		Page:   1,
		PerPage: 100,
	}

	items, _, err := s.receivedRepo.GetFiltered(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending inspection items: %w", err)
	}

	s.cache.SetWithTTL(cacheKey, items, 2*time.Minute)
	return items, nil
}

func (s *receivedService) GetOverdueItems(ctx context.Context) ([]models.ReceivedItem, error) {
	cacheKey := "received:overdue"
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if items, ok := cached.([]models.ReceivedItem); ok {
			return items, nil
		}
	}

	// Fix: Status field is string, not *string
	filters := repository.ReceivedFilters{
		Status: "active",
		Page:   1,
		PerPage: 100,
	}

	items, _, err := s.receivedRepo.GetFiltered(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get items for overdue check: %w", err)
	}

	var overdueItems []models.ReceivedItem
	for _, item := range items {
		if item.IsOverdue() {
			overdueItems = append(overdueItems, item)
		}
	}

	s.cache.SetWithTTL(cacheKey, overdueItems, 5*time.Minute)
	return overdueItems, nil
}

func (s *receivedService) GetByCustomer(ctx context.Context, customerID int, limit, offset int) ([]models.ReceivedItem, int, error) {
	cacheKey := fmt.Sprintf("received:customer:%d:%d:%d", customerID, limit, offset)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(CachedReceivedResult); ok {
			return result.Items, result.Total, nil
		}
	}

	filters := repository.ReceivedFilters{
		CustomerID: &customerID,
		Page:       offset/limit + 1,
		PerPage:    limit,
	}

	items, pagination, err := s.receivedRepo.GetFiltered(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get received items for customer: %w", err)
	}

	total := 0
	if pagination != nil {
		total = pagination.Total
	}

	result := CachedReceivedResult{Items: items, Total: total}
	s.cache.SetWithTTL(cacheKey, result, 3*time.Minute)

	return items, total, nil
}

func (s *receivedService) ValidateWorkOrder(ctx context.Context, workOrder string) error {
	// Check if work order already exists
	existing, _ := s.GetByWorkOrder(ctx, workOrder)
	if existing != nil {
		return fmt.Errorf("work order %s already exists", workOrder)
	}

	// Additional validation rules
	if len(workOrder) < 3 {
		return fmt.Errorf("work order must be at least 3 characters")
	}

	return nil
}

// Helper methods for workflow state transitions
func (s *receivedService) AdvanceToProduction(ctx context.Context, workOrder string) error {
	item, err := s.GetByWorkOrder(ctx, workOrder)
	if err != nil {
		return fmt.Errorf("work order not found: %w", err)
	}

	if item.GetCurrentState() != "received" {
		return fmt.Errorf("item not ready for production, current state: %s", item.GetCurrentState())
	}

	// Update the in_production timestamp
	now := time.Now()
	item.InProduction = &now
	item.WhenUpdated = &now

	if err := s.receivedRepo.Update(ctx, item); err != nil {
		return fmt.Errorf("failed to advance to production: %w", err)
	}

	s.invalidateReceivedCaches()
	return nil
}

func (s *receivedService) MarkComplete(ctx context.Context, workOrder string) error {
	item, err := s.GetByWorkOrder(ctx, workOrder)
	if err != nil {
		return fmt.Errorf("work order not found: %w", err)
	}

	if item.GetCurrentState() != "inspected" {
		return fmt.Errorf("item not ready for completion, current state: %s", item.GetCurrentState())
	}

	// Mark as complete
	item.Complete = true
	now := time.Now()
	item.WhenUpdated = &now

	if err := s.receivedRepo.Update(ctx, item); err != nil {
		return fmt.Errorf("failed to mark complete: %w", err)
	}

	s.invalidateReceivedCaches()
	return nil
}

// convertToRepoFilters helper function
func (s *receivedService) convertToRepoFilters(filters map[string]interface{}) repository.ReceivedFilters {
	repoFilters := repository.ReceivedFilters{
		Page:    1,
		PerPage: 50,
	}

	if customerID, ok := filters["customer_id"]; ok {
		if id, ok := customerID.(int); ok {
			repoFilters.CustomerID = &id
		}
	}

	// Status appears to be a string field, not *string
	if status, ok := filters["status"]; ok {
		if s, ok := status.(string); ok {
			repoFilters.Status = s
		}
	}

	if workOrder, ok := filters["work_order"]; ok {
		if wo, ok := workOrder.(string); ok {
			repoFilters.WorkOrder = &wo
		}
	}

	if grade, ok := filters["grade"]; ok {
		if g, ok := grade.(string); ok {
			repoFilters.Grade = &g
		}
	}

	if size, ok := filters["size"]; ok {
		if s, ok := size.(string); ok {
			repoFilters.Size = &s
		}
	}

	if connection, ok := filters["connection"]; ok {
		if c, ok := connection.(string); ok {
			repoFilters.Connection = &c
		}
	}

	return repoFilters
}

func (s *receivedService) invalidateReceivedCaches() {
	// Clear common cache patterns
	patterns := []string{
		"received:filtered:",
		"received:pending_inspection",
		"received:overdue",
		"received:customer:",
	}

	for _, pattern := range patterns {
		s.cache.DeletePattern(pattern)
	}
}

// Helper types
type CachedReceivedResult struct {
	Items []models.ReceivedItem `json:"items"`
	Total int                   `json:"total"`
}

// Utility functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func stringPtr(s string) *string {
	return &s
}

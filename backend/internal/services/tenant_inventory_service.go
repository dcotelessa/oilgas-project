// backend/internal/services/tenant_inventory_service.go
// Enhanced InventoryService with tenant isolation
package services

import (
	"context"
	"fmt"
	"strings"

	"oilgas-backend/internal/repository"
)

// TenantInventoryService extends InventoryService with tenant capabilities
type TenantInventoryService struct {
	*InventoryService // Embed existing service
	tenantRepo        repository.TenantInventoryRepository
}

func NewTenantInventoryService(repo repository.InventoryRepository, tenantRepo repository.TenantInventoryRepository) *TenantInventoryService {
	return &TenantInventoryService{
		InventoryService: NewInventoryService(repo),
		tenantRepo:       tenantRepo,
	}
}

// Tenant-aware methods
func (s *TenantInventoryService) GetInventoryForTenant(ctx context.Context, tenantID string, filters repository.InventoryFilters) ([]repository.InventoryItem, int, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, 0, err
	}
	
	if filters.Limit <= 0 {
		filters.Limit = 50 // Default limit
	}
	if filters.Limit > 1000 {
		filters.Limit = 1000 // Cap at 1000
	}
	
	items, err := s.tenantRepo.GetAllForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, 0, err
	}
	
	total, err := s.tenantRepo.GetCountForTenant(ctx, tenantID, filters)
	if err != nil {
		return items, 0, err // Return items even if count fails
	}
	
	return items, total, nil
}

func (s *TenantInventoryService) GetInventoryByIDForTenant(ctx context.Context, tenantID string, id int) (*repository.InventoryItem, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	if id <= 0 {
		return nil, fmt.Errorf("invalid inventory ID: %d", id)
	}
	return s.tenantRepo.GetByIDForTenant(ctx, tenantID, id)
}

func (s *TenantInventoryService) GetInventoryByWorkOrderForTenant(ctx context.Context, tenantID, workOrder string) ([]repository.InventoryItem, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	workOrder = strings.TrimSpace(workOrder)
	if workOrder == "" {
		return nil, fmt.Errorf("work order is required")
	}
	return s.tenantRepo.GetByWorkOrderForTenant(ctx, tenantID, workOrder)
}

func (s *TenantInventoryService) SearchInventoryForTenant(ctx context.Context, tenantID, query string) ([]repository.InventoryItem, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}
	return s.tenantRepo.SearchForTenant(ctx, tenantID, query)
}

func (s *TenantInventoryService) GetAvailableInventoryForTenant(ctx context.Context, tenantID string) ([]repository.InventoryItem, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	return s.tenantRepo.GetAvailableForTenant(ctx, tenantID)
}

func (s *TenantInventoryService) GetInventorySummaryForTenant(ctx context.Context, tenantID string, filters repository.InventoryFilters) (*repository.InventorySummary, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	return s.tenantRepo.GetSummaryForTenant(ctx, tenantID, filters)
}

// Work Order methods
func (s *TenantInventoryService) GetWorkOrdersForTenant(ctx context.Context, tenantID string, filters repository.WorkOrderFilters) ([]repository.WorkOrder, int, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, 0, err
	}
	
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 1000 {
		filters.Limit = 1000
	}
	
	workOrders, err := s.tenantRepo.GetWorkOrdersForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, 0, err
	}
	
	total, err := s.tenantRepo.GetWorkOrderCountForTenant(ctx, tenantID, filters)
	if err != nil {
		return workOrders, 0, err
	}
	
	return workOrders, total, nil
}

func (s *TenantInventoryService) GetWorkOrderDetailsForTenant(ctx context.Context, tenantID, workOrderID string) (*repository.WorkOrderDetails, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	workOrderID = strings.TrimSpace(workOrderID)
	if workOrderID == "" {
		return nil, fmt.Errorf("work order ID is required")
	}
	return s.tenantRepo.GetWorkOrderDetailsForTenant(ctx, tenantID, workOrderID)
}

// Tenant-aware CRUD operations
func (s *TenantInventoryService) CreateInventoryItemForTenant(ctx context.Context, tenantID string, item *repository.InventoryItem) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}
	if err := s.validateInventoryItem(item); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Set tenant ID in inventory item
	item.TenantID = tenantID
	
	return s.tenantRepo.CreateForTenant(ctx, tenantID, item)
}

func (s *TenantInventoryService) UpdateInventoryItemForTenant(ctx context.Context, tenantID string, item *repository.InventoryItem) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}
	if item.ID <= 0 {
		return fmt.Errorf("invalid inventory ID: %d", item.ID)
	}
	if err := s.validateInventoryItem(item); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Ensure tenant ID matches
	item.TenantID = tenantID
	
	return s.tenantRepo.UpdateForTenant(ctx, tenantID, item)
}

func (s *TenantInventoryService) DeleteInventoryItemForTenant(ctx context.Context, tenantID string, id int) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}
	if id <= 0 {
		return fmt.Errorf("invalid inventory ID: %d", id)
	}
	return s.tenantRepo.DeleteForTenant(ctx, tenantID, id)
}

func (s *TenantInventoryService) validateTenantID(tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if len(tenantID) < 2 || len(tenantID) > 20 {
		return fmt.Errorf("tenant ID must be between 2 and 20 characters")
	}
	return nil
}

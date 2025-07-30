// internal/services/inventory_service.go
package services

import (
	"context"
	"fmt"
	"strings"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
)

type InventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) GetInventory(ctx context.Context, filters models.InventoryFilters) ([]models.InventoryItem, error) {
	if filters.Limit <= 0 {
		filters.Limit = 100 // Default limit
	}
	if filters.Limit > 1000 {
		return nil, fmt.Errorf("limit too high (max 1000)")
	}
	return s.repo.GetAll(ctx, filters)
}

func (s *InventoryService) GetInventoryByID(ctx context.Context, id int) (*models.InventoryItem, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid inventory ID: %d", id)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *InventoryService) GetInventoryByWorkOrder(ctx context.Context, workOrder string) ([]models.InventoryItem, error) {
	workOrder = strings.TrimSpace(workOrder)
	if workOrder == "" {
		return nil, fmt.Errorf("work order is required")
	}
	return s.repo.GetByWorkOrder(ctx, workOrder)
}

func (s *InventoryService) GetAvailableInventory(ctx context.Context) ([]models.InventoryItem, error) {
	return s.repo.GetAvailable(ctx)
}

func (s *InventoryService) SearchInventory(ctx context.Context, query string) ([]models.InventoryItem, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}
	return s.repo.Search(ctx, query)
}

func (s *InventoryService) CreateInventoryItem(ctx context.Context, item *models.InventoryItem) error {
	if err := s.validateInventoryItem(item); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return s.repo.Create(ctx, item)
}

func (s *InventoryService) UpdateInventoryItem(ctx context.Context, item *models.InventoryItem) error {
	if item.ID <= 0 {
		return fmt.Errorf("invalid inventory ID: %d", item.ID)
	}
	if err := s.validateInventoryItem(item); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return s.repo.Update(ctx, item)
}

func (s *InventoryService) DeleteInventoryItem(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid inventory ID: %d", id)
	}
	return s.repo.Delete(ctx, id)
}

func (s *InventoryService) validateInventoryItem(item *models.InventoryItem) error {
	if item.Joints != nil && *item.Joints < 0 {
		return fmt.Errorf("joints cannot be negative")
	}
	if item.Weight != nil && *item.Weight < 0 {
		return fmt.Errorf("weight cannot be negative")
	}
	return nil
}

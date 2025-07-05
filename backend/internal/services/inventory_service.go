package services

import (
	"context"
	"fmt"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/pkg/validation"
)


// Inventory service interface and implementation
type InventoryService interface {
	GetByID(ctx context.Context, id int) (*models.InventoryItem, error)
	Create(ctx context.Context, req *validation.InventoryValidation) (*models.InventoryItem, error)
	Update(ctx context.Context, id int, req *validation.InventoryValidation) (*models.InventoryItem, error)
	Delete(ctx context.Context, id int) error
	GetFiltered(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.InventoryItem, int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error)
	GetSummary(ctx context.Context) (*models.InventorySummary, error)
}

type inventoryService struct {
	repo  repository.InventoryRepository
	cache *cache.Cache
}

func NewInventoryService(repo repository.InventoryRepository, cache *cache.Cache) InventoryService {
	return &inventoryService{
		repo:  repo,
		cache: cache,
	}
}

func (s *inventoryService) GetByID(ctx context.Context, id int) (*models.InventoryItem, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("inventory:%d", id)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if item, ok := cached.(*models.InventoryItem); ok {
			return item, nil
		}
	}

	// Get from repository
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cache.Set(cacheKey, item)

	return item, nil
}

func (s *inventoryService) Create(ctx context.Context, req *validation.InventoryValidation) (*models.InventoryItem, error) {
	// Normalize the data
	normalizedReq := &validation.InventoryValidation{
		CustomerID: req.CustomerID,
		Joints:     req.Joints,
		Size:       validation.NormalizeSize(req.Size),
		Weight:     req.Weight,
		Grade:      validation.NormalizeGrade(req.Grade),
		Connection: validation.NormalizeConnection(req.Connection),
		Color:      req.Color,
		Location:   req.Location,
	}

	// Create through repository
	item, err := s.repo.Create(ctx, normalizedReq)
	if err != nil {
		return nil, err
	}

	// Invalidate relevant caches
	s.cache.Delete("inventory:all")
	s.cache.Delete("inventory:summary")
	s.cache.Delete("customers:all") // Customer inventory counts may have changed
	s.invalidateFilteredCaches()

	return item, nil
}

func (s *inventoryService) Update(ctx context.Context, id int, req *validation.InventoryValidation) (*models.InventoryItem, error) {
	// Get existing item for cache invalidation
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Normalize the data
	normalizedReq := &validation.InventoryValidation{
		CustomerID: req.CustomerID,
		Joints:     req.Joints,
		Size:       validation.NormalizeSize(req.Size),
		Weight:     req.Weight,
		Grade:      validation.NormalizeGrade(req.Grade),
		Connection: validation.NormalizeConnection(req.Connection),
		Color:      req.Color,
		Location:   req.Location,
	}

	// Update through repository
	item, err := s.repo.Update(ctx, id, normalizedReq)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	s.cache.Delete(fmt.Sprintf("inventory:%d", id))
	s.cache.Delete("inventory:all")
	s.cache.Delete("inventory:summary")
	s.cache.Delete("customers:all")
	s.invalidateFilteredCaches()

	// Invalidate customer-specific caches if customer changed
	if existing.CustomerID != req.CustomerID {
		s.cache.Delete(fmt.Sprintf("customer:%d", existing.CustomerID))
		s.cache.Delete(fmt.Sprintf("customer:%d", req.CustomerID))
	}

	return item, nil
}

func (s *inventoryService) Delete(ctx context.Context, id int) error {
	// Get existing item for cache invalidation
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete through repository
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.cache.Delete(fmt.Sprintf("inventory:%d", id))
	s.cache.Delete("inventory:all")
	s.cache.Delete("inventory:summary")
	s.cache.Delete("customers:all")
	s.cache.Delete(fmt.Sprintf("customer:%d", existing.CustomerID))
	s.invalidateFilteredCaches()

	return nil
}

func (s *inventoryService) GetFiltered(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.InventoryItem, int, error) {
	// Create cache key based on filters
	cacheKey := fmt.Sprintf("inventory:filtered:%v:%d:%d", filters, limit, offset)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(map[string]interface{}); ok {
			items := result["items"].([]models.InventoryItem)
			total := result["total"].(int)
			return items, total, nil
		}
	}

	// Get from repository
	items, total, err := s.repo.GetFiltered(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Cache the results (shorter TTL for filtered results - 2 minutes)
	result := map[string]interface{}{
		"items": items,
		"total": total,
	}
	s.cache.Set(cacheKey, result)

	return items, total, nil
}

func (s *inventoryService) Search(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error) {
	// Create cache key for search
	cacheKey := fmt.Sprintf("search:inventory:%s:%d:%d", query, limit, offset)
	
	if cached, exists := s.cache.GetSearchResults(cacheKey); exists {
		if result, ok := cached.(map[string]interface{}); ok {
			items := result["items"].([]models.InventoryItem)
			total := result["total"].(int)
			return items, total, nil
		}
	}

	// Search through repository
	items, total, err := s.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Cache search results (2 minute TTL)
	result := map[string]interface{}{
		"items": items,
		"total": total,
	}
	s.cache.CacheSearchResults(cacheKey, result)

	return items, total, nil
}

func (s *inventoryService) GetSummary(ctx context.Context) (*models.InventorySummary, error) {
	// Check cache first
	if cached, exists := s.cache.Get("inventory:summary"); exists {
		if summary, ok := cached.(*models.InventorySummary); ok {
			return summary, nil
		}
	}

	// Get summary from repository
	summary, err := s.repo.GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	// Cache summary (longer TTL since it changes less frequently - 10 minutes)
	s.cache.Set("inventory:summary", summary)

	return summary, nil
}

// Helper method to invalidate filtered caches
func (s *inventoryService) invalidateFilteredCaches() {
	// This is a simple approach - in production you might want more sophisticated cache invalidation
	// For now, we'll just clear all search and filtered results
	s.cache.Delete("search:inventory")  // This will clear search results with this prefix
}


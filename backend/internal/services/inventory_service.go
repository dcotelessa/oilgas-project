package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/pkg/validation"
)

// CachedInventoryResult represents cached search/filter results
type CachedInventoryResult struct {
	Items []models.InventoryItem `json:"items"`
	Total int                    `json:"total"`
}

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
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	// Cache the result
	s.cache.Set(cacheKey, item)

	return item, nil
}

func (s *inventoryService) Create(ctx context.Context, req *validation.InventoryValidation) (*models.InventoryItem, error) {
	// Validate the request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", errors)
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

	// Convert validation to model
	item := normalizedReq.ToInventoryModel()

	// Create through repository (returns only error, not item)
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}

	// Cache the new item
	if item.ID != 0 {
		cacheKey := fmt.Sprintf("inventory:%d", item.ID)
		s.cache.Set(cacheKey, item)
	}

	// Invalidate relevant caches
	s.invalidateInventoryCaches()
	s.invalidateFilteredCaches()

	return item, nil
}

func (s *inventoryService) Update(ctx context.Context, id int, req *validation.InventoryValidation) (*models.InventoryItem, error) {
	// Validate the request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	// Get existing item for cache invalidation and validation
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("inventory item not found: %w", err)
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

	// Convert validation to model and preserve ID and timestamps
	item := normalizedReq.ToInventoryModel()
	item.ID = id
	item.CreatedAt = existing.CreatedAt // Preserve creation time

	// Update through repository (expects *models.InventoryItem, returns only error)
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("inventory:%d", id)
	s.cache.Set(cacheKey, item)

	// Invalidate relevant caches
	s.invalidateInventoryCaches()
	s.invalidateFilteredCaches()

	// Invalidate customer-specific caches if customer changed
	if existing.CustomerID != req.CustomerID {
		s.invalidateCustomerCaches(existing.CustomerID)
		s.invalidateCustomerCaches(req.CustomerID)
	}

	return item, nil
}

func (s *inventoryService) Delete(ctx context.Context, id int) error {
	// Get existing item for cache invalidation
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("inventory item not found: %w", err)
	}

	// Soft delete through repository
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("inventory:%d", id)
	s.cache.Delete(cacheKey)

	// Invalidate relevant caches
	s.invalidateInventoryCaches()
	s.invalidateFilteredCaches()
	s.invalidateCustomerCaches(existing.CustomerID)

	return nil
}

func (s *inventoryService) GetFiltered(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.InventoryItem, int, error) {
	// Create cache key based on filters
	cacheKey := fmt.Sprintf("inventory:filtered:%v:%d:%d", filters, limit, offset)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(CachedInventoryResult); ok {
			return result.Items, result.Total, nil
		}
	}

	// Convert map[string]interface{} to repository.InventoryFilters
	repoFilters := s.convertToInventoryFilters(filters, limit, offset)

	// Get from repository (returns items, *models.Pagination, error)
	items, pagination, err := s.repo.GetFiltered(ctx, repoFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get filtered inventory: %w", err)
	}

	// Extract total from pagination
	total := 0
	if pagination != nil {
		total = pagination.Total
	}

	// Cache the results with shorter TTL for filtered results
	result := CachedInventoryResult{Items: items, Total: total}
	s.cache.SetWithTTL(cacheKey, result, 2*time.Minute)

	return items, total, nil
}

func (s *inventoryService) Search(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error) {
	// Validate input
	if strings.TrimSpace(query) == "" {
		return []models.InventoryItem{}, 0, nil
	}

	// Create cache key for search
	cacheKey := fmt.Sprintf("search:inventory:%s:%d:%d", query, limit, offset)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(CachedInventoryResult); ok {
			return result.Items, result.Total, nil
		}
	}

	// Get from repository
	items, total, err := s.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search inventory: %w", err)
	}

	// Cache the results with shorter TTL for search results
	result := CachedInventoryResult{Items: items, Total: total}
	s.cache.SetWithTTL(cacheKey, result, 3*time.Minute)

	return items, total, nil
}

func (s *inventoryService) GetSummary(ctx context.Context) (*models.InventorySummary, error) {
	cacheKey := "inventory:summary"
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if summary, ok := cached.(*models.InventorySummary); ok {
			return summary, nil
		}
	}

	summary, err := s.repo.GetSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory summary: %w", err)
	}

	// Cache summary for longer since it's expensive to compute
	s.cache.SetWithTTL(cacheKey, summary, 5*time.Minute)
	
	return summary, nil
}

// Helper methods
func (s *inventoryService) convertToInventoryFilters(filters map[string]interface{}, limit, offset int) repository.InventoryFilters {
	// Calculate page number (1-based)
	page := 1
	if limit > 0 {
		page = (offset / limit) + 1
	}

	repoFilters := repository.InventoryFilters{
		Page:    page,
		PerPage: limit,
	}

	// Convert customer ID with type safety
	if customerID, exists := filters["customer_id"]; exists {
		switch v := customerID.(type) {
		case int:
			repoFilters.CustomerID = &v
		case float64: // JSON numbers come as float64
			id := int(v)
			repoFilters.CustomerID = &id
		case string:
			if id, err := strconv.Atoi(v); err == nil {
				repoFilters.CustomerID = &id
			}
		}
	}

	// Convert string filters with normalization
	if grade, ok := filters["grade"].(string); ok && grade != "" {
		repoFilters.Grade = validation.NormalizeGrade(grade)
	}

	if size, ok := filters["size"].(string); ok && size != "" {
		repoFilters.Size = validation.NormalizeSize(size)
	}

	if connection, ok := filters["connection"].(string); ok && connection != "" {
		repoFilters.Connection = validation.NormalizeConnection(connection)
	}

	if color, ok := filters["color"].(string); ok && color != "" {
		repoFilters.Color = strings.TrimSpace(color)
	}

	if location, ok := filters["location"].(string); ok && location != "" {
		repoFilters.Location = strings.TrimSpace(location)
	}

	if rack, ok := filters["rack"].(string); ok && rack != "" {
		repoFilters.Rack = strings.TrimSpace(rack)
	}

	// Convert joint count filters with type safety
	if minJoints, exists := filters["min_joints"]; exists {
		switch v := minJoints.(type) {
		case int:
			repoFilters.MinJoints = &v
		case float64:
			joints := int(v)
			repoFilters.MinJoints = &joints
		}
	}

	if maxJoints, exists := filters["max_joints"]; exists {
		switch v := maxJoints.(type) {
		case int:
			repoFilters.MaxJoints = &v
		case float64:
			joints := int(v)
			repoFilters.MaxJoints = &joints
		}
	}

	// Convert date filters
	if dateFrom, ok := filters["date_from"].(string); ok && dateFrom != "" {
		repoFilters.DateFrom = dateFrom
	}

	if dateTo, ok := filters["date_to"].(string); ok && dateTo != "" {
		repoFilters.DateTo = dateTo
	}

	// Handle boolean filters
	if includeDeleted, ok := filters["include_deleted"].(bool); ok {
		repoFilters.IncludeDeleted = includeDeleted
	}

	return repoFilters
}

func (s *inventoryService) invalidateFilteredCaches() {
	// If you have DeletePattern method (recommended):
	if hasDeletePattern := true; hasDeletePattern { // Replace with actual check
		s.cache.DeletePattern("inventory:filtered:")
		s.cache.DeletePattern("search:inventory:")
	} else {
		// Fallback: Clear common cache entries manually
		s.cache.Delete("inventory:filtered")
		s.cache.Delete("search:inventory")
		
		// Clear summary which might include filtered data
		s.cache.Delete("inventory:summary")
	}
}

func (s *inventoryService) invalidateInventoryCaches() {
	s.cache.Delete("inventory:all")
	s.cache.Delete("inventory:summary")
	s.cache.Delete("inventory:count")
	
	// If you have DeletePattern:
	s.cache.DeletePattern("inventory:customer:")
	s.cache.DeletePattern("inventory:grade:")
	s.cache.DeletePattern("inventory:location:")
}

func (s *inventoryService) invalidateCustomerCaches(customerID int) {
	s.cache.Delete(fmt.Sprintf("customer:%d", customerID))
	s.cache.Delete(fmt.Sprintf("customer:%d:inventory", customerID))
	s.cache.Delete(fmt.Sprintf("inventory:customer:%d", customerID))
	s.cache.Delete("customers:all")
}


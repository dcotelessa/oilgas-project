package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/pkg/validation"
)

// Customer service interface and implementation
type CustomerService interface {
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetByID(ctx context.Context, id string) (*models.Customer, error)
	Create(ctx context.Context, req *validation.CustomerValidation) (*models.Customer, error)
	Update(ctx context.Context, id string, req *validation.CustomerValidation) (*models.Customer, error)
	Delete(ctx context.Context, id string) error
}

type customerService struct {
	repo  repository.CustomerRepository
	cache *cache.Cache
}

func NewCustomerService(repo repository.CustomerRepository, cache *cache.Cache) CustomerService {
	return &customerService{
		repo:  repo,
		cache: cache,
	}
}

func (s *customerService) GetAll(ctx context.Context) ([]models.Customer, error) {
	// Check cache first
	if cached, exists := s.cache.Get("customers:all"); exists {
		if customers, ok := cached.([]models.Customer); ok {
			return customers, nil
		}
	}

	// Get from repository
	customers, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the results (5 minute TTL)
	s.cache.Set("customers:all", customers)

	return customers, nil
}

func (s *customerService) GetByID(ctx context.Context, idStr string) (*models.Customer, error) {
	// Convert string ID to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	// Check cache first
	cacheKey := fmt.Sprintf("customer:%d", id)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if customer, ok := cached.(*models.Customer); ok {
			return customer, nil
		}
	}

	// Get from repository
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cache.Set(cacheKey, customer)

	return customer, nil
}

func (s *customerService) Create(ctx context.Context, req *validation.CustomerValidation) (*models.Customer, error) {
	// Normalize customer data for consistency
	req.NormalizeCustomerData()

	// Business logic validation
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check for duplicate customer names
	exists, err := s.repo.ExistsByName(ctx, req.CustomerName)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate customer: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("customer with name '%s' already exists", req.CustomerName)
	}

	// Create through repository
	customer, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Cache the new customer
	cacheKey := fmt.Sprintf("customer:%d", customer.CustomerID)
	s.cache.Set(cacheKey, customer)
	
	// Invalidate relevant caches
	s.invalidateCustomerCaches()

	return customer, nil
}

func (s *customerService) Update(ctx context.Context, idStr string, req *validation.CustomerValidation) (*models.Customer, error) {
	// Convert string ID to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	// Get existing customer for cache invalidation
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Normalize the data
	req.NormalizeCustomerData()

	// Business logic validation
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check for duplicate customer names (excluding current customer)
	exists, err := s.repo.ExistsByName(ctx, req.CustomerName, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate customer: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("customer with name '%s' already exists", req.CustomerName)
	}

	// Update through repository
	customer, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("customer:%d", id)
	s.cache.Set(cacheKey, customer)
	
	// Invalidate relevant caches
	s.invalidateCustomerCaches()

	// If customer name changed, invalidate name-based caches
	if existing.Customer != customer.Customer {
		s.cache.Delete(fmt.Sprintf("customer:name:%s", strings.ToLower(existing.Customer)))
		s.cache.Delete(fmt.Sprintf("customer:name:%s", strings.ToLower(customer.Customer)))
	}

	return customer, nil
}

func (s *customerService) Delete(ctx context.Context, idStr string) error {
	// Convert string ID to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("invalid customer ID: %w", err)
	}

	// Get existing customer for cache invalidation
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Check if customer has active inventory
	hasInventory, err := s.repo.HasActiveInventory(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check customer inventory: %w", err)
	}
	if hasInventory {
		return fmt.Errorf("cannot delete customer with active inventory items")
	}

	// Soft delete through repository
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("customer:%d", id)
	s.cache.Delete(cacheKey)
	s.cache.Delete(fmt.Sprintf("customer:name:%s", strings.ToLower(existing.Customer)))
	
	// Invalidate relevant caches
	s.invalidateCustomerCaches()

	return nil
}

// invalidateCustomerCaches clears all customer-related cache entries
func (s *customerService) invalidateCustomerCaches() {
	// Clear common customer cache keys
	cacheKeys := []string{
		"customers:all",
		"customers:total_count",
		"customer_activity_30",    // 30-day activity
		"customer_activity_7",     // 7-day activity
		"top_customers_5_30",      // Top 5 customers, 30 days
		"top_customers_10_30",     // Top 10 customers, 30 days
	}
	
	for _, key := range cacheKeys {
		s.cache.Delete(key)
	}
}

// backend/internal/services/services.go
package services

import (
	"context"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
)

type Services struct {
	Customer  CustomerService
	Inventory InventoryService
	Cache     *cache.Cache
}

func New(repos *repository.Repositories, cache *cache.Cache) *Services {
	return &Services{
		Customer:  NewCustomerService(repos.Customer, cache),
		Inventory: NewInventoryService(repos.Inventory, cache),
		Cache:     cache,
	}
}

// Customer service interface and implementation
type CustomerService interface {
	GetAll(ctx context.Context) ([]interface{}, error)
	GetByID(ctx context.Context, id string) (interface{}, error)
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

func (s *customerService) GetAll(ctx context.Context) ([]interface{}, error) {
	// TODO: Implement with repository
	return []interface{}{}, nil
}

func (s *customerService) GetByID(ctx context.Context, id string) (interface{}, error) {
	// TODO: Implement with repository and cache
	return map[string]interface{}{
		"id":   id,
		"name": "Sample Customer",
	}, nil
}

// Inventory service interface and implementation
type InventoryService interface {
	GetAll(ctx context.Context) ([]interface{}, error)
	GetByID(ctx context.Context, id string) (interface{}, error)
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

func (s *inventoryService) GetAll(ctx context.Context) ([]interface{}, error) {
	// TODO: Implement with repository
	return []interface{}{}, nil
}

func (s *inventoryService) GetByID(ctx context.Context, id string) (interface{}, error) {
	// TODO: Implement with repository and cache
	return map[string]interface{}{
		"id":    id,
		"grade": "J55",
		"size":  "5.5\"",
	}, nil
}

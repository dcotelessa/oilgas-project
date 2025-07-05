// backend/internal/services/services.go
package services

import (
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

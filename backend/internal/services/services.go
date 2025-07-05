// backend/internal/services/services.go
package services

import (
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
)

// Services aggregates all service interfaces
type Services struct {
	Customer      CustomerService
	Inventory     InventoryService
	Analytics     AnalyticsService
	WorkflowState WorkflowStateService
	Grade         GradeService
	Received      ReceivedService
	Search        SearchService
	Cache         *cache.Cache
}

// New creates a new service manager with all implementations
func New(repos *repository.Repositories, cache *cache.Cache) *Services {
	return &Services{
		Customer:      NewCustomerService(repos.Customer, cache),
		Inventory:     NewInventoryService(repos.Inventory, cache),
		Analytics:     NewAnalyticsService(repos.Analytics, cache),
		WorkflowState: NewWorkflowStateService(repos.WorkflowState, cache),
		Grade:         NewGradeService(repos.Grade, cache),
		Received:      NewReceivedService(repos.Received, cache),
		Search:        NewSearchService(repos, cache),
		Cache:         cache,
	}
}

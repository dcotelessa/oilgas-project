// backend/internal/services/services.go
package services

import (
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
)

// Services aggregates all service interfaces
type Services struct {
	Analytics     AnalyticsService
	Customer      CustomerService
	Grade         GradeService
	Inventory     InventoryService
	Received      ReceivedService
	WorkflowState WorkflowStateService
	Search        SearchService
}

func New(repos *repository.Repositories, cache *cache.Cache) *Services {
	return &Services{
		Analytics: NewAnalyticsService(
			repos.Analytics,
			repos.Customer,
			repos.Inventory,
			repos.Received,
			cache,
		),
		Customer:  NewCustomerService(repos.Customer, cache),
		Grade:	NewGradeService(repos.Grade, cache),
		Inventory: NewInventoryService(repos.Inventory, cache),
		Received: NewReceivedService(
			repos.Received,
			repos.Customer,
			cache,
		),
		WorkflowState: NewWorkflowStateService(
			repos.WorkflowState,
			repos.Received,
			cache,
		),
		Search: NewSearchService(
			repos.Customer,
			repos.Inventory,
			repos.Received,
			repos.Grade,
			cache,
		),
	}
}

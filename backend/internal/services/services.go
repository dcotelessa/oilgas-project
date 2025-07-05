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
	Search        SearchService // Composite service using multiple repos
}

func New(repos *repository.Repositories, cache *cache.Cache) *Services {
	return &Services{
		Analytics:     NewAnalyticsService(repos.Analytics, cache),
		Customer:      NewCustomerService(repos.Customer, cache),
		Grade:         NewGradeService(repos.Grade, cache),
		Inventory:     NewInventoryService(repos.Inventory, cache),
		Received:      NewReceivedService(repos.Received, cache),
		WorkflowState: NewWorkflowStateService(repos.WorkflowState, cache),
		Search:        NewSearchService(repos, cache), // Uses multiple repos
	}
}

type AnalyticsService interface {
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)
	GetCustomerActivity(ctx context.Context, days int, customerID *int) (*models.CustomerActivity, error)
	GetInventoryAnalytics(ctx context.Context) (*models.InventoryAnalytics, error)
	GetGradeDistribution(ctx context.Context) (map[string]int, error)
	GetLocationUtilization(ctx context.Context) (*models.LocationStats, error)
}

type ReceivedService interface {
	GetAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.ReceivedItem, int, error)
	GetByID(ctx context.Context, id int) (*models.ReceivedItem, error)
	Create(ctx context.Context, item *models.ReceivedItem) error
	Update(ctx context.Context, item *models.ReceivedItem) error
	Delete(ctx context.Context, id int) error
	GetByWorkOrder(ctx context.Context, workOrder string) (*models.ReceivedItem, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	GetPendingInspection(ctx context.Context) ([]models.ReceivedItem, error)
}

type WorkflowStateService interface {
	GetCurrentState(ctx context.Context, workOrder string) (*models.WorkflowState, error)
	TransitionTo(ctx context.Context, workOrder string, newState models.WorkflowState, notes string) error
	GetStateHistory(ctx context.Context, workOrder string) ([]models.StateTransition, error)
	GetItemsByState(ctx context.Context, state models.WorkflowState) ([]models.WorkOrderSummary, error)
	ValidateTransition(ctx context.Context, from, to models.WorkflowState) error
}

type SearchService interface {
	GlobalSearch(ctx context.Context, query string, filters map[string]interface{}) (*models.SearchResults, error)
	SearchCustomers(ctx context.Context, query string, limit, offset int) ([]models.Customer, int, error)
	SearchInventory(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error)
	SearchReceived(ctx context.Context, query string, limit, offset int) ([]models.ReceivedItem, int, error)
	GetSuggestions(ctx context.Context, query string, limit int) ([]models.SearchSuggestion, error)
}


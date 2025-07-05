// backend/internal/repository/repository.go
package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories aggregates all repository interfaces
type Repositories struct {
	Customer      CustomerRepository
	Inventory     InventoryRepository
	Grade         GradeRepository
	Received      ReceivedRepository
	Analytics     AnalyticsRepository
	WorkflowState WorkflowStateRepository
}

// New creates a new repository manager with all implementations
func New(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Customer:      NewCustomerRepository(db),
		Inventory:     NewInventoryRepository(db),
		Grade:         NewGradeRepository(db),
		Received:      NewReceivedRepository(db),
		Analytics:     NewAnalyticsRepository(db),
		WorkflowState: NewWorkflowStateRepository(db),
	}
}

// Health check for all repositories
func (r *Repositories) HealthCheck() error {
	// Could add ping tests or validation queries here
	return nil
}

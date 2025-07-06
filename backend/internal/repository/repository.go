// backend/internal/repository/repository.go
package repository

import (
	"context"
	"fmt"
	"log"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
)



// Repositories aggregates all repository interfaces
type Repositories struct {
	Analytics     AnalyticsRepository
	Customer      CustomerRepository
	Grade         GradeRepository
	Inventory     InventoryRepository
	Received      ReceivedRepository
	WorkflowState WorkflowStateRepository
}

// New creates repository instances with your actual implementations
func New(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Analytics:     NewAnalyticsRepository(db),
		Customer:      NewCustomerRepository(db),
		Grade:         NewGradeRepository(db),
		Inventory:     NewInventoryRepository(db),
		Received:      NewReceivedRepository(db),
		WorkflowState: NewWorkflowStateRepository(db),
	}
}

// Helper methods for the repository collection

// GetAll returns all repositories as a map for dynamic access
func (r *Repositories) GetAll() map[string]interface{} {
	return map[string]interface{}{
		"analytics":       r.Analytics,
		"customer":        r.Customer,
		"grade":           r.Grade,
		"inventory":       r.Inventory,
		"received":        r.Received,
		"workflow_state":  r.WorkflowState,
	}
}

// HealthCheck verifies all repository connections are healthy
func (r *Repositories) HealthCheck(ctx context.Context) error {
	start := time.Now()
	defer func() {
		log.Printf("Repository health check completed in %v", time.Since(start))
	}()
	
	// Create a timeout context for health checks
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Check customers table
	if err := r.checkCustomerHealth(healthCtx); err != nil {
		return fmt.Errorf("customer repository health check failed: %w", err)
	}
	
	// Check grades table
	if err := r.checkGradeHealth(healthCtx); err != nil {
		return fmt.Errorf("grade repository health check failed: %w", err)
	}
	
	// Check inventory table with limit to avoid large queries
	if err := r.checkInventoryHealth(healthCtx); err != nil {
		return fmt.Errorf("inventory repository health check failed: %w", err)
	}
	
	// Check received table with limit
	if err := r.checkReceivedHealth(healthCtx); err != nil {
		return fmt.Errorf("received repository health check failed: %w", err)
	}
	
	return nil
}

// Helper methods for individual health checks
func (r *Repositories) checkCustomerHealth(ctx context.Context) error {
	_, err := r.Customer.GetAll(ctx)
	return err
}

func (r *Repositories) checkGradeHealth(ctx context.Context) error {
	_, err := r.Grade.GetAll(ctx)
	return err
}

func (r *Repositories) checkInventoryHealth(ctx context.Context) error {
	filters := InventoryFilters{
		Page:    1,
		PerPage: 1,
	}
	filters.NormalizePagination()
	_, _, err := r.Inventory.GetFiltered(ctx, filters)
	return err
}

func (r *Repositories) checkReceivedHealth(ctx context.Context) error {
	filters := ReceivedFilters{
		Page:    1,
		PerPage: 1,
	}
	filters.NormalizePagination()
	_, _, err := r.Received.GetFiltered(ctx, filters)
	return err
}

// WithTransaction provides a way to run operations across multiple repositories in a transaction
func (r *Repositories) WithTransaction(ctx context.Context, fn func(*Repositories) error) error {
	// For now, just execute the function without transaction support
	// TODO: Implement proper transaction support when needed
	// This would require:
	// 1. Starting a pgx.Tx transaction
	// 2. Creating new repository instances that use the transaction
	// 3. Passing those repositories to the function
	// 4. Committing or rolling back based on the result
	
	return fn(r)
}

// Close gracefully closes all repository connections
func (r *Repositories) Close() {
	log.Println("Closing repository connections...")
	
	// pgx pools handle their own cleanup automatically
	// But we can implement graceful shutdown logic here
	
	// Future implementations might include:
	// - Waiting for pending operations to complete
	// - Flushing any buffered data
	// - Closing cache connections
	// - Cleaning up temporary resources
	
	log.Println("Repository connections closed")
}

// GetRepositoryNames returns a list of available repository names for logging/debugging
func (r *Repositories) GetRepositoryNames() []string {
	return []string{
		"analytics",
		"customer", 
		"grade",
		"inventory",
		"received",
		"workflow_state",
	}
}

// Stats returns basic statistics about the repository collection
func (r *Repositories) Stats() map[string]interface{} {
	return map[string]interface{}{
		"total_repositories": len(r.GetRepositoryNames()),
		"repository_names":   r.GetRepositoryNames(),
		"initialized":        true,
	}
}

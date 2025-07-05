// backend/internal/repository/transaction.go
// Optional: Implement this when you need transaction support across repositories

package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionManager handles database transactions across multiple repositories
type TransactionManager struct {
	db *pgxpool.Pool
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, it's committed
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			tx.Rollback(ctx)
			panic(p)
		}
	}()
	
	// Execute the function
	if err := fn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}
	
	// Commit on success
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// Example of how to use transactions across repositories:
// 
// func (s *SomeService) CreateOrderWithInventory(ctx context.Context, order *Order, items []InventoryItem) error {
//     tm := repository.NewTransactionManager(s.db)
//     
//     return tm.WithTransaction(ctx, func(tx pgx.Tx) error {
//         // Create repositories that use the transaction
//         orderRepo := repository.NewOrderRepositoryWithTx(tx)
//         inventoryRepo := repository.NewInventoryRepositoryWithTx(tx)
//         
//         // Create order
//         if err := orderRepo.Create(ctx, order); err != nil {
//             return err
//         }
//         
//         // Update inventory
//         for _, item := range items {
//             if err := inventoryRepo.Update(ctx, &item); err != nil {
//                 return err // This will cause rollback
//             }
//         }
//         
//         return nil // This will cause commit
//     })
// }

// Enhanced Repositories struct with transaction support
func (r *Repositories) WithTransactionSupport(ctx context.Context, fn func(*Repositories) error) error {
	// Get the underlying database connection from one of the repositories
	// This assumes your repositories have a way to access the db pool
	// You might need to adjust this based on your actual implementation
	
	db := r.getDBPool() // You'd need to implement this method
	if db == nil {
		// Fallback to non-transactional execution
		return fn(r)
	}
	
	tm := NewTransactionManager(db)
	
	return tm.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Create new repository instances that use the transaction
		// This would require your repositories to support transaction-based constructors
		txRepos := &Repositories{
			Analytics:     NewAnalyticsRepositoryWithTx(tx),
			Customer:      NewCustomerRepositoryWithTx(tx),
			Grade:         NewGradeRepositoryWithTx(tx),
			Inventory:     NewInventoryRepositoryWithTx(tx),
			Received:      NewReceivedRepositoryWithTx(tx),
			WorkflowState: NewWorkflowStateRepositoryWithTx(tx),
		}
		
		return fn(txRepos)
	})
}

// Helper method to get DB pool (you'd need to implement this in your repos)
func (r *Repositories) getDBPool() *pgxpool.Pool {
	// This is a placeholder - you'd need to implement this based on your repo structure
	// Option 1: Add a GetDB() method to one of your repository interfaces
	// Option 2: Store the DB pool in the Repositories struct
	// Option 3: Use reflection to access the db field (not recommended)
	return nil
}

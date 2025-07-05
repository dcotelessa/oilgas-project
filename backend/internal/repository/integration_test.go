// backend/internal/repository/integration_test.go
// +build integration

package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oilgas-backend/internal/models"
)

// TestCustomerRepository_Integration tests the customer repository with real database
func TestCustomerRepository_Integration(t *testing.T) {
	RunIntegrationTest(t, func(db *TestDB) {
		repo := NewCustomerRepository(db.Pool)
		
		t.Run("CRUD operations", func(t *testing.T) {
			// Create
			customer := &models.Customer{
				Customer:       "Integration Test Oil Co",
				BillingAddress: "123 Integration St",
				BillingCity:    "Houston",
				BillingState:   "TX",
				BillingZipcode: "77001",
				Contact:        "Test Contact",
				Phone:          "555-1234",
				Email:          "test@integration.com",
			}
			
			err := repo.Create(db.ctx, customer)
			require.NoError(t, err)
			assert.Greater(t, customer.CustomerID, 0)
			assert.False(t, customer.CreatedAt.IsZero())
			
			// Read
			retrieved, err := repo.GetByID(db.ctx, customer.CustomerID)
			require.NoError(t, err)
			AssertCustomersEqual(t, customer, retrieved)
			
			// Update
			retrieved.BillingCity = "Dallas"
			retrieved.Phone = "555-5678"
			err = repo.Update(db.ctx, retrieved)
			require.NoError(t, err)
			
			updated, err := repo.GetByID(db.ctx, customer.CustomerID)
			require.NoError(t, err)
			assert.Equal(t, "Dallas", updated.BillingCity)
			assert.Equal(t, "555-5678", updated.Phone)
			
			// Delete
			err = repo.Delete(db.ctx, customer.CustomerID)
			require.NoError(t, err)
			
			deleted, err := repo.GetByID(db.ctx, customer.CustomerID)
			assert.Error(t, err)
			assert.Nil(t, deleted)
		})
		
		t.Run("duplicate name validation", func(t *testing.T) {
			customer1 := &models.Customer{Customer: "Duplicate Test Company"}
			customer2 := &models.Customer{Customer: "Duplicate Test Company"}
			
			err := repo.Create(db.ctx, customer1)
			require.NoError(t, err)
			
			exists, err := repo.ExistsByName(db.ctx, "Duplicate Test Company")
			require.NoError(t, err)
			assert.True(t, exists)
			
			// Should not exist when excluding this ID
			exists, err = repo.ExistsByName(db.ctx, "Duplicate Test Company", customer1.CustomerID)
			require.NoError(t, err)
			assert.False(t, exists)
			
			// Clean up
			repo.Delete(db.ctx, customer1.CustomerID)
		})
		
		t.Run("search functionality", func(t *testing.T) {
			// Create test customers
			customers := []*models.Customer{
				{Customer: "Search Test Oil Company"},
				{Customer: "Search Test Drilling"},
				{Customer: "Other Company"},
			}
			
			for _, c := range customers {
				err := repo.Create(db.ctx, c)
				require.NoError(t, err)
			}
			
			// Search for "Search Test"
			results, total, err := repo.Search(db.ctx, "Search Test", 10, 0)
			require.NoError(t, err)
			assert.Equal(t, 2, total)
			assert.Len(t, results, 2)
			
			// Clean up
			for _, c := range customers {
				repo.Delete(db.ctx, c.CustomerID)
			}
		})
	})
}

// TestInventoryRepository_Integration tests inventory operations
func TestInventoryRepository_Integration(t *testing.T) {
	RunIntegrationTest(t, func(db *TestDB) {
		customerIDs := db.SeedCustomers(t)
		repo := NewInventoryRepository(db.Pool)
		
		t.Run("create and retrieve inventory", func(t *testing.T) {
			item := &models.InventoryItem{
				CustomerID: customerIDs[0],
				Customer:   "Test Oil Company",
				Joints:     100,
				Size:       "5 1/2\"",
				Weight:     "20",
				Grade:      "J55",
				Connection: "LTC",
				Color:      "RED",
				Location:   "Test Location",
			}
			
			err := repo.Create(db.ctx, item)
			require.NoError(t, err)
			assert.Greater(t, item.ID, 0)
			
			retrieved, err := repo.GetByID(db.ctx, item.ID)
			require.NoError(t, err)
			AssertInventoryEqual(t, item, retrieved)
		})
		
		t.Run("filtering and pagination", func(t *testing.T) {
			itemIDs := db.SeedInventory(t, customerIDs)
			require.Len(t, itemIDs, 3)
			
			// Test filtering by customer
			filters := InventoryFilters{
				CustomerID: &customerIDs[0],
				Page:       1,
				PerPage:    10,
			}
			filters.NormalizePagination()
			
			items, pagination, err := repo.GetFiltered(db.ctx, filters)
			require.NoError(t, err)
			assert.Greater(t, len(items), 0)
			assert.NotNil(t, pagination)
			assert.Equal(t, 1, pagination.Page)
			
			// All items should belong to the filtered customer
			for _, item := range items {
				assert.Equal(t, customerIDs[0], item.CustomerID)
			}
		})
		
		t.Run("search functionality", func(t *testing.T) {
			db.SeedInventory(t, customerIDs)
			
			// Search for "J55" grade
			results, total, err := repo.Search(db.ctx, "J55", 10, 0)
			require.NoError(t, err)
			assert.Greater(t, total, 0)
			
			// All results should contain J55
			for _, item := range results {
				assert.Contains(t, item.Grade, "J55")
			}
		})
		
		t.Run("get summary", func(t *testing.T) {
			db.SeedInventory(t, customerIDs)
			
			summary, err := repo.GetSummary(db.ctx)
			require.NoError(t, err)
			assert.Greater(t, summary.TotalItems, 0)
			assert.Greater(t, summary.TotalJoints, 0)
			assert.Greater(t, len(summary.ItemsByGrade), 0)
			assert.Greater(t, len(summary.ItemsByCustomer), 0)
		})
	})
}

// TestRepositoryPerformance benchmarks repository operations
func TestRepositoryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}
	
	RunIntegrationTest(t, func(db *TestDB) {
		customerIDs := db.SeedCustomers(t)
		db.SeedInventory(t, customerIDs)
		
		repo := NewInventoryRepository(db.Pool)
		
		// Benchmark search operations
		t.Run("search performance", func(t *testing.T) {
			start := time.Now()
			
			for i := 0; i < 100; i++ {
				_, _, err := repo.Search(db.ctx, "J55", 10, 0)
				require.NoError(t, err)
			}
			
			duration := time.Since(start)
			t.Logf("100 searches took: %v (avg: %v per search)", duration, duration/100)
			
			// Should complete within reasonable time
			assert.Less(t, duration.Milliseconds(), int64(5000), "Search operations too slow")
		})
		
		t.Run("filtered query performance", func(t *testing.T) {
			filters := InventoryFilters{
				Grade:   "J55",
				Page:    1,
				PerPage: 50,
			}
			filters.NormalizePagination()
			
			start := time.Now()
			
			for i := 0; i < 50; i++ {
				_, _, err := repo.GetFiltered(db.ctx, filters)
				require.NoError(t, err)
			}
			
			duration := time.Since(start)
			t.Logf("50 filtered queries took: %v (avg: %v per query)", duration, duration/50)
			
			assert.Less(t, duration.Milliseconds(), int64(2500), "Filtered queries too slow")
		})
	})
}

// internal/repository/postgres/inventory_test.go
package postgres

import (
	"context"
	"testing"
	"time"

	"oilgas-backend/internal/repository"
)

func TestInventoryRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	// Create inventory table
	schema := `
	CREATE TABLE IF NOT EXISTS store.inventory (
		id SERIAL PRIMARY KEY,
		work_order VARCHAR(100),
		customer_id INTEGER,
		customer VARCHAR(255),
		joints INTEGER,
		size VARCHAR(50),
		weight DECIMAL(10,2),
		grade VARCHAR(10),
		connection VARCHAR(100),
		date_in DATE,
		date_out DATE,
		well_in VARCHAR(255),
		lease_in VARCHAR(255),
		well_out VARCHAR(255),
		lease_out VARCHAR(255),
		location VARCHAR(100),
		notes TEXT,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create inventory table: %v", err)
	}
	
	repo := NewInventoryRepository(db)
	ctx := context.Background()
	
	item := &repository.InventoryItem{
		WorkOrder:  stringPtr("LB-001000"),
		CustomerID: intPtr(1),
		Customer:   stringPtr("Test Oil Company"),
		Joints:     intPtr(100),
		Size:       stringPtr("5 1/2\""),
		Weight:     float64Ptr(2500.50),
		Grade:      stringPtr("L80"),
		Connection: stringPtr("BTC"),
		DateIn:     timePtr(time.Now()),
		Location:   stringPtr("Yard-A"),
		Notes:      stringPtr("Test inventory item"),
	}
	
	err := repo.Create(ctx, item)
	if err != nil {
		t.Fatalf("Failed to create inventory item: %v", err)
	}
	
	if item.ID == 0 {
		t.Error("Inventory ID should be set after creation")
	}
	
	if item.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set after creation")
	}
}

func TestInventoryRepo_GetByWorkOrder(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	// Create inventory table
	schema := `
	CREATE TABLE IF NOT EXISTS store.inventory (
		id SERIAL PRIMARY KEY,
		work_order VARCHAR(100),
		customer_id INTEGER,
		customer VARCHAR(255),
		joints INTEGER,
		size VARCHAR(50),
		weight DECIMAL(10,2),
		grade VARCHAR(10),
		connection VARCHAR(100),
		date_in DATE,
		date_out DATE,
		well_in VARCHAR(255),
		lease_in VARCHAR(255),
		well_out VARCHAR(255),
		lease_out VARCHAR(255),
		location VARCHAR(100),
		notes TEXT,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create inventory table: %v", err)
	}
	
	repo := NewInventoryRepository(db)
	ctx := context.Background()
	
	// Create test items with same work order
	workOrder := "LB-001000"
	items := []*repository.InventoryItem{
		{
			WorkOrder: stringPtr(workOrder),
			Customer:  stringPtr("Test Oil Company"),
			Joints:    intPtr(50),
			Size:      stringPtr("5 1/2\""),
		},
		{
			WorkOrder: stringPtr(workOrder),
			Customer:  stringPtr("Test Oil Company"),
			Joints:    intPtr(75),
			Size:      stringPtr("7\""),
		},
	}
	
	for _, item := range items {
		err := repo.Create(ctx, item)
		if err != nil {
			t.Fatalf("Failed to create test inventory item: %v", err)
		}
	}
	
	// Test getting items by work order
	results, err := repo.GetByWorkOrder(ctx, workOrder)
	if err != nil {
		t.Fatalf("Failed to get inventory by work order: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 items, got %d", len(results))
	}
	
	for _, item := range results {
		if item.WorkOrder == nil || *item.WorkOrder != workOrder {
			t.Errorf("Expected work order %s, got %v", workOrder, item.WorkOrder)
		}
	}
}

func TestInventoryRepo_GetAvailable(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	// Create inventory table
	schema := `
	CREATE TABLE IF NOT EXISTS store.inventory (
		id SERIAL PRIMARY KEY,
		work_order VARCHAR(100),
		customer_id INTEGER,
		customer VARCHAR(255),
		joints INTEGER,
		size VARCHAR(50),
		weight DECIMAL(10,2),
		grade VARCHAR(10),
		connection VARCHAR(100),
		date_in DATE,
		date_out DATE,
		well_in VARCHAR(255),
		lease_in VARCHAR(255),
		well_out VARCHAR(255),
		lease_out VARCHAR(255),
		location VARCHAR(100),
		notes TEXT,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create inventory table: %v", err)
	}
	
	repo := NewInventoryRepository(db)
	ctx := context.Background()
	
	// Create available item (no date_out)
	availableItem := &repository.InventoryItem{
		WorkOrder: stringPtr("LB-001000"),
		Customer:  stringPtr("Test Oil Company"),
		Joints:    intPtr(100),
		DateIn:    timePtr(time.Now()),
		// DateOut is nil - available
	}
	
	// Create unavailable item (has date_out)
	unavailableItem := &repository.InventoryItem{
		WorkOrder: stringPtr("LB-001001"),
		Customer:  stringPtr("Test Oil Company"),
		Joints:    intPtr(50),
		DateIn:    timePtr(time.Now().AddDate(0, 0, -5)),
		DateOut:   timePtr(time.Now()), // Has date_out - not available
	}
	
	err := repo.Create(ctx, availableItem)
	if err != nil {
		t.Fatalf("Failed to create available item: %v", err)
	}
	
	err = repo.Create(ctx, unavailableItem)
	if err != nil {
		t.Fatalf("Failed to create unavailable item: %v", err)
	}
	
	// Test getting available items
	available, err := repo.GetAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to get available inventory: %v", err)
	}
	
	if len(available) != 1 {
		t.Errorf("Expected 1 available item, got %d", len(available))
	}
	
	if len(available) > 0 {
		if available[0].WorkOrder == nil || *available[0].WorkOrder != "LB-001000" {
			t.Errorf("Expected available item LB-001000, got %v", available[0].WorkOrder)
		}
		if available[0].DateOut != nil {
			t.Error("Available item should not have date_out")
		}
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

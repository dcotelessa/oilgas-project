// internal/repository/postgres/customer_test.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"oilgas-backend/internal/repository"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Load environment variables from project root
	if err := godotenv.Load("../../../../.env.local"); err != nil {
		// Try alternate paths
		if err := godotenv.Load("../../../.env.local"); err != nil {
			t.Logf("Warning: Could not load .env.local: %v", err)
		}
	}
	
	// Use DATABASE_URL from environment, fallback to test database
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:password@localhost:5433/oilgas_inventory_test?sslmode=disable"
		t.Logf("Using fallback test database URL")
	} else {
		t.Logf("Using test database derived from TEST_DATABASE_URL")
	}
	
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	
	// Create test schema
	schema := `
	CREATE SCHEMA IF NOT EXISTS store;
	SET search_path TO store, public;
	
	CREATE TABLE IF NOT EXISTS store.customers (
		customer_id SERIAL PRIMARY KEY,
		customer VARCHAR(255) NOT NULL,
		billing_address TEXT,
		billing_city VARCHAR(100),
		billing_state VARCHAR(50),
		billing_zipcode VARCHAR(20),
		contact VARCHAR(255),
		phone VARCHAR(50),
		fax VARCHAR(50),
		email VARCHAR(255),
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}
	
	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	// Clean up all test data in reverse dependency order
	tables := []string{"inventory", "customers", "grade", "sizes"}
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE store.%s CASCADE", table)); err != nil {
			t.Logf("Warning: Failed to truncate %s: %v", table, err)
		}
	}
	db.Close()
}

func TestCustomerRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	repo := NewCustomerRepository(db)
	ctx := context.Background()
	
	customer := &repository.Customer{
		Customer:       "Test Oil Company",
		BillingAddress: stringPtr("123 Test St"),
		BillingCity:    stringPtr("Houston"),
		BillingState:   stringPtr("TX"),
		Phone:          stringPtr("555-0123"),
		Email:          stringPtr("test@testoil.com"),
	}
	
	err := repo.Create(ctx, customer)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}
	
	if customer.CustomerID == 0 {
		t.Error("Customer ID should be set after creation")
	}
	
	if customer.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set after creation")
	}
}

func TestCustomerRepo_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	repo := NewCustomerRepository(db)
	ctx := context.Background()
	
	// Create test customer
	customer := &repository.Customer{
		Customer: "Test Oil Company",
		Email:    stringPtr("test@testoil.com"),
	}
	err := repo.Create(ctx, customer)
	if err != nil {
		t.Fatalf("Failed to create test customer: %v", err)
	}
	
	// Test getting the customer
	retrieved, err := repo.GetByID(ctx, customer.CustomerID)
	if err != nil {
		t.Fatalf("Failed to get customer: %v", err)
	}
	
	if retrieved == nil {
		t.Fatal("Customer should not be nil")
	}
	
	if retrieved.Customer != customer.Customer {
		t.Errorf("Expected customer name %s, got %s", customer.Customer, retrieved.Customer)
	}
	
	// Test non-existent customer
	nonExistent, err := repo.GetByID(ctx, 99999)
	if err != nil {
		t.Fatalf("Should not error for non-existent customer: %v", err)
	}
	
	if nonExistent != nil {
		t.Error("Non-existent customer should return nil")
	}
}

func TestCustomerRepo_Search(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	repo := NewCustomerRepository(db)
	ctx := context.Background()
	
	// Create test customers
	customers := []*repository.Customer{
		{Customer: "Chevron Corporation", Email: stringPtr("contact@chevron.com")},
		{Customer: "ExxonMobil Corp", Email: stringPtr("info@exxon.com")},
		{Customer: "Shell Oil Company", Email: stringPtr("contact@shell.com")},
	}
	
	for _, c := range customers {
		err := repo.Create(ctx, c)
		if err != nil {
			t.Fatalf("Failed to create test customer: %v", err)
		}
	}
	
	// Test search by company name
	results, err := repo.Search(ctx, "chevron")
	if err != nil {
		t.Fatalf("Failed to search customers: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	
	if len(results) > 0 && results[0].Customer != "Chevron Corporation" {
		t.Errorf("Expected Chevron Corporation, got %s", results[0].Customer)
	}
	
	// Test search by email
	results, err = repo.Search(ctx, "shell.com")
	if err != nil {
		t.Fatalf("Failed to search customers by email: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result for email search, got %d", len(results))
	}
}

func TestCustomerRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	repo := NewCustomerRepository(db)
	ctx := context.Background()
	
	// Create test customer
	customer := &repository.Customer{
		Customer: "Test Oil Company",
		Phone:    stringPtr("555-0123"),
	}
	err := repo.Create(ctx, customer)
	if err != nil {
		t.Fatalf("Failed to create test customer: %v", err)
	}
	
	// Update customer
	customer.Customer = "Updated Oil Company"
	customer.Phone = stringPtr("555-9999")
	
	err = repo.Update(ctx, customer)
	if err != nil {
		t.Fatalf("Failed to update customer: %v", err)
	}
	
	// Verify update
	updated, err := repo.GetByID(ctx, customer.CustomerID)
	if err != nil {
		t.Fatalf("Failed to get updated customer: %v", err)
	}
	
	if updated.Customer != "Updated Oil Company" {
		t.Errorf("Expected updated name, got %s", updated.Customer)
	}
	
	if updated.Phone == nil || *updated.Phone != "555-9999" {
		t.Errorf("Expected updated phone, got %v", updated.Phone)
	}
}

func TestCustomerRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	repo := NewCustomerRepository(db)
	ctx := context.Background()
	
	// Create test customer
	customer := &repository.Customer{
		Customer: "Test Oil Company",
	}
	err := repo.Create(ctx, customer)
	if err != nil {
		t.Fatalf("Failed to create test customer: %v", err)
	}
	
	// Delete customer
	err = repo.Delete(ctx, customer.CustomerID)
	if err != nil {
		t.Fatalf("Failed to delete customer: %v", err)
	}
	
	// Verify deletion (soft delete)
	deleted, err := repo.GetByID(ctx, customer.CustomerID)
	if err != nil {
		t.Fatalf("Should not error when getting deleted customer: %v", err)
	}
	
	if deleted != nil {
		t.Error("Deleted customer should not be returned")
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

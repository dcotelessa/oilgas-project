// backend/internal/repository/test_utils.go
package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"oilgas-backend/internal/models"
)

// TestDB represents a test database connection
type TestDB struct {
	Pool *pgxpool.Pool
	ctx  context.Context
}

// NewTestDB creates a test database connection
// Requires TEST_DATABASE_URL environment variable
func NewTestDB(t *testing.T) *TestDB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	// Test connection
	err = pool.Ping(context.Background())
	require.NoError(t, err)

	return &TestDB{
		Pool: pool,
		ctx:  context.Background(),
	}
}

// Close closes the test database connection
func (db *TestDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Truncate cleans all tables for a fresh test environment
func (db *TestDB) Truncate(t *testing.T) {
	tables := []string{
		"store.inventory",
		"store.received", 
		"store.fletcher",
		"store.bakeout",
		"store.inspected",
		"store.temp",
		"store.tempinv",
		"store.swgc",
		"store.customers",
		"store.users",
	}

	for _, table := range tables {
		_, err := db.Pool.Exec(db.ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(t, err)
	}
}

// SeedCustomers creates test customers and returns their IDs
func (db *TestDB) SeedCustomers(t *testing.T) []int {
	customers := []*models.Customer{
		{
			Customer:       "Test Oil Company",
			BillingAddress: "123 Test Street",
			BillingCity:    "Houston",
			BillingState:   "TX",
			BillingZipcode: "77001",
			Contact:        "John Test",
			Phone:          "555-0123",
			Email:          "john@testoil.com",
		},
		{
			Customer:       "Demo Drilling Services",
			BillingAddress: "456 Demo Avenue",
			BillingCity:    "Dallas", 
			BillingState:   "TX",
			BillingZipcode: "75001",
			Contact:        "Jane Demo",
			Phone:          "555-0456",
			Email:          "jane@demodrill.com",
		},
		{
			Customer:       "Sample Energy Corp",
			BillingAddress: "789 Sample Boulevard",
			BillingCity:    "Austin",
			BillingState:   "TX", 
			BillingZipcode: "78701",
			Contact:        "Bob Sample",
			Phone:          "555-0789",
			Email:          "bob@sampleenergy.com",
		},
	}

	repo := NewCustomerRepository(db.Pool)
	var customerIDs []int

	for _, customer := range customers {
		err := repo.Create(db.ctx, customer)
		require.NoError(t, err)
		customerIDs = append(customerIDs, customer.CustomerID)
	}

	return customerIDs
}

// SeedInventory creates test inventory items
func (db *TestDB) SeedInventory(t *testing.T, customerIDs []int) []int {
	if len(customerIDs) == 0 {
		customerIDs = db.SeedCustomers(t)
	}

	// First, ensure grades exist
	db.SeedGrades(t)

	items := []*models.InventoryItem{
		{
			CustomerID: customerIDs[0],
			Customer:   "Test Oil Company",
			Joints:     100,
			Size:       "5 1/2\"",
			Weight:     "20",
			Grade:      "J55",
			Connection: "LTC",
			Color:      "RED",
			Location:   "North Yard",
			DateIn:     timePtr(time.Now().AddDate(0, 0, -5)),
		},
		{
			CustomerID: customerIDs[1],
			Customer:   "Demo Drilling Services", 
			Joints:     150,
			Size:       "7\"",
			Weight:     "26",
			Grade:      "L80",
			Connection: "BTC",
			Color:      "BLUE",
			Location:   "South Yard",
			DateIn:     timePtr(time.Now().AddDate(0, 0, -3)),
		},
		{
			CustomerID: customerIDs[2],
			Customer:   "Sample Energy Corp",
			Joints:     75,
			Size:       "9 5/8\"", 
			Weight:     "40",
			Grade:      "P110",
			Connection: "PREMIUM",
			Color:      "GREEN",
			Location:   "Premium Storage",
			DateIn:     timePtr(time.Now().AddDate(0, 0, -1)),
		},
	}

	repo := NewInventoryRepository(db.Pool)
	var itemIDs []int

	for _, item := range items {
		err := repo.Create(db.ctx, item)
		require.NoError(t, err)
		itemIDs = append(itemIDs, item.ID)
	}

	return itemIDs
}

// SeedGrades ensures standard grades exist
func (db *TestDB) SeedGrades(t *testing.T) {
	grades := []string{"J55", "JZ55", "L80", "N80", "P105", "P110", "Q125"}
	
	for _, gradeName := range grades {
		_, err := db.Pool.Exec(db.ctx, 
			"INSERT INTO store.grade (grade) VALUES ($1) ON CONFLICT (grade) DO NOTHING", 
			gradeName)
		require.NoError(t, err)
	}
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

// AssertCustomersEqual compares two customers for testing
func AssertCustomersEqual(t *testing.T, expected, actual *models.Customer) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	
	assert.Equal(t, expected.CustomerID, actual.CustomerID)
	assert.Equal(t, expected.Customer, actual.Customer)
	assert.Equal(t, expected.BillingAddress, actual.BillingAddress)
	assert.Equal(t, expected.BillingCity, actual.BillingCity)
	assert.Equal(t, expected.BillingState, actual.BillingState)
	assert.Equal(t, expected.BillingZipcode, actual.BillingZipcode)
	assert.Equal(t, expected.Contact, actual.Contact)
	assert.Equal(t, expected.Phone, actual.Phone)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Equal(t, expected.Deleted, actual.Deleted)
}

// AssertInventoryEqual compares two inventory items for testing
func AssertInventoryEqual(t *testing.T, expected, actual *models.InventoryItem) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.CustomerID, actual.CustomerID)
	assert.Equal(t, expected.Customer, actual.Customer)
	assert.Equal(t, expected.Joints, actual.Joints)
	assert.Equal(t, expected.Size, actual.Size)
	assert.Equal(t, expected.Weight, actual.Weight)
	assert.Equal(t, expected.Grade, actual.Grade)
	assert.Equal(t, expected.Connection, actual.Connection)
	assert.Equal(t, expected.Color, actual.Color)
	assert.Equal(t, expected.Location, actual.Location)
	assert.Equal(t, expected.Deleted, actual.Deleted)
}

// Integration test example
func RunIntegrationTest(t *testing.T, testFunc func(*TestDB)) {
	db := NewTestDB(t)
	defer db.Close()
	
	// Clean slate for each test
	db.Truncate(t)
	
	// Run the test
	testFunc(db)
}

// BenchmarkRepository provides a template for repository benchmarks
func BenchmarkRepository(b *testing.B, setupFunc func() *pgxpool.Pool, testFunc func(*pgxpool.Pool)) {
	pool := setupFunc()
	defer pool.Close()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		testFunc(pool)
	}
}

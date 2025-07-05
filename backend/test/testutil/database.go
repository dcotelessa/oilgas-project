// backend/test/testutil/database.go - Updated for your structure

package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func SetupTestDB(t testing.TB) *pgxpool.Pool {
	dbURL := "postgres://postgres:postgres@localhost:5432/oilgas_inventory_test"
	
	db, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Run test schema setup
	setupTestSchema(t, db)
	
	return db
}

func CleanupTestDB(t testing.TB, db *pgxpool.Pool) {
	// Clean all test data
	cleanupQueries := []string{
		"TRUNCATE TABLE store.inventory CASCADE",
		"TRUNCATE TABLE store.received CASCADE", 
		"TRUNCATE TABLE store.customers CASCADE",
		"TRUNCATE TABLE store.grade CASCADE",
		"TRUNCATE TABLE store.workflow_states CASCADE",
	}

	for _, query := range cleanupQueries {
		_, err := db.Exec(context.Background(), query)
		if err != nil {
			t.Logf("Cleanup warning: %v", err)
		}
	}
	
	db.Close()
}

func setupTestSchema(t testing.TB, db *pgxpool.Pool) {
	schema := `
		CREATE SCHEMA IF NOT EXISTS store;
		
		-- Your essential tables for testing
		CREATE TABLE IF NOT EXISTS store.customers (
			customer_id SERIAL PRIMARY KEY,
			customer VARCHAR(255) NOT NULL,
			billing_address VARCHAR(255),
			billing_city VARCHAR(100),
			billing_state VARCHAR(50),
			billing_zipcode VARCHAR(10),
			phone VARCHAR(20),
			email VARCHAR(255),
			deleted BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Add other essential tables...
		-- Use your actual migration files for complete schema
	`
	
	_, err := db.Exec(context.Background(), schema)
	require.NoError(t, err, "Failed to setup test schema")
}

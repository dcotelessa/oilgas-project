// backend/test/testutil/database.go
package testutil

import (
	"context"
	"testing"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// TestDB wraps the database connection for testing
type TestDB struct {
	Pool *pgxpool.Pool
	ctx  context.Context
}

// SetupTestDB creates and returns a TestDB instance
func SetupTestDB(t testing.TB) *TestDB {
	dbURL := "postgres://postgres:postgres@localhost:5432/oilgas_inventory_test"
	
	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err, "Failed to connect to test database")
	
	db := &TestDB{
		Pool: pool,
		ctx:  context.Background(),
	}
	
	// Run test schema setup
	db.setupTestSchema(t)
	
	return db
}

// CleanupTestDB closes the database connection
func CleanupTestDB(t testing.TB, db *TestDB) {
	if db != nil && db.Pool != nil {
		db.Pool.Close()
	}
}

// Truncate cleans all test data from tables
func (db *TestDB) Truncate(t testing.TB) {
	cleanupQueries := []string{
		"TRUNCATE TABLE store.workflow_state_history CASCADE",
		"TRUNCATE TABLE store.workflow_states CASCADE",
		"TRUNCATE TABLE store.inventory CASCADE",
		"TRUNCATE TABLE store.inspected CASCADE",
		"TRUNCATE TABLE store.received CASCADE", 
		"TRUNCATE TABLE store.fletcher CASCADE",
		"TRUNCATE TABLE store.bakeout CASCADE",
		"TRUNCATE TABLE store.temp CASCADE",
		"TRUNCATE TABLE store.tempinv CASCADE",
		"TRUNCATE TABLE store.swgc CASCADE",
		"TRUNCATE TABLE store.customers CASCADE",
		"TRUNCATE TABLE store.users CASCADE",
		"TRUNCATE TABLE store.grades CASCADE",
	}
	
	for _, query := range cleanupQueries {
		_, err := db.Pool.Exec(db.ctx, query)
		if err != nil {
			t.Logf("Cleanup warning for query '%s': %v", query, err)
		}
	}
}

// SeedGrades creates standard test grades
func (db *TestDB) SeedGrades(t testing.TB) {
	grades := []string{"J55", "K55", "L80", "N80", "P110", "Q125"}
	
	for _, grade := range grades {
		_, err := db.Pool.Exec(db.ctx, 
			"INSERT INTO store.grades (grade) VALUES ($1) ON CONFLICT (grade) DO NOTHING", 
			grade)
		if err != nil {
			t.Logf("Warning seeding grade '%s': %v", grade, err)
		}
	}
}

// setupTestSchema creates the essential test schema
func (db *TestDB) setupTestSchema(t testing.TB) {
	schema := `
		CREATE SCHEMA IF NOT EXISTS store;
		
		-- Customers table
		CREATE TABLE IF NOT EXISTS store.customers (
			customer_id SERIAL PRIMARY KEY,
			customer VARCHAR(255) NOT NULL UNIQUE,
			billing_address VARCHAR(255),
			billing_city VARCHAR(100),
			billing_state VARCHAR(50),
			billing_zipcode VARCHAR(10),
			contact VARCHAR(255),
			phone VARCHAR(20),
			fax VARCHAR(20),
			email VARCHAR(255),
			-- Color fields for customer-specific marking
			color1 VARCHAR(50),
			color2 VARCHAR(50),
			color3 VARCHAR(50),
			color4 VARCHAR(50),
			color5 VARCHAR(50),
			-- Loss fields
			loss1 DECIMAL(5,2),
			loss2 DECIMAL(5,2),
			loss3 DECIMAL(5,2),
			loss4 DECIMAL(5,2),
			loss5 DECIMAL(5,2),
			-- WS (Weight String) color fields
			ws_color1 VARCHAR(50),
			ws_color2 VARCHAR(50),
			ws_color3 VARCHAR(50),
			ws_color4 VARCHAR(50),
			ws_color5 VARCHAR(50),
			-- WS Loss fields
			ws_loss1 DECIMAL(5,2),
			ws_loss2 DECIMAL(5,2),
			ws_loss3 DECIMAL(5,2),
			ws_loss4 DECIMAL(5,2),
			ws_loss5 DECIMAL(5,2),
			deleted BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);

		-- Grades table
		CREATE TABLE IF NOT EXISTS store.grades (
			grade VARCHAR(10) PRIMARY KEY
		);

		-- Received items table
		CREATE TABLE IF NOT EXISTS store.received (
			id SERIAL PRIMARY KEY,
			work_order VARCHAR(50) NOT NULL UNIQUE,
			customer_id INTEGER REFERENCES store.customers(customer_id),
			customer VARCHAR(255) NOT NULL,
			joints INTEGER NOT NULL,
			rack VARCHAR(50),
			size_id INTEGER,
			size VARCHAR(50),
			weight VARCHAR(50),
			grade VARCHAR(50) REFERENCES store.grades(grade),
			connection VARCHAR(50),
			ctd BOOLEAN DEFAULT false,
			w_string BOOLEAN DEFAULT false,
			well VARCHAR(255),
			lease VARCHAR(255),
			ordered_by VARCHAR(255),
			notes TEXT,
			customer_po VARCHAR(100),
			date_received TIMESTAMP,
			background VARCHAR(100),
			norm VARCHAR(100),
			services VARCHAR(255),
			bill_to_id VARCHAR(50),
			entered_by VARCHAR(100),
			when_entered TIMESTAMP DEFAULT NOW(),
			when_updated TIMESTAMP,
			trucking VARCHAR(255),
			trailer VARCHAR(255),
			in_production TIMESTAMP,
			inspected_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Inspected items table
		CREATE TABLE IF NOT EXISTS store.inspected (
			id SERIAL PRIMARY KEY,
			work_order VARCHAR(50) REFERENCES store.received(work_order),
			customer_id INTEGER REFERENCES store.customers(customer_id),
			customer VARCHAR(255) NOT NULL,
			joints INTEGER NOT NULL,
			size VARCHAR(50),
			weight VARCHAR(50),
			grade VARCHAR(50),
			connection VARCHAR(50),
			passed_joints INTEGER DEFAULT 0,
			failed_joints INTEGER DEFAULT 0,
			inspection_date TIMESTAMP,
			inspector VARCHAR(255),
			notes TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);

		-- Inventory table
		CREATE TABLE IF NOT EXISTS store.inventory (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER REFERENCES store.customers(customer_id),
			customer VARCHAR(255) NOT NULL,
			joints INTEGER NOT NULL,
			size VARCHAR(50),
			weight VARCHAR(50),
			grade VARCHAR(50),
			connection VARCHAR(50),
			color VARCHAR(50),
			location VARCHAR(255),
			ctd BOOLEAN DEFAULT false,
			w_string BOOLEAN DEFAULT false,
			date_in TIMESTAMP,
			deleted BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Workflow states table
		CREATE TABLE IF NOT EXISTS store.workflow_states (
			id SERIAL PRIMARY KEY,
			work_order VARCHAR(50) NOT NULL UNIQUE REFERENCES store.received(work_order),
			current_state VARCHAR(50) NOT NULL,
			updated_at TIMESTAMP DEFAULT NOW()
		);

		-- Workflow state history table
		CREATE TABLE IF NOT EXISTS store.workflow_state_history (
			id SERIAL PRIMARY KEY,
			work_order VARCHAR(50) NOT NULL REFERENCES store.received(work_order),
			from_state VARCHAR(50),
			to_state VARCHAR(50) NOT NULL,
			reason TEXT,
			changed_by VARCHAR(100),
			changed_at TIMESTAMP DEFAULT NOW()
		);

		-- Users table (if needed for tests)
		CREATE TABLE IF NOT EXISTS store.users (
			user_id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			access INTEGER DEFAULT 0,
			full_name VARCHAR(255),
			email VARCHAR(255),
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Add other tables as needed for your tests...
		-- Fletcher, Bakeout, SWGC, etc.
	`
	
	_, err := db.Pool.Exec(db.ctx, schema)
	require.NoError(t, err, "Failed to setup test schema")
}

// Legacy function for backward compatibility
func SetupTestDBLegacy(t testing.TB) *pgxpool.Pool {
	db := SetupTestDB(t)
	return db.Pool
}

// Legacy cleanup for backward compatibility  
func CleanupTestDBLegacy(t testing.TB, pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

// backend/test/testutil/database.go

package testutil

import (
	"context"
	"os"
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
// Reads connection string from environment variables
func SetupTestDB(t testing.TB) *TestDB {
	// Try to get database URL from environment variables
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	
	// If still not found, use default test configuration
	if dbURL == "" {
		dbURL = "postgres://postgres:test123@localhost:5433/oilgas_inventory_test"
		t.Logf("No TEST_DATABASE_URL or DATABASE_URL found, using default: %s", dbURL)
	} else {
		t.Logf("Using database URL from environment: %s", dbURL)
	}
	
	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err, "Failed to connect to test database")
	
	// Test the connection
	err = pool.Ping(context.Background())
	require.NoError(t, err, "Failed to ping test database")
	
	db := &TestDB{
		Pool: pool,
		ctx:  context.Background(),
	}
	
	// Run test schema setup - MATCHES YOUR MIGRATION EXACTLY
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
	// Order matters for foreign key constraints - child tables first
	cleanupQueries := []string{
		// Child tables with foreign keys first
		"TRUNCATE TABLE store.inventory CASCADE",
		"TRUNCATE TABLE store.received CASCADE", 
		"TRUNCATE TABLE store.fletcher CASCADE",
		"TRUNCATE TABLE store.bakeout CASCADE",
		"TRUNCATE TABLE store.inspected CASCADE",
		"TRUNCATE TABLE store.swgc CASCADE",
		"TRUNCATE TABLE store.temp CASCADE",
		"TRUNCATE TABLE store.tempinv CASCADE",
		
		// Parent tables after child tables
		"TRUNCATE TABLE store.customers CASCADE",
		"TRUNCATE TABLE store.users CASCADE",
		"TRUNCATE TABLE store.grade CASCADE",  // Fixed: singular table name
		
		// Utility tables (no foreign keys)
		"TRUNCATE TABLE store.test CASCADE",
		"TRUNCATE TABLE store.r_number CASCADE",
		"TRUNCATE TABLE store.wk_number CASCADE",
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
	// Use the same grades as in your migration - FIXED TABLE NAME
	grades := []string{"J55", "JZ55", "K55", "L80", "N80", "P105", "P110", "Q125", "T95", "C90", "C95", "S135"}
	
	for _, grade := range grades {
		_, err := db.Pool.Exec(db.ctx, 
			"INSERT INTO store.grade (grade) VALUES ($1) ON CONFLICT (grade) DO NOTHING", 
			grade)
		if err != nil {
			t.Logf("Warning seeding grade '%s': %v", grade, err)
		}
	}
}

// setupTestSchema creates the COMPLETE test schema - MATCHES YOUR MIGRATIONS EXACTLY
func (db *TestDB) setupTestSchema(t testing.TB) {
	schema := `
		-- Schema matching migrations: 001_initial_schema + 002_enhanced_indexes + 003_add_customer_fields_to_inspected
		CREATE SCHEMA IF NOT EXISTS store;
		SET search_path TO store, public;

		-- Table: customers (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.customers (
			customer_id SERIAL PRIMARY KEY,
			customer VARCHAR(50),
			billing_address VARCHAR(50),
			billing_city VARCHAR(50),
			billing_state VARCHAR(50),
			billing_zipcode VARCHAR(50),
			contact VARCHAR(50),
			phone VARCHAR(50),
			fax VARCHAR(50),
			email VARCHAR(50),
			color1 VARCHAR(50),
			color2 VARCHAR(50),
			color3 VARCHAR(50),
			color4 VARCHAR(50),
			color5 VARCHAR(50),
			loss1 VARCHAR(50),
			loss2 VARCHAR(50),
			loss3 VARCHAR(50),
			loss4 VARCHAR(50),
			loss5 VARCHAR(50),
			wscolor1 VARCHAR(50),
			wscolor2 VARCHAR(50),
			wscolor3 VARCHAR(50),
			wscolor4 VARCHAR(50),
			wscolor5 VARCHAR(50),
			wsloss1 VARCHAR(50),
			wsloss2 VARCHAR(50),
			wsloss3 VARCHAR(50),
			wsloss4 VARCHAR(50),
			wsloss5 VARCHAR(50),
			deleted BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: inventory (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.inventory (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50),
			work_order VARCHAR(50),
			r_number INTEGER,
			customer_id INTEGER REFERENCES store.customers(customer_id),
			customer VARCHAR(50),
			joints INTEGER,
			rack VARCHAR(50),
			size VARCHAR(50),
			weight VARCHAR(50),
			grade VARCHAR(50),
			connection VARCHAR(50),
			ctd BOOLEAN NOT NULL DEFAULT false,
			w_string BOOLEAN NOT NULL DEFAULT false,
			swgcc VARCHAR(50),
			color VARCHAR(50),
			customer_po VARCHAR(50),
			fletcher VARCHAR(50),
			date_in TIMESTAMP,
			date_out TIMESTAMP,
			well_in VARCHAR(50),
			lease_in VARCHAR(50),
			well_out VARCHAR(50),
			lease_out VARCHAR(50),
			trucking VARCHAR(50),
			trailer VARCHAR(50),
			location VARCHAR(50),
			notes TEXT,
			pcode VARCHAR(50),
			cn INTEGER,
			ordered_by VARCHAR(50),
			deleted BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: received (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.received (
			id SERIAL PRIMARY KEY,
			work_order VARCHAR(50),
			customer_id INTEGER REFERENCES store.customers(customer_id),
			customer VARCHAR(50),
			joints INTEGER,
			rack VARCHAR(50),
			size_id INTEGER,
			size VARCHAR(50),
			weight VARCHAR(50),
			grade VARCHAR(50),
			connection VARCHAR(50),
			ctd BOOLEAN NOT NULL DEFAULT false,
			w_string BOOLEAN NOT NULL DEFAULT false,
			well VARCHAR(50),
			lease VARCHAR(50),
			ordered_by VARCHAR(50),
			notes TEXT,
			customer_po VARCHAR(50),
			date_received TIMESTAMP,
			background VARCHAR(50),
			norm VARCHAR(50),
			services VARCHAR(50),
			bill_to_id VARCHAR(50),
			entered_by VARCHAR(50),
			when_entered TIMESTAMP,
			trucking VARCHAR(50),
			trailer VARCHAR(50),
			in_production TIMESTAMP,
			inspected_date TIMESTAMP,
			threading_date TIMESTAMP,
			straighten_required BOOLEAN NOT NULL DEFAULT false,
			excess_material BOOLEAN NOT NULL DEFAULT false,
			complete BOOLEAN NOT NULL DEFAULT false,
			inspected_by VARCHAR(50),
			updated_by VARCHAR(50),
			when_updated TIMESTAMP,
			deleted BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: fletcher (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.fletcher (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50),
			fletcher VARCHAR(50),
			r_number INTEGER,
			customer_id INTEGER REFERENCES store.customers(customer_id),
			customer VARCHAR(50),
			joints INTEGER,
			size VARCHAR(50),
			weight VARCHAR(50),
			grade VARCHAR(50),
			connection VARCHAR(50),
			ctd BOOLEAN NOT NULL DEFAULT false,
			w_string BOOLEAN NOT NULL DEFAULT false,
			swgcc VARCHAR(50),
			color VARCHAR(50),
			customer_po VARCHAR(50),
			date_in TIMESTAMP,
			date_out TIMESTAMP,
			well_in VARCHAR(50),
			lease_in VARCHAR(50),
			well_out VARCHAR(50),
			lease_out VARCHAR(50),
			trucking VARCHAR(50),
			trailer VARCHAR(50),
			location VARCHAR(50),
			notes TEXT,
			pcode VARCHAR(50),
			cn INTEGER,
			ordered_by VARCHAR(50),
			deleted BOOLEAN NOT NULL DEFAULT false,
			complete BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: bakeout (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.bakeout (
			id SERIAL PRIMARY KEY,
			fletcher VARCHAR(50),
			joints INTEGER,
			color VARCHAR(50),
			size VARCHAR(50),
			weight VARCHAR(50),
			grade VARCHAR(50),
			connection VARCHAR(50),
			ctd BOOLEAN NOT NULL DEFAULT false,
			swgcc VARCHAR(50),
			customer_id INTEGER REFERENCES store.customers(customer_id),
			accept INTEGER,
			reject INTEGER,
			pin INTEGER,
			cplg INTEGER,
			pc INTEGER,
			trucking VARCHAR(50),
			trailer VARCHAR(50),
			date_in TIMESTAMP,
			cn INTEGER,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: inspected (ENHANCED - from 001_initial_schema.sql + 003_add_customer_fields_to_inspected.sql)
		CREATE TABLE IF NOT EXISTS store.inspected (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50),
			work_order VARCHAR(50),
			color VARCHAR(50),
			joints INTEGER,
			accept INTEGER,
			reject INTEGER,
			pin INTEGER,
			cplg INTEGER,
			pc INTEGER,
			complete BOOLEAN NOT NULL DEFAULT false,
			rack VARCHAR(50),
			rep_pin INTEGER,
			rep_cplg INTEGER,
			rep_pc INTEGER,
			deleted BOOLEAN NOT NULL DEFAULT false,
			cn INTEGER,
			created_at TIMESTAMP DEFAULT NOW(),
			-- ADDED BY 003 MIGRATION
			customer_id INTEGER REFERENCES store.customers(customer_id),
			customer VARCHAR(50),
			grade VARCHAR(50),
			size VARCHAR(50),
			weight VARCHAR(50),
			connection VARCHAR(50),
			inspector VARCHAR(50),
			inspection_date TIMESTAMP,
			passed_joints INTEGER DEFAULT 0,
			failed_joints INTEGER DEFAULT 0,
			notes TEXT
		);

		-- Table: grade (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.grade (
			grade VARCHAR(50) PRIMARY KEY
		);

		-- Table: swgc (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.swgc (
			size_id INTEGER,
			customer_id INTEGER REFERENCES store.customers(customer_id),
			size VARCHAR(50),
			weight VARCHAR(50),
			connection VARCHAR(50),
			pcode_receive VARCHAR(50),
			pcode_inventory VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: temp (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.temp (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50),
			work_order VARCHAR(50),
			color VARCHAR(50),
			joints INTEGER,
			accept INTEGER,
			reject INTEGER,
			pin INTEGER,
			cplg INTEGER,
			pc INTEGER,
			rack VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: tempinv (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.tempinv (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50),
			work_order VARCHAR(50),
			customer_id INTEGER REFERENCES store.customers(customer_id),
			customer VARCHAR(50),
			joints INTEGER,
			rack VARCHAR(50),
			size VARCHAR(50),
			weight VARCHAR(50),
			grade VARCHAR(50),
			connection VARCHAR(50),
			ctd BOOLEAN NOT NULL DEFAULT false,
			w_string BOOLEAN NOT NULL DEFAULT false,
			swgcc VARCHAR(50),
			color VARCHAR(50),
			customer_po VARCHAR(50),
			fletcher VARCHAR(50),
			date_in TIMESTAMP,
			date_out TIMESTAMP,
			well_in VARCHAR(50),
			lease_in VARCHAR(50),
			well_out VARCHAR(50),
			lease_out VARCHAR(50),
			trucking VARCHAR(50),
			trailer VARCHAR(50),
			location VARCHAR(50),
			notes TEXT,
			pcode VARCHAR(50),
			cn INTEGER,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: test (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.test (
			id SERIAL PRIMARY KEY,
			test VARCHAR(50)
		);

		-- Table: users (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.users (
			user_id SERIAL PRIMARY KEY,
			username VARCHAR(12) UNIQUE,
			password VARCHAR(255),
			access INTEGER,
			full_name VARCHAR(50),
			email VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- Table: r_number (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.r_number (
			r_number INTEGER PRIMARY KEY
		);

		-- Table: wk_number (from 001_initial_schema.sql)
		CREATE TABLE IF NOT EXISTS store.wk_number (
			wk_number INTEGER PRIMARY KEY
		);

		-- Insert standard oil & gas grades (from 001_initial_schema.sql)
		INSERT INTO store.grade (grade) VALUES 
		('J55'), ('JZ55'), ('K55'), ('L80'), ('N80'), 
		('P105'), ('P110'), ('Q125'), ('T95'), ('C90'), ('C95'), ('S135')
		ON CONFLICT (grade) DO NOTHING;

		-- Basic indexes from 001_initial_schema.sql
		CREATE INDEX IF NOT EXISTS idx_customers_customer ON store.customers(customer);
		CREATE INDEX IF NOT EXISTS idx_customers_deleted ON store.customers(deleted) WHERE deleted = false;
		CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
		CREATE INDEX IF NOT EXISTS idx_inventory_grade ON store.inventory(grade);
		CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
		CREATE INDEX IF NOT EXISTS idx_inventory_deleted ON store.inventory(deleted) WHERE deleted = false;
		CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
		CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received);
		CREATE INDEX IF NOT EXISTS idx_received_deleted ON store.received(deleted) WHERE deleted = false;
		CREATE INDEX IF NOT EXISTS idx_fletcher_customer_id ON store.fletcher(customer_id);
		CREATE INDEX IF NOT EXISTS idx_users_username ON store.users(username);

		-- Additional indexes from 003 migration for inspected table
		CREATE INDEX IF NOT EXISTS idx_inspected_customer_id ON store.inspected(customer_id);
		CREATE INDEX IF NOT EXISTS idx_inspected_customer_grade ON store.inspected(customer_id, grade) WHERE customer_id IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_inspected_inspection_date ON store.inspected(inspection_date DESC) WHERE inspection_date IS NOT NULL;

		-- Update statistics
		ANALYZE;
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

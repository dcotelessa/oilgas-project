// backend/tenant_migrator.go
// Extends existing migrator.go for multi-tenant database management
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type TenantMigrator struct {
	*Migrator // Embed existing migrator
	TenantID  string
	TenantDB  *sql.DB
}

func NewTenantMigrator(tenantID string) (*TenantMigrator, error) {
	baseMigrator, err := NewMigrator("local")
	if err != nil {
		return nil, err
	}

	return &TenantMigrator{
		Migrator: baseMigrator,
		TenantID: tenantID,
	}, nil
}

func (tm *TenantMigrator) CreateTenantDatabase() error {
	dbName := fmt.Sprintf("oilgas_%s", tm.TenantID)
	
	log.Printf("Creating tenant database: %s", dbName)

	// Check if database already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)`
	err := tm.db.QueryRow(checkQuery, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		log.Printf("Database %s already exists, skipping creation", dbName)
	} else {
		// Create the database
		createQuery := fmt.Sprintf(`CREATE DATABASE "%s" WITH OWNER = current_user ENCODING = 'UTF8'`, dbName)
		if _, err := tm.db.Exec(createQuery); err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		log.Printf("✅ Database created: %s", dbName)
	}

	// Connect to the new tenant database
	if err := tm.connectToTenantDB(dbName); err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	// Run schema migrations on tenant database
	if err := tm.createTenantSchema(); err != nil {
		return fmt.Errorf("failed to create tenant schema: %w", err)
	}

	log.Printf("✅ Tenant database ready: %s", dbName)
	return nil
}

func (tm *TenantMigrator) connectToTenantDB(dbName string) error {
	// Get connection string from environment and modify for tenant DB
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}

	// Replace database name in connection string
	// Assumes format: postgres://user:password@host:port/database
	parts := strings.Split(databaseURL, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid DATABASE_URL format")
	}

	// Replace the database name
	parts[len(parts)-1] = dbName
	tenantURL := strings.Join(parts, "/")

	var err error
	tm.TenantDB, err = sql.Open("postgres", tenantURL)
	if err != nil {
		return fmt.Errorf("failed to open tenant database connection: %w", err)
	}

	if err := tm.TenantDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping tenant database: %w", err)
	}

	return nil
}

func (tm *TenantMigrator) createTenantSchema() error {
	log.Printf("Creating schema for tenant: %s", tm.TenantID)

	// Use the same schema creation logic from the base migrator
	// but with tenant-specific enhancements
	schema := tm.getTenantSchema()

	if _, err := tm.TenantDB.Exec(schema); err != nil {
		return fmt.Errorf("failed to create tenant schema: %w", err)
	}

	// Insert tenant-specific seed data
	if err := tm.createTenantSeeds(); err != nil {
		return fmt.Errorf("failed to create tenant seeds: %w", err)
	}

	log.Printf("✅ Schema created for tenant: %s", tm.TenantID)
	return nil
}

func (tm *TenantMigrator) getTenantSchema() string {
	return `
	-- Create schemas
	CREATE SCHEMA IF NOT EXISTS store;
	CREATE SCHEMA IF NOT EXISTS migrations;
	
	-- Set search path
	SET search_path TO store, public;
	
	-- Create migrations table
	CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Enhanced customers table with tenant isolation
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
		tenant_id VARCHAR(50) NOT NULL DEFAULT '` + tm.TenantID + `',
		imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Grades table (oil & gas industry standards)
	CREATE TABLE IF NOT EXISTS store.grade (
		grade VARCHAR(10) PRIMARY KEY,
		description TEXT
	);
	
	-- Sizes table
	CREATE TABLE IF NOT EXISTS store.sizes (
		size_id SERIAL PRIMARY KEY,
		size VARCHAR(50) NOT NULL UNIQUE,
		description TEXT
	);
	
	-- Enhanced inventory table with tenant isolation
	CREATE TABLE IF NOT EXISTS store.inventory (
		id SERIAL PRIMARY KEY,
		username VARCHAR(100),
		work_order VARCHAR(100),
		r_number VARCHAR(100),
		customer_id INTEGER REFERENCES store.customers(customer_id),
		customer VARCHAR(255),
		joints INTEGER,
		rack VARCHAR(50),
		size VARCHAR(50),
		weight DECIMAL(10,2),
		grade VARCHAR(10) REFERENCES store.grade(grade),
		connection VARCHAR(100),
		ctd VARCHAR(100),
		w_string VARCHAR(100),
		swgcc VARCHAR(100),
		color VARCHAR(50),
		customer_po VARCHAR(100),
		fletcher VARCHAR(100),
		date_in DATE,
		date_out DATE,
		well_in VARCHAR(255),
		lease_in VARCHAR(255),
		well_out VARCHAR(255),
		lease_out VARCHAR(255),
		trucking VARCHAR(100),
		trailer VARCHAR(100),
		location VARCHAR(100),
		notes TEXT,
		pcode VARCHAR(50),
		cn VARCHAR(50),
		ordered_by VARCHAR(100),
		tenant_id VARCHAR(50) NOT NULL DEFAULT '` + tm.TenantID + `',
		imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Enhanced received table with tenant isolation
	CREATE TABLE IF NOT EXISTS store.received (
		id SERIAL PRIMARY KEY,
		work_order VARCHAR(100),
		customer_id INTEGER REFERENCES store.customers(customer_id),
		customer VARCHAR(255),
		joints INTEGER,
		rack VARCHAR(50),
		size_id INTEGER REFERENCES store.sizes(size_id),
		size VARCHAR(50),
		weight DECIMAL(10,2),
		grade VARCHAR(10) REFERENCES store.grade(grade),
		connection VARCHAR(100),
		ctd VARCHAR(100),
		w_string VARCHAR(100),
		well VARCHAR(255),
		lease VARCHAR(255),
		ordered_by VARCHAR(100),
		notes TEXT,
		customer_po VARCHAR(100),
		date_received DATE,
		background TEXT,
		norm VARCHAR(100),
		services TEXT,
		bill_to_id INTEGER,
		entered_by VARCHAR(100),
		when_entered TIMESTAMP,
		trucking VARCHAR(100),
		trailer VARCHAR(100),
		in_production BOOLEAN DEFAULT FALSE,
		inspected_date DATE,
		threading_date DATE,
		straighten_required BOOLEAN DEFAULT FALSE,
		excess_material BOOLEAN DEFAULT FALSE,
		complete BOOLEAN DEFAULT FALSE,
		inspected_by VARCHAR(100),
		updated_by VARCHAR(100),
		when_updated TIMESTAMP,
		tenant_id VARCHAR(50) NOT NULL DEFAULT '` + tm.TenantID + `',
		imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Create indexes for performance
	CREATE INDEX IF NOT EXISTS idx_customers_tenant ON store.customers(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_customers_customer_id ON store.customers(customer_id);
	CREATE INDEX IF NOT EXISTS idx_customers_search ON store.customers USING gin(to_tsvector('english', customer));
	
	CREATE INDEX IF NOT EXISTS idx_inventory_tenant ON store.inventory(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
	CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
	CREATE INDEX IF NOT EXISTS idx_inventory_date_in ON store.inventory(date_in);
	CREATE INDEX IF NOT EXISTS idx_inventory_search ON store.inventory USING gin(to_tsvector('english', customer || ' ' || COALESCE(work_order, '') || ' ' || COALESCE(notes, '')));
	
	CREATE INDEX IF NOT EXISTS idx_received_tenant ON store.received(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
	CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order);
	CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received);
	
	-- Tenant isolation views for easier querying
	CREATE OR REPLACE VIEW store.tenant_customers AS
	SELECT * FROM store.customers WHERE tenant_id = '` + tm.TenantID + `' AND NOT deleted;
	
	CREATE OR REPLACE VIEW store.tenant_inventory AS
	SELECT * FROM store.inventory WHERE tenant_id = '` + tm.TenantID + `' AND NOT deleted;
	
	CREATE OR REPLACE VIEW store.tenant_received AS
	SELECT * FROM store.received WHERE tenant_id = '` + tm.TenantID + `' AND NOT deleted;
	`
}

func (tm *TenantMigrator) createTenantSeeds() error {
	seeds := `
	-- Set search path
	SET search_path TO store, public;
	
	-- Insert oil & gas industry standard grades
	INSERT INTO store.grade (grade, description) VALUES 
	('J55', 'Standard grade steel casing - most common'),
	('JZ55', 'Enhanced J55 grade with improved properties'),
	('L80', 'Higher strength grade for moderate environments'),
	('N80', 'Medium strength grade for standard applications'),
	('P105', 'High performance grade for demanding conditions'),
	('P110', 'Premium performance grade for extreme environments'),
	('Q125', 'Ultra-high strength grade for specialized applications'),
	('C75', 'Carbon steel grade for basic applications'),
	('C95', 'Higher carbon steel grade'),
	('T95', 'Tough grade for harsh environments')
	ON CONFLICT (grade) DO NOTHING;
	
	-- Insert common pipe sizes
	INSERT INTO store.sizes (size, description) VALUES 
	('4 1/2"', '4.5 inch diameter - small casing'),
	('5"', '5 inch diameter - intermediate casing'),
	('5 1/2"', '5.5 inch diameter - common production casing'),
	('7"', '7 inch diameter - intermediate casing'),
	('8 5/8"', '8.625 inch diameter - surface casing'),
	('9 5/8"', '9.625 inch diameter - surface casing'),
	('10 3/4"', '10.75 inch diameter - surface casing'),
	('13 3/8"', '13.375 inch diameter - surface casing'),
	('16"', '16 inch diameter - conductor casing'),
	('18 5/8"', '18.625 inch diameter - conductor casing'),
	('20"', '20 inch diameter - large conductor casing'),
	('24"', '24 inch diameter - extra large conductor'),
	('30"', '30 inch diameter - structural casing')
	ON CONFLICT (size) DO NOTHING;
	`

	if _, err := tm.TenantDB.Exec(seeds); err != nil {
		return fmt.Errorf("failed to create tenant seeds: %w", err)
	}

	return nil
}

func (tm *TenantMigrator) ImportCSVData(csvDir string) error {
	log.Printf("Importing CSV data for tenant: %s from %s", tm.TenantID, csvDir)

	// Check if CSV directory exists
	if _, err := os.Stat(csvDir); os.IsNotExist(err) {
		return fmt.Errorf("CSV directory not found: %s", csvDir)
	}

	// Look for import script first
	importScript := filepath.Join(csvDir, "import_script.sql")
	if _, err := os.Stat(importScript); err == nil {
		return tm.runImportScript(importScript)
	}

	// Fallback: process CSV files individually
	return tm.importCSVFiles(csvDir)
}

func (tm *TenantMigrator) runImportScript(scriptPath string) error {
	log.Printf("Running import script: %s", scriptPath)

	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read import script: %w", err)
	}

	// Execute the script
	if _, err := tm.TenantDB.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute import script: %w", err)
	}

	log.Printf("✅ Import script executed successfully")
	return nil
}

func (tm *TenantMigrator) importCSVFiles(csvDir string) error {
	entries, err := os.ReadDir(csvDir)
	if err != nil {
		return fmt.Errorf("failed to read CSV directory: %w", err)
	}

	importCount := 0
	for _, entry := range entries {
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			continue
		}

		csvFile := filepath.Join(csvDir, entry.Name())
		tableName := strings.TrimSuffix(strings.ToLower(entry.Name()), ".csv")
		
		if err := tm.importSingleCSV(csvFile, tableName); err != nil {
			log.Printf("⚠️  Failed to import %s: %v", entry.Name(), err)
			continue
		}

		importCount++
		log.Printf("✅ Imported: %s", entry.Name())
	}

	if importCount == 0 {
		return fmt.Errorf("no CSV files successfully imported")
	}

	log.Printf("✅ Import complete: %d files imported", importCount)
	return nil
}

func (tm *TenantMigrator) importSingleCSV(csvFile, tableName string) error {
	// Map CSV table names to actual database tables
	dbTable := mapCSVToDBTable(tableName)
	if dbTable == "" {
		return fmt.Errorf("unknown table mapping for: %s", tableName)
	}

	// Use PostgreSQL COPY command for fast import
	copyCmd := fmt.Sprintf(`COPY store.%s FROM '%s' WITH CSV HEADER DELIMITER ','`, dbTable, csvFile)
	
	_, err := tm.TenantDB.Exec(copyCmd)
	return err
}

func (tm *TenantMigrator) GetTenantStatus() error {
	if tm.TenantDB == nil {
		dbName := fmt.Sprintf("oilgas_%s", tm.TenantID)
		if err := tm.connectToTenantDB(dbName); err != nil {
			return fmt.Errorf("tenant database not accessible: %w", err)
		}
	}

	fmt.Printf("\n=== Tenant Status: %s ===\n", tm.TenantID)
	fmt.Printf("Database: oilgas_%s\n", tm.TenantID)

	// Check table row counts
	tables := []string{"customers", "inventory", "received", "grade", "sizes"}
	fmt.Println("\n📊 Table Status:")
	
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM store.%s", table)
		if table == "customers" || table == "inventory" || table == "received" {
			query += fmt.Sprintf(" WHERE tenant_id = '%s'", tm.TenantID)
		}
		
		err := tm.TenantDB.QueryRow(query).Scan(&count)
		if err != nil {
			fmt.Printf("  ❌ %s: Error checking table\n", table)
		} else {
			fmt.Printf("  ✅ %s: %d records\n", table, count)
		}
	}

	// Check recent imports
	var lastImport sql.NullTime
	err := tm.TenantDB.QueryRow("SELECT MAX(imported_at) FROM store.customers WHERE tenant_id = $1", tm.TenantID).Scan(&lastImport)
	if err == nil && lastImport.Valid {
		fmt.Printf("\n📅 Last Import: %s\n", lastImport.Time.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n✅ Tenant status check complete")
	return nil
}

func (tm *TenantMigrator) DropTenantDatabase() error {
	dbName := fmt.Sprintf("oilgas_%s", tm.TenantID)
	
	log.Printf("⚠️  Dropping tenant database: %s", dbName)

	// Close tenant connection if open
	if tm.TenantDB != nil {
		tm.TenantDB.Close()
		tm.TenantDB = nil
	}

	// Drop the database
	dropQuery := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, dbName)
	if _, err := tm.db.Exec(dropQuery); err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	log.Printf("✅ Tenant database dropped: %s", dbName)
	return nil
}

func mapCSVToDBTable(csvName string) string {
	mapping := map[string]string{
		"customers":  "customers",
		"customer":   "customers",
		"custid":     "customers",
		"inventory":  "inventory",
		"received":   "received",
		"recv":       "received",
		"workorder":  "inventory",
		"workorders": "inventory",
		"grades":     "grade",
		"grade":      "grade",
		"sizes":      "sizes",
		"size":       "sizes",
	}

	if table, exists := mapping[strings.ToLower(csvName)]; exists {
		return table
	}

	return ""
}

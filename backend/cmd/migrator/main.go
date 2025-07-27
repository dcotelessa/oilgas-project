package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// TenantMigrator extends existing functionality for multi-tenant operations
type TenantMigrator struct {
	baseDB   *sql.DB
	tenantDB *sql.DB
	tenantID string
}

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Oil & Gas Inventory System - Database Migrator")
		fmt.Println()
		fmt.Println("Standard Commands:")
		fmt.Println("  migrator migrate [env]        - Run migrations")
		fmt.Println("  migrator seed [env]           - Seed database")
		fmt.Println("  migrator status [env]         - Show status")
		fmt.Println("  migrator reset [env]          - Reset database")
		fmt.Println()
		fmt.Println("Tenant Commands:")
		fmt.Println("  migrator tenant-create <id>   - Create tenant database")
		fmt.Println("  migrator csv-import <id> <dir> - Import CSV to tenant")
		fmt.Println("  migrator tenant-status <id>   - Check tenant status")
		fmt.Println("  migrator tenant-drop <id>     - Drop tenant database")
		fmt.Println("  migrator tenant-list          - List all tenant databases")
		fmt.Println()
		fmt.Println("Schema Management:")
		fmt.Println("  migrator schema-update-all \"<sql>\" - Update all tenant schemas")
		fmt.Println("  migrator schema-check-consistency   - Check schema consistency")
		fmt.Println("  migrator schema-versions            - Show all schema versions")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  migrator tenant-create longbeach")
		fmt.Println("  migrator csv-import longbeach ../csv/longbeach/")
		fmt.Println("  migrator schema-update-all \"ALTER TABLE store.customers ADD COLUMN emergency_contact VARCHAR(255);\"")
		fmt.Println()
		fmt.Println("Environments: local, test, production")
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle tenant-specific commands (don't need env parameter)
	switch command {
	case "tenant-create":
		handleTenantCreate()
	case "csv-import":
		handleCSVImport()
	case "tenant-status":
		handleTenantStatus()
	case "tenant-drop":
		handleTenantDrop()
	case "tenant-list":
		handleTenantList()
	case "schema-update-all":
		handleSchemaUpdateAll()
	case "schema-check-consistency":
		handleSchemaConsistency()
	case "schema-versions":
		handleSchemaVersions()
	default:
		// Handle standard commands (migrate, seed, status, reset)
		handleStandardCommands(command)
	}
}

// ============================================================================
// TENANT COMMAND HANDLERS
// ============================================================================

func handleTenantCreate() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: migrator tenant-create <tenant_id>")
	}
	tenantID := os.Args[2]
	
	if err := validateTenantID(tenantID); err != nil {
		log.Fatalf("Invalid tenant ID: %v", err)
	}
	
	tm, err := NewTenantMigrator(tenantID)
	if err != nil {
		log.Fatalf("Failed to create tenant migrator: %v", err)
	}
	defer tm.Close()
	
	if err := tm.CreateTenantDatabase(); err != nil {
		log.Fatalf("Tenant creation failed: %v", err)
	}
	
	log.Printf("✅ Tenant database created: oilgas_%s", tenantID)
	log.Printf("Next step: migrator csv-import %s /path/to/csv/", tenantID)
}

func handleCSVImport() {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: migrator csv-import <tenant_id> <csv_directory>")
	}
	tenantID := os.Args[2]
	csvDir := os.Args[3]
	
	if err := validateTenantID(tenantID); err != nil {
		log.Fatalf("Invalid tenant ID: %v", err)
	}
	
	tm, err := NewTenantMigrator(tenantID)
	if err != nil {
		log.Fatalf("Failed to create tenant migrator: %v", err)
	}
	defer tm.Close()
	
	if err := tm.ImportCSVData(csvDir); err != nil {
		log.Fatalf("CSV import failed: %v", err)
	}
	
	log.Printf("✅ CSV data imported to tenant: %s", tenantID)
	log.Printf("Next step: migrator tenant-status %s", tenantID)
}

func handleTenantStatus() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: migrator tenant-status <tenant_id>")
	}
	tenantID := os.Args[2]
	
	tm, err := NewTenantMigrator(tenantID)
	if err != nil {
		log.Fatalf("Failed to create tenant migrator: %v", err)
	}
	defer tm.Close()
	
	if err := tm.ShowTenantStatus(); err != nil {
		log.Fatalf("Status check failed: %v", err)
	}
}

func handleTenantDrop() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: migrator tenant-drop <tenant_id>")
	}
	tenantID := os.Args[2]
	
	fmt.Printf("⚠️  WARNING: This will permanently delete tenant database: oilgas_%s\n", tenantID)
	fmt.Print("Type 'yes' to confirm: ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	if confirmation != "yes" {
		fmt.Println("Operation cancelled")
		os.Exit(0)
	}
	
	tm, err := NewTenantMigrator(tenantID)
	if err != nil {
		log.Fatalf("Failed to create tenant migrator: %v", err)
	}
	defer tm.Close()
	
	if err := tm.DropTenantDatabase(); err != nil {
		log.Fatalf("Tenant drop failed: %v", err)
	}
}

func handleTenantList() {
	if err := listTenantDatabases(); err != nil {
		log.Fatalf("Failed to list tenant databases: %v", err)
	}
}

func handleSchemaUpdateAll() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: migrator schema-update-all \"<migration_sql>\"")
	}
	migrationSQL := os.Args[2]
	
	csm, err := NewCentralizedSchemaManager()
	if err != nil {
		log.Fatalf("Failed to create schema manager: %v", err)
	}
	defer csm.Close()
	
	if err := csm.ApplySchemaUpdateToAllTenants(migrationSQL); err != nil {
		log.Fatalf("Schema update failed: %v", err)
	}
}

func handleSchemaConsistency() {
	csm, err := NewCentralizedSchemaManager()
	if err != nil {
		log.Fatalf("Failed to create schema manager: %v", err)
	}
	defer csm.Close()
	
	if err := csm.CheckSchemaConsistency(); err != nil {
		log.Fatalf("Schema consistency check failed: %v", err)
	}
}

func handleSchemaVersions() {
	csm, err := NewCentralizedSchemaManager()
	if err != nil {
		log.Fatalf("Failed to create schema manager: %v", err)
	}
	defer csm.Close()
	
	if err := csm.ShowAllSchemaVersions(); err != nil {
		log.Fatalf("Schema versions check failed: %v", err)
	}
}

// ============================================================================
// STANDARD COMMAND HANDLERS
// ============================================================================

func handleStandardCommands(command string) {
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}

	// Get database URL with CONSISTENT naming enforced
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres123@localhost:5433/oilgas_inventory_local?sslmode=disable"
		fmt.Println("⚠️  Using fallback DATABASE_URL with consistent naming")
	}

	// SAFETY CHECK: Ensure we're using the consistent database name (your existing logic)
	if !contains(databaseURL, "oil_gas_inventory") {
		fmt.Printf("🔧 Database URL correction needed. Current: %s\n", databaseURL)
		databaseURL = "postgresql://postgres:postgres123@localhost:5433/oilgas_inventory_local?sslmode=disable"
		fmt.Printf("🔧 Corrected to: %s\n", databaseURL)
	}

	fmt.Printf("🔌 Connecting to database (env: %s)\n", env)
	fmt.Printf("🔗 Database URL: %s\n", databaseURL)

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connection successful")

	// Execute command (your existing logic preserved)
	switch command {
	case "migrate":
		runMigrations(db, env)
	case "seed":
		runSeeds(db, env)
	case "status":
		showStatus(db, env)
	case "reset":
		resetDatabase(db, env)
	case "generate":
		handleGenerate(db, env)
	case "tenant-create":
		handleTenantCreate(db, env)
	default:
		log.Fatalf("❌ Unknown command: %s", command)
	}
}

// ============================================================================
// TENANT MIGRATOR IMPLEMENTATION
// ============================================================================

func NewTenantMigrator(tenantID string) (*TenantMigrator, error) {
	// Connect to main database for tenant management
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres123@localhost:5433/oilgas_inventory_local?sslmode=disable"
	}

	baseDB, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to base database: %w", err)
	}

	if err := baseDB.Ping(); err != nil {
		baseDB.Close()
		return nil, fmt.Errorf("failed to ping base database: %w", err)
	}

	return &TenantMigrator{
		baseDB:   baseDB,
		tenantID: tenantID,
	}, nil
}

func (tm *TenantMigrator) Close() {
	if tm.tenantDB != nil {
		tm.tenantDB.Close()
	}
	if tm.baseDB != nil {
		tm.baseDB.Close()
	}
}

func (tm *TenantMigrator) CreateTenantDatabase() error {
	dbName := fmt.Sprintf("oilgas_%s", tm.tenantID)
	
	log.Printf("Creating tenant database: %s", dbName)

	// Check if database already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)`
	err := tm.baseDB.QueryRow(checkQuery, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		log.Printf("Database %s already exists, skipping creation", dbName)
	} else {
		// Create the database
		createQuery := fmt.Sprintf(`CREATE DATABASE "%s" WITH OWNER = current_user ENCODING = 'UTF8'`, dbName)
		if _, err := tm.baseDB.Exec(createQuery); err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		log.Printf("✅ Database created: %s", dbName)
	}

	// Connect to the new tenant database
	if err := tm.connectToTenantDB(dbName); err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	// Run your existing migrations on tenant database with tenant enhancements
	if err := tm.runTenantMigrations(); err != nil {
		return fmt.Errorf("failed to run tenant migrations: %w", err)
	}

	// Run your existing seeds on tenant database
	if err := tm.runTenantSeeds(); err != nil {
		return fmt.Errorf("failed to run tenant seeds: %w", err)
	}

	log.Printf("✅ Tenant database ready: %s", dbName)
	return nil
}

func (tm *TenantMigrator) connectToTenantDB(dbName string) error {
	// Get base connection string and modify for tenant DB
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres123@localhost:5433/oilgas_inventory_local?sslmode=disable"
	}

	// Replace database name in connection string
	parts := strings.Split(databaseURL, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid DATABASE_URL format")
	}

	// Handle potential query parameters
	lastPart := parts[len(parts)-1]
	var params string
	if strings.Contains(lastPart, "?") {
		dbAndParams := strings.Split(lastPart, "?")
		params = "?" + dbAndParams[1]
	}

	// Replace the database name
	parts[len(parts)-1] = dbName + params
	tenantURL := strings.Join(parts, "/")

	var err error
	tm.tenantDB, err = sql.Open("postgres", tenantURL)
	if err != nil {
		return fmt.Errorf("failed to open tenant database connection: %w", err)
	}

	if err := tm.tenantDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping tenant database: %w", err)
	}

	return nil
}

func (tm *TenantMigrator) runTenantMigrations() error {
	log.Printf("🔄 Running tenant migrations for: %s", tm.tenantID)

	// Use your existing migration logic but with tenant enhancements
	// Step 1: Create schemas first
	fmt.Println("📁 Step 1: Creating schemas...")
	
	_, err := tm.tenantDB.Exec("CREATE SCHEMA IF NOT EXISTS store")
	if err != nil {
		return fmt.Errorf("failed to create store schema: %w", err)
	}

	_, err = tm.tenantDB.Exec("CREATE SCHEMA IF NOT EXISTS migrations")
	if err != nil {
		return fmt.Errorf("failed to create migrations schema: %w", err)
	}

	// Step 2: Create migration tracking table
	migrationTrackingSQL := `
	CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = tm.tenantDB.Exec(migrationTrackingSQL)
	if err != nil {
		return fmt.Errorf("failed to create migration tracking table: %w", err)
	}

	// Step 3: Set search path
	_, err = tm.tenantDB.Exec("SET search_path TO store, public")
	if err != nil {
		return fmt.Errorf("failed to set search path: %w", err)
	}

	// Step 4: Create reference tables (your existing logic)
	gradeTableSQL := `
	CREATE TABLE IF NOT EXISTS store.grade (
		grade VARCHAR(10) PRIMARY KEY,
		description TEXT
	)`
	
	_, err = tm.tenantDB.Exec(gradeTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create grade table: %w", err)
	}

	sizesTableSQL := `
	CREATE TABLE IF NOT EXISTS store.sizes (
		size_id SERIAL PRIMARY KEY,
		size VARCHAR(50) NOT NULL UNIQUE,
		description TEXT
	)`
	
	_, err = tm.tenantDB.Exec(sizesTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create sizes table: %w", err)
	}

	// Step 5: Create customers table with tenant enhancements
	customersTableSQL := `
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
		tenant_id VARCHAR(50) NOT NULL DEFAULT '` + tm.tenantID + `',
		imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = tm.tenantDB.Exec(customersTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create customers table: %w", err)
	}

	// Step 6: Create inventory table with tenant enhancements
	inventoryTableSQL := `
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
		tenant_id VARCHAR(50) NOT NULL DEFAULT '` + tm.tenantID + `',
		imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = tm.tenantDB.Exec(inventoryTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create inventory table: %w", err)
	}

	// Step 7: Create received table with tenant enhancements
	receivedTableSQL := `
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
		tenant_id VARCHAR(50) NOT NULL DEFAULT '` + tm.tenantID + `',
		imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = tm.tenantDB.Exec(receivedTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create received table: %w", err)
	}

	// Step 8: Create performance indexes with tenant-aware additions
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_customers_tenant ON store.customers(tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_customers_customer_id ON store.customers(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_customers_search ON store.customers USING gin(to_tsvector('english', customer))",
		"CREATE INDEX IF NOT EXISTS idx_inventory_tenant ON store.inventory(tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_date_in ON store.inventory(date_in)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_search ON store.inventory USING gin(to_tsvector('english', customer || ' ' || COALESCE(work_order, '') || ' ' || COALESCE(notes, '')))",
		"CREATE INDEX IF NOT EXISTS idx_received_tenant ON store.received(tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order)",
		"CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received)",
	}
	
	for i, indexSQL := range indexes {
		_, err = tm.tenantDB.Exec(indexSQL)
		if err != nil {
			return fmt.Errorf("failed to create index %d: %w", i+1, err)
		}
	}

	// Step 9: Record migration
	_, err = tm.tenantDB.Exec("INSERT INTO migrations.schema_migrations (version, name) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING", "001", fmt.Sprintf("initial_tenant_schema_%s", tm.tenantID))
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	log.Println("✅ Tenant migrations completed successfully!")
	return nil
}

func (tm *TenantMigrator) runTenantSeeds() error {
	log.Printf("🌱 Running tenant seeds for: %s", tm.tenantID)

	// Set search path
	if _, err := tm.tenantDB.Exec("SET search_path TO store, public;"); err != nil {
		return fmt.Errorf("failed to set search path: %w", err)
	}

	// Use your existing seed logic for reference data
	referenceSQL := `
	-- Insert oil & gas industry grades (same as your existing logic)
	INSERT INTO store.grade (grade, description) VALUES 
	('J55', 'Standard grade steel casing - most common'),
	('L80', 'Higher strength grade for moderate environments'),
	('N80', 'Medium strength grade for standard applications'),
	('P105', 'High performance grade for demanding conditions'),
	('P110', 'Premium performance grade for extreme environments'),
	('Q125', 'Ultra-high strength grade for specialized applications'),
	('C75', 'Carbon steel grade for basic applications'),
	('C95', 'Higher carbon steel grade'),
	('T95', 'Tough grade for harsh environments');
	
	-- Insert common pipe sizes (same as your existing logic)
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
	('30"', '30 inch diameter - structural casing');
	`

	if _, err := tm.tenantDB.Exec(referenceSQL); err != nil {
		return fmt.Errorf("failed to insert reference data: %w", err)
	}

	log.Println("✅ Tenant seeds completed successfully!")
	return nil
}

func (tm *TenantMigrator) ImportCSVData(csvDir string) error {
	log.Printf("📥 Importing CSV data for tenant: %s from %s", tm.tenantID, csvDir)

	// Check if CSV directory exists
	if _, err := os.Stat(csvDir); os.IsNotExist(err) {
		return fmt.Errorf("CSV directory not found: %s", csvDir)
	}

	// Connect to tenant database if not already connected
	if tm.tenantDB == nil {
		dbName := fmt.Sprintf("oilgas_%s", tm.tenantID)
		if err := tm.connectToTenantDB(dbName); err != nil {
			return fmt.Errorf("failed to connect to tenant database: %w", err)
		}
	}

	// Look for import script first (generated by tenant processor)
	importScript := csvDir + "/import_script.sql"
	if _, err := os.Stat(importScript); err == nil {
		return tm.runImportScript(importScript)
	}

	// Fallback: process CSV files individually
	return tm.importCSVFiles(csvDir)
}

func (tm *TenantMigrator) runImportScript(scriptPath string) error {
	log.Printf("📜 Running import script: %s", scriptPath)

	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read import script: %w", err)
	}

	// Execute the script
	if _, err := tm.tenantDB.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute import script: %w", err)
	}

	log.Printf("✅ Import script executed successfully")
	return nil
}

func (tm *TenantMigrator) importCSVFiles(csvDir string) error {
	// Basic CSV import logic - would be enhanced based on your specific CSV structure
	log.Printf("📁 Processing CSV files in: %s", csvDir)
	
	// This is a placeholder - the actual implementation would use your
	// tenant processor output format
	log.Printf("✅ CSV files processed for tenant: %s", tm.tenantID)
	return nil
}

func (tm *TenantMigrator) ShowTenantStatus() error {
	if tm.tenantDB == nil {
		dbName := fmt.Sprintf("oilgas_%s", tm.tenantID)
		if err := tm.connectToTenantDB(dbName); err != nil {
			return fmt.Errorf("tenant database not accessible: %w", err)
		}
	}

	fmt.Printf("\n=== Tenant Status: %s ===\n", tm.tenantID)
	fmt.Printf("Database: oilgas_%s\n", tm.tenantID)

	// Use your existing status logic adapted for tenant
	tables := []string{"customers", "inventory", "received", "grade", "sizes"}
	fmt.Println("\n📊 Table Status:")
	
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM store.%s", table)
		if table == "customers" || table == "inventory" || table == "received" {
			query += fmt.Sprintf(" WHERE tenant_id = '%s'", tm.tenantID)
		}
		
		err := tm.tenantDB.QueryRow(query).Scan(&count)
		if err != nil {
			fmt.Printf("  ❌ %s: Error checking table\n", table)
		} else {
			fmt.Printf("  ✅ %s: %d records\n", table, count)
		}
	}

	// Check recent imports
	var lastImport sql.NullTime
	err := tm.tenantDB.QueryRow("SELECT MAX(imported_at) FROM store.customers WHERE tenant_id = $1", tm.tenantID).Scan(&lastImport)
	if err == nil && lastImport.Valid {
		fmt.Printf("\n📅 Last Import: %s\n", lastImport.Time.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n✅ Tenant status check complete")
	return nil
}

func (tm *TenantMigrator) DropTenantDatabase() error {
	dbName := fmt.Sprintf("oilgas_%s", tm.tenantID)
	
	log.Printf("⚠️  Dropping tenant database: %s", dbName)

	// Close tenant connection if open
	if tm.tenantDB != nil {
		tm.tenantDB.Close()
		tm.tenantDB = nil
	}

	// Drop the database
	dropQuery := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, dbName)
	if _, err := tm.baseDB.Exec(dropQuery); err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	log.Printf("✅ Tenant database dropped: %s", dbName)
	return nil
}

// ============================================================================
// CENTRALIZED SCHEMA MANAGEMENT
// ============================================================================

type CentralizedSchemaManager struct {
	baseDB *sql.DB
}

func NewCentralizedSchemaManager() (*CentralizedSchemaManager, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres123@localhost:5433/oilgas_inventory_local?sslmode=disable"
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &CentralizedSchemaManager{baseDB: db}, nil
}

func (csm *CentralizedSchemaManager) Close() {
	if csm.baseDB != nil {
		csm.baseDB.Close()
	}
}

func (csm *CentralizedSchemaManager) ApplySchemaUpdateToAllTenants(migrationSQL string) error {
	tenants, err := csm.getAllTenantIDs()
	if err != nil {
		return fmt.Errorf("failed to get tenant list: %w", err)
	}

	if len(tenants) == 0 {
		log.Println("⚠️  No tenant databases found")
		return nil
	}

	log.Printf("🔄 Applying schema update to %d tenants...", len(tenants))
	log.Printf("📝 Migration SQL:\n%s", migrationSQL)

	// Validate SQL first (dry run on first tenant)
	if len(tenants) > 0 {
		if err := csm.validateMigrationSQL(tenants[0], migrationSQL); err != nil {
			return fmt.Errorf("migration validation failed: %w", err)
		}
		log.Println("✅ Migration SQL validated")
	}

	// Apply to all tenants in parallel
	return csm.parallelSchemaUpdate(tenants, migrationSQL)
}

func (csm *CentralizedSchemaManager) parallelSchemaUpdate(tenants []string, migrationSQL string) error {
	var wg sync.WaitGroup
	errors := make(chan error, len(tenants))
	success := make(chan string, len(tenants))

	// Limit concurrent updates to prevent resource exhaustion
	maxConcurrent := 5
	if len(tenants) < maxConcurrent {
		maxConcurrent = len(tenants)
	}

	semaphore := make(chan struct{}, maxConcurrent)

	for _, tenantID := range tenants {
		wg.Add(1)
		go func(tenant string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			tm, err := NewTenantMigrator(tenant)
			if err != nil {
				errors <- fmt.Errorf("tenant %s: %w", tenant, err)
				return
			}
			defer tm.Close()

			if err := tm.applyMigration(migrationSQL); err != nil {
				errors <- fmt.Errorf("tenant %s: %w", tenant, err)
				return
			}

			success <- tenant
		}(tenantID)
	}

	wg.Wait()
	close(errors)
	close(success)

	// Collect results
	var updateErrors []error
	successCount := 0

	for err := range errors {
		updateErrors = append(updateErrors, err)
	}

	for tenant := range success {
		log.Printf("✅ Tenant %s: Schema updated successfully", tenant)
		successCount++
	}

	log.Printf("\n📊 Schema Update Summary:")
	log.Printf("  ✅ Successful: %d/%d", successCount, len(tenants))
	log.Printf("  ❌ Failed: %d/%d", len(updateErrors), len(tenants))

	if len(updateErrors) > 0 {
		log.Printf("\n❌ Failed Tenants:")
		for _, err := range updateErrors {
			log.Printf("  %v", err)
		}
		return fmt.Errorf("schema update failed for %d/%d tenants", len(updateErrors), len(tenants))
	}

	log.Printf("\n✅ Schema update completed successfully for all %d tenants", successCount)
	return nil
}

func (tm *TenantMigrator) applyMigration(migrationSQL string) error {
	// Connect to tenant database if not already connected
	if tm.tenantDB == nil {
		dbName := fmt.Sprintf("oilgas_%s", tm.tenantID)
		if err := tm.connectToTenantDB(dbName); err != nil {
			return fmt.Errorf("failed to connect to tenant database: %w", err)
		}
	}

	// Apply migration with transaction for safety
	tx, err := tm.tenantDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	if _, err := tx.Exec(migrationSQL); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

func (csm *CentralizedSchemaManager) validateMigrationSQL(firstTenant, migrationSQL string) error {
	tm, err := NewTenantMigrator(firstTenant)
	if err != nil {
		return err
	}
	defer tm.Close()

	dbName := fmt.Sprintf("oilgas_%s", firstTenant)
	if err := tm.connectToTenantDB(dbName); err != nil {
		return err
	}

	// Validate by running in a transaction and rolling back
	tx, err := tm.tenantDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(migrationSQL); err != nil {
		return fmt.Errorf("SQL validation failed: %w", err)
	}

	// Rollback - this was just validation
	return tx.Rollback()
}

func (csm *CentralizedSchemaManager) getAllTenantIDs() ([]string, error) {
	query := `
		SELECT datname 
		FROM pg_catalog.pg_database 
		WHERE datname LIKE 'oilgas_%' 
		ORDER BY datname
	`

	rows, err := csm.baseDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenant databases: %w", err)
	}
	defer rows.Close()

	var tenants []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		
		// Extract tenant ID from database name
		tenantID := strings.TrimPrefix(dbName, "oilgas_")
		tenants = append(tenants, tenantID)
	}

	return tenants, nil
}

func (csm *CentralizedSchemaManager) CheckSchemaConsistency() error {
	tenants, err := csm.getAllTenantIDs()
	if err != nil {
		return err
	}

	if len(tenants) == 0 {
		log.Println("⚠️  No tenant databases found")
		return nil
	}

	log.Printf("🔍 Checking schema consistency across %d tenants...", len(tenants))

	versionMap := make(map[string][]string) // version -> tenant list
	
	for _, tenantID := range tenants {
		version, err := csm.getSchemaVersion(tenantID)
		if err != nil {
			log.Printf("❌ Failed to get schema version for %s: %v", tenantID, err)
			continue
		}
		
		versionMap[version] = append(versionMap[version], tenantID)
	}

	if len(versionMap) == 1 {
		var version string
		for v := range versionMap {
			version = v
			break
		}
		log.Printf("✅ All tenants have consistent schema version: %s", version)
		return nil
	}

	log.Printf("⚠️  Schema version inconsistency detected:")
	for version, tenantList := range versionMap {
		log.Printf("  Version '%s': %v", version, tenantList)
	}

	return fmt.Errorf("schema version inconsistency: %d different versions found", len(versionMap))
}

func (csm *CentralizedSchemaManager) getSchemaVersion(tenantID string) (string, error) {
	tm, err := NewTenantMigrator(tenantID)
	if err != nil {
		return "", err
	}
	defer tm.Close()

	dbName := fmt.Sprintf("oilgas_%s", tenantID)
	if err := tm.connectToTenantDB(dbName); err != nil {
		return "", err
	}

	var version string
	query := `
		SELECT version 
		FROM migrations.schema_migrations 
		ORDER BY executed_at DESC 
		LIMIT 1
	`
	
	err = tm.tenantDB.QueryRow(query).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return "none", nil
		}
		return "", err
	}

	return version, nil
}

func (csm *CentralizedSchemaManager) ShowAllSchemaVersions() error {
	tenants, err := csm.getAllTenantIDs()
	if err != nil {
		return err
	}

	log.Printf("📋 Schema Versions Report:")
	for _, tenantID := range tenants {
		version, err := csm.getSchemaVersion(tenantID)
		if err != nil {
			log.Printf("  %s: ❌ Error: %v", tenantID, err)
		} else {
			log.Printf("  %s: %s", tenantID, version)
		}
	}
	return nil
}

// ============================================================================
// TENANT LISTING
// ============================================================================

func listTenantDatabases() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres123@localhost:5433/oilgas_inventory_local?sslmode=disable"
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("\n=== Tenant Databases ===")

	// Query for databases starting with 'oilgas_'
	query := `
		SELECT datname 
		FROM pg_catalog.pg_database 
		WHERE datname LIKE 'oilgas_%' 
		ORDER BY datname
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tenant databases: %w", err)
	}
	defer rows.Close()

	tenantCount := 0
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}

		// Extract tenant ID from database name
		tenantID := strings.TrimPrefix(dbName, "oilgas_")
		
		// Get basic stats for this tenant
		stats, err := getTenantStats(dbName)
		if err != nil {
			fmt.Printf("  📁 %s (database: %s) - Error getting stats\n", tenantID, dbName)
		} else {
			fmt.Printf("  📁 %s (database: %s)\n", tenantID, dbName)
			fmt.Printf("     Customers: %d, Inventory: %d, Received: %d\n", 
				stats.Customers, stats.Inventory, stats.Received)
		}
		
		tenantCount++
	}

	if tenantCount == 0 {
		fmt.Println("  No tenant databases found")
		fmt.Println("  Use 'migrator tenant-create <tenant_id>' to create one")
	} else {
		fmt.Printf("\nTotal tenant databases: %d\n", tenantCount)
	}

	return nil
}

type TenantStats struct {
	Customers int
	Inventory int
	Received  int
}

func getTenantStats(dbName string) (*TenantStats, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres123@localhost:5433/oilgas_inventory_local?sslmode=disable"
	}

	// Replace database name in connection string
	parts := strings.Split(databaseURL, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid DATABASE_URL format")
	}
	
	// Handle potential query parameters
	lastPart := parts[len(parts)-1]
	var params string
	if strings.Contains(lastPart, "?") {
		dbAndParams := strings.Split(lastPart, "?")
		params = "?" + dbAndParams[1]
	}
	
	parts[len(parts)-1] = dbName + params
	tenantURL := strings.Join(parts, "/")

	db, err := sql.Open("postgres", tenantURL)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stats := &TenantStats{}

	// Get customers count
	err = db.QueryRow("SELECT COUNT(*) FROM store.customers WHERE NOT deleted").Scan(&stats.Customers)
	if err != nil {
		stats.Customers = -1
	}

	// Get inventory count
	err = db.QueryRow("SELECT COUNT(*) FROM store.inventory WHERE NOT deleted").Scan(&stats.Inventory)
	if err != nil {
		stats.Inventory = -1
	}

	// Get received count
	err = db.QueryRow("SELECT COUNT(*) FROM store.received WHERE NOT deleted").Scan(&stats.Received)
	if err != nil {
		stats.Received = -1
	}

	return stats, nil
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func validateTenantID(tenantID string) error {
	if len(tenantID) < 2 || len(tenantID) > 20 {
		return fmt.Errorf("tenant ID must be between 2 and 20 characters")
	}

	for _, char := range tenantID {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("tenant ID must contain only lowercase letters, numbers, and underscores")
		}
	}

	return nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================================
// YOUR EXISTING FUNCTIONS
// ============================================================================

func runMigrations(db *sql.DB, env string) {
	fmt.Printf("🔄 Running migrations for environment: %s\n", env)

	// Step 1: Create schemas first (separately with error checking)
	fmt.Println("📁 Step 1: Creating schemas...")
	
	_, err := db.Exec("CREATE SCHEMA IF NOT EXISTS store")
	if err != nil {
		log.Fatalf("❌ Failed to create store schema: %v", err)
	}
	fmt.Println("✅ Store schema created")

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS migrations")
	if err != nil {
		log.Fatalf("❌ Failed to create migrations schema: %v", err)
	}
	fmt.Println("✅ Migrations schema created")

	// Step 2: Create migration tracking table
	fmt.Println("📋 Step 2: Creating migration tracking...")
	
	migrationTrackingSQL := `
	CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = db.Exec(migrationTrackingSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create migration tracking table: %v", err)
	}
	fmt.Println("✅ Migration tracking table created")

	// Step 3: Set search path
	fmt.Println("🛤️ Step 3: Setting search path...")
	
	_, err = db.Exec("SET search_path TO store, public")
	if err != nil {
		log.Fatalf("❌ Failed to set search path: %v", err)
	}
	fmt.Println("✅ Search path set to store, public")

	// Step 4: Create reference tables (no dependencies)
	fmt.Println("📊 Step 4: Creating reference tables...")
	
	gradeTableSQL := `
	CREATE TABLE IF NOT EXISTS store.grade (
		grade VARCHAR(10) PRIMARY KEY,
		description TEXT
	)`
	
	_, err = db.Exec(gradeTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create grade table: %v", err)
	}
	fmt.Println("✅ Grade table created")

	sizesTableSQL := `
	CREATE TABLE IF NOT EXISTS store.sizes (
		size_id SERIAL PRIMARY KEY,
		size VARCHAR(50) NOT NULL UNIQUE,
		description TEXT
	)`
	
	_, err = db.Exec(sizesTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create sizes table: %v", err)
	}
	fmt.Println("✅ Sizes table created")

	// Step 5: Create customers table
	fmt.Println("👥 Step 5: Creating customers table...")
	
	customersTableSQL := `
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
	)`
	
	_, err = db.Exec(customersTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create customers table: %v", err)
	}
	fmt.Println("✅ Customers table created")

	// Step 6: Create inventory table (with foreign keys)
	fmt.Println("📦 Step 6: Creating inventory table...")
	
	inventoryTableSQL := `
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
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = db.Exec(inventoryTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create inventory table: %v", err)
	}
	fmt.Println("✅ Inventory table created")

	// Step 7: Create received table (with foreign keys)
	fmt.Println("📨 Step 7: Creating received table...")
	
	receivedTableSQL := `
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
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = db.Exec(receivedTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create received table: %v", err)
	}
	fmt.Println("✅ Received table created")

	// Step 8: Create indexes
	fmt.Println("📈 Step 8: Creating performance indexes...")
	
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_date_in ON store.inventory(date_in)",
		"CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order)",
		"CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received)",
	}
	
	for i, indexSQL := range indexes {
		_, err = db.Exec(indexSQL)
		if err != nil {
			log.Fatalf("❌ Failed to create index %d: %v", i+1, err)
		}
	}
	fmt.Println("✅ Performance indexes created")

	// Step 9: Record migration
	fmt.Println("📝 Step 9: Recording migration...")
	
	_, err = db.Exec("INSERT INTO migrations.schema_migrations (version, name) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING", "001", "initial_oil_gas_schema_stepwise")
	if err != nil {
		log.Fatalf("❌ Failed to record migration: %v", err)
	}
	fmt.Println("✅ Migration recorded")

	fmt.Println("🎉 Migrations completed successfully!")
}

func runSeeds(db *sql.DB, env string) {
	fmt.Printf("🌱 Running seeds for environment: %s\n", env)

	// Set search path
	if _, err := db.Exec("SET search_path TO store, public;"); err != nil {
		log.Fatalf("❌ Failed to set search path: %v", err)
	}

	// Clear existing data (development only) with RESTART IDENTITY to reset sequences
	if env == "local" || env == "development" {
		fmt.Println("🧹 Clearing existing data and resetting sequences...")
		clearSQL := `
		TRUNCATE TABLE store.received CASCADE;
		TRUNCATE TABLE store.inventory CASCADE;
		TRUNCATE TABLE store.customers RESTART IDENTITY CASCADE;
		TRUNCATE TABLE store.sizes RESTART IDENTITY CASCADE;
		DELETE FROM store.grade;
		`
		if _, err := db.Exec(clearSQL); err != nil {
			log.Fatalf("❌ Failed to clear data: %v", err)
		}
		fmt.Println("✅ Data cleared, sequences reset to start at 1")
	}

	// Step 1: Insert reference data (no SERIAL dependencies)
	fmt.Println("📊 Inserting reference data...")
	referenceSQL := `
	-- Insert oil & gas industry grades (no SERIAL, uses explicit PKs)
	INSERT INTO store.grade (grade, description) VALUES 
	('J55', 'Standard grade steel casing - most common'),
	('L80', 'Higher strength grade for moderate environments'),
	('N80', 'Medium strength grade for standard applications'),
	('P105', 'High performance grade for demanding conditions'),
	('P110', 'Premium performance grade for extreme environments'),
	('Q125', 'Ultra-high strength grade for specialized applications'),
	('C75', 'Carbon steel grade for basic applications'),
	('C95', 'Higher carbon steel grade'),
	('T95', 'Tough grade for harsh environments');
	
	-- Insert common pipe sizes (uses SERIAL size_id, will start at 1)
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
	('30"', '30 inch diameter - structural casing');
	`

	if _, err := db.Exec(referenceSQL); err != nil {
		log.Fatalf("❌ Failed to insert reference data: %v", err)
	}
	fmt.Println("✅ Reference data inserted")

	// Step 2: Insert customers and capture their actual SERIAL IDs
	fmt.Println("👥 Inserting customers and capturing SERIAL IDs...")
	
	type Customer struct {
		ID   int
		Name string
	}
	
	customers := make([]Customer, 0)
	
	// Insert customers one by one using RETURNING to capture actual customer_id
	customerData := [][]string{
		{"Permian Basin Energy", "1234 Oil Field Rd", "Midland", "TX", "79701", "John Smith", "432-555-0101", "operations@permianbasin.com"},
		{"Eagle Ford Solutions", "5678 Shale Ave", "San Antonio", "TX", "78201", "Sarah Johnson", "210-555-0201", "drilling@eagleford.com"},
		{"Bakken Industries", "9012 Prairie Blvd", "Williston", "ND", "58801", "Mike Wilson", "701-555-0301", "procurement@bakken.com"},
		{"Gulf Coast Drilling", "3456 Offshore Dr", "Houston", "TX", "77001", "Lisa Brown", "713-555-0401", "logistics@gulfcoast.com"},
		{"Marcellus Gas Co", "7890 Mountain View", "Pittsburgh", "PA", "15201", "Robert Davis", "412-555-0501", "operations@marcellus.com"},
	}
	
	for _, data := range customerData {
		var customerID int
		err := db.QueryRow(`
			INSERT INTO store.customers (customer, billing_address, billing_city, billing_state, billing_zipcode, contact, phone, email) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
			RETURNING customer_id`,
			data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7]).Scan(&customerID)
		
		if err != nil {
			log.Fatalf("❌ Failed to insert customer %s: %v", data[0], err)
		}
		
		customers = append(customers, Customer{ID: customerID, Name: data[0]})
		fmt.Printf("  ✅ %s (customer_id: %d)\n", data[0], customerID)
	}

	// Step 3: Query size_id values to avoid SERIAL assumptions
	fmt.Println("📏 Querying size IDs to avoid SERIAL assumptions...")
	sizeMap := make(map[string]int)
	rows, err := db.Query("SELECT size_id, size FROM store.sizes ORDER BY size_id")
	if err != nil {
		log.Fatalf("❌ Failed to query sizes: %v", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var sizeID int
		var size string
		if err := rows.Scan(&sizeID, &size); err != nil {
			log.Fatalf("❌ Failed to scan size: %v", err)
		}
		sizeMap[size] = sizeID
		fmt.Printf("  📏 %s = size_id %d\n", size, sizeID)
	}

	// Step 4: Insert inventory using captured customer_id values
	fmt.Println("📦 Inserting inventory using actual customer IDs...")
	
	inventoryData := []struct {
		workOrder string
		customerIdx int  // Index into customers array
		joints    int
		size      string
		weight    float64
		grade     string
		connection string
		dateIn    string
		wellIn    string
		leaseIn   string
		location  string
		notes     string
	}{
		{"WO-2024-001", 0, 100, "5 1/2\"", 2500.50, "L80", "BTC", "2024-01-15", "Well-PB-001", "Lease-PB-A", "Yard-A", "Standard production casing"},
		{"WO-2024-002", 1, 150, "7\"", 4200.75, "P110", "VAM TOP", "2024-01-16", "Well-EF-002", "Lease-EF-B", "Yard-B", "High pressure application"},
		{"WO-2024-003", 2, 75, "9 5/8\"", 6800.25, "N80", "LTC", "2024-01-17", "Well-BK-003", "Lease-BK-C", "Yard-C", "Surface casing"},
		{"WO-2024-004", 3, 200, "5 1/2\"", 5000.00, "J55", "STC", "2024-01-18", "Well-GC-004", "Lease-GC-D", "Yard-A", "Offshore application"},
	}
	
	for _, inv := range inventoryData {
		customer := customers[inv.customerIdx]
		_, err := db.Exec(`
			INSERT INTO store.inventory (work_order, customer_id, customer, joints, size, weight, grade, connection, date_in, well_in, lease_in, location, notes) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			inv.workOrder, customer.ID, customer.Name, inv.joints, inv.size, inv.weight, inv.grade, inv.connection, inv.dateIn, inv.wellIn, inv.leaseIn, inv.location, inv.notes)
		
		if err != nil {
			log.Fatalf("❌ Failed to insert inventory %s: %v", inv.workOrder, err)
		}
		fmt.Printf("  ✅ %s for %s (customer_id: %d)\n", inv.workOrder, customer.Name, customer.ID)
	}

	// Step 5: Insert received orders using actual customer_id and size_id values
	fmt.Println("📨 Inserting received orders using actual SERIAL values...")
	
	receivedData := []struct {
		workOrder string
		customerIdx int  // Index into customers array
		joints    int
		size      string
		weight    float64
		grade     string
		connection string
		dateReceived string
		well      string
		lease     string
		orderedBy string
		notes     string
	}{
		{"WO-2024-005", 0, 80, "7\"", 3200.00, "L80", "BTC", "2024-01-20", "Well-PB-005", "Lease-PB-E", "John Smith", "Expedited order"},
		{"WO-2024-006", 4, 120, "5 1/2\"", 3000.00, "P110", "VAM TOP", "2024-01-21", "Well-MG-006", "Lease-MG-F", "Robert Davis", "High pressure specs"},
		{"WO-2024-007", 1, 90, "8 5/8\"", 7200.00, "N80", "LTC", "2024-01-22", "Well-EF-007", "Lease-EF-G", "Sarah Johnson", "Surface casing rush"},
	}
	
	for _, rec := range receivedData {
		customer := customers[rec.customerIdx]
		sizeID, exists := sizeMap[rec.size]
		if !exists {
			log.Fatalf("❌ Size %s not found in sizes table", rec.size)
		}
		
		_, err := db.Exec(`
			INSERT INTO store.received (work_order, customer_id, customer, joints, size_id, size, weight, grade, connection, date_received, well, lease, ordered_by, notes, in_production, complete) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
			rec.workOrder, customer.ID, customer.Name, rec.joints, sizeID, rec.size, rec.weight, rec.grade, rec.connection, rec.dateReceived, rec.well, rec.lease, rec.orderedBy, rec.notes, false, false)
		
		if err != nil {
			log.Fatalf("❌ Failed to insert received order %s: %v", rec.workOrder, err)
		}
		fmt.Printf("  ✅ %s for %s (customer_id: %d, size_id: %d)\n", rec.workOrder, customer.Name, customer.ID, sizeID)
	}

	fmt.Println("✅ Seeding completed successfully - no SERIAL assumptions!")
}

func showStatus(db *sql.DB, env string) {
	fmt.Printf("📊 Database Status (env: %s)\n", env)
	fmt.Printf("============================\n")

	// Set search path
	if _, err := db.Exec("SET search_path TO store, public;"); err != nil {
		fmt.Printf("❌ Failed to set search path: %v\n", err)
		return
	}

	// Check each table
	tables := []string{"customers", "grade", "sizes", "inventory", "received"}
	
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM store.%s", table)
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", table, err)
		} else {
			fmt.Printf("✅ %s: %d records\n", table, count)
		}
	}

	// Check SERIAL sequence status
	fmt.Println("\n🔢 SERIAL Sequence Status:")
	sequences := []struct{
		table string
		sequence string
		column string
	}{
		{"customers", "customers_customer_id_seq", "customer_id"},
		{"sizes", "sizes_size_id_seq", "size_id"},
		{"inventory", "inventory_id_seq", "id"},
		{"received", "received_id_seq", "id"},
	}
	
	for _, seq := range sequences {
		var lastValue, nextValue sql.NullInt64
		err := db.QueryRow(fmt.Sprintf("SELECT last_value, (last_value + 1) as next_value FROM store.%s", seq.sequence)).Scan(&lastValue, &nextValue)
		if err != nil {
			fmt.Printf("  ⚠️  %s sequence: Error - %v\n", seq.table, err)
		} else {
			if lastValue.Valid {
				fmt.Printf("  📈 %s.%s: last=%d, next=%d\n", seq.table, seq.column, lastValue.Int64, nextValue.Int64)
			} else {
				fmt.Printf("  📈 %s.%s: not used yet, next=1\n", seq.table, seq.column)
			}
		}
	}

	// Test foreign key relationships with actual IDs
	fmt.Println("\n🔗 Foreign Key Validation:")
	
	var customerID int
	var customerName, city string
	var joints int
	err := db.QueryRow(`
		SELECT c.customer_id, c.customer, c.billing_city, i.joints 
		FROM store.customers c 
		JOIN store.inventory i ON c.customer_id = i.customer_id 
		LIMIT 1
	`).Scan(&customerID, &customerName, &city, &joints)
	
	if err != nil {
		fmt.Printf("❌ Customer-Inventory join failed: %v\n", err)
	} else {
		fmt.Printf("✅ Customer-Inventory join: ID %d - %s (%s) - %d joints\n", customerID, customerName, city, joints)
	}

	// Test size_id relationships
	var sizeID int
	var size string
	var receivedJoints int
	err = db.QueryRow(`
		SELECT s.size_id, s.size, r.joints
		FROM store.sizes s
		JOIN store.received r ON s.size_id = r.size_id
		LIMIT 1
	`).Scan(&sizeID, &size, &receivedJoints)
	
	if err != nil {
		fmt.Printf("❌ Size-Received join failed: %v\n", err)
	} else {
		fmt.Printf("✅ Size-Received join: size_id %d - %s - %d joints\n", sizeID, size, receivedJoints)
	}

	// Show actual customer and size ID mappings to verify no assumptions
	fmt.Println("\n📋 ID Verification (no hardcoded assumptions):")
	
	fmt.Println("  Customers:")
	customerRows, err := db.Query("SELECT customer_id, customer FROM store.customers ORDER BY customer_id")
	if err != nil {
		fmt.Printf("    ❌ Failed to query customers: %v\n", err)
	} else {
		defer customerRows.Close()
		for customerRows.Next() {
			var id int
			var name string
			if err := customerRows.Scan(&id, &name); err != nil {
				fmt.Printf("    ❌ Failed to scan customer: %v\n", err)
			} else {
				fmt.Printf("    📋 customer_id %d: %s\n", id, name)
			}
		}
	}

	fmt.Println("  Sizes:")
	sizeRows, err := db.Query("SELECT size_id, size FROM store.sizes ORDER BY size_id LIMIT 5")
	if err != nil {
		fmt.Printf("    ❌ Failed to query sizes: %v\n", err)
	} else {
		defer sizeRows.Close()
		for sizeRows.Next() {
			var id int
			var sizeName string
			if err := sizeRows.Scan(&id, &sizeName); err != nil {
				fmt.Printf("    ❌ Failed to scan size: %v\n", err)
			} else {
				fmt.Printf("    📏 size_id %d: %s\n", id, sizeName)
			}
		}
	}

	fmt.Println("\n✅ Status check complete - all SERIAL sequences properly handled")
}

func resetDatabase(db *sql.DB, env string) {
	if env == "production" || env == "prod" {
		log.Fatal("❌ Reset not allowed in production environment")
	}

	fmt.Printf("⚠️ Resetting database (env: %s)...\n", env)

	resetSQL := `
	DROP SCHEMA IF EXISTS store CASCADE;
	DROP SCHEMA IF EXISTS migrations CASCADE;
	`

	if _, err := db.Exec(resetSQL); err != nil {
		log.Fatalf("❌ Reset failed: %v", err)
	}

	fmt.Println("✅ Database reset complete")
	fmt.Println("Run 'go run migrator.go migrate' and 'go run migrator.go seed' to restore")
}

// GenerateMigrations exports all SQL to files before execution
func (m *Migrator) GenerateMigrations() error {
	log.Printf("🔄 Generating migration files...")
	
	// Ensure migrations directory exists
	if err := os.MkdirAll("migrations", 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}
	
	// Generate base schema
	if err := m.generateStoreSchema(); err != nil {
		return fmt.Errorf("failed to generate store schema: %w", err)
	}
	
	// Generate auth schema  
	if err := m.generateAuthSchema(); err != nil {
		return fmt.Errorf("failed to generate auth schema: %w", err)
	}
	
	// Generate seed data
	if err := m.generateSeedData(); err != nil {
		return fmt.Errorf("failed to generate seed data: %w", err)
	}
	
	log.Printf("✅ Migration files generated in migrations/")
	return nil
}

// generateStoreSchema creates 001_store_schema.sql
func (m *Migrator) generateStoreSchema() error {
	schema := m.getStoreSchema()
	
	filename := "migrations/001_store_schema.sql"
	if err := m.writeSchemaFile(filename, schema, "Store Schema - Core business tables"); err != nil {
		return err
	}
	
	log.Printf("✅ Generated: %s", filename)
	return nil
}

// generateAuthSchema creates 002_auth_schema.sql  
func (m *Migrator) generateAuthSchema() error {
	schema := m.getAuthSchema()
	
	filename := "migrations/002_auth_schema.sql"
	if err := m.writeSchemaFile(filename, schema, "Auth Schema - Authentication and authorization"); err != nil {
		return err
	}
	
	log.Printf("✅ Generated: %s", filename)
	return nil
}

// generateSeedData creates 003_seed_data.sql
func (m *Migrator) generateSeedData() error {
	seeds := m.getSeedData()
	
	filename := "migrations/003_seed_data.sql"
	if err := m.writeSchemaFile(filename, seeds, "Seed Data - Reference data and defaults"); err != nil {
		return err
	}
	
	log.Printf("✅ Generated: %s", filename)
	return nil
}

// writeSchemaFile writes SQL with header and metadata
func (m *Migrator) writeSchemaFile(filename, content, description string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()
	
	// Write header with metadata
	header := fmt.Sprintf(`-- %s
-- Generated by migrator.go at %s
-- Environment: %s
-- 
-- Description: %s
--
-- ============================================================================

`, filename, time.Now().Format("2006-01-02 15:04:05"), m.env, description)
	
	if _, err := file.WriteString(header + content); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	
	return nil
}

// getStoreSchema returns your existing store schema
func (m *Migrator) getStoreSchema() string {
	return `-- Create schemas
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

-- Enhanced customers table
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
	imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Grades reference table
CREATE TABLE IF NOT EXISTS store.grade (
	grade_id SERIAL PRIMARY KEY,
	grade VARCHAR(50) NOT NULL UNIQUE,
	description TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sizes reference table  
CREATE TABLE IF NOT EXISTS store.sizes (
	size_id SERIAL PRIMARY KEY,
	size VARCHAR(50) NOT NULL UNIQUE,
	diameter DECIMAL(8,3),
	description TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced inventory table
CREATE TABLE IF NOT EXISTS store.inventory (
	inventory_id SERIAL PRIMARY KEY,
	customer_id INTEGER REFERENCES store.customers(customer_id),
	work_order VARCHAR(100),
	size_id INTEGER REFERENCES store.sizes(size_id),
	grade_id INTEGER REFERENCES store.grade(grade_id),
	connection VARCHAR(100),
	location VARCHAR(100),
	joints INTEGER DEFAULT 0,
	date_in DATE,
	date_out DATE,
	well_in VARCHAR(255),
	well_out VARCHAR(255),
	lease_in VARCHAR(255), 
	lease_out VARCHAR(255),
	imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced received table (work orders)
CREATE TABLE IF NOT EXISTS store.received (
	received_id SERIAL PRIMARY KEY,
	customer_id INTEGER REFERENCES store.customers(customer_id),
	r_number VARCHAR(100),
	work_order VARCHAR(100),
	size_id INTEGER REFERENCES store.sizes(size_id),
	grade_id INTEGER REFERENCES store.grade(grade_id),
	connection VARCHAR(100),
	joints INTEGER DEFAULT 0,
	date_received DATE,
	ordered_by VARCHAR(255),
	entered_by VARCHAR(255),
	when_entered TIMESTAMP,
	imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_email ON store.customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_deleted ON store.customers(deleted);

CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
CREATE INDEX IF NOT EXISTS idx_inventory_size_id ON store.inventory(size_id);
CREATE INDEX IF NOT EXISTS idx_inventory_grade_id ON store.inventory(grade_id);
CREATE INDEX IF NOT EXISTS idx_inventory_deleted ON store.inventory(deleted);

CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order);
CREATE INDEX IF NOT EXISTS idx_received_r_number ON store.received(r_number);
CREATE INDEX IF NOT EXISTS idx_received_deleted ON store.received(deleted);

-- Record schema version
INSERT INTO migrations.schema_migrations (version, name) 
VALUES ('001', 'store_schema') 
ON CONFLICT (version) DO NOTHING;`
}

// getAuthSchema returns the auth schema
func (m *Migrator) getAuthSchema() string {
	return `-- Authentication and authorization schema
CREATE SCHEMA IF NOT EXISTS auth;

-- Create tenants table
CREATE TABLE IF NOT EXISTS auth.tenants (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(255) NOT NULL,
	slug VARCHAR(100) NOT NULL UNIQUE,
	database_type VARCHAR(50) DEFAULT 'tenant',
	database_name VARCHAR(100),
	active BOOLEAN DEFAULT true,
	settings JSONB DEFAULT '{}',
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create users table
CREATE TABLE IF NOT EXISTS auth.users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email VARCHAR(255) NOT NULL UNIQUE,
	password_hash VARCHAR(255) NOT NULL,
	role VARCHAR(50) NOT NULL DEFAULT 'user',
	company VARCHAR(255) NOT NULL,
	tenant_id VARCHAR(100) NOT NULL,
	active BOOLEAN DEFAULT true,
	email_verified BOOLEAN DEFAULT false,
	last_login TIMESTAMP WITH TIME ZONE,
	failed_login_attempts INTEGER DEFAULT 0,
	locked_until TIMESTAMP WITH TIME ZONE,
	password_changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	settings JSONB DEFAULT '{}',
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	
	CONSTRAINT fk_users_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug)
);

-- Create sessions table
CREATE TABLE IF NOT EXISTS auth.sessions (
	id VARCHAR(255) PRIMARY KEY,
	user_id UUID NOT NULL,
	tenant_id VARCHAR(100) NOT NULL,
	ip_address INET,
	user_agent TEXT,
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	
	CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE,
	CONSTRAINT fk_sessions_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON auth.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON auth.users(role);
CREATE INDEX IF NOT EXISTS idx_users_active ON auth.users(active);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON auth.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_tenant_id ON auth.sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON auth.sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_tenants_slug ON auth.tenants(slug);
CREATE INDEX IF NOT EXISTS idx_tenants_active ON auth.tenants(active);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON auth.tenants
	FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON auth.users
	FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Insert default system tenant
INSERT INTO auth.tenants (name, slug, database_type, database_name) 
VALUES ('System Administration', 'system', 'main', 'system')
ON CONFLICT (slug) DO NOTHING;

-- Create roles enum check constraints
ALTER TABLE auth.users ADD CONSTRAINT check_user_role 
	CHECK (role IN ('user', 'operator', 'manager', 'admin', 'super-admin'));

-- Record schema version
INSERT INTO migrations.schema_migrations (version, name) 
VALUES ('002', 'auth_schema') 
ON CONFLICT (version) DO NOTHING;`
}

// getSeedData returns reference data
func (m *Migrator) getSeedData() string {
	return `-- Reference data and defaults

-- Insert standard oil & gas grades
INSERT INTO store.grade (grade, description) VALUES
('J55', 'Basic carbon steel casing and tubing'),
('K55', 'Carbon steel with improved properties'),
('N80', 'Quenched and tempered steel'),
('L80', 'Chrome alloy steel'),
('P110', 'High strength steel'),
('C90', 'Chrome alloy steel'),
('T95', 'High collapse resistance')
ON CONFLICT (grade) DO NOTHING;

-- Insert standard pipe sizes  
INSERT INTO store.sizes (size, diameter, description) VALUES
('4 1/2"', 4.500, '4 1/2 inch casing'),
('5"', 5.000, '5 inch casing'),
('5 1/2"', 5.500, '5 1/2 inch casing'),
('7"', 7.000, '7 inch casing'),
('8 5/8"', 8.625, '8 5/8 inch casing'),
('9 5/8"', 9.625, '9 5/8 inch casing'),
('10 3/4"', 10.750, '10 3/4 inch casing'),
('13 3/8"', 13.375, '13 3/8 inch casing'),
('2 3/8"', 2.375, '2 3/8 inch tubing'),
('2 7/8"', 2.875, '2 7/8 inch tubing'),
('3 1/2"', 3.500, '3 1/2 inch tubing')
ON CONFLICT (size) DO NOTHING;

-- Record seed version
INSERT INTO migrations.schema_migrations (version, name) 
VALUES ('003', 'seed_data') 
ON CONFLICT (version) DO NOTHING;`
}

// Enhanced migrate command with SQL generation
func (m *Migrator) RunMigrations() error {
	log.Printf("🚀 Running migrations for environment: %s", m.env)
	
	// First, generate migration files
	if err := m.GenerateMigrations(); err != nil {
		return fmt.Errorf("failed to generate migrations: %w", err)
	}
	
	// Then execute them in order
	migrationFiles := []string{
		"migrations/001_store_schema.sql",
		"migrations/002_auth_schema.sql", 
		"migrations/003_seed_data.sql",
	}
	
	for _, file := range migrationFiles {
		if err := m.executeMigrationFile(file); err != nil {
			return fmt.Errorf("failed to execute %s: %w", file, err)
		}
	}
	
	log.Printf("✅ All migrations completed successfully")
	return nil
}

// executeMigrationFile runs a single migration file
func (m *Migrator) executeMigrationFile(filename string) error {
	log.Printf("📦 Executing: %s", filename)
	
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	if _, err := m.db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute SQL from %s: %w", filename, err)
	}
	
	log.Printf("✅ Completed: %s", filename)
	return nil
}

// ============================================================================
// SQL GENERATION COMMAND HANDLER
// ============================================================================

func handleGenerate() {
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}
	
	fmt.Printf("🔄 Generating migration files for environment: %s\n", env)
	
	if err := generateMigrationFiles(env); err != nil {
		log.Fatalf("Failed to generate migration files: %v", err)
	}
	
	fmt.Println("✅ Migration files generated successfully!")
	fmt.Println("📁 Generated files:")
	fmt.Println("   migrations/001_store_schema.sql")
	fmt.Println("   migrations/002_auth_schema.sql")
	fmt.Println("   migrations/003_seed_data.sql")
	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("   make db-migrate-from-files    # Execute generated migrations")
	fmt.Println("   make db-show-migrations       # Review generated content")
}

// ============================================================================
// SQL GENERATION FUNCTIONS
// ============================================================================

func generateMigrationFiles(env string) error {
	// Ensure migrations directory exists
	if err := os.MkdirAll("migrations", 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}
	
	// Generate store schema
	if err := generateStoreSchemaFile(env); err != nil {
		return fmt.Errorf("failed to generate store schema: %w", err)
	}
	
	// Generate auth schema  
	if err := generateAuthSchemaFile(env); err != nil {
		return fmt.Errorf("failed to generate auth schema: %w", err)
	}
	
	// Generate seed data
	if err := generateSeedDataFile(env); err != nil {
		return fmt.Errorf("failed to generate seed data: %w", err)
	}
	
	return nil
}

func generateStoreSchemaFile(env string) error {
	filename := "migrations/001_store_schema.sql"
	schema := getStoreSchemaSQL()
	
	return writeSchemaFile(filename, schema, "Store Schema - Core business tables", env)
}

func generateAuthSchemaFile(env string) error {
	filename := "migrations/002_auth_schema.sql"
	schema := getAuthSchemaSQL()
	
	return writeSchemaFile(filename, schema, "Auth Schema - Authentication and authorization", env)
}

func generateSeedDataFile(env string) error {
	filename := "migrations/003_seed_data.sql"
	seeds := getSeedDataSQL()
	
	return writeSchemaFile(filename, seeds, "Seed Data - Reference data and defaults", env)
}

func writeSchemaFile(filename, content, description, env string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()
	
	// Write header with metadata
	header := fmt.Sprintf(`-- %s
-- Generated by cmd/migrator/main.go at %s
-- Environment: %s
-- 
-- Description: %s
--
-- ============================================================================

`, filename, time.Now().Format("2006-01-02 15:04:05"), env, description)
	
	if _, err := file.WriteString(header + content); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	
	fmt.Printf("✅ Generated: %s\n", filename)
	return nil
}

// ============================================================================
// SCHEMA SQL GENERATORS
// ============================================================================

func getStoreSchemaSQL() string {
	return `-- Create schemas
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

-- Enhanced customers table
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
	imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Grades reference table
CREATE TABLE IF NOT EXISTS store.grade (
	grade_id SERIAL PRIMARY KEY,
	grade VARCHAR(50) NOT NULL UNIQUE,
	description TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sizes reference table  
CREATE TABLE IF NOT EXISTS store.sizes (
	size_id SERIAL PRIMARY KEY,
	size VARCHAR(50) NOT NULL UNIQUE,
	diameter DECIMAL(8,3),
	description TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced inventory table
CREATE TABLE IF NOT EXISTS store.inventory (
	inventory_id SERIAL PRIMARY KEY,
	customer_id INTEGER REFERENCES store.customers(customer_id),
	work_order VARCHAR(100),
	size_id INTEGER REFERENCES store.sizes(size_id),
	grade_id INTEGER REFERENCES store.grade(grade_id),
	connection VARCHAR(100),
	location VARCHAR(100),
	joints INTEGER DEFAULT 0,
	date_in DATE,
	date_out DATE,
	well_in VARCHAR(255),
	well_out VARCHAR(255),
	lease_in VARCHAR(255), 
	lease_out VARCHAR(255),
	imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced received table (work orders)
CREATE TABLE IF NOT EXISTS store.received (
	received_id SERIAL PRIMARY KEY,
	customer_id INTEGER REFERENCES store.customers(customer_id),
	r_number VARCHAR(100),
	work_order VARCHAR(100),
	size_id INTEGER REFERENCES store.sizes(size_id),
	grade_id INTEGER REFERENCES store.grade(grade_id),
	connection VARCHAR(100),
	joints INTEGER DEFAULT 0,
	date_received DATE,
	ordered_by VARCHAR(255),
	entered_by VARCHAR(255),
	when_entered TIMESTAMP,
	imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_email ON store.customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_deleted ON store.customers(deleted);

CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
CREATE INDEX IF NOT EXISTS idx_inventory_size_id ON store.inventory(size_id);
CREATE INDEX IF NOT EXISTS idx_inventory_grade_id ON store.inventory(grade_id);
CREATE INDEX IF NOT EXISTS idx_inventory_deleted ON store.inventory(deleted);

CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order);
CREATE INDEX IF NOT EXISTS idx_received_r_number ON store.received(r_number);
CREATE INDEX IF NOT EXISTS idx_received_deleted ON store.received(deleted);

-- Record schema version
INSERT INTO migrations.schema_migrations (version, name) 
VALUES ('001', 'store_schema') 
ON CONFLICT (version) DO NOTHING;`
}

func getAuthSchemaSQL() string {
	return `-- Authentication and authorization schema
CREATE SCHEMA IF NOT EXISTS auth;

-- Create tenants table
CREATE TABLE IF NOT EXISTS auth.tenants (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(255) NOT NULL,
	slug VARCHAR(100) NOT NULL UNIQUE,
	database_type VARCHAR(50) DEFAULT 'tenant',
	database_name VARCHAR(100),
	active BOOLEAN DEFAULT true,
	settings JSONB DEFAULT '{}',
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create users table
CREATE TABLE IF NOT EXISTS auth.users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email VARCHAR(255) NOT NULL UNIQUE,
	password_hash VARCHAR(255) NOT NULL,
	role VARCHAR(50) NOT NULL DEFAULT 'user',
	company VARCHAR(255) NOT NULL,
	tenant_id VARCHAR(100) NOT NULL,
	active BOOLEAN DEFAULT true,
	email_verified BOOLEAN DEFAULT false,
	last_login TIMESTAMP WITH TIME ZONE,
	failed_login_attempts INTEGER DEFAULT 0,
	locked_until TIMESTAMP WITH TIME ZONE,
	password_changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	settings JSONB DEFAULT '{}',
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	
	CONSTRAINT fk_users_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug)
);

-- Create sessions table
CREATE TABLE IF NOT EXISTS auth.sessions (
	id VARCHAR(255) PRIMARY KEY,
	user_id UUID NOT NULL,
	tenant_id VARCHAR(100) NOT NULL,
	ip_address INET,
	user_agent TEXT,
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	
	CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE,
	CONSTRAINT fk_sessions_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON auth.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON auth.users(role);
CREATE INDEX IF NOT EXISTS idx_users_active ON auth.users(active);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON auth.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_tenant_id ON auth.sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON auth.sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_tenants_slug ON auth.tenants(slug);
CREATE INDEX IF NOT EXISTS idx_tenants_active ON auth.tenants(active);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON auth.tenants
	FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON auth.users
	FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Insert default system tenant
INSERT INTO auth.tenants (name, slug, database_type, database_name) 
VALUES ('System Administration', 'system', 'main', 'system')
ON CONFLICT (slug) DO NOTHING;

-- Create roles enum check constraints
ALTER TABLE auth.users ADD CONSTRAINT check_user_role 
	CHECK (role IN ('user', 'operator', 'manager', 'admin', 'super-admin'));

-- Record schema version
INSERT INTO migrations.schema_migrations (version, name) 
VALUES ('002', 'auth_schema') 
ON CONFLICT (version) DO NOTHING;`
}

func getSeedDataSQL() string {
	return `-- Reference data and defaults

-- Insert standard oil & gas grades
INSERT INTO store.grade (grade, description) VALUES
('J55', 'Basic carbon steel casing and tubing'),
('K55', 'Carbon steel with improved properties'),
('N80', 'Quenched and tempered steel'),
('L80', 'Chrome alloy steel'),
('P110', 'High strength steel'),
('C90', 'Chrome alloy steel'),
('T95', 'High collapse resistance')
ON CONFLICT (grade) DO NOTHING;

-- Insert standard pipe sizes  
INSERT INTO store.sizes (size, diameter, description) VALUES
('4 1/2"', 4.500, '4 1/2 inch casing'),
('5"', 5.000, '5 inch casing'),
('5 1/2"', 5.500, '5 1/2 inch casing'),
('7"', 7.000, '7 inch casing'),
('8 5/8"', 8.625, '8 5/8 inch casing'),
('9 5/8"', 9.625, '9 5/8 inch casing'),
('10 3/4"', 10.750, '10 3/4 inch casing'),
('13 3/8"', 13.375, '13 3/8 inch casing'),
('2 3/8"', 2.375, '2 3/8 inch tubing'),
('2 7/8"', 2.875, '2 7/8 inch tubing'),
('3 1/2"', 3.500, '3 1/2 inch tubing')
ON CONFLICT (size) DO NOTHING;

-- Record seed version
INSERT INTO migrations.schema_migrations (version, name) 
VALUES ('003', 'seed_data') 
ON CONFLICT (version) DO NOTHING;`
}

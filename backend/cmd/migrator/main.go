package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// TenantMigrator handles multi-tenant database operations
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
		showHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle tenant-specific commands
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
	case "generate":
		handleGenerate()
	case "migrate":
		handleMigrate()
	case "seed":
		handleSeed()
	case "status":
		handleStatus()
	case "reset":
		handleReset()
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func showHelp() {
	fmt.Println("Oil & Gas Inventory System - Database Migrator")
	fmt.Println()
	fmt.Println("Standard Commands:")
	fmt.Println("  migrator migrate [env]        - Run migrations")
	fmt.Println("  migrator seed [env]           - Seed database")
	fmt.Println("  migrator status [env]         - Show status")
	fmt.Println("  migrator reset [env]          - Reset database")
	fmt.Println("  migrator generate [env]       - Generate migration files")
	fmt.Println()
	fmt.Println("Tenant Commands:")
	fmt.Println("  migrator tenant-create <id>   - Create tenant database")
	fmt.Println("  migrator csv-import <id> <dir> - Import CSV to tenant")
	fmt.Println("  migrator tenant-status <id>   - Check tenant status")
	fmt.Println("  migrator tenant-drop <id>     - Drop tenant database")
	fmt.Println("  migrator tenant-list          - List all tenant databases")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  migrator generate local")
	fmt.Println("  migrator migrate local")
	fmt.Println("  migrator tenant-create longbeach")
	fmt.Println("  migrator csv-import longbeach ../csv/longbeach/")
	fmt.Println()
	fmt.Println("Environments: local, test, production")
}

// ============================================================================
// COMMAND HANDLERS
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
	fmt.Println("   migrator migrate local    # Execute generated migrations")
}

func handleMigrate() {
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}
	
	fmt.Printf("🚀 Running migrations for environment: %s\n", env)
	
	db, err := connectDatabase(env)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	if err := runMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	
	fmt.Println("✅ Migrations completed successfully!")
}

func handleSeed() {
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}
	
	fmt.Printf("🌱 Seeding database for environment: %s\n", env)
	
	db, err := connectDatabase(env)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	if err := runSeeds(db); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}
	
	fmt.Println("✅ Database seeded successfully!")
}

func handleStatus() {
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}
	
	fmt.Printf("📊 Database status for environment: %s\n", env)
	
	db, err := connectDatabase(env)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	if err := showStatus(db); err != nil {
		log.Fatalf("Status check failed: %v", err)
	}
}

func handleReset() {
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}
	
	fmt.Printf("⚠️  WARNING: This will drop all data in %s environment\n", env)
	fmt.Print("Type 'yes' to confirm: ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	if confirmation != "yes" {
		fmt.Println("Operation cancelled")
		os.Exit(0)
	}
	
	db, err := connectDatabase(env)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	if err := resetDatabase(db); err != nil {
		log.Fatalf("Reset failed: %v", err)
	}
	
	fmt.Println("✅ Database reset completed!")
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

// ============================================================================
// DATABASE CONNECTION AND UTILITIES
// ============================================================================

func connectDatabase(env string) (*sql.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}
	
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return db, nil
}

func validateTenantID(tenantID string) error {
	if len(tenantID) < 2 || len(tenantID) > 20 {
		return fmt.Errorf("tenant ID must be 2-20 characters")
	}
	
	for _, char := range tenantID {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("tenant ID must contain only lowercase letters, numbers, and underscores")
		}
	}
	
	return nil
}

// ============================================================================
// SQL GENERATION FUNCTIONS (Function-based approach)
// ============================================================================

func generateMigrationFiles(env string) error {
	if err := os.MkdirAll("migrations", 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}
	
	if err := generateStoreSchemaFile(env); err != nil {
		return fmt.Errorf("failed to generate store schema: %w", err)
	}
	
	if err := generateAuthSchemaFile(env); err != nil {
		return fmt.Errorf("failed to generate auth schema: %w", err)
	}
	
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
	
	header := fmt.Sprintf(`-- %s
-- Generated by cmd/migrator/main.go at %s
-- Environment: %s
-- Description: %s
-- ============================================================================

`, filename, time.Now().Format("2006-01-02 15:04:05"), env, description)
	
	if _, err := file.WriteString(header + content); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	
	return nil
}

func getStoreSchemaSQL() string {
	return `-- Create schemas
CREATE SCHEMA IF NOT EXISTS store;
CREATE SCHEMA IF NOT EXISTS migrations;

-- Create migrations table
CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
	version VARCHAR(255) PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create customers table
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
	tenant_id VARCHAR(100),
	imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create reference tables
CREATE TABLE IF NOT EXISTS store.grade (
	grade_id SERIAL PRIMARY KEY,
	grade VARCHAR(50) NOT NULL UNIQUE,
	description TEXT
);

CREATE TABLE IF NOT EXISTS store.sizes (
	size_id SERIAL PRIMARY KEY,
	size VARCHAR(50) NOT NULL UNIQUE,
	diameter DECIMAL(6,3),
	description TEXT
);

-- Create inventory table
CREATE TABLE IF NOT EXISTS store.inventory (
	id SERIAL PRIMARY KEY,
	work_order VARCHAR(100),
	customer_id INTEGER REFERENCES store.customers(customer_id),
	customer VARCHAR(255),
	joints INTEGER DEFAULT 0,
	size VARCHAR(50),
	weight DECIMAL(10,2),
	grade VARCHAR(50),
	connection VARCHAR(100),
	location VARCHAR(255),
	notes TEXT,
	tenant_id VARCHAR(100),
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

-- Create received table (work orders)
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
CREATE INDEX IF NOT EXISTS idx_inventory_deleted ON store.inventory(deleted);
CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order);
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

-- Insert default system tenant
INSERT INTO auth.tenants (name, slug, database_type, database_name) 
VALUES ('System Administration', 'system', 'main', 'system')
ON CONFLICT (slug) DO NOTHING;

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

// ============================================================================
// MIGRATION EXECUTION FUNCTIONS
// ============================================================================

func runMigrations(db *sql.DB) error {
	migrationFiles := []string{
		"migrations/001_store_schema.sql",
		"migrations/002_auth_schema.sql", 
		"migrations/003_seed_data.sql",
	}
	
	for _, file := range migrationFiles {
		if err := executeMigrationFile(db, file); err != nil {
			return fmt.Errorf("failed to execute %s: %w", file, err)
		}
	}
	
	return nil
}

func executeMigrationFile(db *sql.DB, filename string) error {
	log.Printf("📦 Executing: %s", filename)
	
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute SQL from %s: %w", filename, err)
	}
	
	log.Printf("✅ Completed: %s", filename)
	return nil
}

func runSeeds(db *sql.DB) error {
	seedFile := "migrations/003_seed_data.sql"
	return executeMigrationFile(db, seedFile)
}

func showStatus(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM migrations.schema_migrations").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	
	fmt.Printf("Applied migrations: %d\n", count)
	
	rows, err := db.Query("SELECT version, name, applied_at FROM migrations.schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to query migration details: %w", err)
	}
	defer rows.Close()
	
	fmt.Println("\nMigration History:")
	for rows.Next() {
		var version, name, appliedAt string
		if err := rows.Scan(&version, &name, &appliedAt); err != nil {
			continue
		}
		fmt.Printf("  %s - %s (%s)\n", version, name, appliedAt)
	}
	
	return nil
}

func resetDatabase(db *sql.DB) error {
	commands := []string{
		"DROP SCHEMA IF EXISTS store CASCADE",
		"DROP SCHEMA IF EXISTS auth CASCADE", 
		"DROP SCHEMA IF EXISTS migrations CASCADE",
	}
	
	for _, cmd := range commands {
		if _, err := db.Exec(cmd); err != nil {
			return fmt.Errorf("failed to execute %s: %w", cmd, err)
		}
	}
	
	return nil
}

// ============================================================================
// TENANT MIGRATOR IMPLEMENTATION
// ============================================================================

func NewTenantMigrator(tenantID string) (*TenantMigrator, error) {
	// Connect to base database
	baseDB, err := connectDatabase("local")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to base database: %w", err)
	}
	
	return &TenantMigrator{
		baseDB:   baseDB,
		tenantID: tenantID,
	}, nil
}

func (tm *TenantMigrator) Close() {
	if tm.baseDB != nil {
		tm.baseDB.Close()
	}
	if tm.tenantDB != nil {
		tm.tenantDB.Close()
	}
}

func (tm *TenantMigrator) CreateTenantDatabase() error {
	dbName := fmt.Sprintf("oilgas_%s", tm.tenantID)
	
	// Create database
	createQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
	if _, err := tm.baseDB.Exec(createQuery); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		log.Printf("Database %s already exists", dbName)
	}
	
	// Connect to tenant database and run migrations
	if err := tm.connectToTenantDB(dbName); err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	
	if err := runMigrations(tm.tenantDB); err != nil {
		return fmt.Errorf("failed to run tenant migrations: %w", err)
	}
	
	log.Printf("✅ Tenant database ready: %s", dbName)
	return nil
}

func (tm *TenantMigrator) connectToTenantDB(dbName string) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
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
	
	parts[len(parts)-1] = dbName + params
	tenantURL := strings.Join(parts, "/")
	
	db, err := sql.Open("postgres", tenantURL)
	if err != nil {
		return fmt.Errorf("failed to open tenant database: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping tenant database: %w", err)
	}
	
	tm.tenantDB = db
	return nil
}

func (tm *TenantMigrator) ImportCSVData(csvDir string) error {
	log.Printf("🔄 Importing CSV data from: %s", csvDir)
	// Implementation would go here - CSV import logic
	// This is a placeholder for the CSV import functionality
	log.Printf("✅ CSV import completed (placeholder)")
	return nil
}

func (tm *TenantMigrator) ShowTenantStatus() error {
	dbName := fmt.Sprintf("oilgas_%s", tm.tenantID)
	log.Printf("📊 Tenant Status: %s", dbName)
	
	if err := tm.connectToTenantDB(dbName); err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	
	return showStatus(tm.tenantDB)
}

func (tm *TenantMigrator) DropTenantDatabase() error {
	dbName := fmt.Sprintf("oilgas_%s", tm.tenantID)
	
	// Close any existing connections
	if tm.tenantDB != nil {
		tm.tenantDB.Close()
		tm.tenantDB = nil
	}
	
	// Drop database
	dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	if _, err := tm.baseDB.Exec(dropQuery); err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}
	
	log.Printf("✅ Tenant database dropped: %s", dbName)
	return nil
}

func listTenantDatabases() error {
	db, err := connectDatabase("local")
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	
	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datname LIKE 'oilgas_%' 
		ORDER BY datname
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tenant databases: %w", err)
	}
	defer rows.Close()
	
	fmt.Println("📋 Tenant Databases:")
	count := 0
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		
		// Extract tenant ID from database name
		tenantID := strings.TrimPrefix(dbName, "oilgas_")
		fmt.Printf("  • %s (tenant: %s)\n", dbName, tenantID)
		count++
	}
	
	if count == 0 {
		fmt.Println("  No tenant databases found")
	} else {
		fmt.Printf("\nTotal tenant databases: %d\n", count)
	}
	
	return nil
}

package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Migration struct {
	Version    string
	Name       string
	SQL        string
	ExecutedAt *time.Time
}

type Migrator struct {
	db  *sql.DB
	env string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Oil & Gas Inventory System - Database Migrator")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  migrator migrate [env]  - Run migrations")
		fmt.Println("  migrator seed [env]     - Seed database")
		fmt.Println("  migrator status [env]   - Show migration status")
		fmt.Println("  migrator reset [env]    - Reset database")
		fmt.Println()
		fmt.Println("Environments: local, test, production")
		os.Exit(1)
	}

	command := os.Args[1]
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}

	// Load environment variables
	if err := loadEnv(env); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	migrator, err := NewMigrator(env)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer migrator.Close()

	switch command {
	case "migrate":
		if err := migrator.RunMigrations(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	case "seed":
		if err := migrator.RunSeeds(); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
	case "status":
		if err := migrator.ShowStatus(); err != nil {
			log.Fatalf("Status check failed: %v", err)
		}
	case "reset":
		if err := migrator.ResetDatabase(); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func loadEnv(env string) error {
	// Try multiple environment file locations
	envFiles := []string{
		".env",
		".env.local",
		fmt.Sprintf(".env.%s", env),
		"../.env",
		"../.env.local",
		fmt.Sprintf("../.env.%s", env),
	}

	loaded := false
	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			if err := godotenv.Load(file); err == nil {
				log.Printf("Loaded environment from: %s", file)
				loaded = true
				break
			}
		}
	}

	if !loaded {
		log.Printf("Warning: No .env file found, using system environment")
	}

	return nil
}

func NewMigrator(env string) (*Migrator, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✅ Connected to database successfully")
	return &Migrator{
		db:  db,
		env: env,
	}, nil
}

func (m *Migrator) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE SCHEMA IF NOT EXISTS migrations;
		
		CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS idx_schema_migrations_executed_at 
		ON migrations.schema_migrations(executed_at);
	`
	
	_, err := m.db.Exec(query)
	return err
}

func (m *Migrator) RunMigrations() error {
	log.Printf("Running migrations for environment: %s", m.env)
	
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	migrations, err := m.loadMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}
	
	executed, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}
	
	var pending []Migration
	for _, migration := range migrations {
		if _, exists := executed[migration.Version]; !exists {
			pending = append(pending, migration)
		}
	}
	
	if len(pending) == 0 {
		log.Println("✅ No pending migrations found")
		return nil
	}
	
	log.Printf("Found %d pending migrations", len(pending))
	
	for _, migration := range pending {
		log.Printf("🔄 Executing migration: %s - %s", migration.Version, migration.Name)
		
		// Handle migration 002 without transaction due to complex DDL
		if migration.Version == "002" {
			log.Printf("⚠️  Executing migration %s without transaction (complex DDL)", migration.Version)
			
			// Execute migration directly
			if _, err := m.db.Exec(migration.SQL); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
			}
			
			// Record migration separately with conflict handling
			recordSQL := `INSERT INTO migrations.schema_migrations (version, name) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING`
			if _, err := m.db.Exec(recordSQL, migration.Version, migration.Name); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}
		} else {
			// Use transaction for other migrations
			tx, err := m.db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
			
			if _, err := tx.Exec(migration.SQL); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
			}
			
			recordSQL := `INSERT INTO migrations.schema_migrations (version, name) VALUES ($1, $2)`
			if _, err := tx.Exec(recordSQL, migration.Version, migration.Name); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}
			
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
			}
		}
		
		log.Printf("✅ Migration %s completed", migration.Version)
	}
	
	log.Println("✅ All migrations completed successfully")
	return nil
}

func (m *Migrator) RunSeeds() error {
	log.Printf("Running seeds for environment: %s", m.env)

	var seedFile string
	switch m.env {
	case "local":
		seedFile = "seeds/local_seeds.sql"
	case "test":
		seedFile = "seeds/test_seeds.sql"
	case "production", "prod":
		seedFile = "seeds/production_seeds.sql"
	default:
		seedFile = "seeds/local_seeds.sql"
	}

	if _, err := os.Stat(seedFile); os.IsNotExist(err) {
		log.Printf("Seed file %s not found, creating basic seed data...", seedFile)
		return m.createBasicSeeds()
	}

	content, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("failed to read seed file %s: %w", seedFile, err)
	}

	log.Printf("Executing seed file: %s", seedFile)
	if _, err := m.db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute seeds: %w", err)
	}

	log.Println("✅ Seeds completed successfully")
	return nil
}

func (m *Migrator) createBasicSeeds() error {
	log.Println("Creating basic seed data...")

	seeds := `
	-- Set search path
	SET search_path TO store, public;
	
	-- Insert oil & gas industry standard grades
	INSERT INTO store.grade (grade, description) VALUES 
	('J55', 'Standard grade steel casing'),
	('JZ55', 'Enhanced J55 grade'),
	('L80', 'Higher strength grade'),
	('N80', 'Medium strength grade'),
	('P105', 'High performance grade'),
	('P110', 'Premium performance grade')
	ON CONFLICT (grade) DO NOTHING;
	
	-- Insert common pipe sizes
	INSERT INTO store.sizes (size, description) VALUES 
	('5 1/2"', '5.5 inch diameter'),
	('7"', '7 inch diameter'),
	('9 5/8"', '9.625 inch diameter'),
	('13 3/8"', '13.375 inch diameter'),
	('20"', '20 inch diameter')
	ON CONFLICT (size) DO NOTHING;
	
	-- Insert sample customer (development only)
	INSERT INTO store.customers (customer, billing_address, billing_city, billing_state, phone, email) VALUES 
	('Sample Oil Company', '123 Main St', 'Houston', 'TX', '555-0123', 'contact@sampleoil.com')
	ON CONFLICT DO NOTHING;
	`

	if _, err := m.db.Exec(seeds); err != nil {
		return fmt.Errorf("failed to create basic seeds: %w", err)
	}

	log.Println("✅ Basic seed data created successfully")
	return nil
}

func (m *Migrator) ShowStatus() error {
	fmt.Printf("\n=== Migration Status (Environment: %s) ===\n", m.env)

	// Check if migrations table exists
	var exists bool
	err := m.db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'migrations' AND table_name = 'schema_migrations')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !exists {
		fmt.Println("❌ Migrations table not found - run 'migrator migrate' first")
		return nil
	}

	migrations, err := m.loadMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	executed, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	fmt.Printf("%-15s %-30s %-10s %s\n", "Version", "Name", "Status", "Executed At")
	fmt.Println(strings.Repeat("-", 80))

	for _, migration := range migrations {
		if exec, exists := executed[migration.Version]; exists {
			fmt.Printf("%-15s %-30s %-10s %s\n", 
				migration.Version, 
				migration.Name, 
				"✅ Applied", 
				exec.ExecutedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("%-15s %-30s %-10s %s\n", 
				migration.Version, 
				migration.Name, 
				"⏳ Pending", 
				"-")
		}
	}

	fmt.Printf("\nTotal: %d migrations, %d applied, %d pending\n", 
		len(migrations), len(executed), len(migrations)-len(executed))

	// Check basic tables
	tables := []string{"customers", "grade", "sizes", "inventory", "received"}
	fmt.Println("\n📊 Schema Status:")
	for _, table := range tables {
		var count int
		err := m.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM store.%s", table)).Scan(&count)
		if err != nil {
			fmt.Printf("  ❌ %s: Error checking table\n", table)
		} else {
			fmt.Printf("  ✅ %s: %d records\n", table, count)
		}
	}

	fmt.Println("\n✅ Database status check complete")
	return nil
}

func (m *Migrator) ResetDatabase() error {
	log.Printf("⚠️  Resetting database for environment: %s", m.env)

	// Drop and recreate schemas
	dropSQL := `
	DROP SCHEMA IF EXISTS store CASCADE;
	DROP SCHEMA IF EXISTS auth CASCADE;
	DROP SCHEMA IF EXISTS migrations CASCADE;
	`

	if _, err := m.db.Exec(dropSQL); err != nil {
		return fmt.Errorf("failed to drop schemas: %w", err)
	}

	log.Println("✅ Database reset complete")
	log.Println("Run 'migrator migrate' and 'migrator seed' to restore")
	return nil
}

func (m *Migrator) loadMigrationFiles() ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir("migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		filename := filepath.Base(path)
		parts := strings.SplitN(strings.TrimSuffix(filename, ".sql"), "_", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename format: %s", filename)
		}

		migration := Migration{
			Version: parts[0],
			Name:    strings.ReplaceAll(parts[1], "_", " "),
			SQL:     string(content),
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) getExecutedMigrations() (map[string]Migration, error) {
	query := `SELECT version, name, executed_at FROM migrations.schema_migrations ORDER BY executed_at`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[string]Migration)
	for rows.Next() {
		var migration Migration
		if err := rows.Scan(&migration.Version, &migration.Name, &migration.ExecutedAt); err != nil {
			return nil, err
		}
		executed[migration.Version] = migration
	}

	return executed, rows.Err()
}

func (m *Migrator) createBasicSchema() error {
	log.Println("Creating basic schema...")

	schema := `
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
	
	-- Create basic tables for oil & gas inventory
	
	-- Customers table
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
	
	-- Inventory table
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
	);
	
	-- Received table
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
	);
	
	-- Create indexes for performance
	CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
	CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
	CREATE INDEX IF NOT EXISTS idx_inventory_date_in ON store.inventory(date_in);
	CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
	CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order);
	CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received);
	`

	if _, err := m.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create basic schema: %w", err)
	}

	log.Println("✅ Basic schema created successfully")
	return nil
}

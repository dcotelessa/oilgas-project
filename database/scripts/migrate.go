// database/scripts/migrate.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	
	_ "github.com/lib/pq"
)

type Migration struct {
	Version  string
	Name     string
	UpPath   string
	DownPath string
	Database string
}

type MigrationStatus struct {
	Version   string
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: migrate <database> <direction> [options]")
		fmt.Println("  database: auth|longbeach|bakersfield|colorado|all")
		fmt.Println("  direction: up|down|status")
		fmt.Println("  options: --steps=N (for rollback)")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  migrate auth up")
		fmt.Println("  migrate longbeach down --steps=1")
		fmt.Println("  migrate all status")
		os.Exit(1)
	}
	
	// Handle special case of 'all' database
	if os.Args[1] == "all" {
		handleAllDatabases(os.Args[2])
		return
	}
	
	database := os.Args[1]
	direction := os.Args[2]
	
	var dbURL string
	switch database {
	case "auth":
		dbURL = os.Getenv("CENTRAL_AUTH_DB_URL")
	case "longbeach":
		dbURL = os.Getenv("LONGBEACH_DB_URL")
	case "bakersfield":
		dbURL = os.Getenv("BAKERSFIELD_DB_URL")
	case "colorado":
		dbURL = os.Getenv("COLORADO_DB_URL")
	default:
		log.Fatal("Invalid database. Use: auth|longbeach|bakersfield|colorado")
	}
	
	if dbURL == "" {
		log.Fatal("Database URL not set in environment")
	}
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	
	// Ensure migrations table exists
	createMigrationsTable(db)
	
	// Get migrations
	migrations, err := getMigrations(database)
	if err != nil {
		log.Fatal("Failed to get migrations:", err)
	}
	
	switch direction {
	case "up":
		runMigrationsUp(db, migrations)
	case "down":
		runMigrationsDown(db, migrations)
	case "status":
		showMigrationStatus(db, migrations)
	default:
		log.Fatal("Invalid direction. Use: up|down|status")
	}
}

func createMigrationsTable(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			database_name VARCHAR(50) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			checksum VARCHAR(64) -- For migration integrity
		)`
	
	if _, err := db.Exec(query); err != nil {
		log.Fatal("Failed to create migrations table:", err)
	}
}

func getMigrations(database string) ([]Migration, error) {
	// Check database-specific directory first
	dbSpecificDir := filepath.Join("..", "migrations", database)
	
	// Then check root migrations directory for database-specific files
	rootDir := filepath.Join("..", "migrations")
	
	var migrations []Migration
	
	// Get database-specific migrations
	if _, err := os.Stat(dbSpecificDir); err == nil {
		dbMigs, err := scanMigrationsInDir(dbSpecificDir, database)
		if err != nil {
			return nil, fmt.Errorf("error scanning %s: %v", dbSpecificDir, err)
		}
		migrations = append(migrations, dbMigs...)
	}
	
	// Get root migrations that match our database
	rootMigs, err := scanMigrationsInDir(rootDir, database)
	if err != nil {
		return nil, fmt.Errorf("error scanning root migrations: %v", err)
	}
	migrations = append(migrations, rootMigs...)
	
	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	
	return migrations, nil
}

func scanMigrationsInDir(dir, database string) ([]Migration, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	var migrations []Migration
	migrationMap := make(map[string]*Migration)
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		filename := file.Name()
		
		// Handle .up.sql and .down.sql files
		if strings.HasSuffix(filename, ".up.sql") {
			version, name := parseFilename(filename, ".up.sql")
			if version == "" {
				continue
			}
			
			key := version + "_" + name
			if migrationMap[key] == nil {
				migrationMap[key] = &Migration{
					Version:  version,
					Name:     name,
					Database: database,
				}
			}
			migrationMap[key].UpPath = filepath.Join(dir, filename)
			
		} else if strings.HasSuffix(filename, ".down.sql") {
			version, name := parseFilename(filename, ".down.sql")
			if version == "" {
				continue
			}
			
			key := version + "_" + name
			if migrationMap[key] == nil {
				migrationMap[key] = &Migration{
					Version:  version,
					Name:     name,
					Database: database,
				}
			}
			migrationMap[key].DownPath = filepath.Join(dir, filename)
			
		} else if strings.HasSuffix(filename, ".sql") {
			// Handle single .sql files (assume up migration)
			version, name := parseFilename(filename, ".sql")
			if version == "" {
				continue
			}
			
			key := version + "_" + name
			migrationMap[key] = &Migration{
				Version:  version,
				Name:     name,
				UpPath:   filepath.Join(dir, filename),
				Database: database,
			}
		}
	}
	
	// Convert map to slice
	for _, migration := range migrationMap {
		migrations = append(migrations, *migration)
	}
	
	return migrations, nil
}

func parseFilename(filename, suffix string) (version, name string) {
	base := strings.TrimSuffix(filename, suffix)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) >= 1 {
		version = parts[0]
	}
	if len(parts) >= 2 {
		name = parts[1]
	}
	return
}

func runMigrationsUp(db *sql.DB, migrations []Migration) {
	for _, migration := range migrations {
		// Check if already applied
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", migration.Version).Scan(&exists)
		if err != nil {
			log.Printf("Error checking migration %s: %v", migration.Version, err)
			continue
		}
		
		if exists {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}
		
		// Read and execute migration
		if migration.UpPath == "" {
			log.Printf("No up migration file for %s, skipping", migration.Version)
			continue
		}
		
		content, err := os.ReadFile(migration.UpPath)
		if err != nil {
			log.Printf("Error reading migration %s: %v", migration.Version, err)
			continue
		}
		
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Error starting transaction for migration %s: %v", migration.Version, err)
			continue
		}
		
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			log.Printf("Error executing migration %s: %v", migration.Version, err)
			continue
		}
		
		if _, err := tx.Exec("INSERT INTO schema_migrations (version, name, database_name) VALUES ($1, $2, $3)", 
			migration.Version, migration.Name, migration.Database); err != nil {
			tx.Rollback()
			log.Printf("Error recording migration %s: %v", migration.Version, err)
			continue
		}
		
		if err := tx.Commit(); err != nil {
			log.Printf("Error committing migration %s: %v", migration.Version, err)
			continue
		}
		
		log.Printf("Applied migration %s: %s", migration.Version, migration.Name)
	}
}

func runMigrationsDown(db *sql.DB, migrations []Migration) {
	// Parse steps parameter if provided
	steps := parseStepsParam()
	if steps <= 0 {
		steps = 1 // Default to rolling back 1 migration
	}
	
	// Get applied migrations in reverse order
	var appliedMigrations []Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		
		// Check if applied
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", migration.Version).Scan(&exists)
		if err != nil {
			log.Printf("Error checking migration %s: %v", migration.Version, err)
			continue
		}
		
		if exists {
			appliedMigrations = append(appliedMigrations, migration)
			if len(appliedMigrations) >= steps {
				break
			}
		}
	}
	
	if len(appliedMigrations) == 0 {
		log.Println("No migrations to roll back")
		return
	}
	
	// Execute rollbacks
	for _, migration := range appliedMigrations {
		if err := rollbackMigration(db, migration); err != nil {
			log.Printf("Error rolling back migration %s: %v", migration.Version, err)
			// Stop on first error to avoid inconsistent state
			break
		}
		log.Printf("Rolled back migration %s: %s", migration.Version, migration.Name)
	}
}

func rollbackMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()
	
	// Execute down migration if available
	if migration.DownPath != "" {
		content, err := os.ReadFile(migration.DownPath)
		if err != nil {
			return fmt.Errorf("failed to read down migration file: %v", err)
		}
		
		if _, err := tx.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute down migration: %v", err)
		}
		log.Printf("Executed down migration SQL for %s", migration.Version)
	} else {
		log.Printf("WARNING: No down migration file for %s - only removing version record", migration.Version)
	}
	
	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %v", err)
	}
	
	return tx.Commit()
}

func parseStepsParam() int {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--steps=") {
			stepsStr := strings.TrimPrefix(arg, "--steps=")
			if steps := parseIntOrDefault(stepsStr, 1); steps > 0 {
				return steps
			}
		}
	}
	return 1
}

func parseIntOrDefault(str string, def int) int {
	var val int
	if n, err := fmt.Sscanf(str, "%d", &val); n == 1 && err == nil {
		return val
	}
	return def
}

func handleAllDatabases(direction string) {
	databases := []string{"auth", "longbeach", "bakersfield", "colorado"}
	
	for _, database := range databases {
		fmt.Printf("\n=== Processing database: %s ===\n", database)
		
		var dbURL string
		switch database {
		case "auth":
			dbURL = os.Getenv("CENTRAL_AUTH_DB_URL")
		case "longbeach":
			dbURL = os.Getenv("LONGBEACH_DB_URL")
		case "bakersfield":
			dbURL = os.Getenv("BAKERSFIELD_DB_URL")
		case "colorado":
			dbURL = os.Getenv("COLORADO_DB_URL")
		}
		
		if dbURL == "" {
			log.Printf("Skipping %s database: URL not set in environment", database)
			continue
		}
		
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("Failed to connect to %s database: %v", database, err)
			continue
		}
		defer db.Close()
		
		// Ensure migrations table exists
		createMigrationsTable(db)
		
		// Get migrations
		migrations, err := getMigrations(database)
		if err != nil {
			log.Printf("Failed to get migrations for %s: %v", database, err)
			continue
		}
		
		switch direction {
		case "up":
			runMigrationsUp(db, migrations)
		case "down":
			runMigrationsDown(db, migrations)
		case "status":
			showMigrationStatus(db, migrations)
		default:
			log.Printf("Invalid direction for %s database: %s", database, direction)
		}
	}
}

func showMigrationStatus(db *sql.DB, migrations []Migration) {
	fmt.Printf("%-20s %-40s %-10s %-20s\n", "Version", "Name", "Applied", "Applied At")
	fmt.Println(strings.Repeat("-", 90))
	
	for _, migration := range migrations {
		var applied bool
		var appliedAt *time.Time
		
		err := db.QueryRow(`
			SELECT true, applied_at 
			FROM schema_migrations 
			WHERE version = $1
		`, migration.Version).Scan(&applied, &appliedAt)
		
		if err != nil {
			applied = false
			appliedAt = nil
		}
		
		status := "No"
		dateStr := "-"
		
		if applied {
			status = "Yes"
			if appliedAt != nil {
				dateStr = appliedAt.Format("2006-01-02 15:04:05")
			}
		}
		
		fmt.Printf("%-20s %-40s %-10s %-20s\n", 
			migration.Version, 
			migration.Name, 
			status, 
			dateStr)
	}
}

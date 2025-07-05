// backend/migrator.go
package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Migration struct {
	Version     string
	Name        string
	SQL         string
	ExecutedAt  *time.Time
}

type Migrator struct {
	db   *pgxpool.Pool
	env  string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrator <command> [env]")
		fmt.Println("Commands:")
		fmt.Println("  migrate   - Run pending migrations")
		fmt.Println("  seed      - Run seed data for environment")
		fmt.Println("  status    - Show migration status")
		fmt.Println("")
		fmt.Println("Environments: local, dev, production")
		os.Exit(1)
	}

	command := os.Args[1]
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}

	// Load environment variables from root directory
	envFile := fmt.Sprintf("../.env.%s", env)
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Could not load %s", envFile)
		// Try root .env as fallback
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: Could not load ../.env either")
			// Try local files as last resort
			localEnvFile := fmt.Sprintf(".env.%s", env)
			if err := godotenv.Load(localEnvFile); err != nil {
				log.Printf("Warning: Could not load %s, trying .env", localEnvFile)
				godotenv.Load(".env")
			}
		}
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
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

// Rest of the migrator code stays the same...
func NewMigrator(env string) (*Migrator, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

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
	
	_, err := m.db.Exec(context.Background(), query)
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
		log.Println("No pending migrations")
		return nil
	}

	log.Printf("Found %d pending migrations", len(pending))

	for _, migration := range pending {
		log.Printf("Executing migration: %s - %s", migration.Version, migration.Name)
		
		tx, err := m.db.Begin(context.Background())
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(context.Background(), migration.SQL); err != nil {
			tx.Rollback(context.Background())
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}

		recordSQL := `INSERT INTO migrations.schema_migrations (version, name) VALUES ($1, $2)`
		if _, err := tx.Exec(context.Background(), recordSQL, migration.Version, migration.Name); err != nil {
			tx.Rollback(context.Background())
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		if err := tx.Commit(context.Background()); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
		}

		log.Printf("✅ Migration %s completed", migration.Version)
	}

	log.Println("All migrations completed successfully")
	return nil
}

func (m *Migrator) RunSeeds() error {
	log.Printf("Running seeds for environment: %s", m.env)
	
	var seedFile string
	switch m.env {
	case "local":
		seedFile = "seeds/local_seeds.sql"
	case "production", "prod":
		seedFile = "seeds/production_seeds.sql"
	default:
		seedFile = "seeds/local_seeds.sql"
	}

	content, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("failed to read seed file %s: %w", seedFile, err)
	}

	log.Printf("Executing seed file: %s", seedFile)
	if _, err := m.db.Exec(context.Background(), string(content)); err != nil {
		return fmt.Errorf("failed to execute seeds: %w", err)
	}

	log.Println("✅ Seeds completed successfully")
	return nil
}

func (m *Migrator) ShowStatus() error {
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

	fmt.Printf("\n=== Migration Status (Environment: %s) ===\n", m.env)
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

	rows, err := m.db.Query(context.Background(), query)
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

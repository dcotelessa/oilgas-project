// backend/cmd/migrator/main.go
package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type MigrationConfig struct {
	CentralAuthURL  string
	LongBeachURL    string
	DataPath        string
	LogPath         string
	TenantID        string
	BatchSize       int
	DryRun          bool
}

type MigrationStats struct {
	TablesProcessed int
	RecordsInserted int
	RecordsSkipped  int
	ErrorCount      int
	StartTime       time.Time
	EndTime         time.Time
}

func main() {
	config := MigrationConfig{
		CentralAuthURL:  getEnv("CENTRAL_AUTH_DB_URL", ""),
		LongBeachURL:    getEnv("LONGBEACH_DB_URL", ""),
		DataPath:        getEnv("DATA_PATH", "/app/data"),
		LogPath:         getEnv("LOG_PATH", "/app/logs"),
		TenantID:        getEnv("TENANT_ID", "local-dev"),
		BatchSize:       getEnvInt("BATCH_SIZE", 1000),
		DryRun:          getEnvBool("DRY_RUN", false),
	}

	log.Printf("Starting MDB to PostgreSQL migration...")
	log.Printf("Data path: %s", config.DataPath)
	log.Printf("Tenant ID: %s", config.TenantID)
	log.Printf("Dry run: %t", config.DryRun)

	// Create log directory
	os.MkdirAll(config.LogPath, 0755)

	// Run migration for dev database
	if err := runMigration("development", config.DevDatabaseURL, config); err != nil {
		log.Fatalf("Development migration failed: %v", err)
	}

	// Run migration for test database  
	if err := runMigration("test", config.TestDatabaseURL, config); err != nil {
		log.Fatalf("Test migration failed: %v", err)
	}

	log.Printf("Migration completed successfully!")
}

func runMigration(environment, databaseURL string, config MigrationConfig) error {
	log.Printf("Migrating to %s database...", environment)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to %s database: %w", environment, err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping %s database: %w", environment, err)
	}

	stats := &MigrationStats{StartTime: time.Now()}

	// Set tenant context
	if err := setTenantContext(db, config.TenantID); err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Migrate customers data
	if err := migrateCustomers(db, config, stats); err != nil {
		return fmt.Errorf("customer migration failed: %w", err)
	}

	stats.EndTime = time.Now()
	duration := stats.EndTime.Sub(stats.StartTime)

	log.Printf("%s Migration Summary:", strings.Title(environment))
	log.Printf("  Tables processed: %d", stats.TablesProcessed)
	log.Printf("  Records inserted: %d", stats.RecordsInserted)
	log.Printf("  Records skipped: %d", stats.RecordsSkipped)
	log.Printf("  Errors: %d", stats.ErrorCount)
	log.Printf("  Duration: %v", duration)

	// Write detailed log
	logFile := filepath.Join(config.LogPath, fmt.Sprintf("migration_%s_%s.log", environment, time.Now().Format("20060102_150405")))
	writeLogFile(logFile, environment, stats, config)

	return nil
}

func migrateCustomers(db *sql.DB, config MigrationConfig, stats *MigrationStats) error {
	log.Printf("Migrating customers data...")

	// Look for customer CSV files from Phase 1
	customerFiles := []string{
		filepath.Join(config.DataPath, "customers.csv"),
		filepath.Join(config.DataPath, "customer.csv"),
		filepath.Join(config.DataPath, "cust.csv"),
	}

	var customerFile string
	for _, file := range customerFiles {
		if _, err := os.Stat(file); err == nil {
			customerFile = file
			break
		}
	}

	if customerFile == "" {
		log.Printf("No customer CSV file found in %s, skipping customer migration", config.DataPath)
		return nil
	}

	log.Printf("Found customer data file: %s", customerFile)

	file, err := os.Open(customerFile)
	if err != nil {
		return fmt.Errorf("failed to open customer file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		log.Printf("Customer file is empty, skipping")
		return nil
	}

	// Get headers and map columns
	headers := records[0]
	columnMap := mapCustomerColumns(headers)
	
	log.Printf("Found %d customer records to process", len(records)-1)
	log.Printf("Column mapping: %+v", columnMap)

	// Process in batches
	recordCount := 0
	for i := 1; i < len(records); i += config.BatchSize {
		end := i + config.BatchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		inserted, skipped, err := insertCustomerBatch(db, batch, columnMap, config)
		if err != nil {
			log.Printf("Batch %d-%d failed: %v", i, end-1, err)
			stats.ErrorCount++
			continue
		}

		recordCount += inserted
		stats.RecordsInserted += inserted
		stats.RecordsSkipped += skipped

		log.Printf("Processed batch %d-%d: %d inserted, %d skipped", i, end-1, inserted, skipped)
	}

	stats.TablesProcessed++
	log.Printf("Customer migration completed: %d records processed", recordCount)
	return nil
}

func mapCustomerColumns(headers []string) map[string]int {
	// Map Access column names to positions
	columnMap := make(map[string]int)
	
	// Mapping based on common Access field names from Phase 1 analysis
	fieldMappings := map[string][]string{
		"customer_id":      {"custid", "customerid", "customer_id", "id"},
		"customer":         {"custname", "customername", "customer", "name", "company"},
		"billing_address":  {"billaddr", "billing_address", "address", "addr"},
		"billing_city":     {"billcity", "billing_city", "city"},
		"billing_state":    {"billstate", "billing_state", "state", "st"},
		"billing_zipcode":  {"billzip", "billing_zipcode", "zipcode", "zip"},
		"contact":          {"contact", "contactname", "contact_name"},
		"phone":            {"phoneno", "phone", "phonenum", "telephone"},
		"fax":              {"fax", "faxno", "faxnum"},
		"email":            {"email", "emailaddr", "email_address"},
	}

	for i, header := range headers {
		normalizedHeader := strings.ToLower(strings.TrimSpace(header))
		
		// Direct mapping
		for field, variations := range fieldMappings {
			for _, variation := range variations {
				if normalizedHeader == variation {
					columnMap[field] = i
					break
				}
			}
		}
	}

	return columnMap
}

func insertCustomerBatch(db *sql.DB, batch [][]string, columnMap map[string]int, config MigrationConfig) (int, int, error) {
	if config.DryRun {
		log.Printf("DRY RUN: Would insert %d customer records", len(batch))
		return len(batch), 0, nil
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO store.customers (
			customer, billing_address, billing_city, billing_state, billing_zipcode,
			contact, phone, fax, email, tenant_id, imported_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT DO NOTHING`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	skipped := 0
	now := time.Now()

	for _, record := range batch {
		customer := extractCustomerData(record, columnMap)
		
		// Skip if required fields are missing
		if customer.Customer == "" {
			skipped++
			continue
		}

		// Validate state code
		if customer.BillingState != "" && len(customer.BillingState) != 2 {
			customer.BillingState = "" // Clear invalid state
		}

		// Execute insert
		result, err := stmt.Exec(
			customer.Customer,
			nullString(customer.BillingAddress),
			nullString(customer.BillingCity),
			nullString(customer.BillingState),
			nullString(customer.BillingZipcode),
			nullString(customer.Contact),
			nullString(customer.Phone),
			nullString(customer.Fax),
			nullString(customer.Email),
			config.TenantID,
			now,
			now,
		)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				// Handle specific PostgreSQL errors
				switch pqErr.Code {
				case "23505": // unique_violation
					skipped++
					continue
				case "23514": // check_violation
					log.Printf("Data validation error for customer %s: %v", customer.Customer, err)
					skipped++
					continue
				}
			}
			return inserted, skipped, fmt.Errorf("failed to insert customer %s: %w", customer.Customer, err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			inserted++
		} else {
			skipped++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return inserted, skipped, nil
}

type CustomerRecord struct {
	CustomerID      string
	Customer        string
	BillingAddress  string
	BillingCity     string
	BillingState    string
	BillingZipcode  string
	Contact         string
	Phone           string
	Fax             string
	Email           string
}

func extractCustomerData(record []string, columnMap map[string]int) CustomerRecord {
	customer := CustomerRecord{}

	if idx, ok := columnMap["customer_id"]; ok && idx < len(record) {
		customer.CustomerID = strings.TrimSpace(record[idx])
	}
	if idx, ok := columnMap["customer"]; ok && idx < len(record) {
		customer.Customer = strings.TrimSpace(record[idx])
	}
	if idx, ok := columnMap["billing_address"]; ok && idx < len(record) {
		customer.BillingAddress = strings.TrimSpace(record[idx])
	}
	if idx, ok := columnMap["billing_city"]; ok && idx < len(record) {
		customer.BillingCity = strings.TrimSpace(record[idx])
	}
	if idx, ok := columnMap["billing_state"]; ok && idx < len(record) {
		state := strings.TrimSpace(strings.ToUpper(record[idx]))
		if len(state) <= 2 {
			customer.BillingState = state
		}
	}
	if idx, ok := columnMap["billing_zipcode"]; ok && idx < len(record) {
		customer.BillingZipcode = strings.TrimSpace(record[idx])
	}
	if idx, ok := columnMap["contact"]; ok && idx < len(record) {
		customer.Contact = strings.TrimSpace(record[idx])
	}
	if idx, ok := columnMap["phone"]; ok && idx < len(record) {
		customer.Phone = cleanPhoneNumber(strings.TrimSpace(record[idx]))
	}
	if idx, ok := columnMap["fax"]; ok && idx < len(record) {
		customer.Fax = cleanPhoneNumber(strings.TrimSpace(record[idx]))
	}
	if idx, ok := columnMap["email"]; ok && idx < len(record) {
		email := strings.TrimSpace(strings.ToLower(record[idx]))
		if isValidEmail(email) {
			customer.Email = email
		}
	}

	return customer
}

func setTenantContext(db *sql.DB, tenantID string) error {
	_, err := db.Exec("SELECT set_tenant_context($1)", tenantID)
	return err
}

func cleanPhoneNumber(phone string) string {
	if phone == "" {
		return ""
	}
	// Basic phone cleaning - remove extra characters but keep structure
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	return strings.TrimSpace(phone)
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Basic email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) > 5
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func writeLogFile(filename, environment string, stats *MigrationStats, config MigrationConfig) {
	content := fmt.Sprintf(`Migration Log - %s Environment
========================================
Start Time: %s
End Time: %s
Duration: %v

Configuration:
- Tenant ID: %s
- Data Path: %s
- Batch Size: %d
- Dry Run: %t

Results:
- Tables Processed: %d
- Records Inserted: %d
- Records Skipped: %d
- Errors: %d

Status: %s
`,
		strings.Title(environment),
		stats.StartTime.Format(time.RFC3339),
		stats.EndTime.Format(time.RFC3339),
		stats.EndTime.Sub(stats.StartTime),
		config.TenantID,
		config.DataPath,
		config.BatchSize,
		config.DryRun,
		stats.TablesProcessed,
		stats.RecordsInserted,
		stats.RecordsSkipped,
		stats.ErrorCount,
		func() string {
			if stats.ErrorCount == 0 {
				return "SUCCESS"
			}
			return "COMPLETED WITH ERRORS"
		}(),
	)

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		log.Printf("Failed to write log file %s: %v", filename, err)
	} else {
		log.Printf("Migration log written to: %s", filename)
	}
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

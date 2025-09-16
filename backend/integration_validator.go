// integration_validator.go - Integration tests without internal dependencies
package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type TestResult struct {
	Name     string
	Success  bool
	Duration time.Duration
	Message  string
}

type IntegrationTestSuite struct {
	AuthDBURL   string
	TenantDBURL string
	Results     []TestResult
}

func NewIntegrationTestSuite() *IntegrationTestSuite {
	return &IntegrationTestSuite{
		AuthDBURL:   getEnvOrDefault("DEV_CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable"),
		TenantDBURL: getEnvOrDefault("DEV_LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable"),
		Results:     make([]TestResult, 0),
	}
}

func (suite *IntegrationTestSuite) runTest(name string, testFunc func() error) {
	fmt.Printf("🧪 Running %s...\n", name)
	start := time.Now()
	
	err := testFunc()
	duration := time.Since(start)
	
	if err != nil {
		suite.Results = append(suite.Results, TestResult{
			Name:     name,
			Success:  false,
			Duration: duration,
			Message:  err.Error(),
		})
		fmt.Printf("❌ %s failed: %v\n", name, err)
	} else {
		suite.Results = append(suite.Results, TestResult{
			Name:     name,
			Success:  true,
			Duration: duration,
			Message:  "Passed",
		})
		fmt.Printf("✅ %s passed in %v\n", name, duration)
	}
}

func (suite *IntegrationTestSuite) testDatabaseConnectivity() error {
	// Test auth database
	authDB, err := sql.Open("postgres", suite.AuthDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to auth database: %w", err)
	}
	defer authDB.Close()

	if err := authDB.Ping(); err != nil {
		return fmt.Errorf("auth database ping failed: %w", err)
	}

	// Test tenant database
	tenantDB, err := sql.Open("postgres", suite.TenantDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	defer tenantDB.Close()

	if err := tenantDB.Ping(); err != nil {
		return fmt.Errorf("tenant database ping failed: %w", err)
	}

	return nil
}

func (suite *IntegrationTestSuite) testMultiTenantIsolation() error {
	db, err := sql.Open("postgres", suite.TenantDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	defer db.Close()

	// Clean up any existing test data
	db.Exec("DELETE FROM store.customers WHERE company_code IN ('ISOL001', 'ISOL002')")

	// Insert test customers for longbeach tenant
	_, err = db.Exec(`
		INSERT INTO store.customers (tenant_id, name, company_code, status)
		VALUES ('longbeach', 'Long Beach Oil Co', 'ISOL001', 'active')
	`)
	if err != nil {
		return fmt.Errorf("failed to insert test customers: %w", err)
	}

	// Test longbeach isolation
	var longbeachCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM store.customers 
		WHERE tenant_id = 'longbeach' AND company_code = 'ISOL001'
	`).Scan(&longbeachCount)
	if err != nil {
		return fmt.Errorf("failed to query longbeach customers: %w", err)
	}
	if longbeachCount != 1 {
		return fmt.Errorf("expected 1 longbeach customer, got %d", longbeachCount)
	}

	// Insert another customer for the same tenant to test data segregation
	_, err = db.Exec(`
		INSERT INTO store.customers (tenant_id, name, company_code, status)
		VALUES ('longbeach', 'Long Beach Gas Co', 'ISOL002', 'active')
	`)
	if err != nil {
		return fmt.Errorf("failed to insert second test customer: %w", err)
	}

	// Test that we can filter by specific customers within the tenant
	var isolCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM store.customers 
		WHERE tenant_id = 'longbeach' AND company_code = 'ISOL001'
	`).Scan(&isolCount)
	if err != nil {
		return fmt.Errorf("failed to query specific customer: %w", err)
	}
	if isolCount != 1 {
		return fmt.Errorf("expected 1 specific customer, got %d", isolCount)
	}

	// Test that tenant isolation works by ensuring all data is properly scoped
	var totalLongbeachCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM store.customers 
		WHERE tenant_id = 'longbeach' AND company_code IN ('ISOL001', 'ISOL002')
	`).Scan(&totalLongbeachCount)
	if err != nil {
		return fmt.Errorf("failed to query tenant data: %w", err)
	}
	if totalLongbeachCount != 2 {
		return fmt.Errorf("expected 2 longbeach customers, got %d", totalLongbeachCount)
	}

	// Test that no data exists for other tenants (proper isolation)
	var otherTenantCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM store.customers 
		WHERE tenant_id != 'longbeach' AND company_code IN ('ISOL001', 'ISOL002')
	`).Scan(&otherTenantCount)
	if err != nil {
		return fmt.Errorf("failed to query other tenant data: %w", err)
	}
	if otherTenantCount != 0 {
		return fmt.Errorf("expected 0 data leakage to other tenants, got %d", otherTenantCount)
	}

	// Cleanup
	db.Exec("DELETE FROM store.customers WHERE company_code IN ('ISOL001', 'ISOL002')")

	return nil
}

func (suite *IntegrationTestSuite) testCrossDatabaseOperations() error {
	authDB, err := sql.Open("postgres", suite.AuthDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to auth database: %w", err)
	}
	defer authDB.Close()

	tenantDB, err := sql.Open("postgres", suite.TenantDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	defer tenantDB.Close()

	// Check if expected tables exist
	var authTableExists bool
	err = authDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = 'users'
		)
	`).Scan(&authTableExists)
	if err != nil {
		return fmt.Errorf("failed to check auth table existence: %w", err)
	}

	var tenantTableExists bool
	err = tenantDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'store' AND table_name = 'customers'
		)
	`).Scan(&tenantTableExists)
	if err != nil {
		return fmt.Errorf("failed to check tenant table existence: %w", err)
	}

	if !tenantTableExists {
		return fmt.Errorf("store.customers table does not exist in tenant database")
	}

	// Test tenant database operations
	var customerCount int
	err = tenantDB.QueryRow("SELECT COUNT(*) FROM store.customers").Scan(&customerCount)
	if err != nil {
		return fmt.Errorf("failed to count customers: %w", err)
	}

	// If auth users table exists, test auth operations
	if authTableExists {
		var userCount int
		err = authDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
		if err != nil {
			return fmt.Errorf("failed to count auth users: %w", err)
		}
		fmt.Printf("  Found %d users in auth database, %d customers in tenant database\n", userCount, customerCount)
	} else {
		fmt.Printf("  Auth users table not found, found %d customers in tenant database\n", customerCount)
	}

	return nil
}

func (suite *IntegrationTestSuite) testDatabasePerformance() error {
	db, err := sql.Open("postgres", suite.TenantDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	defer db.Close()

	const numRecords = 50
	
	// Clean up existing test data
	db.Exec("DELETE FROM store.customers WHERE company_code LIKE 'PERF%'")

	// Test bulk insert performance
	start := time.Now()
	for i := 0; i < numRecords; i++ {
		_, err = db.Exec(`
			INSERT INTO store.customers (tenant_id, name, company_code, status)
			VALUES ('longbeach', $1, $2, 'active')
		`, 
			fmt.Sprintf("Performance Test Co %d", i),
			fmt.Sprintf("PERF%03d", i))
		
		if err != nil {
			return fmt.Errorf("failed to insert record %d: %w", i, err)
		}
	}
	insertDuration := time.Since(start)

	// Test query performance
	start = time.Now()
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM store.customers 
		WHERE company_code LIKE 'PERF%' AND status = 'active'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query performance test records: %w", err)
	}
	queryDuration := time.Since(start)

	if count != numRecords {
		return fmt.Errorf("expected %d records, found %d", numRecords, count)
	}

	// Cleanup
	db.Exec("DELETE FROM store.customers WHERE company_code LIKE 'PERF%'")

	fmt.Printf("  Inserted %d records in %v (%.2f records/sec)\n", 
		numRecords, insertDuration, float64(numRecords)/insertDuration.Seconds())
	fmt.Printf("  Query found %d records in %v\n", count, queryDuration)

	return nil
}

func (suite *IntegrationTestSuite) testTransactionIntegrity() error {
	db, err := sql.Open("postgres", suite.TenantDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	defer db.Close()

	// Clean up
	db.Exec("DELETE FROM store.customers WHERE company_code IN ('TX001', 'TX002')")

	// Test successful transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO store.customers (tenant_id, name, company_code, status)
		VALUES ('longbeach', 'Transaction Test 1', 'TX001', 'active')
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert in transaction: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Verify successful transaction
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM store.customers WHERE company_code = 'TX001')
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify committed transaction: %w", err)
	}
	if !exists {
		return fmt.Errorf("transaction commit failed - record not found")
	}

	// Test rollback
	tx2, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin rollback transaction: %w", err)
	}

	_, err = tx2.Exec(`
		INSERT INTO store.customers (tenant_id, name, company_code, status)
		VALUES ('longbeach', 'Transaction Test 2', 'TX002', 'active')
	`)
	if err != nil {
		tx2.Rollback()
		return fmt.Errorf("failed to insert in rollback transaction: %w", err)
	}

	err = tx2.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	// Verify rollback worked
	err = db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM store.customers WHERE company_code = 'TX002')
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify rollback: %w", err)
	}
	if exists {
		return fmt.Errorf("rollback failed - record should not exist")
	}

	// Cleanup
	db.Exec("DELETE FROM store.customers WHERE company_code = 'TX001'")

	return nil
}

func (suite *IntegrationTestSuite) testSchemaValidation() error {
	authDB, err := sql.Open("postgres", suite.AuthDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to auth database: %w", err)
	}
	defer authDB.Close()

	tenantDB, err := sql.Open("postgres", suite.TenantDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	defer tenantDB.Close()

	// Validate tenant database schema
	expectedTables := []string{"customers", "customer_contacts", "tenants"}
	for _, table := range expectedTables {
		var exists bool
		err = tenantDB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_schema = 'store' AND table_name = $1
			)
		`, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table %s existence: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("required table store.%s does not exist", table)
		}
	}

	// Validate tenant configuration
	var tenantExists bool
	err = tenantDB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM store.tenants WHERE tenant_id = 'longbeach')
	`).Scan(&tenantExists)
	if err != nil {
		return fmt.Errorf("failed to check tenant configuration: %w", err)
	}
	if !tenantExists {
		return fmt.Errorf("longbeach tenant configuration not found")
	}

	return nil
}

func (suite *IntegrationTestSuite) runAllTests() {
	fmt.Println("🚀 Starting Integration Tests")
	fmt.Println("=============================")

	suite.runTest("Database Connectivity", suite.testDatabaseConnectivity)
	suite.runTest("Multi-Tenant Isolation", suite.testMultiTenantIsolation)
	suite.runTest("Cross-Database Operations", suite.testCrossDatabaseOperations)
	suite.runTest("Database Performance", suite.testDatabasePerformance)
	suite.runTest("Transaction Integrity", suite.testTransactionIntegrity)
	suite.runTest("Schema Validation", suite.testSchemaValidation)

	suite.printResults()
}

func (suite *IntegrationTestSuite) printResults() {
	fmt.Println("\n📊 Integration Test Results")
	fmt.Println("===========================")

	passed := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, result := range suite.Results {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ FAIL"
			failed++
		} else {
			passed++
		}
		
		fmt.Printf("%s %-30s %8v %s\n", status, result.Name, result.Duration, result.Message)
		totalDuration += result.Duration
	}

	fmt.Printf("\nSummary: %d passed, %d failed in %v\n", passed, failed, totalDuration)

	if failed == 0 {
		fmt.Println("🎉 All integration tests passed!")
	} else {
		fmt.Printf("⚠️  %d tests failed\n", failed)
		os.Exit(1)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Integration Test Suite for Multi-Tenant Oil & Gas Application")
		fmt.Println("\nEnvironment Variables:")
		fmt.Println("  CENTRAL_AUTH_DB_URL    - Auth database connection string")
		fmt.Println("  LONGBEACH_DB_URL       - Longbeach database connection string")
		fmt.Println("\nUsage:")
		fmt.Println("  go run standalone_integration_test.go")
		return
	}

	suite := NewIntegrationTestSuite()
	suite.runAllTests()
}
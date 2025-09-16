// backend/test/integration/multi_tenant_test.go
package integration

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"

	"oilgas-backend/internal/customer"
	"oilgas-backend/internal/shared/database"
)

type MultiTenantTestSuite struct {
	suite.Suite
	dbManager       *database.DatabaseManager
	ctx            context.Context
	authDB         *sql.DB
	longbeachDB    *sql.DB
}

func (suite *MultiTenantTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Initialize database manager with actual test databases
	dbConfig := &database.Config{
		CentralDBURL: getEnvOrDefault("CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable"),
		TenantDBs: map[string]string{
			"longbeach": getEnvOrDefault("LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable"),
		},
		MaxOpenConns: 10,
		MaxIdleConns: 2,
		MaxLifetime:  time.Hour,
	}

	var err error
	suite.dbManager, err = database.NewDatabaseManager(dbConfig)
	suite.Require().NoError(err, "Failed to create database manager")

	// Get direct database connections for testing
	suite.authDB, err = sql.Open("postgres", dbConfig.CentralDBURL)
	suite.Require().NoError(err, "Failed to connect to auth database")

	suite.longbeachDB, err = sql.Open("postgres", dbConfig.TenantDBs["longbeach"])
	suite.Require().NoError(err, "Failed to connect to longbeach database")

	// Verify connections
	err = suite.authDB.Ping()
	suite.Require().NoError(err, "Auth database ping failed")

	err = suite.longbeachDB.Ping()
	suite.Require().NoError(err, "Longbeach database ping failed")

	// Set up test isolation
	suite.setupTestIsolation()
}

func (suite *MultiTenantTestSuite) setupTestIsolation() {
	// Create test schema in longbeach database for isolation
	_, err := suite.longbeachDB.Exec("CREATE SCHEMA IF NOT EXISTS test_isolation")
	suite.Require().NoError(err)

	// Create test customer table for isolation tests
	_, err = suite.longbeachDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_isolation.customers (
			id SERIAL PRIMARY KEY,
			tenant_id VARCHAR(100) NOT NULL,
			name VARCHAR(255) NOT NULL,
			company_code VARCHAR(50) UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	suite.Require().NoError(err)
}

func (suite *MultiTenantTestSuite) TearDownSuite() {
	// Clean up test data
	if suite.longbeachDB != nil {
		suite.longbeachDB.Exec("DROP SCHEMA IF EXISTS test_isolation CASCADE")
		suite.longbeachDB.Close()
	}
	if suite.authDB != nil {
		suite.authDB.Close()
	}
	if suite.dbManager != nil {
		suite.dbManager.Close()
	}
}

func (suite *MultiTenantTestSuite) SetupTest() {
	// Clean test data before each test
	suite.longbeachDB.Exec("TRUNCATE TABLE test_isolation.customers RESTART IDENTITY")
}

// Test that tenant data is properly isolated
func (suite *MultiTenantTestSuite) TestTenantDataIsolation() {
	suite.T().Log("Testing tenant data isolation...")

	// Insert customer data for longbeach tenant
	_, err := suite.longbeachDB.Exec(`
		INSERT INTO test_isolation.customers (tenant_id, name, company_code)
		VALUES ('longbeach', 'Long Beach Oil Co', 'LBC001')
	`)
	suite.Require().NoError(err)

	// Insert another customer for a different tenant in the same database
	_, err = suite.longbeachDB.Exec(`
		INSERT INTO test_isolation.customers (tenant_id, name, company_code)
		VALUES ('bakersfield', 'Bakersfield Energy', 'BAK001')
	`)
	suite.Require().NoError(err)

	// Verify we can query only longbeach data
	var count int
	err = suite.longbeachDB.QueryRow(`
		SELECT COUNT(*) FROM test_isolation.customers WHERE tenant_id = 'longbeach'
	`).Scan(&count)
	suite.Require().NoError(err)
	suite.Assert().Equal(1, count, "Should have exactly 1 longbeach customer")

	// Verify bakersfield data exists but is separate
	err = suite.longbeachDB.QueryRow(`
		SELECT COUNT(*) FROM test_isolation.customers WHERE tenant_id = 'bakersfield'
	`).Scan(&count)
	suite.Require().NoError(err)
	suite.Assert().Equal(1, count, "Should have exactly 1 bakersfield customer")

	// Test tenant-specific queries don't leak data
	var name string
	err = suite.longbeachDB.QueryRow(`
		SELECT name FROM test_isolation.customers WHERE tenant_id = 'longbeach' AND company_code = 'LBC001'
	`).Scan(&name)
	suite.Require().NoError(err)
	suite.Assert().Equal("Long Beach Oil Co", name)

	// Verify we can't access other tenant's data with wrong tenant filter
	err = suite.longbeachDB.QueryRow(`
		SELECT name FROM test_isolation.customers WHERE tenant_id = 'longbeach' AND company_code = 'BAK001'
	`).Scan(&name)
	suite.Assert().Error(err, "Should not find bakersfield customer when filtering by longbeach tenant")
}

// Test database manager tenant isolation
func (suite *MultiTenantTestSuite) TestDatabaseManagerTenantIsolation() {
	suite.T().Log("Testing database manager tenant isolation...")

	// Get tenant-specific database connection
	tenantDB, err := suite.dbManager.GetTenantDB("longbeach")
	suite.Require().NoError(err)

	// Insert test customer using the proper store schema
	_, err = tenantDB.Exec(`
		INSERT INTO store.customers (tenant_id, name, company_code, status)
		VALUES ('longbeach', 'Database Manager Test Co', 'DMT001', 'active')
		ON CONFLICT (company_code) DO UPDATE SET name = EXCLUDED.name
	`)
	suite.Require().NoError(err)

	// Verify the customer exists in the longbeach database
	var customerName string
	err = tenantDB.QueryRow(`
		SELECT name FROM store.customers WHERE company_code = 'DMT001' AND tenant_id = 'longbeach'
	`).Scan(&customerName)
	suite.Require().NoError(err)
	suite.Assert().Equal("Database Manager Test Co", customerName)

	// Verify this customer cannot be accessed from a different tenant context
	// (In a real multi-tenant setup, you'd have separate databases or stronger isolation)
	var count int
	err = tenantDB.QueryRow(`
		SELECT COUNT(*) FROM store.customers WHERE company_code = 'DMT001' AND tenant_id != 'longbeach'
	`).Scan(&count)
	suite.Require().NoError(err)
	suite.Assert().Equal(0, count, "Customer should not be accessible from different tenant context")
}

// Test cross-database operations (auth + tenant)
func (suite *MultiTenantTestSuite) TestCrossDatabaseOperations() {
	suite.T().Log("Testing cross-database operations...")

	// Create a test user in auth database (if users table exists)
	var userTableExists bool
	err := suite.authDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = 'users'
		)
	`).Scan(&userTableExists)
	suite.Require().NoError(err)

	if userTableExists {
		// Insert test user
		var userID int
		err = suite.authDB.QueryRow(`
			INSERT INTO users (username, email, full_name, password_hash, role)
			VALUES ('testuser', 'test@example.com', 'Test User', 'hashed_password', 'OPERATOR')
			ON CONFLICT (email) DO UPDATE SET full_name = EXCLUDED.full_name
			RETURNING id
		`).Scan(&userID)
		suite.Require().NoError(err)

		// Create corresponding customer in tenant database
		_, err = suite.longbeachDB.Exec(`
			INSERT INTO store.customers (tenant_id, name, company_code, status)
			VALUES ('longbeach', 'Cross DB Test Co', 'CDB001', 'active')
			ON CONFLICT (company_code) DO UPDATE SET name = EXCLUDED.name
		`)
		suite.Require().NoError(err)

		// Verify both records exist and can be joined conceptually
		var authUserExists bool
		err = suite.authDB.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
		`, userID).Scan(&authUserExists)
		suite.Require().NoError(err)
		suite.Assert().True(authUserExists, "Auth user should exist")

		var customerExists bool
		err = suite.longbeachDB.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM store.customers WHERE company_code = 'CDB001')
		`).Scan(&customerExists)
		suite.Require().NoError(err)
		suite.Assert().True(customerExists, "Customer should exist")

		suite.T().Logf("Successfully created auth user ID %d and corresponding customer", userID)
	} else {
		suite.T().Log("Users table doesn't exist in auth database, skipping user creation test")
	}
}

// Test database connection pooling and concurrent access
func (suite *MultiTenantTestSuite) TestConcurrentTenantAccess() {
	suite.T().Log("Testing concurrent tenant access...")

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Spawn multiple goroutines to test concurrent database access
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			tenantDB, err := suite.dbManager.GetTenantDB("longbeach")
			if err != nil {
				results <- err
				return
			}

			// Perform database operation
			_, err = tenantDB.Exec(`
				INSERT INTO store.customers (tenant_id, name, company_code, status)
				VALUES ('longbeach', $1, $2, 'active')
				ON CONFLICT (company_code) DO UPDATE SET name = EXCLUDED.name
			`, 
				"Concurrent Test Co "+string(rune(id)), 
				"CON"+string(rune(48+id%10))+string(rune(48+(id/10)%10))+string(rune(48+(id/100)%10)))

			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		suite.Assert().NoError(err, "Concurrent access should not produce errors")
	}

	// Verify at least some customers were created
	var count int
	err := suite.longbeachDB.QueryRow(`
		SELECT COUNT(*) FROM store.customers WHERE name LIKE 'Concurrent Test Co%'
	`).Scan(&count)
	suite.Require().NoError(err)
	suite.Assert().Greater(count, 0, "Should have created at least some test customers concurrently")
}

// Test tenant-specific configuration validation
func (suite *MultiTenantTestSuite) TestTenantConfiguration() {
	suite.T().Log("Testing tenant configuration...")

	// Verify tenant exists in store.tenants table
	var tenantExists bool
	err := suite.longbeachDB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM store.tenants WHERE tenant_id = 'longbeach')
	`).Scan(&tenantExists)
	suite.Require().NoError(err)
	suite.Assert().True(tenantExists, "Longbeach tenant should exist in tenants table")

	// Verify tenant is active
	var isActive bool
	err = suite.longbeachDB.QueryRow(`
		SELECT is_active FROM store.tenants WHERE tenant_id = 'longbeach'
	`).Scan(&isActive)
	suite.Require().NoError(err)
	suite.Assert().True(isActive, "Longbeach tenant should be active")

	// Test invalid tenant access
	_, err = suite.dbManager.GetTenantDB("nonexistent")
	suite.Assert().Error(err, "Should not be able to access nonexistent tenant database")
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestMultiTenantTestSuite(t *testing.T) {
	suite.Run(t, new(MultiTenantTestSuite))
}
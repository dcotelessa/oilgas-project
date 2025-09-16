// backend/test/integration/integration_test_runner.go
package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// IntegrationTestRunner orchestrates all integration tests
type IntegrationTestRunner struct {
	suite.Suite
	fixture *DatabaseFixture
	ctx     context.Context
}

func (suite *IntegrationTestRunner) SetupSuite() {
	suite.ctx = context.Background()
	
	log.Println("🚀 Starting Integration Test Suite")
	log.Println("==================================")

	// Create database fixture
	var err error
	suite.fixture, err = NewDatabaseFixture()
	suite.Require().NoError(err, "Failed to create database fixture")

	// Set up test schemas and seed data
	err = suite.fixture.SetupTestSchemas()
	suite.Require().NoError(err, "Failed to set up test schemas")

	err = suite.fixture.SeedTestData()
	suite.Require().NoError(err, "Failed to seed test data")

	log.Println("✅ Integration test environment ready")
}

func (suite *IntegrationTestRunner) TearDownSuite() {
	if suite.fixture != nil {
		suite.fixture.DropTestSchemas()
		suite.fixture.Close()
	}
	log.Println("🧹 Integration test cleanup completed")
}

// TestDatabaseConnectivity validates basic database connectivity
func (suite *IntegrationTestRunner) TestDatabaseConnectivity() {
	suite.T().Log("Testing database connectivity...")

	err := suite.fixture.VerifyConnections()
	suite.Assert().NoError(err, "All database connections should be working")

	// Test database manager
	tenantDB, err := suite.fixture.DBManager.GetTenantDB("longbeach")
	suite.Require().NoError(err)

	var result string
	err = tenantDB.QueryRow("SELECT 'Connection OK'").Scan(&result)
	suite.Require().NoError(err)
	suite.Assert().Equal("Connection OK", result)
}

// TestMultiTenantIsolationBasic tests basic tenant isolation
func (suite *IntegrationTestRunner) TestMultiTenantIsolationBasic() {
	suite.T().Log("Testing basic multi-tenant isolation...")

	// Create tenant-specific data
	customerID, err := suite.fixture.CreateTestCustomer("Isolation Test Co", "ISO001", "active")
	suite.Require().NoError(err)

	// Verify data exists for correct tenant
	var count int
	err = suite.fixture.TenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_integration.customers 
		WHERE tenant_id = 'longbeach' AND id = $1
	`, customerID).Scan(&count)
	suite.Require().NoError(err)
	suite.Assert().Equal(1, count, "Customer should exist in longbeach tenant")

	// Verify data doesn't leak to other tenants
	err = suite.fixture.TenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_integration.customers 
		WHERE tenant_id != 'longbeach' AND id = $1
	`, customerID).Scan(&count)
	suite.Require().NoError(err)
	suite.Assert().Equal(0, count, "Customer should not exist in other tenants")
}

// TestAuthIntegrationBasic tests basic auth integration
func (suite *IntegrationTestRunner) TestAuthIntegrationBasic() {
	suite.T().Log("Testing basic auth integration...")

	// Create test user
	userID, err := suite.fixture.CreateTestUser("integration_test", "integration@test.com", "Integration Test User", "OPERATOR")
	suite.Require().NoError(err)

	// Grant tenant access
	err = suite.fixture.GrantTenantAccess(userID, "longbeach", "OPERATOR", []string{"view", "edit"})
	suite.Require().NoError(err)

	// Verify user has correct access
	var hasAccess bool
	err = suite.fixture.AuthDB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM test_integration.user_tenant_access
			WHERE user_id = $1 AND tenant_id = 'longbeach' AND is_active = true
		)
	`, userID).Scan(&hasAccess)
	suite.Require().NoError(err)
	suite.Assert().True(hasAccess, "User should have access to longbeach tenant")
}

// TestCrossDatabaseOperations tests operations across both databases
func (suite *IntegrationTestRunner) TestCrossDatabaseOperations() {
	suite.T().Log("Testing cross-database operations...")

	// Create user in auth database
	userID, err := suite.fixture.CreateTestUser("cross_db_user", "crossdb@test.com", "Cross DB User", "CUSTOMER_CONTACT")
	suite.Require().NoError(err)

	// Create customer in tenant database
	customerID, err := suite.fixture.CreateTestCustomer("Cross DB Test Co", "CDB001", "active")
	suite.Require().NoError(err)

	// Link them together
	err = suite.fixture.LinkCustomerContact(customerID, userID, "PRIMARY", true)
	suite.Require().NoError(err)

	// Verify the relationship exists
	var contactExists bool
	err = suite.fixture.TenantDB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM test_integration.customer_contacts
			WHERE customer_id = $1 AND auth_user_id = $2
		)
	`, customerID, userID).Scan(&contactExists)
	suite.Require().NoError(err)
	suite.Assert().True(contactExists, "Customer contact relationship should exist")

	// Verify we can join across conceptual databases
	var customerName, userName string
	err = suite.fixture.TenantDB.QueryRow(`
		SELECT c.name, cc.full_name
		FROM test_integration.customers c
		JOIN test_integration.customer_contacts cc ON c.id = cc.customer_id
		WHERE c.id = $1 AND cc.auth_user_id = $2
	`, customerID, userID).Scan(&customerName, &userName)
	suite.Require().NoError(err)
	suite.Assert().Equal("Cross DB Test Co", customerName)
}

// TestDataConsistency tests data consistency across databases
func (suite *IntegrationTestRunner) TestDataConsistency() {
	suite.T().Log("Testing data consistency across databases...")

	start := time.Now()

	// Perform multiple operations that should maintain consistency
	const numOperations = 10
	for i := 0; i < numOperations; i++ {
		userID, err := suite.fixture.CreateTestUser(
			fmt.Sprintf("consistency_user_%d", i),
			fmt.Sprintf("consistency%d@test.com", i),
			fmt.Sprintf("Consistency User %d", i),
			"OPERATOR")
		suite.Require().NoError(err)

		err = suite.fixture.GrantTenantAccess(userID, "longbeach", "OPERATOR", []string{"view"})
		suite.Require().NoError(err)

		customerID, err := suite.fixture.CreateTestCustomer(
			fmt.Sprintf("Consistency Co %d", i),
			fmt.Sprintf("CON%03d", i),
			"active")
		suite.Require().NoError(err)

		err = suite.fixture.LinkCustomerContact(customerID, userID, "PRIMARY", true)
		suite.Require().NoError(err)
	}

	operationDuration := time.Since(start)
	suite.T().Logf("Created %d consistent user-customer pairs in %v", numOperations, operationDuration)

	// Verify consistency
	var authUserCount, tenantCustomerCount, linkCount int

	err := suite.fixture.AuthDB.QueryRow(`
		SELECT COUNT(*) FROM test_integration.users WHERE username LIKE 'consistency_user_%'
	`).Scan(&authUserCount)
	suite.Require().NoError(err)

	err = suite.fixture.TenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_integration.customers WHERE company_code LIKE 'CON%'
	`).Scan(&tenantCustomerCount)
	suite.Require().NoError(err)

	err = suite.fixture.TenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_integration.customer_contacts cc
		JOIN test_integration.customers c ON cc.customer_id = c.id
		WHERE c.company_code LIKE 'CON%'
	`).Scan(&linkCount)
	suite.Require().NoError(err)

	suite.Assert().Equal(numOperations, authUserCount, "Should have created all auth users")
	suite.Assert().Equal(numOperations, tenantCustomerCount, "Should have created all customers")
	suite.Assert().Equal(numOperations, linkCount, "Should have created all customer-user links")
}

// TestPerformanceUnderLoad tests system performance under load
func (suite *IntegrationTestRunner) TestPerformanceUnderLoad() {
	suite.T().Log("Testing performance under load...")

	const numGoroutines = 5
	const operationsPerGoroutine = 20

	start := time.Now()
	results := make(chan error, numGoroutines*operationsPerGoroutine)

	// Spawn goroutines to create load
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			for j := 0; j < operationsPerGoroutine; j++ {
				// Create customer
				_, err := suite.fixture.CreateTestCustomer(
					fmt.Sprintf("Load Test Co %d-%d", workerID, j),
					fmt.Sprintf("LT%d%02d", workerID, j),
					"active")
				results <- err
			}
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < numGoroutines*operationsPerGoroutine; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	totalDuration := time.Since(start)
	totalOperations := numGoroutines * operationsPerGoroutine

	suite.Assert().Empty(errors, "No errors should occur during load test")
	suite.T().Logf("Completed %d operations in %v (%.2f ops/sec)",
		totalOperations, totalDuration, float64(totalOperations)/totalDuration.Seconds())

	// Verify all operations completed
	var loadTestCount int
	err := suite.fixture.TenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_integration.customers WHERE company_code LIKE 'LT%'
	`).Scan(&loadTestCount)
	suite.Require().NoError(err)
	suite.Assert().Equal(totalOperations, loadTestCount, "All load test customers should be created")
}

// RunAllIntegrationTests runs all integration test suites
func RunAllIntegrationTests(t *testing.T) {
	// Check if integration tests should be skipped
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration tests (SKIP_INTEGRATION_TESTS=true)")
	}

	// Run the main integration test runner
	suite.Run(t, new(IntegrationTestRunner))

	// Run specialized test suites
	suite.Run(t, new(MultiTenantTestSuite))
	suite.Run(t, new(AuthIntegrationTestSuite))
	suite.Run(t, new(DatabaseOperationsTestSuite))
}

func TestAllIntegrationTests(t *testing.T) {
	RunAllIntegrationTests(t)
}
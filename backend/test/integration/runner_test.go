// backend/test/integration/runner_test.go
package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"oilgas-backend/test/testutil"
)

// IntegrationRunner orchestrates all integration test suites
type IntegrationRunner struct {
	suite.Suite
	testDB *testutil.TestDB
}

func (r *IntegrationRunner) SetupSuite() {
	// Ensure we're running in test mode
	if os.Getenv("APP_ENV") != "test" && os.Getenv("TEST_DATABASE_URL") == "" {
		r.T().Skip("Integration tests require TEST_DATABASE_URL or APP_ENV=test")
	}
	
	// Setup test database once for all suites
	r.testDB = testutil.SetupTestDB(r.T())
	
	// Verify database is clean and ready
	r.testDB.Truncate(r.T())
	r.testDB.SeedGrades(r.T())
}

func (r *IntegrationRunner) TearDownSuite() {
	if r.testDB != nil {
		testutil.CleanupTestDB(r.T(), r.testDB)
	}
}

// TestAllIntegrationSuites runs all integration test suites in order
func (r *IntegrationRunner) TestAllIntegrationSuites() {
	r.T().Log("Starting comprehensive integration test suite")
	
	// Run repository integration tests first (foundation)
	r.T().Run("RepositoryIntegration", func(t *testing.T) {
		suite.Run(t, new(IntegrationTestSuite))
	})
	
	// Run service integration tests (business logic)
	r.T().Run("ServiceIntegration", func(t *testing.T) {
		suite.Run(t, new(ServiceIntegrationTestSuite))
	})
	
	// Run API integration tests (HTTP layer)
	r.T().Run("APIIntegration", func(t *testing.T) {
		suite.Run(t, new(APIIntegrationTestSuite))
	})
	
	r.T().Log("All integration test suites completed successfully")
}

func TestIntegrationRunner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive integration tests in short mode")
	}
	
	suite.Run(t, new(IntegrationRunner))
}

// backend/test/integration/setup_test.go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/suite"
    "oilgas-backend/internal/config"
    "oilgas-backend/internal/shared/database"
    "oilgas-backend/internal/shared/events"
)

type IntegrationTestSuite struct {
    suite.Suite
    tenantManager *database.TenantManager
    eventBus      *events.EventBus
    ctx           context.Context
}

func (suite *IntegrationTestSuite) SetupSuite() {
    // Load test configuration
    cfg := config.NewTestConfig()
    
    // Initialize tenant manager
    tm, err := database.NewTenantManager(cfg.DatabaseConfig)
    suite.Require().NoError(err)
    suite.tenantManager = tm
    
    // Initialize event bus
    eventStore := events.NewDatabaseEventStore(tm.GetAuthDB())
    suite.eventBus = events.NewEventBus(eventStore)
    
    suite.ctx = context.Background()
    
    // Run migrations for all databases
    suite.runMigrations()
}

func (suite *IntegrationTestSuite) runMigrations() {
    // Run auth database migrations
    authDB := suite.tenantManager.GetAuthDB()
    suite.runAuthMigrations(authDB)
    
    // Run tenant database migrations
    for tenantID := range suite.tenantManager.GetAllTenantDBs() {
        tenantDB, err := suite.tenantManager.GetTenantDB(tenantID)
        suite.Require().NoError(err)
        suite.runTenantMigrations(tenantDB, tenantID)
    }
}

func (suite *IntegrationTestSuite) TearDownSuite() {
    // Clean up test data
    suite.cleanupTestData()
}

// Test multi-tenant isolation
func (suite *IntegrationTestSuite) TestMultiTenantIsolation() {
    // Create customer in Houston
    houstonDB, err := suite.tenantManager.GetTenantDB("houston")
    suite.Require().NoError(err)
    
    houstonCustomer := createTestCustomer(suite.ctx, houstonDB, "Houston Oil Co")
    suite.Require().NotNil(houstonCustomer)
    
    // Verify customer doesn't appear in Dallas
    dallasDB, err := suite.tenantManager.GetTenantDB("dallas")  
    suite.Require().NoError(err)
    
    customers := getAllCustomers(suite.ctx, dallasDB)
    suite.Require().NotContains(customers, houstonCustomer.ID)
}

// Test cross-tenant admin access
func (suite *IntegrationTestSuite) TestCrossTenantAdminAccess() {
    // Create admin user with access to both tenants
    adminUser := suite.createAdminUser("admin@company.com", []string{"houston", "dallas"})
    
    // Verify admin can access both tenant databases
    err := suite.tenantManager.ValidateTenantAccess(suite.ctx, adminUser.ID, "houston")
    suite.Require().NoError(err)
    
    err = suite.tenantManager.ValidateTenantAccess(suite.ctx, adminUser.ID, "dallas") 
    suite.Require().NoError(err)
}

// Test event-driven communication
func (suite *IntegrationTestSuite) TestEventDrivenWorkflow() {
    // Set up event handlers
    eventsReceived := make([]events.Event, 0)
    suite.eventBus.Subscribe("workorder.created", func(ctx context.Context, event events.Event) error {
        eventsReceived = append(eventsReceived, event)
        return nil
    })
    
    // Create work order (should emit event)
    workOrder := suite.createTestWorkOrder("houston", 1)
    
    // Wait for async event processing
    time.Sleep(100 * time.Millisecond)
    
    // Verify event was received
    suite.Require().Len(eventsReceived, 1)
    suite.Require().Equal("workorder.created", eventsReceived[0].EventType())
}

func TestIntegrationSuite(t *testing.T) {
    suite.Run(t, new(IntegrationTestSuite))
}

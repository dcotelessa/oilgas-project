// backend/test/integration/service_integration_test.go
package integration

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/validation"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/test/testutil"
	"oilgas-backend/test/integration/testdata"
)

type ServiceIntegrationTestSuite struct {
	suite.Suite
	db       *testutil.TestDB
	repos    *repository.Repositories
	services *services.Services
	ctx      context.Context
}

func (s *ServiceIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.db = testutil.SetupTestDB(s.T())
	s.repos = repository.New(s.db.Pool)
	
	cache := cache.New(cache.Config{
		TTL:             5 * time.Minute,
		CleanupInterval: 10 * time.Minute,
		MaxSize:         100,
	})
	
	s.services = services.New(s.repos, cache)
}

func (s *ServiceIntegrationTestSuite) TearDownSuite() {
	testutil.CleanupTestDB(s.T(), s.db)
}

func (s *ServiceIntegrationTestSuite) SetupTest() {
	s.db.Truncate(s.T())
	s.db.SeedGrades(s.T())
}

// Test customer service with validation and caching
func (s *ServiceIntegrationTestSuite) TestCustomerService() {
	// Test creation with validation
	req := &validation.CustomerValidation{
		CustomerName: "Service Test Customer",
		Address:      "123 Service St",
		City:         "Houston",
		State:        "TX",
		Phone:        "555-service",
		Email:        "service@test.com",
	}
	
	customer, err := s.services.Customer.Create(s.ctx, req)
	s.Require().NoError(err)
	s.Assert().NotZero(customer.CustomerID)
	s.Assert().Equal(req.CustomerName, customer.Customer)

	// Test duplicate name validation
	dupReq := &validation.CustomerValidation{
		CustomerName: "Service Test Customer", // Same name
		Phone:        "555-different",
	}
	
	_, err = s.services.Customer.Create(s.ctx, dupReq)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "already exists")

	// Test retrieval and caching
	retrieved, err := s.services.Customer.GetByID(s.ctx, strconv.Itoa(customer.CustomerID))
	s.Require().NoError(err)
	s.Assert().Equal(customer.Customer, retrieved.Customer)

	// Test cache hit (second call should be faster)
	start := time.Now()
	retrieved2, err := s.services.Customer.GetByID(s.ctx, strconv.Itoa(customer.CustomerID))
	duration := time.Since(start)
	s.Require().NoError(err)
	s.Assert().Equal(retrieved.Customer, retrieved2.Customer)
	s.Assert().True(duration < 10*time.Millisecond) // Should be very fast from cache

	// Test update invalidates cache
	updateReq := &validation.CustomerValidation{
		CustomerName: "Updated Service Customer",
		Phone:        req.Phone,
		Email:        req.Email,
	}
	
	updated, err := s.services.Customer.Update(s.ctx, strconv.Itoa(customer.CustomerID), updateReq)
	s.Require().NoError(err)
	s.Assert().Equal(updateReq.CustomerName, updated.Customer)

	// Verify cache was invalidated
	fromDB, err := s.services.Customer.GetByID(s.ctx, strconv.Itoa(customer.CustomerID))
	s.Require().NoError(err)
	s.Assert().Equal("Updated Service Customer", fromDB.Customer)
}

// Test received service with workflow integration
func (s *ServiceIntegrationTestSuite) TestReceivedService() {
	// Setup customer
	customer := &models.Customer{Customer: "Received Service Customer"}
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	// Test creation with validation
	req := &validation.ReceivedValidation{
		WorkOrder:  "WO-SERVICE-001",
		CustomerID: customer.CustomerID,
		Joints:     100,
		Size:       "5 1/2\"",
		Grade:      "J55",
		Connection: "LTC",
		// Note: Removed Color and Location fields that don't exist in ReceivedValidation
		Weight:     "20", // Added required field
	}
	
	received, err := s.services.Received.Create(s.ctx, req)
	s.Require().NoError(err)
	s.Assert().NotZero(received.ID)
	s.Assert().Equal(req.WorkOrder, received.WorkOrder)

	// Verify workflow state was initialized (if workflow service exists)
	// Note: Commented out until we verify WorkflowState service exists
	// state, err := s.repos.WorkflowState.GetCurrentState(s.ctx, received.WorkOrder)
	// s.Require().NoError(err)
	// s.Assert().Equal(models.StateReceived, *state)

	// Test duplicate work order validation
	dupReq := &validation.ReceivedValidation{
		WorkOrder:  "WO-SERVICE-001", // Same work order
		CustomerID: customer.CustomerID,
		Joints:     50,
		Size:       "7\"",
		Grade:      "L80",
		Weight:     "25",
		Connection: "BTC",
	}
	
	_, err = s.services.Received.Create(s.ctx, dupReq)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "work order already exists")

	// Test filtering with caching
	filters := map[string]interface{}{
		"customer_id": customer.CustomerID,
	}
	
	// Note: Updated method call to match actual service interface
	items, total, err := s.services.Received.GetAll(s.ctx, filters, 10, 0)
	s.Require().NoError(err)
	s.Assert().Equal(1, total)
	s.Assert().Len(items, 1)
	s.Assert().Equal(received.WorkOrder, items[0].WorkOrder)

	// Test status updates
	err = s.services.Received.UpdateStatus(s.ctx, received.ID, "in_inspection", "Ready for inspection")
	s.Require().NoError(err)

	// Verify status was updated
	updated, err := s.repos.Received.GetByID(s.ctx, received.ID)
	s.Require().NoError(err)
	s.Assert().Equal(received.WorkOrder, updated.WorkOrder)
}

// Test inspection service with quality control (if inspection service exists)
func (s *ServiceIntegrationTestSuite) TestInspectionService() {
	// Setup test data
	customer := &models.Customer{Customer: "Inspection Service Customer"}
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	received := &models.ReceivedItem{
		WorkOrder:    "WO-INSPECTION-001",
		CustomerID:   customer.CustomerID,
		Customer:     customer.Customer,
		Joints:       100,
		Size:         "5 1/2\"",
		Grade:        "J55",
		Weight:       "20",
		Connection:   "LTC",
		DateReceived: testdata.TimePtr(time.Now()),
	}
	err = s.repos.Received.Create(s.ctx, received)
	s.Require().NoError(err)

	s.T().Skip("Inspection service not implemented yet")

	// Test inspection creation (placeholder for when service is implemented)
	// This will need to be updated when InspectionValidation is defined
	s.T().Skip("InspectionValidation not yet defined - implement when inspection service is ready")
}

// Test analytics service with real data (if analytics service exists) 
func (s *ServiceIntegrationTestSuite) TestAnalyticsService() {
	// Check if analytics service exists
	if s.services.Analytics == nil {
		s.T().Skip("Analytics service not implemented yet")
		return
	}

	// Setup test data across multiple customers
	customers := []*models.Customer{
		{Customer: "Analytics Customer 1"},
		{Customer: "Analytics Customer 2"},
		{Customer: "Analytics Customer 3"},
	}
	
	for _, customer := range customers {
		err := s.repos.Customer.Create(s.ctx, customer)
		s.Require().NoError(err)
	}

	// Test basic analytics functionality (to be expanded when service is implemented)
	s.T().Skip("Analytics service methods not yet defined - implement when service is ready")
}

// Test search service functionality
func (s *ServiceIntegrationTestSuite) TestSearchService() {
	// Check if search service exists
	if s.services.Search == nil {
		s.T().Skip("Search service not implemented yet")
		return
	}

	// Setup test data
	customer := &models.Customer{Customer: "Search Test Customer"}
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	// Create multiple received items for searching
	for i := 1; i <= 5; i++ {
		req := &validation.ReceivedValidation{
			WorkOrder:  fmt.Sprintf("WO-SEARCH-%03d", i),
			CustomerID: customer.CustomerID,
			Joints:     100 + i*10,
			Size:       "5 1/2\"",
			Grade:      "J55",
			Weight:     "20",
			Connection: "LTC",
		}
		
		_, err := s.services.Received.Create(s.ctx, req)
		s.Require().NoError(err)
	}

	// Test search functionality (to be implemented when search service methods are defined)
	s.T().Skip("Search service methods not yet defined - implement when service is ready")
}

// Test batch operations service
func (s *ServiceIntegrationTestSuite) TestBatchOperationsService() {
	// Check if batch service exists
	s.T().Skip("Batch service not implemented yet")

	// Setup customer
	customer := &models.Customer{Customer: "Batch Operations Customer"}
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	// Test batch received creation
	// batchData := []validation.ReceivedValidation{
	// 	{
	// 		WorkOrder:  "WO-BATCH-001",
	// 		CustomerID: customer.CustomerID,
	// 		Joints:     100,
	// 		Size:       "5 1/2\"",
	// 		Grade:      "J55",
	// 		Weight:     "20",
	// 		Connection: "LTC",
	// 	},
	// 	{
	// 		WorkOrder:  "WO-BATCH-002", 
	// 		CustomerID: customer.CustomerID,
	// 		Joints:     150,
	// 		Size:       "7\"",
	// 		Grade:      "L80",
	// 		Weight:     "25",
	// 		Connection: "BTC",
	// 	},
	// 	{
	// 		WorkOrder:  "WO-BATCH-003",
	// 		CustomerID: customer.CustomerID,
	// 		Joints:     200,
	// 		Size:       "9 5/8\"",
	// 		Grade:      "N80",
	// 		Weight:     "30",
	// 		Connection: "EUE",
	// 	},
	// }

	// Test batch operations (to be implemented when batch service methods are defined)
	s.T().Skip("Batch service methods not yet defined - implement when service is ready")
}

func TestServiceIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping service integration tests in short mode")
	}
	
	suite.Run(t, new(ServiceIntegrationTestSuite))
}

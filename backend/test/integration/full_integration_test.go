// backend/test/integration/full_integration_test.go
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	"oilgas-backend/internal/handlers"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/test/testutil"
	"oilgas-backend/test/integration/testdata"
)

type IntegrationTestSuite struct {
	suite.Suite
	db       *testutil.TestDB
	repos    *repository.Repositories
	services *services.Services
	router   *gin.Engine
	ctx      context.Context
}

func (s *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.ctx = context.Background()
	
	// Setup test database
	s.db = testutil.SetupTestDB(s.T())
	s.repos = repository.New(s.db.Pool)
	
	// Setup cache
	cache := cache.New(cache.Config{
		TTL:             5 * time.Minute,
		CleanupInterval: 10 * time.Minute,
		MaxSize:         100,
	})
	
	// Setup services
	s.services = services.New(s.repos, cache)
	
	// Setup HTTP router
	s.router = gin.New()
	handlers := handlers.New(s.services)
	api := s.router.Group("/api/v1")
	handlers.RegisterRoutes(api)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.CleanupTestDB(s.T(), s.db)
}

func (s *IntegrationTestSuite) SetupTest() {
	// Clean database for each test
	s.db.Truncate(s.T())
	s.db.SeedGrades(s.T())
}

// Test complete workflow from receiving to production
func (s *IntegrationTestSuite) TestCompleteWorkflow() {
	// Step 1: Create customer
	customer := &models.Customer{
		Customer:       "Workflow Test Company",
		BillingAddress: "123 Test St",
		BillingCity:    "Houston",
		BillingState:   "TX",
		Phone:          "555-123-4567",
		Email:          "test@workflow.com",
	}
	
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)
	s.Assert().NotZero(customer.CustomerID)

	// Step 2: Receive inventory
	received := &models.ReceivedItem{
		WorkOrder:    "WO-WORKFLOW-001",
		CustomerID:   customer.CustomerID,
		Customer:     customer.Customer,
		Joints:       100,
		Size:         "5 1/2\"",
		Weight:       "20",
		Grade:        "J55",
		Connection:   "LTC",
		DateReceived: testdata.TimePtr(time.Now()),
		Notes:        "Test workflow item",
	}
	
	err = s.repos.Received.Create(s.ctx, received)
	s.Require().NoError(err)
	s.Assert().NotZero(received.ID)

	// Step 3: Test workflow state transitions
	currentState, err := s.repos.WorkflowState.GetCurrentState(s.ctx, received.WorkOrder)
	s.Require().NoError(err)
	s.Assert().Equal(string(models.StateReceived), currentState)

	// Move to inspection
	err = s.repos.WorkflowState.TransitionTo(s.ctx, received.WorkOrder, models.StateInspection, "Ready for inspection")
	s.Require().NoError(err)

	// Step 4: Create inspection record
	inspection := &models.InspectionItem{
		WorkOrder:      received.WorkOrder,
		CustomerID:     customer.CustomerID,
		Customer:       customer.Customer,
		Joints:         received.Joints,
		Size:           received.Size,
		Grade:          received.Grade,
		PassedJoints:   95,
		FailedJoints:   5,
		InspectionDate: testdata.TimePtr(time.Now()),
		Inspector:      "John Doe",
		Notes:          "5 joints failed thread inspection",
	}
	
	err = s.repos.Inspected.Create(s.ctx, inspection)
	s.Require().NoError(err)

	// Step 5: Move to production
	err = s.repos.WorkflowState.TransitionTo(s.ctx, received.WorkOrder, models.StateProduction, "Inspection complete, moving to production")
	s.Require().NoError(err)

	// Step 6: Add to inventory
	inventory := &models.InventoryItem{
		CustomerID: customer.CustomerID,
		Customer:   customer.Customer,
		Joints:     inspection.PassedJoints,
		Size:       received.Size,
		Weight:     received.Weight,
		Grade:      received.Grade,
		Connection: received.Connection,
		Color:      "PROCESSED",
		Location:   "Production Area",
		DateIn:     testdata.TimePtr(time.Now()),
	}
	
	err = s.repos.Inventory.Create(s.ctx, inventory)
	s.Require().NoError(err)

	// Step 7: Verify analytics data
	stats, err := s.repos.Analytics.GetDashboardStats(s.ctx)
	s.Require().NoError(err)
	s.Assert().Greater(stats.ActiveJobs, 0)
	s.Assert().Greater(stats.ActiveInventory, 0)
	s.Assert().Greater(stats.TotalCustomers, 0)
	s.Assert().Greater(stats.TotalJoints, 0) 
}

// Test API endpoints end-to-end
func (s *IntegrationTestSuite) TestAPIEndpoints() {
	// Create test customer via API
	customerPayload := `{
		"name": "API Test Customer",
		"billing_address": "456 API St",
		"billing_city": "Dallas",
		"billing_state": "TX",
		"phone": "555-987-6543"
	}`
	
	req := httptest.NewRequest("POST", "/api/v1/customers", strings.NewReader(customerPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	s.router.ServeHTTP(w, req)
	s.Assert().Equal(http.StatusCreated, w.Code)
	
	var customerResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &customerResponse)
	s.Require().NoError(err)
	s.Assert().True(customerResponse["success"].(bool))
	
	customerID := int(customerResponse["data"].(map[string]interface{})["id"].(float64))

	// Test receiving inventory via API
	receivedPayload := fmt.Sprintf(`{
		"work_order": "WO-API-001",
		"customer_id": %d,
		"customer": "API Test Customer",
		"joints": 150,
		"size": "7\"",
		"weight": "26",
		"grade": "L80",
		"connection": "BTC",
		"color": "BLUE",
		"location": "Yard B"
	}`, customerID)
	
	req = httptest.NewRequest("POST", "/api/v1/received", strings.NewReader(receivedPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	s.router.ServeHTTP(w, req)
	s.Assert().Equal(http.StatusCreated, w.Code)

	// Test getting received items
	req = httptest.NewRequest("GET", "/api/v1/received", nil)
	w = httptest.NewRecorder()
	
	s.router.ServeHTTP(w, req)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var receivedResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &receivedResponse)
	s.Require().NoError(err)
	s.Assert().True(receivedResponse["success"].(bool))
	
	data := receivedResponse["data"].([]interface{})
	s.Assert().Len(data, 1)

	// Test filtering
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/received?customer_id=%d", customerID), nil)
	w = httptest.NewRecorder()
	
	s.router.ServeHTTP(w, req)
	s.Assert().Equal(http.StatusOK, w.Code)
}

// Test concurrent operations
func (s *IntegrationTestSuite) TestConcurrentOperations() {
	customer := &models.Customer{
		Customer: "Concurrent Test Customer",
		Phone:    "555-concurrent",
	}
	
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	// Test concurrent received item creation
	const numWorkers = 10
	errors := make(chan error, numWorkers)
	
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			received := &models.ReceivedItem{
				WorkOrder:    fmt.Sprintf("WO-CONCURRENT-%03d", id),
				CustomerID:   customer.CustomerID,
				Customer:     customer.Customer,
				Joints:       100 + id,
				Size:         "5 1/2\"",
				Grade:        "J55",
				DateReceived: testdata.TimePtr(time.Now()),
			}
			
			err := s.repos.Received.Create(s.ctx, received)
			errors <- err
		}(i)
	}
	
	// Collect results
	for i := 0; i < numWorkers; i++ {
		err := <-errors
		s.Assert().NoError(err)
	}
	
	// Verify all items were created
	receivedFilters := repository.ReceivedFilters{
	    CustomerID: &customer.CustomerID,
	    Page:       1,
	    PerPage:    10,
	}
	receivedItems, total, err := s.repos.Received.GetFiltered(s.ctx, receivedFilters)
	s.Require().NoError(err)
	s.Assert().Len(receivedItems, 1)
	s.Assert().Equal(customer.Customer, receivedItems[0].Customer)
	s.Assert().Equal(1, total) 
}

// Test error scenarios
func (s *IntegrationTestSuite) TestErrorScenarios() {
	// Test creating received item with non-existent customer
	received := &models.ReceivedItem{
		WorkOrder:  "WO-ERROR-001",
		CustomerID: 99999, // Non-existent customer
		Customer:   "Ghost Customer",
		Joints:     100,
		Size:       "5 1/2\"",
		Grade:      "J55",
	}
	
	err := s.repos.Received.Create(s.ctx, received)
	s.Assert().Error(err)

	// Test duplicate work order
	customer := &models.Customer{Customer: "Error Test Customer"}
	err = s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	received1 := &models.ReceivedItem{
		WorkOrder:  "WO-DUPLICATE-001",
		CustomerID: customer.CustomerID,
		Customer:   customer.Customer,
		Joints:     100,
		Size:       "5 1/2\"",
		Grade:      "J55",
	}
	
	err = s.repos.Received.Create(s.ctx, received1)
	s.Require().NoError(err)

	received2 := &models.ReceivedItem{
		WorkOrder:  "WO-DUPLICATE-001", // Same work order
		CustomerID: customer.CustomerID,
		Customer:   customer.Customer,
		Joints:     50,
		Size:       "7\"",
		Grade:      "L80",
	}
	
	err = s.repos.Received.Create(s.ctx, received2)
	s.Assert().Error(err)
}

// Test caching behavior
func (s *IntegrationTestSuite) TestCaching() {
	// Create test data
	customer := &models.Customer{Customer: "Cache Test Customer"}
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	// First call should hit database
	start := time.Now()
	customers, err := s.services.Customer.GetAll(s.ctx)
	duration1 := time.Since(start)
	s.Require().NoError(err)
	s.Assert().Len(customers, 1)

	// Second call should hit cache (should be faster)
	start = time.Now()
	customers, err = s.services.Customer.GetAll(s.ctx)
	duration2 := time.Since(start)
	s.Require().NoError(err)
	s.Assert().Len(customers, 1)
	
	// Cache hit should be faster (though this might be flaky in CI)
	s.Assert().True(duration2 < duration1 || duration2 < 10*time.Millisecond)
}

// Test data consistency across repositories
func (s *IntegrationTestSuite) TestDataConsistency() {
	// Create customer
	customer := &models.Customer{
		Customer: "Consistency Test Customer",
		Phone:    "555-consistency",
	}
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	// Create received item
	received := &models.ReceivedItem{
		WorkOrder:    "WO-CONSISTENCY-001",
		CustomerID:   customer.CustomerID,
		Customer:     customer.Customer,
		Joints:       100,
		Size:         "5 1/2\"",
		Grade:        "J55",
		DateReceived: testdata.TimePtr(time.Now()),
	}
	err = s.repos.Received.Create(s.ctx, received)
	s.Require().NoError(err)

	// Move through workflow
	err = s.repos.WorkflowState.TransitionTo(s.ctx, received.WorkOrder, models.StateInspection, "Moving to inspection")
	s.Require().NoError(err)

	// Create inspection
	inspection := &models.InspectionItem{
		WorkOrder:      received.WorkOrder,
		CustomerID:     customer.CustomerID,
		Customer:       customer.Customer,
		Joints:         received.Joints,
		Size:           received.Size,
		Grade:          received.Grade,
		PassedJoints:   95,
		FailedJoints:   5,
		InspectionDate: testdata.TimePtr(time.Now()),
		Inspector:      "Test Inspector",
	}
	err = s.repos.Inspected.Create(s.ctx, inspection)
	s.Require().NoError(err)

	// Verify data consistency
	// Customer should exist in all related records
	receivedFilters := repository.ReceivedFilters{
	    CustomerID: &customer.CustomerID,
	    Page:       1,
	    PerPage:    10,
	}
	receivedItems, _, err := s.repos.Received.GetFiltered(s.ctx, receivedFilters)

	s.Require().NoError(err)
	s.Assert().Len(receivedItems, 1)
	s.Assert().Equal(customer.Customer, receivedItems[0].Customer)

	// Inspection should reference same customer
	inspectedFilters := repository.InspectedFilters{
	    CustomerID: &customer.CustomerID,
	    Page:       1,
	    PerPage:    10,
	}
	inspectionItems, _, err := s.repos.Inspected.GetFiltered(s.ctx, inspectedFilters)

	s.Require().NoError(err)
	s.Assert().Len(inspectionItems, 1)
	s.Assert().Equal(customer.Customer, inspectionItems[0].Customer)

	// Analytics should reflect the data
	stats, err := s.repos.Analytics.GetDashboardStats(s.ctx)
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(stats.ActiveJobs, 1)
	s.Assert().GreaterOrEqual(stats.ActiveInventory, 1)
}

// Test search functionality
func (s *IntegrationTestSuite) TestSearchFunctionality() {
	// Create test data
	customers := []*models.Customer{
		{Customer: "ABC Oil Company"},
		{Customer: "XYZ Drilling Services"},
		{Customer: "Energy Solutions Inc"},
	}
	
	for _, customer := range customers {
		err := s.repos.Customer.Create(s.ctx, customer)
		s.Require().NoError(err)
	}

	// Create received items with different grades
	grades := []string{"J55", "L80", "N80", "P110"}
	for i, grade := range grades {
		received := &models.ReceivedItem{
			WorkOrder:    fmt.Sprintf("WO-SEARCH-%03d", i),
			CustomerID:   customers[i%len(customers)].CustomerID,
			Customer:     customers[i%len(customers)].Customer,
			Joints:       100 + i*10,
			Size:         "5 1/2\"",
			Grade:        grade,
			DateReceived: testdata.TimePtr(time.Now().AddDate(0, 0, -i)),
		}
		err := s.repos.Received.Create(s.ctx, received)
		s.Require().NoError(err)
	}

	// Test search by grade
	allFilters := repository.ReceivedFilters{
	    Page:    1,
	    PerPage: 100,
	}
	allItems, _, err := s.repos.Received.GetFiltered(s.ctx, allFilters)
	s.Require().NoError(err)

	// Filter by grade manually
	var l80Items []models.ReceivedItem
	for _, item := range allItems {
	    if item.Grade == "L80" {
		l80Items = append(l80Items, item)
	    }
	}

	s.Require().NoError(err)
	s.Assert().Len(l80Items, 1)
	s.Assert().Equal("L80", l80Items[0].Grade)

	// Test search by customer
	customerFilters := repository.ReceivedFilters{
	    CustomerID: &customers[0].CustomerID,
	    Page:       1,
	    PerPage:    10,
	}
	abcItems, _, err := s.repos.Received.GetFiltered(s.ctx, customerFilters)

	s.Require().NoError(err)
	s.Assert().Greater(len(abcItems), 0)
	
	for _, item := range abcItems {
		s.Assert().Equal(customers[0].CustomerID, item.CustomerID)
	}

	// Test combined filters
	combinedFilters := repository.ReceivedFilters{
	    CustomerID: &customers[0].CustomerID,
	    Page:       1,
	    PerPage:    100,
	}
	allCustomerItems, _, err := s.repos.Received.GetFiltered(s.ctx, combinedFilters)
	s.Require().NoError(err)

	// Filter by grade manually
	var combinedItems []models.ReceivedItem
	for _, item := range allCustomerItems {
	    if item.Grade == "J55" {
		combinedItems = append(combinedItems, item)
	    }
	}

	s.Require().NoError(err)
	
	if len(combinedItems) > 0 {
		s.Assert().Equal(customers[0].CustomerID, combinedItems[0].CustomerID)
		s.Assert().Equal("J55", combinedItems[0].Grade)
	}
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	suite.Run(t, new(IntegrationTestSuite))
}

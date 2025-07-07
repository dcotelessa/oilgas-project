// backend/test/integration/service_integration_test.go
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/validation"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/test/testutil"
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
		CustomerName:   "Service Test Customer",
		BillingAddress: "123 Service St",
		BillingCity:    "Houston",
		BillingState:   "TX",
		Phone:          "555-service",
		Email:          "service@test.com",
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
	retrieved, err := s.services.Customer.GetByID(s.ctx, customer.CustomerID)
	s.Require().NoError(err)
	s.Assert().Equal(customer.Customer, retrieved.Customer)

	// Test cache hit (second call should be faster)
	start := time.Now()
	retrieved2, err := s.services.Customer.GetByID(s.ctx, customer.CustomerID)
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
	
	updated, err := s.services.Customer.Update(s.ctx, customer.CustomerID, updateReq)
	s.Require().NoError(err)
	s.Assert().Equal(updateReq.CustomerName, updated.Customer)

	// Verify cache was invalidated
	fromDB, err := s.services.Customer.GetByID(s.ctx, customer.CustomerID)
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
		Color:      "RED",
		Location:   "Test Location",
	}
	
	received, err := s.services.Received.Create(s.ctx, req)
	s.Require().NoError(err)
	s.Assert().NotZero(received.ID)
	s.Assert().Equal(req.WorkOrder, received.WorkOrder)

	// Verify workflow state was initialized
	state, err := s.repos.WorkflowState.GetCurrentState(s.ctx, received.WorkOrder)
	s.Require().NoError(err)
	s.Assert().Equal(models.StateReceived, *state)

	// Test duplicate work order validation
	dupReq := &validation.ReceivedValidation{
		WorkOrder:  "WO-SERVICE-001", // Same work order
		CustomerID: customer.CustomerID,
		Joints:     50,
		Size:       "7\"",
		Grade:      "L80",
	}
	
	_, err = s.services.Received.Create(s.ctx, dupReq)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "work order already exists")

	// Test filtering with caching
	filters := map[string]interface{}{
		"customer_id": customer.CustomerID,
	}
	
	items, total, err := s.services.Received.GetFiltered(s.ctx, filters, 10, 0)
	s.Require().NoError(err)
	s.Assert().Equal(1, total)
	s.Assert().Len(items, 1)
	s.Assert().Equal(received.WorkOrder, items[0].WorkOrder)

	// Test status updates
	err = s.services.Received.UpdateStatus(s.ctx, received.ID, "in_inspection")
	s.Require().NoError(err)

	// Verify status was updated
	updated, err := s.repos.Received.GetByID(s.ctx, received.ID)
	s.Require().NoError(err)
	s.Assert().Equal("in_inspection", updated.Status)
}

// Test inspection service with quality control
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
		DateReceived: timePtr(time.Now()),
	}
	err = s.repos.Received.Create(s.ctx, received)
	s.Require().NoError(err)

	// Move to inspection state
	err = s.repos.WorkflowState.TransitionTo(s.ctx, received.WorkOrder, models.StateInspection, "Ready for inspection")
	s.Require().NoError(err)

	// Test inspection creation
	req := &validation.InspectionValidation{
		WorkOrder:    received.WorkOrder,
		PassedJoints: 95,
		FailedJoints: 5,
		Inspector:    "Test Inspector",
		Notes:        "5 joints failed thread inspection",
	}
	
	inspection, err := s.services.Inspection.Create(s.ctx, req)
	s.Require().NoError(err)
	s.Assert().NotZero(inspection.ID)
	s.Assert().Equal(req.PassedJoints, inspection.PassedJoints)
	s.Assert().Equal(req.FailedJoints, inspection.FailedJoints)

	// Test validation rules
	// Total joints should match received joints
	invalidReq := &validation.InspectionValidation{
		WorkOrder:    received.WorkOrder,
		PassedJoints: 50,
		FailedJoints: 60, // 50 + 60 = 110, but received only 100
		Inspector:    "Test Inspector",
	}
	
	_, err = s.services.Inspection.Create(s.ctx, invalidReq)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "total joints mismatch")

	// Test quality metrics calculation
	metrics, err := s.services.Inspection.GetQualityMetrics(s.ctx, customer.CustomerID, 30)
	s.Require().NoError(err)
	s.Assert().Equal(float64(95), metrics.PassRate) // 95% pass rate
	s.Assert().Equal(1, metrics.TotalInspections)
	s.Assert().Equal(100, metrics.TotalJoints)
}

// Test analytics service with real data
func (s *ServiceIntegrationTestSuite) TestAnalyticsService() {
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

	// Create received items over time
	dates := []time.Time{
		time.Now().AddDate(0, 0, -7),  // 7 days ago
		time.Now().AddDate(0, 0, -5),  // 5 days ago
		time.Now().AddDate(0, 0, -3),  // 3 days ago
		time.Now().AddDate(0, 0, -1),  // 1 day ago
	}
	
			for i, date := range dates {
		received := &models.ReceivedItem{
			WorkOrder:    fmt.Sprintf("WO-ANALYTICS-%03d", i),
			CustomerID:   customers[i%len(customers)].CustomerID,
			Customer:     customers[i%len(customers)].Customer,
			Joints:       100 + i*50,
			Size:         "5 1/2\"",
			Grade:        []string{"J55", "L80", "N80", "P110"}[i%4],
			DateReceived: &date,
		}
		err := s.repos.Received.Create(s.ctx, received)
		s.Require().NoError(err)

		// Create some inspections
		if i%2 == 0 { // Every other item gets inspected
			inspection := &models.InspectionItem{
				WorkOrder:      received.WorkOrder,
				CustomerID:     received.CustomerID,
				Customer:       received.Customer,
				Joints:         received.Joints,
				Size:           received.Size,
				Grade:          received.Grade,
				PassedJoints:   received.Joints - 5, // 5 failed joints
				FailedJoints:   5,
				InspectionDate: timePtr(date.Add(24 * time.Hour)),
				Inspector:      "Analytics Inspector",
			}
			err = s.repos.Inspected.Create(s.ctx, inspection)
			s.Require().NoError(err)

			// Add to inventory
			inventory := &models.InventoryItem{
				CustomerID: received.CustomerID,
				Customer:   received.Customer,
				Joints:     inspection.PassedJoints,
				Size:       received.Size,
				Grade:      received.Grade,
				Location:   "Main Yard",
				DateIn:     timePtr(date.Add(48 * time.Hour)),
			}
			err = s.repos.Inventory.Create(s.ctx, inventory)
			s.Require().NoError(err)
		}
	}

	// Test dashboard stats
	stats, err := s.services.Analytics.GetDashboardStats(s.ctx)
	s.Require().NoError(err)
	s.Assert().Equal(4, stats.TotalReceived)
	s.Assert().Equal(2, stats.TotalInspected)
	s.Assert().Equal(2, stats.TotalInventory)
	s.Assert().Greater(stats.TotalJoints, 0)

	// Test customer activity
	activity, err := s.services.Analytics.GetCustomerActivity(s.ctx, 7, &customers[0].CustomerID)
	s.Require().NoError(err)
	s.Assert().NotNil(activity)
	s.Assert().Greater(len(activity.DailyReceived), 0)

	// Test grade distribution
	gradeStats, err := s.services.Analytics.GetGradeDistribution(s.ctx, 30)
	s.Require().NoError(err)
	s.Assert().Greater(len(gradeStats), 0)
	
	// Verify all grades are represented
	gradeMap := make(map[string]int)
	for _, stat := range gradeStats {
		gradeMap[stat.Grade] = stat.Count
	}
	s.Assert().Contains(gradeMap, "J55")
	s.Assert().Contains(gradeMap, "L80")
}

// Test workflow service with state transitions
func (s *ServiceIntegrationTestSuite) TestWorkflowService() {
	// Setup test data
	customer := &models.Customer{Customer: "Workflow Service Customer"}
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	received := &models.ReceivedItem{
		WorkOrder:    "WO-WORKFLOW-SERVICE-001",
		CustomerID:   customer.CustomerID,
		Customer:     customer.Customer,
		Joints:       100,
		Size:         "5 1/2\"",
		Grade:        "J55",
		DateReceived: timePtr(time.Now()),
	}
	err = s.repos.Received.Create(s.ctx, received)
	s.Require().NoError(err)

	// Test getting current state
	state, err := s.services.Workflow.GetCurrentState(s.ctx, received.WorkOrder)
	s.Require().NoError(err)
	s.Assert().Equal(models.StateReceived, state.State)

	// Test valid state transition
	err = s.services.Workflow.TransitionTo(s.ctx, received.WorkOrder, models.StateInspection, "Moving to inspection", "Test User")
	s.Require().NoError(err)

	// Verify state changed
	newState, err := s.services.Workflow.GetCurrentState(s.ctx, received.WorkOrder)
	s.Require().NoError(err)
	s.Assert().Equal(models.StateInspection, newState.State)

	// Test invalid state transition
	err = s.services.Workflow.TransitionTo(s.ctx, received.WorkOrder, models.StateShipped, "Invalid transition", "Test User")
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "invalid state transition")

	// Test getting state history
	history, err := s.services.Workflow.GetStateHistory(s.ctx, received.WorkOrder)
	s.Require().NoError(err)
	s.Assert().Len(history, 2) // Initial state + transition

	// Test getting items by state
	items, err := s.services.Workflow.GetItemsByState(s.ctx, models.StateInspection)
	s.Require().NoError(err)
	s.Assert().Contains(items, received.WorkOrder)

	// Test workflow completion
	err = s.services.Workflow.TransitionTo(s.ctx, received.WorkOrder, models.StateProduction, "Inspection complete", "Test User")
	s.Require().NoError(err)

	err = s.services.Workflow.TransitionTo(s.ctx, received.WorkOrder, models.StateShipped, "Ready to ship", "Test User")
	s.Require().NoError(err)

	// Verify final state
	finalState, err := s.services.Workflow.GetCurrentState(s.ctx, received.WorkOrder)
	s.Require().NoError(err)
	s.Assert().Equal(models.StateShipped, finalState.State)

	// Test workflow metrics
	metrics, err := s.services.Workflow.GetWorkflowMetrics(s.ctx, 30)
	s.Require().NoError(err)
	s.Assert().Greater(metrics.AverageProcessingTime, time.Duration(0))
	s.Assert().Equal(1, metrics.CompletedWorkflows)
}

// Test search service with complex filters
func (s *ServiceIntegrationTestSuite) TestSearchService() {
	// Setup diverse test data
	customers := []*models.Customer{
		{Customer: "Alpha Oil Company"},
		{Customer: "Beta Drilling Services"},
		{Customer: "Gamma Energy Solutions"},
	}
	
	for _, customer := range customers {
		err := s.repos.Customer.Create(s.ctx, customer)
		s.Require().NoError(err)
	}

	// Create received items with various attributes
	testData := []struct {
		WorkOrder string
		Customer  int
		Size      string
		Grade     string
		Joints    int
		Location  string
		DaysAgo   int
	}{
		{"WO-SEARCH-001", 0, "5 1/2\"", "J55", 100, "Yard A", 1},
		{"WO-SEARCH-002", 1, "7\"", "L80", 150, "Yard B", 2},
		{"WO-SEARCH-003", 2, "5 1/2\"", "N80", 200, "Yard A", 3},
		{"WO-SEARCH-004", 0, "9 5/8\"", "P110", 75, "Yard C", 4},
		{"WO-SEARCH-005", 1, "7\"", "J55", 125, "Yard B", 5},
	}

	for _, data := range testData {
		received := &models.ReceivedItem{
			WorkOrder:    data.WorkOrder,
			CustomerID:   customers[data.Customer].CustomerID,
			Customer:     customers[data.Customer].Customer,
			Joints:       data.Joints,
			Size:         data.Size,
			Grade:        data.Grade,
			Location:     data.Location,
			DateReceived: timePtr(time.Now().AddDate(0, 0, -data.DaysAgo)),
		}
		err := s.repos.Received.Create(s.ctx, received)
		s.Require().NoError(err)
	}

	// Test search by single criteria
	results, err := s.services.Search.SearchReceived(s.ctx, map[string]interface{}{
		"size": "5 1/2\"",
	}, 10, 0)
	s.Require().NoError(err)
	s.Assert().Len(results.Items, 2) // WO-SEARCH-001 and WO-SEARCH-003

	// Test search by multiple criteria
	results, err = s.services.Search.SearchReceived(s.ctx, map[string]interface{}{
		"customer_id": customers[1].CustomerID,
		"grade":       "L80",
	}, 10, 0)
	s.Require().NoError(err)
	s.Assert().Len(results.Items, 1) // Only WO-SEARCH-002

	// Test search with date range
	results, err = s.services.Search.SearchReceived(s.ctx, map[string]interface{}{
		"date_from": time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
		"date_to":   time.Now().Format("2006-01-02"),
	}, 10, 0)
	s.Require().NoError(err)
	s.Assert().Len(results.Items, 3) // Last 3 days

	// Test search with text query
	results, err = s.services.Search.SearchReceived(s.ctx, map[string]interface{}{
		"q": "Alpha",
	}, 10, 0)
	s.Require().NoError(err)
	s.Assert().Len(results.Items, 2) // Both Alpha Oil Company items

	// Test pagination
	results, err = s.services.Search.SearchReceived(s.ctx, map[string]interface{}{}, 2, 0)
	s.Require().NoError(err)
	s.Assert().Len(results.Items, 2)
	s.Assert().Equal(5, results.Total)

	results, err = s.services.Search.SearchReceived(s.ctx, map[string]interface{}{}, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(results.Items, 2)
	s.Assert().Equal(5, results.Total)
}

// Test batch operations service
func (s *ServiceIntegrationTestSuite) TestBatchOperationsService() {
	// Setup customer
	customer := &models.Customer{Customer: "Batch Operations Customer"}
	err := s.repos.Customer.Create(s.ctx, customer)
	s.Require().NoError(err)

	// Test batch received creation
	batchData := []validation.ReceivedValidation{
		{
			WorkOrder:  "WO-BATCH-001",
			CustomerID: customer.CustomerID,
			Joints:     100,
			Size:       "5 1/2\"",
			Grade:      "J55",
		},
		{
			WorkOrder:  "WO-BATCH-002",
			CustomerID: customer.CustomerID,
			Joints:     150,
			Size:       "7\"",
			Grade:      "L80",
		},
		{
			WorkOrder:  "WO-BATCH-003",
			CustomerID: customer.CustomerID,
			Joints:     200,
			Size:       "9 5/8\"",
			Grade:      "N80",
		},
	}

	results, err := s.services.Batch.CreateReceivedItems(s.ctx, batchData)
	s.Require().NoError(err)
	s.Assert().Len(results.Succeeded, 3)
	s.Assert().Len(results.Failed, 0)

	// Test batch with some failures
	batchWithErrors := []validation.ReceivedValidation{
		{
			WorkOrder:  "WO-BATCH-004",
			CustomerID: customer.CustomerID,
			Joints:     75,
			Size:       "5 1/2\"",
			Grade:      "P110",
		},
		{
			WorkOrder:  "WO-BATCH-001", // Duplicate work order
			CustomerID: customer.CustomerID,
			Joints:     100,
			Size:       "7\"",
			Grade:      "J55",
		},
	}

	results, err = s.services.Batch.CreateReceivedItems(s.ctx, batchWithErrors)
	s.Require().NoError(err)
	s.Assert().Len(results.Succeeded, 1) // Only WO-BATCH-004
	s.Assert().Len(results.Failed, 1)    // WO-BATCH-001 duplicate

	// Test batch status updates
	workOrders := []string{"WO-BATCH-001", "WO-BATCH-002", "WO-BATCH-003"}
	updateResults, err := s.services.Batch.UpdateReceivedStatus(s.ctx, workOrders, "in_inspection")
	s.Require().NoError(err)
	s.Assert().Len(updateResults.Succeeded, 3)
	s.Assert().Len(updateResults.Failed, 0)

	// Verify updates
	for _, workOrder := range workOrders {
		item, err := s.repos.Received.GetByWorkOrder(s.ctx, workOrder)
		s.Require().NoError(err)
		s.Assert().Equal("in_inspection", item.Status)
	}
}

func TestServiceIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping service integration tests in short mode")
	}
	
	suite.Run(t, new(ServiceIntegrationTestSuite))
}

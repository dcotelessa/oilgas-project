// backend/test/integration/api_integration_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"oilgas-backend/internal/handlers"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/test/testutil"
)

type APIIntegrationTestSuite struct {
	suite.Suite
	db       *testutil.TestDB
	repos    *repository.Repositories
	services *services.Services
	router   *gin.Engine
}

func (s *APIIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	s.db = testutil.SetupTestDB(s.T())
	s.repos = repository.New(s.db.Pool)
	
	cache := cache.New(cache.Config{
		TTL:             5 * time.Minute,
		CleanupInterval: 10 * time.Minute,
		MaxSize:         100,
	})
	
	s.services = services.New(s.repos, cache)
	
	s.router = gin.New()
	handlers := handlers.New(s.services)
	api := s.router.Group("/api/v1")
	handlers.RegisterRoutes(api)
}

func (s *APIIntegrationTestSuite) TearDownSuite() {
	testutil.CleanupTestDB(s.T(), s.db)
}

func (s *APIIntegrationTestSuite) SetupTest() {
	s.db.Truncate(s.T())
	s.db.SeedGrades(s.T())
}

// Helper function to make JSON requests
func (s *APIIntegrationTestSuite) makeRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	
	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	s.router.ServeHTTP(w, req)
	return w
}

// Helper to parse JSON response
func (s *APIIntegrationTestSuite) parseResponse(w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	s.Require().NoError(err)
}

// Test customer API endpoints
func (s *APIIntegrationTestSuite) TestCustomerAPI() {
	// Test creating customer
	customerData := map[string]interface{}{
		"name":            "API Test Customer",
		"billing_address": "123 API St",
		"billing_city":    "Houston",
		"billing_state":   "TX",
		"phone":           "555-api-test",
		"email":           "api@test.com",
	}
	
	w := s.makeRequest("POST", "/api/v1/customers", customerData)
	s.Assert().Equal(http.StatusCreated, w.Code)
	
	var createResponse map[string]interface{}
	s.parseResponse(w, &createResponse)
	s.Assert().True(createResponse["success"].(bool))
	
	customerID := int(createResponse["data"].(map[string]interface{})["id"].(float64))
	s.Assert().NotZero(customerID)

	// Test getting customer by ID
	w = s.makeRequest("GET", fmt.Sprintf("/api/v1/customers/%d", customerID), nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var getResponse map[string]interface{}
	s.parseResponse(w, &getResponse)
	s.Assert().True(getResponse["success"].(bool))
	
	customer := getResponse["data"].(map[string]interface{})
	s.Assert().Equal("API Test Customer", customer["name"])

	// Test updating customer
	updateData := map[string]interface{}{
		"name":  "Updated API Customer",
		"phone": "555-updated",
		"email": "updated@test.com",
	}
	
	w = s.makeRequest("PUT", fmt.Sprintf("/api/v1/customers/%d", customerID), updateData)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var updateResponse map[string]interface{}
	s.parseResponse(w, &updateResponse)
	s.Assert().True(updateResponse["success"].(bool))

	// Test getting all customers
	w = s.makeRequest("GET", "/api/v1/customers", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var listResponse map[string]interface{}
	s.parseResponse(w, &listResponse)
	s.Assert().True(listResponse["success"].(bool))
	
	customers := listResponse["data"].([]interface{})
	s.Assert().Len(customers, 1)

	// Test validation errors
	invalidData := map[string]interface{}{
		"name": "", // Empty name should fail
	}
	
	w = s.makeRequest("POST", "/api/v1/customers", invalidData)
	s.Assert().Equal(http.StatusBadRequest, w.Code)
	
	var errorResponse map[string]interface{}
	s.parseResponse(w, &errorResponse)
	s.Assert().False(errorResponse["success"].(bool))
	s.Assert().Contains(errorResponse["error"], "validation")

	// Test duplicate customer name
	w = s.makeRequest("POST", "/api/v1/customers", customerData)
	s.Assert().Equal(http.StatusConflict, w.Code)
}

// Test received items API
func (s *APIIntegrationTestSuite) TestReceivedAPI() {
	// Setup customer first
	customer := &models.Customer{Customer: "Received API Customer"}
	err := s.repos.Customer.Create(context.Background(), customer)
	s.Require().NoError(err)

	// Test creating received item
	receivedData := map[string]interface{}{
		"work_order":  "WO-API-001",
		"customer_id": customer.CustomerID,
		"joints":      100,
		"size":        "5 1/2\"",
		"weight":      "20",
		"grade":       "J55",
		"connection":  "LTC",
		"color":       "RED",
		"location":    "API Test Yard",
	}
	
	w := s.makeRequest("POST", "/api/v1/received", receivedData)
	s.Assert().Equal(http.StatusCreated, w.Code)
	
	var createResponse map[string]interface{}
	s.parseResponse(w, &createResponse)
	s.Assert().True(createResponse["success"].(bool))
	
	receivedID := int(createResponse["data"].(map[string]interface{})["id"].(float64))

	// Test getting received items
	w = s.makeRequest("GET", "/api/v1/received", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var listResponse map[string]interface{}
	s.parseResponse(w, &listResponse)
	s.Assert().True(listResponse["success"].(bool))
	
	items := listResponse["data"].([]interface{})
	s.Assert().Len(items, 1)

	// Test filtering by customer
	w = s.makeRequest("GET", fmt.Sprintf("/api/v1/received?customer_id=%d", customer.CustomerID), nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &listResponse)
	items = listResponse["data"].([]interface{})
	s.Assert().Len(items, 1)

	// Test filtering by grade
	w = s.makeRequest("GET", "/api/v1/received?grade=J55", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &listResponse)
	items = listResponse["data"].([]interface{})
	s.Assert().Len(items, 1)

	// Test pagination
	w = s.makeRequest("GET", "/api/v1/received?limit=5&offset=0", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &listResponse)
	s.Assert().Contains(listResponse, "pagination")

	// Test getting by work order
	w = s.makeRequest("GET", "/api/v1/received/work-order/WO-API-001", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var workOrderResponse map[string]interface{}
	s.parseResponse(w, &workOrderResponse)
	s.Assert().True(workOrderResponse["success"].(bool))

	// Test updating status
	statusData := map[string]interface{}{
		"status": "in_inspection",
	}
	
	w = s.makeRequest("PATCH", fmt.Sprintf("/api/v1/received/%d/status", receivedID), statusData)
	s.Assert().Equal(http.StatusOK, w.Code)

	// Test getting pending inspection
	w = s.makeRequest("GET", "/api/v1/received/pending-inspection", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &listResponse)
	// Should be empty since we didn't move to inspection workflow state
	items = listResponse["data"].([]interface{})
	s.Assert().Len(items, 0)

	// Test duplicate work order
	w = s.makeRequest("POST", "/api/v1/received", receivedData)
	s.Assert().Equal(http.StatusConflict, w.Code)
}

// Test workflow API
func (s *APIIntegrationTestSuite) TestWorkflowAPI() {
	// Setup test data
	customer := &models.Customer{Customer: "Workflow API Customer"}
	err := s.repos.Customer.Create(context.Background(), customer)
	s.Require().NoError(err)

	received := &models.ReceivedItem{
		WorkOrder:    "WO-WORKFLOW-API-001",
		CustomerID:   customer.CustomerID,
		Customer:     customer.Customer,
		Joints:       100,
		Size:         "5 1/2\"",
		Grade:        "J55",
		DateReceived: timePtr(time.Now()),
	}
	err = s.repos.Received.Create(context.Background(), received)
	s.Require().NoError(err)

	// Test getting current state
	w := s.makeRequest("GET", fmt.Sprintf("/api/v1/workflow/state/%s", received.WorkOrder), nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var stateResponse map[string]interface{}
	s.parseResponse(w, &stateResponse)
	s.Assert().True(stateResponse["success"].(bool))
	s.Assert().Equal("received", stateResponse["data"].(map[string]interface{})["state"])

	// Test state transition
	transitionData := map[string]interface{}{
		"new_state": "inspection",
		"notes":     "Moving to inspection",
		"user":      "API Test User",
	}
	
	w = s.makeRequest("POST", fmt.Sprintf("/api/v1/workflow/transition/%s", received.WorkOrder), transitionData)
	s.Assert().Equal(http.StatusOK, w.Code)

	// Verify state changed
	w = s.makeRequest("GET", fmt.Sprintf("/api/v1/workflow/state/%s", received.WorkOrder), nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &stateResponse)
	s.Assert().Equal("inspection", stateResponse["data"].(map[string]interface{})["state"])

	// Test getting state history
	w = s.makeRequest("GET", fmt.Sprintf("/api/v1/workflow/history/%s", received.WorkOrder), nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var historyResponse map[string]interface{}
	s.parseResponse(w, &historyResponse)
	s.Assert().True(historyResponse["success"].(bool))
	
	history := historyResponse["data"].([]interface{})
	s.Assert().Len(history, 2) // Initial state + transition

	// Test getting items by state
	w = s.makeRequest("GET", "/api/v1/workflow/items/inspection", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var itemsResponse map[string]interface{}
	s.parseResponse(w, &itemsResponse)
	s.Assert().True(itemsResponse["success"].(bool))
	
	items := itemsResponse["data"].([]interface{})
	s.Assert().Len(items, 1)

	// Test invalid state transition
	invalidTransition := map[string]interface{}{
		"new_state": "shipped", // Can't go from inspection to shipped
		"notes":     "Invalid transition",
		"user":      "API Test User",
	}
	
	w = s.makeRequest("POST", fmt.Sprintf("/api/v1/workflow/transition/%s", received.WorkOrder), invalidTransition)
	s.Assert().Equal(http.StatusBadRequest, w.Code)
}

// Test analytics API
func (s *APIIntegrationTestSuite) TestAnalyticsAPI() {
	// Setup test data
	customer := &models.Customer{Customer: "Analytics API Customer"}
	err := s.repos.Customer.Create(context.Background(), customer)
	s.Require().NoError(err)

	// Create some test data
	for i := 0; i < 5; i++ {
		received := &models.ReceivedItem{
			WorkOrder:    fmt.Sprintf("WO-ANALYTICS-API-%03d", i),
			CustomerID:   customer.CustomerID,
			Customer:     customer.Customer,
			Joints:       100 + i*10,
			Size:         "5 1/2\"",
			Grade:        []string{"J55", "L80", "N80"}[i%3],
			DateReceived: timePtr(time.Now().AddDate(0, 0, -i)),
		}
		err = s.repos.Received.Create(context.Background(), received)
		s.Require().NoError(err)

		// Create some inspections
		if i%2 == 0 {
			inspection := &models.InspectionItem{
				WorkOrder:      received.WorkOrder,
				CustomerID:     received.CustomerID,
				Customer:       received.Customer,
				Joints:         received.Joints,
				Size:           received.Size,
				Grade:          received.Grade,
				PassedJoints:   received.Joints - 5,
				FailedJoints:   5,
				InspectionDate: timePtr(time.Now().AddDate(0, 0, -i+1)),
				Inspector:      "API Inspector",
			}
			err = s.repos.Inspected.Create(context.Background(), inspection)
			s.Require().NoError(err)
		}
	}

	// Test dashboard stats
	w := s.makeRequest("GET", "/api/v1/analytics/dashboard", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var dashboardResponse map[string]interface{}
	s.parseResponse(w, &dashboardResponse)
	s.Assert().True(dashboardResponse["success"].(bool))
	
	stats := dashboardResponse["data"].(map[string]interface{})
	s.Assert().Equal(float64(5), stats["total_received"])
	s.Assert().Equal(float64(3), stats["total_inspected"])
	s.Assert().Greater(stats["total_joints"], float64(0))

	// Test customer activity
	w = s.makeRequest("GET", fmt.Sprintf("/api/v1/analytics/customer-activity?customer_id=%d&days=7", customer.CustomerID), nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var activityResponse map[string]interface{}
	s.parseResponse(w, &activityResponse)
	s.Assert().True(activityResponse["success"].(bool))

	// Test grade distribution
	w = s.makeRequest("GET", "/api/v1/analytics/grade-distribution?days=30", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var gradeResponse map[string]interface{}
	s.parseResponse(w, &gradeResponse)
	s.Assert().True(gradeResponse["success"].(bool))
	
	grades := gradeResponse["data"].([]interface{})
	s.Assert().Greater(len(grades), 0)

	// Test workflow metrics
	w = s.makeRequest("GET", "/api/v1/analytics/workflow-metrics?days=30", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var workflowResponse map[string]interface{}
	s.parseResponse(w, &workflowResponse)
	s.Assert().True(workflowResponse["success"].(bool))
}

// Test search API
func (s *APIIntegrationTestSuite) TestSearchAPI() {
	// Setup test data
	customer := &models.Customer{Customer: "Search API Customer"}
	err := s.repos.Customer.Create(context.Background(), customer)
	s.Require().NoError(err)

	// Create diverse test data
	testItems := []map[string]interface{}{
		{
			"work_order":  "WO-SEARCH-001",
			"customer_id": customer.CustomerID,
			"joints":      100,
			"size":        "5 1/2\"",
			"grade":       "J55",
			"location":    "Yard A",
		},
		{
			"work_order":  "WO-SEARCH-002",
			"customer_id": customer.CustomerID,
			"joints":      150,
			"size":        "7\"",
			"grade":       "L80",
			"location":    "Yard B",
		},
		{
			"work_order":  "WO-SEARCH-003",
			"customer_id": customer.CustomerID,
			"joints":      200,
			"size":        "5 1/2\"",
			"grade":       "N80",
			"location":    "Yard A",
		},
	}

	for _, item := range testItems {
		w := s.makeRequest("POST", "/api/v1/received", item)
		s.Assert().Equal(http.StatusCreated, w.Code)
	}

	// Test basic search
	w := s.makeRequest("GET", "/api/v1/search?q=WO-SEARCH", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var searchResponse map[string]interface{}
	s.parseResponse(w, &searchResponse)
	s.Assert().True(searchResponse["success"].(bool))
	
	results := searchResponse["data"].(map[string]interface{})
	items := results["items"].([]interface{})
	s.Assert().Len(items, 3)

	// Test search by size
	w = s.makeRequest("GET", "/api/v1/search?size=5+1/2\"", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &searchResponse)
	results = searchResponse["data"].(map[string]interface{})
	items = results["items"].([]interface{})
	s.Assert().Len(items, 2) // WO-SEARCH-001 and WO-SEARCH-003

	// Test search by grade
	w = s.makeRequest("GET", "/api/v1/search?grade=L80", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &searchResponse)
	results = searchResponse["data"].(map[string]interface{})
	items = results["items"].([]interface{})
	s.Assert().Len(items, 1)

	// Test search by location
	w = s.makeRequest("GET", "/api/v1/search?location=Yard+A", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &searchResponse)
	results = searchResponse["data"].(map[string]interface{})
	items = results["items"].([]interface{})
	s.Assert().Len(items, 2)

	// Test combined search
	w = s.makeRequest("GET", "/api/v1/search?size=5+1/2\"&location=Yard+A", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &searchResponse)
	results = searchResponse["data"].(map[string]interface{})
	items = results["items"].([]interface{})
	s.Assert().Len(items, 2)

	// Test pagination
	w = s.makeRequest("GET", "/api/v1/search?limit=2&offset=0", nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &searchResponse)
	results = searchResponse["data"].(map[string]interface{})
	items = results["items"].([]interface{})
	s.Assert().Len(items, 2)
	s.Assert().Equal(float64(3), results["total"])
}

// Test batch operations API
func (s *APIIntegrationTestSuite) TestBatchAPI() {
	// Setup customer
	customer := &models.Customer{Customer: "Batch API Customer"}
	err := s.repos.Customer.Create(context.Background(), customer)
	s.Require().NoError(err)

	// Test batch create received items
	batchData := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"work_order":  "WO-BATCH-001",
				"customer_id": customer.CustomerID,
				"joints":      100,
				"size":        "5 1/2\"",
				"grade":       "J55",
			},
			{
				"work_order":  "WO-BATCH-002",
				"customer_id": customer.CustomerID,
				"joints":      150,
				"size":        "7\"",
				"grade":       "L80",
			},
			{
				"work_order":  "WO-BATCH-003",
				"customer_id": customer.CustomerID,
				"joints":      200,
				"size":        "9 5/8\"",
				"grade":       "N80",
			},
		},
	}

	w := s.makeRequest("POST", "/api/v1/batch/received", batchData)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var batchResponse map[string]interface{}
	s.parseResponse(w, &batchResponse)
	s.Assert().True(batchResponse["success"].(bool))
	
	results := batchResponse["data"].(map[string]interface{})
	succeeded := results["succeeded"].([]interface{})
	failed := results["failed"].([]interface{})
	s.Assert().Len(succeeded, 3)
	s.Assert().Len(failed, 0)

	// Test batch with some failures
	batchWithErrors := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"work_order":  "WO-BATCH-004",
				"customer_id": customer.CustomerID,
				"joints":      75,
				"size":        "5 1/2\"",
				"grade":       "P110",
			},
			{
				"work_order":  "WO-BATCH-001", // Duplicate
				"customer_id": customer.CustomerID,
				"joints":      100,
				"size":        "7\"",
				"grade":       "J55",
			},
			{
				"work_order":  "WO-BATCH-005",
				"customer_id": 99999, // Invalid customer
				"joints":      50,
				"size":        "7\"",
				"grade":       "L80",
			},
		},
	}

	w = s.makeRequest("POST", "/api/v1/batch/received", batchWithErrors)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &batchResponse)
	results = batchResponse["data"].(map[string]interface{})
	succeeded = results["succeeded"].([]interface{})
	failed = results["failed"].([]interface{})
	s.Assert().Len(succeeded, 1) // Only WO-BATCH-004
	s.Assert().Len(failed, 2)    // Duplicate and invalid customer

	// Test batch status update
	statusUpdateData := map[string]interface{}{
		"work_orders": []string{"WO-BATCH-001", "WO-BATCH-002", "WO-BATCH-003"},
		"status":      "in_inspection",
	}

	w = s.makeRequest("PATCH", "/api/v1/batch/received/status", statusUpdateData)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	s.parseResponse(w, &batchResponse)
	results = batchResponse["data"].(map[string]interface{})
	succeeded = results["succeeded"].([]interface{})
	failed = results["failed"].([]interface{})
	s.Assert().Len(succeeded, 3)
	s.Assert().Len(failed, 0)
}

// Test error handling and edge cases
func (s *APIIntegrationTestSuite) TestErrorHandling() {
	// Test 404 errors
	w := s.makeRequest("GET", "/api/v1/customers/99999", nil)
	s.Assert().Equal(http.StatusNotFound, w.Code)
	
	var errorResponse map[string]interface{}
	s.parseResponse(w, &errorResponse)
	s.Assert().False(errorResponse["success"].(bool))

	// Test invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/customers", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Assert().Equal(http.StatusBadRequest, w.Code)

	// Test missing required fields
	invalidCustomer := map[string]interface{}{
		"billing_address": "123 Test St",
		// Missing required name field
	}
	
	w = s.makeRequest("POST", "/api/v1/customers", invalidCustomer)
	s.Assert().Equal(http.StatusBadRequest, w.Code)

	// Test invalid work order format
	customer := &models.Customer{Customer: "Error Test Customer"}
	err := s.repos.Customer.Create(context.Background(), customer)
	s.Require().NoError(err)

	invalidReceived := map[string]interface{}{
		"work_order":  "", // Empty work order
		"customer_id": customer.CustomerID,
		"joints":      100,
		"size":        "5 1/2\"",
		"grade":       "J55",
	}
	
	w = s.makeRequest("POST", "/api/v1/received", invalidReceived)
	s.Assert().Equal(http.StatusBadRequest, w.Code)

	// Test non-existent work order
	w = s.makeRequest("GET", "/api/v1/workflow/state/NON-EXISTENT", nil)
	s.Assert().Equal(http.StatusNotFound, w.Code)
}

// Test rate limiting and performance
func (s *APIIntegrationTestSuite) TestPerformance() {
	customer := &models.Customer{Customer: "Performance Test Customer"}
	err := s.repos.Customer.Create(context.Background(), customer)
	s.Require().NoError(err)

	// Test concurrent requests
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			receivedData := map[string]interface{}{
				"work_order":  fmt.Sprintf("WO-PERF-%03d", id),
				"customer_id": customer.CustomerID,
				"joints":      100 + id,
				"size":        "5 1/2\"",
				"grade":       "J55",
			}
			
			w := s.makeRequest("POST", "/api/v1/received", receivedData)
			results <- w.Code
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		code := <-results
		if code == http.StatusCreated {
			successCount++
		}
	}

	s.Assert().Equal(numRequests, successCount)

	// Verify all items were created
	w := s.makeRequest("GET", fmt.Sprintf("/api/v1/received?customer_id=%d", customer.CustomerID), nil)
	s.Assert().Equal(http.StatusOK, w.Code)
	
	var listResponse map[string]interface{}
	s.parseResponse(w, &listResponse)
	items := listResponse["data"].([]interface{})
	s.Assert().Len(items, numRequests)
}

// Test API versioning and backwards compatibility
func (s *APIIntegrationTestSuite) TestAPIVersioning() {
	// Test that v1 endpoints are accessible
	w := s.makeRequest("GET", "/api/v1/health", nil)
	s.Assert().Equal(http.StatusOK, w.Code)

	// Test CORS headers
	req := httptest.NewRequest("OPTIONS", "/api/v1/customers", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w = httptest.NewRecorder()
	
	s.router.ServeHTTP(w, req)
	s.Assert().Contains(w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

// Test comprehensive workflow through API
func (s *APIIntegrationTestSuite) TestCompleteWorkflowAPI() {
	// Step 1: Create customer
	customerData := map[string]interface{}{
		"name":  "Complete Workflow Customer",
		"phone": "555-workflow",
	}
	
	w := s.makeRequest("POST", "/api/v1/customers", customerData)
	s.Require().Equal(http.StatusCreated, w.Code)
	
	var customerResponse map[string]interface{}
	s.parseResponse(w, &customerResponse)
	customerID := int(customerResponse["data"].(map[string]interface{})["id"].(float64))

	// Step 2: Receive inventory
	receivedData := map[string]interface{}{
		"work_order":  "WO-COMPLETE-001",
		"customer_id": customerID,
		"joints":      100,
		"size":        "5 1/2\"",
		"grade":       "J55",
		"location":    "Main Yard",
	}
	
	w = s.makeRequest("POST", "/api/v1/received", receivedData)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Step 3: Move to inspection
	transitionData := map[string]interface{}{
		"new_state": "inspection",
		"notes":     "Ready for inspection",
		"user":      "API Workflow Test",
	}
	
	w = s.makeRequest("POST", "/api/v1/workflow/transition/WO-COMPLETE-001", transitionData)
	s.Require().Equal(http.StatusOK, w.Code)

	// Step 4: Create inspection
	inspectionData := map[string]interface{}{
		"work_order":    "WO-COMPLETE-001",
		"passed_joints": 95,
		"failed_joints": 5,
		"inspector":     "API Inspector",
		"notes":         "Minor thread issues on 5 joints",
	}
	
	w = s.makeRequest("POST", "/api/v1/inspections", inspectionData)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Step 5: Move to production
	transitionData["new_state"] = "production"
	transitionData["notes"] = "Inspection complete, moving to production"
	
	w = s.makeRequest("POST", "/api/v1/workflow/transition/WO-COMPLETE-001", transitionData)
	s.Require().Equal(http.StatusOK, w.Code)

	// Step 6: Verify analytics reflect the workflow
	w = s.makeRequest("GET", "/api/v1/analytics/dashboard", nil)
	s.Require().Equal(http.StatusOK, w.Code)
	
	var dashboardResponse map[string]interface{}
	s.parseResponse(w, &dashboardResponse)
	stats := dashboardResponse["data"].(map[string]interface{})
	s.Assert().Greater(stats["total_received"], float64(0))
	s.Assert().Greater(stats["total_inspected"], float64(0))

	// Step 7: Verify workflow history
	w = s.makeRequest("GET", "/api/v1/workflow/history/WO-COMPLETE-001", nil)
	s.Require().Equal(http.StatusOK, w.Code)
	
	var historyResponse map[string]interface{}
	s.parseResponse(w, &historyResponse)
	history := historyResponse["data"].([]interface{})
	s.Assert().Len(history, 3) // received -> inspection -> production
}

func TestAPIIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API integration tests in short mode")
	}
	
	suite.Run(t, new(APIIntegrationTestSuite))
}

func timePtr(t time.Time) *time.Time {
	return &t
}

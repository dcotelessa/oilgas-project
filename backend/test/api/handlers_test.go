// backend/test/api/handlers_test.go

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oilgas-backend/internal/handlers"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/test/testutil"
)

func setupTestAPI(t *testing.T) (*gin.Engine, *repository.Repositories) {
	gin.SetMode(gin.TestMode)

	db := testutil.SetupTestDB(t)
	repos := repository.New(db)
	cache := cache.New(cache.Config{
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		MaxSize:         100,
	})
	services := services.New(repos, cache)
	handlers := handlers.New(services)

	r := gin.New()
	api := r.Group("/api/v1")
	handlers.RegisterRoutes(api)

	return r, repos
}

func TestReceivedAPI(t *testing.T) {
	r, repos := setupTestAPI(t)
	defer testutil.CleanupTestDB(t, repos.Customer.(*customerRepository).db)

	ctx := context.Background()

	// Setup test customer
	customer := &models.Customer{Name: "API Test Customer"}
	err := repos.Customer.Create(ctx, customer)
	require.NoError(t, err)

	t.Run("POST /api/v1/received", func(t *testing.T) {
		received := models.ReceivedItem{
			WorkOrder:  "WO-API-001",
			CustomerID: customer.ID,
			Customer:   customer.Name,
			Joints:     100,
			Size:       "5 1/2\"",
			Grade:      "J55",
		}

		body, _ := json.Marshal(received)
		req := httptest.NewRequest("POST", "/api/v1/received", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	t.Run("GET /api/v1/received", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/received", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "data")
	})

	t.Run("GET /api/v1/received/pending-inspection", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/received/pending-inspection", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestWorkflowAPI(t *testing.T) {
	r, repos := setupTestAPI(t)
	defer testutil.CleanupTestDB(t, repos.Customer.(*customerRepository).db)

	ctx := context.Background()

	// Setup test data
	customer := &models.Customer{Name: "Workflow Test Customer"}
	err := repos.Customer.Create(ctx, customer)
	require.NoError(t, err)

	received := &models.ReceivedItem{
		WorkOrder:  "WO-WORKFLOW-API",
		CustomerID: customer.ID,
		Customer:   customer.Name,
		Joints:     50,
	}
	err = repos.Received.Create(ctx, received)
	require.NoError(t, err)

	t.Run("GET /api/v1/workflow/:workOrder/state", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/workflow/WO-WORKFLOW-API/state", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	t.Run("POST /api/v1/workflow/:workOrder/transition", func(t *testing.T) {
		transition := map[string]interface{}{
			"new_state": "PRODUCTION",
			"notes":     "API test transition",
		}

		body, _ := json.Marshal(transition)
		req := httptest.NewRequest("POST", "/api/v1/workflow/WO-WORKFLOW-API/transition", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

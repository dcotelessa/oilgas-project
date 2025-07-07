// backend/test/integration/repository_integration_test.go

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/test/testutil"
	"oilgas-backend/test/integration/testdata"
)

func TestRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repos := repository.New(db.Pool)
	ctx := context.Background()

	t.Run("Customer Repository", func(t *testing.T) {
		// Test customer CRUD operations
		customer := &models.Customer{
			Customer:           "Test Oil Company",
			BillingAddress: "123 Test St",
			BillingCity:    "Houston",
			BillingState:   "TX",
			Phone:          "555-123-4567",
		}

		// Create
		err := repos.Customer.Create(ctx, customer)
		require.NoError(t, err)
		assert.NotZero(t, customer.CustomerID)

		// Read
		retrieved, err := repos.Customer.GetByID(ctx, customer.CustomerID)
		require.NoError(t, err)
		assert.Equal(t, customer.Customer, retrieved.Customer)

		// Update
		customer.Customer = "Updated Oil Company"
		err = repos.Customer.Update(ctx, customer)
		require.NoError(t, err)

		// Verify update
		updated, err := repos.Customer.GetByID(ctx, customer.CustomerID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Oil Company", updated.Customer)

		// Delete
		err = repos.Customer.Delete(ctx, customer.CustomerID)
		require.NoError(t, err)
	})

	t.Run("Received Repository", func(t *testing.T) {
		// First create a customer
		customer := &models.Customer{Customer: "Test Customer"}
		err := repos.Customer.Create(ctx, customer)
		require.NoError(t, err)

		received := &models.ReceivedItem{
			WorkOrder:    "WO-TEST-001",
			CustomerID:   customer.CustomerID,
			Customer:     customer.Customer,
			Joints:       100,
			Size:         "5 1/2\"",
			Weight:       "20",
			Grade:        "J55",
			Connection:   "LTC",
			DateReceived: testdata.TimePtr(time.Now()),
		}

		// Create
		err = repos.Received.Create(ctx, received)
		require.NoError(t, err)
		assert.NotZero(t, received.CustomerID)

		// Read
		retrieved, err := repos.Received.GetByID(ctx, received.CustomerID)
		require.NoError(t, err)
		assert.Equal(t, received.WorkOrder, retrieved.WorkOrder)

		// Get by work order
		byWorkOrder, err := repos.Received.GetByWorkOrder(ctx, received.WorkOrder)
		require.NoError(t, err)
		assert.Equal(t, received.CustomerID, byWorkOrder.CustomerID)

		// Update status
		err = repos.Received.UpdateStatus(ctx, received.ID, "in_production", "Moving to production")
		require.NoError(t, err)

		// Get filtered
		filters := repository.ReceivedFilters{
		    CustomerID: &customer.CustomerID,
		    Page:       1,
		    PerPage:    10,
		}
		items, pagination, err := repos.Received.GetFiltered(ctx, filters)

		require.NoError(t, err)
		assert.Len(t, items, 1)
		if pagination != nil {
		    assert.Equal(t, 1, pagination.Total)
		}

		// Delete
		err = repos.Received.Delete(ctx, received.CustomerID)
		require.NoError(t, err)
	})

	t.Run("Grade Repository", func(t *testing.T) {
		// Get all grades
		grades, err := repos.Grade.GetAll(ctx)
		require.NoError(t, err)
		assert.Contains(t, grades, "J55")
		assert.Contains(t, grades, "L80")

		// Create new grade
		testGrade := &models.Grade{
		    Grade: "TEST123",
		}
		err = repos.Grade.Create(ctx, testGrade)
		require.NoError(t, err)

		// Verify it exists
		updatedGrades, err := repos.Grade.GetAll(ctx)
		require.NoError(t, err)
		assert.Contains(t, updatedGrades, "TEST123")

		// Check if in use (should be false)
		inUse, err := repos.Grade.IsInUse(ctx, "TEST123")
		require.NoError(t, err)
		assert.False(t, inUse)

		// Delete test grade
		err = repos.Grade.Delete(ctx, "TEST123")
		require.NoError(t, err)

		// Verify deletion
		finalGrades, err := repos.Grade.GetAll(ctx)
		require.NoError(t, err)
		assert.NotContains(t, finalGrades, "TEST123")
	})

	t.Run("Analytics Repository", func(t *testing.T) {
		// Test dashboard stats
		stats, err := repos.Analytics.GetDashboardStats(ctx)
		require.NoError(t, err)
		assert.NotNil(t, stats)

		// Test customer activity
		activity, err := repos.Analytics.GetCustomerActivity(ctx, 30, nil)
		require.NoError(t, err)
		assert.NotNil(t, activity)
	})

	t.Run("WorkflowState Repository", func(t *testing.T) {
		// First create a received item to test workflow on
		customer := &models.Customer{Customer: "Workflow Test Customer"}
		err := repos.Customer.Create(ctx, customer)
		require.NoError(t, err)

		received := &models.ReceivedItem{
			WorkOrder:  "WO-WORKFLOW-001",
			CustomerID: customer.CustomerID,
			Customer:   customer.Customer,
			Joints:     50,
			Size:       "7\"",
			Grade:      "N80",
		}
		err = repos.Received.Create(ctx, received)
		require.NoError(t, err)

		// Test workflow state operations
		currentState, err := repos.WorkflowState.GetCurrentState(ctx, received.WorkOrder)
		require.NoError(t, err)
		assert.NotNil(t, currentState)

		// Test state transition
		newState := models.StateProduction
		err = repos.WorkflowState.TransitionTo(ctx, received.WorkOrder, newState, "Moving to production")
		require.NoError(t, err)

		// Verify state changed
		updatedState, err := repos.WorkflowState.GetCurrentState(ctx, received.WorkOrder)
		require.NoError(t, err)
		assert.Equal(t, newState, *updatedState)
	})
}

// backend/test/integration/repository_integration_test.go

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oilgas-backend/internal/repository"
	"oilgas-backend/test/testutil"
)

func TestRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repos := repository.New(db)
	ctx := context.Background()

	t.Run("Customer Repository", func(t *testing.T) {
		// Test customer CRUD operations
		customer := &models.Customer{
			Name:           "Test Oil Company",
			BillingAddress: "123 Test St",
			BillingCity:    "Houston",
			BillingState:   "TX",
			Phone:          "555-123-4567",
		}

		// Create
		err := repos.Customer.Create(ctx, customer)
		require.NoError(t, err)
		assert.NotZero(t, customer.ID)

		// Read
		retrieved, err := repos.Customer.GetByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, customer.Name, retrieved.Name)

		// Update
		customer.Name = "Updated Oil Company"
		err = repos.Customer.Update(ctx, customer)
		require.NoError(t, err)

		// Verify update
		updated, err := repos.Customer.GetByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Oil Company", updated.Name)

		// Delete
		err = repos.Customer.Delete(ctx, customer.ID)
		require.NoError(t, err)
	})

	t.Run("Received Repository", func(t *testing.T) {
		// First create a customer
		customer := &models.Customer{Name: "Test Customer"}
		err := repos.Customer.Create(ctx, customer)
		require.NoError(t, err)

		received := &models.ReceivedItem{
			WorkOrder:    "WO-TEST-001",
			CustomerID:   customer.ID,
			Customer:     customer.Name,
			Joints:       100,
			Size:         "5 1/2\"",
			Weight:       "20",
			Grade:        "J55",
			Connection:   "LTC",
			DateReceived: timePtr(time.Now()),
		}

		// Create
		err = repos.Received.Create(ctx, received)
		require.NoError(t, err)
		assert.NotZero(t, received.ID)

		// Read
		retrieved, err := repos.Received.GetByID(ctx, received.ID)
		require.NoError(t, err)
		assert.Equal(t, received.WorkOrder, retrieved.WorkOrder)

		// Get by work order
		byWorkOrder, err := repos.Received.GetByWorkOrder(ctx, received.WorkOrder)
		require.NoError(t, err)
		assert.Equal(t, received.ID, byWorkOrder.ID)

		// Update status
		err = repos.Received.UpdateStatus(ctx, received.ID, "in_production")
		require.NoError(t, err)

		// Get filtered
		filters := map[string]interface{}{
			"customer_id": customer.ID,
		}
		items, total, err := repos.Received.GetFiltered(ctx, filters, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, items, 1)

		// Delete
		err = repos.Received.Delete(ctx, received.ID)
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
			Grade:       "TEST123",
			Description: "Test grade for integration testing",
		}
		err = repos.Grade.Create(ctx, testGrade.Grade, testGrade.Description)
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
		customer := &models.Customer{Name: "Workflow Test Customer"}
		err := repos.Customer.Create(ctx, customer)
		require.NoError(t, err)

		received := &models.ReceivedItem{
			WorkOrder:  "WO-WORKFLOW-001",
			CustomerID: customer.ID,
			Customer:   customer.Name,
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

		// Test state history
		history, err := repos.WorkflowState.GetStateHistory(ctx, received.WorkOrder)
		require.NoError(t, err)
		assert.Len(t, history, 2) // Initial state + transition

		// Test get items by state
		items, err := repos.WorkflowState.GetItemsByState(ctx, newState)
		require.NoError(t, err)
		assert.Contains(t, items, received.WorkOrder)
	})
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// backend/test/integration/helpers.go
package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
)

// IntegrationHelper provides common test helper functions
type IntegrationHelper struct {
	repos *repository.Repositories
	ctx   context.Context
}

func NewIntegrationHelper(repos *repository.Repositories) *IntegrationHelper {
	return &IntegrationHelper{
		repos: repos,
		ctx:   context.Background(),
	}
}

// CreateCompleteWorkflow creates a complete workflow from received to inventory
func (h *IntegrationHelper) CreateCompleteWorkflow(t *testing.T, customer *models.Customer, workOrderSuffix string) (*models.ReceivedItem, *models.InspectionItem, *models.InventoryItem) {
	// Create received item
	received := &models.ReceivedItem{
		WorkOrder:    fmt.Sprintf("WO-COMPLETE-%s", workOrderSuffix),
		CustomerID:   customer.CustomerID,
		Customer:     customer.Customer,
		Joints:       100,
		Size:         "5 1/2\"",
		Grade:        "J55",
		Connection:   "LTC",
		Color:        "RED",
		Location:     "Test Yard",
		DateReceived: timePtr(time.Now()),
	}
	
	err := h.repos.Received.Create(h.ctx, received)
	require.NoError(t, err)

	// Move through workflow states
	err = h.repos.WorkflowState.TransitionTo(h.ctx, received.WorkOrder, models.StateInspection, "Moving to inspection")
	require.NoError(t, err)

	// Create inspection
	inspection := &models.InspectionItem{
		WorkOrder:      received.WorkOrder,
		CustomerID:     received.CustomerID,
		Customer:       received.Customer,
		Joints:         received.Joints,
		Size:           received.Size,
		Grade:          received.Grade,
		PassedJoints:   95,
		FailedJoints:   5,
		InspectionDate: timePtr(time.Now().Add(time.Hour)),
		Inspector:      "Test Inspector",
		Notes:          "Standard inspection completed",
	}
	
	err = h.repos.Inspected.Create(h.ctx, inspection)
	require.NoError(t, err)

	// Move to production
	err = h.repos.WorkflowState.TransitionTo(h.ctx, received.WorkOrder, models.StateProduction, "Moving to production")
	require.NoError(t, err)

	// Create inventory item
	inventory := &models.InventoryItem{
		CustomerID: received.CustomerID,
		Customer:   received.Customer,
		Joints:     inspection.PassedJoints,
		Size:       received.Size,
		Weight:     received.Weight,
		Grade:      received.Grade,
		Connection: received.Connection,
		Color:      "PROCESSED",
		Location:   "Main Storage",
		DateIn:     timePtr(time.Now().Add(2 * time.Hour)),
	}
	
	err = h.repos.Inventory.Create(h.ctx, inventory)
	require.NoError(t, err)

	return received, inspection, inventory
}

// SetupMultiCustomerData creates a realistic multi-customer dataset
func (h *IntegrationHelper) SetupMultiCustomerData(t *testing.T) ([]*models.Customer, []*models.ReceivedItem, []*models.InspectionItem) {
	fixtures := NewTestFixtures()
	customers := fixtures.StandardCustomers()
	
	// Create customers
	for _, customer := range customers {
		err := h.repos.Customer.Create(h.ctx, customer)
		require.NoError(t, err)
	}

	var allReceived []*models.ReceivedItem
	var allInspections []*models.InspectionItem

	// Create data for each customer
	for i, customer := range customers {
		itemCount := 3 + i // Varying amounts per customer
		receivedItems := fixtures.ReceivedItemsForCustomer(customer, itemCount)
		
		// Create received items
		for _, item := range receivedItems {
			err := h.repos.Received.Create(h.ctx, item)
			require.NoError(t, err)
			allReceived = append(allReceived, item)
		}

		// Create inspections for some items (varying pass rates by customer)
		passRate := 0.85 + float64(i)*0.05 // 85%, 90%, 95%
		inspections := fixtures.InspectionItemsFromReceived(receivedItems[:itemCount-1], passRate) // Leave one uninspected
		
		for _, inspection := range inspections {
			err := h.repos.Inspected.Create(h.ctx, inspection)
			require.NoError(t, err)
			allInspections = append(allInspections, inspection)
		}
	}

	return customers, allReceived, allInspections
}

// VerifyDataConsistency checks that data relationships are maintained
func (h *IntegrationHelper) VerifyDataConsistency(t *testing.T) {
	// Check that all received items have valid customers
	received, _, err := h.repos.Received.GetFiltered(h.ctx, map[string]interface{}{}, 1000, 0)
	require.NoError(t, err)
	
	for _, item := range received {
		customer, err := h.repos.Customer.GetByID(h.ctx, item.CustomerID)
		require.NoError(t, err)
		require.Equal(t, customer.Customer, item.Customer)
	}

	// Check that all inspections reference valid received items
	inspections, _, err := h.repos.Inspected.GetFiltered(h.ctx, map[string]interface{}{}, 1000, 0)
	require.NoError(t, err)
	
	for _, inspection := range inspections {
		receivedItem, err := h.repos.Received.GetByWorkOrder(h.ctx, inspection.WorkOrder)
		require.NoError(t, err)
		require.Equal(t, receivedItem.CustomerID, inspection.CustomerID)
		require.Equal(t, receivedItem.Joints, inspection.Joints)
	}

	// Check that workflow states exist for all received items
	for _, item := range received {
		state, err := h.repos.WorkflowState.GetCurrentState(h.ctx, item.WorkOrder)
		require.NoError(t, err)
		require.NotNil(t, state)
	}
}

// CleanupTestData removes all test data (useful for teardown)
func (h *IntegrationHelper) CleanupTestData(t *testing.T) {
	tables := []string{
		"store.workflow_state_history",
		"store.workflow_states", 
		"store.inventory",
		"store.inspected",
		"store.received",
		"store.customers",
	}

	for _, table := range tables {
		_, err := h.repos.Customer.(*customerRepository).db.Exec(h.ctx, fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err)
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// backend/test/unit/services_test.go

package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/cache"
)

func TestReceivedService(t *testing.T) {
	mockRepo := new(MockReceivedRepository)
	cache := cache.New(cache.Config{
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		MaxSize:         100,
	})
	
	service := services.NewReceivedService(mockRepo, cache)
	ctx := context.Background()

	t.Run("GetByID with cache", func(t *testing.T) {
		expectedItem := &models.ReceivedItem{
			ID:        1,
			WorkOrder: "WO-TEST-001",
			Customer:  "Test Customer",
		}

		mockRepo.On("GetByID", ctx, 1).Return(expectedItem, nil).Once()

		// First call - should hit repository
		item, err := service.GetByID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, expectedItem, item)

		// Second call - should hit cache (no additional repo call)
		item2, err := service.GetByID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, expectedItem, item2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Create with cache invalidation", func(t *testing.T) {
		newItem := &models.ReceivedItem{
			WorkOrder: "WO-TEST-002",
			Customer:  "Test Customer",
		}

		mockRepo.On("Create", ctx, newItem).Return(nil).Once()

		err := service.Create(ctx, newItem)
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		mockRepo.On("UpdateStatus", ctx, 1, "in_production").Return(nil).Once()

		err := service.UpdateStatus(ctx, 1, "in_production")
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("GetPendingInspection with cache", func(t *testing.T) {
		expectedItems := []models.ReceivedItem{
			{ID: 1, WorkOrder: "WO-001"},
			{ID: 2, WorkOrder: "WO-002"},
		}

		mockRepo.On("GetPendingInspection", ctx).Return(expectedItems, nil).Once()

		// First call
		items, err := service.GetPendingInspection(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedItems, items)

		// Second call - should hit cache
		items2, err := service.GetPendingInspection(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedItems, items2)

		mockRepo.AssertExpectations(t)
	})
}

func TestWorkflowStateService(t *testing.T) {
	mockRepo := new(MockWorkflowStateRepository)
	cache := cache.New(cache.Config{
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		MaxSize:         100,
	})
	
	service := services.NewWorkflowStateService(mockRepo, cache)
	ctx := context.Background()

	t.Run("GetCurrentState", func(t *testing.T) {
		workOrder := "WO-TEST-001"
		expectedState := models.StateProduction

		mockRepo.On("GetCurrentState", ctx, workOrder).Return(&expectedState, nil).Once()

		state, err := service.GetCurrentState(ctx, workOrder)
		require.NoError(t, err)
		assert.Equal(t, &expectedState, state)

		mockRepo.AssertExpectations(t)
	})

	t.Run("TransitionTo valid transition", func(t *testing.T) {
		workOrder := "WO-TEST-001"
		newState := models.StateInspection
		notes := "Moving to inspection"

		mockRepo.On("TransitionTo", ctx, workOrder, newState, notes).Return(nil).Once()

		err := service.TransitionTo(ctx, workOrder, newState, notes)
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}

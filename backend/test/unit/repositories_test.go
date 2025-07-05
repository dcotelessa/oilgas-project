// backend/test/unit/repositories_test.go

package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
)

// Mock implementations for your actual repositories

type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.DashboardStats), args.Error(1)
}

func (m *MockAnalyticsRepository) GetCustomerActivity(ctx context.Context, days int, customerID *int) (*models.CustomerActivity, error) {
	args := m.Called(ctx, days, customerID)
	return args.Get(0).(*models.CustomerActivity), args.Error(1)
}

type MockReceivedRepository struct {
	mock.Mock
}

func (m *MockReceivedRepository) GetFiltered(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.ReceivedItem, int, error) {
	args := m.Called(ctx, filters, limit, offset)
	return args.Get(0).([]models.ReceivedItem), args.Int(1), args.Error(2)
}

func (m *MockReceivedRepository) GetByID(ctx context.Context, id int) (*models.ReceivedItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ReceivedItem), args.Error(1)
}

func (m *MockReceivedRepository) Create(ctx context.Context, item *models.ReceivedItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockReceivedRepository) Update(ctx context.Context, item *models.ReceivedItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockReceivedRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReceivedRepository) GetByWorkOrder(ctx context.Context, workOrder string) (*models.ReceivedItem, error) {
	args := m.Called(ctx, workOrder)
	return args.Get(0).(*models.ReceivedItem), args.Error(1)
}

func (m *MockReceivedRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockReceivedRepository) GetPendingInspection(ctx context.Context) ([]models.ReceivedItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.ReceivedItem), args.Error(1)
}

type MockWorkflowStateRepository struct {
	mock.Mock
}

func (m *MockWorkflowStateRepository) GetCurrentState(ctx context.Context, workOrder string) (*models.WorkflowState, error) {
	args := m.Called(ctx, workOrder)
	return args.Get(0).(*models.WorkflowState), args.Error(1)
}

func (m *MockWorkflowStateRepository) TransitionTo(ctx context.Context, workOrder string, newState models.WorkflowState, notes string) error {
	args := m.Called(ctx, workOrder, newState, notes)
	return args.Error(0)
}

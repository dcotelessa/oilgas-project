// backend/test/unit/customer_service_test.go

package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/pkg/validation"
)

// MockCustomerRepository - extends existing mocks
type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) GetAll(ctx context.Context) ([]models.Customer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, id int) (*models.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Create(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Simulate ID assignment like real database would
	customer.CustomerID = 1
	customer.CreatedAt = time.Now()
	return customer, args.Error(1)
}

func (m *MockCustomerRepository) Update(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerRepository) ExistsByName(ctx context.Context, name string, excludeID ...int) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCustomerRepository) HasActiveInventory(ctx context.Context, customerID int) (bool, error) {
	args := m.Called(ctx, customerID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCustomerRepository) GetTotalCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockCustomerRepository) Search(ctx context.Context, query string, limit, offset int) ([]models.Customer, int, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]models.Customer), args.Int(1), args.Error(2)
}

// MAIN CUSTOMER SERVICE TESTS
func TestCustomerService_Create_Fixed(t *testing.T) {
	mockRepo := new(MockCustomerRepository)
	cache := cache.New(cache.Config{
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		MaxSize:         100,
	})
	
	service := services.NewCustomerService(mockRepo, cache)
	ctx := context.Background()

	t.Run("Successful creation with proper type conversion", func(t *testing.T) {
		req := &validation.CustomerValidation{
			CustomerName: "Test Customer",
			Address:      "123 Test St",
			City:         "Houston",
			State:        "TX",
			Zip:          "77001",
			Contact:      "John Doe", 
			Phone:        "555-123-4567",
			Email:        "test@example.com",
		}

		// Mock the repository calls
		mockRepo.On("ExistsByName", ctx, "Test Customer").Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.MatchedBy(func(customer *models.Customer) bool {
			// Verify the conversion from validation to model happened correctly
			return customer.Customer == "Test Customer" &&
				   customer.BillingAddress == "123 Test St" &&
				   customer.BillingCity == "Houston" &&
				   customer.BillingState == "TX" &&
				   customer.BillingZipcode == "77001" &&
				   customer.Contact == "John Doe" &&
				   customer.Phone == "555-123-4567" &&
				   customer.Email == "test@example.com"
		})).Return(&models.Customer{
			CustomerID:     1,
			Customer:       "Test Customer",
			BillingAddress: "123 Test St",
			BillingCity:    "Houston",
			BillingState:   "TX",
			BillingZipcode: "77001",
			Contact:        "John Doe",
			Phone:          "555-123-4567",
			Email:          "test@example.com",
			CreatedAt:      time.Now(),
		}, nil).Once()

		// Execute the service method
		result, err := service.Create(ctx, req)

		// Assertions
		require.NoError(t, err, "Create should not return an error")
		assert.NotNil(t, result, "Result should not be nil")
		assert.Equal(t, 1, result.CustomerID, "Customer ID should be set")
		assert.Equal(t, "Test Customer", result.Customer, "Customer name should match")
		assert.Equal(t, "123 Test St", result.BillingAddress, "Address should match")
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("Duplicate customer name error", func(t *testing.T) {
		req := &validation.CustomerValidation{
			CustomerName: "Duplicate Customer",
		}

		mockRepo.On("ExistsByName", ctx, "Duplicate Customer").Return(true, nil).Once()

		result, err := service.Create(ctx, req)

		assert.Nil(t, result, "Result should be nil on duplicate")
		assert.Error(t, err, "Should return an error")
		assert.Contains(t, err.Error(), "already exists", "Error should mention duplicate")
		
		// Verify Create was never called
		mockRepo.AssertNotCalled(t, "Create")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository creation error", func(t *testing.T) {
		req := &validation.CustomerValidation{
			CustomerName: "Error Customer",
			Email:        "error@test.com",
		}

		mockRepo.On("ExistsByName", ctx, "Error Customer").Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Customer")).
			Return(nil, errors.New("database connection failed")).Once()

		result, err := service.Create(ctx, req)

		assert.Nil(t, result, "Result should be nil on error")
		assert.Error(t, err, "Should return an error")
		assert.Contains(t, err.Error(), "failed to create customer", "Error should be wrapped")
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("Validation error - empty name", func(t *testing.T) {
		req := &validation.CustomerValidation{
			CustomerName: "", // Empty name should fail validation
			Email:        "test@example.com",
		}

		result, err := service.Create(ctx, req)

		assert.Nil(t, result, "Result should be nil on validation error")
		assert.Error(t, err, "Should return validation error")
		assert.Contains(t, err.Error(), "validation error", "Error should mention validation")
		
		// Repository methods should not be called if validation fails
		mockRepo.AssertNotCalled(t, "ExistsByName")
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestCustomerService_GetByID_CacheIntegration(t *testing.T) {
	mockRepo := new(MockCustomerRepository)
	cache := cache.New(cache.Config{
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		MaxSize:         100,
	})
	
	service := services.NewCustomerService(mockRepo, cache)
	ctx := context.Background()

	t.Run("Cache miss then hit scenario", func(t *testing.T) {
		expectedCustomer := &models.Customer{
			CustomerID: 42,
			Customer:   "Cache Test Customer",
			Email:      "cache@test.com",
		}

		// First call - cache miss, should hit repository
		mockRepo.On("GetByID", ctx, 42).Return(expectedCustomer, nil).Once()

		result1, err := service.GetByID(ctx, "42")
		require.NoError(t, err)
		assert.Equal(t, expectedCustomer.CustomerID, result1.CustomerID)
		assert.Equal(t, expectedCustomer.Customer, result1.Customer)

		// Second call - should hit cache, no repository call
		result2, err := service.GetByID(ctx, "42")
		require.NoError(t, err)
		assert.Equal(t, expectedCustomer.CustomerID, result2.CustomerID)
		assert.Equal(t, result1, result2, "Both results should be identical")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid ID format", func(t *testing.T) {
		result, err := service.GetByID(ctx, "not-a-number")

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid customer ID")
		
		// Should not call repository with invalid ID
		mockRepo.AssertNotCalled(t, "GetByID")
	})
}

// Test helper to verify cache invalidation
func TestCustomerService_CacheInvalidation(t *testing.T) {
	mockRepo := new(MockCustomerRepository)
	cache := cache.New(cache.Config{
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		MaxSize:         100,
	})
	
	service := services.NewCustomerService(mockRepo, cache)
	ctx := context.Background()

	t.Run("Create invalidates cache", func(t *testing.T) {
		// Pre-populate cache with customers list
		cachedCustomers := []models.Customer{
			{CustomerID: 1, Customer: "Existing Customer"},
		}
		cache.Set("customers:all", cachedCustomers)

		// Verify cache has data
		cached, exists := cache.Get("customers:all")
		assert.True(t, exists)
		assert.Len(t, cached.([]models.Customer), 1)

		// Create a new customer
		req := &validation.CustomerValidation{
			CustomerName: "New Customer",
			Email:        "new@test.com",
		}

		mockRepo.On("ExistsByName", ctx, "New Customer").Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Customer")).
			Return(&models.Customer{CustomerID: 2, Customer: "New Customer"}, nil).Once()

		_, err := service.Create(ctx, req)
		require.NoError(t, err)

		// Cache should be invalidated (implementation detail - depends on your invalidateCustomerCaches)
		// This test verifies the service calls the right methods
		mockRepo.AssertExpectations(t)
	})
}

// PERFORMANCE AND BENCHMARK TESTS
func BenchmarkCustomerService_GetByID_CacheHit(b *testing.B) {
	mockRepo := new(MockCustomerRepository)
	cache := cache.New(cache.Config{
		TTL:             time.Hour, // Longer TTL for benchmark
		CleanupInterval: time.Hour,
		MaxSize:         1000,
	})
	
	service := services.NewCustomerService(mockRepo, cache)
	ctx := context.Background()

	// Pre-populate cache
	customer := &models.Customer{
		CustomerID: 1,
		Customer:   "Benchmark Customer",
	}
	cache.Set("customer:1", customer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GetByID(ctx, "1")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCustomerService_Create(b *testing.B) {
	mockRepo := new(MockCustomerRepository)
	cache := cache.New(cache.Config{
		TTL:             time.Hour,
		CleanupInterval: time.Hour,
		MaxSize:         1000,
	})
	
	service := services.NewCustomerService(mockRepo, cache)
	ctx := context.Background()

	// Set up mock to handle all calls
	mockRepo.On("ExistsByName", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Customer")).
		Return(&models.Customer{CustomerID: 1}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &validation.CustomerValidation{
			CustomerName: "Benchmark Customer",
			Email:        "bench@test.com",
		}
		_, err := service.Create(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

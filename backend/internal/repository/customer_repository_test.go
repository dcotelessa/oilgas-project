// backend/internal/repository/customer_repository_test.go
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oilgas-backend/internal/models"
)

func TestCustomerRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCustomerRepository(mock)
	ctx := context.Background()

	t.Run("customer found", func(t *testing.T) {
		customerID := 123
		expectedCustomer := &models.Customer{
			CustomerID:      customerID,
			Customer:        "Test Oil Company",
			BillingAddress:  "123 Test St",
			BillingCity:     "Houston",
			BillingState:    "TX",
			BillingZipcode:  "77001",
			Contact:         "John Doe",
			Phone:           "555-0123",
			Email:           "john@testoil.com",
			Deleted:         false,
			CreatedAt:       time.Now(),
		}

		// Mock the query
		rows := pgxmock.NewRows([]string{
			"customer_id", "customer", "billing_address", "billing_city", "billing_state",
			"billing_zipcode", "contact", "phone", "fax", "email",
			"color1", "color2", "color3", "color4", "color5",
			"loss1", "loss2", "loss3", "loss4", "loss5",
			"wscolor1", "wscolor2", "wscolor3", "wscolor4", "wscolor5",
			"wsloss1", "wsloss2", "wsloss3", "wsloss4", "wsloss5",
			"deleted", "created_at",
		}).AddRow(
			expectedCustomer.CustomerID, expectedCustomer.Customer, expectedCustomer.BillingAddress,
			expectedCustomer.BillingCity, expectedCustomer.BillingState, expectedCustomer.BillingZipcode,
			expectedCustomer.Contact, expectedCustomer.Phone, "", expectedCustomer.Email,
			"", "", "", "", "", "", "", "", "", "",
			"", "", "", "", "", "", "", "", "", "",
			expectedCustomer.Deleted, expectedCustomer.CreatedAt,
		)

		mock.ExpectQuery(`SELECT customer_id, customer, billing_address`).
			WithArgs(customerID).
			WillReturnRows(rows)

		// Execute
		result, err := repo.GetByID(ctx, customerID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedCustomer.CustomerID, result.CustomerID)
		assert.Equal(t, expectedCustomer.Customer, result.Customer)
		assert.Equal(t, expectedCustomer.BillingCity, result.BillingCity)
		assert.Equal(t, expectedCustomer.Email, result.Email)
		
		// Ensure all expectations were met
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("customer not found", func(t *testing.T) {
		customerID := 999

		mock.ExpectQuery(`SELECT customer_id, customer, billing_address`).
			WithArgs(customerID).
			WillReturnError(pgxmock.ErrNoRows)

		// Execute
		result, err := repo.GetByID(ctx, customerID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "customer not found")
		
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCustomerRepository(mock)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		customer := &models.Customer{
			Customer:       "New Oil Company",
			BillingAddress: "456 New St",
			BillingCity:    "Dallas",
			BillingState:   "TX",
			BillingZipcode: "75001",
			Contact:        "Jane Smith",
			Phone:          "555-0456",
			Email:          "jane@newoil.com",
		}

		expectedID := 124
		expectedTime := time.Now()

		// Mock the INSERT query
		mock.ExpectQuery(`INSERT INTO store.customers`).
			WithArgs(
				customer.Customer, customer.BillingAddress, customer.BillingCity,
				customer.BillingState, customer.BillingZipcode, customer.Contact,
				customer.Phone, customer.Fax, customer.Email,
				customer.Color1, customer.Color2, customer.Color3, customer.Color4, customer.Color5,
				customer.Loss1, customer.Loss2, customer.Loss3, customer.Loss4, customer.Loss5,
				customer.WSColor1, customer.WSColor2, customer.WSColor3, customer.WSColor4, customer.WSColor5,
				customer.WSLoss1, customer.WSLoss2, customer.WSLoss3, customer.WSLoss4, customer.WSLoss5,
			).
			WillReturnRows(pgxmock.NewRows([]string{"customer_id", "created_at"}).
				AddRow(expectedID, expectedTime))

		// Execute
		err := repo.Create(ctx, customer)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedID, customer.CustomerID)
		assert.Equal(t, expectedTime, customer.CreatedAt)
		
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate customer error", func(t *testing.T) {
		customer := &models.Customer{
			Customer: "Duplicate Company",
		}

		// Mock duplicate error
		mock.ExpectQuery(`INSERT INTO store.customers`).
			WillReturnError(fmt.Errorf("duplicate key value violates unique constraint"))

		// Execute
		err := repo.Create(ctx, customer)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create customer")
		
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_ExistsByName(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCustomerRepository(mock)
	ctx := context.Background()

	t.Run("customer exists", func(t *testing.T) {
		customerName := "Existing Company"

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(customerName).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

		// Execute
		exists, err := repo.ExistsByName(ctx, customerName)

		// Assert
		require.NoError(t, err)
		assert.True(t, exists)
		
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("customer does not exist", func(t *testing.T) {
		customerName := "Non-existent Company"

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(customerName).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

		// Execute
		exists, err := repo.ExistsByName(ctx, customerName)

		// Assert
		require.NoError(t, err)
		assert.False(t, exists)
		
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exclude specific ID", func(t *testing.T) {
		customerName := "Test Company"
		excludeID := 5

		mock.ExpectQuery(`SELECT EXISTS.*AND customer_id != \$2`).
			WithArgs(customerName, excludeID).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

		// Execute
		exists, err := repo.ExistsByName(ctx, customerName, excludeID)

		// Assert
		require.NoError(t, err)
		assert.False(t, exists)
		
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_HasActiveInventory(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCustomerRepository(mock)
	ctx := context.Background()

	t.Run("has active inventory", func(t *testing.T) {
		customerID := 123

		mock.ExpectQuery(`SELECT EXISTS.*FROM store.inventory.*WHERE customer_id = \$1`).
			WithArgs(customerID).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

		// Execute
		hasInventory, err := repo.HasActiveInventory(ctx, customerID)

		// Assert
		require.NoError(t, err)
		assert.True(t, hasInventory)
		
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no active inventory", func(t *testing.T) {
		customerID := 456

		mock.ExpectQuery(`SELECT EXISTS.*FROM store.inventory.*WHERE customer_id = \$1`).
			WithArgs(customerID).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

		// Execute
		hasInventory, err := repo.HasActiveInventory(ctx, customerID)

		// Assert
		require.NoError(t, err)
		assert.False(t, hasInventory)
		
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_Search(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCustomerRepository(mock)
	ctx := context.Background()

	t.Run("successful search", func(t *testing.T) {
		query := "oil"
		limit := 10
		offset := 0
		expectedCount := 2

		// Mock count query
		mock.ExpectQuery(`SELECT COUNT.*FROM store.customers.*WHERE deleted = false AND`).
			WithArgs("%oil%").
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(expectedCount))

		// Mock search results
		searchRows := pgxmock.NewRows([]string{
			"customer_id", "customer", "billing_address", "billing_city", "billing_state",
			"billing_zipcode", "contact", "phone", "fax", "email",
			"color1", "color2", "color3", "color4", "color5",
			"loss1", "loss2", "loss3", "loss4", "loss5",
			"wscolor1", "wscolor2", "wscolor3", "wscolor4", "wscolor5",
			"wsloss1", "wsloss2", "wsloss3", "wsloss4", "wsloss5",
			"deleted", "created_at",
		}).
			AddRow(1, "First Oil Co", "123 St", "Houston", "TX", "77001", 
				"John", "555-0001", "", "john@first.com",
				"", "", "", "", "", "", "", "", "", "",
				"", "", "", "", "", "", "", "", "", "",
				false, time.Now()).
			AddRow(2, "Second Oil Services", "456 Ave", "Dallas", "TX", "75001", 
				"Jane", "555-0002", "", "jane@second.com",
				"", "", "", "", "", "", "", "", "", "",
				"", "", "", "", "", "", "", "", "", "",
				false, time.Now())

		mock.ExpectQuery(`SELECT customer_id, customer.*ORDER BY.*LIMIT.*OFFSET`).
			WithArgs("%oil%", limit, offset).
			WillReturnRows(searchRows)

		// Execute
		customers, total, err := repo.Search(ctx, query, limit, offset)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedCount, total)
		assert.Len(t, customers, 2)
		assert.Contains(t, customers[0].Customer, "Oil")
		assert.Contains(t, customers[1].Customer, "Oil")
		
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

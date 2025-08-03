// backend/internal/customer/service_test.go
// Test suite aligned with existing Customer struct
package customer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCustomer(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		customer *Customer
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid customer",
			customer: &Customer{
				Customer: "Test Company",
				TenantID: "local-dev",
				Email:    stringPtr("test@example.com"),
				BillingState: stringPtr("TX"),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			customer: &Customer{
				TenantID: "local-dev",
			},
			wantErr: true,
			errMsg:  "customer name is required",
		},
		{
			name: "missing tenant ID",
			customer: &Customer{
				Customer: "Test Company",
			},
			wantErr: true,
			errMsg:  "tenant ID is required",
		},
		{
			name: "invalid email",
			customer: &Customer{
				Customer: "Test Company",
				TenantID: "local-dev",
				Email:    stringPtr("invalid-email"),
			},
			wantErr: true,
			errMsg:  "invalid email address format",
		},
		{
			name: "invalid state code",
			customer: &Customer{
				Customer:     "Test Company",
				TenantID:     "local-dev",
				BillingState: stringPtr("TEXAS"),
			},
			wantErr: true,
			errMsg:  "billing state should be 2-letter code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateCustomer(tt.customer)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user@domain.org", true},
		{"invalid", false},
		{"no-at-symbol.com", false},
		{"no-dot@domain", false},
		{"@domain.com", false},
		{"user@.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestNewService(t *testing.T) {
	// This test doesn't require a real database connection
	// We're just testing the constructor
	service := NewService(nil)
	assert.NotNil(t, service)
	assert.Nil(t, service.db) // Since we passed nil
}

func TestCustomerFilter(t *testing.T) {
	filter := CustomerFilter{
		TenantID:       "local-dev",
		State:          "TX",
		Search:         "oil",
		IncludeDeleted: false,
		Limit:          10,
		Offset:         0,
	}

	assert.Equal(t, "local-dev", filter.TenantID)
	assert.Equal(t, "TX", filter.State)
	assert.Equal(t, "oil", filter.Search)
	assert.False(t, filter.IncludeDeleted)
	assert.Equal(t, 10, filter.Limit)
}

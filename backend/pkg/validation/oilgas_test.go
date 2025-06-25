// backend/pkg/validation/oilgas_test.go
package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGrade(t *testing.T) {
	tests := []struct {
		name        string
		grade       string
		expectError bool
	}{
		{"Valid J55", "J55", false},
		{"Valid L80", "L80", false},
		{"Valid P110", "P110", false},
		{"Valid lowercase", "j55", false},
		{"Valid with spaces", " N80 ", false},
		{"Invalid grade", "INVALID", true},
		{"Empty grade", "", true},
		{"Number only", "55", true},
		{"Valid Q125", "Q125", false},
		{"Valid S135", "S135", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGrade(tt.grade)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSize(t *testing.T) {
	tests := []struct {
		name        string
		size        string
		expectError bool
	}{
		{"Valid 5.5 inch", "5 1/2\"", false},
		{"Valid 7 inch", "7\"", false},
		{"Valid 9 5/8", "9 5/8\"", false},
		{"Valid 4.5 inch", "4 1/2\"", false},
		{"Invalid size", "12.5\"", true},
		{"Empty size", "", true},
		{"No quotes", "5.5", true}, // Should fail without quotes
		{"Invalid format", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSize(tt.size)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWeight(t *testing.T) {
	tests := []struct {
		name        string
		weight      string
		expectError bool
	}{
		{"Valid weight 20", "20", false},
		{"Valid weight 26.4", "26.4", false},
		{"Valid with units", "20 lbs/ft", false},
		{"Valid with # symbol", "20#", false},
		{"Valid with lb/ft", "26.4 lb/ft", false},
		{"Invalid text", "heavy", true},
		{"Empty weight", "", true},
		{"Negative weight", "-5", true},
		{"Too high weight", "500", true},
		{"Too low weight", "1", true},
		{"Valid float", "32.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWeight(tt.weight)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConnection(t *testing.T) {
	tests := []struct {
		name        string
		connection  string
		expectError bool
	}{
		{"Valid LTC", "LTC", false},
		{"Valid BTC", "BTC", false},
		{"Valid EUE", "EUE", false},
		{"Valid lowercase", "ltc", false},
		{"Valid with spaces", " BTC ", false},
		{"Valid PREMIUM", "PREMIUM", false},
		{"Valid VAM", "VAM", false},
		{"Invalid connection", "INVALID", true},
		{"Empty connection", "", true},
		{"Numbers", "123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnection(tt.connection)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateJointsCount(t *testing.T) {
	tests := []struct {
		name        string
		joints      int
		expectError bool
	}{
		{"Valid count 100", 100, false},
		{"Valid count 1", 1, false},
		{"Valid count 5000", 5000, false},
		{"Zero joints", 0, true},
		{"Negative joints", -10, true},
		{"Too many joints", 20000, true},
		{"Valid high count", 9999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJointsCount(tt.joints)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryValidation(t *testing.T) {
	t.Run("Valid inventory item", func(t *testing.T) {
		inv := InventoryValidation{
			CustomerID: 123,
			Joints:     100,
			Size:       "5 1/2\"",
			Weight:     "20",
			Grade:      "J55",
			Connection: "LTC",
		}

		errors := inv.Validate()
		assert.Empty(t, errors)
	})

	t.Run("Invalid inventory item", func(t *testing.T) {
		inv := InventoryValidation{
			CustomerID: -1,        // Invalid
			Joints:     0,         // Invalid
			Size:       "invalid", // Invalid
			Weight:     "heavy",   // Invalid
			Grade:      "WRONG",   // Invalid
			Connection: "BAD",     // Invalid
		}

		errors := inv.Validate()
		assert.Len(t, errors, 6) // Should have 6 validation errors

		// Check specific error fields
		fields := make(map[string]bool)
		for _, err := range errors {
			fields[err.Field] = true
		}

		assert.True(t, fields["customer_id"])
		assert.True(t, fields["joints"])
		assert.True(t, fields["size"])
		assert.True(t, fields["weight"])
		assert.True(t, fields["grade"])
		assert.True(t, fields["connection"])
	})

	t.Run("Partial validation errors", func(t *testing.T) {
		inv := InventoryValidation{
			CustomerID: 123,   // Valid
			Joints:     100,   // Valid
			Size:       "7\"", // Valid
			Weight:     "26.4", // Valid
			Grade:      "INVALID", // Invalid
			Connection: "LTC",     // Valid
		}

		errors := inv.Validate()
		assert.Len(t, errors, 1)
		assert.Equal(t, "grade", errors[0].Field)
	})
}

func TestCustomerValidation(t *testing.T) {
	t.Run("Valid customer", func(t *testing.T) {
		customer := CustomerValidation{
			Name:    "Acme Oil Company",
			Address: "123 Oil Street",
			City:    "Houston",
			State:   "TX",
			Zipcode: "77001",
			Phone:   "555-123-4567",
			Email:   "contact@acmeoil.com",
		}

		errors := customer.Validate()
		assert.Empty(t, errors)
	})

	t.Run("Valid customer minimal", func(t *testing.T) {
		customer := CustomerValidation{
			Name: "Oil Co",
		}

		errors := customer.Validate()
		assert.Empty(t, errors)
	})

	t.Run("Invalid customer", func(t *testing.T) {
		customer := CustomerValidation{
			Name:    "", // Invalid - required
			Address: strings.Repeat("a", 60), // Invalid - too long
			City:    strings.Repeat("b", 60), // Invalid - too long
			Phone:   "123", // Invalid - too short
			Email:   "invalid-email", // Invalid format
			State:   "ZZ", // Invalid state
			Zipcode: "123", // Invalid format
		}

		errors := customer.Validate()
		assert.Len(t, errors, 7)

		fields := make(map[string]bool)
		for _, err := range errors {
			fields[err.Field] = true
		}

		assert.True(t, fields["name"])
		assert.True(t, fields["address"])
		assert.True(t, fields["city"])
		assert.True(t, fields["phone"])
		assert.True(t, fields["email"])
		assert.True(t, fields["state"])
		assert.True(t, fields["zipcode"])
	})
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		expectError bool
	}{
		{"Valid 10 digit", "5551234567", false},
		{"Valid 11 digit", "15551234567", false},
		{"Valid with formatting", "(555) 123-4567", false},
		{"Valid with dots", "555.123.4567", false},
		{"Valid with spaces", "555 123 4567", false},
		{"Too short", "123456", true},
		{"Too long", "555123456789", true},
		{"Empty", "", true},
		{"Letters", "555abcd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhone(tt.phone)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{"Valid simple", "test@example.com", false},
		{"Valid with subdomain", "user@mail.example.com", false},
		{"Valid with plus", "user+tag@example.com", false},
		{"Valid with numbers", "user123@example123.com", false},
		{"Invalid no @", "testexample.com", true},
		{"Invalid no domain", "test@", true},
		{"Invalid no TLD", "test@example", true},
		{"Invalid double @", "test@@example.com", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateState(t *testing.T) {
	tests := []struct {
		name        string
		state       string
		expectError bool
	}{
		{"Valid TX", "TX", false},
		{"Valid Texas", "TEXAS", false},
		{"Valid lowercase", "tx", false},
		{"Valid OK", "OK", false},
		{"Valid Oklahoma", "OKLAHOMA", false},
		{"Valid with spaces", " TX ", false},
		{"Invalid state", "ZZ", true},
		{"Non oil state", "HI", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateState(tt.state)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateZipcode(t *testing.T) {
	tests := []struct {
		name        string
		zipcode     string
		expectError bool
	}{
		{"Valid 5 digit", "77001", false},
		{"Valid 9 digit", "77001-1234", false},
		{"Invalid 4 digit", "1234", true},
		{"Invalid 6 digit", "123456", true},
		{"Invalid format", "12345-123", true},
		{"Letters", "abcde", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateZipcode(tt.zipcode)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizationFunctions(t *testing.T) {
	t.Run("NormalizeGrade", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"j55", "J55"},
			{" L80 ", "L80"},
			{"p110", "P110"},
			{"N80", "N80"},
		}

		for _, tt := range tests {
			result := NormalizeGrade(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("NormalizeSize", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"5.5", "5.5\""},
			{"7\"", "7\""},
			{" 9 5/8\" ", "9 5/8\""},
			{"4.5 in", "4.5 in"},
		}

		for _, tt := range tests {
			result := NormalizeSize(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("NormalizeConnection", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"ltc", "LTC"},
			{" BTC ", "BTC"},
			{"eue", "EUE"},
			{"PREMIUM", "PREMIUM"},
		}

		for _, tt := range tests {
			result := NormalizeConnection(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})
}

// Benchmark tests for validation performance
func BenchmarkValidateGrade(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateGrade("J55")
	}
}

func BenchmarkValidateInventory(b *testing.B) {
	inv := InventoryValidation{
		CustomerID: 123,
		Joints:     100,
		Size:       "5 1/2\"",
		Weight:     "20",
		Grade:      "J55",
		Connection: "LTC",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inv.Validate()
	}
}

func BenchmarkValidateCustomer(b *testing.B) {
	customer := CustomerValidation{
		Name:    "Test Oil Company",
		Address: "123 Test Street",
		City:    "Houston", 
		State:   "TX",
		Zipcode: "77001",
		Phone:   "555-123-4567",
		Email:   "test@oilcompany.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		customer.Validate()
	}
}

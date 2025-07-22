// tools/internal/validation/validator_test.go
package validation

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestBusinessRuleValidator(t *testing.T) {
	validator := NewBusinessRuleValidator()

	t.Run("work_order_format", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected bool
		}{
			{"Valid LB format", "LB-001001", true},
			{"Valid without dash", "LB001001", true},
			{"Invalid prefix", "XX-001001", false},
			{"Too short", "LB-001", false},
			{"Too long", "LB-0010011", false},
			{"No numbers", "LB-ABCDEF", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validator.ValidateWorkOrder(tt.input)
				assert.Equal(t, tt.expected, result.IsValid)
			})
		}
	})

	t.Run("customer_name_normalization", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"chevron corp", "Chevron Corp"},
			{"EXXON MOBIL", "Exxon Mobil"},
			{"ConocoPhillips Company", "ConocoPhillips Company"},
			{"test customer llc", "Test Customer LLC"},
		}

		for _, tt := range tests {
			result := validator.NormalizeCustomerName(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})
}

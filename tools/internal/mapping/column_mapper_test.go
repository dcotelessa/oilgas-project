// tools/internal/mapping/column_mapper_test.go
package mapping

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGradeNormalizer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		valid    bool
	}{
		{"Standard J55", "J55", "J55", true},
		{"Hyphenated J-55", "J-55", "J55", true},
		{"Lowercase j55", "j55", "J55", true},
		{"With spaces", " J 55 ", "J55", true},
		{"L80 variant", "L-80", "L80", true},
		{"Invalid grade", "X99", "", false},
		{"Empty input", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := NormalizeGrade(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestSizeNormalizer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		valid    bool
	}{
		{"Decimal to fraction", "5.5", "5 1/2\"", true},
		{"With inch symbol", "5.5\"", "5 1/2\"", true},
		{"With inch word", "5.5 inch", "5 1/2\"", true},
		{"Already fractional", "5 1/2\"", "5 1/2\"", true},
		{"Large pipe", "20", "20\"", true},
		{"Invalid size", "99.99", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := NormalizeSize(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

package validation

import (
	"fmt"
)

// Validator handles basic validation
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRecord performs basic record validation
func (v *Validator) ValidateRecord(record []string) []string {
	var issues []string
	for i, field := range record {
		if len(field) > 1000 {
			issues = append(issues, fmt.Sprintf("Field %d too long", i))
		}
	}
	return issues
}

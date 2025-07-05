// backend/pkg/validation/oilgas.go
package validation

import (
	"fmt"
	"strconv"
	"strings"
)

// Oil & Gas Industry Standards Validation

// Standard pipe grades used in oil & gas industry
var ValidGrades = map[string]bool{
	"J55":   true,
	"JZ55":  true,
	"K55":   true,
	"L80":   true,
	"N80":   true,
	"P105":  true,
	"P110":  true,
	"Q125":  true,
	"T95":   true,
	"C90":   true,
	"C95":   true,
	"S135":  true,
}

// Standard pipe sizes (OD in inches)
var ValidSizes = map[string]bool{
	"2 3/8\"":  true,
	"2 7/8\"":  true,
	"3 1/2\"":  true,
	"4\"":      true,
	"4 1/2\"":  true,
	"5\"":      true,
	"5 1/2\"":  true,
	"6 5/8\"":  true,
	"7\"":      true,
	"7 5/8\"":  true,
	"8 5/8\"":  true,
	"9 5/8\"":  true,
	"10 3/4\"": true,
	"11 3/4\"": true,
	"13 3/8\"": true,
	"16\"":     true,
	"18 5/8\"": true,
	"20\"":     true,
}

// Standard connection types
var ValidConnections = map[string]bool{
	"LTC":    true, // Long Thread Casing
	"STC":    true, // Short Thread Casing
	"BTC":    true, // Buttress Thread Casing
	"EUE":    true, // External Upset End
	"NUE":    true, // Non-upset End
	"PREMIUM": true, // Premium connections
	"VAM":    true, // VAM connection
	"NEW VAM": true,
	"TENARIS": true,
}

// Validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Value)
}

// Inventory item validation
type InventoryValidation struct {
	CustomerID int    `json:"customer_id"`
	Joints     int    `json:"joints"`
	Size       string `json:"size"`
	Weight     string `json:"weight"`
	Grade      string `json:"grade"`
	Connection string `json:"connection"`
	Color      string `json:"color,omitempty"`
	Location   string `json:"location,omitempty"`
}

func (iv *InventoryValidation) Validate() []ValidationError {
	var errors []ValidationError

	// Validate grade
	if err := ValidateGrade(iv.Grade); err != nil {
		errors = append(errors, ValidationError{
			Field:   "grade",
			Value:   iv.Grade,
			Message: err.Error(),
		})
	}

	// Validate size
	if err := ValidateSize(iv.Size); err != nil {
		errors = append(errors, ValidationError{
			Field:   "size",
			Value:   iv.Size,
			Message: err.Error(),
		})
	}

	// Validate weight
	if err := ValidateWeight(iv.Weight); err != nil {
		errors = append(errors, ValidationError{
			Field:   "weight",
			Value:   iv.Weight,
			Message: err.Error(),
		})
	}

	// Validate connection
	if err := ValidateConnection(iv.Connection); err != nil {
		errors = append(errors, ValidationError{
			Field:   "connection",
			Value:   iv.Connection,
			Message: err.Error(),
		})
	}

	// Validate joints count
	if err := ValidateJointsCount(iv.Joints); err != nil {
		errors = append(errors, ValidationError{
			Field:   "joints",
			Value:   fmt.Sprintf("%d", iv.Joints),
			Message: err.Error(),
		})
	}

	// Validate customer ID
	if iv.CustomerID <= 0 {
		errors = append(errors, ValidationError{
			Field:   "customer_id",
			Value:   fmt.Sprintf("%d", iv.CustomerID),
			Message: "customer ID must be positive",
		})
	}

	return errors
}

// Individual validation functions
func ValidateGrade(grade string) error {
	if grade == "" {
		return fmt.Errorf("grade is required")
	}
	
	normalizedGrade := strings.ToUpper(strings.TrimSpace(grade))
	if !ValidGrades[normalizedGrade] {
		return fmt.Errorf("invalid grade - must be one of: J55, JZ55, K55, L80, N80, P105, P110, Q125, T95, C90, C95, S135")
	}
	
	return nil
}

func ValidateSize(size string) error {
	if size == "" {
		return fmt.Errorf("size is required")
	}
	
	normalizedSize := strings.TrimSpace(size)
	if !ValidSizes[normalizedSize] {
		// Try common variations
		variations := []string{
			normalizedSize + "\"",
			strings.Replace(normalizedSize, "in", "\"", -1),
			strings.Replace(normalizedSize, " in", "\"", -1),
		}
		
		for _, variation := range variations {
			if ValidSizes[variation] {
				return nil
			}
		}
		
		return fmt.Errorf("invalid pipe size - common sizes: 4 1/2\", 5 1/2\", 7\", 9 5/8\", etc.")
	}
	
	return nil
}

func ValidateWeight(weight string) error {
	if weight == "" {
		return fmt.Errorf("weight is required")
	}
	
	// Remove common suffixes and normalize
	normalizedWeight := strings.TrimSpace(weight)
	normalizedWeight = strings.Replace(normalizedWeight, " lbs/ft", "", -1)
	normalizedWeight = strings.Replace(normalizedWeight, " lb/ft", "", -1)
	normalizedWeight = strings.Replace(normalizedWeight, "lbs/ft", "", -1)
	normalizedWeight = strings.Replace(normalizedWeight, "lb/ft", "", -1)
	normalizedWeight = strings.Replace(normalizedWeight, "#", "", -1)
	normalizedWeight = strings.TrimSpace(normalizedWeight)
	
	// Try to parse as float
	if weightFloat, err := strconv.ParseFloat(normalizedWeight, 64); err != nil {
		return fmt.Errorf("weight must be a valid number (e.g., 20.0, 26.4, 32.3)")
	} else {
		// Reasonable weight range for oil & gas tubing/casing
		if weightFloat < 4.0 || weightFloat > 200.0 {
			return fmt.Errorf("weight %.1f lbs/ft seems unrealistic for oil & gas pipe", weightFloat)
		}
	}
	
	return nil
}

func ValidateConnection(connection string) error {
	if connection == "" {
		return fmt.Errorf("connection type is required")
	}
	
	normalizedConnection := strings.ToUpper(strings.TrimSpace(connection))
	if !ValidConnections[normalizedConnection] {
		return fmt.Errorf("invalid connection type - must be one of: LTC, STC, BTC, EUE, NUE, PREMIUM, VAM")
	}
	
	return nil
}

func ValidateJointsCount(joints int) error {
	if joints <= 0 {
		return fmt.Errorf("joints count must be positive")
	}
	
	if joints > 10000 {
		return fmt.Errorf("joints count %d seems unrealistically high", joints)
	}
	
	return nil
}

// Helper function to normalize grades for database storage
func NormalizeGrade(grade string) string {
	return strings.ToUpper(strings.TrimSpace(grade))
}

// Helper function to normalize sizes
func NormalizeSize(size string) string {
	normalized := strings.TrimSpace(size)
	// Ensure proper quote format
	if !strings.Contains(normalized, "\"") && !strings.Contains(normalized, "in") {
		normalized += "\""
	}
	return normalized
}

// Helper function to normalize connections
func NormalizeConnection(connection string) string {
	return strings.ToUpper(strings.TrimSpace(connection))
}

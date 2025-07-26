package utils

import (
	"strings"
)

// Common oil & gas industry grades
var ValidGrades = map[string]string{
	"J55":  "Standard grade steel casing",
	"JZ55": "Enhanced J55 grade",
	"L80":  "Higher strength grade",
	"N80":  "Medium strength grade",
	"P105": "High performance grade",
	"P110": "Premium performance grade",
	"Q125": "Ultra-high strength grade",
	"C75":  "Carbon steel grade",
	"C95":  "Higher carbon steel grade",
	"T95":  "Tough grade for harsh environments",
}

// ValidateGrade checks if grade is a valid oil & gas grade
func ValidateGrade(grade string) bool {
	_, exists := ValidGrades[strings.ToUpper(grade)]
	return exists
}

// NormalizeGrade converts grade to standard format
func NormalizeGrade(grade string) string {
	return strings.ToUpper(strings.TrimSpace(grade))
}

// Common pipe size formats
func NormalizePipeSize(size string) string {
	size = strings.TrimSpace(size)
	// Add common normalization rules here
	return size
}

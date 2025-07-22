package mapping

import (
	"regexp"
	"strconv"
	"strings"
	"oilgas-tools/internal/config"
)

// ColumnMapper handles column name normalization and data transformation
type ColumnMapper struct {
	config *config.Config
}

// NewColumnMapper creates a new column mapper
func NewColumnMapper(config *config.Config) *ColumnMapper {
	return &ColumnMapper{
		config: config,
	}
}

// NormalizeHeaders normalizes column headers using industry mappings
func (cm *ColumnMapper) NormalizeHeaders(headers []string) []string {
	normalized := make([]string, len(headers))
	for i, header := range headers {
		normalized[i] = cm.normalizeColumnName(header)
	}
	return normalized
}

// normalizeColumnName converts a column name to PostgreSQL-friendly format
func (cm *ColumnMapper) normalizeColumnName(colName string) string {
	if colName == "" {
		return colName
	}

	// Basic normalization
	normalized := strings.ToLower(strings.TrimSpace(colName))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	
	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	normalized = reg.ReplaceAllString(normalized, "_")
	
	// Collapse multiple underscores
	reg = regexp.MustCompile(`_+`)
	normalized = reg.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")

	return normalized
}

// NormalizeGrade normalizes oil & gas grade values
func (cm *ColumnMapper) NormalizeGrade(grade string) (string, bool) {
	if grade == "" {
		return "", false
	}

	// Clean input
	cleaned := strings.ToUpper(strings.TrimSpace(grade))
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	// Standard grades
	gradeMap := map[string]string{
		"J55": "J55", "L80": "L80", "N80": "N80", "P110": "P110",
	}

	if normalized, exists := gradeMap[cleaned]; exists {
		return normalized, true
	}
	return "", false
}

// NormalizeSize normalizes pipe size values
func (cm *ColumnMapper) NormalizeSize(size string) (string, bool) {
	if size == "" {
		return "", false
	}

	cleaned := strings.TrimSpace(size)
	cleaned = strings.ReplaceAll(cleaned, " inch", "")
	cleaned = strings.ReplaceAll(cleaned, "\"", "")

	// Handle decimal to fraction
	if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
		whole := int(val)
		decimal := val - float64(whole)
		
		if decimal >= 0.4 && decimal <= 0.6 {
			return strconv.Itoa(whole) + " 1/2\"", true
		}
		return strconv.Itoa(whole) + "\"", true
	}

	return "", false
}

// NormalizeCustomerName normalizes customer names
func (cm *ColumnMapper) NormalizeCustomerName(name string) string {
	if name == "" {
		return ""
	}
	return strings.Title(strings.ToLower(strings.TrimSpace(name)))
}

// NormalizeConnection normalizes connection types
func (cm *ColumnMapper) NormalizeConnection(connection string) string {
	if connection == "" {
		return ""
	}

	cleaned := strings.ToUpper(strings.TrimSpace(connection))
	connectionMap := map[string]string{
		"BUTTRESS THREAD": "BTC",
		"LONG THREAD":     "LTC",
		"SHORT THREAD":    "STC",
	}

	if normalized, exists := connectionMap[cleaned]; exists {
		return normalized
	}
	return connection
}

// internal/mapping/column_mapper.go
// Column mapping and transformation logic for oil & gas industry
package mapping

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tools/internal/config"
	"tools/internal/processor"
)

// ColumnMapper handles column mapping and data transformation
type ColumnMapper struct {
	config           *config.Config
	gradeNormalizer  *GradeNormalizer
	sizeNormalizer   *SizeNormalizer
	connectionMapper *ConnectionMapper
	customerMapper   *CustomerMapper
	workOrderMapper  *WorkOrderMapper
}

// New creates a new column mapper
func New(cfg *config.Config) *ColumnMapper {
	return &ColumnMapper{
		config:           cfg,
		gradeNormalizer:  NewGradeNormalizer(cfg.BusinessRules.ValidGrades),
		sizeNormalizer:   NewSizeNormalizer(cfg.BusinessRules.ValidSizes),
		connectionMapper: NewConnectionMapper(cfg.BusinessRules.ValidConnections),
		customerMapper:   NewCustomerMapper(cfg.BusinessRules.CustomerRules),
		workOrderMapper:  NewWorkOrderMapper(cfg.BusinessRules.WorkOrderFormat),
	}
}

// TransformTable transforms an entire table's data according to mappings
func (cm *ColumnMapper) TransformTable(table processor.TableInfo, data []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(data) == 0 {
		return data, nil
	}

	var transformedData []map[string]interface{}
	
	for i, record := range data {
		transformedRecord, err := cm.transformRecord(table, record, i+1)
		if err != nil {
			// Log error but continue processing
			continue
		}
		
		if transformedRecord != nil {
			transformedData = append(transformedData, transformedRecord)
		}
	}
	
	return transformedData, nil
}

// transformRecord transforms a single record
func (cm *ColumnMapper) transformRecord(table processor.TableInfo, record map[string]interface{}, recordNum int) (map[string]interface{}, error) {
	transformed := make(map[string]interface{})
	recordID := fmt.Sprintf("%s_record_%d", table.Name, recordNum)
	
	for _, column := range table.Columns {
		originalValue, exists := record[column.OriginalName]
		if !exists {
			// Handle missing columns
			if cm.isRequiredColumn(column) {
				return nil, fmt.Errorf("required column %s missing in record %s", column.OriginalName, recordID)
			}
			// Use default value if available
			if column.DefaultValue != nil {
				transformed[column.Name] = *column.DefaultValue
			}
			continue
		}
		
		// Transform the value
		transformedValue, err := cm.transformValue(column, originalValue, recordID)
		if err != nil {
			return nil, fmt.Errorf("transformation failed for %s in record %s: %w", column.Name, recordID, err)
		}
		
		transformed[column.Name] = transformedValue
	}
	
	return transformed, nil
}

// transformValue transforms a single value according to column configuration
func (cm *ColumnMapper) transformValue(column processor.ColumnInfo, value interface{}, recordID string) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	
	stringValue := fmt.Sprintf("%v", value)
	stringValue = strings.TrimSpace(stringValue)
	
	// Handle empty strings
	if stringValue == "" {
		if column.IsNullable {
			return nil, nil
		}
		if column.DefaultValue != nil {
			return *column.DefaultValue, nil
		}
		return nil, fmt.Errorf("empty value for non-nullable column")
	}
	
	// Apply business rule transformations
	for _, rule := range column.BusinessRules {
		var err error
		stringValue, err = cm.applyBusinessRule(rule, stringValue, recordID)
		if err != nil {
			return nil, fmt.Errorf("business rule %s failed: %w", rule, err)
		}
	}
	
	// Convert to target PostgreSQL type
	return cm.convertToPostgreSQLType(column.PostgreSQLType, stringValue)
}

// applyBusinessRule applies a specific business rule transformation
func (cm *ColumnMapper) applyBusinessRule(rule string, value string, recordID string) (string, error) {
	switch rule {
	case "validate_grade":
		return cm.gradeNormalizer.Normalize(value)
	case "validate_size":
		return cm.sizeNormalizer.Normalize(value)
	case "validate_connection":
		return cm.connectionMapper.Normalize(value)
	case "normalize_customer_name":
		return cm.customerMapper.NormalizeName(value)
	case "validate_work_order_format":
		return cm.workOrderMapper.Normalize(value)
	case "normalize_phone":
		return normalizePhoneNumber(value), nil
	case "normalize_email":
		return normalizeEmail(value), nil
	case "uppercase":
		return strings.ToUpper(value), nil
	case "lowercase":
		return strings.ToLower(value), nil
	case "trim_whitespace":
		return strings.TrimSpace(value), nil
	case "remove_special_chars":
		return removeSpecialCharacters(value), nil
	default:
		return value, nil // Unknown rule, pass through
	}
}

// convertToPostgreSQLType converts string value to appropriate PostgreSQL type
func (cm *ColumnMapper) convertToPostgreSQLType(pgType string, value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}
	
	switch strings.ToUpper(pgType) {
	case "INTEGER", "INT", "SERIAL":
		return strconv.Atoi(value)
	
	case "BIGINT", "BIGSERIAL":
		return strconv.ParseInt(value, 10, 64)
	
	case "DECIMAL", "NUMERIC":
		return strconv.ParseFloat(value, 64)
		
	case "BOOLEAN", "BOOL":
		return parseBoolean(value)
		
	case "DATE":
		return parseDate(value)
		
	case "TIMESTAMP", "TIMESTAMPTZ":
		return parseTimestamp(value)
		
	case "VARCHAR", "TEXT", "CHAR":
		return value, nil
		
	default:
		// Default to string for unknown types
		return value, nil
	}
}

// isRequiredColumn checks if a column is required
func (cm *ColumnMapper) isRequiredColumn(column processor.ColumnInfo) bool {
	// Check if column is marked as required in business rules
	for _, rule := range cm.config.BusinessRules.CustomerRules.RequiredFields {
		if rule == column.Name {
			return true
		}
	}
	
	// Check if it's a primary key (typically required)
	if strings.HasSuffix(column.Name, "_id") || column.Name == "id" {
		return true
	}
	
	return false
}

// GradeNormalizer handles oil & gas grade normalization
type GradeNormalizer struct {
	validGrades map[string]config.GradeInfo
}

func NewGradeNormalizer(grades []config.GradeInfo) *GradeNormalizer {
	gradeMap := make(map[string]config.GradeInfo)
	for _, grade := range grades {
		gradeMap[strings.ToUpper(grade.Code)] = grade
	}
	return &GradeNormalizer{validGrades: gradeMap}
}

func (gn *GradeNormalizer) Normalize(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	
	// Direct match
	if grade, exists := gn.validGrades[normalized]; exists && !grade.Deprecated {
		return grade.Code, nil
	}
	
	// Try common variations
	variations := map[string]string{
		"J-55":  "J55",
		"JZ-55": "JZ55",
		"L-80":  "L80",
		"N-80":  "N80",
		"P-105": "P105",
		"P-110": "P110",
		"Q-125": "Q125",
		"C-75":  "C75",
		"C-95":  "C95",
		"T-95":  "T95",
	}
	
	if standardForm, exists := variations[normalized]; exists {
		if grade, validExists := gn.validGrades[standardForm]; validExists && !grade.Deprecated {
			return grade.Code, nil
		}
	}
	
	// No valid grade found
	return value, fmt.Errorf("invalid grade: %s", value)
}

// SizeNormalizer handles pipe size normalization
type SizeNormalizer struct {
	validSizes map[string]config.SizeInfo
}

func NewSizeNormalizer(sizes []config.SizeInfo) *SizeNormalizer {
	sizeMap := make(map[string]config.SizeInfo)
	for _, size := range sizes {
		sizeMap[size.Size] = size
	}
	return &SizeNormalizer{validSizes: sizeMap}
}

func (sn *SizeNormalizer) Normalize(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	
	// Direct match
	if _, exists := sn.validSizes[normalized]; exists {
		return normalized, nil
	}
	
	// Try to parse and standardize format
	standardized := standardizePipeSize(normalized)
	if _, exists := sn.validSizes[standardized]; exists {
		return standardized, nil
	}
	
	return value, fmt.Errorf("invalid pipe size: %s", value)
}

// ConnectionMapper handles connection type mapping
type ConnectionMapper struct {
	validConnections map[string]config.ConnectionInfo
}

func NewConnectionMapper(connections []config.ConnectionInfo) *ConnectionMapper {
	connMap := make(map[string]config.ConnectionInfo)
	for _, conn := range connections {
		connMap[strings.ToUpper(conn.Type)] = conn
	}
	return &ConnectionMapper{validConnections: connMap}
}

func (cm *ConnectionMapper) Normalize(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	
	// Direct match
	if conn, exists := cm.validConnections[normalized]; exists && !conn.Deprecated {
		return conn.Type, nil
	}
	
	// Try common variations
	variations := map[string]string{
		"BUTTRESS THREAD CASING": "BTC",
		"LONG THREAD CASING":     "LTC",
		"SHORT THREAD CASING":    "STC",
		"VAM-TOP":               "VAM TOP",
		"VAM_TOP":               "VAM TOP",
		"VAM-ACE":               "VAM ACE",
		"VAM_ACE":               "VAM ACE",
		"NEW-VAM":               "NEW VAM",
		"NEW_VAM":               "NEW VAM",
		"TSH-BLUE":              "TSH BLUE",
		"TSH_BLUE":              "TSH BLUE",
	}
	
	if standardForm, exists := variations[normalized]; exists {
		if conn, validExists := cm.validConnections[standardForm]; validExists && !conn.Deprecated {
			return conn.Type, nil
		}
	}
	
	return value, fmt.Errorf("invalid connection type: %s", value)
}

// CustomerMapper handles customer name normalization
type CustomerMapper struct {
	rules config.CustomerRules
}

func NewCustomerMapper(rules config.CustomerRules) *CustomerMapper {
	return &CustomerMapper{rules: rules}
}

func (cm *CustomerMapper) NormalizeName(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	
	// Remove extra whitespace
	re := regexp.MustCompile(`\s+`)
	normalized = re.ReplaceAllString(normalized, " ")
	
	// Apply normalization rule
	switch cm.rules.NameNormalization {
	case "uppercase":
		normalized = strings.ToUpper(normalized)
	case "lowercase":
		normalized = strings.ToLower(normalized)
	case "title_case":
		normalized = strings.Title(strings.ToLower(normalized))
	case "uppercase_first_letter":
		if len(normalized) > 0 {
			normalized = strings.ToUpper(normalized[:1]) + normalized[1:]
		}
	}
	
	// Remove special characters that might cause issues
	re = regexp.MustCompile(`[^\w\s&\.\-,]`)
	normalized = re.ReplaceAllString(normalized, "")
	
	return normalized, nil
}

// WorkOrderMapper handles work order format normalization
type WorkOrderMapper struct {
	format config.WorkOrderFormat
	regex  *regexp.Regexp
}

func NewWorkOrderMapper(format config.WorkOrderFormat) *WorkOrderMapper {
	regex, _ := regexp.Compile(format.Pattern)
	return &WorkOrderMapper{
		format: format,
		regex:  regex,
	}
}

func (wm *WorkOrderMapper) Normalize(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	
	// Check if already in correct format
	if wm.regex != nil && wm.regex.MatchString(normalized) {
		return normalized, nil
	}
	
	// Try to extract and reformat
	// Look for pattern like: letters followed by numbers
	re := regexp.MustCompile(`^([A-Z]+)[-_\s]*(\d+)$`)
	matches := re.FindStringSubmatch(normalized)
	
	if len(matches) == 3 {
		prefix := matches[1]
		number := matches[2]
		
		// Pad number to required length
		if len(number) < wm.format.ZeroPad {
			number = fmt.Sprintf("%0*s", wm.format.ZeroPad, number)
		}
		
		formatted := fmt.Sprintf("%s-%s", prefix, number)
		
		// Validate formatted version
		if wm.regex != nil && wm.regex.MatchString(formatted) {
			return formatted, nil
		}
	}
	
	return value, fmt.Errorf("invalid work order format: %s (expected pattern: %s)", value, wm.format.Pattern)
}

// Utility functions

func parseBoolean(value string) (bool, error) {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch lower {
	case "true", "t", "yes", "y", "1", "on":
		return true, nil
	case "false", "f", "no", "n", "0", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

func parseDate(value string) (time.Time, error) {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"02-01-2006",
		"2-1-2006",
		"2006/01/02",
		"2006/1/2",
		"January 2, 2006",
		"Jan 2, 2006",
		"2 January 2006",
		"2 Jan 2006",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", value)
}

func parseTimestamp(value string) (time.Time, error) {
	// Try common timestamp formats
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"01/02/2006 15:04:05",
		"1/2/2006 3:04:05 PM",
		"2006-01-02 15:04:05.000",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", value)
}

func normalizePhoneNumber(value string) string {
	// Remove all non-digit characters
	re := regexp.MustCompile(`[^\d]`)
	digits := re.ReplaceAllString(value, "")
	
	// Format as (XXX) XXX-XXXX if US number
	if len(digits) == 10 {
		return fmt.Sprintf("(%s) %s-%s", digits[0:3], digits[3:6], digits[6:10])
	}
	if len(digits) == 11 && digits[0] == '1' {
		return fmt.Sprintf("1 (%s) %s-%s", digits[1:4], digits[4:7], digits[7:11])
	}
	
	// Return original if not standard format
	return value
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func removeSpecialCharacters(value string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\s\-_.]`)
	return re.ReplaceAllString(value, "")
}

func standardizePipeSize(value string) string {
	// Try to parse size like "5.5", "5 1/2", "5.5 inch", etc.
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "inch", "")
	value = strings.ReplaceAll(value, "in", "")
	value = strings.ReplaceAll(value, "\"", "")
	value = strings.TrimSpace(value)
	
	// Handle fractional sizes
	fractionMap := map[string]string{
		"1/2": ".5",
		"1/4": ".25",
		"3/4": ".75",
		"3/8": ".375",
		"5/8": ".625",
		"7/8": ".875",
	}
	
	for fraction, decimal := range fractionMap {
		if strings.Contains(value, fraction) {
			value = strings.ReplaceAll(value, fraction, decimal)
		}
	}
	
	// Convert back to standard format with quote
	if val, err := strconv.ParseFloat(value, 64); err == nil {
		if val == float64(int(val)) {
			return fmt.Sprintf("%.0f\"", val)
		} else {
			// Convert decimal back to fraction for standard format
			if val == 4.5 {
				return "4 1/2\""
			}
			if val == 5.5 {
				return "5 1/2\""
			}
			if val == 8.625 {
				return "8 5/8\""
			}
			if val == 9.625 {
				return "9 5/8\""
			}
			if val == 10.75 {
				return "10 3/4\""
			}
			if val == 13.375 {
				return "13 3/8\""
			}
			if val == 18.625 {
				return "18 5/8\""
			}
		}
	}
	
	return value + "\""
}

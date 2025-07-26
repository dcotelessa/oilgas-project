// tools/internal/config/models.go
// Adds multi-tenant support and lowercase table mapping
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the complete configuration for MDB processing
type Config struct {
	Company           string                      `json:"company"`
	TenantID          string                      `json:"tenant_id"`
	DatabaseURL       string                      `json:"database_url"`
	OilGasMappings    map[string]ColumnMapping    `json:"oil_gas_mappings"`
	ProcessingOptions ProcessingOptions           `json:"processing_options"`
	DatabaseConfig    DatabaseConfig              `json:"database_config"`
	ValidationRules   ValidationRules             `json:"validation_rules"`
	// Multi-tenant specific settings
	TenantSettings    TenantSettings              `json:"tenant_settings"`
	TableMappings     map[string]TableMapping     `json:"table_mappings"`
	SequenceConfig    SequenceConfig              `json:"sequence_config"`
}

// TenantSettings holds tenant-specific configuration
type TenantSettings struct {
	DefaultTenant     string            `json:"default_tenant"`
	TenantDatabases   map[string]string `json:"tenant_databases"`
	TenantPrefixes    map[string]string `json:"tenant_prefixes"`
	IsolationMode     string            `json:"isolation_mode"` // "database" or "schema"
}

// TableMapping defines how to handle table name transformations
type TableMapping struct {
	SourceName      string            `json:"source_name"`      // Original table name (e.g., "RECEIVED")
	TargetName      string            `json:"target_name"`      // Target table name (e.g., "received")
	ColumnMappings  map[string]string `json:"column_mappings"`  // Column name mappings
	IsCounterTable  bool              `json:"is_counter_table"` // RNUMBER, WKNUMBER tables
	SequenceName    string            `json:"sequence_name"`    // Target sequence name
}

// SequenceConfig handles sequence generation settings
type SequenceConfig struct {
	RNumberStart    int64  `json:"r_number_start"`
	WorkOrderStart  int64  `json:"work_order_start"`
	WorkOrderPrefix string `json:"work_order_prefix"`
	WorkOrderFormat string `json:"work_order_format"`
}

// ColumnMapping defines how to transform a column
type ColumnMapping struct {
	SourceColumn string                 `json:"source_column"`
	TargetColumn string                 `json:"target_column"`
	DataType     string                 `json:"data_type"`
	Required     bool                   `json:"required"`
	Rules        []TransformationRule   `json:"rules"`
	TenantSpecific map[string]interface{} `json:"tenant_specific,omitempty"`
}

// TransformationRule defines data transformation rules
type TransformationRule struct {
	Type        string                 `json:"type"`        // "normalize", "validate", "format"
	Parameters  map[string]interface{} `json:"parameters"`
	Description string                 `json:"description"`
}

// ProcessingOptions controls processing behavior
type ProcessingOptions struct {
	BatchSize          int    `json:"batch_size"`
	Workers            int    `json:"workers"`
	ContinueOnError    bool   `json:"continue_on_error"`
	ValidateBeforeImport bool  `json:"validate_before_import"`
	CreateTenantDB     bool   `json:"create_tenant_db"`
	MigrateTenant      bool   `json:"migrate_tenant"`
	BackupBeforeImport bool   `json:"backup_before_import"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"database_name"`
	SSLMode      string `json:"ssl_mode"`
	MaxConns     int    `json:"max_conns"`
	IdleConns    int    `json:"idle_conns"`
	TenantDBTemplate string `json:"tenant_db_template"` // Template for tenant database names
	AdminDatabase    string `json:"admin_database"`     // Admin database for creating tenants
}

// ValidationRules defines data validation rules
type ValidationRules struct {
	RequiredTables   []string          `json:"required_tables"`
	RequiredColumns  map[string][]string `json:"required_columns"`
	DataTypes        map[string]string `json:"data_types"`
	BusinessRules    []BusinessRule    `json:"business_rules"`
	TenantValidation map[string][]BusinessRule `json:"tenant_validation"`
}

// BusinessRule defines a business validation rule
type BusinessRule struct {
	Name        string                 `json:"name"`
	Table       string                 `json:"table"`
	Column      string                 `json:"column"`
	RuleType    string                 `json:"rule_type"`    // "regex", "range", "enum", "custom"
	Parameters  map[string]interface{} `json:"parameters"`
	ErrorMessage string                `json:"error_message"`
	Severity    string                 `json:"severity"`     // "error", "warning", "info"
}

// Load loads configuration from file with enhanced error handling
func Load(filename string) (*Config, error) {
	// Default configuration
	config := &Config{
		TenantSettings: TenantSettings{
			DefaultTenant: "location_longbeach",
			IsolationMode: "database",
			TenantDatabases: map[string]string{
				"location_longbeach": "oilgas_location_longbeach",
				"location_lasvegas":  "oilgas_location_lasvegas",
				"location_colorado":  "oilgas_location_colorado",
			},
		},
		TableMappings: getDefaultTableMappings(),
		SequenceConfig: SequenceConfig{
			RNumberStart:    1000,
			WorkOrderStart:  1000,
			WorkOrderPrefix: "WO",
			WorkOrderFormat: "WO-%06d",
		},
		ProcessingOptions: ProcessingOptions{
			BatchSize:            1000,
			Workers:              4,
			ContinueOnError:      true,
			ValidateBeforeImport: true,
			CreateTenantDB:       false,
			MigrateTenant:        true,
			BackupBeforeImport:   true,
		},
		DatabaseConfig: DatabaseConfig{
			Host:             "localhost",
			Port:             5432,
			SSLMode:          "disable",
			MaxConns:         25,
			IdleConns:        10,
			TenantDBTemplate: "oilgas_%s",
			AdminDatabase:    "postgres",
		},
		OilGasMappings: getDefaultOilGasMappings(),
		ValidationRules: getDefaultValidationRules(),
	}

	// Load from file if it exists
	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return config, nil
}

// GetTenantDatabaseName returns the database name for a tenant
func (c *Config) GetTenantDatabaseName(tenantID string) string {
	if dbName, exists := c.TenantSettings.TenantDatabases[tenantID]; exists {
		return dbName
	}
	return fmt.Sprintf(c.DatabaseConfig.TenantDBTemplate, tenantID)
}

// GetTableMapping returns the mapping for a table
func (c *Config) GetTableMapping(tableName string) (TableMapping, bool) {
	mapping, exists := c.TableMappings[tableName]
	return mapping, exists
}

// IsCounterTable checks if a table is a counter table that should be converted to sequence
func (c *Config) IsCounterTable(tableName string) bool {
	if mapping, exists := c.TableMappings[tableName]; exists {
		return mapping.IsCounterTable
	}
	return false
}

// getDefaultTableMappings returns the default table mappings for oil & gas industry
func getDefaultTableMappings() map[string]TableMapping {
	return map[string]TableMapping{
		"RECEIVED": {
			SourceName: "RECEIVED",
			TargetName: "received",
			ColumnMappings: map[string]string{
				"ID":           "id",
				"WKORDER":      "work_order",
				"CUSTID":       "customer_id",
				"CUSTOMER":     "customer_name",
				"DATERECVD":    "date_received",
				"BILLTOID":     "bill_to_id",
				"WHEN1":        "when_entered",
				"WHEN2":        "when_updated",
				"COMPLETE":     "is_complete",
				"DELETED":      "is_deleted",
			},
		},
		"RNUMBER": {
			SourceName:     "RNUMBER",
			TargetName:     "", // No target table, becomes sequence
			IsCounterTable: true,
			SequenceName:   "r_number_seq",
		},
		"WKNUMBER": {
			SourceName:     "WKNUMBER", 
			TargetName:     "", // No target table, becomes sequence
			IsCounterTable: true,
			SequenceName:   "work_order_seq",
		},
		"customers": {
			SourceName: "customers",
			TargetName: "customers", // Already lowercase
			ColumnMappings: map[string]string{
				"custid":   "customer_id",
				"customer": "customer_name",
				"deleted":  "is_deleted",
			},
		},
		"fletcher": {
			SourceName: "fletcher",
			TargetName: "fletcher",
			ColumnMappings: map[string]string{
				"custid":    "customer_id",
				"customer":  "customer_name", 
				"orderedby": "ordered_by",
				"location":  "location_code",
				"complete":  "is_complete",
				"deleted":   "is_deleted",
			},
		},
		"inventory": {
			SourceName: "inventory",
			TargetName: "inventory",
			ColumnMappings: map[string]string{
				"wkorder":   "work_order",
				"custid":    "customer_id",
				"customer":  "customer_name",
				"orderedby": "ordered_by",
				"location":  "location_code",
				"deleted":   "is_deleted",
			},
		},
		"bakeout": {
			SourceName: "bakeout",
			TargetName: "bakeout",
			ColumnMappings: map[string]string{
				"custid": "customer_id",
				"datein": "date_in",
			},
		},
		"inspected": {
			SourceName: "inspected",
			TargetName: "inspected",
			ColumnMappings: map[string]string{
				"wkorder":  "work_order",
				"complete": "is_complete",
				"deleted":  "is_deleted",
			},
		},
	}
}

// getDefaultOilGasMappings returns industry-specific column mappings
func getDefaultOilGasMappings() map[string]ColumnMapping {
	return map[string]ColumnMapping{
		"customer_name": {
			SourceColumn: "customer",
			TargetColumn: "customer_name",
			DataType:     "string",
			Required:     true,
			Rules: []TransformationRule{
				{
					Type: "normalize",
					Parameters: map[string]interface{}{
						"case": "title",
						"trim": true,
					},
					Description: "Convert to title case and trim whitespace",
				},
			},
		},
		"grade": {
			SourceColumn: "grade",
			TargetColumn: "grade",
			DataType:     "string",
			Required:     false,
			Rules: []TransformationRule{
				{
					Type: "normalize",
					Parameters: map[string]interface{}{
						"case": "upper",
						"trim": true,
					},
					Description: "Convert to uppercase and trim",
				},
			},
		},
		"connection": {
			SourceColumn: "connection",
			TargetColumn: "connection", 
			DataType:     "string",
			Required:     false,
			Rules: []TransformationRule{
				{
					Type: "normalize",
					Parameters: map[string]interface{}{
						"case": "upper",
						"trim": true,
					},
					Description: "Convert to uppercase and trim",
				},
			},
		},
	}
}

// getDefaultValidationRules returns default validation rules
func getDefaultValidationRules() ValidationRules {
	return ValidationRules{
		RequiredTables: []string{"customers", "received"},
		RequiredColumns: map[string][]string{
			"customers": {"customer_id", "customer_name"},
			"received":  {"work_order", "customer_id"},
		},
		BusinessRules: []BusinessRule{
			{
				Name:         "valid_grade",
				Table:        "*",
				Column:       "grade",
				RuleType:     "enum",
				Parameters:   map[string]interface{}{
					"values": []string{"J55", "JZ55", "L80", "N80", "P105", "P110", "Q125", "C75", "C95", "T95"},
				},
				ErrorMessage: "Invalid grade value",
				Severity:     "warning",
			},
			{
				Name:         "customer_name_length",
				Table:        "customers",
				Column:       "customer_name",
				RuleType:     "range",
				Parameters:   map[string]interface{}{
					"min_length": 1,
					"max_length": 255,
				},
				ErrorMessage: "Customer name must be between 1 and 255 characters",
				Severity:     "error",
			},
		},
	}
}

// internal/config/models.go
// Configuration models for MDB processor
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the complete configuration for MDB processing
type Config struct {
	OilGasMappings     map[string]ColumnMapping `json:"oil_gas_mappings"`
	CompanyTemplate    CompanyTemplate          `json:"company_template"`
	ValidationRules    []ValidationRule         `json:"validation_rules"`
	OutputSettings     OutputSettings           `json:"output_settings"`
	ProcessingOptions  ProcessingOptions        `json:"processing_options"`
	DatabaseConfig     DatabaseConfig           `json:"database_config"`
	BusinessRules      BusinessRules            `json:"business_rules"`
}

// ColumnMapping defines how to map and transform a column
type ColumnMapping struct {
	PostgreSQLName   string            `json:"postgresql_name"`
	DataType         string            `json:"data_type"`
	BusinessRules    []string          `json:"business_rules"`
	Transformations  map[string]string `json:"transformations"`
	ValidationRules  []string          `json:"validation_rules"`
	Required         bool              `json:"required"`
	DefaultValue     *string           `json:"default_value,omitempty"`
	MaxLength        *int              `json:"max_length,omitempty"`
	AllowedValues    []string          `json:"allowed_values,omitempty"`
}

// CompanyTemplate defines company-specific configurations
type CompanyTemplate struct {
	CompanyName      string            `json:"company_name"`
	WorkOrderPrefix  string            `json:"work_order_prefix"`
	CustomMappings   map[string]string `json:"custom_mappings"`
	SpecialRules     []string          `json:"special_rules"`
	ContactInfo      ContactInfo       `json:"contact_info"`
	BusinessSettings BusinessSettings  `json:"business_settings"`
}

type ContactInfo struct {
	PrimaryContact string `json:"primary_contact"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
}

type BusinessSettings struct {
	FiscalYearStart    string            `json:"fiscal_year_start"`
	DefaultCurrency    string            `json:"default_currency"`
	TimeZone          string            `json:"time_zone"`
	CustomFields      map[string]string `json:"custom_fields"`
}

// ValidationRule defines a validation constraint
type ValidationRule struct {
	FieldName    string   `json:"field_name"`
	RuleType     string   `json:"rule_type"`     // "required", "format", "range", "enum", "custom"
	Parameters   []string `json:"parameters"`
	ErrorAction  string   `json:"error_action"` // "reject", "warn", "fix"
	ErrorMessage string   `json:"error_message"`
	Priority     int      `json:"priority"`     // 1=critical, 2=important, 3=warning
}

// OutputSettings controls what outputs are generated
type OutputSettings struct {
	CSVOutput         bool   `json:"csv_output"`
	SQLOutput         bool   `json:"sql_output"`
	PostgreSQLDirect  bool   `json:"postgresql_direct"`
	ValidationReport  bool   `json:"validation_report"`
	BusinessReport    bool   `json:"business_report"`
	PerformanceReport bool   `json:"performance_report"`
	OutputDir         string `json:"output_dir"`
	FileNaming        string `json:"file_naming"` // "timestamp", "sequential", "job_id"
}

// ProcessingOptions controls processing behavior
type ProcessingOptions struct {
	Workers           int  `json:"workers"`
	BatchSize         int  `json:"batch_size"`
	MemoryLimit       int  `json:"memory_limit_mb"`
	TimeoutMinutes    int  `json:"timeout_minutes"`
	DryRun           bool `json:"dry_run"`
	ContinueOnError  bool `json:"continue_on_error"`
	ValidateOnly     bool `json:"validate_only"`
	ProgressInterval int  `json:"progress_interval_seconds"`
}

// DatabaseConfig controls database connections and behavior
type DatabaseConfig struct {
	MaxConnections    int    `json:"max_connections"`
	IdleConnections   int    `json:"idle_connections"`
	ConnectionTimeout int    `json:"connection_timeout_seconds"`
	QueryTimeout      int    `json:"query_timeout_seconds"`
	Schema           string `json:"schema"`
	CreateTables     bool   `json:"create_tables"`
	DropFirst        bool   `json:"drop_first"`
	UseTransactions  bool   `json:"use_transactions"`
	BatchInserts     bool   `json:"batch_inserts"`
}

// BusinessRules contains oil & gas industry specific rules
type BusinessRules struct {
	ValidGrades      []GradeInfo      `json:"valid_grades"`
	ValidSizes       []SizeInfo       `json:"valid_sizes"`
	ValidConnections []ConnectionInfo `json:"valid_connections"`
	WorkOrderFormat  WorkOrderFormat  `json:"work_order_format"`
	CustomerRules    CustomerRules    `json:"customer_rules"`
	InventoryRules   InventoryRules   `json:"inventory_rules"`
}

type GradeInfo struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	MinYield    *int    `json:"min_yield,omitempty"`
	MaxYield    *int    `json:"max_yield,omitempty"`
	Deprecated  bool    `json:"deprecated"`
}

type SizeInfo struct {
	Size        string  `json:"size"`
	Description string  `json:"description"`
	OuterDiam   float64 `json:"outer_diameter"`
	InnerDiam   float64 `json:"inner_diameter"`
	WeightRange string  `json:"weight_range"`
}

type ConnectionInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Standard    string `json:"standard"`
	Deprecated  bool   `json:"deprecated"`
}

type WorkOrderFormat struct {
	Pattern     string `json:"pattern"`
	PrefixRules string `json:"prefix_rules"`
	NumberStart int    `json:"number_start"`
	ZeroPad     int    `json:"zero_pad"`
}

type CustomerRules struct {
	RequiredFields     []string          `json:"required_fields"`
	NameNormalization string            `json:"name_normalization"`
	DuplicateHandling string            `json:"duplicate_handling"`
	AddressValidation bool              `json:"address_validation"`
	CustomValidations map[string]string `json:"custom_validations"`
}

type InventoryRules struct {
	LocationFormat     string            `json:"location_format"`
	SerialNumberFormat string            `json:"serial_number_format"`
	StatusValidation   []string          `json:"status_validation"`
	WeightValidation   WeightValidation  `json:"weight_validation"`
	QualityRules      QualityRules      `json:"quality_rules"`
}

type WeightValidation struct {
	MinWeight     float64 `json:"min_weight"`
	MaxWeight     float64 `json:"max_weight"`
	TolerancePerc float64 `json:"tolerance_percentage"`
}

type QualityRules struct {
	RequiredInspections []string          `json:"required_inspections"`
	CertificationRules map[string]string `json:"certification_rules"`
	RejectCriteria     []string          `json:"reject_criteria"`
}

// Load loads configuration from file with defaults
func Load(configPath string) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()

	// Load from file if it exists
	if configPath != "" {
		if err := config.LoadFromFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadFromFile loads configuration from JSON file
func (c *Config) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(c); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate processing options
	if c.ProcessingOptions.Workers < 1 || c.ProcessingOptions.Workers > 32 {
		return fmt.Errorf("workers must be between 1 and 32, got %d", c.ProcessingOptions.Workers)
	}

	if c.ProcessingOptions.BatchSize < 100 || c.ProcessingOptions.BatchSize > 50000 {
		return fmt.Errorf("batch size must be between 100 and 50000, got %d", c.ProcessingOptions.BatchSize)
	}

	// Validate database config
	if c.DatabaseConfig.MaxConnections < 1 || c.DatabaseConfig.MaxConnections > 100 {
		return fmt.Errorf("max connections must be between 1 and 100, got %d", c.DatabaseConfig.MaxConnections)
	}

	// Validate business rules
	if len(c.BusinessRules.ValidGrades) == 0 {
		return fmt.Errorf("at least one valid grade must be specified")
	}

	// Validate output settings
	if !c.OutputSettings.CSVOutput && !c.OutputSettings.SQLOutput && !c.OutputSettings.PostgreSQLDirect {
		return fmt.Errorf("at least one output format must be enabled")
	}

	return nil
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		OilGasMappings: make(map[string]ColumnMapping),
		CompanyTemplate: CompanyTemplate{
			CompanyName:     "Default Company",
			WorkOrderPrefix: "WO",
			CustomMappings:  make(map[string]string),
			SpecialRules:    []string{},
			ContactInfo: ContactInfo{
				PrimaryContact: "Unknown",
				Email:         "",
				Phone:         "",
				Address:       "",
			},
			BusinessSettings: BusinessSettings{
				FiscalYearStart: "01-01",
				DefaultCurrency: "USD",
				TimeZone:       "UTC",
				CustomFields:   make(map[string]string),
			},
		},
		ValidationRules: []ValidationRule{},
		OutputSettings: OutputSettings{
			CSVOutput:         true,
			SQLOutput:         true,
			PostgreSQLDirect:  false,
			ValidationReport:  true,
			BusinessReport:    true,
			PerformanceReport: false,
			OutputDir:        "output",
			FileNaming:       "timestamp",
		},
		ProcessingOptions: ProcessingOptions{
			Workers:           4,
			BatchSize:         1000,
			MemoryLimit:       4096,
			TimeoutMinutes:    60,
			DryRun:           false,
			ContinueOnError:  true,
			ValidateOnly:     false,
			ProgressInterval: 5,
		},
		DatabaseConfig: DatabaseConfig{
			MaxConnections:    10,
			IdleConnections:   2,
			ConnectionTimeout: 30,
			QueryTimeout:      300,
			Schema:           "store",
			CreateTables:     true,
			DropFirst:        false,
			UseTransactions:  true,
			BatchInserts:     true,
		},
		BusinessRules: BusinessRules{
			ValidGrades: []GradeInfo{
				{Code: "J55", Description: "Standard grade steel casing", Deprecated: false},
				{Code: "JZ55", Description: "Enhanced J55 grade", Deprecated: false},
				{Code: "L80", Description: "Higher strength grade", Deprecated: false},
				{Code: "N80", Description: "Medium strength grade", Deprecated: false},
				{Code: "P105", Description: "High performance grade", Deprecated: false},
				{Code: "P110", Description: "Premium performance grade", Deprecated: false},
				{Code: "Q125", Description: "Ultra-high strength grade", Deprecated: false},
				{Code: "C75", Description: "Carbon steel grade", Deprecated: false},
				{Code: "C95", Description: "Higher carbon steel grade", Deprecated: false},
				{Code: "T95", Description: "Tough grade for harsh environments", Deprecated: false},
			},
			ValidSizes: []SizeInfo{
				{Size: "4 1/2\"", Description: "4.5 inch diameter", OuterDiam: 4.5, InnerDiam: 3.958},
				{Size: "5\"", Description: "5 inch diameter", OuterDiam: 5.0, InnerDiam: 4.408},
				{Size: "5 1/2\"", Description: "5.5 inch diameter", OuterDiam: 5.5, InnerDiam: 4.892},
				{Size: "7\"", Description: "7 inch diameter", OuterDiam: 7.0, InnerDiam: 6.366},
				{Size: "8 5/8\"", Description: "8.625 inch diameter", OuterDiam: 8.625, InnerDiam: 7.921},
				{Size: "9 5/8\"", Description: "9.625 inch diameter", OuterDiam: 9.625, InnerDiam: 8.921},
				{Size: "10 3/4\"", Description: "10.75 inch diameter", OuterDiam: 10.75, InnerDiam: 9.894},
				{Size: "13 3/8\"", Description: "13.375 inch diameter", OuterDiam: 13.375, InnerDiam: 12.459},
				{Size: "16\"", Description: "16 inch diameter", OuterDiam: 16.0, InnerDiam: 15.0},
				{Size: "18 5/8\"", Description: "18.625 inch diameter", OuterDiam: 18.625, InnerDiam: 17.755},
				{Size: "20\"", Description: "20 inch diameter", OuterDiam: 20.0, InnerDiam: 19.0},
				{Size: "24\"", Description: "24 inch diameter", OuterDiam: 24.0, InnerDiam: 23.0},
				{Size: "30\"", Description: "30 inch diameter", OuterDiam: 30.0, InnerDiam: 29.0},
			},
			ValidConnections: []ConnectionInfo{
				{Type: "BTC", Description: "Buttress Thread Casing", Standard: "API", Deprecated: false},
				{Type: "LTC", Description: "Long Thread Casing", Standard: "API", Deprecated: false},
				{Type: "STC", Description: "Short Thread Casing", Standard: "API", Deprecated: false},
				{Type: "VAM TOP", Description: "VAM Top Premium", Standard: "Vallourec", Deprecated: false},
				{Type: "VAM ACE", Description: "VAM Ace Premium", Standard: "Vallourec", Deprecated: false},
				{Type: "NEW VAM", Description: "New VAM Premium", Standard: "Vallourec", Deprecated: false},
				{Type: "TSH BLUE", Description: "TSH Blue Premium", Standard: "Tenaris", Deprecated: false},
				{Type: "PREMIUM", Description: "Generic Premium", Standard: "Various", Deprecated: false},
				{Type: "BUTTRESS", Description: "Buttress Thread", Standard: "API", Deprecated: false},
			},
			WorkOrderFormat: WorkOrderFormat{
				Pattern:     "^[A-Z]{1,3}-\\d{6}$",
				PrefixRules: "Company initials or location code",
				NumberStart: 1,
				ZeroPad:     6,
			},
			CustomerRules: CustomerRules{
				RequiredFields:     []string{"customer_name", "billing_address", "contact"},
				NameNormalization: "uppercase_first_letter",
				DuplicateHandling: "merge",
				AddressValidation: false,
				CustomValidations: make(map[string]string),
			},
			InventoryRules: InventoryRules{
				LocationFormat:     "^[A-Z]+-[A-Z0-9]+$",
				SerialNumberFormat: "^[A-Z0-9]{8,}$",
				StatusValidation:   []string{"received", "processing", "ready", "shipped", "returned"},
				WeightValidation: WeightValidation{
					MinWeight:     0.1,
					MaxWeight:     50000.0,
					TolerancePerc: 5.0,
				},
				QualityRules: QualityRules{
					RequiredInspections: []string{"visual", "dimensional"},
					CertificationRules:  make(map[string]string),
					RejectCriteria:     []string{"cracked", "damaged", "out_of_spec"},
				},
			},
		},
	}
}

// GetColumnMapping returns the mapping for a specific column
func (c *Config) GetColumnMapping(columnName string) (ColumnMapping, bool) {
	mapping, exists := c.OilGasMappings[columnName]
	return mapping, exists
}

// IsValidGrade checks if a grade is valid
func (c *Config) IsValidGrade(grade string) bool {
	for _, validGrade := range c.BusinessRules.ValidGrades {
		if validGrade.Code == grade && !validGrade.Deprecated {
			return true
		}
	}
	return false
}

// IsValidSize checks if a size is valid
func (c *Config) IsValidSize(size string) bool {
	for _, validSize := range c.BusinessRules.ValidSizes {
		if validSize.Size == size {
			return true
		}
	}
	return false
}

// IsValidConnection checks if a connection type is valid
func (c *Config) IsValidConnection(connection string) bool {
	for _, validConn := range c.BusinessRules.ValidConnections {
		if validConn.Type == connection && !validConn.Deprecated {
			return true
		}
	}
	return false
}

// GetValidationRulesForField returns validation rules for a specific field
func (c *Config) GetValidationRulesForField(fieldName string) []ValidationRule {
	var rules []ValidationRule
	for _, rule := range c.ValidationRules {
		if rule.FieldName == fieldName {
			rules = append(rules, rule)
		}
	}
	return rules
}

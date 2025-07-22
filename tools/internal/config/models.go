package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the complete configuration for MDB processing
type Config struct {
	OilGasMappings   map[string]ColumnMapping `json:"oil_gas_mappings"`
	ProcessingOptions ProcessingOptions        `json:"processing_options"`
	DatabaseConfig   DatabaseConfig           `json:"database_config"`
	ValidationRules  ValidationRules          `json:"validation_rules"`
}

// ColumnMapping defines how to transform a column
type ColumnMapping struct {
	PostgreSQLName  string   `json:"postgresql_name"`
	DataType        string   `json:"data_type"`
	BusinessRules   []string `json:"business_rules"`
	Required        bool     `json:"required"`
	DefaultValue    string   `json:"default_value,omitempty"`
	ValidationRules []string `json:"validation_rules,omitempty"`
}

// ProcessingOptions controls processing behavior
type ProcessingOptions struct {
	Workers       int `json:"workers"`
	BatchSize     int `json:"batch_size"`
	MemoryLimitMB int `json:"memory_limit_mb"`
	TimeoutMinutes int `json:"timeout_minutes"`
}

// DatabaseConfig for PostgreSQL connections
type DatabaseConfig struct {
	MaxConnections     int `json:"max_connections"`
	IdleConnections    int `json:"idle_connections"`
	QueryTimeoutSeconds int `json:"query_timeout_seconds"`
}

// ValidationRules for data quality
type ValidationRules struct {
	OilGasGrades     []string `json:"oil_gas_grades"`
	PipeSizes        []string `json:"pipe_sizes"`
	ConnectionTypes  []string `json:"connection_types"`
	RequiredFields   []string `json:"required_fields"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filename string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// getDefaultConfig returns a default configuration for oil & gas processing
func getDefaultConfig() *Config {
	return &Config{
		OilGasMappings: map[string]ColumnMapping{
			"customer": {
				PostgreSQLName: "customer",
				DataType:       "varchar(255)",
				BusinessRules:  []string{"normalize_customer_name"},
				Required:       true,
			},
			"work_order": {
				PostgreSQLName: "work_order",
				DataType:       "varchar(50)",
				BusinessRules:  []string{"validate_work_order_format"},
				Required:       true,
			},
		},
		ProcessingOptions: ProcessingOptions{
			Workers:        4,
			BatchSize:      1000,
			MemoryLimitMB:  2048,
			TimeoutMinutes: 60,
		},
		DatabaseConfig: DatabaseConfig{
			MaxConnections:      25,
			IdleConnections:     10,
			QueryTimeoutSeconds: 300,
		},
		ValidationRules: ValidationRules{
			OilGasGrades:    []string{"J55", "L80", "N80", "P110"},
			PipeSizes:       []string{"4 1/2\"", "5\"", "5 1/2\"", "7\""},
			ConnectionTypes: []string{"BTC", "LTC", "STC", "VAM TOP"},
			RequiredFields:  []string{"work_order", "customer"},
		},
	}
}

// validateConfig ensures the configuration is valid
func validateConfig(config *Config) error {
	if config.ProcessingOptions.Workers < 1 {
		return fmt.Errorf("workers must be at least 1")
	}
	return nil
}

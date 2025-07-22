package processor

import (
	"time"
	"oilgas-tools/internal/config"
)

// Config holds the processor configuration
type Config struct {
	Company       string
	Workers       int
	BatchSize     int
	OutputDir     string
	Verbose       bool
	MappingConfig *config.Config
}

// ProcessingResult contains the results of processing
type ProcessingResult struct {
	RecordsProcessed int
	ValidRecords     int
	Warnings         int
	Errors           int
	ValidationIssues []ValidationIssue
	OutputFiles      []string
	Duration         time.Duration
}

// ValidationIssue represents a data validation issue
type ValidationIssue struct {
	Type        string
	Description string
	Row         int
	Column      string
	Value       string
	Severity    string
}

// TableInfo contains metadata about a table being processed
type TableInfo struct {
	Name       string
	RowCount   int
	Columns    []string
	DataSample [][]string
	FileSize   int64
}

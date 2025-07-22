// cmd/mdb_processor.go
// Main entry point for MDB processor - handles CLI and orchestration
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"tools/internal/config"
	"tools/internal/processor"
	"tools/internal/reporting"
)

const (
	Version = "1.0.0"
	AppName = "MDB Processor"
)

type CLIOptions struct {
	MDBFile      string
	CompanyName  string
	ConfigFile   string
	DBConnection string
	OutputDir    string
	Format       string
	Verbose      bool
	DryRun       bool
	Workers      int
	BatchSize    int
	ShowVersion  bool
	ShowHelp     bool
}

func main() {
	opts := parseFlags()

	if opts.ShowVersion {
		fmt.Printf("%s v%s\n", AppName, Version)
		os.Exit(0)
	}

	if opts.ShowHelp || opts.MDBFile == "" || opts.CompanyName == "" {
		showUsage()
		os.Exit(0)
	}

	// Setup logging
	setupLogging(opts.Verbose)

	log.Printf("Starting %s v%s", AppName, Version)
	log.Printf("Processing: %s for company: %s", opts.MDBFile, opts.CompanyName)

	// Validate inputs
	if err := validateInputs(opts); err != nil {
		log.Fatalf("Input validation failed: %v", err)
	}

	// Load configuration
	cfg, err := config.Load(opts.ConfigFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Apply CLI overrides to config
	applyCliOverrides(cfg, opts)

	// Create processor
	proc, err := processor.New(cfg, opts.DBConnection)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}
	defer proc.Close()

	// Setup progress monitoring
	progressChan := make(chan processor.ProgressUpdate, 100)
	errorChan := make(chan processor.ProcessingError, 1000)

	go monitorProgress(progressChan)
	go monitorErrors(errorChan)

	// Process MDB file
	startTime := time.Now()
	job, err := proc.ProcessMDB(processor.ProcessingRequest{
		SourceFile:   opts.MDBFile,
		CompanyName:  opts.CompanyName,
		OutputDir:    opts.OutputDir,
		DryRun:       opts.DryRun,
		ProgressChan: progressChan,
		ErrorChan:    errorChan,
	})

	duration := time.Since(startTime)

	// Close channels
	close(progressChan)
	close(errorChan)

	if err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	// Generate final report
	if err := generateFinalReport(job, duration, opts.OutputDir); err != nil {
		log.Printf("Warning: Failed to generate final report: %v", err)
	}

	// Print summary
	printSummary(job, duration)

	if job.HasErrors() {
		os.Exit(1)
	}

	log.Println("Processing completed successfully")
}

func parseFlags() CLIOptions {
	var opts CLIOptions

	flag.StringVar(&opts.MDBFile, "file", "", "Path to MDB file (required)")
	flag.StringVar(&opts.CompanyName, "company", "", "Company name (required)")
	flag.StringVar(&opts.ConfigFile, "config", "config/oil_gas_mappings.json", "Configuration file path")
	flag.StringVar(&opts.DBConnection, "db", "", "Database connection string (uses DATABASE_URL env if not provided)")
	flag.StringVar(&opts.OutputDir, "output", "output", "Output directory")
	flag.StringVar(&opts.Format, "format", "all", "Output format: csv,sql,direct,all")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Verbose logging")
	flag.BoolVar(&opts.DryRun, "dry-run", false, "Validate only, don't write outputs")
	flag.IntVar(&opts.Workers, "workers", 4, "Number of worker goroutines")
	flag.IntVar(&opts.BatchSize, "batch-size", 1000, "Batch size for processing")
	flag.BoolVar(&opts.ShowVersion, "version", false, "Show version")
	flag.BoolVar(&opts.ShowHelp, "help", false, "Show help")

	flag.Parse()

	// Set defaults
	if opts.DBConnection == "" {
		opts.DBConnection = os.Getenv("DATABASE_URL")
	}

	return opts
}

func showUsage() {
	fmt.Printf(`%s v%s - Oil & Gas Industry MDB to PostgreSQL Converter

USAGE:
    %s -file <mdb_file> -company <company_name> [options]

EXAMPLES:
    # Basic conversion
    %s -file company.mdb -company "Acme Oil"
    
    # With custom config and verbose output
    %s -file company.mdb -company "Acme Oil" -config custom.json -verbose
    
    # Dry run validation only
    %s -file company.mdb -company "Acme Oil" -dry-run
    
    # High performance with more workers
    %s -file large.mdb -company "Big Oil Corp" -workers 8 -batch-size 5000

OPTIONS:
`, AppName, Version, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])

	flag.PrintDefaults()

	fmt.Printf(`
SUPPORTED FORMATS:
    csv     - Normalized CSV files
    sql     - PostgreSQL CREATE TABLE and INSERT statements  
    direct  - Direct database insertion
    all     - All formats (default)

CONFIGURATION:
    The processor uses industry-standard oil & gas mappings and validation rules.
    Customize behavior via the configuration file.

ENVIRONMENT VARIABLES:
    DATABASE_URL    - PostgreSQL connection string
    LOG_LEVEL       - Logging level (debug, info, warn, error)
    
For more information, see: https://github.com/dcotelessa/oilgas-project/tools
`)
}

func validateInputs(opts CLIOptions) error {
	// Check if MDB file exists
	if _, err := os.Stat(opts.MDBFile); os.IsNotExist(err) {
		return fmt.Errorf("MDB file does not exist: %s", opts.MDBFile)
	}

	// Validate company name
	if len(opts.CompanyName) < 2 {
		return fmt.Errorf("company name must be at least 2 characters")
	}

	// Check config file exists
	if _, err := os.Stat(opts.ConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", opts.ConfigFile)
	}

	// Validate output directory
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	// Validate format
	validFormats := map[string]bool{
		"csv":    true,
		"sql":    true,
		"direct": true,
		"all":    true,
	}
	if !validFormats[opts.Format] {
		return fmt.Errorf("invalid format: %s", opts.Format)
	}

	// Validate workers and batch size
	if opts.Workers < 1 || opts.Workers > 16 {
		return fmt.Errorf("workers must be between 1 and 16")
	}

	if opts.BatchSize < 100 || opts.BatchSize > 10000 {
		return fmt.Errorf("batch size must be between 100 and 10000")
	}

	return nil
}

func setupLogging(verbose bool) {
	logLevel := os.Getenv("LOG_LEVEL")
	if verbose {
		logLevel = "debug"
	}
	if logLevel == "" {
		logLevel = "info"
	}

	// Configure log format
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// TODO: Implement proper structured logging with levels
	// For now, using standard log package
	if verbose {
		log.Println("Verbose logging enabled")
	}
}

func applyCliOverrides(cfg *config.Config, opts CLIOptions) {
	// Apply CLI overrides to configuration
	cfg.ProcessingOptions.Workers = opts.Workers
	cfg.ProcessingOptions.BatchSize = opts.BatchSize
	cfg.ProcessingOptions.DryRun = opts.DryRun

	// Set output formats based on CLI flag
	switch opts.Format {
	case "csv":
		cfg.OutputSettings = config.OutputSettings{CSVOutput: true}
	case "sql":
		cfg.OutputSettings = config.OutputSettings{SQLOutput: true}
	case "direct":
		cfg.OutputSettings = config.OutputSettings{PostgreSQLDirect: true}
	case "all":
		cfg.OutputSettings = config.OutputSettings{
			CSVOutput:        true,
			SQLOutput:        true,
			PostgreSQLDirect: true,
			ValidationReport: true,
		}
	}
}

func monitorProgress(progressChan <-chan processor.ProgressUpdate) {
	for update := range progressChan {
		log.Printf("Progress: %s - %.1f%% (%d/%d) - %s",
			update.Table, update.Percentage, update.RecordsProcessed, update.TotalRecords, update.Message)
	}
}

func monitorErrors(errorChan <-chan processor.ProcessingError) {
	for err := range errorChan {
		log.Printf("Error: Record %s, Field %s - %s: %s",
			err.RecordID, err.FieldName, err.ErrorType, err.Description)
	}
}

func generateFinalReport(job *processor.ConversionJob, duration time.Duration, outputDir string) error {
	reporter := reporting.New()
	
	report := reporting.FinalReport{
		JobID:            job.ID,
		CompanyName:      job.CompanyName,
		StartTime:        job.StartTime,
		EndTime:          time.Now(),
		Duration:         duration,
		RecordsTotal:     job.RecordsTotal,
		RecordsValid:     job.RecordsValid,
		RecordsInvalid:   job.RecordsInvalid,
		TablesProcessed:  job.TablesProcessed,
		OutputFiles:      job.OutputFiles,
		Errors:           job.Errors,
		Status:           job.Status,
		PerformanceStats: job.PerformanceStats,
	}

	reportPath := filepath.Join(outputDir, "reports", fmt.Sprintf("final_report_%s.json", job.ID))
	return reporter.GenerateFinalReport(report, reportPath)
}

func printSummary(job *processor.ConversionJob, duration time.Duration) {
	fmt.Printf("\n" + "="*60 + "\n")
	fmt.Printf("🎯 CONVERSION COMPLETE\n")
	fmt.Printf("="*60 + "\n")
	fmt.Printf("Job ID:           %s\n", job.ID)
	fmt.Printf("Company:          %s\n", job.CompanyName)
	fmt.Printf("Status:           %s\n", formatStatus(job.Status))
	fmt.Printf("Duration:         %v\n", duration.Round(time.Second))
	fmt.Printf("Tables Processed: %d\n", job.TablesProcessed)
	fmt.Printf("Records Total:    %d\n", job.RecordsTotal)
	fmt.Printf("Records Valid:    %d\n", job.RecordsValid)
	fmt.Printf("Records Invalid:  %d\n", job.RecordsInvalid)
	fmt.Printf("Success Rate:     %.2f%%\n", job.SuccessRate())

	if len(job.Errors) > 0 {
		fmt.Printf("\n⚠️  ERRORS ENCOUNTERED: %d\n", len(job.Errors))
		
		// Group errors by type
		errorCounts := make(map[string]int)
		for _, err := range job.Errors {
			errorCounts[err.ErrorType]++
		}
		
		for errorType, count := range errorCounts {
			fmt.Printf("  - %s: %d\n", errorType, count)
		}
		
		// Show first few critical errors
		criticalErrors := 0
		for _, err := range job.Errors {
			if err.ErrorType == "critical" && criticalErrors < 3 {
				fmt.Printf("  ❌ %s\n", err.Description)
				criticalErrors++
			}
		}
	}

	if len(job.OutputFiles) > 0 {
		fmt.Printf("\n📁 OUTPUT FILES (%d):\n", len(job.OutputFiles))
		for _, file := range job.OutputFiles {
			if fileInfo, err := os.Stat(file); err == nil {
				fmt.Printf("  ✅ %s (%s)\n", file, formatFileSize(fileInfo.Size()))
			} else {
				fmt.Printf("  ❌ %s (missing)\n", file)
			}
		}
	}

	if job.PerformanceStats != nil {
		fmt.Printf("\n📊 PERFORMANCE:\n")
		fmt.Printf("  Records/sec:  %.0f\n", job.PerformanceStats.RecordsPerSecond)
		fmt.Printf("  Peak Memory:  %s\n", formatFileSize(job.PerformanceStats.PeakMemoryUsage))
		fmt.Printf("  Avg CPU:      %.1f%%\n", job.PerformanceStats.AvgCPUUsage)
	}

	fmt.Printf("\n💡 NEXT STEPS:\n")
	if job.Status == "completed" {
		fmt.Printf("  1. Review validation report for data quality\n")
		fmt.Printf("  2. Import SQL files or verify direct database insertion\n")
		fmt.Printf("  3. Run business logic validation tests\n")
		fmt.Printf("  4. Update application configuration for new schema\n")
	} else {
		fmt.Printf("  1. Review error log for issues\n")
		fmt.Printf("  2. Fix data quality problems\n")
		fmt.Printf("  3. Re-run conversion with corrected data\n")
	}
	
	fmt.Printf("="*60 + "\n")
}

func formatStatus(status string) string {
	switch status {
	case "completed":
		return "✅ " + status
	case "failed":
		return "❌ " + status
	case "partial":
		return "⚠️  " + status
	default:
		return "🔄 " + status
	}
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// tools/cmd/mdb_processor.go
// Implements multi-tenant support, lowercase table mapping, and sequence handling
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	
	"oilgas-tools/internal/config"
	"oilgas-tools/internal/processor"
	"oilgas-tools/internal/reporting"
	"oilgas-tools/internal/database"
)

const version = "1.1.0" // Updated for multi-tenant support

type CLIConfig struct {
	InputFile     string
	Company       string
	TenantID      string
	ConfigFile    string
	OutputDir     string
	DatabaseURL   string
	Workers       int
	BatchSize     int
	Verbose       bool
	ShowHelp      bool
	ShowVersion   bool
	DirectImport  bool
	ValidateOnly  bool
}

func main() {
	cfg := parseCLIArgs()
	
	if cfg.ShowHelp {
		showHelp()
		return
	}
	
	if cfg.ShowVersion {
		fmt.Printf("MDB Processor v%s (Multi-Tenant)\n", version)
		return
	}
	
	if err := validateArgs(cfg); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	
	// Load configuration
	config, err := loadConfig(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	
	if err := processWithTenantSupport(cfg, config); err != nil {
		log.Fatalf("❌ Processing failed: %v", err)
	}
}

func parseCLIArgs() *CLIConfig {
	cfg := &CLIConfig{}
	
	flag.StringVar(&cfg.InputFile, "input", "", "Input MDB file path")
	flag.StringVar(&cfg.Company, "company", "", "Company name")
	flag.StringVar(&cfg.TenantID, "tenant", "location_longbeach", "Tenant ID (e.g., location_longbeach)")
	flag.StringVar(&cfg.ConfigFile, "config", "config.json", "Configuration file path")
	flag.StringVar(&cfg.OutputDir, "output", "tools/output", "Output directory")
	flag.StringVar(&cfg.DatabaseURL, "db", "", "Database connection URL")
	flag.IntVar(&cfg.Workers, "workers", 4, "Number of parallel workers")
	flag.IntVar(&cfg.BatchSize, "batch", 1000, "Batch size for processing")
	flag.BoolVar(&cfg.DirectImport, "direct", false, "Import directly to database (skip CSV)")
	flag.BoolVar(&cfg.ValidateOnly, "validate", false, "Validate only, don't process")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&cfg.ShowHelp, "help", false, "Show help")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Show version")
	
	flag.Parse()
	return cfg
}

func validateArgs(cfg *CLIConfig) error {
	if cfg.InputFile == "" {
		return fmt.Errorf("input file is required")
	}
	
	if _, err := os.Stat(cfg.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", cfg.InputFile)
	}
	
	if cfg.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	
	if cfg.DirectImport && cfg.DatabaseURL == "" {
		return fmt.Errorf("database URL required for direct import")
	}
	
	return nil
}

func loadConfig(cliCfg *CLIConfig) (*config.Config, error) {
	cfg, err := config.Load(cliCfg.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	if cliCfg.Company != "" {
		cfg.Company = cliCfg.Company
	}
	if cliCfg.TenantID != "" {
		cfg.TenantID = cliCfg.TenantID
	}
	if cliCfg.DatabaseURL != "" {
		cfg.DatabaseURL = cliCfg.DatabaseURL
	}
	
	if cfg.DatabaseURL != "" && cfg.TenantID != "" {
		cfg.DatabaseURL = updateDatabaseForTenant(cfg.DatabaseURL, cfg.TenantID)
	}
	
	return cfg, nil
}

func updateDatabaseForTenant(dbURL, tenantID string) string {
	if tenantID == "" {
		return dbURL
	}
	
	// Simple implementation - TODO: you might want to use a proper URL parser
	if contains := strings.Contains(dbURL, "dbname="); contains {
		// Replace existing dbname
		re := regexp.MustCompile(`dbname=\w+`)
		return re.ReplaceAllString(dbURL, fmt.Sprintf("dbname=oilgas_%s", tenantID))
	} else {
		// Add dbname
		return fmt.Sprintf("%s dbname=oilgas_%s", dbURL, tenantID)
	}
}

func processWithTenantSupport(cliCfg *CLIConfig, cfg *config.Config) error {
	ctx := context.Background()
	
	fmt.Printf("🚀 Processing MDB file for tenant: %s\n", cfg.TenantID)
	fmt.Printf("📂 Input: %s\n", cliCfg.InputFile)
	fmt.Printf("📁 Output: %s\n", cliCfg.OutputDir)
	
	if cliCfg.DirectImport {
		fmt.Printf("🗄️ Database: %s\n", maskDatabaseURL(cfg.DatabaseURL))
	}
	
	proc, err := processor.NewWithTenantSupport(cfg)
	if err != nil {
		return fmt.Errorf("failed to create processor: %w", err)
	}
	defer proc.Close()
	
	if cliCfg.ValidateOnly {
		return runValidationOnly(ctx, proc, cliCfg.InputFile)
	}
	
	result, err := proc.ProcessMDBWithTenant(ctx, &processor.ProcessRequest{
		InputFile:    cliCfg.InputFile,
		OutputDir:    cliCfg.OutputDir,
		TenantID:     cfg.TenantID,
		DirectImport: cliCfg.DirectImport,
		Workers:      cliCfg.Workers,
		BatchSize:    cliCfg.BatchSize,
		Verbose:      cliCfg.Verbose,
	})
	
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}
	
	return generateTenantReport(result, cfg.TenantID, cliCfg.OutputDir)
}

func runValidationOnly(ctx context.Context, proc *processor.Processor, inputFile string) error {
	fmt.Println("🔍 Running validation only...")
	
	validation, err := proc.ValidateMDBFile(ctx, inputFile)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Print validation results
	fmt.Printf("✅ Validation Results:\n")
	fmt.Printf("  Tables found: %d\n", validation.TablesFound)
	fmt.Printf("  Total records: %d\n", validation.TotalRecords)
	fmt.Printf("  Issues found: %d\n", len(validation.Issues))
	
	if len(validation.Issues) > 0 {
		fmt.Printf("\n⚠️ Issues:\n")
		for _, issue := range validation.Issues {
			fmt.Printf("  • %s: %s\n", issue.Table, issue.Description)
		}
	}
	
	return nil
}

func generateTenantReport(result *processor.ProcessResult, tenantID, outputDir string) error {
	reporter := reporting.New()
	
	report := &reporting.TenantReport{
		TenantID:        tenantID,
		ProcessedAt:     time.Now(),
		TablesProcessed: result.TablesProcessed,
		RecordsImported: result.RecordsImported,
		RecordsSkipped:  result.RecordsSkipped,
		Errors:          result.Errors,
		Performance:     result.Performance,
	}
	
	// Generate report in multiple formats
	reportPath := fmt.Sprintf("%s/processing_report_%s.json", outputDir, tenantID)
	if err := reporter.GenerateJSON(report, reportPath); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	
	// Console summary
	fmt.Printf("\n📊 Processing Summary - Tenant: %s\n", tenantID)
	fmt.Printf("================================\n")
	fmt.Printf("Tables processed: %d\n", result.TablesProcessed)
	fmt.Printf("Records imported: %d\n", result.RecordsImported)
	fmt.Printf("Records skipped: %d\n", result.RecordsSkipped)
	fmt.Printf("Processing time: %v\n", result.Performance.Duration)
	
	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️ Errors encountered: %d\n", len(result.Errors))
		for i, err := range result.Errors {
			if i >= 5 {
				fmt.Printf("  ... and %d more errors (see full report)\n", len(result.Errors)-5)
				break
			}
			fmt.Printf("  • %s\n", err)
		}
	}
	
	fmt.Printf("\n📄 Full report: %s\n", reportPath)
	return nil
}

func maskDatabaseURL(dbURL string) string {
	// Mask password in database URL for logging
	re := regexp.MustCompile(`password=\w+`)
	return re.ReplaceAllString(dbURL, "password=***")
}

func showHelp() {
	fmt.Printf("MDB Processor v%s - Multi-Tenant Oil & Gas Data Processing\n\n", version)
	fmt.Println("USAGE:")
	fmt.Println("  mdb_processor [OPTIONS]")
	fmt.Println("")
	fmt.Println("REQUIRED:")
	fmt.Println("  -input <file>     Input MDB file path")
	fmt.Println("")
	fmt.Println("OPTIONS:")
	fmt.Println("  -tenant <id>      Tenant ID (default: location_longbeach)")
	fmt.Println("  -company <name>   Company name")
	fmt.Println("  -config <file>    Configuration file (default: config.json)")
	fmt.Println("  -output <dir>     Output directory (default: tools/output)")
	fmt.Println("  -db <url>         Database connection URL")
	fmt.Println("  -workers <n>      Number of parallel workers (default: 4)")
	fmt.Println("  -batch <n>        Batch size for processing (default: 1000)")
	fmt.Println("  -direct           Import directly to database (skip CSV)")
	fmt.Println("  -validate         Validate only, don't process")
	fmt.Println("  -verbose          Verbose output")
	fmt.Println("  -version          Show version")
	fmt.Println("  -help             Show this help")
	fmt.Println("")
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Process MDB file for Long Beach location")
	fmt.Println("  mdb_processor -input data.mdb -tenant location_longbeach")
	fmt.Println("")
	fmt.Println("  # Direct import to database")
	fmt.Println("  mdb_processor -input data.mdb -tenant location_lasvegas \\")
	fmt.Println("    -db 'host=localhost user=postgres password=secret' -direct")
	fmt.Println("")
	fmt.Println("  # Validate MDB file without processing")
	fmt.Println("  mdb_processor -input data.mdb -validate")
	fmt.Println("")
	fmt.Println("TENANT IDs:")
	fmt.Println("  location_longbeach   Long Beach operations")
	fmt.Println("  location_lasvegas    Las Vegas operations") 
	fmt.Println("  location_colorado    Colorado operations")
}

func OutputTenantCSV(data [][]string, headers []string, tenant string, tableName string) {
    outputFile := fmt.Sprintf("../csv/%s/%s.csv", tenant, strings.ToLower(tableName))
    
    normalizedHeaders := make([]string, len(headers))
    for i, header := range headers {
        normalizedHeaders[i] = NormalizeColumnName(header)
    }
    
    // Write tenant-ready CSV
    file, _ := os.Create(outputFile)
    writer := csv.NewWriter(file)
    writer.Write(normalizedHeaders)
    writer.WriteAll(data)
    writer.Flush()
    file.Close()
    
    fmt.Printf("✅ Tenant CSV ready: %s\n", outputFile)
}

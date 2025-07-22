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
)

const version = "1.0.0"

type CLIConfig struct {
	InputFile    string
	Company      string
	ConfigFile   string
	OutputDir    string
	Workers      int
	BatchSize    int
	Verbose      bool
	ShowHelp     bool
	ShowVersion  bool
}

func main() {
	cfg := parseCLIArgs()

	if cfg.ShowHelp {
		showHelp()
		return
	}

	if cfg.ShowVersion {
		fmt.Printf("MDB Processor v%s\n", version)
		return
	}

	if err := validateArgs(cfg); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	if err := runProcessor(cfg); err != nil {
		log.Fatalf("❌ Processing failed: %v", err)
	}

	fmt.Println("✅ Processing completed successfully!")
}

func parseCLIArgs() *CLIConfig {
	cfg := &CLIConfig{}

	flag.StringVar(&cfg.InputFile, "file", "", "Input CSV file to process")
	flag.StringVar(&cfg.Company, "company", "", "Company name for this conversion")
	flag.StringVar(&cfg.ConfigFile, "config", "config/oil_gas_mappings.json", "Configuration file path")
	flag.StringVar(&cfg.OutputDir, "output", "output", "Output directory")
	flag.IntVar(&cfg.Workers, "workers", 4, "Number of worker threads")
	flag.IntVar(&cfg.BatchSize, "batch-size", 1000, "Batch size")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&cfg.ShowHelp, "help", false, "Show help")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Show version")

	flag.Parse()
	return cfg
}

func validateArgs(cfg *CLIConfig) error {
	if cfg.InputFile == "" {
		return fmt.Errorf("input file is required (-file)")
	}

	if _, err := os.Stat(cfg.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", cfg.InputFile)
	}

	if cfg.Company == "" {
		return fmt.Errorf("company name is required (-company)")
	}

	return nil
}

func runProcessor(cfg *CLIConfig) error {
	if cfg.Verbose {
		fmt.Printf("🚀 MDB Processor v%s\n", version)
		fmt.Printf("📁 Input: %s\n", cfg.InputFile)
		fmt.Printf("🏢 Company: %s\n", cfg.Company)
	}

	// Load configuration
	mappingConfig, err := config.LoadConfig(cfg.ConfigFile)
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	// Create processor config
	processorConfig := &processor.Config{
		Company:       cfg.Company,
		Workers:       cfg.Workers,
		BatchSize:     cfg.BatchSize,
		OutputDir:     cfg.OutputDir,
		Verbose:       cfg.Verbose,
		MappingConfig: mappingConfig,
	}

	// Create processor
	proc := processor.New(processorConfig)

	// Process file
	ctx := context.Background()
	start := time.Now()

	result, err := proc.ProcessFile(ctx, cfg.InputFile)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	duration := time.Since(start)

	// Show results
	if cfg.Verbose {
		fmt.Printf("\n📊 Results:\n")
		fmt.Printf("  ⏱️  Duration: %v\n", duration)
		fmt.Printf("  📝 Records: %d\n", result.RecordsProcessed)
		fmt.Printf("  ✅ Valid: %d\n", result.ValidRecords)
		fmt.Printf("  ❌ Errors: %d\n", result.Errors)
		fmt.Printf("  📁 Output: %s\n", cfg.OutputDir)
	}

	// Generate report
	reporter := reporting.New(cfg.OutputDir)
	if err := reporter.GenerateReport(cfg.Company, cfg.InputFile, result); err != nil {
		fmt.Printf("⚠️  Report generation warning: %v\n", err)
	}

	return nil
}

func showHelp() {
	fmt.Printf(`MDB Processor v%s - Oil & Gas Data Conversion

USAGE:
  mdb_processor -file <input.csv> -company <name> [options]

EXAMPLES:
  mdb_processor -file data.csv -company "Acme Oil"
  mdb_processor -file data.csv -company "Big Oil" -output results -verbose

OPTIONS:
  -file string       Input CSV file (required)
  -company string    Company name (required)
  -config string     Config file (default "config/oil_gas_mappings.json")
  -output string     Output directory (default "output")
  -workers int       Worker threads (default 4)
  -batch-size int    Batch size (default 1000)
  -verbose          Enable verbose output
  -help             Show help
  -version          Show version

OUTPUT:
  csv/     - Normalized CSV files
  sql/     - PostgreSQL scripts
  reports/ - Processing reports

FEATURES:
  • Grade normalization: J-55 → J55
  • Size conversion: 5.5 → 5 1/2"
  • Customer formatting: chevron corp → Chevron Corp
  • Connection mapping: BUTTRESS THREAD → BTC
`, version)
}

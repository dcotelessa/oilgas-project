// backend/cmd/tools/mdb-processor/main.go
// Integrates existing Phase 1 logic with new MDB processing capabilities
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Existing proven column mappings from Phase 1
var ColumnMapping = map[string]string{
	// Customer fields
	"custid":           "customer_id",
	"customerid":       "customer_id", 
	"custname":         "customer",
	"customername":     "customer",
	"customerpo":       "customer_po",
	"custpo":           "customer_po",
	// Work order fields
	"wkorder":          "work_order",
	"workorder":        "work_order",
	"wo":               "work_order",
	"rnumber":          "r_number",
	"rnum":             "r_number",
	// Date fields
	"datein":           "date_in",
	"dateout":          "date_out",
	"datereceived":     "date_received",
	"daterecvd":        "date_received",
	// Location fields
	"wellin":           "well_in",
	"wellout":          "well_out", 
	"leasein":          "lease_in",
	"leaseout":         "lease_out",
	// Address fields
	"billaddr":         "billing_address",
	"billcity":         "billing_city",
	"billstate":        "billing_state",
	"billzip":          "billing_zipcode",
	"billtoid":         "bill_to_id",
	// Contact fields
	"phoneno":          "phone",
	"phonenum":         "phone",
	"emailaddr":        "email",
	// Technical fields
	"wstring":          "w_string",
	"sizeid":           "size_id",
	"conntype":         "connection",
	"conn":             "connection",
	"locationcode":     "location",
	"loc":              "location",
	// Audit fields
	"orderedby":        "ordered_by",
	"enteredby":        "entered_by",
	"whenentered":      "when_entered",
	"when1":            "when_entered",
	"whenupdated":      "when_updated",
	"when2":            "when_updated",
	"updatedby":        "updated_by",
	"inspectedby":      "inspected_by",
	"inspecteddate":    "inspected_date",
	"inspected":        "inspected_date",
	"threadingdate":    "threading_date",
	"threading":        "threading_date",
	// Boolean fields
	"straightenreq":    "straighten_required",
	"straighten":       "straighten_required",
	"excessmat":        "excess_material",
	"excess":           "excess_material",
	"inproduction":     "in_production",
	"isdeleted":        "deleted",
	"createdat":        "created_at",
	"complete":         "complete",
}

// Enhanced MDB processor combining Phase 1 logic with mdb-tools
type MDBProcessor struct {
	mdbFile     string
	tenantSlug  string
	outputDir   string
	workingDir  string
}

func NewMDBProcessor(mdbFile, tenantSlug, outputDir string) *MDBProcessor {
	workingDir := filepath.Join(outputDir, fmt.Sprintf("working_%s", tenantSlug))
	return &MDBProcessor{
		mdbFile:    mdbFile,
		tenantSlug: tenantSlug,
		outputDir:  outputDir,
		workingDir: workingDir,
	}
}

// Phase 1 proven normalization logic
func NormalizeColumnName(colName string) string {
	if colName == "" {
		return colName
	}

	// Basic normalization
	normalized := strings.ToLower(strings.TrimSpace(colName))
	normalized = strings.ReplaceAll(normalized, `"`, "")
	normalized = strings.ReplaceAll(normalized, `'`, "")

	// Convert non-alphanumeric to underscores
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	normalized = reg.ReplaceAllString(normalized, "_")
	
	// Collapse multiple underscores
	reg = regexp.MustCompile(`_+`)
	normalized = reg.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")

	// Apply industry-specific mappings
	if mapped, exists := ColumnMapping[normalized]; exists {
		return mapped
	}

	// Convert common ID patterns
	if strings.HasSuffix(normalized, "id") && !strings.HasSuffix(normalized, "_id") {
		return normalized[:len(normalized)-2] + "_id"
	}

	return normalized
}

// Enhanced MDB analysis with tenant context
func (mp *MDBProcessor) AnalyzeMDB() error {
	fmt.Printf("🔍 Analyzing MDB file for tenant: %s\n", mp.tenantSlug)
	fmt.Printf("📁 Source: %s\n", mp.mdbFile)
	
	// Extract table list using mdb-tools
	cmd := exec.Command("mdb-tables", "-1", mp.mdbFile)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to extract tables: %w", err)
	}
	
	tableNames := strings.Fields(string(output))
	var cleanTables []string
	
	fmt.Printf("📋 Found %d tables:\n", len(tableNames))
	
	for _, table := range tableNames {
		// Skip system tables
		if !strings.HasPrefix(table, "MSys") && 
		   !strings.HasPrefix(table, "~") && 
		   len(table) > 0 {
			cleanTables = append(cleanTables, table)
			
			// Analyze table structure
			err := mp.analyzeTable(table)
			if err != nil {
				fmt.Printf("  ❌ %s: Error - %v\n", table, err)
			}
		}
	}
	
	fmt.Printf("\n📊 Analysis Summary:\n")
	fmt.Printf("  Tables to process: %d\n", len(cleanTables))
	fmt.Printf("  Tenant: %s\n", mp.tenantSlug)
	fmt.Printf("  Ready for extraction: ✅\n")
	
	return nil
}

func (mp *MDBProcessor) analyzeTable(tableName string) error {
	// Export sample data for analysis
	cmd := exec.Command("mdb-export", mp.mdbFile, tableName)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	// Parse first few rows
	reader := csv.NewReader(strings.NewReader(string(output)))
	reader.FieldsPerRecord = -1
	
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	
	// Count data rows
	rowCount := 0
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err == nil {
			rowCount++
		}
	}
	
	// Analyze column mappings
	mappings := make([]string, 0)
	for _, header := range headers {
		normalized := NormalizeColumnName(header)
		if strings.ToLower(header) != normalized {
			mappings = append(mappings, fmt.Sprintf("%s→%s", header, normalized))
		}
	}
	
	fmt.Printf("  ✅ %s: %d columns, %d rows", tableName, len(headers), rowCount)
	if len(mappings) > 0 {
		fmt.Printf(" (%d mapped)", len(mappings))
	}
	fmt.Println()
	
	return nil
}

// Complete MDB to clean CSV pipeline
func (mp *MDBProcessor) ProcessComplete() error {
	fmt.Printf("🚀 Starting complete MDB processing pipeline for tenant: %s\n", mp.tenantSlug)
	
	// Step 1: Create working directories
	if err := os.MkdirAll(mp.workingDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}
	
	rawCSVDir := filepath.Join(mp.workingDir, "raw_csv")
	cleanCSVDir := filepath.Join(mp.workingDir, "clean_csv")
	
	// Step 2: Extract MDB to raw CSV
	fmt.Println("📤 Step 1: Extracting MDB to CSV...")
	if err := mp.extractToCSV(rawCSVDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	
	// Step 3: Clean and normalize CSV files
	fmt.Println("🧹 Step 2: Cleaning and normalizing CSV...")
	if err := mp.cleanCSVFiles(rawCSVDir, cleanCSVDir); err != nil {
		return fmt.Errorf("cleaning failed: %w", err)
	}
	
	// Step 4: Generate processing report
	fmt.Println("📊 Step 3: Generating processing report...")
	if err := mp.generateReport(cleanCSVDir); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}
	
	// Step 5: Move final files to output
	finalDir := filepath.Join(mp.outputDir, fmt.Sprintf("tenant_%s_processed", mp.tenantSlug))
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return err
	}
	
	// Copy clean files to final location
	cmd := exec.Command("cp", "-r", cleanCSVDir+"/.", finalDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy final files: %w", err)
	}
	
	fmt.Printf("✅ Processing complete! Files ready at: %s\n", finalDir)
	return nil
}

func (mp *MDBProcessor) extractToCSV(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	
	// Get table list
	cmd := exec.Command("mdb-tables", "-1", mp.mdbFile)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	tableNames := strings.Fields(string(output))
	processedCount := 0
	
	for _, tableName := range tableNames {
		// Skip system tables
		if strings.HasPrefix(tableName, "MSys") || strings.HasPrefix(tableName, "~") {
			continue
		}
		
		outputFile := filepath.Join(outputDir, strings.ToLower(tableName)+".csv")
		
		// Export table
		exportCmd := exec.Command("mdb-export", mp.mdbFile, tableName)
		csvData, err := exportCmd.Output()
		if err != nil {
			fmt.Printf("  ❌ Failed to export %s: %v\n", tableName, err)
			continue
		}
		
		// Write to file
		if err := os.WriteFile(outputFile, csvData, 0644); err != nil {
			fmt.Printf("  ❌ Failed to write %s: %v\n", outputFile, err)
			continue
		}
		
		// Count rows for reporting
		lines := strings.Count(string(csvData), "\n")
		fmt.Printf("  ✅ %s → %s (%d rows)\n", tableName, filepath.Base(outputFile), lines-1)
		processedCount++
	}
	
	fmt.Printf("📤 Extraction complete: %d tables processed\n", processedCount)
	return nil
}

// Use proven Phase 1 CSV cleaning logic
func (mp *MDBProcessor) cleanCSVFiles(inputDir, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	
	processedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			inputFile := filepath.Join(inputDir, entry.Name())
			outputFile := filepath.Join(outputDir, entry.Name())
			
			if err := mp.processCSVFile(inputFile, outputFile); err != nil {
				fmt.Printf("  ❌ Error processing %s: %v\n", entry.Name(), err)
			} else {
				processedCount++
			}
		}
	}
	
	fmt.Printf("🧹 Cleaning complete: %d files processed\n", processedCount)
	return nil
}

// Enhanced version of Phase 1 ProcessCSVFile with better tenant context
func (mp *MDBProcessor) processCSVFile(inputFile, outputFile string) error {
	// Check file accessibility and size to prevent EOF errors
	fileInfo, err := os.Stat(inputFile)
	if err != nil {
		return fmt.Errorf("cannot access file %s: %w", inputFile, err)
	}

	if fileInfo.Size() == 0 {
		fmt.Printf("  ⚠️  Skipping empty file: %s\n", filepath.Base(inputFile))
		return nil
	}

	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", inputFile, err)
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", outputFile, err)
	}
	defer outFile.Close()

	reader := csv.NewReader(inFile)
	reader.FieldsPerRecord = -1 // Allow variable field counts
	reader.TrimLeadingSpace = true

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Read headers with EOF protection
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			fmt.Printf("  ⚠️  Skipping file with no data (EOF): %s\n", filepath.Base(inputFile))
			return nil
		}
		return fmt.Errorf("failed to read headers from %s: %w", inputFile, err)
	}

	if len(headers) == 0 {
		fmt.Printf("  ⚠️  Skipping file with no columns: %s\n", filepath.Base(inputFile))
		return nil
	}

	// Normalize headers and track mappings
	normalizedHeaders := make([]string, len(headers))
	mappingCount := 0
	for i, header := range headers {
		if strings.TrimSpace(header) == "" {
			normalizedHeaders[i] = fmt.Sprintf("column_%d", i+1)
		} else {
			original := strings.ToLower(header)
			normalized := NormalizeColumnName(header)
			normalizedHeaders[i] = normalized
			if original != normalized {
				mappingCount++
			}
		}
	}

	// Write normalized headers
	if err := writer.Write(normalizedHeaders); err != nil {
		return fmt.Errorf("failed to write headers to %s: %w", outputFile, err)
	}

	// Process data rows
	rowCount := 0
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			// Log but continue on parse errors
			fmt.Printf("  ⚠️  Parse error in %s at row %d: %v\n", filepath.Base(inputFile), rowCount+2, err)
			continue
		}

		// Normalize record length to match headers
		for len(record) < len(normalizedHeaders) {
			record = append(record, "")
		}
		if len(record) > len(normalizedHeaders) {
			record = record[:len(normalizedHeaders)]
		}

		if err := writer.Write(record); err != nil {
			fmt.Printf("  ⚠️  Write error in %s at row %d: %v\n", filepath.Base(inputFile), rowCount+2, err)
			continue
		}
		rowCount++
	}

	// Report results
	fmt.Printf("  ✅ %s: %d rows, %d columns", filepath.Base(outputFile), rowCount, len(headers))
	if mappingCount > 0 {
		fmt.Printf(" (%d columns mapped)", mappingCount)
	}
	fmt.Println()

	return nil
}

func (mp *MDBProcessor) generateReport(cleanCSVDir string) error {
	reportFile := filepath.Join(mp.outputDir, fmt.Sprintf("processing_report_%s.md", mp.tenantSlug))
	
	report := fmt.Sprintf(`# MDB Processing Report - Tenant: %s

Generated: %s
Source: %s
Output: %s

## Processing Summary

`, mp.tenantSlug, time.Now().Format("2006-01-02 15:04:05"), mp.mdbFile, cleanCSVDir)

	// Analyze processed files
	entries, err := os.ReadDir(cleanCSVDir)
	if err != nil {
		return err
	}

	totalRows := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".csv") {
			filePath := filepath.Join(cleanCSVDir, entry.Name())
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}
			
			reader := csv.NewReader(file)
			rowCount := 0
			for {
				_, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err == nil {
					rowCount++
				}
			}
			file.Close()
			
			report += fmt.Sprintf("- **%s**: %d rows\n", entry.Name(), rowCount-1) // -1 for header
			totalRows += rowCount - 1
		}
	}

	report += fmt.Sprintf(`
## Total Records: %d

## Ready for Import
Files are cleaned and normalized, ready for PostgreSQL import.

Use command:
make data-import TENANT=%s CSV_DIR=%s
`, totalRows, mp.tenantSlug, cleanCSVDir)

	return os.WriteFile(reportFile, []byte(report), 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Enhanced MDB Processor - Integrating Phase 1 Logic")
		fmt.Println("Usage:")
		fmt.Println("  mdb-processor analyze <mdb_file> [tenant_code]")
		fmt.Println("  mdb-processor process <mdb_file> <tenant_code> <output_dir>")
		fmt.Println("  mdb-processor extract <mdb_file> <output_dir>")
		fmt.Println("  mdb-processor clean <input_csv_dir> <output_csv_dir>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  mdb-processor analyze longbeach.mdb longbeach")
		fmt.Println("  mdb-processor process longbeach.mdb longbeach ./processed")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "analyze":
		if len(os.Args) < 3 {
			log.Fatal("Usage: mdb-processor analyze <mdb_file> [tenant_code]")
		}
		mdbFile := os.Args[2]
		tenantSlug := "default"
		if len(os.Args) > 3 {
			tenantSlug = os.Args[3]
		}
		
		processor := NewMDBProcessor(mdbFile, tenantSlug, "")
		if err := processor.AnalyzeMDB(); err != nil {
			log.Fatal(err)
		}

	case "process":
		if len(os.Args) != 5 {
			log.Fatal("Usage: mdb-processor process <mdb_file> <tenant_code> <output_dir>")
		}
		processor := NewMDBProcessor(os.Args[2], os.Args[3], os.Args[4])
		if err := processor.ProcessComplete(); err != nil {
			log.Fatal(err)
		}

	case "extract":
		if len(os.Args) != 4 {
			log.Fatal("Usage: mdb-processor extract <mdb_file> <output_dir>")
		}
		processor := NewMDBProcessor(os.Args[2], "default", os.Args[3])
		if err := processor.extractToCSV(os.Args[3]); err != nil {
			log.Fatal(err)
		}

	case "clean":
		if len(os.Args) != 4 {
			log.Fatal("Usage: mdb-processor clean <input_csv_dir> <output_csv_dir>")
		}
		processor := NewMDBProcessor("", "default", "")
		if err := processor.cleanCSVFiles(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

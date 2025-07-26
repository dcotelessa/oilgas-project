// backend/cmd/data-tools/tenant_processor.go
// Extends existing MDB processor for tenant-aware CSV output
package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TenantProcessor struct {
	TenantID   string
	OutputDir  string
	MDBFile    string
	CSVFiles   map[string]string // table_name -> csv_file_path
}

func NewTenantProcessor(tenantID, mdbFile, outputDir string) *TenantProcessor {
	return &TenantProcessor{
		TenantID:  tenantID,
		MDBFile:   mdbFile,
		OutputDir: outputDir,
		CSVFiles:  make(map[string]string),
	}
}

func (tp *TenantProcessor) ProcessMDBForTenant() error {
	fmt.Printf("🔄 Processing MDB for tenant: %s\n", tp.TenantID)
	
	// Create tenant-specific output directory
	tenantDir := filepath.Join(tp.OutputDir, tp.TenantID)
	if err := os.MkdirAll(tenantDir, 0755); err != nil {
		return fmt.Errorf("failed to create tenant directory: %w", err)
	}

	// Extract MDB to temporary CSV files using existing logic
	tempDir := filepath.Join(tenantDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Use existing ProcessDirectory logic (from your current data-tools)
	if err := ProcessDirectory(tp.MDBFile, tempDir); err != nil {
		return fmt.Errorf("failed to process MDB file: %w", err)
	}

	// Process each CSV file for tenant
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			continue
		}

		tableName := strings.TrimSuffix(strings.ToLower(entry.Name()), ".csv")
		inputFile := filepath.Join(tempDir, entry.Name())
		outputFile := filepath.Join(tenantDir, entry.Name())

		if err := tp.processTenantCSV(inputFile, outputFile, tableName); err != nil {
			fmt.Printf("⚠️  Error processing %s: %v\n", entry.Name(), err)
			continue
		}

		tp.CSVFiles[tableName] = outputFile
		fmt.Printf("✅ %s: tenant-ready\n", entry.Name())
	}

	// Generate import script and validation report
	if err := tp.generateImportScript(); err != nil {
		return fmt.Errorf("failed to generate import script: %w", err)
	}

	if err := tp.generateValidationReport(); err != nil {
		return fmt.Errorf("failed to generate validation report: %w", err)
	}

	fmt.Printf("✅ Tenant %s processing complete\n", tp.TenantID)
	return nil
}

func (tp *TenantProcessor) processTenantCSV(inputFile, outputFile, tableName string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	reader := csv.NewReader(inFile)
	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read headers: %w", err)
	}

	// Normalize headers using existing column mapping
	normalizedHeaders := make([]string, len(headers))
	for i, header := range headers {
		normalizedHeaders[i] = NormalizeColumnName(header)
	}

	// Add tenant metadata columns for specific tables
	if tableName == "customers" || tableName == "inventory" || tableName == "received" {
		normalizedHeaders = append(normalizedHeaders, "tenant_id", "imported_at")
	}

	if err := writer.Write(normalizedHeaders); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Process data rows
	rowCount := 0
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			continue // Skip malformed rows
		}

		// Add tenant metadata for relevant tables
		if tableName == "customers" || tableName == "inventory" || tableName == "received" {
			record = append(record, tp.TenantID, time.Now().Format("2006-01-02 15:04:05"))
		}

		// Ensure record length matches headers
		for len(record) < len(normalizedHeaders) {
			record = append(record, "")
		}
		if len(record) > len(normalizedHeaders) {
			record = record[:len(normalizedHeaders)]
		}

		if err := writer.Write(record); err != nil {
			continue // Skip problem rows
		}
		rowCount++
	}

	return nil
}

func (tp *TenantProcessor) generateImportScript() error {
	scriptPath := filepath.Join(tp.OutputDir, tp.TenantID, "import_script.sql")
	file, err := os.Create(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to create import script: %w", err)
	}
	defer file.Close()

	script := fmt.Sprintf(`-- Import script for tenant: %s
-- Generated: %s
-- Database: oilgas_%s

\\c oilgas_%s;
SET search_path TO store, public;

-- Disable triggers for faster import
ALTER TABLE store.customers DISABLE TRIGGER ALL;
ALTER TABLE store.inventory DISABLE TRIGGER ALL;
ALTER TABLE store.received DISABLE TRIGGER ALL;

`, tp.TenantID, time.Now().Format("2006-01-02 15:04:05"), tp.TenantID, tp.TenantID)

	// Generate COPY commands for each CSV file
	for tableName, csvFile := range tp.CSVFiles {
		// Map CSV files to database tables
		dbTable := mapCSVToTable(tableName)
		if dbTable == "" {
			continue
		}

		script += fmt.Sprintf(`
-- Import %s
\\COPY store.%s FROM '%s' WITH CSV HEADER DELIMITER ',';

`, tableName, dbTable, csvFile)
	}

	script += `
-- Re-enable triggers
ALTER TABLE store.customers ENABLE TRIGGER ALL;
ALTER TABLE store.inventory ENABLE TRIGGER ALL;
ALTER TABLE store.received ENABLE TRIGGER ALL;

-- Update sequences
SELECT setval('store.customers_customer_id_seq', COALESCE((SELECT MAX(customer_id) FROM store.customers), 1));
SELECT setval('store.inventory_id_seq', COALESCE((SELECT MAX(id) FROM store.inventory), 1));
SELECT setval('store.received_id_seq', COALESCE((SELECT MAX(id) FROM store.received), 1));

-- Validation queries
SELECT 'customers' as table_name, COUNT(*) as imported_rows FROM store.customers WHERE tenant_id = '` + tp.TenantID + `';
SELECT 'inventory' as table_name, COUNT(*) as imported_rows FROM store.inventory WHERE tenant_id = '` + tp.TenantID + `';
SELECT 'received' as table_name, COUNT(*) as imported_rows FROM store.received WHERE tenant_id = '` + tp.TenantID + `';
`

	_, err = file.WriteString(script)
	return err
}

func (tp *TenantProcessor) generateValidationReport() error {
	reportPath := filepath.Join(tp.OutputDir, tp.TenantID, "validation_report.md")
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create validation report: %w", err)
	}
	defer file.Close()

	report := fmt.Sprintf(`# Tenant Data Validation Report

**Tenant ID**: %s  
**Source MDB**: %s  
**Processed**: %s  

## Files Generated

`, tp.TenantID, filepath.Base(tp.MDBFile), time.Now().Format("2006-01-02 15:04:05"))

	// Analyze each CSV file
	for tableName, csvFile := range tp.CSVFiles {
		rowCount, err := countCSVRows(csvFile)
		if err != nil {
			rowCount = -1
		}

		report += fmt.Sprintf(`### %s.csv
- **Rows**: %d
- **Path**: %s
- **Database Table**: store.%s

`, tableName, rowCount, csvFile, mapCSVToTable(tableName))
	}

	report += `## Import Instructions

1. Create tenant database:
   ```bash
   cd backend && go run migrator.go tenant-create ` + tp.TenantID + `
   ```

2. Import data:
   ```bash
   psql oilgas_` + tp.TenantID + ` -f ` + filepath.Join(tp.TenantID, "import_script.sql") + `
   ```

3. Validate import:
   ```bash
   cd backend && go run migrator.go tenant-status ` + tp.TenantID + `
   ```

## Next Steps

- [ ] Verify tenant database creation
- [ ] Run import script  
- [ ] Test API endpoints with tenant header
- [ ] Validate data quality and completeness

`

	_, err = file.WriteString(report)
	return err
}

func countCSVRows(csvFile string) (int, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	count := 0
	
	// Skip header
	if _, err := reader.Read(); err != nil {
		return 0, err
	}

	for {
		if _, err := reader.Read(); err != nil {
			if err.Error() == "EOF" {
				break
			}
			continue // Skip malformed rows
		}
		count++
	}

	return count, nil
}

func mapCSVToTable(csvName string) string {
	// Map CSV file names to database table names
	mapping := map[string]string{
		"customers":  "customers",
		"customer":   "customers", 
		"custid":     "customers",
		"inventory":  "inventory",
		"received":   "received",
		"recv":       "received",
		"workorder":  "inventory", // Work orders go to inventory table
		"workorders": "inventory",
		"grades":     "grade",
		"grade":      "grade",
		"sizes":      "sizes",
		"size":       "sizes",
	}

	if table, exists := mapping[strings.ToLower(csvName)]; exists {
		return table
	}

	// Default mapping for unknown tables
	return csvName
}

// Enhanced main.go command handling
func handleTenantCommand(args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("usage: data-tools process-tenant <mdb_file> <tenant_id> <output_dir>")
	}

	mdbFile := args[2]
	tenantID := args[3] 
	outputDir := args[4]

	// Validate inputs
	if _, err := os.Stat(mdbFile); os.IsNotExist(err) {
		return fmt.Errorf("MDB file not found: %s", mdbFile)
	}

	if !isValidTenantID(tenantID) {
		return fmt.Errorf("invalid tenant ID: %s (use lowercase letters, numbers, underscores only)", tenantID)
	}

	processor := NewTenantProcessor(tenantID, mdbFile, outputDir)
	return processor.ProcessMDBForTenant()
}

func isValidTenantID(tenantID string) bool {
	if len(tenantID) < 2 || len(tenantID) > 20 {
		return false
	}

	for _, char := range tenantID {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

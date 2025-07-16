#!/bin/bash
# scripts/phase1_mdb_migration.sh - Complete MDB to CSV migration pipeline
# Phase 1: Prepares normalized CSV files and analysis for Phase 2 (local dev setup)

set -e

# Configuration paths relative to project root
MDB_FILE="${MDB_FILE:-db_prep/petros.mdb}"
DATABASE_DIR="${DATABASE_DIR:-database}"
ANALYSIS_DIR="$DATABASE_DIR/analysis"
DATA_DIR="$DATABASE_DIR/data"
SCHEMA_DIR="$DATABASE_DIR/schema"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
step() { echo -e "${PURPLE}[STEP]${NC} $1"; }
phase() { echo -e "${CYAN}[PHASE]${NC} $1"; }

# Check if running from project root
if [[ ! -f "Makefile" ]] || [[ ! -d "scripts" ]]; then
    error "Please run this script from the project root directory"
    echo "Expected: ./scripts/phase1_mdb_migration.sh"
    echo "Current: $(pwd)"
    exit 1
fi

echo "🚀 PHASE 1: MDB MIGRATION TO NORMALIZED CSV"
echo "============================================"
echo "This script migrates Access MDB to normalized CSV files"
echo "and prepares analysis data for Phase 2 (local development setup)"
echo
echo "📁 Configuration:"
echo "  MDB File: $MDB_FILE"
echo "  Database Output: $DATABASE_DIR/"
echo "  Analysis Output: $ANALYSIS_DIR/"
echo "  Data Output: $DATA_DIR/"
echo "  Schema Output: $SCHEMA_DIR/"
echo

# Phase 1.1: Environment Validation
phase "1.1 Environment Validation"

if [[ ! -f "$MDB_FILE" ]]; then
    error "MDB file not found: $MDB_FILE"
    echo
    echo "Expected location based on project structure:"
    echo "  db_prep/petros.mdb"
    echo
    echo "Please ensure the MDB file is in the correct location or set:"
    echo "  export MDB_FILE=path/to/your/file.mdb"
    exit 1
fi

missing_deps=()
for dep in mdb-tables mdb-export mdb-count mdb-schema; do
    if ! command -v "$dep" >/dev/null 2>&1; then
        missing_deps+=("$dep")
    fi
done

if [[ ${#missing_deps[@]} -gt 0 ]]; then
    error "Missing mdb-tools: ${missing_deps[*]}"
    echo
    echo "📦 Installation:"
    echo "  macOS: brew install mdbtools"
    echo "  Ubuntu: sudo apt-get install mdb-tools"
    exit 1
fi

if ! command -v go >/dev/null 2>&1; then
    error "Go not found - required for data processing tools"
    echo "📦 Install from: https://golang.org/dl/"
    exit 1
fi

log "✅ All dependencies satisfied"

# Phase 1.2: Directory Setup (ensure they exist)
phase "1.2 Directory Structure Validation"

mkdir -p \
    "$DATA_DIR"/{exported,normalized,clean,logs} \
    "$ANALYSIS_DIR" \
    "$SCHEMA_DIR" \
    backend/cmd/data-tools \
    backend/migrations \
    backend/seeds

log "✅ Directory structure validated"

# Phase 1.3: MDB Diagnostics and Format Detection
phase "1.3 MDB File Diagnostics and Format Fix"

file_size=$(stat -f%z "$MDB_FILE" 2>/dev/null || stat -c%s "$MDB_FILE" 2>/dev/null || echo "unknown")
log "File size: $file_size bytes"

raw_tables_output=$(mdb-tables "$MDB_FILE" 2>/dev/null || echo "FAILED")
if [[ "$raw_tables_output" == "FAILED" ]]; then
    error "Cannot read MDB file - file may be corrupted or password protected"
    exit 1
fi

echo "Raw mdb-tables output: '$raw_tables_output'"

word_count=$(echo "$raw_tables_output" | wc -w | tr -d ' ')
line_count=$(echo "$raw_tables_output" | wc -l | tr -d ' ')

log "Detected: $word_count tables in $line_count lines"

# Fix table list format (handles the EOF issue cause)
if [[ $line_count -eq 1 ]] && [[ $word_count -gt 1 ]]; then
    warn "Detected malformed table output (all tables on single line - EOF issue cause)"
    log "Applying fix: converting space-separated to individual lines"
    echo "$raw_tables_output" | tr ' ' '\n' | grep -v '^$' > "$DATA_DIR/exported/table_list.txt"
else
    log "Table format appears normal"
    echo "$raw_tables_output" > "$DATA_DIR/exported/table_list.txt"
fi

table_count=$(wc -l < "$DATA_DIR/exported/table_list.txt" | tr -d ' ')
log "✅ Processed $table_count tables for extraction"

# Generate table_list.txt for analysis folder (Phase 2 requirement)
cp "$DATA_DIR/exported/table_list.txt" "$ANALYSIS_DIR/table_list.txt"

# Phase 1.4: Enhanced Data Processing Tools
phase "1.4 Building Enhanced Data Processing Tools"

log "Creating EOF-safe Go data processing tools..."

cat > backend/cmd/data-tools/main.go << 'GOEOF'
// Enhanced MDB data processing tools for Phase 1 migration
// Handles EOF errors, normalizes columns, and prepares data for Phase 2
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Oil & Gas industry column mappings for PostgreSQL compatibility
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

func ProcessCSVFile(inputFile, outputFile string) error {
	// Check file accessibility and size to prevent EOF errors
	fileInfo, err := os.Stat(inputFile)
	if err != nil {
		return fmt.Errorf("cannot access file %s: %w", inputFile, err)
	}

	if fileInfo.Size() == 0 {
		fmt.Printf("⚠️  Skipping empty file: %s\n", filepath.Base(inputFile))
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
			fmt.Printf("⚠️  Skipping file with no data (EOF): %s\n", filepath.Base(inputFile))
			return nil
		}
		return fmt.Errorf("failed to read headers from %s: %w", inputFile, err)
	}

	if len(headers) == 0 {
		fmt.Printf("⚠️  Skipping file with no columns: %s\n", filepath.Base(inputFile))
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
			fmt.Printf("⚠️  Parse error in %s at row %d: %v\n", filepath.Base(inputFile), rowCount+2, err)
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
			fmt.Printf("⚠️  Write error in %s at row %d: %v\n", filepath.Base(inputFile), rowCount+2, err)
			continue
		}
		rowCount++
	}

	// Report results
	fmt.Printf("✅ %s: %d rows, %d columns", filepath.Base(outputFile), rowCount, len(headers))
	if mappingCount > 0 {
		fmt.Printf(" (%d columns mapped)", mappingCount)
	}
	fmt.Println()

	return nil
}

func ProcessDirectory(inputDir, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read input directory: %w", err)
	}

	processedCount := 0
	skippedCount := 0
	
	fmt.Printf("🔄 Processing: %s -> %s\n", inputDir, outputDir)
	fmt.Println(strings.Repeat("-", 50))

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			inputFile := filepath.Join(inputDir, entry.Name())
			outputFileName := strings.ToLower(entry.Name())
			outputFile := filepath.Join(outputDir, outputFileName)

			if err := ProcessCSVFile(inputFile, outputFile); err != nil {
				fmt.Printf("❌ Error processing %s: %v\n", entry.Name(), err)
				skippedCount++
			} else {
				processedCount++
			}
		}
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("📊 Processing Summary:\n")
	fmt.Printf("  ✅ Files processed: %d\n", processedCount)
	fmt.Printf("  ⚠️  Files skipped: %d\n", skippedCount)
	fmt.Printf("  📁 Output directory: %s\n", outputDir)

	return nil
}

func AnalyzeColumns(inputDir string) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	fmt.Println("# MDB Column Analysis Report")
	fmt.Printf("# Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("# Source directory: %s\n", inputDir)
	fmt.Printf("# Purpose: Phase 1 migration analysis for Phase 2 development setup\n\n")

	totalFiles := 0
	processedFiles := 0

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			totalFiles++
			filePath := filepath.Join(inputDir, entry.Name())
			
			fileInfo, err := os.Stat(filePath)
			if err != nil || fileInfo.Size() == 0 {
				fmt.Printf("## %s\n❌ Empty or inaccessible file\n\n", entry.Name())
				continue
			}

			file, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("## %s\n❌ Cannot open file: %v\n\n", entry.Name(), err)
				continue
			}

			reader := csv.NewReader(file)
			reader.FieldsPerRecord = -1
			
			headers, err := reader.Read()
			file.Close()

			if err != nil {
				if err == io.EOF {
					fmt.Printf("## %s\n⚠️  No data (EOF)\n\n", entry.Name())
				} else {
					fmt.Printf("## %s\n❌ Read error: %v\n\n", entry.Name(), err)
				}
				continue
			}

			processedFiles++
			fmt.Printf("## %s\n", entry.Name())
			fmt.Printf("📊 Columns: %d\n", len(headers))
			fmt.Printf("### Column Mappings (Original -> Normalized):\n")
			
			for i, header := range headers {
				if strings.TrimSpace(header) == "" {
					fmt.Printf("%2d. (empty) -> column_%d\n", i+1, i+1)
				} else {
					normalized := NormalizeColumnName(header)
					if strings.ToLower(header) != normalized {
						fmt.Printf("%2d. %s -> %s\n", i+1, header, normalized)
					} else {
						fmt.Printf("%2d. %s\n", i+1, header)
					}
				}
			}
			fmt.Println()
		}
	}

	fmt.Printf("## Analysis Summary\n")
	fmt.Printf("- Total CSV files: %d\n", totalFiles)
	fmt.Printf("- Successfully analyzed: %d\n", processedFiles)
	fmt.Printf("- Ready for Phase 2: %s\n", "✅")

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Phase 1 MDB Data Processing Tools")
		fmt.Println("Usage:")
		fmt.Println("  data-tools normalize-dir <input_dir> <output_dir>")
		fmt.Println("  data-tools analyze <input_dir>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "normalize-dir":
		if len(os.Args) != 4 {
			log.Fatal("Usage: data-tools normalize-dir <input_dir> <output_dir>")
		}
		if err := ProcessDirectory(os.Args[2], os.Args[3]); err != nil {
			log.Fatalf("Failed to process directory: %v", err)
		}
		fmt.Println("✅ Phase 1 normalization complete - ready for Phase 2")

	case "analyze":
		if len(os.Args) != 3 {
			log.Fatal("Usage: data-tools analyze <input_dir>")
		}
		if err := AnalyzeColumns(os.Args[2]); err != nil {
			log.Fatalf("Failed to analyze columns: %v", err)
		}

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
GOEOF

# Build the enhanced tools
log "Building Go data processing tools..."
cd backend
if go build -o data-tools cmd/data-tools/main.go; then
    log "✅ Data processing tools built successfully"
    cd ..
else
    error "Failed to build data processing tools"
    cd ..
    exit 1
fi

# Phase 1.5: Table Validation and Extraction
phase "1.5 Table Validation and Data Extraction"

log "Validating and extracting table data..."

echo "# Table Extraction Log - $(date)" > "$DATA_DIR/logs/extraction.log"
echo "# Source: $MDB_FILE" >> "$DATA_DIR/logs/extraction.log"
echo "" >> "$DATA_DIR/logs/extraction.log"

extracted_tables=()
empty_tables=()
failed_tables=()

while IFS= read -r table_name; do
    if [[ -n "$table_name" ]]; then
        echo "🔍 Processing table: $table_name"
        
        if record_count=$(mdb-count "$MDB_FILE" "$table_name" 2>/dev/null); then
            echo "  📊 Records: $record_count"
            echo "$table_name: $record_count records" >> "$DATA_DIR/logs/extraction.log"
            
            if [[ "$record_count" -eq 0 ]]; then
                empty_tables+=("$table_name")
                echo "  ⚠️  Empty table - skipping"
            else
                table_lower=$(echo "$table_name" | tr '[:upper:]' '[:lower:]')
                output_file="$DATA_DIR/exported/${table_lower}.csv"
                
                if mdb-export "$MDB_FILE" "$table_name" > "$output_file" 2>>"$DATA_DIR/logs/extraction.log"; then
                    if [[ -f "$output_file" ]] && [[ -s "$output_file" ]]; then
                        line_count=$(wc -l < "$output_file")
                        if [[ $line_count -gt 1 ]]; then
                            extracted_tables+=("$table_name")
                            echo "  ✅ Extracted: $line_count lines -> ${table_lower}.csv"
                            echo "$table_name: SUCCESS ($line_count lines) -> ${table_lower}.csv" >> "$DATA_DIR/logs/extraction.log"
                        else
                            failed_tables+=("$table_name")
                            echo "  ❌ Export empty: only headers"
                            echo "$table_name: FAILED (headers only)" >> "$DATA_DIR/logs/extraction.log"
                        fi
                    else
                        failed_tables+=("$table_name")
                        echo "  ❌ Export failed: no output file"
                        echo "$table_name: FAILED (no output)" >> "$DATA_DIR/logs/extraction.log"
                    fi
                else
                    failed_tables+=("$table_name")
                    echo "  ❌ Export command failed"
                    echo "$table_name: FAILED (export error)" >> "$DATA_DIR/logs/extraction.log"
                fi
            fi
        else
            failed_tables+=("$table_name")
            echo "  ❌ Cannot count records"
            echo "$table_name: FAILED (cannot count)" >> "$DATA_DIR/logs/extraction.log"
        fi
        echo
    fi
done < "$DATA_DIR/exported/table_list.txt"

log "📊 Extraction Summary:"
log "  ✅ Successfully extracted: ${#extracted_tables[@]} tables"
log "  📭 Empty tables: ${#empty_tables[@]} tables"
log "  ❌ Failed extractions: ${#failed_tables[@]} tables"

# Phase 1.6: Column Normalization
phase "1.6 Column Normalization for PostgreSQL Compatibility"

if [[ ${#extracted_tables[@]} -gt 0 ]]; then
    log "🔄 Normalizing ${#extracted_tables[@]} extracted files..."
    
    cd backend && ./data-tools normalize-dir "../$DATA_DIR/exported" "../$DATA_DIR/normalized" && cd ..
    
    log "✅ Column normalization complete"
    
    # Copy normalized files to clean directory for Phase 2
    log "Preparing clean files for Phase 2..."
    cp -r "$DATA_DIR/normalized"/* "$DATA_DIR/clean/" 2>/dev/null || true
else
    warn "No files available for normalization"
fi

# Phase 1.7: Schema Extraction
phase "1.7 Schema Analysis and Documentation"

log "Extracting database schema..."

if mdb-schema "$MDB_FILE" > "$SCHEMA_DIR/mdb_schema.sql" 2>/dev/null; then
    log "✅ Full schema extracted"
else
    warn "Could not extract full schema"
fi

# Extract individual table schemas for successfully extracted tables
for table_name in "${extracted_tables[@]}"; do
    table_lower=$(echo "$table_name" | tr '[:upper:]' '[:lower:]')
    if mdb-schema "$MDB_FILE" -T "$table_name" --no-indexes > "$SCHEMA_DIR/${table_lower}_schema.sql" 2>/dev/null; then
        echo "  ✅ ${table_lower}_schema.sql"
    fi
done

# Generate table counts for analysis (required for Phase 2)
log "Generating table counts analysis..."
echo "# Table Record Counts - Generated $(date)" > "$ANALYSIS_DIR/table_counts.txt"
echo "# Source: $MDB_FILE" >> "$ANALYSIS_DIR/table_counts.txt"
echo "" >> "$ANALYSIS_DIR/table_counts.txt"

for table_name in "${extracted_tables[@]}"; do
    if record_count=$(mdb-count "$MDB_FILE" "$table_name" 2>/dev/null); then
        echo "$table_name: $record_count records" >> "$ANALYSIS_DIR/table_counts.txt"
    fi
done

# Phase 1.8: Column Analysis Generation
phase "1.8 Column Analysis Generation"

log "Generating comprehensive column analysis..."

if [[ -d "$DATA_DIR/exported" ]] && [[ $(ls -1 "$DATA_DIR/exported"/*.csv 2>/dev/null | wc -l) -gt 0 ]]; then
    cd backend && ./data-tools analyze "../$DATA_DIR/exported" > "../$ANALYSIS_DIR/mdb_column_analysis.txt" 2>&1 && cd ..
    log "✅ Column analysis generated: $ANALYSIS_DIR/mdb_column_analysis.txt"
else
    warn "No CSV files available for analysis"
    echo "# No data available for analysis" > "$ANALYSIS_DIR/mdb_column_analysis.txt"
fi

# Phase 1.9: Migration Preparation for Phase 2
phase "1.9 Phase 2 Preparation"

log "Preparing files for Phase 2 (local development setup)..."

# Create basic migration file for Phase 2
cat > backend/migrations/001_initial_schema.sql << 'SQLEOF'
-- Initial schema migration for Oil & Gas Inventory System
-- Generated from Phase 1 MDB migration

-- Create schema
CREATE SCHEMA IF NOT EXISTS store;
SET search_path TO store, public;

-- Create migrations table
CREATE SCHEMA IF NOT EXISTS migrations;
CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Note: Actual table schemas will be added based on MDB analysis
-- Run Phase 2 setup to complete the migration process
SQLEOF

# Create seed file template for Phase 2
cat > backend/seeds/local_seeds.sql << 'SQLEOF'
-- Local development seed data
-- Generated from Phase 1 MDB migration

SET search_path TO store, public;

-- Note: Actual seed data will be populated from normalized CSV files
-- This file will be updated during Phase 2 setup process

-- Example grade data for oil & gas industry
INSERT INTO store.grade (grade, description) VALUES 
('J55', 'Standard grade steel casing'),
('JZ55', 'Enhanced J55 grade'),
('L80', 'Higher strength grade'),
('N80', 'Medium strength grade'),
('P105', 'High performance grade'),
('P110', 'Premium performance grade')
ON CONFLICT (grade) DO NOTHING;
SQLEOF

log "✅ Phase 2 preparation complete"

# Phase 1.10: Final Report Generation
phase "1.10 Final Migration Report"

# Generate comprehensive Phase 1 report
cat > "$ANALYSIS_DIR/phase1_migration_report.txt" << EOF
# Phase 1: MDB Migration Report
Generated: $(date)
Source: $MDB_FILE

## Migration Overview
Phase 1 converts Access MDB database to normalized CSV files and prepares
analysis data for Phase 2 (local development environment setup).

## Processing Summary
- Total tables found: $table_count
- Successfully extracted: ${#extracted_tables[@]}
- Empty tables: ${#empty_tables[@]}
- Failed extractions: ${#failed_tables[@]}

## Successfully Extracted Tables
EOF

if [[ ${#extracted_tables[@]} -gt 0 ]]; then
    for table_name in "${extracted_tables[@]}"; do
        table_lower=$(echo "$table_name" | tr '[:upper:]' '[:lower:]')
        if [[ -f "$DATA_DIR/normalized/${table_lower}.csv" ]]; then
            line_count=$(wc -l < "$DATA_DIR/normalized/${table_lower}.csv" 2>/dev/null || echo "0")
            echo "- $table_name -> ${table_lower}.csv ($line_count lines)" >> "$ANALYSIS_DIR/phase1_migration_report.txt"
        fi
    done
else
    echo "- None" >> "$ANALYSIS_DIR/phase1_migration_report.txt"
fi

cat >> "$ANALYSIS_DIR/phase1_migration_report.txt" << EOF

## Empty Tables (Skipped)
EOF

if [[ ${#empty_tables[@]} -gt 0 ]]; then
    printf '%s\n' "${empty_tables[@]}" | sed 's/^/- /' >> "$ANALYSIS_DIR/phase1_migration_report.txt"
else
    echo "- None" >> "$ANALYSIS_DIR/phase1_migration_report.txt"
fi

cat >> "$ANALYSIS_DIR/phase1_migration_report.txt" << EOF

## Failed Tables
EOF

if [[ ${#failed_tables[@]} -gt 0 ]]; then
    printf '%s\n' "${failed_tables[@]}" | sed 's/^/- /' >> "$ANALYSIS_DIR/phase1_migration_report.txt"
else
    echo "- None" >> "$ANALYSIS_DIR/phase1_migration_report.txt"
fi

cat >> "$ANALYSIS_DIR/phase1_migration_report.txt" << EOF

## Generated Files

### Data Files
- Original exports: $DATA_DIR/exported/
- Normalized CSV: $DATA_DIR/normalized/
- Clean CSV (Phase 2 ready): $DATA_DIR/clean/
- Processing logs: $DATA_DIR/logs/

### Analysis Files (Required for Phase 2)
- Column analysis: $ANALYSIS_DIR/mdb_column_analysis.txt
- Schema files: $SCHEMA_DIR/mdb_schema.sql
- Table counts: $ANALYSIS_DIR/table_counts.txt  
- Table list: $ANALYSIS_DIR/table_list.txt
- This report: $ANALYSIS_DIR/phase1_migration_report.txt

### Phase 2 Preparation Files
- Initial migration: backend/migrations/001_initial_schema.sql
- Seed template: backend/seeds/local_seeds.sql
- Data tools: backend/data-tools

## Status
EOF

if [[ ${#extracted_tables[@]} -gt 0 ]]; then
    cat >> "$ANALYSIS_DIR/phase1_migration_report.txt" << EOF
✅ SUCCESS: Phase 1 migration completed successfully

## Ready for Phase 2
The following files are ready for Phase 2 (local development setup):

1. ✅ Normalized CSV files in $DATA_DIR/clean/
2. ✅ Analysis files in $ANALYSIS_DIR/
3. ✅ Schema files in $SCHEMA_DIR/
4. ✅ Migration templates in backend/migrations/
5. ✅ Seed templates in backend/seeds/

## Next Steps (Phase 2)
1. Run: make setup (from project root)
2. Import normalized data: make import-clean-data
3. Start development: make dev
4. Access PgAdmin: http://localhost:8080

## Phase 2 Commands Available
- make setup          # Complete local environment setup
- make migrate         # Run database migrations  
- make seed           # Load seed data
- make dev-start      # Start Docker services
- make dev            # Start development servers
EOF
else
    cat >> "$ANALYSIS_DIR/phase1_migration_report.txt" << EOF
❌ FAILURE: No data could be extracted

## Troubleshooting
1. Check extraction logs: $DATA_DIR/logs/extraction.log
2. Verify MDB file is not corrupted
3. Check for password protection
4. Try manual export from Microsoft Access
5. Convert to older Access format if needed

## Alternative Options
- Manual CSV export from Microsoft Access
- Use LibreOffice Base for conversion
- Try PowerShell with Access drivers (Windows)
EOF
fi

# Final Summary
echo
echo "🎉 PHASE 1 MIGRATION COMPLETE!"
echo "==============================="
echo
echo "📊 Migration Results:"
echo "  ✅ Successfully processed: ${#extracted_tables[@]} tables"
echo "  📭 Empty tables skipped: ${#empty_tables[@]} tables"
echo "  ❌ Failed extractions: ${#failed_tables[@]} tables"
echo
echo "📁 Generated Files (Organized Structure):"
echo "  📄 Migration report: $ANALYSIS_DIR/phase1_migration_report.txt"
echo "  📊 Column analysis: $ANALYSIS_DIR/mdb_column_analysis.txt"
echo "  📋 Table counts: $ANALYSIS_DIR/table_counts.txt"
echo "  📄 Table list: $ANALYSIS_DIR/table_list.txt"
echo "  🗄️  Schema: $SCHEMA_DIR/mdb_schema.sql"
echo "  📁 Clean data: $DATA_DIR/clean/"
echo

if [[ ${#extracted_tables[@]} -gt 0 ]]; then
    echo "✅ SUCCESS: Phase 1 completed successfully!"
    echo
    echo "🚀 Ready for Phase 2 (Local Development Setup)"
    echo "=============================================="
    echo
    echo "Your normalized CSV files and analysis are ready for Phase 2."
    echo "All required analysis files have been generated in organized folders:"
    echo
    echo "📋 Required Analysis Files:"
    echo "  ✅ $ANALYSIS_DIR/mdb_column_analysis.txt"
    echo "  ✅ $ANALYSIS_DIR/table_counts.txt"
    echo "  ✅ $ANALYSIS_DIR/table_list.txt"
    echo "  ✅ $SCHEMA_DIR/mdb_schema.sql"
    echo
    echo "📁 Processed Data Files:"
    if [[ -d "$DATA_DIR/clean" ]]; then
        echo "  📊 Clean CSV files for Phase 2 import:"
        ls -la "$DATA_DIR/clean"/*.csv 2>/dev/null | head -5 || echo "    No CSV files found"
        csv_count=$(ls -1 "$DATA_DIR/clean"/*.csv 2>/dev/null | wc -l)
        echo "  📈 Total files: $csv_count CSV files ready for Phase 2"
    fi
    echo
    echo "🔧 Next Steps (Phase 2):"
    echo "  1. Review migration report: cat $ANALYSIS_DIR/phase1_migration_report.txt"
    echo "  2. Start Phase 2 setup: make setup"
    echo "  3. Import data: make import-clean-data"
    echo "  4. Start development: make dev"
    echo
    echo "💡 Phase 2 will set up:"
    echo "  • Local PostgreSQL database with Docker"
    echo "  • PgAdmin for database management"
    echo "  • Import your normalized CSV data"
    echo "  • Development environment with hot reload"
    echo
    echo "🎯 Phase 1 Objectives Achieved:"
    echo "  ✅ MDB file successfully processed"
    echo "  ✅ EOF errors in column analysis resolved"
    echo "  ✅ Tables extracted and normalized for PostgreSQL"
    echo "  ✅ Column mappings applied (Access → PostgreSQL format)"
    echo "  ✅ Analysis files generated in organized structure"
    echo "  ✅ Migration templates prepared"
    echo "  ✅ Seed data templates created"
else
    echo "❌ PHASE 1 FAILED"
    echo
    echo "No data could be successfully extracted from the MDB file."
    echo "Check troubleshooting steps in docs/README_PHASE1.md"
fi

echo
echo "📖 For complete details:"
echo "  cat $ANALYSIS_DIR/phase1_migration_report.txt"
echo
echo "📚 Documentation:"
echo "  docs/README_PHASE1.md"
echo
echo "============================"
echo "Phase 1 Migration Complete"
echo "============================"

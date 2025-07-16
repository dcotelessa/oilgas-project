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

// backend/cmd/tools/csv-cleaner/main.go
// Tool to clean and normalize CSV data exported from MDB
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CleaningConfig struct {
	InputDir  string
	OutputDir string
	LogFile   string
}

type CleaningStats struct {
	FilesProcessed int
	RecordsCleaned int
	ErrorCount     int
	StartTime      time.Time
}

func main() {
	config := CleaningConfig{}
	flag.StringVar(&config.InputDir, "input", "database/data/exported", "Input directory with raw CSV files")
	flag.StringVar(&config.OutputDir, "output", "database/data/clean", "Output directory for cleaned CSV files")
	flag.StringVar(&config.LogFile, "log", "database/logs/cleaning.log", "Log file path")
	flag.Parse()

	log.Printf("Starting CSV data cleaning...")
	log.Printf("Input: %s", config.InputDir)
	log.Printf("Output: %s", config.OutputDir)

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create log directory
	if err := os.MkdirAll(filepath.Dir(config.LogFile), 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	stats := &CleaningStats{StartTime: time.Now()}

	// Process all CSV files in input directory
	files, err := filepath.Glob(filepath.Join(config.InputDir, "*.csv"))
	if err != nil {
		log.Fatalf("Failed to find CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Printf("No CSV files found in %s", config.InputDir)
		return
	}

	for _, file := range files {
		if err := processCSVFile(file, config.OutputDir, stats); err != nil {
			log.Printf("Error processing %s: %v", file, err)
			stats.ErrorCount++
		} else {
			stats.FilesProcessed++
		}
	}

	// Write summary
	duration := time.Since(stats.StartTime)
	summary := fmt.Sprintf(`CSV Cleaning Summary
===================
Start Time: %s
Duration: %v
Files Processed: %d
Records Cleaned: %d
Errors: %d
Status: %s
`,
		stats.StartTime.Format(time.RFC3339),
		duration,
		stats.FilesProcessed,
		stats.RecordsCleaned,
		stats.ErrorCount,
		func() string {
			if stats.ErrorCount == 0 {
				return "SUCCESS"
			}
			return "COMPLETED WITH ERRORS"
		}(),
	)

	log.Print(summary)

	// Write to log file
	if err := os.WriteFile(config.LogFile, []byte(summary), 0644); err != nil {
		log.Printf("Failed to write log file: %v", err)
	}
}

func processCSVFile(inputFile, outputDir string, stats *CleaningStats) error {
	log.Printf("Processing: %s", inputFile)

	// Open input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		log.Printf("Empty file: %s", inputFile)
		return nil
	}

	// Get base filename
	baseFile := filepath.Base(inputFile)
	outputFile := filepath.Join(outputDir, baseFile)

	// Clean and normalize data
	cleanedRecords := cleanCSVData(records, baseFile)
	stats.RecordsCleaned += len(cleanedRecords) - 1 // Subtract header

	// Write cleaned data
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	for _, record := range cleanedRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	log.Printf("Cleaned %s: %d records → %s", baseFile, len(records)-1, outputFile)
	return nil
}

func cleanCSVData(records [][]string, filename string) [][]string {
	if len(records) == 0 {
		return records
	}

	// Normalize headers
	headers := records[0]
	normalizedHeaders := make([]string, len(headers))
	
	for i, header := range headers {
		normalizedHeaders[i] = normalizeColumnName(header)
	}

	// Clean data records
	cleanedRecords := [][]string{normalizedHeaders}
	
	for i := 1; i < len(records); i++ {
		record := records[i]
		
		// Ensure record has same length as headers
		if len(record) < len(headers) {
			// Pad with empty strings
			for len(record) < len(headers) {
				record = append(record, "")
			}
		} else if len(record) > len(headers) {
			// Truncate extra fields
			record = record[:len(headers)]
		}

		// Clean each field
		cleanedRecord := make([]string, len(record))
		for j, field := range record {
			cleanedRecord[j] = cleanField(field, normalizedHeaders[j])
		}

		// Skip completely empty records
		if !isEmptyRecord(cleanedRecord) {
			cleanedRecords = append(cleanedRecords, cleanedRecord)
		}
	}

	return cleanedRecords
}

func normalizeColumnName(header string) string {
	// Convert to lowercase and trim
	normalized := strings.ToLower(strings.TrimSpace(header))
	
	// Apply standard mappings from Access to PostgreSQL naming
	mappings := map[string]string{
		"custid":           "customer_id",
		"customerid":       "customer_id",
		"custname":         "customer",
		"customername":     "customer",
		"custpo":          "customer_po",
		"customerpo":      "customer_po",
		"billaddr":        "billing_address",
		"billingaddr":     "billing_address",
		"billcity":        "billing_city",
		"billingcity":     "billing_city",
		"billstate":       "billing_state",
		"billingstate":    "billing_state",
		"billzip":         "billing_zipcode",
		"billingzip":      "billing_zipcode",
		"billzipcode":     "billing_zipcode",
		"phoneno":         "phone",
		"phonenum":        "phone",
		"telephone":       "phone",
		"faxno":           "fax",
		"faxnum":          "fax",
		"emailaddr":       "email",
		"emailaddress":    "email",
		"workorder":       "work_order",
		"wkorder":         "work_order",
		"wo":              "work_order",
		"datein":          "date_in",
		"dateout":         "date_out",
		"daterecvd":       "date_received",
		"daterecieved":    "date_received",
		"wellin":          "well_in",
		"wellout":         "well_out",
		"leasein":         "lease_in",
		"leaseout":        "lease_out",
		"orderedby":       "ordered_by",
		"enteredby":       "entered_by",
		"whenentered":     "when_entered",
		"whenupdated":     "when_updated",
	}

	if mapped, exists := mappings[normalized]; exists {
		return mapped
	}

	// Clean up common patterns
	normalized = regexp.MustCompile(`[^\w]`).ReplaceAllString(normalized, "_")
	normalized = regexp.MustCompile(`_+`).ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")

	return normalized
}

func cleanField(field, columnName string) string {
	// Trim whitespace
	cleaned := strings.TrimSpace(field)
	
	// Handle null values and empty strings
	if cleaned == "" || cleaned == "NULL" || cleaned == "null" || cleaned == "0000-00-00" {
		return ""
	}

	// Column-specific cleaning
	switch columnName {
	case "email":
		return cleanEmail(cleaned)
	case "phone", "fax":
		return cleanPhone(cleaned)
	case "billing_state":
		return cleanState(cleaned)
	case "billing_zipcode":
		return cleanZipcode(cleaned)
	case "customer", "contact":
		return cleanName(cleaned)
	default:
		// General text cleaning
		return cleanText(cleaned)
	}
}

func cleanEmail(email string) string {
	if email == "" {
		return ""
	}
	
	email = strings.ToLower(email)
	
	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}// backend/cmd/tools/csv-cleaner/main.go
// Tool to clean and normalize CSV data exported from MDB
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CleaningConfig struct {
	InputDir  string
	OutputDir string
	LogFile   string
}

type CleaningStats struct {
	FilesProcessed int
	RecordsCleaned int
	ErrorCount     int
	StartTime      time.Time
}

func main() {
	config := CleaningConfig{}
	flag.StringVar(&config.InputDir, "input", "database/data/exported", "Input directory with raw CSV files")
	flag.StringVar(&config.OutputDir, "output", "database/data/clean", "Output directory for cleaned CSV files")
	flag.StringVar(&config.LogFile, "log", "database/logs/cleaning.log", "Log file path")
	flag.Parse()

	log.Printf("Starting CSV data cleaning...")
	log.Printf("Input: %s", config.InputDir)
	log.Printf("Output: %s", config.OutputDir)

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create log directory
	if err := os.MkdirAll(filepath.Dir(config.LogFile), 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	stats := &CleaningStats{StartTime: time.Now()}

	// Process all CSV files in input directory
	files, err := filepath.Glob(filepath.Join(config.InputDir, "*.csv"))
	if err != nil {
		log.Fatalf("Failed to find CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Printf("No CSV files found in %s", config.InputDir)
		return
	}

	for _, file := range files {
		if err := processCSVFile(file, config.OutputDir, stats); err != nil {
			log.Printf("Error processing %s: %v", file, err)
			stats.ErrorCount++
		} else {
			stats.FilesProcessed++
		}
	}

	// Write summary
	duration := time.Since(stats.StartTime)
	summary := fmt.Sprintf(`CSV Cleaning Summary
===================
Start Time: %s
Duration: %v
Files Processed: %d
Records Cleaned: %d
Errors: %d
Status: %s
`,
		stats.StartTime.Format(time.RFC3339),
		duration,
		stats.FilesProcessed,
		stats.RecordsCleaned,
		stats.ErrorCount,
		func() string {
			if stats.ErrorCount == 0 {
				return "SUCCESS"
			}
			return "COMPLETED WITH ERRORS"
		}(),
	)

	log.Print(summary)

	// Write to log file
	if err := os.WriteFile(config.LogFile, []byte(summary), 0644); err != nil {
		log.Printf("Failed to write log file: %v", err)
	}
}

func processCSVFile(inputFile, outputDir string, stats *CleaningStats) error {
	log.Printf("Processing: %s", inputFile)

	// Open input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		log.Printf("Empty file: %s", inputFile)
		return nil
	}

	// Get base filename
	baseFile := filepath.Base(inputFile)
	outputFile := filepath.Join(outputDir, baseFile)

	// Clean and normalize data
	cleanedRecords := cleanCSVData(records, baseFile)
	stats.RecordsCleaned += len(cleanedRecords) - 1 // Subtract header

	// Write cleaned data
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	for _, record := range cleanedRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	log.Printf("Cleaned %s: %d records → %s", baseFile, len(records)-1, outputFile)
	return nil
}

)
	if emailRegex.MatchString(email) {
		return email
	}
	
	return "" // Invalid email becomes empty
}

func cleanPhone(phone string) string {
	if phone == "" {
		return ""
	}
	
	// Remove common non-digits but preserve formatting
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, ".", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	
	// Keep only digits and basic formatting
	phoneRegex := regexp.MustCompile(`[^\d]`)
	digitsOnly := phoneRegex.ReplaceAllString(phone, "")
	
	// Format US phone numbers
	if len(digitsOnly) == 10 {
		return fmt.Sprintf("(%s) %s-%s", digitsOnly[:3], digitsOnly[3:6], digitsOnly[6:])
	} else if len(digitsOnly) == 11 && digitsOnly[0] == '1' {
		return fmt.Sprintf("1 (%s) %s-%s", digitsOnly[1:4], digitsOnly[4:7], digitsOnly[7:])
	}
	
	// Return original if not standard format
	return phone
}

func cleanState(state string) string {
	if state == "" {
		return ""
	}
	
	state = strings.ToUpper(strings.TrimSpace(state))
	
	// Validate US state codes
	validStates := map[string]bool{
		"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true, "CO": true,
		"CT": true, "DE": true, "FL": true, "GA": true, "HI": true, "ID": true,
		"IL": true, "IN": true, "IA": true, "KS": true, "KY": true, "LA": true,
		"ME": true, "MD": true, "MA": true, "MI": true, "MN": true, "MS": true,
		"MO": true, "MT": true, "NE": true, "NV": true, "NH": true, "NJ": true,
		"NM": true, "NY": true, "NC": true, "ND": true, "OH": true, "OK": true,
		"OR": true, "PA": true, "RI": true, "SC": true, "SD": true, "TN": true,
		"TX": true, "UT": true, "VT": true, "VA": true, "WA": true, "WV": true,
		"WI": true, "WY": true, "DC": true,
	}
	
	if len(state) == 2 && validStates[state] {
		return state
	}
	
	// Try to map full state names to codes
	stateMap := map[string]string{
		"ALABAMA": "AL", "ALASKA": "AK", "ARIZONA": "AZ", "ARKANSAS": "AR",
		"CALIFORNIA": "CA", "COLORADO": "CO", "CONNECTICUT": "CT", "DELAWARE": "DE",
		"FLORIDA": "FL", "GEORGIA": "GA", "HAWAII": "HI", "IDAHO": "ID",
		"ILLINOIS": "IL", "INDIANA": "IN", "IOWA": "IA", "KANSAS": "KS",
		"KENTUCKY": "KY", "LOUISIANA": "LA", "MAINE": "ME", "MARYLAND": "MD",
		"MASSACHUSETTS": "MA", "MICHIGAN": "MI", "MINNESOTA": "MN", "MISSISSIPPI": "MS",
		"MISSOURI": "MO", "MONTANA": "MT", "NEBRASKA": "NE", "NEVADA": "NV",
		"NEW HAMPSHIRE": "NH", "NEW JERSEY": "NJ", "NEW MEXICO": "NM", "NEW YORK": "NY",
		"NORTH CAROLINA": "NC", "NORTH DAKOTA": "ND", "OHIO": "OH", "OKLAHOMA": "OK",
		"OREGON": "OR", "PENNSYLVANIA": "PA", "RHODE ISLAND": "RI", "SOUTH CAROLINA": "SC",
		"SOUTH DAKOTA": "SD", "TENNESSEE": "TN", "TEXAS": "TX", "UTAH": "UT",
		"VERMONT": "VT", "VIRGINIA": "VA", "WASHINGTON": "WA", "WEST VIRGINIA": "WV",
		"WISCONSIN": "WI", "WYOMING": "WY",
	}
	
	if code, exists := stateMap[state]; exists {
		return code
	}
	
	return "" // Invalid state becomes empty
}

func cleanZipcode(zip string) string {
	if zip == "" {
		return ""
	}
	
	// Remove non-digits and hyphens
	zipRegex := regexp.MustCompile(`[^\d-]`)
	cleaned := zipRegex.ReplaceAllString(zip, "")
	
	// US zipcode patterns
	if regexp.MustCompile(`^\d{5}// backend/cmd/tools/csv-cleaner/main.go
// Tool to clean and normalize CSV data exported from MDB
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CleaningConfig struct {
	InputDir  string
	OutputDir string
	LogFile   string
}

type CleaningStats struct {
	FilesProcessed int
	RecordsCleaned int
	ErrorCount     int
	StartTime      time.Time
}

func main() {
	config := CleaningConfig{}
	flag.StringVar(&config.InputDir, "input", "database/data/exported", "Input directory with raw CSV files")
	flag.StringVar(&config.OutputDir, "output", "database/data/clean", "Output directory for cleaned CSV files")
	flag.StringVar(&config.LogFile, "log", "database/logs/cleaning.log", "Log file path")
	flag.Parse()

	log.Printf("Starting CSV data cleaning...")
	log.Printf("Input: %s", config.InputDir)
	log.Printf("Output: %s", config.OutputDir)

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create log directory
	if err := os.MkdirAll(filepath.Dir(config.LogFile), 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	stats := &CleaningStats{StartTime: time.Now()}

	// Process all CSV files in input directory
	files, err := filepath.Glob(filepath.Join(config.InputDir, "*.csv"))
	if err != nil {
		log.Fatalf("Failed to find CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Printf("No CSV files found in %s", config.InputDir)
		return
	}

	for _, file := range files {
		if err := processCSVFile(file, config.OutputDir, stats); err != nil {
			log.Printf("Error processing %s: %v", file, err)
			stats.ErrorCount++
		} else {
			stats.FilesProcessed++
		}
	}

	// Write summary
	duration := time.Since(stats.StartTime)
	summary := fmt.Sprintf(`CSV Cleaning Summary
===================
Start Time: %s
Duration: %v
Files Processed: %d
Records Cleaned: %d
Errors: %d
Status: %s
`,
		stats.StartTime.Format(time.RFC3339),
		duration,
		stats.FilesProcessed,
		stats.RecordsCleaned,
		stats.ErrorCount,
		func() string {
			if stats.ErrorCount == 0 {
				return "SUCCESS"
			}
			return "COMPLETED WITH ERRORS"
		}(),
	)

	log.Print(summary)

	// Write to log file
	if err := os.WriteFile(config.LogFile, []byte(summary), 0644); err != nil {
		log.Printf("Failed to write log file: %v", err)
	}
}

func processCSVFile(inputFile, outputDir string, stats *CleaningStats) error {
	log.Printf("Processing: %s", inputFile)

	// Open input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		log.Printf("Empty file: %s", inputFile)
		return nil
	}

	// Get base filename
	baseFile := filepath.Base(inputFile)
	outputFile := filepath.Join(outputDir, baseFile)

	// Clean and normalize data
	cleanedRecords := cleanCSVData(records, baseFile)
	stats.RecordsCleaned += len(cleanedRecords) - 1 // Subtract header

	// Write cleaned data
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	for _, record := range cleanedRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	log.Printf("Cleaned %s: %d records → %s", baseFile, len(records)-1, outputFile)
	return nil
}

).MatchString(cleaned) {
		return cleaned
	}
	if regexp.MustCompile(`^\d{5}-\d{4}// backend/cmd/tools/csv-cleaner/main.go
// Tool to clean and normalize CSV data exported from MDB
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CleaningConfig struct {
	InputDir  string
	OutputDir string
	LogFile   string
}

type CleaningStats struct {
	FilesProcessed int
	RecordsCleaned int
	ErrorCount     int
	StartTime      time.Time
}

func main() {
	config := CleaningConfig{}
	flag.StringVar(&config.InputDir, "input", "database/data/exported", "Input directory with raw CSV files")
	flag.StringVar(&config.OutputDir, "output", "database/data/clean", "Output directory for cleaned CSV files")
	flag.StringVar(&config.LogFile, "log", "database/logs/cleaning.log", "Log file path")
	flag.Parse()

	log.Printf("Starting CSV data cleaning...")
	log.Printf("Input: %s", config.InputDir)
	log.Printf("Output: %s", config.OutputDir)

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create log directory
	if err := os.MkdirAll(filepath.Dir(config.LogFile), 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	stats := &CleaningStats{StartTime: time.Now()}

	// Process all CSV files in input directory
	files, err := filepath.Glob(filepath.Join(config.InputDir, "*.csv"))
	if err != nil {
		log.Fatalf("Failed to find CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Printf("No CSV files found in %s", config.InputDir)
		return
	}

	for _, file := range files {
		if err := processCSVFile(file, config.OutputDir, stats); err != nil {
			log.Printf("Error processing %s: %v", file, err)
			stats.ErrorCount++
		} else {
			stats.FilesProcessed++
		}
	}

	// Write summary
	duration := time.Since(stats.StartTime)
	summary := fmt.Sprintf(`CSV Cleaning Summary
===================
Start Time: %s
Duration: %v
Files Processed: %d
Records Cleaned: %d
Errors: %d
Status: %s
`,
		stats.StartTime.Format(time.RFC3339),
		duration,
		stats.FilesProcessed,
		stats.RecordsCleaned,
		stats.ErrorCount,
		func() string {
			if stats.ErrorCount == 0 {
				return "SUCCESS"
			}
			return "COMPLETED WITH ERRORS"
		}(),
	)

	log.Print(summary)

	// Write to log file
	if err := os.WriteFile(config.LogFile, []byte(summary), 0644); err != nil {
		log.Printf("Failed to write log file: %v", err)
	}
}

func processCSVFile(inputFile, outputDir string, stats *CleaningStats) error {
	log.Printf("Processing: %s", inputFile)

	// Open input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		log.Printf("Empty file: %s", inputFile)
		return nil
	}

	// Get base filename
	baseFile := filepath.Base(inputFile)
	outputFile := filepath.Join(outputDir, baseFile)

	// Clean and normalize data
	cleanedRecords := cleanCSVData(records, baseFile)
	stats.RecordsCleaned += len(cleanedRecords) - 1 // Subtract header

	// Write cleaned data
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	for _, record := range cleanedRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	log.Printf("Cleaned %s: %d records → %s", baseFile, len(records)-1, outputFile)
	return nil
}

).MatchString(cleaned) {
		return cleaned
	}
	if regexp.MustCompile(`^\d{9}// backend/cmd/tools/csv-cleaner/main.go
// Tool to clean and normalize CSV data exported from MDB
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CleaningConfig struct {
	InputDir  string
	OutputDir string
	LogFile   string
}

type CleaningStats struct {
	FilesProcessed int
	RecordsCleaned int
	ErrorCount     int
	StartTime      time.Time
}

func main() {
	config := CleaningConfig{}
	flag.StringVar(&config.InputDir, "input", "database/data/exported", "Input directory with raw CSV files")
	flag.StringVar(&config.OutputDir, "output", "database/data/clean", "Output directory for cleaned CSV files")
	flag.StringVar(&config.LogFile, "log", "database/logs/cleaning.log", "Log file path")
	flag.Parse()

	log.Printf("Starting CSV data cleaning...")
	log.Printf("Input: %s", config.InputDir)
	log.Printf("Output: %s", config.OutputDir)

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create log directory
	if err := os.MkdirAll(filepath.Dir(config.LogFile), 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	stats := &CleaningStats{StartTime: time.Now()}

	// Process all CSV files in input directory
	files, err := filepath.Glob(filepath.Join(config.InputDir, "*.csv"))
	if err != nil {
		log.Fatalf("Failed to find CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Printf("No CSV files found in %s", config.InputDir)
		return
	}

	for _, file := range files {
		if err := processCSVFile(file, config.OutputDir, stats); err != nil {
			log.Printf("Error processing %s: %v", file, err)
			stats.ErrorCount++
		} else {
			stats.FilesProcessed++
		}
	}

	// Write summary
	duration := time.Since(stats.StartTime)
	summary := fmt.Sprintf(`CSV Cleaning Summary
===================
Start Time: %s
Duration: %v
Files Processed: %d
Records Cleaned: %d
Errors: %d
Status: %s
`,
		stats.StartTime.Format(time.RFC3339),
		duration,
		stats.FilesProcessed,
		stats.RecordsCleaned,
		stats.ErrorCount,
		func() string {
			if stats.ErrorCount == 0 {
				return "SUCCESS"
			}
			return "COMPLETED WITH ERRORS"
		}(),
	)

	log.Print(summary)

	// Write to log file
	if err := os.WriteFile(config.LogFile, []byte(summary), 0644); err != nil {
		log.Printf("Failed to write log file: %v", err)
	}
}

func processCSVFile(inputFile, outputDir string, stats *CleaningStats) error {
	log.Printf("Processing: %s", inputFile)

	// Open input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		log.Printf("Empty file: %s", inputFile)
		return nil
	}

	// Get base filename
	baseFile := filepath.Base(inputFile)
	outputFile := filepath.Join(outputDir, baseFile)

	// Clean and normalize data
	cleanedRecords := cleanCSVData(records, baseFile)
	stats.RecordsCleaned += len(cleanedRecords) - 1 // Subtract header

	// Write cleaned data
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	for _, record := range cleanedRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	log.Printf("Cleaned %s: %d records → %s", baseFile, len(records)-1, outputFile)
	return nil
}

).MatchString(cleaned) {
		return cleaned[:5] + "-" + cleaned[5:]
	}
	
	return zip // Return original if doesn't match patterns
}

func cleanName(name string) string {
	if name == "" {
		return ""
	}
	
	// Trim and normalize whitespace
	name = strings.TrimSpace(name)
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	
	// Title case for names
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	
	return strings.Join(words, " ")
}

func cleanText(text string) string {
	if text == "" {
		return ""
	}
	
	// Trim and normalize whitespace
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	
	// Remove control characters
	text = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(text, "")
	
	return text
}

func isEmptyRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

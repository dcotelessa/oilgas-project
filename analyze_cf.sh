#!/bin/bash

# ColdFusion to PostgreSQL Migration Analysis Script
# Analyzes CF queries, extracts MDB schema/data, generates PostgreSQL-ready CSVs

set -e

# Configuration
CF_DIR="${CF_DIR:-./coldfusion_files}"
MDB_FILE="${MDB_FILE:-./petros.mdb}"
OUTPUT_DIR="${OUTPUT_DIR:-./migration_output}"
SCHEMA_DIR="${OUTPUT_DIR}/schema"
DATA_DIR="${OUTPUT_DIR}/data"
ANALYSIS_DIR="${OUTPUT_DIR}/analysis"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check dependencies
check_dependencies() {
    log "Checking dependencies..."
    
    local deps=("mdb-export" "mdb-schema" "mdb-tables" "jq")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v $dep &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        error "Missing required dependencies: ${missing_deps[*]}"
        echo
        echo "Installation commands:"
        echo "  macOS: brew install mdbtools jq"
        echo "  Ubuntu/Debian: sudo apt-get install mdb-tools jq"
        exit 1
    fi
    
    # Check for Go (optional)
    if ! command -v go &> /dev/null; then
        warn "Go not found - ColdFusion analysis will be skipped"
        warn "Install Go from https://golang.org/dl/"
    fi
    
    log "Dependencies satisfied"
}

# Process CSV data with bash (convert dates, lowercase headers, clean data)
process_csv_data() {
    local input_file="$1"
    local output_file="$2"
    
    if [[ ! -f "$input_file" ]]; then
        error "Input file not found: $input_file"
        return 1
    fi
    
    # Read header line and convert to lowercase
    local header_line=$(head -n 1 "$input_file")
    local lowercase_header=$(echo "$header_line" | tr '[:upper:]' '[:lower:]')
    
    # Write lowercase header to output
    echo "$lowercase_header" > "$output_file"
    
    # Process data rows with awk for better CSV handling
    tail -n +2 "$input_file" | awk -F',' '
    BEGIN { OFS="," }
    {
        for (i = 1; i <= NF; i++) {
            # Remove quotes and trim whitespace
            gsub(/^"|"$/, "", $i)
            gsub(/^[ \t]+|[ \t]+$/, "", $i)
            
            # Check for date pattern MM/DD/YY or MM/DD/YYYY
            if ($i ~ /^[0-9]{1,2}\/[0-9]{1,2}\/[0-9]{2,4}$/) {
                split($i, date_parts, "/")
                month = date_parts[1]
                day = date_parts[2]
                year = date_parts[3]
                
                # Handle invalid dates
                if (month == "00" || day == "00") {
                    $i = ""
                } else {
                    # Convert 2-digit year to 4-digit
                    if (length(year) == 2) {
                        if (year < 50) {
                            year = "20" year
                        } else {
                            year = "19" year
                        }
                    }
                    
                    # Format as YYYY-MM-DD with zero padding
                    $i = sprintf("%04d-%02d-%02d", year, month, day)
                }
            }
            
            # Quote field if it contains commas or quotes
            if ($i ~ /[,"]/) {
                gsub(/"/, "\"\"", $i)  # Escape quotes
                $i = "\"" $i "\""
            }
        }
        print
    }' >> "$output_file"
}

# Extract MDB schema
extract_schema() {
    log "Extracting MDB schema..."
    
    if [[ ! -f "$MDB_FILE" ]]; then
        error "MDB file not found: $MDB_FILE"
        exit 1
    fi
    
    # Get table list
    mdb-tables "$MDB_FILE" | tr ' ' '\n' > "$SCHEMA_DIR/tables.txt"
    log "Found $(wc -l < "$SCHEMA_DIR/tables.txt") tables"
    
    # Extract schema for each table
    while IFS= read -r table; do
        [[ -z "$table" ]] && continue
        local table_lower=$(echo "$table" | tr '[:upper:]' '[:lower:]')
        log "Extracting schema for table: $table"
        mdb-schema "$MDB_FILE" -T "$table" --no-indexes > "$SCHEMA_DIR/${table_lower}.sql"
    done < "$SCHEMA_DIR/tables.txt"
}

# Convert Access schema to clean PostgreSQL
convert_access_to_postgresql() {
    local input_file="$1"
    local output_file="$2"
    
    if [[ ! -f "$input_file" ]]; then
        return 1
    fi
    
    # Process the Access SQL and convert to PostgreSQL using sed and bash
    local table_name=""
    local in_table=false
    local columns=""
    
    while IFS= read -r line; do
        # Skip MDB Tools headers
        if [[ "$line" =~ ^--.*MDB\ Tools ]] || [[ "$line" =~ ^--.*Copyright ]] || [[ "$line" =~ ^--.*Files\ in ]] || [[ "$line" =~ ^--.*Check\ out ]] || [[ "$line" =~ ^--.*encoding ]] || [[ "$line" =~ ^--\ ------ ]]; then
            continue
        fi
        
        # Skip empty lines
        if [[ -z "$line" ]]; then
            continue
        fi
        
        # Start of table definition
        if [[ "$line" =~ ^CREATE\ TABLE\ \[([^\]]+)\] ]]; then
            # Extract table name using bash regex
            if [[ "$line" =~ \[([^\]]+)\] ]]; then
                table_name=$(echo "${BASH_REMATCH[1]}" | tr '[:upper:]' '[:lower:]')
                in_table=true
                columns=""
                echo "-- Table: $table_name" >> "$output_file"
                echo "CREATE TABLE IF NOT EXISTS store.$table_name (" >> "$output_file"
            fi
            continue
        fi
        
        # Column definitions
        if [[ "$in_table" == true ]] && [[ "$line" =~ ^\	\[([^\]]+)\] ]]; then
            # Extract column name and type
            local col_def=$(echo "$line" | sed 's/^	//; s/,$//')
            if [[ "$col_def" =~ \[([^\]]+)\][[:space:]]+(.+) ]]; then
                local col_name=$(echo "${BASH_REMATCH[1]}" | tr '[:upper:]' '[:lower:]')
                local col_type="${BASH_REMATCH[2]}"
                local pg_type=""
                
                # Convert Access types to PostgreSQL
                if [[ "$col_type" =~ Long\ Integer ]] && [[ "$col_name" =~ ^(id|.*id)$ ]]; then
                    if [[ "$col_name" == "id" ]] || [[ "$col_name" == "custid" ]] || [[ "$col_name" == "userid" ]]; then
                        pg_type="SERIAL PRIMARY KEY"
                    else
                        pg_type="INTEGER"
                    fi
                elif [[ "$col_type" =~ Long\ Integer ]]; then
                    pg_type="INTEGER"
                elif [[ "$col_type" =~ Text\ \(([0-9]+)\) ]]; then
                    local size=$(echo "$col_type" | sed -n 's/.*Text (\([0-9]*\)).*/\1/p')
                    pg_type="VARCHAR($size)"
                elif [[ "$col_type" =~ Text ]]; then
                    pg_type="VARCHAR(255)"
                elif [[ "$col_type" =~ Memo/Hyperlink ]]; then
                    pg_type="TEXT"
                elif [[ "$col_type" =~ DateTime ]]; then
                    pg_type="TIMESTAMP"
                elif [[ "$col_type" =~ Boolean\ NOT\ NULL ]]; then
                    pg_type="BOOLEAN NOT NULL DEFAULT false"
                elif [[ "$col_type" =~ Boolean ]]; then
                    pg_type="BOOLEAN DEFAULT false"
                else
                    pg_type="VARCHAR(255)"
                fi
                
                # Add column to list
                if [[ -n "$columns" ]]; then
                    columns="$columns,\n"
                fi
                columns="$columns    $col_name $pg_type"
            fi
            continue
        fi
        
        # End of table definition
        if [[ "$in_table" == true ]] && [[ "$line" =~ ^\)\; ]]; then
            in_table=false
            # Add created_at timestamp if not a simple lookup table
            if [[ ! "$table_name" =~ ^(grade|rnumber|wknumber|test)$ ]]; then
                if [[ -n "$columns" ]]; then
                    columns="$columns,\n"
                fi
                columns="$columns    created_at TIMESTAMP DEFAULT NOW()"
            fi
            echo -e "$columns" >> "$output_file"
            echo ");" >> "$output_file"
            echo "" >> "$output_file"
        fi
    done < "$input_file"
}

# Generate PostgreSQL compatible schema
generate_postgresql_schema() {
    log "Generating PostgreSQL schema..."
    
    cat > "$SCHEMA_DIR/postgresql_schema.sql" << 'EOF'
-- PostgreSQL Schema for Oil & Gas Inventory System
-- Generated from ColdFusion/MDB migration

-- Create schema
CREATE SCHEMA IF NOT EXISTS store;

-- Set search path
SET search_path TO store, public;

EOF

    # Check if we have actual MDB tables to convert
    local has_mdb_tables=false
    if [[ -f "$SCHEMA_DIR/tables.txt" ]]; then
        while IFS= read -r table; do
            [[ -z "$table" ]] && continue
            local table_lower=$(echo "$table" | tr '[:upper:]' '[:lower:]')
            
            if [[ -f "$SCHEMA_DIR/${table_lower}.sql" ]]; then
                log "Converting Access schema for table: $table_lower"
                convert_access_to_postgresql "$SCHEMA_DIR/${table_lower}.sql" "$SCHEMA_DIR/postgresql_schema.sql"
                has_mdb_tables=true
            fi
        done < "$SCHEMA_DIR/tables.txt"
    fi
    
    # Add indexes and constraints
    if [[ "$has_mdb_tables" == "true" ]]; then
        cat >> "$SCHEMA_DIR/postgresql_schema.sql" << 'EOF'

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_customers_customer ON store.customers(customer);
CREATE INDEX IF NOT EXISTS idx_customers_custid ON store.customers(custid);
CREATE INDEX IF NOT EXISTS idx_inventory_custid ON store.inventory(custid);
CREATE INDEX IF NOT EXISTS idx_inventory_grade ON store.inventory(grade);
CREATE INDEX IF NOT EXISTS idx_inventory_wkorder ON store.inventory(wkorder);
CREATE INDEX IF NOT EXISTS idx_received_custid ON store.received(custid);
CREATE INDEX IF NOT EXISTS idx_received_daterecvd ON store.received(daterecvd);
CREATE INDEX IF NOT EXISTS idx_fletcher_custid ON store.fletcher(custid);
CREATE INDEX IF NOT EXISTS idx_grade_grade ON store.grade(grade);

-- Insert required oil & gas grades
INSERT INTO store.grade (grade) VALUES 
('J55'),
('JZ55'),
('L80'),
('N80'),
('P105'),
('P110')
ON CONFLICT (grade) DO NOTHING;

EOF
    else
        # Add sample schema if no MDB tables found
        cat >> "$SCHEMA_DIR/postgresql_schema.sql" << 'EOF'

-- Sample Oil & Gas Inventory Tables (no MDB file processed)

-- Grades table for oil & gas pipe grades
CREATE TABLE IF NOT EXISTS store.grade (
    id SERIAL PRIMARY KEY,
    grade VARCHAR(50) UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Customers table
CREATE TABLE IF NOT EXISTS store.customers (
    custid SERIAL PRIMARY KEY,
    customer VARCHAR(255) NOT NULL,
    billingaddress VARCHAR(255),
    billingcity VARCHAR(100),
    billingstate VARCHAR(50),
    billingzipcode VARCHAR(10),
    phone VARCHAR(20),
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Basic inventory table
CREATE TABLE IF NOT EXISTS store.inventory (
    id SERIAL PRIMARY KEY,
    wkorder VARCHAR(50),
    custid INTEGER,
    grade VARCHAR(50),
    joints INTEGER,
    size VARCHAR(50),
    weight VARCHAR(50),
    location VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert required oil & gas grades
INSERT INTO store.grade (grade, description) VALUES 
('J55', 'Standard grade steel casing'),
('JZ55', 'Enhanced J55 grade'),
('L80', 'Higher strength grade'),
('N80', 'Medium strength grade'),
('P105', 'High performance grade'),
('P110', 'Premium performance grade')
ON CONFLICT (grade) DO NOTHING;

EOF
    fi
    
    # Add validation check
    cat >> "$SCHEMA_DIR/postgresql_schema.sql" << 'EOF'
-- Validation: Ensure grade table contains required values
-- Expected grades: J55, JZ55, L80, N80, P105, P110

EOF
}

# Extract and clean data from MDB
extract_data() {
    log "Extracting data from MDB..."
    
    while IFS= read -r table; do
        [[ -z "$table" ]] && continue
        local table_lower=$(echo "$table" | tr '[:upper:]' '[:lower:]')
        
        log "Extracting data for table: $table_lower"
        
        # Export to CSV with proper handling using bash
        mdb-export "$MDB_FILE" "$table" > "$DATA_DIR/${table_lower}_raw.csv"
        
        # Process CSV with bash (date conversion and case handling)
        process_csv_data "$DATA_DIR/${table_lower}_raw.csv" "$DATA_DIR/${table_lower}.csv"
        
        # Remove raw file
        rm "$DATA_DIR/${table_lower}_raw.csv"
        
        log "Exported $(tail -n +2 "$DATA_DIR/${table_lower}.csv" | wc -l) rows for $table_lower"
    done < "$SCHEMA_DIR/tables.txt"
}

# Create empty CF analysis file
create_empty_cf_analysis() {
    cat > "$ANALYSIS_DIR/queries.json" << 'EOF'
{
  "total_queries": 0,
  "total_files": 0,
  "queried_files": 0,
  "queries": [],
  "table_usage": [],
  "complex_queries": [],
  "recommendations": ["ColdFusion analysis skipped"]
}
EOF
}

# Analyze ColdFusion queries
analyze_cf_queries() {
    log "Analyzing ColdFusion queries..."
    
    if [[ ! -d "$CF_DIR" ]]; then
        warn "ColdFusion directory not found: $CF_DIR - skipping CF analysis"
        warn "Set CF_DIR environment variable to analyze ColdFusion files"
        create_empty_cf_analysis
        return
    fi
    
    if ! command -v go &> /dev/null; then
        warn "Go not found - skipping ColdFusion analysis"
        create_empty_cf_analysis
        return
    fi
    
    # Check if we should skip CF analysis
    if [[ "${SKIP_CF_ANALYSIS:-false}" == "true" ]]; then
        log "Skipping ColdFusion analysis (SKIP_CF_ANALYSIS=true)"
        create_empty_cf_analysis
        return
    fi
    
    # Find all CF files
    find "$CF_DIR" -type f \( -name "*.cfm" -o -name "*.cfc" \) > "$ANALYSIS_DIR/cf_files.txt"
    
    local cf_count=$(wc -l < "$ANALYSIS_DIR/cf_files.txt")
    log "Found $cf_count ColdFusion files"
    
    if [[ $cf_count -eq 0 ]]; then
        warn "No ColdFusion files found in $CF_DIR"
        create_empty_cf_analysis
        return
    fi
    
    # Generate Go-based CF analyzer
    cat > "$ANALYSIS_DIR/cf_analyzer.go" << 'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"crypto/md5"
	"encoding/hex"
)

type Query struct {
	File           string   `json:"file"`
	QueryName      string   `json:"query_name"`
	Datasource     string   `json:"datasource"`
	SQL            string   `json:"sql"`
	SQLHash        string   `json:"sql_hash"`
	Tables         []string `json:"tables"`
	Operations     []string `json:"operations"`
	HasJoins       bool     `json:"has_joins"`
	HasWhere       bool     `json:"has_where"`
	HasOrderBy     bool     `json:"has_order_by"`
	HasGroupBy     bool     `json:"has_group_by"`
	ComplexityScore int     `json:"complexity_score"`
}

type TableUsage struct {
	TableName   string   `json:"table_name"`
	QueryCount  int      `json:"query_count"`
	FileCount   int      `json:"file_count"`
	Files       []string `json:"files"`
	Operations  []string `json:"operations"`
}

type Analysis struct {
	TotalQueries    int          `json:"total_queries"`
	TotalFiles      int          `json:"total_files"`
	QueriedFiles    int          `json:"queried_files"`
	Queries         []Query      `json:"queries"`
	TableUsage      []TableUsage `json:"table_usage"`
	ComplexQueries  []Query      `json:"complex_queries"`
	Recommendations []string     `json:"recommendations"`
}

var (
	cfQueryPattern = regexp.MustCompile(`(?is)<cfquery\s+([^>]*?)>(.*?)</cfquery>`)
	namePattern    = regexp.MustCompile(`(?i)name\s*=\s*["']([^"']*)["']`)
	dsPattern      = regexp.MustCompile(`(?i)datasource\s*=\s*["']([^"']*)["']`)
	tablePattern   = regexp.MustCompile(`(?i)\b(?:FROM|JOIN|INTO|UPDATE)\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	joinPattern    = regexp.MustCompile(`(?i)\b(?:INNER|LEFT|RIGHT|FULL|CROSS)?\s*JOIN\b`)
	wherePattern   = regexp.MustCompile(`(?i)\bWHERE\b`)
	orderPattern   = regexp.MustCompile(`(?i)\bORDER\s+BY\b`)
	groupPattern   = regexp.MustCompile(`(?i)\bGROUP\s+BY\b`)
	cfTagPattern   = regexp.MustCompile(`<[^>]*>`)
	cfVarPattern   = regexp.MustCompile(`#[^#]*#`)
	spacePattern   = regexp.MustCompile(`\s+`)
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: cf_analyzer <cf_directory> [output_file]")
	}

	cfDir := os.Args[1]
	outputFile := "cf_analysis.json"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	analysis, err := analyzeCFDirectory(cfDir)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	// Write JSON output
	jsonData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	// Print summary
	printSummary(analysis)
	fmt.Printf("\nDetailed analysis saved to: %s\n", outputFile)
}

func analyzeCFDirectory(cfDir string) (*Analysis, error) {
	var allQueries []Query
	fileCount := 0
	queriedFiles := 0
	
	// Check if directory exists using os.Stat
	info, err := os.Stat(cfDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Analysis{
				TotalQueries:    0,
				TotalFiles:      0,
				QueriedFiles:    0,
				Queries:         []Query{},
				TableUsage:      []TableUsage{},
				ComplexQueries:  []Query{},
				Recommendations: []string{"ColdFusion directory not found"},
			}, nil
		}
		return nil, fmt.Errorf("cannot access directory %s: %w", cfDir, err)
	}
	
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", cfDir)
	}
	
	// Use filepath.Walk for better compatibility
	err = filepath.Walk(cfDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Warning: Skipping %s: %v\n", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".cfm" && ext != ".cfc" {
			return nil
		}

		fileCount++
		queries, err := extractQueriesFromFile(path)
		if err != nil {
			fmt.Printf("Warning: Failed to process %s: %v\n", path, err)
			return nil
		}

		if len(queries) > 0 {
			queriedFiles++
			allQueries = append(allQueries, queries...)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", cfDir, err)
	}

	// Analyze table usage
	tableUsage := analyzeTableUsage(allQueries)
	
	// Find complex queries
	var complexQueries []Query
	for _, query := range allQueries {
		if query.ComplexityScore > 5 {
			complexQueries = append(complexQueries, query)
		}
	}

	// Generate recommendations
	recommendations := generateRecommendations(allQueries, tableUsage)

	return &Analysis{
		TotalQueries:    len(allQueries),
		TotalFiles:      fileCount,
		QueriedFiles:    queriedFiles,
		Queries:         allQueries,
		TableUsage:      tableUsage,
		ComplexQueries:  complexQueries,
		Recommendations: recommendations,
	}, nil
}

func extractQueriesFromFile(filePath string) ([]Query, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var queries []Query
	matches := cfQueryPattern.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		attributes := match[1]
		sqlContent := match[2]

		// Extract query name
		nameMatch := namePattern.FindStringSubmatch(attributes)
		queryName := "unnamed_query"
		if len(nameMatch) > 1 {
			queryName = nameMatch[1]
		}

		// Extract datasource
		dsMatch := dsPattern.FindStringSubmatch(attributes)
		datasource := "default"
		if len(dsMatch) > 1 {
			datasource = dsMatch[1]
		}

		// Clean SQL
		sqlClean := cfTagPattern.ReplaceAllString(sqlContent, "")
		sqlClean = cfVarPattern.ReplaceAllString(sqlClean, "?")
		sqlClean = spacePattern.ReplaceAllString(sqlClean, " ")
		sqlClean = strings.TrimSpace(sqlClean)

		// Generate hash
		hash := md5.Sum([]byte(sqlClean))
		sqlHash := hex.EncodeToString(hash[:])[:8]

		// Extract tables
		tables := extractTables(sqlClean)
		
		// Extract operations
		operations := extractOperations(sqlClean)
		
		// Analyze patterns
		hasJoins := joinPattern.MatchString(sqlClean)
		hasWhere := wherePattern.MatchString(sqlClean)
		hasOrderBy := orderPattern.MatchString(sqlClean)
		hasGroupBy := groupPattern.MatchString(sqlClean)

		// Calculate complexity
		complexity := len(tables)
		if hasJoins { complexity += 2 }
		if hasWhere { complexity += 1 }
		if hasOrderBy { complexity += 1 }
		if hasGroupBy { complexity += 2 }

		// Truncate SQL for output
		displaySQL := sqlClean
		if len(displaySQL) > 500 {
			displaySQL = displaySQL[:500] + "..."
		}

		query := Query{
			File:           filePath,
			QueryName:      queryName,
			Datasource:     datasource,
			SQL:            displaySQL,
			SQLHash:        sqlHash,
			Tables:         tables,
			Operations:     operations,
			HasJoins:       hasJoins,
			HasWhere:       hasWhere,
			HasOrderBy:     hasOrderBy,
			HasGroupBy:     hasGroupBy,
			ComplexityScore: complexity,
		}

		queries = append(queries, query)
	}

	return queries, nil
}

func extractTables(sql string) []string {
	matches := tablePattern.FindAllStringSubmatch(sql, -1)
	seen := make(map[string]bool)
	var tables []string

	for _, match := range matches {
		if len(match) > 1 {
			table := strings.ToLower(strings.TrimSpace(match[1]))
			if table != "" && !seen[table] {
				tables = append(tables, table)
				seen[table] = true
			}
		}
	}

	return tables
}

func extractOperations(sql string) []string {
	var operations []string
	sqlUpper := strings.ToUpper(sql)

	if strings.Contains(sqlUpper, "SELECT") {
		operations = append(operations, "SELECT")
	}
	if strings.Contains(sqlUpper, "INSERT") {
		operations = append(operations, "INSERT")
	}
	if strings.Contains(sqlUpper, "UPDATE") {
		operations = append(operations, "UPDATE")
	}
	if strings.Contains(sqlUpper, "DELETE") {
		operations = append(operations, "DELETE")
	}

	return operations
}

func analyzeTableUsage(queries []Query) []TableUsage {
	usage := make(map[string]*TableUsage)

	for _, query := range queries {
		for _, table := range query.Tables {
			if _, exists := usage[table]; !exists {
				usage[table] = &TableUsage{
					TableName:  table,
					Files:      []string{},
					Operations: []string{},
				}
			}

			u := usage[table]
			u.QueryCount++
			
			// Add file if not already present
			fileExists := false
			for _, f := range u.Files {
				if f == query.File {
					fileExists = true
					break
				}
			}
			if !fileExists {
				u.Files = append(u.Files, query.File)
			}

			// Add operations if not already present
			for _, op := range query.Operations {
				opExists := false
				for _, existingOp := range u.Operations {
					if existingOp == op {
						opExists = true
						break
					}
				}
				if !opExists {
					u.Operations = append(u.Operations, op)
				}
			}
		}
	}

	// Convert to slice and calculate file counts
	var result []TableUsage
	for _, u := range usage {
		u.FileCount = len(u.Files)
		result = append(result, *u)
	}

	// Sort by query count
	sort.Slice(result, func(i, j int) bool {
		return result[i].QueryCount > result[j].QueryCount
	})

	return result
}

func generateRecommendations(queries []Query, tableUsage []TableUsage) []string {
	var recommendations []string

	// Check for queries without WHERE clauses
	noWhereCount := 0
	for _, query := range queries {
		if !query.HasWhere && contains(query.Operations, "SELECT") {
			noWhereCount++
		}
	}
	if noWhereCount > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("⚠️  Found %d SELECT queries without WHERE clauses - review for performance", noWhereCount))
	}

	// Check for complex queries
	complexCount := 0
	for _, query := range queries {
		if query.ComplexityScore > 7 {
			complexCount++
		}
	}
	if complexCount > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("⚠️  Found %d highly complex queries (score > 7) - consider optimization", complexCount))
	}

	// Index recommendations for high-usage tables
	if len(tableUsage) > 0 {
		var highUsageTables []string
		for i, usage := range tableUsage {
			if i >= 5 { break } // Top 5 tables
			if usage.QueryCount > 10 {
				highUsageTables = append(highUsageTables, usage.TableName)
			}
		}
		if len(highUsageTables) > 0 {
			recommendations = append(recommendations, 
				fmt.Sprintf("💡 Consider indexing high-usage tables: %s", strings.Join(highUsageTables, ", ")))
		}
	}

	// Migration-specific recommendations
	recommendations = append(recommendations, "🔄 Convert all table/column names to lowercase for PostgreSQL compatibility")
	recommendations = append(recommendations, "📅 Review date handling - ColdFusion date formats may need conversion")
	recommendations = append(recommendations, "🔒 Implement prepared statements in Go for security")

	return recommendations
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func printSummary(analysis *Analysis) {
	fmt.Printf("\n=== ColdFusion Analysis Summary ===\n")
	fmt.Printf("Total Files Scanned: %d\n", analysis.TotalFiles)
	fmt.Printf("Files with Queries: %d\n", analysis.QueriedFiles)
	fmt.Printf("Total Queries Found: %d\n", analysis.TotalQueries)
	fmt.Printf("Complex Queries (score > 5): %d\n", len(analysis.ComplexQueries))

	if len(analysis.TableUsage) > 0 {
		fmt.Printf("\n=== Top 10 Most Queried Tables ===\n")
		for i, usage := range analysis.TableUsage {
			if i >= 10 { break }
			fmt.Printf("%2d. %-20s %d queries in %d files\n", 
				i+1, usage.TableName, usage.QueryCount, usage.FileCount)
		}
	}

	fmt.Printf("\n=== Recommendations ===\n")
	for _, rec := range analysis.Recommendations {
		fmt.Printf("  %s\n", rec)
	}
}
EOF

    # Build and run the analyzer
    log "Building ColdFusion analyzer..."
    cd "$ANALYSIS_DIR"
    
    # Initialize Go module if needed
    if [[ ! -f "go.mod" ]]; then
        go mod init cf-analyzer 2>/dev/null || true
    fi
    
    # Build the analyzer
    if go build -o cf_analyzer cf_analyzer.go 2>/dev/null; then
        log "Running ColdFusion analysis..."
        
        # Run analyzer with error handling
        if ./cf_analyzer "$CF_DIR" queries.json 2>/dev/null; then
            log "✅ ColdFusion analysis completed successfully"
        else
            warn "❌ ColdFusion analysis failed - creating empty analysis file"
            create_empty_cf_analysis
        fi
    else
        warn "❌ Failed to build ColdFusion analyzer - skipping CF analysis"
        create_empty_cf_analysis
    fi
    
    cd - > /dev/null
    log "ColdFusion analysis stage complete"
}


# Generate migrations and seed data
generate_migrations() {
    log "Generating seed data..."
    
    # Add the converted schema if tables exist
    if [[ -f "$SCHEMA_DIR/postgresql_schema.sql" ]]; then
        log "Adding converted PostgreSQL schema to migration"
        
        # Skip the header comments and add the actual schema
        sed '/^-- Set search_path/,$!d' "$SCHEMA_DIR/postgresql_schema.sql" >> "$MIGRATIONS_DIR/001_initial_schema.sql"
    fi

    # Generate seed data
    generate_seed_data_production "$SEEDS_DIR"
}

# Generate production seed data (from real data)
generate_seed_data_production() {
    local seeds_dir="$1"
    log "Generating production seed data from real data..."
    
    cat > "$seeds_dir/production_seeds.sql" << 'EOF'
-- Production environment seed data
-- Real data from MDB migration

SET search_path TO store, public;

EOF

    # For each table, generate INSERT statements from CSV data
    if [[ -f "$SCHEMA_DIR/tables.txt" ]]; then
        while IFS= read -r table; do
            [[ -z "$table" ]] && continue
            local table_lower=$(echo "$table" | tr '[:upper:]' '[:lower:]')
            
            if [[ -f "$DATA_DIR/${table_lower}.csv" ]]; then
                log "Generating production seed data for: $table_lower"
                
                # Convert CSV to INSERT statements
                cat >> "$seeds_dir/production_seeds.sql" << EOF
-- Data for $table_lower
\\echo 'Seeding $table_lower...'
\\copy store.$table_lower FROM 'data/${table_lower}.csv' WITH (FORMAT CSV, HEADER true, NULL '');

EOF
            fi
        done < "$SCHEMA_DIR/tables.txt"
    fi
    
    # Add validation
    cat >> "$seeds_dir/production_seeds.sql" << 'EOF'

-- Validate critical data
\echo 'Validating production seed data...'

-- Check grade table
DO $
DECLARE
    expected_grades TEXT[] := ARRAY['J55', 'JZ55', 'L80', 'N80', 'P105', 'P110'];
    missing_count INTEGER := 0;
BEGIN
    FOR i IN 1..array_length(expected_grades, 1) LOOP
        IF NOT EXISTS (SELECT 1 FROM store.grade WHERE UPPER(grade_name) = expected_grades[i]) THEN
            missing_count := missing_count + 1;
            RAISE WARNING 'Missing grade: %', expected_grades[i];
        END IF;
    END LOOP;
    
    IF missing_count = 0 THEN
        RAISE NOTICE '✅ All expected grades found';
    ELSE
        RAISE WARNING '❌ Missing % grades', missing_count;
    END IF;
END $;

-- Update statistics
ANALYZE;
EOF
}


# Main execution
main() {
    log "Starting ColdFusion to PostgreSQL migration analysis..."
    
    check_dependencies
    extract_schema
    generate_postgresql_schema
    extract_data
    analyze_cf_queries
    generate_migrations
    
    log "Migration system setup complete!"
    echo
    echo "Usage:"
    echo "1. Set CF_DIR and MDB_FILE if not in current directory"
    echo "2. Set up environment: cp .env.local .env"
    echo "3. Start database: make dev-start"
    echo "4. Build migrator: make build"
    echo "5. Run migrations: make migrate"
    echo "6. Seed data: make seed"
    echo "7. Check status: make status"
    echo
    echo "Analysis Results:"
    if [[ -f "$ANALYSIS_DIR/queries.json" ]]; then
        if command -v jq &> /dev/null; then
            local query_count=$(jq '.total_queries // 0' "$ANALYSIS_DIR/queries.json" 2>/dev/null || echo "0")
            echo "  📊 Found $query_count ColdFusion queries"
        else
            echo "  📊 ColdFusion analysis file created (install jq for details)"
        fi
    else
        echo "  📊 No ColdFusion analysis performed"
    fi
    if [[ -f "$SCHEMA_DIR/tables.txt" ]]; then
        local table_count=$(wc -l < "$SCHEMA_DIR/tables.txt" 2>/dev/null || echo "0")
        echo "  🗄️  Migrated $table_count database tables"
    else
        echo "  🗄️  No database tables found"
    fi
}

# Run main function
main "$@"


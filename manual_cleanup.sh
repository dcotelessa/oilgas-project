#!/bin/bash
# manual_cleanup.sh - Simple manual cleanup steps

echo "🧹 Manual Repository Cleanup for Customer Migration"
echo "=================================================="
echo ""

# 1. Remove unnecessary backend tools
echo "1. Cleaning backend/cmd/tools..."
if [ -f "backend/cmd/tools/main.go" ]; then
    rm backend/cmd/tools/main.go
    echo "   ✅ Removed backend/cmd/tools/main.go"
fi

# 2. Remove tools directory if it exists
if [ -d "tools/" ]; then
    rm -rf tools/
    echo "   ✅ Removed tools/ directory"
fi

# 3. Remove unnecessary scripts
echo "2. Cleaning scripts..."
if [ -d "backend/scripts/" ]; then
    rm -rf backend/scripts/
    echo "   ✅ Removed backend/scripts/"
fi

if [ -d "scripts/" ]; then
    rm -rf scripts/
    echo "   ✅ Removed scripts/"
fi

# 4. Clean up docs
echo "3. Cleaning documentation..."
if [ -f "docs/README_PHASE1.md" ]; then
    rm docs/README_PHASE1.md
    echo "   ✅ Removed docs/README_PHASE1.md"
fi

# 5. Remove unnecessary config files
echo "4. Removing unnecessary files..."
UNNECESSARY_FILES=(
    ".env.local"
    "frontend/README.md"
    "setup_phase1.sh"
)

for file in "${UNNECESSARY_FILES[@]}"; do
    if [ -f "$file" ]; then
        rm "$file"
        echo "   ✅ Removed $file"
    fi
done

# 6. Create essential directories
echo "5. Creating essential directories..."
mkdir -p backend/cmd/tools/customer-analyzer
mkdir -p backend/cmd/customer-cleaner  
mkdir -p backend/cmd/standardized-importer
mkdir -p database/{data/{exported,clean},logs,migrations,init}
mkdir -p db_prep
mkdir -p docs
echo "   ✅ Created essential directories"

# 7. Create customer analyzer tool
echo "6. Creating customer analyzer tool..."
cat > backend/cmd/tools/customer-analyzer/main.go << 'EOF'
// backend/cmd/tools/customer-analyzer/main.go
// Analyzes customer CSV structure from MDB export
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: customer-analyzer <customers.csv>")
	}
	
	csvFile := os.Args[1]
	
	file, err := os.Open(csvFile)
	if err != nil {
		log.Fatal("Failed to open CSV:", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Failed to read CSV:", err)
	}
	
	if len(records) == 0 {
		log.Fatal("Empty CSV file")
	}
	
	headers := records[0]
	dataRow := []string{}
	if len(records) > 1 {
		dataRow = records[1]
	}
	
	fmt.Printf("📊 Customer CSV Analysis\n")
	fmt.Printf("File: %s\n", csvFile)
	fmt.Printf("Rows: %d (including header)\n", len(records))
	fmt.Printf("Columns: %d\n", len(headers))
	fmt.Printf("\n")
	
	fmt.Printf("Column Structure:\n")
	fmt.Printf("%-3s %-25s %-15s %s\n", "#", "Column Name", "Sample Data", "Standardized Name")
	fmt.Printf("%s\n", strings.Repeat("-", 70))
	
	for i, header := range headers {
		sample := ""
		if i < len(dataRow) {
			sample = dataRow[i]
			if len(sample) > 12 {
				sample = sample[:12] + "..."
			}
		}
		
		standardized := standardizeColumnName(header)
		fmt.Printf("%-3d %-25s %-15s %s\n", i+1, header, sample, standardized)
	}
}

func standardizeColumnName(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	
	mappings := map[string]string{
		"custid":         "customer_id",
		"customer":       "customer_name", 
		"billingaddress": "billing_address",
		"billingcity":    "billing_city",
		"billingstate":   "billing_state",
		"billingzipcode": "billing_zip_code",
		"contact":        "contact_name",
		"phone":          "phone_number",
		"fax":            "fax_number",
		"email":          "email_address",
		"color1":         "color_grade_1",
		"color2":         "color_grade_2",
		"color3":         "color_grade_3",
		"color4":         "color_grade_4",
		"color5":         "color_grade_5",
		"loss1":          "wall_loss_1",
		"loss2":          "wall_loss_2",
		"loss3":          "wall_loss_3",
		"loss4":          "wall_loss_4",
		"loss5":          "wall_loss_5",
		"wscolor1":       "wstring_color_1",
		"wscolor2":       "wstring_color_2",
		"wscolor3":       "wstring_color_3",
		"wscolor4":       "wstring_color_4",
		"wscolor5":       "wstring_color_5",
		"wsloss1":        "wstring_loss_1",
		"wsloss2":        "wstring_loss_2",
		"wsloss3":        "wstring_loss_3",
		"wsloss4":        "wstring_loss_4",
		"wsloss5":        "wstring_loss_5",
		"deleted":        "is_deleted",
	}
	
	if standardized, exists := mappings[lower]; exists {
		return standardized
	}
	
	return lower
}
EOF
echo "   ✅ Created customer analyzer tool"

# 8. Create streamlined documentation
echo "7. Creating streamlined documentation..."
cat > docs/CUSTOMER_MIGRATION.md << 'EOF'
# Customer Migration Quick Guide

## Setup
1. Place `petros-lb.mdb` in `db_prep/` folder
2. Run: `make setup-customers`
3. Verify: `make verify-customers`

## Commands
- `make check-mdb` - Verify MDB file
- `make analyze-mdb` - Export customers from Access
- `make setup-customers` - Complete workflow
- `make verify-customers` - Check results

## Files
- `db_prep/petros-lb.mdb` - Your Access database
- `customers.csv` - Exported data (auto-generated)
- PostgreSQL database with standardized schema
EOF

cat > docs/DATABASE_SCHEMA.md << 'EOF'
# Customer Database Schema

## Table: store.customers

Standardized customer table with no abbreviations:

- `customer_id` - Primary key (auto-generated)
- `original_customer_id` - Reference to MDB custid
- `customer_name` - Company name (required)
- `billing_address` - Full billing address
- `billing_city` - Billing city
- `billing_state` - 2-letter state code
- `billing_zip_code` - ZIP/postal code
- `contact_name` - Primary contact person
- `phone_number` - Formatted phone number
- `email_address` - Email address
- Color grades: `color_grade_1` through `color_grade_5`
- Wall losses: `wall_loss_1` through `wall_loss_5`
- W-String colors: `wstring_color_1` through `wstring_color_5`
- W-String losses: `wstring_loss_1` through `wstring_loss_5`
- `is_deleted` - Soft delete flag
- `tenant_id` - Multi-tenant isolation

## Features
- No abbreviations in column names
- Automatic duplicate detection
- Multi-tenant row-level security
- Data validation and cleaning
EOF
echo "   ✅ Created documentation"

# 9. Update .gitignore
echo "8. Updating .gitignore..."
cat > .gitignore << 'EOF'
# Database files
*.mdb
*.accdb
*.db

# Environment files  
.env
.env.production
.env.staging

# Build artifacts
*.exe
vendor/
node_modules/
dist/

# Generated data
customers.csv
customers_cleaned.csv
database/data/exported/*
database/data/clean/*
database/logs/*

# Tools
customer-cleaner
customer-analyzer
standardized-importer

# IDE
.vscode/
.idea/
*.swp
.DS_Store

# Docker
postgres_data/
pgadmin_data/
EOF
echo "   ✅ Updated .gitignore"

echo ""
echo "✅ Manual cleanup completed!"
echo ""
echo "📁 Your repository structure is now:"
echo "├── backend/cmd/tools/customer-analyzer/main.go"
echo "├── backend/cmd/customer-cleaner/                (create this)"
echo "├── backend/cmd/standardized-importer/           (create this)"  
echo "├── docs/CUSTOMER_MIGRATION.md"
echo "├── docs/DATABASE_SCHEMA.md"
echo "├── db_prep/                                     (place petros-lb.mdb here)"
echo "└── Makefile                                     (use optimized version)"
echo ""
echo "🚀 Next steps:"
echo "1. Copy petros-lb.mdb to db_prep/ folder"
echo "2. Replace Makefile with optimized version"
echo "3. Add customer-cleaner and standardized-importer tools"
echo "4. Run: make check-mdb"

#!/bin/bash
# cleanup_repository.sh - Clean up repository for optimized customer migration

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🧹 Repository Cleanup and Optimization${NC}"
echo -e "${BLUE}=====================================Cleanup and Optimization${NC}"
echo ""

print_status() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }

# =============================================================================
# 1. REMOVE UNNECESSARY TOOLS
# =============================================================================

echo -e "${YELLOW}📂 Cleaning up backend/cmd/tools folder...${NC}"

# Remove tools we don't need for customer setup
UNNECESSARY_TOOLS=(
    "backend/cmd/tools/cf_query_analyzer.go"
    "backend/cmd/tools/conversion_tester.go"
    "backend/cmd/tools/mdb_processor.go"
    "tools/"
)

for tool in "${UNNECESSARY_TOOLS[@]}"; do
    if [ -e "$tool" ]; then
        print_info "Removing unnecessary tool: $tool"
        rm -rf "$tool"
    fi
done

# Keep only essential tools for customer migration
ESSENTIAL_TOOLS_DIR="backend/cmd/tools"
mkdir -p "$ESSENTIAL_TOOLS_DIR"

# Create only the tools we need
cat > "$ESSENTIAL_TOOLS_DIR/customer-analyzer/main.go" << 'EOF'
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

print_status "Created essential customer analysis tool"

# =============================================================================
# 2. REMOVE UNNECESSARY SCRIPTS
# =============================================================================

echo -e "${YELLOW}📜 Cleaning up backend/scripts folder...${NC}"

# Remove scripts we don't need
UNNECESSARY_SCRIPTS=(
    "backend/scripts/utilities/"
    "backend/scripts/phase1_mdb_migration.sh"
    "scripts/"
)

for script in "${UNNECESSARY_SCRIPTS[@]}"; do
    if [ -e "$script" ]; then
        print_info "Removing unnecessary script: $script"
        rm -rf "$script"
    fi
done

print_status "Removed unnecessary scripts"

# =============================================================================
# 3. CLEAN UP DOCS FOLDER
# =============================================================================

echo -e "${YELLOW}📚 Cleaning up docs folder...${NC}"

# Remove outdated documentation
OUTDATED_DOCS=(
    "docs/README_PHASE1.md"
    "docs/README_PHASE2.md"
)

for doc in "${OUTDATED_DOCS[@]}"; do
    if [ -e "$doc" ]; then
        print_info "Removing outdated doc: $doc"
        rm -f "$doc"
    fi
done

# Create streamlined documentation
mkdir -p docs

cat > "docs/CUSTOMER_MIGRATION.md" << 'EOF'
# Customer Migration Guide

## Quick Start

### 1. Setup Database
```bash
make setup-db
```

### 2. Export from Access
```bash
# Export customers table to CSV
mdb-export db_prep/petros-lb.mdb customers > customers.csv
```

### 3. Analyze Structure  
```bash
make analyze-customers
```

### 4. Clean and Import
```bash
make import-customers
```

### 5. Verify
```bash
make verify-customers
```

## Files
- `db_prep/petros-lb.mdb` - Your Access database
- `customers.csv` - Exported customer data
- Database: PostgreSQL with standardized schema

## Standards Applied
- No abbreviations: `custid` → `customer_id`
- Deduplication: Automatic duplicate detection
- Multi-tenant: Tenant isolation with RLS
- Data cleaning: Phone, email, address standardization
EOF

cat > "docs/DATABASE_SCHEMA.md" << 'EOF'
# Database Schema

## Customer Table Structure

```sql
CREATE TABLE store.customers (
    customer_id SERIAL PRIMARY KEY,
    original_customer_id INTEGER,           -- custid from MDB
    customer_name VARCHAR(255) NOT NULL,    -- customer
    billing_address TEXT,                   -- billingaddress
    billing_city VARCHAR(100),             -- billingcity
    billing_state VARCHAR(2),              -- billingstate
    billing_zip_code VARCHAR(20),          -- billingzipcode
    contact_name VARCHAR(255),             -- contact
    phone_number VARCHAR(50),              -- phone
    fax_number VARCHAR(50),                -- fax
    email_address VARCHAR(255),            -- email
    
    -- CFM Color System
    color_grade_1 VARCHAR(50),             -- color1
    color_grade_2 VARCHAR(50),             -- color2
    color_grade_3 VARCHAR(50),             -- color3
    color_grade_4 VARCHAR(50),             -- color4
    color_grade_5 VARCHAR(50),             -- color5
    
    -- Loss Percentages
    wall_loss_1 DECIMAL(5,2),              -- loss1
    wall_loss_2 DECIMAL(5,2),              -- loss2
    wall_loss_3 DECIMAL(5,2),              -- loss3
    wall_loss_4 DECIMAL(5,2),              -- loss4
    wall_loss_5 DECIMAL(5,2),              -- loss5
    
    -- W-String System
    wstring_color_1 VARCHAR(50),           -- wscolor1
    wstring_color_2 VARCHAR(50),           -- wscolor2
    wstring_color_3 VARCHAR(50),           -- wscolor3
    wstring_color_4 VARCHAR(50),           -- wscolor4
    wstring_color_5 VARCHAR(50),           -- wscolor5
    wstring_loss_1 DECIMAL(5,2),           -- wsloss1
    wstring_loss_2 DECIMAL(5,2),           -- wsloss2
    wstring_loss_3 DECIMAL(5,2),           -- wsloss3
    wstring_loss_4 DECIMAL(5,2),           -- wsloss4
    wstring_loss_5 DECIMAL(5,2),           -- wsloss5
    
    is_deleted BOOLEAN DEFAULT false,       -- deleted
    tenant_id VARCHAR(50) NOT NULL,
    
    -- Deduplication fields
    normalized_company_name VARCHAR(255),
    address_hash VARCHAR(32),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Key Features
- **No abbreviations** in column names
- **Deduplication support** with normalized names
- **Multi-tenant isolation** with RLS
- **CFM color systems** preserved exactly as in Access
EOF

print_status "Created streamlined documentation"

# =============================================================================
# 4. REMOVE UNNECESSARY FILES
# =============================================================================

echo -e "${YELLOW}🗑️  Removing unnecessary files...${NC}"

UNNECESSARY_FILES=(
    ".env.local"
    "frontend/README.md"
    "setup_phase1.sh"
    "enhanced_customer_migration.sh"
    "complete_migration_workflow.sh"
)

for file in "${UNNECESSARY_FILES[@]}"; do
    if [ -e "$file" ]; then
        print_info "Removing unnecessary file: $file"
        rm -f "$file"
    fi
done

print_status "Removed unnecessary files"

# =============================================================================
# 5. ENSURE CLEAN DATABASE SETUP
# =============================================================================

echo -e "${YELLOW}🗄️  Setting up for petros-lb.mdb...${NC}"

# Ensure db_prep directory exists
mkdir -p db_prep

# Check for the MDB file
if [ ! -f "db_prep/petros-lb.mdb" ]; then
    print_warning "MDB file not found: db_prep/petros-lb.mdb"
    print_info "Please place your Access database at: db_prep/petros-lb.mdb"
else
    print_status "Found MDB file: db_prep/petros-lb.mdb"
fi

# Create database directories
mkdir -p database/{data/{exported,clean},logs,migrations,init}

print_status "Database directories ready"

# =============================================================================
# 6. UPDATE .gitignore
# =============================================================================

echo -e "${YELLOW}📝 Updating .gitignore...${NC}"

cat > ".gitignore" << 'EOF'
# ===== SENSITIVE DATA =====
.env
.env.production
.env.staging
*.mdb
*.accdb
real_*.csv
production_*.csv
backup_*.sql

# ===== BUILD ARTIFACTS =====
*.exe
*.dll
*.so
*.dylib
vendor/
node_modules/
dist/
build/

# ===== DEVELOPMENT =====
.vscode/
.idea/
*.swp
*.log
.DS_Store
Thumbs.db

# ===== GENERATED DATA =====
customers.csv
customers_cleaned.csv
database/data/exported/*
database/data/clean/*
database/logs/*

# ===== DOCKER =====
postgres_data/
pgadmin_data/

# ===== TOOLS =====
customer-cleaner
customer-analyzer
standardized-importer

# ===== SAFE FILES =====
!.env.example
!README.md
!docs/**
EOF

print_status "Updated .gitignore"

# =============================================================================
# 7. SUMMARY
# =============================================================================

echo ""
echo -e "${GREEN}🎉 Repository Cleanup Complete!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""

echo -e "${BLUE}📋 What was cleaned up:${NC}"
echo "✅ Removed unnecessary tools from backend/cmd/tools/"
echo "✅ Removed unused scripts from backend/scripts/"
echo "✅ Removed outdated documentation"
echo "✅ Removed unnecessary configuration files"
echo "✅ Updated .gitignore for clean repository"
echo ""

echo -e "${BLUE}📁 Current clean structure:${NC}"
echo "backend/cmd/tools/customer-analyzer/     # Essential customer analysis tool"
echo "docs/CUSTOMER_MIGRATION.md              # Streamlined migration guide"
echo "docs/DATABASE_SCHEMA.md                 # Schema documentation"
echo "db_prep/                                # Place petros-lb.mdb here"
echo ""

echo -e "${BLUE}🚀 Next steps:${NC}"
echo "1. Place your petros-lb.mdb file in db_prep/"
echo "2. Run: make analyze-mdb"
echo "3. Run: make setup-customers"
echo "4. Run: make import-customers"
echo ""

if [ ! -f "db_prep/petros-lb.mdb" ]; then
    echo -e "${YELLOW}⚠️  Don't forget to place petros-lb.mdb in db_prep/ folder!${NC}"
fi

echo -e "${GREEN}Repository is now optimized for customer migration! 🎯${NC}"

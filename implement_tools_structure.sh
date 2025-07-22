#!/bin/bash
# implement_tools_structure.sh
# Execute this script from your project root to implement Option 1 structure

set -euo pipefail

PROJECT_ROOT="$(pwd)"
TOOLS_DIR="$PROJECT_ROOT/tools"
MDB_CONVERSION_DIR="$TOOLS_DIR/mdb-conversion"

echo "🚀 Implementing Tools Directory Structure (Option 1)"
echo "Project Root: $PROJECT_ROOT"

# Verify we're in the right location
if [ ! -d "backend" ] && [ ! -f "scripts/phase1_mdb_migration.sh" ]; then
    echo "❌ Error: This doesn't appear to be your project root"
    echo "Expected to find 'backend/' directory or 'scripts/phase1_mdb_migration.sh'"
    echo "Please run this script from your project root directory"
    exit 1
fi

echo "✅ Project root verified"

# Create directory structure
create_directories() {
    echo ""
    echo "📁 Creating directory structure..."
    
    # Main tools directories
    mkdir -p "$MDB_CONVERSION_DIR"/{cmd,internal/{mapping,validation,testing,models},config,scripts,docs,test/{data,output}}
    mkdir -p "$TOOLS_DIR/scripts"
    mkdir -p "$PROJECT_ROOT/output/conversion"
    
    echo "✅ Created directories:"
    echo "  📂 $TOOLS_DIR/mdb-conversion/"
    echo "    📂 cmd/                    # Main conversion executables"
    echo "    📂 internal/               # Private packages"
    echo "    📂 config/                 # Configuration files"
    echo "    📂 scripts/                # Helper scripts"
    echo "    📂 docs/                   # Tool documentation"
    echo "    📂 test/                   # Test data and outputs"
    echo "  📂 $TOOLS_DIR/scripts/       # Utility scripts"
    echo "  📂 $PROJECT_ROOT/output/     # Shared output directory"
}

# Create Go module
create_go_module() {
    echo ""
    echo "🔧 Creating Go module for conversion tools..."
    
    cd "$MDB_CONVERSION_DIR"
    
    # Determine organization name from existing backend module if available
    ORG_NAME="dcotelessa"
    if [ -f "$PROJECT_ROOT/backend/go.mod" ]; then
        ORG_NAME=$(grep "^module" "$PROJECT_ROOT/backend/go.mod" | awk '{print $2}' | cut -d'/' -f1-2 || echo "dcotelessa")
    fi
    
    cat > go.mod << EOF
module $ORG_NAME/mdb-conversion-tools

go 1.21

require (
    github.com/lib/pq v1.10.9
)
EOF

    echo "✅ Go module created: $ORG_NAME/mdb-conversion-tools"
    cd "$PROJECT_ROOT"
}

# Create configuration files
create_config_files() {
    echo ""
    echo "⚙️  Creating configuration files..."
    
    # Oil & Gas mappings
    cat > "$MDB_CONVERSION_DIR/config/oil_gas_mappings.json" << 'EOF'
{
  "domain_mappings": {
    "work_order": {
      "patterns": ["wo", "wkorder", "workorder", "work_order"],
      "normalized": "work_order",
      "type": "VARCHAR(100)",
      "description": "Work order number or identifier"
    },
    "customer": {
      "patterns": ["custid", "customerid", "customer_id", "custname", "customername"],
      "normalized": "customer",
      "type": "VARCHAR(255)",
      "description": "Customer company name or ID"
    },
    "pipe_size": {
      "patterns": ["size", "pipesize", "diameter", "od"],
      "normalized": "size",
      "type": "VARCHAR(50)",
      "description": "Pipe outside diameter (e.g., 5 1/2\", 7\", 9 5/8\")"
    },
    "grade": {
      "patterns": ["grade", "steel_grade", "material_grade"],
      "normalized": "grade", 
      "type": "VARCHAR(10)",
      "description": "Steel grade (J55, L80, N80, P110, etc.)"
    },
    "connection": {
      "patterns": ["conn", "connection", "conntype", "thread"],
      "normalized": "connection",
      "type": "VARCHAR(100)",
      "description": "Thread connection type (BTC, LTC, STC, VAM, etc.)"
    },
    "weight": {
      "patterns": ["weight", "wt", "pipe_weight"],
      "normalized": "weight",
      "type": "DECIMAL(10,2)",
      "description": "Total weight in pounds"
    },
    "joints": {
      "patterns": ["joints", "joint_count", "pieces", "qty"],
      "normalized": "joints",
      "type": "INTEGER",
      "description": "Number of pipe joints"
    }
  },
  "transformation_rules": {
    "date_fields": {
      "patterns": ["date", "when"],
      "transforms": [
        "Replace empty strings with NULL",
        "Convert MM/DD/YYYY to YYYY-MM-DD",
        "Handle 'Now()' function calls"
      ]
    },
    "boolean_fields": {
      "patterns": ["deleted", "active", "complete", "required"],
      "transforms": [
        "Convert 'Yes'/'No' to true/false",
        "Convert 1/0 to true/false",
        "Default to false for NULL values"
      ]
    }
  }
}
EOF

    # Company template
    cat > "$MDB_CONVERSION_DIR/config/company_template.json" << 'EOF'
{
  "company_info": {
    "name": "",
    "industry": "oil_gas",
    "database_type": "access_mdb",
    "conversion_date": "",
    "contact": ""
  },
  "file_locations": {
    "mdb_files": [],
    "cf_application": "",
    "output_directory": ""
  },
  "custom_mappings": {
    "tables": {},
    "columns": {},
    "business_rules": []
  },
  "validation_rules": {
    "required_tables": ["customers", "inventory", "received"],
    "required_columns": ["customer", "work_order", "size", "grade"],
    "data_quality_checks": [
      "No duplicate work orders",
      "Valid customer references", 
      "Standard pipe sizes",
      "Valid steel grades"
    ]
  }
}
EOF

    echo "✅ Configuration files created"
}

# Create the tools Makefile
create_tools_makefile() {
    echo ""
    echo "🔧 Creating tools Makefile..."
    
    # Copy the comprehensive Makefile we created
    cat > "$MDB_CONVERSION_DIR/Makefile" << 'EOF'
# tools/mdb-conversion/Makefile
# MDB Conversion Tools - Separate from main application

.PHONY: help setup clean test build
.DEFAULT_GOAL := help

# Project paths
TOOLS_ROOT := $(shell pwd)
PROJECT_ROOT := $(shell cd ../.. && pwd)
OUTPUT_DIR := $(PROJECT_ROOT)/output
CONFIG_DIR := $(TOOLS_ROOT)/config
TEST_DIR := $(TOOLS_ROOT)/test
BUILD_DIR := $(TOOLS_ROOT)/build
BIN_DIR := $(BUILD_DIR)/bin

# Go settings
GO := go
GOFLAGS := -v
GOBUILD := $(GO) build $(GOFLAGS)
GOTEST := $(GO) test $(GOFLAGS)

# Database settings for testing
TEST_DATABASE_URL ?= postgres://postgres:password@localhost/oil_gas_test?sslmode=disable

help: ## Show available commands
	@echo "🛠️  MDB Conversion Tools"
	@echo "========================"
	@echo ""
	@echo "SETUP:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(setup|install|build)"
	@echo ""
	@echo "CONVERSION:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(convert|mdb|csv|sql)"
	@echo ""
	@echo "ANALYSIS:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(analyze|cf|query|extract)"
	@echo ""
	@echo "TESTING:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(test|validate|check)"
	@echo ""
	@echo "UTILITIES:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(clean|status|report)"
	@echo ""
	@echo "EXAMPLES:"
	@echo "  make convert-mdb FILE=database.mdb      # Convert MDB to CSV/SQL"
	@echo "  make analyze-cf DIR=./cf_app            # Analyze ColdFusion app"
	@echo "  make full-analysis COMPANY=acme_oil     # Complete analysis pipeline"
	@echo "  make test-conversion                    # Run all tests"

# =============================================================================
# SETUP COMMANDS
# =============================================================================

setup: setup-dirs install-deps build-tools verify-setup ## Complete setup
	@echo "✅ MDB conversion tools ready!"
	@echo "Next steps:"
	@echo "  1. make convert-mdb FILE=yourfile.mdb"
	@echo "  2. make analyze-cf DIR=your_cf_app"

setup-dirs: ## Create necessary directories
	@echo "📁 Creating directories..."
	@mkdir -p $(BUILD_DIR) $(BIN_DIR) $(OUTPUT_DIR)/conversion $(TEST_DIR)/output
	@echo "✅ Directories created"

install-deps: ## Install system dependencies
	@echo "🔧 Installing dependencies..."
	@if command -v brew >/dev/null 2>&1; then \
		echo "📦 Installing via Homebrew..."; \
		brew install mdb-tools || echo "⚠️  mdb-tools may already be installed"; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "📦 Installing via apt..."; \
		sudo apt-get update && sudo apt-get install -y mdb-tools || echo "⚠️  apt installation failed"; \
	else \
		echo "⚠️  Please install mdb-tools manually"; \
		echo "  macOS: brew install mdb-tools"; \
		echo "  Ubuntu/Debian: sudo apt-get install mdb-tools"; \
	fi
	@echo "✅ Dependencies installation complete"

build-tools: ## Build Go conversion tools
	@echo "🔨 Building conversion tools..."
	@$(GOBUILD) -o $(BIN_DIR)/mdb_processor ./cmd/mdb_processor.go
	@$(GOBUILD) -o $(BIN_DIR)/cf_query_analyzer ./cmd/cf_query_analyzer.go
	@$(GOBUILD) -o $(BIN_DIR)/conversion_tester ./cmd/conversion_tester.go
	@echo "✅ Tools built successfully"

verify-setup: ## Verify setup is working
	@echo "🔍 Verifying setup..."
	@if command -v mdb-ver >/dev/null 2>&1; then \
		echo "  ✅ mdb-tools: available"; \
	else \
		echo "  ❌ mdb-tools: not found"; \
	fi
	@for tool in mdb_processor cf_query_analyzer conversion_tester; do \
		if [ -f "$(BIN_DIR)/$$tool" ]; then \
			echo "  ✅ $$tool: built"; \
		else \
			echo "  ❌ $$tool: missing"; \
		fi; \
	done
	@if [ -f "$(CONFIG_DIR)/oil_gas_mappings.json" ]; then \
		echo "  ✅ Configuration: found"; \
	else \
		echo "  ❌ Configuration: missing"; \
	fi

# =============================================================================
# PLACEHOLDER COMMANDS (to be implemented)
# =============================================================================

convert-mdb: ## Convert MDB file (usage: make convert-mdb FILE=database.mdb [FORMAT=csv])
	@echo "🔄 MDB conversion placeholder"
	@echo "TODO: Implement full MDB conversion"
	@echo "FILE: $(FILE)"
	@echo "FORMAT: $(or $(FORMAT),csv)"

analyze-cf: ## Analyze ColdFusion app (usage: make analyze-cf DIR=cf_app)
	@echo "🔍 CF analysis placeholder"
	@echo "TODO: Implement ColdFusion analysis"
	@echo "DIR: $(DIR)"

test-conversion: ## Run conversion tool tests
	@echo "🧪 Test placeholder"
	@echo "TODO: Implement comprehensive testing"

clean: ## Clean build artifacts and outputs
	@echo "🧹 Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(TEST_DIR)/output/*
	@echo "✅ Cleanup complete"

status: ## Show tool status
	@echo "📊 MDB Conversion Tools Status"
	@echo "=============================="
	@echo "Tools Directory: $(TOOLS_ROOT)"
	@echo "Project Root:    $(PROJECT_ROOT)"
	@echo "Output Dir:      $(OUTPUT_DIR)"
	@echo ""
	@echo "🔧 System Tools:"
	@if command -v mdb-ver >/dev/null 2>&1; then \
		echo "  ✅ mdb-tools: installed"; \
	else \
		echo "  ❌ mdb-tools: missing"; \
	fi
	@if command -v go >/dev/null 2>&1; then \
		echo "  ✅ Go: $$(go version | awk '{print $$3}')"; \
	else \
		echo "  ❌ Go: missing"; \
	fi
EOF

    echo "✅ Tools Makefile created"
}

# Create placeholder Go files
create_placeholder_go_files() {
    echo ""
    echo "📄 Creating placeholder Go files..."
    
    # Main executables
    cat > "$MDB_CONVERSION_DIR/cmd/mdb_processor.go" << 'EOF'
package main

import (
    "fmt"
    "os"
)

// MDB Processor - Convert MDB files to CSV/SQL
// TODO: Implement the comprehensive MDB processor from earlier artifacts

func main() {
    if len(os.Args) < 2 {
        fmt.Println("MDB Processor v1.0")
        fmt.Println("Usage: mdb_processor <command> [args]")
        fmt.Println("Commands:")
        fmt.Println("  convert <mdb_file> <output_dir> <format>")
        fmt.Println("  analyze <mdb_file>")
        fmt.Println("  test")
        os.Exit(1)
    }
    
    command := os.Args[1]
    
    switch command {
    case "convert":
        fmt.Println("🔄 MDB conversion - placeholder")
        fmt.Println("TODO: Implement full conversion logic from artifacts")
    case "analyze":
        fmt.Println("🔍 MDB analysis - placeholder")
        fmt.Println("TODO: Implement structure analysis")
    case "test":
        fmt.Println("🧪 Running basic test...")
        fmt.Println("✅ MDB processor placeholder working")
    default:
        fmt.Printf("❌ Unknown command: %s\n", command)
        os.Exit(1)
    }
}
EOF

    cat > "$MDB_CONVERSION_DIR/cmd/cf_query_analyzer.go" << 'EOF'
package main

import (
    "fmt"
    "os"
)

// ColdFusion Query Analyzer
// TODO: Implement the comprehensive CF analyzer from earlier artifacts

func main() {
    if len(os.Args) < 2 {
        fmt.Println("ColdFusion Query Analyzer v1.0")
        fmt.Println("Usage: cf_query_analyzer <command> [args]")
        fmt.Println("Commands:")
        fmt.Println("  analyze <cf_directory> [output_dir]")
        fmt.Println("  extract <cf_directory>")
        fmt.Println("  test")
        os.Exit(1)
    }
    
    command := os.Args[1]
    
    switch command {
    case "analyze":
        fmt.Println("🔍 CF analysis - placeholder")
        fmt.Println("TODO: Implement ColdFusion query extraction from artifacts")
    case "extract":
        fmt.Println("📄 CF extraction - placeholder") 
        fmt.Println("TODO: Implement SQL query extraction")
    case "test":
        fmt.Println("🧪 Running basic test...")
        fmt.Println("✅ CF analyzer placeholder working")
    default:
        fmt.Printf("❌ Unknown command: %s\n", command)
        os.Exit(1)
    }
}
EOF

    cat > "$MDB_CONVERSION_DIR/cmd/conversion_tester.go" << 'EOF'
package main

import (
    "fmt"
    "os"
)

// Conversion Tester - Test suite for conversion tools
// TODO: Implement the comprehensive test framework from earlier artifacts

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Conversion Tester v1.0")
        fmt.Println("Usage: conversion_tester <command>")
        fmt.Println("Commands:")
        fmt.Println("  unit        - Run unit tests")
        fmt.Println("  integration - Run integration tests") 
        fmt.Println("  performance - Run performance tests")
        fmt.Println("  all         - Run all tests")
        os.Exit(1)
    }
    
    command := os.Args[1]
    
    switch command {
    case "unit":
        fmt.Println("🧪 Unit tests - placeholder")
        fmt.Println("TODO: Implement unit test suite from artifacts")
    case "integration":
        fmt.Println("🔗 Integration tests - placeholder")
        fmt.Println("TODO: Implement integration test suite")
    case "performance":
        fmt.Println("📊 Performance tests - placeholder")
        fmt.Println("TODO: Implement performance benchmarks")
    case "all":
        fmt.Println("🧪 Running all tests...")
        fmt.Println("✅ Conversion tester placeholder working")
    default:
        fmt.Printf("❌ Unknown command: %s\n", command)
        os.Exit(1)
    }
}
EOF

    # Internal packages
    mkdir -p "$MDB_CONVERSION_DIR/internal/mapping"
    cat > "$MDB_CONVERSION_DIR/internal/mapping/column_mapper.go" << 'EOF'
package mapping

// ColumnMapper handles column name normalization and mapping
// TODO: Implement the full ColumnMapping struct from earlier artifacts

type ColumnMapper struct {
    OilGasTerms   map[string]string
    CommonTerms   map[string]string
    DataTypes     map[string]string
}

func NewColumnMapper() *ColumnMapper {
    return &ColumnMapper{
        OilGasTerms: make(map[string]string),
        CommonTerms: make(map[string]string),
        DataTypes:   make(map[string]string),
    }
}

func (cm *ColumnMapper) NormalizeColumn(original string) string {
    // TODO: Implement full normalization logic from artifacts
    return original
}
EOF

    echo "✅ Placeholder Go files created"
}

# Create documentation
create_documentation() {
    echo ""
    echo "📚 Creating documentation..."
    
    cat > "$MDB_CONVERSION_DIR/README.md" << 'EOF'
# MDB Conversion Tools

Comprehensive tools for converting Microsoft Access (MDB) databases and analyzing ColdFusion applications for migration to PostgreSQL + Go systems.

## Quick Start

```bash
# Setup environment
make setup

# Convert MDB file
make convert-mdb FILE=database.mdb

# Analyze ColdFusion application  
make analyze-cf DIR=cf_application

# Run tests
make test-conversion
```

## Status: Implementation Phase

This is the initial structure setup. The comprehensive conversion tools from the artifacts need to be implemented to replace the current placeholders.

### TODO: Implementation Tasks

1. **Replace placeholder Go files** with full implementations from artifacts:
   - `cmd/mdb_processor.go` - Complete MDB conversion logic
   - `cmd/cf_query_analyzer.go` - ColdFusion analysis logic  
   - `cmd/conversion_tester.go` - Comprehensive test framework

2. **Implement internal packages**:
   - `internal/mapping/` - Column mapping and normalization
   - `internal/validation/` - Business rule validation
   - `internal/testing/` - Test utilities
   - `internal/models/` - Data structures

3. **Complete Makefile commands**:
   - All conversion commands are currently placeholders
   - Need to implement actual conversion logic

## Directory Structure

```
mdb-conversion/
├── cmd/                    # Executables (placeholders)
├── internal/               # Private packages (placeholders)
├── config/                 # Configuration files ✅
├── scripts/                # Helper scripts
├── test/                   # Test data
└── docs/                   # Documentation ✅
```

## Integration with Main Project

These tools are designed to work independently from your main application:

- **Separate Go module**: Independent versioning and dependencies
- **Dedicated Makefile**: Tool-specific commands
- **Shared output**: Results go to `../../output/conversion/`
- **Root integration**: Accessible via root Makefile shortcuts

## Next Steps

1. Implement the conversion tools using the comprehensive code from earlier artifacts
2. Test with sample MDB files
3. Integrate with your existing oil & gas backend system
4. Add company-specific validation rules
EOF

    cat > "$MDB_CONVERSION_DIR/docs/IMPLEMENTATION_GUIDE.md" << 'EOF'
# Implementation Guide

## Overview

This directory contains the structure for Option 1 (separate tools directory). The comprehensive conversion tools from the earlier artifacts need to be implemented to replace the current placeholders.

## Implementation Steps

### 1. Replace Placeholder Files

The current Go files in `cmd/` are placeholders. Replace them with the comprehensive implementations from the artifacts:

#### `cmd/mdb_processor.go`
- Implement the full MDB processor with column mapping
- Add oil & gas specific business rules
- Include CSV/SQL export capabilities
- Add error handling and progress tracking

#### `cmd/cf_query_analyzer.go`  
- Implement ColdFusion file parsing
- Add SQL query extraction logic
- Include complexity analysis
- Add reporting capabilities

#### `cmd/conversion_tester.go`
- Implement comprehensive test framework
- Add unit tests for column mapping
- Include integration tests for full workflow
- Add performance benchmarks

### 2. Implement Internal Packages

#### `internal/mapping/`
- Column name normalization
- Oil & gas specific mappings
- Data type inference
- Business rule validation

#### `internal/validation/`
- Data quality checks
- Business rule enforcement  
- Referential integrity validation
- Performance monitoring

#### `internal/testing/`
- Test utilities and frameworks
- Sample data generation
- Validation helpers
- Reporting tools

#### `internal/models/`
- Data structures for conversion jobs
- Table and column mappings
- Validation rules
- Configuration models

### 3. Complete Makefile Implementation

The current Makefile has placeholder commands. Implement:

- `convert-mdb`: Full MDB conversion pipeline
- `analyze-cf`: Complete ColdFusion analysis
- `test-conversion`: Comprehensive test suite
- Integration commands for backend import

### 4. Testing Strategy

1. **Unit Tests**: Test individual functions and mappings
2. **Integration Tests**: Test complete conversion workflows  
3. **Performance Tests**: Validate with large datasets
4. **Business Logic Tests**: Verify oil & gas specific rules

### 5. Configuration Management

The configuration files are in place:
- `config/oil_gas_mappings.json` - Industry-specific mappings
- `config/company_template.json` - Company-specific templates

Extend these as needed for specific business requirements.

## Quality Assurance

Before marking implementation complete:

1. ✅ All placeholder code replaced
2. ✅ Comprehensive test coverage
3. ✅ Documentation updated
4. ✅ Integration with main project tested
5. ✅ Performance benchmarks passed
6. ✅ Business stakeholder approval

## Integration Points

### With Main Application
- Import converted data via PostgreSQL
- Generate migrations for schema changes
- Validate data integrity after import

### With Root Makefile
- `make convert-mdb FILE=database.mdb` (from project root)
- `make analyze-cf DIR=cf_app` (from project root)
- `make tools` (show tools help)

## Success Criteria

1. **Conversion Accuracy**: >99.9% data fidelity
2. **Performance**: Handle 100k+ records efficiently
3. **Usability**: Simple commands for common tasks
4. **Maintainability**: Clear code structure and documentation
5. **Reusability**: Works for multiple company acquisitions
EOF

    echo "✅ Documentation created"
}

# Create test data
create_test_data() {
    echo ""
    echo "🧪 Creating test data..."
    
    cat > "$MDB_CONVERSION_DIR/test/data/sample_customers.csv" << 'EOF'
"CustID","CustName","BillAddr","BillCity","BillState","BillZip","Phone","Email"
1,"Permian Basin Energy","1234 Oil Field Rd","Midland","TX","79701","432-555-0101","ops@permianbasin.com"
2,"Eagle Ford Solutions","5678 Shale Ave","San Antonio","TX","78201","210-555-0201","drilling@eagleford.com"
3,"Bakken Industries","9012 Prairie Blvd","Williston","ND","58801","701-555-0301","procurement@bakken.com"
EOF

    cat > "$MDB_CONVERSION_DIR/test/data/sample_inventory.csv" << 'EOF'
"WkOrder","CustID","Joints","Size","Weight","Grade","Connection","DateIn","WellIn","LeaseIn"
"WO-001",1,100,"5 1/2""",2500.50,"L80","BTC","2024-01-15","Well-PB-001","Lease-PB-A"
"WO-002",2,150,"7""",4200.75,"P110","VAM TOP","2024-01-16","Well-EF-002","Lease-EF-B"
"WO-003",3,75,"9 5/8""",6800.25,"N80","LTC","2024-01-17","Well-BK-003","Lease-BK-C"
EOF

    cat > "$MDB_CONVERSION_DIR/test/data/sample.cfm" << 'EOF'
<cfquery name="getCustomers" datasource="inventory">
    SELECT CustID, CustName, BillAddr 
    FROM Customers 
    WHERE IsDeleted = 0
    ORDER BY CustName
</cfquery>

<cfquery name="getInventory" datasource="inventory">
    SELECT i.WkOrder, i.Joints, i.Size, i.Grade, c.CustName
    FROM Inventory i
    INNER JOIN Customers c ON i.CustID = c.CustID
    WHERE i.DateIn >= #CreateDate(2024,1,1)#
</cfquery>
EOF

    echo "✅ Test data created"
}

# Create root orchestrator Makefile
create_root_makefile() {
    echo ""
    echo "🔧 Creating root orchestrator Makefile..."
    
    # Check if root Makefile already exists
    if [ -f "$PROJECT_ROOT/Makefile" ]; then
        echo "⚠️  Root Makefile already exists"
        echo "Creating backup and adding integration commands..."
        cp "$PROJECT_ROOT/Makefile" "$PROJECT_ROOT/Makefile.backup.$(date +%Y%m%d_%H%M%S)"
        
        # Add integration commands to existing Makefile
        cat >> "$PROJECT_ROOT/Makefile" << 'EOF'

# =============================================================================
# MDB CONVERSION TOOLS INTEGRATION
# =============================================================================

tools: ## Show conversion tool commands
	@if [ -d "tools/mdb-conversion" ]; then \
		echo "🔗 Conversion tool commands (cd tools/mdb-conversion && make <command>):"; \
		cd tools/mdb-conversion && make help 2>/dev/null || echo "  Run 'make setup-tools' first"; \
	else \
		echo "❌ Tools directory not found: tools/mdb-conversion"; \
		echo "  Run 'make setup-tools' to create it"; \
	fi

setup-tools: ## Setup conversion tools
	@echo "🔧 Setting up conversion tools..."
	@if [ ! -d "tools/mdb-conversion" ]; then \
		echo "❌ Tools directory not found - structure needs to be created first"; \
		exit 1; \
	fi
	@cd tools/mdb-conversion && make setup
	@echo "✅ Conversion tools ready!"

# Convenience shortcuts for conversion tools
convert-mdb: ## Convert MDB file (usage: make convert-mdb FILE=database.mdb)
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Usage: make convert-mdb FILE=path/to/database.mdb"; \
		exit 1; \
	fi
	@cd tools/mdb-conversion && make convert-mdb FILE="$(abspath $(FILE))"

analyze-cf: ## Analyze ColdFusion app (usage: make analyze-cf DIR=cf_app)
	@if [ -z "$(DIR)" ]; then \
		echo "❌ Usage: make analyze-cf DIR=path/to/cf_application"; \
		exit 1; \
	fi
	@cd tools/mdb-conversion && make analyze-cf DIR="$(abspath $(DIR))"

test-conversion: ## Run conversion tool tests
	@cd tools/mdb-conversion && make test-conversion

full-analysis: ## Complete analysis (usage: make full-analysis COMPANY=company_name)
	@if [ -z "$(COMPANY)" ]; then \
		echo "❌ Usage: make full-analysis COMPANY=company_name"; \
		exit 1; \
	fi
	@cd tools/mdb-conversion && make full-analysis COMPANY="$(COMPANY)"
EOF
        
        echo "✅ Integration commands added to existing Makefile"
        echo "💾 Backup created: Makefile.backup.$(date +%Y%m%d_%H%M%S)"
    else
        # Create new root Makefile
        cat > "$PROJECT_ROOT/Makefile" << 'EOF'
# Root Project Makefile - Orchestrates all components
.PHONY: help backend tools conversion
.DEFAULT_GOAL := help

# Project paths
BACKEND_DIR := backend
TOOLS_DIR := tools/mdb-conversion

help: ## Show all available commands
	@echo "🏗️  Oil & Gas Inventory System"
	@echo "=============================="
	@echo ""
	@echo "MAIN COMPONENTS:"
	@echo "  make backend           - Backend application commands"
	@echo "  make tools             - MDB conversion tools commands"
	@echo "  make setup-all         - Setup entire project"
	@echo ""
	@echo "CONVERSION SHORTCUTS:"
	@echo "  make convert-mdb FILE=database.mdb    - Convert MDB file"
	@echo "  make analyze-cf DIR=cf_app            - Analyze ColdFusion app"
	@echo "  make test-conversion                  - Test conversion tools"
	@echo ""
	@echo "For detailed commands, run:"
	@echo "  make backend    # Shows backend-specific commands"
	@echo "  make tools      # Shows conversion tool commands"

backend: ## Show backend application commands
	@if [ -d "$(BACKEND_DIR)" ]; then \
		echo "🔗 Backend commands (cd backend && make <command>):"; \
		cd $(BACKEND_DIR) && make help 2>/dev/null || echo "  No help available - check backend/Makefile"; \
	else \
		echo "❌ Backend directory not found: $(BACKEND_DIR)"; \
	fi

tools: ## Show conversion tool commands
	@if [ -d "$(TOOLS_DIR)" ]; then \
		echo "🔗 Conversion tool commands (cd tools/mdb-conversion && make <command>):"; \
		cd $(TOOLS_DIR) && make help 2>/dev/null || echo "  Run 'make setup-tools' first"; \
	else \
		echo "❌ Tools directory not found: $(TOOLS_DIR)"; \
		echo "  Run 'make setup-tools' to create it"; \
	fi

setup-all: ## Setup entire project (backend + tools)
	@echo "🚀 Setting up complete project..."
	@$(MAKE) setup-tools
	@if [ -d "$(BACKEND_DIR)" ] && [ -f "$(BACKEND_DIR)/Makefile" ]; then \
		echo "Setting up backend..."; \
		cd $(BACKEND_DIR) && make setup 2>/dev/null || echo "⚠️  Backend setup failed or not available"; \
	fi
	@echo "✅ Project setup complete!"

setup-tools: ## Setup conversion tools
	@echo "🔧 Setting up conversion tools..."
	@if [ ! -d "$(TOOLS_DIR)" ]; then \
		echo "❌ Tools directory not found - run the implementation script first"; \
		exit 1; \
	fi
	@cd $(TOOLS_DIR) && make setup
	@echo "✅ Conversion tools ready!"

# Convenience shortcuts for conversion tools
convert-mdb: ## Convert MDB file (usage: make convert-mdb FILE=database.mdb)
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Usage: make convert-mdb FILE=path/to/database.mdb"; \
		exit 1; \
	fi
	@cd $(TOOLS_DIR) && make convert-mdb FILE="$(abspath $(FILE))"

analyze-cf: ## Analyze ColdFusion app (usage: make analyze-cf DIR=cf_app)  
	@if [ -z "$(DIR)" ]; then \
		echo "❌ Usage: make analyze-cf DIR=path/to/cf_application"; \
		exit 1; \
	fi
	@cd $(TOOLS_DIR) && make analyze-cf DIR="$(abspath $(DIR))"

test-conversion: ## Run conversion tool tests
	@cd $(TOOLS_DIR) && make test-conversion

full-analysis: ## Complete analysis (usage: make full-analysis COMPANY=company_name)
	@if [ -z "$(COMPANY)" ]; then \
		echo "❌ Usage: make full-analysis COMPANY=company_name"; \
		exit 1; \
	fi
	@cd $(TOOLS_DIR) && make full-analysis COMPANY="$(COMPANY)"

# Status and reporting
status: ## Show overall project status
	@echo "📊 Project Status"
	@echo "================="
	@echo "Backend:"
	@if [ -d "$(BACKEND_DIR)" ]; then \
		echo "  ✅ Directory exists"; \
		if [ -f "$(BACKEND_DIR)/go.mod" ]; then echo "  ✅ Go module found"; else echo "  ⚠️  No Go module"; fi; \
	else \
		echo "  ❌ Backend directory missing"; \
	fi
	@echo "Tools:"
	@if [ -d "$(TOOLS_DIR)" ]; then \
		echo "  ✅ Directory exists"; \
		if [ -f "$(TOOLS_DIR)/go.mod" ]; then echo "  ✅ Go module found"; else echo "  ⚠️  No Go module"; fi; \
	else \
		echo "  ❌ Tools directory missing - run implementation script"; \
	fi
	@echo "Output:"
	@if [ -d "output" ]; then \
		echo "  ✅ Output directory exists"; \
		conv_count=$(find output -name "*.csv" -o -name "*.sql" 2>/dev/null | wc -l); \
		echo "  📄 Conversion files: $conv_count"; \
	else \
		echo "  ⚠️  No output directory"; \
	fi

clean: ## Clean all build artifacts
	@echo "🧹 Cleaning project..."
	@if [ -d "$(BACKEND_DIR)" ]; then cd $(BACKEND_DIR) && make clean 2>/dev/null || true; fi
	@if [ -d "$(TOOLS_DIR)" ]; then cd $(TOOLS_DIR) && make clean 2>/dev/null || true; fi
	@rm -rf output/conversion/* 2>/dev/null || true
	@echo "✅ Cleanup complete"
EOF
        
        echo "✅ Root Makefile created"
    fi
}

# Create .gitignore files
create_gitignore_files() {
    echo ""
    echo "📝 Creating .gitignore files..."
    
    cat > "$MDB_CONVERSION_DIR/.gitignore" << 'EOF'
# Go build artifacts
/build/
/bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test outputs
/test/output/
*.test
*.out

# Conversion outputs (keep structure, ignore data)
*.csv
*.sql
*.log

# Temporary files
*.tmp
*.temp
.DS_Store
.vscode/
.idea/

# Go mod cache
go.sum
EOF

    cat > "$TOOLS_DIR/.gitignore" << 'EOF'
# Script outputs
*.log
*.tmp

# OS files
.DS_Store
Thumbs.db
EOF

    echo "✅ .gitignore files created"
}

# Verify implementation
verify_implementation() {
    echo ""
    echo "🔍 Verifying implementation..."
    
    local errors=0
    
    # Check directory structure
    echo "📁 Checking directory structure..."
    local required_dirs=(
        "$MDB_CONVERSION_DIR"
        "$MDB_CONVERSION_DIR/cmd"
        "$MDB_CONVERSION_DIR/internal"
        "$MDB_CONVERSION_DIR/config"
        "$MDB_CONVERSION_DIR/test/data"
        "$PROJECT_ROOT/output"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [ -d "$dir" ]; then
            echo "  ✅ $dir"
        else
            echo "  ❌ $dir - missing"
            errors=$((errors + 1))
        fi
    done
    
    # Check key files
    echo ""
    echo "📄 Checking key files..."
    local required_files=(
        "$MDB_CONVERSION_DIR/go.mod"
        "$MDB_CONVERSION_DIR/Makefile"
        "$MDB_CONVERSION_DIR/README.md"
        "$MDB_CONVERSION_DIR/config/oil_gas_mappings.json"
        "$MDB_CONVERSION_DIR/cmd/mdb_processor.go"
        "$MDB_CONVERSION_DIR/cmd/cf_query_analyzer.go"
        "$MDB_CONVERSION_DIR/cmd/conversion_tester.go"
        "$PROJECT_ROOT/Makefile"
    )
    
    for file in "${required_files[@]}"; do
        if [ -f "$file" ]; then
            echo "  ✅ $(basename "$file")"
        else
            echo "  ❌ $(basename "$file") - missing"
            errors=$((errors + 1))
        fi
    done
    
    # Check Go module validity
    echo ""
    echo "🔧 Checking Go module..."
    cd "$MDB_CONVERSION_DIR"
    if go mod verify >/dev/null 2>&1; then
        echo "  ✅ Go module valid"
    else
        echo "  ⚠️  Go module issues (expected for placeholders)"
    fi
    cd "$PROJECT_ROOT"
    
    # Summary
    echo ""
    if [ $errors -eq 0 ]; then
        echo "✅ Implementation verification passed!"
    else
        echo "⚠️  Implementation has $errors issues"
    fi
    
    return $errors
}

# Main execution
main() {
    echo "🚀 Starting Option 1 implementation..."
    
    create_directories
    create_go_module
    create_config_files
    create_tools_makefile
    create_placeholder_go_files
    create_documentation
    create_test_data
    create_root_makefile
    create_gitignore_files
    
    verify_implementation
    local verification_result=$?
    
    echo ""
    echo "🎉 Option 1 Implementation Complete!"
    echo ""
    echo "📁 Structure Created:"
    echo "  📂 tools/mdb-conversion/           # Conversion tools module"
    echo "  📂 tools/mdb-conversion/cmd/       # Main executables (placeholders)"
    echo "  📂 tools/mdb-conversion/internal/  # Private packages (placeholders)"
    echo "  📂 tools/mdb-conversion/config/    # Configuration files ✅"
    echo "  📂 output/conversion/              # Shared output directory"
    echo "  📄 Makefile                       # Root orchestrator"
    echo ""
    echo "🚀 Next Steps:"
    echo ""
    echo "1. **Test the structure:**"
    echo "   cd tools/mdb-conversion"
    echo "   make status"
    echo "   make setup    # (will build placeholder tools)"
    echo ""
    echo "2. **From project root:**"
    echo "   make tools    # Show tools help"
    echo "   make status   # Overall project status"
    echo ""
    echo "3. **Replace placeholders with full implementations:**"
    echo "   - Replace cmd/*.go files with comprehensive code from artifacts"
    echo "   - Implement internal packages (mapping, validation, testing, models)"
    echo "   - Complete Makefile commands (currently placeholders)"
    echo ""
    echo "4. **Test full implementation:**"
    echo "   make test-conversion"
    echo "   make convert-mdb FILE=sample.mdb"
    echo "   make analyze-cf DIR=sample_cf_app"
    echo ""
    echo "📚 Documentation:"
    echo "  - tools/mdb-conversion/README.md - Overview and status"
    echo "  - tools/mdb-conversion/docs/IMPLEMENTATION_GUIDE.md - Detailed next steps"
    echo ""
    if [ $verification_result -eq 0 ]; then
        echo "✅ All structure verification checks passed!"
    else
        echo "⚠️  Some verification issues found - check output above"
    fi
    echo ""
    echo "🎯 Ready for implementation phase!"
}

# Execute main function
main "$@"

#!/bin/bash
# adaptive_migrate_structure.sh
# Smart migration script that adapts to your current structure

set -e

echo "🔄 Adaptive Structure Migration"
echo "==============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# Check current directory
if [ ! -f "Makefile" ] || [ ! -d "backend" ]; then
    log_error "This script must be run from the project root directory"
    exit 1
fi

log_info "Analyzing current structure..."

# Backup timestamp
BACKUP_DIR="migration_backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Step 1: Analyze what we have
echo ""
log_info "Step 1: Current structure analysis"

if [ -d "tools" ]; then
    log_info "Found existing tools directory"
    cp -r tools "$BACKUP_DIR/tools_existing"
    
    # List what's actually in tools
    echo "Current tools structure:"
    find tools -type f | head -20 | sed 's/^/  /'
    
    # Check for old mdb-conversion structure
    if [ -d "tools/mdb-conversion" ]; then
        log_warning "Found old mdb-conversion structure"
        echo "Contents of tools/mdb-conversion:"
        find tools/mdb-conversion -type f | head -10 | sed 's/^/  /'
    fi
else
    log_info "No existing tools directory found"
fi

# Step 2: Create new structure
echo ""
log_info "Step 2: Creating new modular structure"

# Create the complete new structure
mkdir -p tools/{cmd,internal/{config,processor,mapping,validation,exporters,reporting},test/{fixtures,integration,performance,business},config,docs,bin,output/{csv,sql,reports,generated},coverage}

log_success "New directory structure created"

# Step 3: Smart file migration
echo ""
log_info "Step 3: Smart file migration"

migrate_file_smart() {
    local search_pattern="$1"
    local dest_dir="$2"
    local desc="$3"
    
    log_info "Looking for $desc files..."
    
    # Find files matching pattern anywhere in the project
    found_files=$(find . -name "$search_pattern" -not -path "./.git/*" -not -path "./migration_backup*" -not -path "./tools/cmd/*" -not -path "./tools/internal/*" 2>/dev/null || true)
    
    if [ -n "$found_files" ]; then
        echo "$found_files" | while read -r file; do
            if [ -f "$file" ]; then
                log_info "Migrating $desc: $file → $dest_dir/"
                mkdir -p "$dest_dir"
                cp "$file" "$dest_dir/"
                log_success "✓ $(basename "$file") migrated"
            fi
        done
    else
        log_warning "No $desc files found - will create defaults"
    fi
}

# Migrate configuration files
migrate_file_smart "*oil_gas*.json" "tools/config" "oil & gas mappings"
migrate_file_smart "*company*.json" "tools/config" "company templates"
migrate_file_smart "*validation*.json" "tools/config" "validation rules"
migrate_file_smart "*workflow*.json" "tools/config" "workflow configs"

# Migrate any existing Go files
migrate_file_smart "*processor*.go" "tools/cmd" "processor"
migrate_file_smart "*analyzer*.go" "tools/cmd" "analyzer"
migrate_file_smart "*tester*.go" "tools/cmd" "tester"

# Migrate test data
if [ -d "tools/mdb-conversion/test" ] || [ -d "tools/test" ]; then
    log_info "Migrating test data..."
    find tools -name "*.csv" -o -name "*.mdb" 2>/dev/null | while read -r file; do
        if [ -f "$file" ]; then
            cp "$file" tools/test/fixtures/
            log_success "✓ Test file migrated: $(basename "$file")"
        fi
    done
fi

# Migrate documentation
migrate_file_smart "*README*.md" "tools/docs" "documentation"

# Step 4: Create default files
echo ""
log_info "Step 4: Creating default configuration files"

# Create go.mod if it doesn't exist
if [ ! -f "tools/go.mod" ]; then
    log_info "Creating Go module..."
    cat > tools/go.mod << 'EOF'
module github.com/dcotelessa/oil-gas-inventory/tools

go 1.21

require (
	github.com/lib/pq v1.10.9
)
EOF
    log_success "Go module created"
fi

# Create default oil & gas mappings if none exist
if [ ! -f "tools/config/oil_gas_mappings.json" ]; then
    log_info "Creating default oil & gas mappings..."
    cat > tools/config/oil_gas_mappings.json << 'EOF'
{
  "oil_gas_mappings": {
    "customer": {
      "postgresql_name": "customers",
      "data_type": "varchar(255)",
      "business_rules": ["normalize_customer_name"],
      "required": true
    },
    "work_order": {
      "postgresql_name": "work_order", 
      "data_type": "varchar(50)",
      "business_rules": ["validate_work_order_format"],
      "required": true
    },
    "grade": {
      "postgresql_name": "grade",
      "data_type": "varchar(10)",
      "business_rules": ["validate_grade"],
      "required": false
    },
    "size": {
      "postgresql_name": "size",
      "data_type": "varchar(20)",
      "business_rules": ["validate_size"],
      "required": false
    },
    "connection": {
      "postgresql_name": "connection",
      "data_type": "varchar(50)",
      "business_rules": ["validate_connection"],
      "required": false
    }
  },
  "processing_options": {
    "workers": 4,
    "batch_size": 1000,
    "memory_limit_mb": 4096,
    "timeout_minutes": 60,
    "dry_run": false,
    "continue_on_error": true
  },
  "database_config": {
    "max_connections": 10,
    "idle_connections": 2,
    "connection_timeout_seconds": 30,
    "query_timeout_seconds": 300,
    "schema": "store",
    "create_tables": true,
    "use_transactions": true
  },
  "output_settings": {
    "csv_output": true,
    "sql_output": true,
    "postgresql_direct": false,
    "validation_report": true,
    "business_report": true,
    "output_dir": "output",
    "file_naming": "timestamp"
  },
  "business_rules": {
    "valid_grades": [
      {"code": "J55", "description": "Standard grade steel casing"},
      {"code": "JZ55", "description": "Enhanced J55 grade"},
      {"code": "L80", "description": "Higher strength grade"},
      {"code": "N80", "description": "Medium strength grade"},
      {"code": "P105", "description": "High performance grade"},
      {"code": "P110", "description": "Premium performance grade"},
      {"code": "Q125", "description": "Ultra-high strength grade"}
    ],
    "valid_sizes": [
      {"size": "4 1/2\"", "description": "4.5 inch diameter"},
      {"size": "5\"", "description": "5 inch diameter"},
      {"size": "5 1/2\"", "description": "5.5 inch diameter"},
      {"size": "7\"", "description": "7 inch diameter"},
      {"size": "8 5/8\"", "description": "8.625 inch diameter"},
      {"size": "9 5/8\"", "description": "9.625 inch diameter"},
      {"size": "10 3/4\"", "description": "10.75 inch diameter"}
    ],
    "valid_connections": [
      {"type": "BTC", "description": "Buttress Thread Casing"},
      {"type": "LTC", "description": "Long Thread Casing"},
      {"type": "STC", "description": "Short Thread Casing"},
      {"type": "VAM TOP", "description": "VAM Top Premium"},
      {"type": "VAM ACE", "description": "VAM Ace Premium"}
    ]
  }
}
EOF
    log_success "Default oil & gas mappings created"
fi

# Create company template if none exists
if [ ! -f "tools/config/company_template.json" ]; then
    log_info "Creating company template..."
    cat > tools/config/company_template.json << 'EOF'
{
  "company_name": "Default Company",
  "work_order_prefix": "WO",
  "custom_mappings": {},
  "contact_info": {
    "primary_contact": "Unknown",
    "email": "",
    "phone": "",
    "address": ""
  },
  "business_settings": {
    "fiscal_year_start": "01-01",
    "default_currency": "USD",
    "time_zone": "UTC",
    "custom_fields": {}
  }
}
EOF
    log_success "Company template created"
fi

# Create validation rules if none exist
if [ ! -f "tools/config/validation_rules.json" ]; then
    log_info "Creating validation rules..."
    cat > tools/config/validation_rules.json << 'EOF'
[
  {
    "field_name": "grade",
    "rule_type": "enum",
    "parameters": ["J55", "JZ55", "L80", "N80", "P105", "P110", "Q125"],
    "error_action": "warn",
    "error_message": "Invalid grade specified",
    "priority": 2
  },
  {
    "field_name": "work_order",
    "rule_type": "format",
    "parameters": ["^[A-Z]{1,3}-\\d{6}$"],
    "error_action": "reject",
    "error_message": "Work order must follow pattern: ABC-123456",
    "priority": 1
  },
  {
    "field_name": "customer",
    "rule_type": "required",
    "parameters": [],
    "error_action": "reject",
    "error_message": "Customer name is required",
    "priority": 1
  }
]
EOF
    log_success "Validation rules created"
fi

# Create workflow config
if [ ! -f "tools/config/workflow_config.json" ]; then
    log_info "Creating workflow config..."
    cat > tools/config/workflow_config.json << 'EOF'
{
  "workflows": [
    {
      "name": "mdb_conversion",
      "description": "Convert MDB files to PostgreSQL",
      "steps": [
        {"name": "analyze", "required": true},
        {"name": "extract", "required": true},
        {"name": "validate", "required": true},
        {"name": "transform", "required": true},
        {"name": "export", "required": true}
      ]
    }
  ]
}
EOF
    log_success "Workflow config created"
fi

# Step 5: Create sample test data
echo ""
log_info "Step 5: Creating sample test data"

if [ ! -f "tools/test/fixtures/sample_customers.csv" ]; then
    cat > tools/test/fixtures/sample_customers.csv << 'EOF'
CustID,CustomerName,BillingAddress,City,State,Phone
1,"Permian Basin Energy","1234 Oil Field Rd","Midland","TX","432-555-0101"
2,"Eagle Ford Solutions","5678 Shale Ave","San Antonio","TX","210-555-0201"
3,"Bakken Industries","9012 Prairie Blvd","Williston","ND","701-555-0301"
EOF
    log_success "Sample customers data created"
fi

if [ ! -f "tools/test/fixtures/sample_inventory.csv" ]; then
    cat > tools/test/fixtures/sample_inventory.csv << 'EOF'
WorkOrder,CustomerID,Customer,Joints,Size,Weight,Grade,Connection,DateIn
"LB-001001",1,"Permian Basin Energy",100,"5 1/2\"",2500.50,"L80","BTC","2024-01-15"
"LB-001002",2,"Eagle Ford Solutions",150,"7\"",4200.75,"P110","VAM TOP","2024-01-16"
"LB-001003",3,"Bakken Industries",75,"9 5/8\"",6800.25,"N80","LTC","2024-01-17"
EOF
    log_success "Sample inventory data created"
fi

# Step 6: Create README
if [ ! -f "tools/README.md" ]; then
    log_info "Creating tools README..."
    cat > tools/README.md << 'EOF'
# Oil & Gas Inventory - Migration Tools

Modern Go-based tools for converting legacy MDB files and ColdFusion applications to PostgreSQL.

## Quick Start

```bash
# From project root
make migration::setup
make migration::build
make migration::convert FILE=database.mdb COMPANY="Company Name"
```

## Structure

- `cmd/` - Command-line tools (mdb_processor, cf_analyzer, etc.)
- `internal/` - Internal packages (config, processor, mapping, validation)
- `config/` - Configuration files for oil & gas industry standards
- `test/` - Test files and fixtures
- `output/` - Generated conversion outputs

## Configuration

The tools use JSON configuration files in `config/`:

- `oil_gas_mappings.json` - Industry-specific column mappings and business rules
- `company_template.json` - Company-specific customizations
- `validation_rules.json` - Data validation rules
- `workflow_config.json` - Workflow definitions

## Development

```bash
# Build tools
make migration::build

# Run tests  
make migration::test

# Test with sample data
make migration::test-sample
```

See the main project Makefile for all available commands.
EOF
    log_success "Tools README created"
fi

# Step 7: Update .gitignore
echo ""
log_info "Step 7: Updating .gitignore"

if [ -f ".gitignore" ]; then
    if ! grep -q "tools/bin/" ".gitignore"; then
        echo "" >> .gitignore
        echo "# Tools build artifacts" >> .gitignore
        echo "tools/bin/" >> .gitignore
        echo "tools/output/" >> .gitignore
        echo "tools/coverage/" >> .gitignore
        echo "tools/*.tmp" >> .gitignore
        log_success "Updated .gitignore"
    else
        log_info ".gitignore already contains tools entries"
    fi
fi

# Step 8: Create cleanup script
echo ""
log_info "Step 8: Creating cleanup script"

cat > clean_old_structure.sh << 'EOF'
#!/bin/bash
# clean_old_structure.sh
# Removes old structure after migration verification

echo "🗑️  Cleaning old structure..."

# Remove old mdb-conversion directory if it exists
if [ -d "tools/mdb-conversion" ]; then
    echo "Removing tools/mdb-conversion..."
    rm -rf tools/mdb-conversion
    echo "✅ Old mdb-conversion structure removed"
fi

# Remove any other old structure files
find tools -name "*.old" -delete 2>/dev/null || true

echo "✅ Cleanup complete"
EOF

chmod +x clean_old_structure.sh
log_success "Cleanup script created: clean_old_structure.sh"

# Final summary
echo ""
echo "🎯 Adaptive Migration Complete!"
echo "==============================="
echo ""
log_success "Repository structure successfully migrated to new modular system"
echo ""
echo "📋 What was done:"
echo "  ✅ Created new modular directory structure"
echo "  ✅ Migrated any existing files found"
echo "  ✅ Created comprehensive default configurations"
echo "  ✅ Generated sample test data"
echo "  ✅ Updated .gitignore"
echo ""
echo "📁 New structure:"
find tools -type d | head -15 | sed 's/^/  /'
echo ""
echo "📄 Configuration files created:"
find tools/config -name "*.json" | sed 's/^/  /'
echo ""
echo "📋 Next Steps:"
echo "  1. Review the new structure: ls -la tools/"
echo "  2. Test the setup: make migration::setup"
echo "  3. Initialize Go modules: cd tools && go mod tidy"
echo "  4. Verify: make migration::status"
echo "  5. Clean old structure: ./clean_old_structure.sh"
echo ""
echo "📁 Backup Location: $BACKUP_DIR/"
echo ""
log_warning "Important: Test the new structure before removing backups!"

echo ""
log_info "Migration script completed successfully! 🎉"

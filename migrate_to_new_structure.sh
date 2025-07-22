#!/bin/bash
# migrate_to_new_structure.sh
# Migrates from old tools/mdb-conversion structure to new modular tools/ structure

set -e

echo "🔄 Oil & Gas Inventory - Structure Migration"
echo "============================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
OLD_TOOLS_DIR="tools/mdb-conversion"
NEW_TOOLS_DIR="tools"
BACKUP_DIR="migration_backup_$(date +%Y%m%d_%H%M%S)"

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in the right directory
if [ ! -f "Makefile" ] || [ ! -d "backend" ]; then
    log_error "This script must be run from the project root directory"
    exit 1
fi

log_info "Starting repository structure migration..."

# Step 1: Create backup
echo ""
log_info "Step 1: Creating backup of current structure"
mkdir -p "$BACKUP_DIR"

if [ -d "$OLD_TOOLS_DIR" ]; then
    log_info "Backing up existing tools directory..."
    cp -r "$OLD_TOOLS_DIR" "$BACKUP_DIR/"
    log_success "Backup created: $BACKUP_DIR/"
else
    log_warning "Old tools directory not found: $OLD_TOOLS_DIR"
fi

# Backup any existing tools/ files
if [ -d "tools" ]; then
    log_info "Backing up existing tools/ directory..."
    cp -r tools "$BACKUP_DIR/tools_existing"
fi

# Step 2: Create new directory structure
echo ""
log_info "Step 2: Creating new modular directory structure"

# Create main tools structure
mkdir -p tools/{cmd,internal/{config,processor,mapping,validation,exporters,reporting},test/{fixtures,integration,performance,business},config,docs}

# Create internal package subdirectories
mkdir -p tools/internal/processor
mkdir -p tools/internal/mapping
mkdir -p tools/internal/validation
mkdir -p tools/internal/exporters
mkdir -p tools/internal/reporting

# Create output directories
mkdir -p tools/output/{csv,sql,reports,generated}

# Create test directories
mkdir -p tools/test/{fixtures,integration,performance,business}

# Create binary and coverage directories
mkdir -p tools/{bin,coverage}

log_success "New directory structure created"

# Step 3: Migrate existing files
echo ""
log_info "Step 3: Migrating existing files to new structure"

# Function to migrate files with intelligent mapping
migrate_file() {
    local src="$1"
    local dest="$2"
    local desc="$3"
    
    if [ -f "$src" ]; then
        log_info "Migrating $desc: $src → $dest"
        mkdir -p "$(dirname "$dest")"
        cp "$src" "$dest"
        log_success "✓ $desc migrated"
    else
        log_warning "Source file not found: $src"
    fi
}

# Migrate configuration files
if [ -d "$OLD_TOOLS_DIR" ]; then
    migrate_file "$OLD_TOOLS_DIR/config/oil_gas_mappings.json" "tools/config/oil_gas_mappings.json" "Oil & gas mappings"
    migrate_file "$OLD_TOOLS_DIR/config/company_template.json" "tools/config/company_template.json" "Company template"
    migrate_file "$OLD_TOOLS_DIR/config/validation_rules.json" "tools/config/validation_rules.json" "Validation rules"
    
    # Migrate any existing Go files
    if [ -f "$OLD_TOOLS_DIR/main.go" ]; then
        migrate_file "$OLD_TOOLS_DIR/main.go" "tools/cmd/mdb_processor.go" "Main MDB processor"
    fi
    
    # Migrate test data
    if [ -d "$OLD_TOOLS_DIR/test" ]; then
        log_info "Migrating test data..."
        cp -r "$OLD_TOOLS_DIR/test"/* tools/test/fixtures/ 2>/dev/null || true
    fi
    
    # Migrate any documentation
    if [ -f "$OLD_TOOLS_DIR/README.md" ]; then
        migrate_file "$OLD_TOOLS_DIR/README.md" "tools/docs/migration_from_old.md" "Documentation"
    fi
fi

# Step 4: Create new Go module and files
echo ""
log_info "Step 4: Initializing new Go module and creating skeleton files"

# Create go.mod for tools
if [ ! -f "tools/go.mod" ]; then
    log_info "Creating Go module for tools..."
    cat > tools/go.mod << 'EOF'
module github.com/your-org/oil-gas-inventory/tools

go 1.21

require (
	github.com/lib/pq v1.10.9
)
EOF
    log_success "Go module created"
fi

# Create basic configuration files
log_info "Creating default configuration files..."

# Default oil & gas mappings
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
    }
  },
  "processing_options": {
    "workers": 4,
    "batch_size": 1000,
    "memory_limit_mb": 4096,
    "timeout_minutes": 60
  },
  "database_config": {
    "max_connections": 10,
    "idle_connections": 2,
    "schema": "store"
  },
  "output_settings": {
    "csv_output": true,
    "sql_output": true,
    "postgresql_direct": false,
    "validation_report": true
  }
}
EOF

# Company template
cat > tools/config/company_template.json << 'EOF'
{
  "company_name": "Default Company",
  "work_order_prefix": "WO",
  "custom_mappings": {},
  "contact_info": {
    "primary_contact": "Unknown",
    "email": "",
    "phone": ""
  }
}
EOF

# Validation rules
cat > tools/config/validation_rules.json << 'EOF'
[
  {
    "field_name": "grade",
    "rule_type": "enum",
    "parameters": ["J55", "JZ55", "L80", "N80", "P105", "P110", "Q125"],
    "error_action": "warn",
    "error_message": "Invalid grade specified"
  },
  {
    "field_name": "work_order",
    "rule_type": "format",
    "parameters": ["^[A-Z]{1,3}-\\d{6}$"],
    "error_action": "reject",
    "error_message": "Work order must follow pattern: ABC-123456"
  }
]
EOF

log_success "Configuration files created"

# Create basic README for tools
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

## Development

```bash
# Build tools
go build -o bin/mdb_processor cmd/mdb_processor.go

# Run tests  
go test ./...

# Generate documentation
go doc ./...
```

See the main project Makefile for all available commands.
EOF

# Step 5: Create sample test files
echo ""
log_info "Step 5: Creating sample test files"

# Sample CSV test data
cat > tools/test/fixtures/sample_customers.csv << 'EOF'
CustID,CustomerName,BillingAddress,City,State,Phone
1,"Permian Basin Energy","1234 Oil Field Rd","Midland","TX","432-555-0101"
2,"Eagle Ford Solutions","5678 Shale Ave","San Antonio","TX","210-555-0201"
3,"Bakken Industries","9012 Prairie Blvd","Williston","ND","701-555-0301"
EOF

cat > tools/test/fixtures/sample_inventory.csv << 'EOF'
WorkOrder,CustomerID,Customer,Joints,Size,Weight,Grade,Connection,DateIn
"LB-001001",1,"Permian Basin Energy",100,"5 1/2\"",2500.50,"L80","BTC","2024-01-15"
"LB-001002",2,"Eagle Ford Solutions",150,"7\"",4200.75,"P110","VAM TOP","2024-01-16"
"LB-001003",3,"Bakken Industries",75,"9 5/8\"",6800.25,"N80","LTC","2024-01-17"
EOF

log_success "Sample test files created"

# Step 6: Generate file inventory
echo ""
log_info "Step 6: Generating migration report"

echo ""
echo "📋 Migration Report"
echo "=================="
echo ""

echo "📁 New Directory Structure:"
find tools -type d | sort | sed 's/^/  /'

echo ""
echo "📄 Files Created/Migrated:"
find tools -type f | sort | sed 's/^/  /'

if [ -d "$OLD_TOOLS_DIR" ]; then
    echo ""
    echo "📦 Original Files (backed up to $BACKUP_DIR):"
    find "$OLD_TOOLS_DIR" -type f | sort | sed 's/^/  /'
fi

# Step 7: Update .gitignore if needed
echo ""
log_info "Step 7: Updating .gitignore"

if [ -f ".gitignore" ]; then
    # Add tools-specific ignores if they don't exist
    if ! grep -q "tools/bin/" ".gitignore"; then
        echo "" >> .gitignore
        echo "# Tools build artifacts" >> .gitignore
        echo "tools/bin/" >> .gitignore
        echo "tools/output/" >> .gitignore
        echo "tools/coverage/" >> .gitignore
        echo "tools/*.tmp" >> .gitignore
        log_success "Updated .gitignore with tools entries"
    else
        log_info ".gitignore already contains tools entries"
    fi
else
    log_warning ".gitignore not found - consider creating one"
fi

# Step 8: Create cleanup script
echo ""
log_info "Step 8: Creating cleanup script"

cat > remove_old_structure.sh << 'EOF'
#!/bin/bash
# remove_old_structure.sh
# Removes the old tools/mdb-conversion structure after migration verification

echo "🗑️  Removing old tools structure..."

if [ -d "tools/mdb-conversion" ]; then
    echo "Removing tools/mdb-conversion..."
    rm -rf tools/mdb-conversion
    echo "✅ Old structure removed"
else
    echo "❌ Old structure not found"
fi

echo "✅ Cleanup complete"
EOF

chmod +x remove_old_structure.sh
log_success "Cleanup script created: remove_old_structure.sh"

# Final summary
echo ""
echo "🎯 Migration Complete!"
echo "====================="
echo ""
log_success "Repository structure successfully migrated to new modular system"
echo ""
echo "📋 Next Steps:"
echo "  1. Review migrated files in tools/ directory"
echo "  2. Test the new structure: make migration::setup"
echo "  3. Build tools: make migration::build"
echo "  4. Run tests: make migration::test"
echo "  5. If everything works, run: ./remove_old_structure.sh"
echo ""
echo "📁 Backup Location: $BACKUP_DIR/"
echo "📄 Cleanup Script: ./remove_old_structure.sh"
echo ""
log_warning "Important: Test the new structure thoroughly before removing backups!"

# Display directory tree
if command -v tree >/dev/null 2>&1; then
    echo ""
    echo "📊 New Tools Directory Structure:"
    tree tools/ -I 'bin|output|coverage' || true
fi

echo ""
log_info "Migration script completed successfully! 🎉"

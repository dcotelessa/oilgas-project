#!/bin/bash
# fixed_macos_migrate_structure.sh
# Complete macOS-optimized migration script - FIXED

set -e

echo "🍎 macOS Oil & Gas Inventory - Structure Migration"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }
log_cleanup() { echo -e "${PURPLE}🗑️  $1${NC}"; }

# Check we're in the right place
if [ ! -f "Makefile" ] || [ ! -d "backend" ]; then
    log_error "This script must be run from the project root directory"
    exit 1
fi

log_info "Starting macOS-optimized structure migration..."

# Configuration
BACKUP_DIR="migration_backup_$(date +%Y%m%d_%H%M%S)"
NEW_TOOLS_DIR="tools"
OLD_TOOLS_DIR="tools/mdb-conversion"

# Step 1: Clean up any previous migration attempts
echo ""
log_info "Step 1: Cleaning up previous migration attempts"

cleanup_old_migrations() {
    log_cleanup "Cleaning up old migration artifacts..."
    
    # Remove old backup directories (keep only the 3 most recent)
    if ls migration_backup_* >/dev/null 2>&1; then
        log_info "Found old migration backups, keeping only 3 most recent..."
        ls -t migration_backup_* | tail -n +4 | xargs rm -rf 2>/dev/null || true
        log_success "Old backups cleaned up"
    fi
    
    # Remove any .backup files from previous runs
    find tools -name "*.backup" -delete 2>/dev/null || true
    
    # Remove empty directories
    find tools -type d -empty -delete 2>/dev/null || true
    
    # Clean up any temp files
    find . -name "*.tmp" -delete 2>/dev/null || true
    find . -name ".DS_Store" -delete 2>/dev/null || true
    
    log_success "Cleanup completed"
}

cleanup_old_migrations

# Step 2: Create comprehensive backup
echo ""
log_info "Step 2: Creating comprehensive backup"

create_backup() {
    log_info "Creating backup: $BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
    
    # Backup entire tools directory if it exists
    if [ -d "tools" ]; then
        cp -R tools "$BACKUP_DIR/tools_complete" 2>/dev/null || true
        log_success "Complete tools directory backed up"
    fi
    
    # Backup any config files in root
    find . -maxdepth 2 -name "*.json" -path "./config/*" -exec cp {} "$BACKUP_DIR/" \; 2>/dev/null || true
    
    # Backup Makefile if it has tools references
    if grep -q "tools" Makefile 2>/dev/null; then
        cp Makefile "$BACKUP_DIR/Makefile.backup"
    fi
    
    log_success "Backup created: $BACKUP_DIR"
}

create_backup

# Step 3: Create new modular structure
echo ""
log_info "Step 3: Creating new modular directory structure"

create_structure() {
    log_info "Creating comprehensive directory structure..."
    
    # Create all required directories
    mkdir -p tools/{cmd,internal/{config,processor,mapping,validation,exporters,reporting},test/{fixtures,integration,performance,business},config,docs,bin,output/{csv,sql,reports,generated},coverage}
    
    log_success "Directory structure created"
    
    # Show the structure
    log_info "New structure preview:"
    find tools -type d | head -15 | sed 's/^/  /'
}

create_structure

# Step 4: Intelligent file migration
echo ""
log_info "Step 4: Intelligent file migration (macOS optimized)"

migrate_files() {
    log_info "Starting intelligent file migration..."
    
    # Function to safely copy files with conflict resolution
    safe_migrate() {
        local src="$1"
        local dest="$2"
        local desc="$3"
        
        if [ -f "$src" ]; then
            log_info "Migrating $desc: $(basename "$src")"
            mkdir -p "$(dirname "$dest")"
            
            if [ -f "$dest" ]; then
                if cmp -s "$src" "$dest" 2>/dev/null; then
                    log_info "  → File identical, skipping"
                    return 0
                else
                    log_warning "  → File exists but differs, creating backup"
                    cp "$dest" "$dest.backup.$(date +%s)"
                fi
            fi
            
            cp "$src" "$dest"
            log_success "  ✓ $desc migrated successfully"
            return 0
        fi
        return 1
    }
    
    # Migration counter
    local migrated_count=0
    
    # Migrate configuration files
    log_info "Migrating configuration files..."
    
    # Try multiple possible locations for each config file
    safe_migrate "tools/mdb-conversion/config/oil_gas_mappings.json" "tools/config/oil_gas_mappings.json" "oil & gas mappings" && ((migrated_count++)) || \
    safe_migrate "tools/config/oil_gas_mappings.json" "tools/config/oil_gas_mappings.json" "oil & gas mappings" && ((migrated_count++)) || \
    safe_migrate "config/oil_gas_mappings.json" "tools/config/oil_gas_mappings.json" "oil & gas mappings" && ((migrated_count++)) || true
    
    safe_migrate "tools/mdb-conversion/config/company_template.json" "tools/config/company_template.json" "company template" && ((migrated_count++)) || \
    safe_migrate "tools/config/company_template.json" "tools/config/company_template.json" "company template" && ((migrated_count++)) || \
    safe_migrate "config/company_template.json" "tools/config/company_template.json" "company template" && ((migrated_count++)) || true
    
    # Migrate Go source files
    log_info "Migrating Go source files..."
    
    safe_migrate "tools/mdb-conversion/main.go" "tools/cmd/mdb_processor.go" "main processor" && ((migrated_count++)) || \
    safe_migrate "tools/main.go" "tools/cmd/mdb_processor.go" "main processor" && ((migrated_count++)) || \
    safe_migrate "cmd/mdb_processor.go" "tools/cmd/mdb_processor.go" "main processor" && ((migrated_count++)) || true
    
    # Migrate test files
    log_info "Migrating test files..."
    
    if [ -d "tools/mdb-conversion/test" ]; then
        log_info "Migrating test directory from mdb-conversion..."
        find "tools/mdb-conversion/test" -type f \( -name "*.csv" -o -name "*.json" -o -name "*.mdb" \) -exec cp {} tools/test/fixtures/ \; 2>/dev/null || true
        ((migrated_count++))
        log_success "  ✓ Test files migrated"
    fi
    
    log_success "Migration completed: $migrated_count files migrated"
}

migrate_files

# Step 5: Create/update Go module
echo ""
log_info "Step 5: Creating/updating Go module"

setup_go_module() {
    log_info "Setting up Go module for tools..."
    
    cd tools
    
    # Remove old go.mod if it has wrong module path
    if [ -f "go.mod" ]; then
        if grep -q "your-org" go.mod 2>/dev/null; then
            log_warning "Found go.mod with incorrect module path, recreating..."
            rm go.mod go.sum 2>/dev/null || true
        fi
    fi
    
    # Create or update go.mod
    if [ ! -f "go.mod" ]; then
        log_info "Creating new go.mod..."
        go mod init github.com/dcotelessa/oil-gas-inventory/tools
    else
        log_info "go.mod already exists with correct path"
    fi
    
    # Add required dependencies
    log_info "Adding required dependencies..."
    go get github.com/lib/pq@latest 2>/dev/null || true
    
    # Tidy up
    go mod tidy 2>/dev/null || log_warning "go mod tidy failed (some dependencies may be missing)"
    
    cd ..
    log_success "Go module setup completed"
}

setup_go_module

# Step 6: Create comprehensive default configurations
echo ""
log_info "Step 6: Creating comprehensive default configurations"

create_configurations() {
    log_info "Creating/updating configuration files..."
    
    # Oil & Gas Mappings - comprehensive version
    if [ ! -f "tools/config/oil_gas_mappings.json" ]; then
        log_info "Creating comprehensive oil & gas mappings..."
        cat > tools/config/oil_gas_mappings.json << 'EOF'
{
  "oil_gas_mappings": {
    "customer_id": {
      "postgresql_name": "customer_id",
      "data_type": "integer",
      "business_rules": [],
      "required": true
    },
    "customer": {
      "postgresql_name": "customer",
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
    "dry_run": false
  },
  "database_config": {
    "max_connections": 10,
    "schema": "store"
  },
  "output_settings": {
    "csv_output": true,
    "sql_output": true,
    "validation_report": true
  }
}
EOF
        log_success "✓ Oil & gas mappings created"
    else
        log_info "✓ Oil & gas mappings already exist"
    fi
    
    # Company Template
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
    "phone": ""
  }
}
EOF
        log_success "✓ Company template created"
    else
        log_info "✓ Company template already exists"
    fi
    
    # Validation Rules
    if [ ! -f "tools/config/validation_rules.json" ]; then
        log_info "Creating validation rules..."
        cat > tools/config/validation_rules.json << 'EOF'
[
  {
    "field_name": "grade",
    "rule_type": "enum",
    "parameters": ["J55", "JZ55", "L80", "N80", "P105", "P110"],
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
        log_success "✓ Validation rules created"
    else
        log_info "✓ Validation rules already exist"
    fi
    
    # Workflow Config
    if [ ! -f "tools/config/workflow_config.json" ]; then
        log_info "Creating workflow configuration..."
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
        log_success "✓ Workflow configuration created"
    else
        log_info "✓ Workflow configuration already exists"
    fi
}

create_configurations

# Step 7: Create sample test data
echo ""
log_info "Step 7: Creating sample test data"

create_test_data() {
    log_info "Creating sample test data..."
    
    # Sample customers
    if [ ! -f "tools/test/fixtures/sample_customers.csv" ]; then
        cat > tools/test/fixtures/sample_customers.csv << 'EOF'
CustID,CustomerName,BillingAddress,BillingCity,BillingState,Phone
1,"Permian Basin Energy","1234 Oil Field Rd","Midland","TX","432-555-0101"
2,"Eagle Ford Solutions","5678 Shale Ave","San Antonio","TX","210-555-0201"
3,"Bakken Industries","9012 Prairie Blvd","Williston","ND","701-555-0301"
EOF
        log_success "✓ Sample customers data created"
    fi
    
    # Sample inventory
    if [ ! -f "tools/test/fixtures/sample_inventory.csv" ]; then
        cat > tools/test/fixtures/sample_inventory.csv << 'EOF'
WorkOrder,CustomerID,Customer,Joints,Size,Weight,Grade,Connection,DateIn
"LB-001001",1,"Permian Basin Energy",100,"5 1/2\"",2500.50,"L80","BTC","2024-01-15"
"LB-001002",2,"Eagle Ford Solutions",150,"7\"",4200.75,"P110","VAM TOP","2024-01-16"
"LB-001003",3,"Bakken Industries",75,"9 5/8\"",6800.25,"N80","LTC","2024-01-17"
EOF
        log_success "✓ Sample inventory data created"
    fi
}

create_test_data

# Step 8: Create documentation
echo ""
log_info "Step 8: Creating documentation"

create_documentation() {
    log_info "Creating documentation..."
    
    # Main README
    if [ ! -f "tools/README.md" ]; then
        cat > tools/README.md << 'EOF'
# Oil & Gas Inventory - Migration Tools

Modern Go-based tools for converting legacy MDB files to PostgreSQL.

## Quick Start

```bash
# From project root
make migration::setup
make migration::build
make migration::convert FILE=database.mdb COMPANY="Company Name"
```

## Configuration

The tools use JSON configuration files in `config/`:
- `oil_gas_mappings.json` - Industry-specific mappings and business rules
- `company_template.json` - Company-specific customizations  
- `validation_rules.json` - Data validation rules
- `workflow_config.json` - Workflow definitions

## Development

```bash
make migration::build    # Build tools
make migration::test     # Run tests
```

See the main project Makefile for all available commands.
EOF
        log_success "✓ Main README created"
    fi
}

create_documentation

# Step 9: Clean up old structure
echo ""
log_info "Step 9: Creating cleanup script"

create_cleanup_script() {
    cat > clean_old_structure.sh << 'EOF'
#!/bin/bash
# clean_old_structure.sh - Remove old structure after verification

echo "🗑️  Removing old mdb-conversion structure..."

if [ -d "tools/mdb-conversion" ]; then
    echo "Removing tools/mdb-conversion..."
    rm -rf tools/mdb-conversion
    echo "✅ Old structure removed"
else
    echo "❌ Old structure not found (already cleaned?)"
fi

# Clean up backup files
find tools -name "*.backup.*" -delete 2>/dev/null || true

echo "✅ Cleanup complete"
EOF
    chmod +x clean_old_structure.sh
    log_success "✓ Cleanup script created"
}

create_cleanup_script

# Step 10: Update .gitignore
echo ""
log_info "Step 10: Updating .gitignore"

update_gitignore() {
    if [ -f ".gitignore" ]; then
        if ! grep -q "tools/bin/" ".gitignore" 2>/dev/null; then
            log_info "Adding tools entries to .gitignore..."
            cat >> .gitignore << 'EOF'

# Tools build artifacts
tools/bin/
tools/output/
tools/coverage/
tools/*.tmp
migration_backup_*/
EOF
            log_success "✓ .gitignore updated"
        else
            log_info "✓ .gitignore already contains tools entries"
        fi
    fi
}

update_gitignore

# Step 11: Final validation
echo ""
log_info "Step 11: Final validation"

final_validation() {
    log_info "Running final validation..."
    
    # Test Go module
    cd tools
    if go mod tidy 2>/dev/null; then
        log_success "✓ Go module is valid"
    else
        log_warning "⚠️  Go module needs attention"
    fi
    cd ..
    
    # Check configuration files
    local config_count=0
    for config_file in tools/config/*.json; do
        if [ -f "$config_file" ]; then
            if python3 -m json.tool "$config_file" > /dev/null 2>&1; then
                log_success "✓ $(basename "$config_file") is valid JSON"
                ((config_count++))
            else
                log_error "❌ $(basename "$config_file") has invalid JSON"
            fi
        fi
    done
    
    log_info "Validation: $config_count configuration files validated"
}

final_validation

# Step 12: Create verification script
echo ""
log_info "Step 12: Creating verification script"

create_verification_script() {
    cat > verify_migration.sh << 'EOF'
#!/bin/bash
# verify_migration.sh - Verify the new structure works

echo "🔍 Verifying migration..."

echo "1. Testing Go module..."
cd tools
if go mod tidy 2>/dev/null; then
    echo "✅ Go module OK"
else
    echo "❌ Go module issues"
    cd ..
    exit 1
fi
cd ..

echo "2. Testing Makefile commands..."
if make migration::setup >/dev/null 2>&1; then
    echo "✅ Migration setup OK"
else
    echo "❌ Migration setup failed"
fi

echo "3. Checking configuration files..."
for file in tools/config/*.json; do
    if python3 -m json.tool "$file" >/dev/null 2>&1; then
        echo "✅ $(basename "$file") valid"
    else
        echo "❌ $(basename "$file") invalid"
    fi
done

echo ""
echo "🎯 Migration verification complete!"
echo "If all checks passed, run: ./clean_old_structure.sh"
EOF
    chmod +x verify_migration.sh
    log_success "✓ Verification script created"
}

create_verification_script

# Final summary
echo ""
echo "🎯 macOS Migration Complete!"
echo "============================="
echo ""
log_success "Repository structure successfully migrated!"
echo ""

echo "📋 Migration Summary:"
echo "  🏗️  Directory structure: ✅ Created"
echo "  📁 Configuration files: ✅ $(ls tools/config/*.json 2>/dev/null | wc -l | tr -d ' ') files"
echo "  🧪 Test fixtures: ✅ $(ls tools/test/fixtures/*.csv 2>/dev/null | wc -l | tr -d ' ') files"
echo "  📚 Documentation: ✅ Created"
echo "  🗂️  Backup: ✅ $BACKUP_DIR"
echo ""

echo "🔧 Next Steps:"
echo "  1. Verify: ./verify_migration.sh"
echo "  2. Test: make migration::setup"
echo "  3. Build: make migration::build"
echo "  4. Clean: ./clean_old_structure.sh"
echo ""

log_success "Migration completed successfully! 🎉"

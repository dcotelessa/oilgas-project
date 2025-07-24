#!/bin/bash
# scripts/fix_makefile_conflicts.sh
# Fix Makefile conflicts and complete API integration setup
# Run from project root: ./scripts/fix_makefile_conflicts.sh

set -e

PROJECT_ROOT="$(pwd)"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"
MAKE_DIR="$PROJECT_ROOT/make"

echo "🔧 Fixing Makefile Conflicts for API Integration"
echo "================================================"
echo "Project root: $PROJECT_ROOT"
echo ""

# Validate we're in the right place
if [ ! -f "Makefile" ]; then
    echo "❌ Error: Run this script from project root (where Makefile is located)"
    exit 1
fi

# Create make directory if it doesn't exist
if [ ! -d "make" ]; then
    echo "📁 Creating make/ directory for modular Makefiles..."
    mkdir -p make
    echo "✅ Created make/ directory"
fi

echo "1. Analyzing current Makefile structure..."
echo "=========================================="

# Show current state
echo "📁 Current make/ modules:"
if [ -d make ] && ls make/*.mk >/dev/null 2>&1; then
    ls -la make/*.mk | awk '{print "   " $9}'
else
    echo "   No .mk files found (will create modular structure)"
fi

echo ""
echo "📄 Current Makefile includes:"
if grep "include make/" Makefile >/dev/null 2>&1; then
    grep "include make/" Makefile | awk '{print "   " $0}'
else
    echo "   No modular includes found (will add them)"
fi

echo ""
echo "2. Identifying conflicts..."
echo "=========================="

# Find duplicate targets (only if make directory has files)
echo "🔍 Checking for duplicate targets..."
if [ -d make ] && ls make/*.mk >/dev/null 2>&1; then
    DUPLICATES=$(grep -h "^[a-zA-Z_-]*:" make/*.mk 2>/dev/null | sort | uniq -d | cut -d: -f1 || true)
    
    if [ -n "$DUPLICATES" ]; then
        echo "❌ Found duplicate targets:"
        for target in $DUPLICATES; do
            echo "   $target found in:"
            grep -l "^$target:" make/*.mk | sed 's/^/     /'
        done
    else
        echo "✅ No duplicate targets found"
    fi
else
    echo "✅ No existing modules - creating clean structure"
fi

echo ""
echo "3. Creating compatibility layer..."
echo "================================="

# Create make/compatibility.mk with missing targets
cat > "$MAKE_DIR/compatibility.mk" << 'EOF'
# =============================================================================
# Compatibility Layer - Missing Targets and Dependencies
# Created by scripts/fix_makefile_conflicts.sh
# =============================================================================

.PHONY: dev-ensure-db dev-wait dev-db-reset

## Ensure database is accessible (compatibility target)
dev-ensure-db:
	@echo "🔍 Checking database accessibility..."
	@$(MAKE) db-health 2>/dev/null || (echo "❌ Database not accessible - run: make docker-up && make dev-setup" && exit 1)

## Wait for services to be ready (compatibility target)  
dev-wait:
	@echo "⏳ Waiting for services to be ready..."
	@sleep 3
	@echo "✅ Services should be ready"

## Reset development database (compatibility target)
dev-db-reset:
	@echo "🔄 Resetting development database..."
	@$(MAKE) test-db-reset 2>/dev/null || echo "⚠️  test-db-reset not available"

## Additional compatibility aliases
.PHONY: ensure-db wait-services db-reset

ensure-db: dev-ensure-db
wait-services: dev-wait
db-reset: dev-db-reset
EOF

echo "✅ Created make/compatibility.mk"

echo ""
echo "4. Resolving duplicate targets..."
echo "================================"

# Only try to fix duplicates if make directory exists with files
if [ -d make ] && ls make/*.mk >/dev/null 2>&1; then
    # Fix test-db-setup conflict (remove from database.mk if exists)
    if [ -f make/database.mk ] && grep -q "^test-db-setup:" make/database.mk 2>/dev/null; then
        echo "🔧 Removing test-db-setup from database.mk (keeping in testing.mk)..."
        # Create backup
        cp make/database.mk make/database.mk.backup
        # Remove the target and its commands (up to next target or EOF)
        sed -i '/^test-db-setup:/,/^[a-zA-Z_-]*:/{ /^[a-zA-Z_-]*:/!d; }' make/database.mk
        # Handle case where it's the last target
        sed -i '/^test-db-setup:/d' make/database.mk
        echo "   ✅ Removed from database.mk"
    fi

    # Fix import-mdb-data conflict (rename in database.mk to avoid conflict with data.mk)
    if [ -f make/database.mk ] && grep -q "^import-mdb-data:" make/database.mk 2>/dev/null; then
        echo "🔧 Renaming import-mdb-data to db-import-legacy in database.mk..."
        # Create backup if not already created
        [ ! -f make/database.mk.backup ] && cp make/database.mk make/database.mk.backup
        # Rename the target
        sed -i 's/^import-mdb-data:/db-import-legacy:/' make/database.mk
        echo "   ✅ Renamed to db-import-legacy in database.mk"
    fi
else
    echo "✅ No existing modules to fix - creating clean structure"
fi

echo ""
echo "5. Creating updated module files..."
echo "=================================="

# Create the fixed API module
cat > "$MAKE_DIR/api.mk" << 'EOF'
# =============================================================================
# API Development Commands
# =============================================================================

.PHONY: api-start api-test api-test-quick api-dev api-logs api-examples api-curl-examples api-check-db

## Start API server in development mode
api-start: dev-ensure-db
	@echo "🚀 Starting API server..."
	@echo "📋 Health: http://localhost:8000/health"
	@echo "🔌 API: http://localhost:8000/api/v1"
	@echo "Press Ctrl+C to stop"
	@cd backend && go run cmd/server/main.go

## Test API integration with repository layer
api-test:
	@echo "🧪 Testing API integration..."
	@test -f scripts/test_api_integration.sh || (echo "❌ Test script missing" && exit 1)
	@chmod +x scripts/test_api_integration.sh
	@scripts/test_api_integration.sh

## Quick API health check
api-test-quick:
	@echo "⚡ Quick API test..."
	@curl -s http://localhost:8000/health | jq -r '"Status: " + .status + " | Service: " + .service' 2>/dev/null || echo "❌ API not responding (is it running?)"

## Start API in development mode with auto-reload
api-dev: dev-ensure-db
	@echo "🔄 Starting API with auto-reload..."
	@which air > /dev/null || (echo "💡 Install air: cd backend && go install github.com/cosmtrek/air@latest" && exit 1)
	@cd backend && air

## Show API usage examples
api-examples:
	@echo "🔍 API Usage Examples"
	@echo "===================="
	@echo "Health Check:"
	@echo "  curl http://localhost:8000/health"
	@echo ""
	@echo "Get All Customers:"
	@echo "  curl http://localhost:8000/api/v1/customers | jq"
	@echo ""
	@echo "Search Customers:"
	@echo "  curl 'http://localhost:8000/api/v1/customers/search?q=oil' | jq"
	@echo ""
	@echo "Get Customer by ID:"
	@echo "  curl http://localhost:8000/api/v1/customers/1 | jq"

## Check if database is ready for API
api-check-db:
	@echo "🔍 Checking database readiness for API..."
	@$(MAKE) db-health && echo "✅ Database accessible" || (echo "❌ Database not accessible" && exit 1)
	@echo "📊 Sample data check:"
	@$(MAKE) db-exec SQL="SELECT COUNT(*) as customers FROM store.customers;" 2>/dev/null || echo "❌ Cannot access customers table"
EOF

echo "✅ Created updated make/api.mk"

# Create the data module without conflicts
cat > "$MAKE_DIR/data.mk" << 'EOF'
# =============================================================================
# Data Import Commands (MDB Processing)
# =============================================================================

.PHONY: data-check data-convert data-import data-status data-setup data-backup data-stats

## Check MDB files and conversion readiness
data-check:
	@echo "📂 Checking for MDB data files..."
	@test -d data/mdb && echo "✅ data/mdb directory exists" || echo "❌ Create: mkdir -p data/mdb"
	@test -d data/csv && echo "✅ data/csv directory exists" || echo "❌ Create: mkdir -p data/csv"
	@echo ""
	@echo "Expected structure:"
	@echo "  data/mdb/customers.mdb"
	@echo "  data/mdb/inventory.mdb"
	@echo "  data/csv/ (converted files)"

## Convert MDB files to CSV
data-convert:
	@echo "🔄 Converting MDB files to CSV..."
	@mkdir -p data/csv
	@echo "⚠️  MDB converter needs customization"
	@echo "   Available in Phase 3B implementation"

## Import converted data
data-import: dev-ensure-db
	@echo "📥 Importing MDB data..."
	@echo "⚠️  Import process needs customization"
	@echo "   Available in Phase 3B implementation"

## Show import status and data counts
data-status: dev-ensure-db
	@echo "📊 Data Import Status"
	@echo "====================="
	@$(MAKE) db-exec SQL="SELECT 'Customers' as table_name, count(*) as records FROM store.customers UNION ALL SELECT 'Inventory', count(*) FROM store.inventory UNION ALL SELECT 'Grades', count(*) FROM store.grade UNION ALL SELECT 'Sizes', count(*) FROM store.sizes;" 2>/dev/null || echo "❌ Cannot query tables"

## Create data directories
data-setup:
	@echo "📁 Creating data directories..."
	@mkdir -p data/mdb data/csv data/backup
	@echo "✅ Data directories created"

## Show data statistics
data-stats: dev-ensure-db
	@echo "📊 Database Statistics"
	@echo "======================"
	@$(MAKE) db-exec SQL="SELECT 'customers' as table_name, count(*) FROM store.customers UNION ALL SELECT 'inventory', count(*) FROM store.inventory;" 2>/dev/null || echo "❌ Cannot query tables"

## Legacy aliases for backward compatibility
.PHONY: import-check convert-mdb import-mdb-data import-status

import-check: data-check
convert-mdb: data-convert
import-mdb-data: data-import
import-status: data-status
EOF

echo "✅ Created updated make/data.mk"

echo ""
echo "6. Updating main Makefile..."
echo "============================"

# Backup current Makefile
cp Makefile Makefile.backup
echo "💾 Backed up Makefile to Makefile.backup"

# Remove existing include statements for our modules
grep -v "include make/compatibility.mk" Makefile > Makefile.tmp && mv Makefile.tmp Makefile
grep -v "include make/api.mk" Makefile > Makefile.tmp && mv Makefile.tmp Makefile
grep -v "include make/data.mk" Makefile > Makefile.tmp && mv Makefile.tmp Makefile

# Add includes in correct order
if grep -q "include make/" Makefile; then
    # Add after last existing include
    echo "🔧 Adding new includes after existing ones..."
    # Find the last include line and add after it
    last_include_line=$(grep -n "include make/" Makefile | tail -1 | cut -d: -f1)
    {
        head -n "$last_include_line" Makefile
        echo "include make/compatibility.mk"
        echo "include make/api.mk" 
        echo "include make/data.mk"
        tail -n "+$(($last_include_line + 1))" Makefile
    } > Makefile.tmp && mv Makefile.tmp Makefile
else
    # Add at end of file
    echo "🔧 Adding module includes to end of Makefile..."
    echo "" >> Makefile
    echo "# Module includes added by fix_makefile_conflicts.sh" >> Makefile
    echo "include make/compatibility.mk" >> Makefile
    echo "include make/api.mk" >> Makefile
    echo "include make/data.mk" >> Makefile
fi

echo "✅ Updated Makefile with correct include order"

echo ""
echo "7. Creating API test script..."
echo "============================="

# Create the test script
mkdir -p scripts
cat > "$SCRIPTS_DIR/test_api_integration.sh" << 'EOF'
#!/bin/bash
# Test API integration after repository wiring

set -e

API_BASE="http://localhost:8000"
TIMEOUT=10

echo "🧪 Testing API Integration"
echo "================================"

# Check if server is running
echo -n "Checking if API server is running... "
if curl -s --max-time 5 "$API_BASE/health" > /dev/null; then
    echo "✅ Server responding"
else
    echo "❌ Server not responding"
    echo ""
    echo "Start the server first:"
    echo "  make api-start"
    exit 1
fi

# Function to test endpoint
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local expected_status="$3"
    local description="$4"
    
    echo -n "Testing $description... "
    
    response=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT \
        -X "$method" "$API_BASE$endpoint" \
        -H "Content-Type: application/json")
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        echo "✅ ($http_code)"
        if [ "$method" = "GET" ] && [ -n "$body" ]; then
            count=$(echo "$body" | jq -r '.count // empty' 2>/dev/null || echo "")
            if [ -n "$count" ]; then
                echo "   📊 Records: $count"
            fi
        fi
    else
        echo "❌ ($http_code)"
        echo "   Expected: $expected_status"
        return 1
    fi
}

# Test endpoints
echo ""
echo "1. Health Checks"
echo "----------------"
test_endpoint "GET" "/health" "200" "Health check"
test_endpoint "GET" "/api/v1/status" "200" "API status"

echo ""
echo "2. Customer Endpoints"
echo "---------------------"
test_endpoint "GET" "/api/v1/customers" "200" "Get all customers"
test_endpoint "GET" "/api/v1/customers/search?q=oil" "200" "Search customers"

echo ""
echo "3. Inventory Endpoints"
echo "----------------------"
test_endpoint "GET" "/api/v1/inventory" "200" "Get all inventory"

echo ""
echo "4. Reference Data"
echo "-----------------"
test_endpoint "GET" "/api/v1/grades" "200" "Get all grades"
test_endpoint "GET" "/api/v1/sizes" "200" "Get all sizes"

echo ""
echo "🎯 API Integration Test Summary"
echo "================================"
echo "✅ Server running and responding"
echo "✅ Repository layer connected"
echo "✅ Basic endpoints working"
echo ""
echo "🚀 API Integration Complete!"
echo ""
echo "Next steps:"
echo "1. Import your MDB data: make data-convert && make data-import"
echo "2. View API examples: make api-examples"
echo "3. Start development: make api-dev"
EOF

chmod +x "$SCRIPTS_DIR/test_api_integration.sh"
echo "✅ Created executable scripts/test_api_integration.sh"

echo ""
echo "8. Testing the fixes..."
echo "======================"

# Test Makefile syntax
echo -n "Testing Makefile syntax... "
if make -n help > /dev/null 2>&1; then
    echo "✅ Syntax OK"
else
    echo "❌ Still has issues"
    echo "   Run: make -n help"
fi

# Test API target
echo -n "Testing api-start target... "
if make -n api-start > /dev/null 2>&1; then
    echo "✅ Dependencies resolved"
else
    echo "❌ Still missing dependencies"
    make -n api-start 2>&1 | grep "No rule" | head -2
fi

echo ""
echo "9. Verification and next steps..."
echo "================================"

echo ""
echo "📋 Created files:"
echo "   make/compatibility.mk    # Missing target compatibility layer"
echo "   make/api.mk             # Updated API commands"
echo "   make/data.mk            # Updated data import commands" 
echo "   scripts/test_api_integration.sh  # API testing script"
echo "   Makefile.backup         # Backup of original Makefile"
echo ""

if [ -f make/database.mk.backup ]; then
    echo "📋 Backup files:"
    echo "   make/database.mk.backup # Backup before removing duplicates"
    echo ""
fi

echo "🧪 Test commands:"
echo "   make help               # Test overall Makefile syntax"
echo "   make api-check-db       # Check database readiness"
echo "   make api-start          # Start API server"
echo "   make api-test           # Test API integration (run in another terminal)"
echo ""

echo "🎯 Current Status Check:"
echo "========================"

# Quick status check
echo -n "Makefile loads: "
make -n help >/dev/null 2>&1 && echo "✅" || echo "❌"

echo -n "API target resolves: "
make -n api-start >/dev/null 2>&1 && echo "✅" || echo "❌"

echo -n "Test script exists: "
[ -x scripts/test_api_integration.sh ] && echo "✅" || echo "❌"

echo ""
echo "🚀 Ready for API Integration!"
echo "=============================="
echo "1. Run: make help                    # Verify everything loads"
echo "2. Run: make api-check-db           # Ensure database ready"
echo "3. Run: make api-start              # Start API server"
echo "4. In another terminal: make api-test  # Test integration"
echo ""
echo "✅ API Integration Setup Complete!"
EOF

chmod +x "$SCRIPTS_DIR/fix_makefile_conflicts.sh"

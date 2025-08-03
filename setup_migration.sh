#!/bin/bash
# setup_migration.sh - Complete MDB to PostgreSQL migration setup

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MDB_FILE="${MDB_FILE:-db_prep/petros.mdb}"
TENANT_ID="${TENANT_ID:-local-dev}"
SKIP_MDB_EXPORT="${SKIP_MDB_EXPORT:-false}"

echo -e "${BLUE}🚀 Oil & Gas Inventory System - Complete Migration Setup${NC}"
echo -e "${BLUE}==========================================================${NC}"
echo ""

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check prerequisites
echo -e "${YELLOW}📋 Checking prerequisites...${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is required but not installed"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi
print_status "Docker found"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is required but not installed"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi
print_status "Docker Compose found"

# Check Go
if ! command -v go &> /dev/null; then
    print_error "Go is required but not installed"
    echo "Please install Go: https://golang.org/doc/install"
    exit 1
fi
print_status "Go found"

# Check for MDB file and tools (only if not skipping MDB export)
if [ "$SKIP_MDB_EXPORT" != "true" ]; then
    if [ ! -f "$MDB_FILE" ]; then
        print_warning "MDB file not found at: $MDB_FILE"
        echo "Please ensure your Microsoft Access database is at: $MDB_FILE"
        echo "Or set MDB_FILE environment variable to the correct path"
        echo "Alternatively, set SKIP_MDB_EXPORT=true if you already have CSV files"
        exit 1
    fi
    print_status "MDB file found: $MDB_FILE"

    if ! command -v mdb-tables &> /dev/null; then
        print_warning "mdb-tools not found"
        echo "Installing mdb-tools..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            if command -v brew &> /dev/null; then
                brew install mdbtools
            else
                print_error "Homebrew not found. Please install mdb-tools manually"
                exit 1
            fi
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            # Linux
            if command -v apt-get &> /dev/null; then
                sudo apt-get update && sudo apt-get install -y mdb-tools
            elif command -v yum &> /dev/null; then
                sudo yum install -y mdb-tools
            else
                print_error "Package manager not supported. Please install mdb-tools manually"
                exit 1
            fi
        else
            print_error "OS not supported for automatic mdb-tools installation"
            exit 1
        fi
    fi
    print_status "mdb-tools available"
else
    print_info "Skipping MDB export (SKIP_MDB_EXPORT=true)"
fi

echo ""

# Create necessary directories
echo -e "${YELLOW}📁 Creating directories...${NC}"
mkdir -p database/{data/{exported,clean},logs,migrations,seeds,init}
mkdir -p docker
print_status "Directories created"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚙️  Creating .env file...${NC}"
    cat > .env << EOF
# Database Configuration
DEV_DATABASE_URL=postgresql://oilgas_user:oilgas_pass@localhost:5432/oilgas_dev
TEST_DATABASE_URL=postgresql://oilgas_test_user:oilgas_test_pass@localhost:5433/oilgas_test

# Migration Configuration
MDB_FILE=$MDB_FILE
TENANT_ID=$TENANT_ID
DATA_PATH=database/data/clean

# Docker Configuration
POSTGRES_DB=oilgas_dev
POSTGRES_USER=oilgas_user
POSTGRES_PASSWORD=oilgas_pass

POSTGRES_TEST_DB=oilgas_test
POSTGRES_TEST_USER=oilgas_test_user
POSTGRES_TEST_PASSWORD=oilgas_test_pass
EOF
    print_status ".env file created"
else
    print_info ".env file already exists"
fi

# Phase 1: Export and clean MDB data (if not skipping)
if [ "$SKIP_MDB_EXPORT" != "true" ]; then
    echo -e "${YELLOW}📤 Phase 1: Exporting MDB data...${NC}"
    
    # Export customers table
    echo "Exporting customers..."
    if mdb-export "$MDB_FILE" customers > database/data/exported/customers.csv 2>/dev/null; then
        print_status "Customers table exported"
    else
        print_warning "Customers table not found or empty"
    fi
    
    # Try alternative table names
    for table in customer cust; do
        if mdb-export "$MDB_FILE" "$table" > "database/data/exported/${table}.csv" 2>/dev/null; then
            print_status "Table '$table' exported"
        fi
    done
    
    echo -e "${YELLOW}🧹 Cleaning exported data...${NC}"
    
    # Build CSV cleaner
    if [ ! -f backend/cmd/tools/csv-cleaner/main.go ]; then
        print_error "CSV cleaner source not found"
        exit 1
    fi
    
    # Run CSV cleaner
    cd backend && go run cmd/tools/csv-cleaner/main.go \
        -input ../database/data/exported \
        -output ../database/data/clean \
        -log ../database/logs/cleaning.log
    cd ..
    
    print_status "Data cleaning completed"
else
    print_info "Skipping MDB export - using existing CSV files"
fi

echo ""

# Phase 2: Setup PostgreSQL containers
echo -e "${YELLOW}🐳 Phase 2: Setting up PostgreSQL containers...${NC}"

# Stop any existing containers
docker-compose down 2>/dev/null || true

# Start PostgreSQL containers
docker-compose up -d postgres postgres-test

# Wait for databases to be ready
echo "Waiting for databases to start..."
timeout=60
while [ $timeout -gt 0 ]; do
    if docker-compose exec -T postgres pg_isready -U oilgas_user -d oilgas_dev >/dev/null 2>&1 && \
       docker-compose exec -T postgres-test pg_isready -U oilgas_test_user -d oilgas_test >/dev/null 2>&1; then
        print_status "Databases are ready"
        break
    fi
    echo "Waiting for databases... ($timeout seconds remaining)"
    sleep 2
    timeout=$((timeout-2))
done

if [ $timeout -eq 0 ]; then
    print_error "Databases failed to start within timeout"
    exit 1
fi

echo ""

# Phase 3: Run database migrations
echo -e "${YELLOW}📊 Phase 3: Running database migrations...${NC}"

# Run initial setup
docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -f /docker-entrypoint-initdb.d/001_init_database.sql
docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -f /docker-entrypoint-initdb.d/001_init_database.sql

# Run customer domain migration
docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -f /docker-entrypoint-initdb.d/migrations/001_create_customers.sql
docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -f /docker-entrypoint-initdb.d/migrations/001_create_customers.sql

print_status "Database migrations completed"

echo ""

# Phase 4: Import customer data
echo -e "${YELLOW}📥 Phase 4: Importing customer data...${NC}"

# Check if we have customer data to import
customer_files=(database/data/clean/customers.csv database/data/clean/customer.csv database/data/clean/cust.csv)
customer_file=""

for file in "${customer_files[@]}"; do
    if [ -f "$file" ]; then
        customer_file="$file"
        break
    fi
done

if [ -n "$customer_file" ]; then
    print_info "Found customer data: $customer_file"
    
    # Build migrator
    cd backend && go build -o ../migrator cmd/migrator/main.go
    cd ..
    
    # Run data import
    DEV_DATABASE_URL="postgresql://oilgas_user:oilgas_pass@localhost:5432/oilgas_dev" \
    TEST_DATABASE_URL="postgresql://oilgas_test_user:oilgas_test_pass@localhost:5433/oilgas_test" \
    DATA_PATH="database/data/clean" \
    TENANT_ID="$TENANT_ID" \
    ./migrator
    
    print_status "Customer data imported"
else
    print_warning "No customer data files found - skipping data import"
    print_info "Available files:"
    ls -la database/data/clean/ 2>/dev/null || echo "No files in clean directory"
fi

echo ""

# Phase 5: Verification and summary
echo -e "${YELLOW}🔍 Phase 5: Verification...${NC}"

# Test database connections
if docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT COUNT(*) FROM store.customers;" >/dev/null 2>&1; then
    customer_count=$(docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -tAc "SELECT COUNT(*) FROM store.customers WHERE deleted = false;")
    print_status "Development database: $customer_count customers"
else
    print_warning "Could not verify development database"
fi

if docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -c "SELECT COUNT(*) FROM store.customers;" >/dev/null 2>&1; then
    test_customer_count=$(docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -tAc "SELECT COUNT(*) FROM store.customers WHERE deleted = false;")
    print_status "Test database: $test_customer_count customers"
else
    print_warning "Could not verify test database"
fi

echo ""
echo -e "${GREEN}🎉 Migration Setup Complete!${NC}"
echo -e "${GREEN}=============================${NC}"
echo ""
echo -e "${BLUE}📊 Summary:${NC}"
echo "✅ PostgreSQL containers running"
echo "✅ Customer domain schema created"
echo "✅ Multi-tenant architecture enabled"
echo "✅ Row-level security configured"
if [ -n "$customer_file" ]; then
    echo "✅ Customer data imported"
else
    echo "⚠️  Customer data import skipped (no data files found)"
fi
echo ""
echo -e "${BLUE}🎯 Access Points:${NC}"
echo "• Development DB:  postgresql://oilgas_user:oilgas_pass@localhost:5432/oilgas_dev"
echo "• Test DB:         postgresql://oilgas_test_user:oilgas_test_pass@localhost:5433/oilgas_test"
echo "• PgAdmin:         Run 'docker-compose up -d pgadmin' then http://localhost:8080"
echo ""
echo -e "${BLUE}🚀 Next Steps:${NC}"
echo "1. Start development server: make dev"
echo "2. Run customer tests: make test-customer"
echo "3. View migration logs: cat database/logs/migration_*.log"
echo "4. Start building your Go backend with the enhanced customer domain!"
echo ""
echo -e "${BLUE}📁 Important Files:${NC}"
echo "• Customer domain: backend/internal/customer/"
echo "• Migration logs: database/logs/"
echo "• Cleaned data: database/data/clean/"
echo "• Database config: .env"
echo ""
echo -e "${YELLOW}💡 Tip:${NC} Use 'make db-status' to check database status anytime"

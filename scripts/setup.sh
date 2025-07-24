#!/bin/bash
# scripts/phase3_5_multitenant_architecture.sh
# Phase 3.5: Multi-Tenant Architecture Implementation
# Implements tenant isolation with proper RLS policies

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🏗️  Phase 3.5: Multi-Tenant Architecture Implementation${NC}"
echo "=============================================================="

# Check and setup environment configuration
setup_environment() {
    echo -e "${BLUE}📋 Checking environment configuration...${NC}"
    
    # Detect existing environment file
    if [ -f ".env.local" ]; then
        echo -e "${GREEN}✅ Found .env.local file${NC}"
        ENV_FILE=".env.local"
    elif [ -f ".env" ]; then
        echo -e "${GREEN}✅ Found .env file${NC}"
        ENV_FILE=".env"
    else
        echo -e "${RED}❌ No environment file found${NC}"
        echo ""
        echo -e "${YELLOW}Environment file required at root level:${NC}"
        echo "  • .env.local (preferred for local development)"
        echo "  • .env (fallback)"
        echo ""
        echo -e "${YELLOW}Please create .env.local with your database configuration:${NC}"
        echo ""
        echo "# Database Configuration"
        echo "DATABASE_URL=postgresql://user:password@host:port/database_name"
        echo "DB_HOST=localhost"
        echo "DB_PORT=5432"
        echo "DB_NAME=your_database_name"
        echo "DB_USER=your_username"
        echo "DB_PASSWORD=your_password"
        echo ""
        echo "# Application Configuration"
        echo "APP_ENV=local"
        echo "APP_PORT=8000"
        echo ""
        echo -e "${RED}Setup cannot continue without environment configuration.${NC}"
        exit 1
    fi
    
    # Load environment variables
    echo -e "${BLUE}📄 Loading environment from: $ENV_FILE${NC}"
    set -o allexport
    source "$ENV_FILE"
    set +o allexport
    
    # Validate required environment variables and normalize them
    if [ -z "$DATABASE_URL" ] && [ -z "$POSTGRES_DB" ] && [ -z "$DB_NAME" ]; then
        echo -e "${RED}❌ Missing database configuration in $ENV_FILE${NC}"
        echo ""
        echo -e "${YELLOW}Required variables (one of these patterns):${NC}"
        echo "  Pattern 1: DATABASE_URL=postgresql://user:password@host:port/database"
        echo "  Pattern 2: POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD, etc."
        echo "  Pattern 3: DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD"
        echo ""
        exit 1
    fi
    
    # Normalize database variables from different naming conventions
    # Use POSTGRES_* variables if they exist, otherwise fall back to DB_*
    DB_NAME=${DB_NAME:-${POSTGRES_DB:-oilgas_inventory_local}}
    DB_USER=${DB_USER:-${POSTGRES_USER:-postgres}}
    DB_PASSWORD=${DB_PASSWORD:-${POSTGRES_PASSWORD:-postgres123}}
    DB_HOST=${DB_HOST:-${POSTGRES_HOST:-localhost}}
    DB_PORT=${DB_PORT:-${POSTGRES_PORT:-5433}}
    
    # Normalize DATABASE_URL format (postgres:// to postgresql://)
    if [ -n "$DATABASE_URL" ]; then
        # Convert postgres:// to postgresql:// for compatibility
        if echo "$DATABASE_URL" | grep -q "^postgres://"; then
            DATABASE_URL=$(echo "$DATABASE_URL" | sed 's/^postgres:/postgresql:/')
            echo -e "${BLUE}🔄 Normalized DATABASE_URL format (postgres:// → postgresql://)${NC}"
        fi
    else
        # Build DATABASE_URL from components if not provided
        DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
        echo -e "${BLUE}🔧 Built DATABASE_URL from component variables${NC}"
    fi
    
    # Export normalized variables for use by other tools
    export DATABASE_URL DB_NAME DB_USER DB_PASSWORD DB_HOST DB_PORT
    
    # Show configuration summary
    echo -e "${BLUE}📊 Environment Configuration:${NC}"
    echo "  Environment: ${APP_ENV:-local}"
    echo "  Database: ${DB_NAME}"
    echo "  Host: ${DB_HOST}:${DB_PORT}"
    echo "  User: ${DB_USER}"
    echo "  DATABASE_URL: ${DATABASE_URL}"
    echo -e "${GREEN}✅ Environment configuration loaded successfully${NC}"
}

# Test database connection and setup
setup_database() {
    echo -e "${BLUE}🔌 Setting up database connection...${NC}"
    
    echo "Testing connection to: ${DB_HOST}:${DB_PORT} as ${DB_USER}"
    
    # Test if PostgreSQL is running
    if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" >/dev/null 2>&1; then
        echo -e "${RED}❌ PostgreSQL is not running or not accessible${NC}"
        echo ""
        echo -e "${YELLOW}💡 To start PostgreSQL:${NC}"
        echo "   macOS (Homebrew): brew services start postgresql"
        echo "   Ubuntu/Debian:    sudo systemctl start postgresql"
        echo "   Docker:           docker run --name postgres -e POSTGRES_PASSWORD=$DB_PASSWORD -p 5433:5432 -d postgres"
        echo ""
        exit 1
    fi
    
    echo -e "${GREEN}✅ PostgreSQL is running${NC}"
    
    # Create database if it doesn't exist
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        echo -e "${GREEN}✅ Database $DB_NAME already exists${NC}"
    else
        echo -e "${YELLOW}📦 Creating database $DB_NAME...${NC}"
        createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"
        echo -e "${GREEN}✅ Database $DB_NAME created${NC}"
    fi
}

# Setup basic schema if not exists
setup_basic_schema() {
    echo -e "${BLUE}🏗️  Setting up basic schema...${NC}"
    
    # Check if basic tables exist
    if psql "$DATABASE_URL" -c "SELECT 1 FROM store.customers LIMIT 1;" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Basic schema already exists${NC}"
        return
    fi
    
    echo -e "${BLUE}🗄️  Creating basic schema and tables...${NC}"
    
    # Create the basic schema SQL
    psql "$DATABASE_URL" << 'EOF'
-- Basic schema setup for oil & gas inventory system

-- Create schemas
CREATE SCHEMA IF NOT EXISTS store;
CREATE SCHEMA IF NOT EXISTS migrations;

-- Set search path
SET search_path TO store, public;

-- Create migrations table
CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Customers table
CREATE TABLE IF NOT EXISTS store.customers (
    customer_id SERIAL PRIMARY KEY,
    customer VARCHAR(255) NOT NULL,
    billing_address TEXT,
    billing_city VARCHAR(100),
    billing_state VARCHAR(50),
    billing_zipcode VARCHAR(20),
    contact VARCHAR(255),
    phone VARCHAR(50),
    fax VARCHAR(50),
    email VARCHAR(255),
    deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Grades table
CREATE TABLE IF NOT EXISTS store.grade (
    grade VARCHAR(10) PRIMARY KEY,
    description TEXT
);

-- Sizes table
CREATE TABLE IF NOT EXISTS store.sizes (
    size_id SERIAL PRIMARY KEY,
    size VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

-- Users table
CREATE TABLE IF NOT EXISTS store.users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Inventory table
CREATE TABLE IF NOT EXISTS store.inventory (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100),
    work_order VARCHAR(100),
    r_number VARCHAR(100),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(255),
    joints INTEGER,
    rack VARCHAR(50),
    size VARCHAR(50),
    weight DECIMAL(10,2),
    grade VARCHAR(10) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd VARCHAR(100),
    w_string VARCHAR(100),
    swgcc VARCHAR(100),
    color VARCHAR(50),
    customer_po VARCHAR(100),
    fletcher VARCHAR(100),
    date_in DATE,
    date_out DATE,
    well_in VARCHAR(255),
    lease_in VARCHAR(255),
    well_out VARCHAR(255),
    lease_out VARCHAR(255),
    trucking VARCHAR(100),
    trailer VARCHAR(100),
    location VARCHAR(100),
    notes TEXT,
    pcode VARCHAR(50),
    cn VARCHAR(50),
    ordered_by VARCHAR(100),
    deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Received table
CREATE TABLE IF NOT EXISTS store.received (
    id SERIAL PRIMARY KEY,
    work_order VARCHAR(100),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(255),
    joints INTEGER,
    rack VARCHAR(50),
    size_id INTEGER REFERENCES store.sizes(size_id),
    size VARCHAR(50),
    weight DECIMAL(10,2),
    grade VARCHAR(10) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd VARCHAR(100),
    w_string VARCHAR(100),
    well VARCHAR(255),
    lease VARCHAR(255),
    ordered_by VARCHAR(100),
    notes TEXT,
    customer_po VARCHAR(100),
    date_received DATE,
    background TEXT,
    norm VARCHAR(100),
    services TEXT,
    bill_to_id INTEGER,
    entered_by VARCHAR(100),
    when_entered TIMESTAMP,
    trucking VARCHAR(100),
    trailer VARCHAR(100),
    in_production BOOLEAN DEFAULT FALSE,
    inspected_date DATE,
    threading_date DATE,
    straighten_required BOOLEAN DEFAULT FALSE,
    excess_material BOOLEAN DEFAULT FALSE,
    complete BOOLEAN DEFAULT FALSE,
    inspected_by VARCHAR(100),
    updated_by VARCHAR(100),
    when_updated TIMESTAMP,
    deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order);

-- Insert reference data
INSERT INTO store.grade (grade, description) VALUES 
('J55', 'Standard grade steel casing'),
('L80', 'Higher strength grade'),
('N80', 'Medium strength grade'),
('P110', 'Premium performance grade')
ON CONFLICT (grade) DO NOTHING;

INSERT INTO store.sizes (size, description) VALUES 
('5 1/2"', '5.5 inch diameter'),
('7"', '7 inch diameter'),
('9 5/8"', '9.625 inch diameter')
ON CONFLICT (size) DO NOTHING;

INSERT INTO store.customers (customer, billing_city, billing_state, contact, phone, email) VALUES 
('Permian Basin Energy', 'Midland', 'TX', 'John Smith', '432-555-0101', 'operations@permianbasin.com'),
('Eagle Ford Solutions', 'San Antonio', 'TX', 'Sarah Johnson', '210-555-0201', 'drilling@eagleford.com')
ON CONFLICT DO NOTHING;

-- Record migration
INSERT INTO migrations.schema_migrations (version, name) VALUES 
('001', 'Initial schema setup')
ON CONFLICT (version) DO NOTHING;
EOF
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Basic schema setup completed${NC}"
    else
        echo -e "${RED}❌ Schema setup failed${NC}"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}📋 Checking prerequisites...${NC}"
    
    # Check if Go backend structure exists
    if [ ! -d "backend" ]; then
        echo -e "${YELLOW}📁 Creating backend directory structure...${NC}"
        mkdir -p backend/{cmd/server,internal/{models,handlers,middleware,repository,services},migrations,seeds,test}
        
        # Create basic go.mod if it doesn't exist
        if [ ! -f "backend/go.mod" ]; then
            cat > backend/go.mod << 'EOF'
module oilgas-backend

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-migrate/migrate/v4 v4.16.2
    github.com/joho/godotenv v1.4.0
    github.com/lib/pq v1.10.9
)
EOF
        fi
    fi
    
    echo -e "${GREEN}✅ Prerequisites satisfied${NC}"
}

# Create tenant migration files
create_tenant_migrations() {
    echo -e "${BLUE}🗄️  Creating tenant migration files...${NC}"
    
    mkdir -p backend/migrations
    
    cat > backend/migrations/001_add_tenant_architecture.sql << 'EOF'
-- Migration 003: Add Tenant Architecture
BEGIN;

-- Create tenants table
CREATE TABLE IF NOT EXISTS store.tenants (
    tenant_id SERIAL PRIMARY KEY,
    tenant_name VARCHAR(255) NOT NULL UNIQUE,
    tenant_slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    contact_email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    active BOOLEAN DEFAULT TRUE,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add tenant_id to existing tables
ALTER TABLE store.users ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES store.tenants(tenant_id);
ALTER TABLE store.inventory ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES store.tenants(tenant_id);
ALTER TABLE store.received ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES store.tenants(tenant_id);

-- Customer-tenant relationship
CREATE TABLE IF NOT EXISTS store.customer_tenant_assignments (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES store.customers(customer_id) ON DELETE CASCADE,
    tenant_id INTEGER REFERENCES store.tenants(tenant_id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) DEFAULT 'primary',
    active BOOLEAN DEFAULT TRUE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by INTEGER REFERENCES store.users(user_id),
    notes TEXT,
    UNIQUE(customer_id, tenant_id)
);

-- User roles within tenants
CREATE TABLE IF NOT EXISTS store.user_tenant_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES store.users(user_id) ON DELETE CASCADE,
    tenant_id INTEGER REFERENCES store.tenants(tenant_id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by INTEGER REFERENCES store.users(user_id),
    active BOOLEAN DEFAULT TRUE,
    UNIQUE(user_id, tenant_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON store.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_id ON store.inventory(tenant_id);
CREATE INDEX IF NOT EXISTS idx_received_tenant_id ON store.received(tenant_id);

COMMIT;
EOF

    cat > backend/migrations/002_tenant_rls_policies.sql << 'EOF'
-- Migration 004: Tenant Row-Level Security Policies
BEGIN;

-- Enable RLS on tenant-sensitive tables
ALTER TABLE store.inventory ENABLE ROW LEVEL SECURITY;
ALTER TABLE store.received ENABLE ROW LEVEL SECURITY;
ALTER TABLE store.customer_tenant_assignments ENABLE ROW LEVEL SECURITY;

-- Create function to get current tenant context
CREATE OR REPLACE FUNCTION get_current_tenant_id() 
RETURNS INTEGER AS $$
BEGIN
    BEGIN
        RETURN current_setting('app.current_tenant_id')::INTEGER;
    EXCEPTION
        WHEN OTHERS THEN
            RETURN (
                SELECT tenant_id 
                FROM store.users 
                WHERE user_id = current_setting('app.current_user_id')::INTEGER
            );
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to check if user is admin
CREATE OR REPLACE FUNCTION is_admin_user()
RETURNS BOOLEAN AS $$
BEGIN
    BEGIN
        RETURN current_setting('app.user_role') = 'admin';
    EXCEPTION
        WHEN OTHERS THEN
            RETURN FALSE;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- RLS policies
CREATE POLICY inventory_tenant_isolation ON store.inventory
    FOR ALL TO authenticated_users
    USING (
        tenant_id = get_current_tenant_id() 
        OR is_admin_user()
        OR tenant_id IS NULL
    );

CREATE POLICY received_tenant_isolation ON store.received
    FOR ALL TO authenticated_users
    USING (
        tenant_id = get_current_tenant_id()
        OR is_admin_user()
        OR tenant_id IS NULL
    );

COMMIT;
EOF

    echo -e "${GREEN}✅ Tenant migration files created${NC}"
}

# Create tenant seed data
create_tenant_seeds() {
    echo -e "${BLUE}🌱 Creating tenant seed data...${NC}"
    
    mkdir -p backend/seeds
    
    cat > backend/seeds/tenant_seeds.sql << 'EOF'
-- Tenant seed data
SET search_path TO store, public;

-- Insert default tenant
INSERT INTO store.tenants (tenant_name, tenant_slug, description, contact_email, phone) 
VALUES (
    'Petros',
    'petros',
    'Main division for Petros operations',
    'operations@petros.com',
    '432-555-0100'
) ON CONFLICT (tenant_slug) DO NOTHING;

-- Assign existing data to default tenant
DO $$
DECLARE
    petros_tenant_id INTEGER;
BEGIN
    SELECT tenant_id INTO petros_tenant_id 
    FROM store.tenants 
    WHERE tenant_slug = 'petros';

    -- Update existing data
    UPDATE store.users SET tenant_id = petros_tenant_id WHERE tenant_id IS NULL;
    UPDATE store.inventory SET tenant_id = petros_tenant_id WHERE tenant_id IS NULL;
    UPDATE store.received SET tenant_id = petros_tenant_id WHERE tenant_id IS NULL;

    -- Assign customers to default tenant
    INSERT INTO store.customer_tenant_assignments (customer_id, tenant_id, relationship_type)
    SELECT customer_id, petros_tenant_id, 'primary'
    FROM store.customers
    ON CONFLICT (customer_id, tenant_id) DO NOTHING;
END $$;
EOF

    echo -e "${GREEN}✅ Tenant seed data created${NC}"
}

# Update Makefile with tenant commands
update_makefile() {
    echo -e "${BLUE}📝 Updating Makefile...${NC}"
    
    if [ ! -f "Makefile" ]; then
        cat > Makefile << 'EOF'
# Oil & Gas Inventory System - Makefile
# Load environment variables from .env.local or .env
ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
else ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Fallback database URL if not set in environment
DATABASE_URL ?= postgresql://postgres:postgres123@localhost:5433/oil_gas_inventory_local

.PHONY: help setup-phase35 migrate-phase35 seed-tenants db-status env-info debug-migrations

help:
	@echo "Oil & Gas Inventory System"
	@echo "=========================="
	@echo "Environment: $(or $(APP_ENV),local)"
	@echo "Database: $(or $(DB_NAME),unknown)"
	@echo "DATABASE_URL: $(DATABASE_URL)"
	@echo ""
	@echo "Commands:"
	@echo "  make env-info           - Show environment configuration"
	@echo "  make setup-phase35      - Complete Phase 3.5 setup"
	@echo "  make migrate-phase35    - Run tenant migrations"
	@echo "  make seed-tenants       - Seed tenant data"
	@echo "  make debug-migrations   - Debug migration status"

env-info:
	@echo "🔧 Environment Information"
	@echo "========================="
	@echo "APP_ENV: $(APP_ENV)"
	@echo "DB_NAME: $(or $(DB_NAME),$(POSTGRES_DB))"
	@echo "DB_HOST: $(or $(DB_HOST),$(POSTGRES_HOST))"
	@echo "DB_PORT: $(or $(DB_PORT),$(POSTGRES_PORT))"
	@echo "DB_USER: $(or $(DB_USER),$(POSTGRES_USER))"
	@echo "DATABASE_URL: $(DATABASE_URL)"

setup-phase35: migrate-phase35 seed-tenants
	@echo "🎯 Phase 3.5 setup completed!"
	@echo "✅ Tenant architecture implemented"
	@echo "✅ Default tenant 'Petros' created"

migrate-phase35:
	@echo "🗄️  Running Phase 3.5 tenant migrations..."
	@echo "Using DATABASE_URL: $(DATABASE_URL)"
	@echo "Checking if migration files exist..."
	@if [ ! -f "backend/migrations/001_add_tenant_architecture.sql" ]; then \
		echo "❌ Migration file backend/migrations/001_add_tenant_architecture.sql not found"; \
		exit 1; \
	fi
	@if [ ! -f "backend/migrations/002_tenant_rls_policies.sql" ]; then \
		echo "❌ Migration file backend/migrations/002_tenant_rls_policies.sql not found"; \
		exit 1; \
	fi
	@echo "Running migration 001_add_tenant_architecture.sql..."
	@psql "$(DATABASE_URL)" -f backend/migrations/001_add_tenant_architecture.sql || (echo "❌ Migration 003 failed" && exit 1)
	@echo "Running migration 002_tenant_rls_policies.sql..."
	@psql "$(DATABASE_URL)" -f backend/migrations/002_tenant_rls_policies.sql || (echo "❌ Migration 004 failed" && exit 1)
	@echo "Verifying tenant tables were created..."
	@psql "$(DATABASE_URL)" -c "SELECT 'tenants table exists' FROM store.tenants LIMIT 0;" || (echo "❌ Tenants table not created" && exit 1)
	@echo "✅ Phase 3.5 migrations completed successfully"

seed-tenants:
	@echo "🌱 Seeding tenant data..."
	@echo "Using DATABASE_URL: $(DATABASE_URL)"
	@echo "Checking if seed file exists..."
	@if [ ! -f "backend/seeds/tenant_seeds.sql" ]; then \
		echo "❌ Seed file backend/seeds/tenant_seeds.sql not found"; \
		exit 1; \
	fi
	@echo "Verifying tenants table exists before seeding..."
	@psql "$(DATABASE_URL)" -c "SELECT 'tenants table exists' FROM store.tenants LIMIT 0;" || (echo "❌ Tenants table missing - run migrate-phase35 first" && exit 1)
	@echo "Running tenant seeds..."
	@psql "$(DATABASE_URL)" -f backend/seeds/tenant_seeds.sql || (echo "❌ Tenant seeding failed" && exit 1)
	@echo "Verifying tenant data was inserted..."
	@psql "$(DATABASE_URL)" -c "SELECT 'Tenants: ' || count(*) FROM store.tenants;" || (echo "❌ Tenant verification failed" && exit 1)
	@echo "✅ Tenant seed data loaded successfully"

db-status:
	@echo "📊 Database Status:"
	@echo "Using DATABASE_URL: $(DATABASE_URL)"
	@psql "$(DATABASE_URL)" -c "SELECT 'Connected to database: ' || current_database() as status;" 2>/dev/null || echo "❌ Database connection failed"
	@echo ""
	@echo "📋 Tables in store schema:"
	@psql "$(DATABASE_URL)" -c "SELECT schemaname, tablename FROM pg_tables WHERE schemaname = 'store' ORDER BY tablename;" 2>/dev/null || echo "No tables found - run setup first"

# Development server
dev:
	@echo "🚀 Starting development server..."
	@echo "Environment: $(APP_ENV)"
	@echo "Database: $(DB_NAME)"
	@cd backend && go run cmd/server/main.go

# Clean up
clean:
	@echo "🧹 Cleaning temporary files..."
	@find . -name "*.tmp" -delete
	@find . -name ".DS_Store" -delete
EOF
    else
        echo -e "${YELLOW}📝 Makefile already exists. Checking if update needed...${NC}"
        
        # Check if Makefile has proper environment loading
        if ! grep -q "include .env.local" Makefile; then
            echo -e "${YELLOW}🔄 Updating existing Makefile with environment support...${NC}"
            
            # Backup existing Makefile
            cp Makefile Makefile.backup
            echo -e "${BLUE}💾 Backed up existing Makefile to Makefile.backup${NC}"
            
            # Create updated Makefile with environment support
            cat > Makefile << 'EOF'
# Oil & Gas Inventory System - Makefile
# Load environment variables from .env.local or .env
ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
else ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Fallback database URL if not set in environment
DATABASE_URL ?= postgresql://postgres:password@localhost:5434/oil_gas_inventory

.PHONY: help setup-phase35 migrate-phase35 seed-tenants db-status env-info

help:
	@echo "Oil & Gas Inventory System"
	@echo "=========================="
	@echo "Environment: $(or $(APP_ENV),local)"
	@echo "Database: $(or $(DB_NAME),unknown)"
	@echo "DATABASE_URL: $(DATABASE_URL)"
	@echo ""
	@echo "Commands:"
	@echo "  make env-info           - Show environment configuration"
	@echo "  make setup-phase35      - Complete Phase 3.5 setup"
	@echo "  make migrate-phase35    - Run tenant migrations"
	@echo "  make seed-tenants       - Seed tenant data"
	@echo "  make db-status          - Check database status"

env-info:
	@echo "🔧 Environment Information"
	@echo "========================="
	@echo "APP_ENV: $(APP_ENV)"
	@echo "DB_NAME: $(DB_NAME)"
	@echo "DB_HOST: $(DB_HOST)"
	@echo "DB_PORT: $(DB_PORT)"
	@echo "DB_USER: $(DB_USER)"
	@echo "DATABASE_URL: $(DATABASE_URL)"

setup-phase35: migrate-phase35 seed-tenants
	@echo "🎯 Phase 3.5 setup completed!"

migrate-phase35:
	@echo "🗄️  Running Phase 3.5 tenant migrations..."
	@echo "Using DATABASE_URL: $(DATABASE_URL)"
	@psql "$(DATABASE_URL)" -f backend/migrations/001_add_tenant_architecture.sql
	@psql "$(DATABASE_URL)" -f backend/migrations/002_tenant_rls_policies.sql
	@echo "✅ Phase 3.5 migrations completed"

seed-tenants:
	@echo "🌱 Seeding tenant data..."
	@psql "$(DATABASE_URL)" -f backend/seeds/tenant_seeds.sql
	@echo "✅ Tenant seed data loaded"

db-status:
	@echo "📊 Database Status:"
	@echo "Using DATABASE_URL: $(DATABASE_URL)"
	@psql "$(DATABASE_URL)" -c "SELECT 'Connected to database: ' || current_database() as status;"
	@echo ""
	@echo "📋 Tables in store schema:"
	@psql "$(DATABASE_URL)" -c "SELECT schemaname, tablename FROM pg_tables WHERE schemaname = 'store' ORDER BY tablename;" 2>/dev/null || echo "No tables found"

debug-migrations:
	@echo "🔍 Migration Debug Information"
	@echo "=============================="
	@echo "DATABASE_URL: $(DATABASE_URL)"
	@echo ""
	@echo "📁 Checking migration files:"
	@ls -la backend/migrations/ 2>/dev/null || echo "❌ backend/migrations/ directory not found"
	@echo ""
	@echo "📊 Current database tables:"
	@psql "$(DATABASE_URL)" -c "SELECT schemaname, tablename FROM pg_tables WHERE schemaname IN ('store', 'migrations') ORDER BY schemaname, tablename;" 2>/dev/null || echo "❌ Database connection failed"
	@echo ""
	@echo "📋 Migration history:"
	@psql "$(DATABASE_URL)" -c "SELECT version, name, executed_at FROM migrations.schema_migrations ORDER BY executed_at;" 2>/dev/null || echo "❌ No migration history found"
	@echo ""
	@echo "🏢 Checking for tenant tables specifically:"
	@psql "$(DATABASE_URL)" -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'store' AND table_name LIKE '%tenant%';" 2>/dev/null || echo "❌ No tenant tables found"
EOF
            echo -e "${GREEN}✅ Makefile updated with environment support${NC}"
        else
            echo -e "${GREEN}✅ Makefile already has environment support${NC}"
        fi
    fi
    
    echo -e "${GREEN}✅ Makefile configured to use your environment variables${NC}"
}

# Main execution flow
main() {
    echo -e "${BLUE}Starting Phase 3.5: Multi-Tenant Architecture Implementation${NC}"
    echo ""
    
    setup_environment
    setup_database  
    setup_basic_schema
    check_prerequisites
    create_tenant_migrations
    create_tenant_seeds
    update_makefile
    
    echo ""
    echo -e "${GREEN}🎉 Phase 3.5 Multi-Tenant Architecture Setup Complete!${NC}"
    echo "=============================================================="
    echo ""
    echo -e "${YELLOW}📋 What was implemented:${NC}"
    echo "✅ Environment configuration loaded from: $ENV_FILE"
    echo "✅ Database: $DB_NAME (verified/created)"
    echo "✅ Basic schema with all required tables"
    echo "✅ Multi-tenant database architecture"
    echo "✅ Row-Level Security (RLS) migration files"
    echo "✅ Tenant seed data"
    echo "✅ Updated Makefile with environment support"
    echo ""
    echo -e "${YELLOW}🚀 Next Steps:${NC}"
    echo "1. Run Phase 3.5 setup:       make setup-phase35"
    echo "2. Check database status:     make db-status"
    echo "3. Verify tenant setup works properly"
    echo ""
    echo -e "${GREEN}✅ Ready to run tenant setup!${NC}"
}

# Execute main function
main "$@"

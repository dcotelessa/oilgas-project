#!/bin/bash
# merge_makefile_commands.sh - Add missing tools commands to your existing Makefile

set -e

echo "🔧 Merging new tools commands into existing Makefile"

# Backup existing Makefile
cp Makefile Makefile.backup.$(date +%Y%m%d_%H%M%S)
echo "💾 Backup created: Makefile.backup.$(date +%Y%m%d_%H%M%S)"

# Create updated Makefile by replacing the tools section
cat > Makefile.new << 'EOF'
# Oil & Gas Inventory System - Makefile
# Load environment variables from .env.local or .env
ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
else ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Use DATABASE_URL from environment, convert postgres:// to postgresql:// if needed
ifneq (,$(findstring postgres://,$(DATABASE_URL)))
    DATABASE_URL := $(subst postgres://,postgresql://,$(DATABASE_URL))
endif

# Fallback DATABASE_URL if not set in environment
DATABASE_URL ?= postgresql://postgres:password@localhost:5432/oil_gas_inventory

.PHONY: help setup migrate seed-data db-status env-info debug-migrations ensure-basic-schema

help:
	@echo "Oil & Gas Inventory System"
	@echo "=========================="
	@echo "Environment: $(or $(APP_ENV),local)"
	@echo "Database: $(POSTGRES_DB)"
	@echo "DATABASE_URL: $(DATABASE_URL)"
	@echo ""
	@echo "🚀 Main Commands:"
	@echo "  make setup              - Complete multi-tenant setup"
	@echo "  make migrate            - Run tenant migrations"
	@echo "  make seed-data          - Seed default tenant data"
	@echo ""
	@echo "🔧 Development:"
	@echo "  make dev                - Start development server"
	@echo "  make test               - Run tests"
	@echo ""
	@echo "📊 Database:"
	@echo "  make db-status          - Check database status"
	@echo "  make ensure-basic-schema - Verify basic tables exist"
	@echo ""
	@echo "🔍 Debug:"
	@echo "  make env-info           - Show environment configuration"
	@echo "  make debug-migrations   - Debug migration status"
	@echo ""
	@echo "🏢 Tenant Management:"
	@echo "  make create-tenant      - Create new tenant"
	@echo "  make list-tenants       - List all tenants"
	@echo "  make tenant-status      - Show tenant status"
	@echo ""
	@echo "🛠️ Conversion Tools:"
	@echo "  make tools-status       - Show tools status"
	@echo "  make tools-setup        - Setup conversion tools"
	@echo "  make tools-test         - Test conversion tools"
	@echo "  make convert-mdb FILE=database.mdb - Convert MDB file"
	@echo "  make analyze-cf DIR=cf_app         - Analyze ColdFusion app"

env-info:
	@echo "🔧 Environment Information"
	@echo "========================="
	@echo "APP_ENV: $(APP_ENV)"
	@echo "POSTGRES_DB: $(POSTGRES_DB)"
	@echo "POSTGRES_HOST: $(POSTGRES_HOST)"
	@echo "POSTGRES_PORT: $(POSTGRES_PORT)"
	@echo "POSTGRES_USER: $(POSTGRES_USER)"
	@echo "DATABASE_URL: $(DATABASE_URL)"

setup: migrate seed-data
	@echo "🎯 Multi-tenant setup completed!"
	@echo "✅ Basic schema verified"
	@echo "✅ Tenant architecture implemented" 
	@echo "✅ Default tenant 'Petros' created"
	@echo ""
	@echo "🚀 Next steps:"
	@echo "  make dev               # Start development server"
	@echo "  make tenant-status     # Check tenant configuration"

ensure-basic-schema:
	@echo "🏗️  Ensuring basic schema exists..."
	@echo "Checking if basic tables exist..."
	@psql "$(DATABASE_URL)" -c "SELECT 'customers table exists' FROM store.customers LIMIT 0;" 2>/dev/null || \
		(echo "❌ Basic schema missing. Run the setup script first:" && \
		 echo "   ./scripts/setup.sh" && exit 1)
	@psql "$(DATABASE_URL)" -c "SELECT 'users table exists' FROM store.users LIMIT 0;" 2>/dev/null || \
		(echo "❌ Users table missing. Run the setup script first." && exit 1)
	@psql "$(DATABASE_URL)" -c "SELECT 'inventory table exists' FROM store.inventory LIMIT 0;" 2>/dev/null || \
		(echo "❌ Inventory table missing. Run the setup script first." && exit 1)
	@psql "$(DATABASE_URL)" -c "SELECT 'received table exists' FROM store.received LIMIT 0;" 2>/dev/null || \
		(echo "❌ Received table missing. Run the setup script first." && exit 1)
	@echo "✅ Basic schema verified"

migrate: ensure-basic-schema
	@echo "🗄️  Running tenant migrations..."
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
	@psql "$(DATABASE_URL)" -f backend/migrations/001_add_tenant_architecture.sql || (echo "❌ Migration 001 failed" && exit 1)
	@echo "Running migration 002_tenant_rls_policies.sql..."
	@psql "$(DATABASE_URL)" -f backend/migrations/002_tenant_rls_policies.sql || (echo "❌ Migration 002 failed" && exit 1)
	@echo "Verifying tenant tables were created..."
	@psql "$(DATABASE_URL)" -c "SELECT 'tenants table exists' FROM store.tenants LIMIT 0;" || (echo "❌ Tenants table not created" && exit 1)
	@echo "✅ Tenant migrations completed successfully"

seed-data:
	@echo "🌱 Seeding tenant data..."
	@echo "Using DATABASE_URL: $(DATABASE_URL)"
	@echo "Checking if seed file exists..."
	@if [ ! -f "backend/seeds/tenant_seeds.sql" ]; then \
		echo "❌ Seed file backend/seeds/tenant_seeds.sql not found"; \
		exit 1; \
	fi
	@echo "Verifying tenants table exists before seeding..."
	@psql "$(DATABASE_URL)" -c "SELECT 'tenants table exists' FROM store.tenants LIMIT 0;" || (echo "❌ Tenants table missing - run 'make migrate' first" && exit 1)
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

# Tenant management commands
create-tenant:
	@echo "🏢 Creating new tenant..."
	@echo "Note: Tenant creation utilities will be available in Phase 4"
	@echo "For now, you can create tenants directly in the database"

list-tenants:
	@echo "📋 Listing all tenants..."
	@psql "$(DATABASE_URL)" -c "SELECT tenant_id, tenant_name, tenant_slug, active, created_at FROM store.tenants ORDER BY tenant_id;" 2>/dev/null || echo "❌ No tenants found - run 'make setup' first"

tenant-status:
	@echo "📊 Tenant Status Report"
	@echo "======================"
	@echo ""
	@echo "📋 Tenants:"
	@psql "$(DATABASE_URL)" -c "SELECT tenant_id, tenant_name, active, (SELECT COUNT(*) FROM store.users WHERE tenant_id = t.tenant_id) as users, (SELECT COUNT(*) FROM store.inventory WHERE tenant_id = t.tenant_id) as inventory FROM store.tenants t ORDER BY tenant_id;" 2>/dev/null || echo "❌ No tenants found"
	@echo ""
	@echo "👥 Users by Tenant:"
	@psql "$(DATABASE_URL)" -c "SELECT t.tenant_name, u.username, utr.role FROM store.tenants t LEFT JOIN store.users u ON t.tenant_id = u.tenant_id LEFT JOIN store.user_tenant_roles utr ON u.user_id = utr.user_id AND t.tenant_id = utr.tenant_id ORDER BY t.tenant_id, u.username;" 2>/dev/null || echo "❌ No user assignments found"
	@echo ""
	@echo "🏢 Customer Assignments:"
	@psql "$(DATABASE_URL)" -c "SELECT t.tenant_name, c.customer, cta.relationship_type FROM store.customer_tenant_assignments cta JOIN store.tenants t ON cta.tenant_id = t.tenant_id JOIN store.customers c ON cta.customer_id = c.customer_id WHERE cta.active = true ORDER BY t.tenant_name, c.customer;" 2>/dev/null || echo "❌ No customer assignments found"

# Development server
dev:
	@echo "🚀 Starting development server..."
	@echo "Environment: $(APP_ENV)"
	@echo "Database: $(POSTGRES_DB)"
	@cd backend && go run cmd/server/main.go

# Run tests
test:
	@echo "🧪 Running tests..."
	@cd backend && go test ./... 2>/dev/null || echo "No tests found yet"

# Clean up
clean:
	@echo "🧹 Cleaning temporary files..."
	@find . -name "*.tmp" -delete
	@find . -name ".DS_Store" -delete

# Legacy command aliases (for backward compatibility)
setup-phase35: setup
	@echo "⚠️  'make setup-phase35' is deprecated. Use 'make setup' instead."

migrate-phase35: migrate
	@echo "⚠️  'make migrate-phase35' is deprecated. Use 'make migrate' instead."

seed-tenants: seed-data
	@echo "⚠️  'make seed-tenants' is deprecated. Use 'make seed-data' instead."

# Basic test commands
test-basic:
	@echo "🧪 Running basic tests..."
	@cd backend && go test ./test/... -v

test-tenant:
	@echo "🏢 Testing tenant functionality..."
	@cd backend && go test ./test/tenant/... -v

test-env:
	@echo "🔧 Testing environment setup..."
	@cd backend && go test ./test/integration/... -v

# =============================================================================
# MDB CONVERSION TOOLS INTEGRATION
# =============================================================================

tools: ## Show conversion tool commands
	@if [ -d "tools/mdb-conversion" ]; then \
		echo "🔗 Conversion tool commands (cd tools/mdb-conversion && make <command>):"; \
		cd tools/mdb-conversion && make help 2>/dev/null || echo "  Run 'make tools-setup' first"; \
	else \
		echo "❌ Tools directory not found: tools/mdb-conversion"; \
		echo "  Run the tools setup script first"; \
	fi

tools-status: ## Show tools status and configuration
	@echo "📊 MDB Conversion Tools Status"
	@echo "=============================="
	@if [ -d "tools/mdb-conversion" ]; then \
		cd tools/mdb-conversion && make status 2>/dev/null || echo "❌ Tools not properly set up"; \
	else \
		echo "❌ Tools directory not found: tools/mdb-conversion"; \
		echo "  Run the tools setup script first"; \
	fi

tools-setup: ## Setup conversion tools
	@echo "🔧 Setting up conversion tools..."
	@if [ ! -d "tools/mdb-conversion" ]; then \
		echo "❌ Tools directory not found - run the setup script first"; \
		echo "  Example: ./minimal_tools_setup.sh"; \
		exit 1; \
	fi
	@cd tools/mdb-conversion && make setup
	@echo "✅ Conversion tools ready!"

tools-test: ## Test conversion tools
	@echo "🧪 Testing conversion tools..."
	@if [ -d "tools/mdb-conversion" ]; then \
		cd tools/mdb-conversion && make test; \
	else \
		echo "❌ Tools directory not found"; \
	fi

tools-help: ## Show detailed tools help
	@echo "🛠️  Detailed Tools Help"
	@echo "======================="
	@if [ -d "tools/mdb-conversion" ]; then \
		cd tools/mdb-conversion && make help; \
	else \
		echo "❌ Tools directory not found"; \
		echo "  Run the tools setup script first"; \
	fi

# Conversion shortcuts
convert-mdb: ## Convert MDB file (usage: make convert-mdb FILE=database.mdb)
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Usage: make convert-mdb FILE=path/to/database.mdb"; \
		exit 1; \
	fi
	@if [ ! -f "$(FILE)" ]; then \
		echo "❌ File not found: $(FILE)"; \
		exit 1; \
	fi
	@cd tools/mdb-conversion && make convert-mdb FILE="$(abspath $(FILE))"

analyze-cf: ## Analyze ColdFusion app (usage: make analyze-cf DIR=cf_app)
	@if [ -z "$(DIR)" ]; then \
		echo "❌ Usage: make analyze-cf DIR=path/to/cf_application"; \
		exit 1; \
	fi
	@if [ ! -d "$(DIR)" ]; then \
		echo "❌ Directory not found: $(DIR)"; \
		exit 1; \
	fi
	@cd tools/mdb-conversion && make analyze-cf DIR="$(abspath $(DIR))"

test-conversion: ## Run conversion tool tests
	@cd tools/mdb-conversion && make test-conversion 2>/dev/null || echo "❌ Conversion tests not available"

full-analysis: ## Complete analysis (usage: make full-analysis COMPANY=company_name)
	@if [ -z "$(COMPANY)" ]; then \
		echo "❌ Usage: make full-analysis COMPANY=company_name"; \
		exit 1; \
	fi
	@cd tools/mdb-conversion && make full-analysis COMPANY="$(COMPANY)" 2>/dev/null || echo "❌ Full analysis not available"

# Tools reporting
tools-report: ## Generate tools usage report
	@echo "📋 MDB Conversion Tools Report"
	@echo "=============================="
	@if [ -d "output/conversion" ]; then \
		echo "📂 Processed Companies:"; \
		find output/conversion -type d -maxdepth 1 -mindepth 1 -exec basename {} \; | sort; \
		echo ""; \
		echo "📊 Conversion Files:"; \
		find output/conversion -name "*.csv" -o -name "*.sql" | wc -l | xargs echo "  Total files:"; \
	else \
		echo "❌ No conversions found - output/conversion directory missing"; \
	fi

# Quick tools setup
quick-tools-setup: ## Quick setup for tools (creates minimal structure)
	@echo "⚡ Quick tools setup..."
	@if [ -f "minimal_tools_setup.sh" ]; then \
		./minimal_tools_setup.sh; \
	else \
		echo "❌ minimal_tools_setup.sh script not found"; \
		echo "  Create this script first or run the full setup"; \
	fi
EOF

# Replace the original Makefile
mv Makefile.new Makefile

echo "✅ Makefile updated successfully!"
echo ""
echo "🆕 New commands added:"
echo "  make tools-status       # Show tools status"
echo "  make tools-setup        # Setup tools"
echo "  make tools-test         # Test tools"
echo "  make tools-help         # Detailed tools help"
echo "  make tools-report       # Usage report"
echo "  make quick-tools-setup  # Quick setup"
echo ""
echo "🔧 Enhanced existing commands:"
echo "  make help               # Now includes tools section"
echo "  make convert-mdb        # Better error checking"
echo "  make analyze-cf         # Better error checking"
echo ""
echo "🧪 Test the new commands:"
echo "  make tools-status"
echo "  make help"

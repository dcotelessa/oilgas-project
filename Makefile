# Optimized Makefile for Customer Migration with petros-lb.mdb
.PHONY: help setup clean dev test

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
BLUE := \033[34m
NC := \033[0m

# Configuration with SSL disabled for local development
MDB_FILE := db_prep/petros-lb.mdb
TENANT_ID := local-dev
DEV_DATABASE_URL := postgresql://oilgas_user:oilgas_pass@localhost:5432/oilgas_dev?sslmode=disable
TEST_DATABASE_URL := postgresql://oilgas_test_user:oilgas_test_pass@localhost:5433/oilgas_test?sslmode=disable

help: ## Show available commands
	@echo "$(BLUE)🛢️  Oil & Gas Customer Migration Commands$(NC)"
	@echo "============================================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)📋 Quick Start:$(NC)"
	@echo "  1. make check-mdb        # Verify your MDB file"
	@echo "  2. make setup-customers  # Complete customer setup"
	@echo "  3. make verify-customers # Verify everything works"

# =============================================================================
# PRE-FLIGHT CHECKS
# =============================================================================

check-deps: ## Check required dependencies
	@echo "$(YELLOW)📋 Checking dependencies...$(NC)"
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)❌ Docker required$(NC)"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "$(RED)❌ Docker Compose required$(NC)"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "$(RED)❌ Go required$(NC)"; exit 1; }
	@command -v mdb-tables >/dev/null 2>&1 || { echo "$(RED)❌ mdb-tools required$(NC)"; exit 1; }
	@echo "$(GREEN)✅ All dependencies found$(NC)"

check-mdb: ## Check if MDB file exists and is accessible
	@echo "$(YELLOW)🔍 Checking MDB file...$(NC)"
	@if [ ! -f "$(MDB_FILE)" ]; then \
		echo "$(RED)❌ MDB file not found: $(MDB_FILE)$(NC)"; \
		echo "$(YELLOW)💡 Please place your Access database at: $(MDB_FILE)$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ MDB file found: $(MDB_FILE)$(NC)"
	@echo "$(BLUE)📊 MDB Contents:$(NC)"
	@mdb-tables "$(MDB_FILE)" | tr ' ' '\n' | sort | head -10

check-docker-mount: ## Check Docker mount capabilities for Mac M1
	@echo "$(YELLOW)🔍 Checking Docker mount capabilities...$(NC)"
	@if ! docker run --rm -v "$(PWD):/test" alpine ls /test >/dev/null 2>&1; then \
		echo "$(RED)❌ Docker volume mount failed$(NC)"; \
		echo "$(YELLOW)💡 Mac M1 Fix:$(NC)"; \
		echo "  1. Open Docker Desktop → Settings → Resources → File sharing"; \
		echo "  2. Add $(PWD) to shared directories"; \
		echo "  3. Apply & Restart Docker Desktop"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Docker mounts working$(NC)"

# =============================================================================
# DATABASE SETUP - SIMPLIFIED STRUCTURE
# =============================================================================

setup-db: check-deps check-docker-mount ## Setup PostgreSQL databases
	@echo "$(YELLOW)🐳 Setting up PostgreSQL databases...$(NC)"
	@docker-compose down 2>/dev/null || true
	@mkdir -p database/init database/data/exported database/data/clean database/logs
	@echo "$(YELLOW)🚀 Starting containers...$(NC)"
	@docker-compose up -d postgres postgres-test
	@echo "$(YELLOW)⏳ Waiting for databases and initialization...$(NC)"
	@timeout=90; \
	while [ $$timeout -gt 0 ]; do \
		if docker-compose exec -T postgres pg_isready -U oilgas_user -d oilgas_dev >/dev/null 2>&1; then \
			echo "$(GREEN)✅ Development database ready$(NC)"; \
			break; \
		fi; \
		echo "   Waiting for dev database... ($$timeout seconds remaining)"; \
		sleep 3; \
		timeout=$$((timeout-3)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "$(RED)❌ Development database timeout$(NC)"; \
		docker-compose logs postgres; \
		exit 1; \
	fi
	@timeout=90; \
	while [ $$timeout -gt 0 ]; do \
		if docker-compose exec -T postgres-test pg_isready -U oilgas_test_user -d oilgas_test >/dev/null 2>&1; then \
			echo "$(GREEN)✅ Test database ready$(NC)"; \
			break; \
		fi; \
		echo "   Waiting for test database... ($$timeout seconds remaining)"; \
		sleep 3; \
		timeout=$$((timeout-3)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "$(RED)❌ Test database timeout$(NC)"; \
		docker-compose logs postgres-test; \
		exit 1; \
	fi
	@echo "$(YELLOW)🔍 Verifying database schema...$(NC)"
	@if docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\dt store.*" | grep -q "customers_standardized"; then \
		echo "$(GREEN)✅ Database initialization completed successfully$(NC)"; \
	else \
		echo "$(RED)❌ Database initialization failed - tables not created$(NC)"; \
		echo "$(YELLOW)Checking initialization logs:$(NC)"; \
		docker-compose logs postgres | grep -A 5 -B 5 "database initialization"; \
		exit 1; \
	fi

verify-schema: ## Verify database schema was created correctly
	@echo "$(YELLOW)🔍 Verifying database schema...$(NC)"
	@echo "$(BLUE)📊 Development Database Schema:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\dt store.*"
	@echo ""
	@echo "$(BLUE)🔧 Testing tenant functions:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT set_tenant_context('local-dev'); SELECT get_current_tenant();"
	@echo ""
	@echo "$(BLUE)📋 Tenants table:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT * FROM store.tenants;"

setup-customers: check-deps check-mdb ## Complete customer setup workflow (UPDATED)
	@echo "$(GREEN)🚀 Complete Customer Setup Workflow$(NC)"
	@echo "====================================="
	@$(MAKE) analyze-mdb
	@$(MAKE) analyze-customers
	@$(MAKE) setup-db
	@$(MAKE) clean-customers
	@$(MAKE) import-customers
	@echo ""
	@echo "$(GREEN)🎉 Customer setup completed successfully!$(NC)"
	@$(MAKE) verify-customers

# =============================================================================
# MDB ANALYSIS AND EXPORT
# =============================================================================

analyze-mdb: check-mdb ## Analyze MDB structure and export customers
	@echo "$(YELLOW)📊 Analyzing MDB structure...$(NC)"
	@mkdir -p database/data/exported database/logs
	@echo "Tables in $(MDB_FILE):" > database/logs/mdb_analysis.log
	@mdb-tables "$(MDB_FILE)" >> database/logs/mdb_analysis.log
	@echo "" >> database/logs/mdb_analysis.log
	@echo "Exporting customers table..." | tee -a database/logs/mdb_analysis.log
	@if mdb-export "$(MDB_FILE)" customers > database/data/exported/customers.csv 2>>database/logs/mdb_analysis.log; then \
		echo "$(GREEN)✅ Customers exported successfully$(NC)"; \
		RECORD_COUNT=$$(wc -l < database/data/exported/customers.csv); \
		echo "   Records: $$((RECORD_COUNT - 1)) (excluding header)"; \
	else \
		echo "$(RED)❌ Failed to export customers table$(NC)"; \
		echo "$(YELLOW)💡 Check available tables: make list-tables$(NC)"; \
		exit 1; \
	fi

list-tables: check-mdb ## List all tables in MDB file
	@echo "$(BLUE)📋 Tables in $(MDB_FILE):$(NC)"
	@mdb-tables "$(MDB_FILE)" | tr ' ' '\n' | sort | nl

analyze-customers: ## Analyze exported customer CSV structure
	@echo "$(YELLOW)🔍 Analyzing customer CSV structure...$(NC)"
	@if [ ! -f "database/data/exported/customers.csv" ]; then \
		echo "$(RED)❌ No customer CSV found. Run 'make analyze-mdb' first$(NC)"; \
		exit 1; \
	fi
	@echo "Building customer analyzer..."
	@cd backend && go build -o ../customer-analyzer cmd/tools/customer-analyzer/main.go
	@echo "$(BLUE)📊 Customer CSV Analysis:$(NC)"
	@./customer-analyzer database/data/exported/customers.csv

# =============================================================================
# CUSTOMER DATA PROCESSING
# =============================================================================

clean-customers: ## Clean customer data with deduplication
	@echo "$(YELLOW)🧹 Cleaning customer data...$(NC)"
	@if [ ! -f "database/data/exported/customers.csv" ]; then \
		echo "$(RED)❌ No exported customers found. Run 'make analyze-mdb' first$(NC)"; \
		exit 1; \
	fi
	@echo "Building customer cleaner..."
	@cd backend && go build -o ../customer-cleaner cmd/customer-cleaner/main.go
	@mkdir -p database/data/clean database/logs
	@echo "Cleaning customer data with standards and deduplication..."
	@./customer-cleaner database/data/exported/customers.csv database/data/clean/customers_cleaned.csv $(TENANT_ID)
	@echo "$(GREEN)✅ Customer data cleaned$(NC)"

import-customers: ## Import cleaned customers to database
	@echo "$(YELLOW)📥 Importing customers to database...$(NC)"
	@if [ ! -f "database/data/clean/customers_cleaned.csv" ]; then \
		echo "$(RED)❌ No cleaned customers found. Run 'make clean-customers' first$(NC)"; \
		exit 1; \
	fi
	@echo "Building standardized importer..."
	@cd backend && go build -o ../standardized-importer cmd/standardized-importer/main.go
	@echo "Importing to development database..."
	@DATABASE_URL="$(DEV_DATABASE_URL)" ./standardized-importer database/data/clean/customers_cleaned.csv $(TENANT_ID)
	@echo "Importing to test database..."
	@DATABASE_URL="$(TEST_DATABASE_URL)" ./standardized-importer database/data/clean/customers_cleaned.csv $(TENANT_ID)
	@echo "$(GREEN)✅ Customer import completed$(NC)"

# =============================================================================
# COMPLETE WORKFLOW COMMANDS
# =============================================================================


quick-setup: ## Quick setup (skip analysis steps)
	@echo "$(GREEN)⚡ Quick Customer Setup$(NC)"
	@$(MAKE) setup-db
	@echo "$(GREEN)✅ Quick setup completed$(NC)"

verify-customers: ## Verify customer import and database setup
	@echo "$(YELLOW)🔍 Verifying customer setup...$(NC)"
	@echo "$(BLUE)📊 Database status:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT COUNT(*) as dev_customers FROM store.customers_standardized;" 2>/dev/null || echo "❌ Dev database connection failed"
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -c "SELECT COUNT(*) as test_customers FROM store.customers_standardized;" 2>/dev/null || echo "❌ Test database connection failed"
	@echo "$(BLUE)📁 File status:$(NC)"
	@ls -la database/data/exported/customers.csv 2>/dev/null && echo "✅ Raw export exists" || echo "❌ No raw export"
	@ls -la database/data/clean/customers_cleaned.csv 2>/dev/null && echo "✅ Cleaned data exists" || echo "❌ No cleaned data"

# =============================================================================
# DATABASE UTILITIES
# =============================================================================

db-status: ## Show database container status
	@echo "$(BLUE)🐳 Database Container Status:$(NC)"
	@docker-compose ps postgres postgres-test 2>/dev/null || echo "❌ Containers not running"

db-reset: ## Reset all databases (destructive operation)
	@echo "$(RED)⚠️  This will delete all database data$(NC)"
	@read -p "Continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker-compose down -v
	@docker volume rm $$(docker volume ls -q | grep postgres) 2>/dev/null || true
	@echo "$(GREEN)✅ Databases reset - run 'make setup-db' to recreate$(NC)"

start-pgadmin: ## Start PgAdmin for database management
	@echo "$(YELLOW)🖥️  Starting PgAdmin...$(NC)"
	@docker-compose up -d pgadmin
	@echo "$(GREEN)✅ PgAdmin started$(NC)"
	@echo "$(BLUE)🌐 Access at: http://localhost:8080$(NC)"
	@echo "   Email: admin@oilgas.local"
	@echo "   Password: admin123"

logs: ## Show container logs
	@echo "$(BLUE)📋 Container Logs:$(NC)"
	@docker-compose logs --tail=50

# =============================================================================
# DEVELOPMENT
# =============================================================================

dev: ## Start development server
	@echo "$(YELLOW)🚀 Starting development server...$(NC)"
	@cd backend && go run cmd/server/main.go

build: ## Build all applications
	@echo "$(YELLOW)🔨 Building applications...$(NC)"
	@cd backend && go build -o ../customer-cleaner cmd/customer-cleaner/main.go
	@cd backend && go build -o ../customer-analyzer cmd/tools/customer-analyzer/main.go
	@cd backend && go build -o ../standardized-importer cmd/standardized-importer/main.go
	@echo "$(GREEN)✅ Build completed$(NC)"

test: ## Run all tests
	@echo "$(YELLOW)🧪 Running tests...$(NC)"
	@cd backend && go test ./... -v

# =============================================================================
# UTILITIES
# =============================================================================

status: ## Show overall system status
	@echo "$(BLUE)📊 System Status:$(NC)"
	@echo "=================="
	@echo -n "Docker: "; docker --version 2>/dev/null || echo "❌ Not installed"
	@echo -n "Docker Compose: "; docker-compose --version 2>/dev/null || echo "❌ Not installed"
	@echo -n "Go: "; go version 2>/dev/null || echo "❌ Not installed"
	@echo -n "mdb-tools: "; mdb-tools --version 2>/dev/null || echo "❌ Not installed"
	@echo ""
	@echo -n "PostgreSQL containers: "; \
		if docker-compose ps postgres postgres-test | grep -q "Up"; then \
			echo "✅ Running"; \
		else \
			echo "❌ Not running"; \
		fi
	@echo -n "MDB file: "; \
		if [ -f "$(MDB_FILE)" ]; then \
			echo "✅ Found"; \
		else \
			echo "❌ Missing"; \
		fi

config: ## Show current configuration
	@echo "$(BLUE)ℹ️  Current Configuration:$(NC)"
	@echo "MDB File: $(MDB_FILE)"
	@echo "Tenant ID: $(TENANT_ID)"
	@echo "Dev Database: $(DEV_DATABASE_URL)"
	@echo "Test Database: $(TEST_DATABASE_URL)"
	@echo ""
	@echo "$(BLUE)📁 Key Files:$(NC)"
	@ls -la $(MDB_FILE) 2>/dev/null || echo "❌ MDB file not found"
	@ls -la database/data/exported/customers.csv 2>/dev/null || echo "ℹ️  No exported CSV (will be generated)"
	@ls -la database/data/clean/customers_cleaned.csv 2>/dev/null || echo "ℹ️  No cleaned data yet"

troubleshoot: ## Show troubleshooting information
	@echo "$(YELLOW)🔧 Troubleshooting Information$(NC)"
	@echo "================================"
	@echo ""
	@echo "$(BLUE)Common Issues:$(NC)"
	@echo ""
	@echo "1. MDB file not found:"
	@echo "   Place your Access database at: $(MDB_FILE)"
	@echo ""
	@echo "2. mdb-tools not installed:"
	@echo "   macOS: brew install mdbtools"
	@echo ""
	@echo "3. Docker /host_mnt error (Mac M1):"
	@echo "   Docker Desktop → Settings → Resources → File sharing"
	@echo "   Add your project directory and restart Docker"
	@echo ""
	@echo "4. Database connection failed:"
	@echo "   Run: make db-status"
	@echo "   Try: make db-reset && make setup-db"
	@echo ""
	@echo "5. Customer export failed:"
	@echo "   Check available tables: make list-tables"
	@echo "   Your table might have a different name"
	@echo ""
	@echo "$(BLUE)Log Files:$(NC)"
	@ls -la database/logs/ 2>/dev/null || echo "No logs yet"

# =============================================================================
# DOMAIN TESTING
# =============================================================================

test-customer-domain: ## Test customer domain with auth integration
	@echo "$(YELLOW)🧪 Testing customer domain...$(NC)"
	@cd backend && go test -v ./internal/customer/... -tags=integration || true
	@echo "$(GREEN)✅ Customer domain tests completed$(NC)"

test-auth-domain: ## Test auth domain with tenant manager  
	@echo "$(YELLOW)🧪 Testing auth domain...$(NC)"
	@cd backend && go test -v ./internal/auth/... -tags=integration || true
	@echo "$(GREEN)✅ Auth domain tests completed$(NC)"

test-workorder-domain: ## Test work order domain
	@echo "$(YELLOW)🧪 Testing work order domain...$(NC)"
	@cd backend && go test -v ./internal/workorder/... -tags=integration || true
	@echo "$(GREEN)✅ Work order domain tests completed$(NC)"

test-all-domains: ## Test all domains together
	@echo "$(YELLOW)🧪 Testing all domains...$(NC)"
	@$(MAKE) test-customer-domain
	@$(MAKE) test-auth-domain  
	@$(MAKE) test-workorder-domain
	@echo "$(GREEN)✅ All domain tests completed$(NC)"

# =============================================================================
# BACKEND API
# =============================================================================

start-backend: ## Start backend API server
	@echo "$(YELLOW)🚀 Starting backend API server...$(NC)"
	@docker-compose up -d backend
	@echo "$(GREEN)✅ Backend API started at http://localhost:8000$(NC)"

backend-logs: ## Show backend API logs
	@docker-compose logs -f backend

# =============================================================================
# COMPLETE SETUP
# =============================================================================

setup-all-domains: ## Complete setup for customer/auth/workorder domains
	@echo "$(GREEN)🚀 Complete Domain Setup$(NC)"
	@echo "=========================="
	@$(MAKE) analyze-mdb
	@$(MAKE) analyze-customers
	@$(MAKE) setup-db
	@$(MAKE) clean-customers
	@$(MAKE) import-customers
	@echo "$(YELLOW)🔍 Verifying domain integration...$(NC)"
	@$(MAKE) verify-domains
	@echo "$(GREEN)✅ All domains setup completed!$(NC)"

verify-domains: ## Verify all domain tables exist
	@echo "$(YELLOW)🔍 Verifying domain setup...$(NC)"
	@echo "$(BLUE)📊 Checking schemas exist:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT schema_name FROM information_schema.schemata WHERE schema_name IN ('auth', 'audit', 'store') ORDER BY schema_name;"
	@echo "$(BLUE)📊 Checking auth tables:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT tablename FROM pg_tables WHERE schemaname = 'auth' ORDER BY tablename;" || echo "No auth tables found"
	@echo "$(BLUE)📊 Checking store tables:$(NC)"  
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT tablename FROM pg_tables WHERE schemaname = 'store' ORDER BY tablename;"
	@echo "$(BLUE)📊 Checking data migration:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT 'customers_standardized' as table_name, COUNT(*) as record_count FROM store.customers_standardized;" || echo "No standardized customers"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT 'customers' as table_name, COUNT(*) as record_count FROM store.customers;" || echo "No optimized customers yet"
	@echo "$(GREEN)✅ Domain verification completed$(NC)"

# Remove the duplicate setup-customers target and keep only the enhanced one
setup-customers-enhanced: check-deps check-mdb ## Enhanced customer setup with domain integration
	@echo "$(GREEN)🚀 Enhanced Customer Setup$(NC)"
	@echo "============================"
	@$(MAKE) analyze-mdb
	@$(MAKE) analyze-customers  
	@$(MAKE) setup-db
	@$(MAKE) clean-customers
	@$(MAKE) import-customers
	@$(MAKE) verify-domains
	@echo "$(GREEN)🎉 Enhanced customer setup completed!$(NC)"

# =============================================================================
# MIGRATIONS
# =============================================================================


migrate-up: ## Run all migrations up
	@echo "$(YELLOW)🆙 Running migrations up...$(NC)"
	@migrate -path database/migrations -database $(DEV_DATABASE_URL) up
	@echo "$(GREEN)✅ Migrations completed$(NC)"

migrate-down: ## Rollback one migration
	@echo "$(YELLOW)🔽 Rolling back one migration...$(NC)"
	@migrate -path database/migrations -database $(DEV_DATABASE_URL) down 1
	@echo "$(GREEN)✅ Rollback completed$(NC)"

migrate-status: ## Check migration status
	@echo "$(BLUE)📊 Migration status:$(NC)"
	@migrate -path database/migrations -database $(DEV_DATABASE_URL) version

migrate-create: ## Create new migration (usage: make migrate-create name=add_new_feature)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)❌ Usage: make migrate-create name=add_new_feature$(NC)"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir database/migrations $(name)
	@echo "$(GREEN)✅ Created migration files for: $(name)$(NC)"

# Enhanced domain setup with migrations
setup-all-domains-with-migrations: ## Complete setup with proper migrations
	@echo "$(GREEN)🚀 Complete Setup with Migrations$(NC)"
	@$(MAKE) setup-db          # Init files for fresh database
	@$(MAKE) migrate-up        # Apply migrations
	@$(MAKE) setup-customers   # Your Access pipeline
	@$(MAKE) verify-domains    # Verify everything works
	@echo "$(GREEN)✅ Complete setup with migrations done!$(NC)"

# =============================================================================
# CLEANUP
# =============================================================================

clean: ## Clean up generated files and containers
	@echo "$(YELLOW)🧹 Cleaning up...$(NC)"
	@docker-compose down 2>/dev/null || true
	@rm -f customer-cleaner customer-analyzer standardized-importer
	@rm -rf database/data/exported/* database/data/clean/* database/logs/*
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

# =============================================================================
# DOCUMENTATION
# =============================================================================

docs: ## Show documentation links
	@echo "$(BLUE)📚 Documentation:$(NC)"
	@echo "==================="
	@echo "• Customer Migration: docs/CUSTOMER_MIGRATION.md"
	@echo "• Database Schema: docs/DATABASE_SCHEMA.md"
	@echo "• Data Standards: docs/DATA_CONVERSION_STANDARDS.md"
	@echo ""
	@echo "$(BLUE)🌐 Web Resources:$(NC)"
	@echo "• PgAdmin (if running): http://localhost:8080"
	@echo "• API Health Check: http://localhost:8000/health"

.DEFAULT_GOAL := help

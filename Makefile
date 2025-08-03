# Makefile - Enhanced with MDB to PostgreSQL migration support
.PHONY: help setup clean dev test migrate seed

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
NC := \033[0m

# Environment configuration (eventually in .env.local or..)
DEV_DATABASE_URL ?= postgresql://oilgas_user:oilgas_pass@localhost:5432/oilgas_dev
TEST_DATABASE_URL ?= postgresql://oilgas_test_user:oilgas_test_pass@localhost:5433/oilgas_test
MDB_FILE ?= db_prep/petros.mdb
DATA_PATH ?= database/data/clean
TENANT_ID ?= local-dev

help: ## Show this help message
	@echo "Oil & Gas Inventory System - Migration Commands"
	@echo "=============================================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Setup and Infrastructure
# =============================================================================

setup: ## Set up complete development environment
	@echo "$(GREEN)🚀 Setting up Oil & Gas Inventory System...$(NC)"
	@$(MAKE) check-dependencies
	@$(MAKE) setup-docker
	@$(MAKE) wait-for-db
	@$(MAKE) migrate-db
	@$(MAKE) setup-complete

check-dependencies: ## Check required dependencies
	@echo "$(YELLOW)📋 Checking dependencies...$(NC)"
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)❌ Docker is required but not installed$(NC)"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "$(RED)❌ Docker Compose is required but not installed$(NC)"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "$(RED)❌ Go is required but not installed$(NC)"; exit 1; }
	@echo "$(GREEN)✅ All dependencies found$(NC)"

setup-docker: ## Start Docker containers
	@echo "$(YELLOW)🐳 Starting Docker containers...$(NC)"
	@docker-compose up -d postgres postgres-test
	@echo "$(GREEN)✅ PostgreSQL containers started$(NC)"

setup-docker-full: ## Start all Docker containers including PgAdmin
	@echo "$(YELLOW)🐳 Starting all Docker containers...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)✅ All containers started$(NC)"
	@echo "$(GREEN)🎯 Access points:$(NC)"
	@echo "  • PostgreSQL Dev:  localhost:5432"
	@echo "  • PostgreSQL Test: localhost:5433" 
	@echo "  • PgAdmin:         http://localhost:8080"

wait-for-db: ## Wait for databases to be ready
	@echo "$(YELLOW)⏳ Waiting for databases to be ready...$(NC)"
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if docker-compose exec -T postgres pg_isready -U oilgas_user -d oilgas_dev >/dev/null 2>&1 && \
		   docker-compose exec -T postgres-test pg_isready -U oilgas_test_user -d oilgas_test >/dev/null 2>&1; then \
			echo "$(GREEN)✅ Databases are ready$(NC)"; \
			break; \
		fi; \
		echo "Waiting for databases... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "$(RED)❌ Databases failed to start within timeout$(NC)"; \
		exit 1; \
	fi

# =============================================================================
# Database Migration
# =============================================================================

migrate-db: ## Run database migrations
	@echo "$(YELLOW)📊 Running database migrations...$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -f /docker-entrypoint-initdb.d/001_init_database.sql
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -f /docker-entrypoint-initdb.d/migrations/001_create_customers.sql
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -f /docker-entrypoint-initdb.d/001_init_database.sql
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -f /docker-entrypoint-initdb.d/migrations/001_create_customers.sql
	@echo "$(GREEN)✅ Database migrations completed$(NC)"

# =============================================================================
# MDB Migration (Phase 1 to PostgreSQL)
# =============================================================================

check-mdb: ## Check if MDB file exists and is accessible
	@echo "$(YELLOW)🔍 Checking MDB file...$(NC)"
	@if [ ! -f "$(MDB_FILE)" ]; then \
		echo "$(RED)❌ MDB file not found: $(MDB_FILE)$(NC)"; \
		echo "$(YELLOW)💡 Please ensure your Access database is at: $(MDB_FILE)$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ MDB file found: $(MDB_FILE)$(NC)"
	@if command -v mdb-tables >/dev/null 2>&1; then \
		echo "$(GREEN)✅ mdb-tools available$(NC)"; \
	else \
		echo "$(RED)❌ mdb-tools not installed$(NC)"; \
		echo "$(YELLOW)💡 Install with: brew install mdbtools (macOS) or apt-get install mdb-tools (Ubuntu)$(NC)"; \
		exit 1; \
	fi

phase1-export: check-mdb ## Export MDB data to CSV (Phase 1)
	@echo "$(YELLOW)📤 Exporting MDB data to CSV...$(NC)"
	@mkdir -p database/data/exported database/data/clean database/logs
	@echo "Exporting customers table..."
	@if mdb-export "$(MDB_FILE)" customers > database/data/exported/customers.csv 2>/dev/null; then \
		echo "$(GREEN)✅ Customers exported$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  Customers table not found or empty$(NC)"; \
	fi
	@echo "$(GREEN)✅ MDB export completed$(NC)"

phase1-clean: ## Clean and normalize exported CSV data
	@echo "$(YELLOW)🧹 Cleaning and normalizing CSV data...$(NC)"
	@go run backend/cmd/tools/csv-cleaner/main.go \
		-input database/data/exported \
		-output database/data/clean \
		-log database/logs/cleaning.log
	@echo "$(GREEN)✅ Data cleaning completed$(NC)"

phase1-complete: ## Complete Phase 1 migration (export + clean)
	@echo "$(GREEN)🎯 Running complete Phase 1 migration...$(NC)"
	@$(MAKE) phase1-export
	@$(MAKE) phase1-clean
	@echo "$(GREEN)✅ Phase 1 migration completed$(NC)"
	@echo "$(GREEN)📁 Cleaned data available in: database/data/clean/$(NC)"

# =============================================================================
# Data Import (Phase 1 CSV to PostgreSQL)
# =============================================================================

build-migrator: ## Build the data migration tool
	@echo "$(YELLOW)🔨 Building migration tool...$(NC)"
	@cd backend && go build -o ../migrator cmd/migrator/main.go
	@echo "$(GREEN)✅ Migration tool built$(NC)"

import-data: build-migrator ## Import cleaned CSV data to PostgreSQL
	@echo "$(YELLOW)📥 Importing data to PostgreSQL...$(NC)"
	@DEV_DATABASE_URL="$(DEV_DATABASE_URL)" \
	 TEST_DATABASE_URL="$(TEST_DATABASE_URL)" \
	 DATA_PATH="$(DATA_PATH)" \
	 TENANT_ID="$(TENANT_ID)" \
	 ./migrator
	@echo "$(GREEN)✅ Data import completed$(NC)"

import-data-dry-run: build-migrator ## Dry run data import (no actual changes)
	@echo "$(YELLOW)🔍 Running data import dry run...$(NC)"
	@DEV_DATABASE_URL="$(DEV_DATABASE_URL)" \
	 TEST_DATABASE_URL="$(TEST_DATABASE_URL)" \
	 DATA_PATH="$(DATA_PATH)" \
	 TENANT_ID="$(TENANT_ID)" \
	 DRY_RUN=true \
	 ./migrator
	@echo "$(GREEN)✅ Dry run completed$(NC)"

import-docker: ## Run data import using Docker
	@echo "$(YELLOW)🐳 Running data import with Docker...$(NC)"
	@docker-compose run --rm migrator
	@echo "$(GREEN)✅ Docker data import completed$(NC)"

# =============================================================================
# Complete Migration Workflow
# =============================================================================

migrate-complete: ## Complete migration workflow (MDB → CSV → PostgreSQL)
	@echo "$(GREEN)🚀 Starting complete migration workflow...$(NC)"
	@echo "$(GREEN)Step 1: Phase 1 - Export and clean MDB data$(NC)"
	@$(MAKE) phase1-complete
	@echo ""
	@echo "$(GREEN)Step 2: Setup PostgreSQL containers$(NC)"
	@$(MAKE) setup-docker
	@$(MAKE) wait-for-db
	@echo ""
	@echo "$(GREEN)Step 3: Run database migrations$(NC)"
	@$(MAKE) migrate-db
	@echo ""
	@echo "$(GREEN)Step 4: Import customer data$(NC)"
	@$(MAKE) import-data
	@echo ""
	@echo "$(GREEN)🎉 Complete migration workflow finished!$(NC)"
	@$(MAKE) migration-summary

migration-summary: ## Show migration summary and next steps
	@echo ""
	@echo "$(GREEN)📊 Migration Summary$(NC)"
	@echo "===================="
	@echo "$(GREEN)✅ MDB data exported and cleaned$(NC)"
	@echo "$(GREEN)✅ PostgreSQL containers running$(NC)"
	@echo "$(GREEN)✅ Database schema created$(NC)"
	@echo "$(GREEN)✅ Customer data imported$(NC)"
	@echo ""
	@echo "$(YELLOW)🎯 Next Steps:$(NC)"
	@echo "1. Start development server: make dev"
	@echo "2. Run tests: make test"
	@echo "3. Access PgAdmin: make setup-docker-full then http://localhost:8080"
	@echo ""
	@echo "$(YELLOW)📁 Important Files:$(NC)"
	@echo "• Migration logs: database/logs/"
	@echo "• Cleaned data: database/data/clean/"
	@echo "• Database URL: $(DEV_DATABASE_URL)"

# =============================================================================
# Development and Testing
# =============================================================================

dev: ## Start development server
	@echo "$(YELLOW)🚀 Starting development server...$(NC)"
	@cd backend && go run cmd/server/main.go

test: ## Run tests
	@echo "$(YELLOW)🧪 Running tests...$(NC)"
	@cd backend && go test ./internal/...

test-customer: ## Run customer domain tests
	@echo "$(YELLOW)🧪 Running customer domain tests...$(NC)"
	@cd backend && go test ./internal/customer/...

# =============================================================================
# Database Management
# =============================================================================

db-status: ## Show database status
	@echo "$(YELLOW)📊 Database Status$(NC)"
	@echo "=================="
	@echo "Dev Database:"
	@docker-compose exec -T postgres psql "$(DEV_DATABASE_URL)" -c "\l" 2>/dev/null || echo "$(RED)❌ Dev database not accessible$(NC)"
	@echo ""
	@echo "Test Database:"
	@docker-compose exec -T postgres-test psql "$(TEST_DATABASE_URL)" -c "\l" 2>/dev/null || echo "$(RED)❌ Test database not accessible$(NC)"

db-reset: ## Reset databases (WARNING: destroys all data)
	@echo "$(RED)⚠️  WARNING: This will destroy all database data!$(NC)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ]
	@docker-compose down -v
	@docker volume rm oilgas-postgres_postgres_data oilgas-postgres_postgres_test_data 2>/dev/null || true
	@$(MAKE) setup-docker
	@$(MAKE) wait-for-db
	@$(MAKE) migrate-db
	@echo "$(GREEN)✅ Databases reset$(NC)"

db-logs: ## Show database logs
	@echo "$(YELLOW)📋 Database Logs$(NC)"
	@docker-compose logs postgres postgres-test

# =============================================================================
# Cleanup
# =============================================================================

clean: ## Clean up generated files and containers
	@echo "$(YELLOW)🧹 Cleaning up...$(NC)"
	@docker-compose down
	@rm -f migrator
	@rm -rf database/logs/*
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

clean-data: ## Clean up data files (keeps containers)
	@echo "$(YELLOW)🧹 Cleaning data files...$(NC)"
	@rm -rf database/data/exported/*
	@rm -rf database/data/clean/*
	@rm -rf database/logs/*
	@echo "$(GREEN)✅ Data files cleaned$(NC)"

setup-complete:
	@echo ""
	@echo "$(GREEN)🎉 Setup completed successfully!$(NC)"
	@echo ""
	@echo "$(YELLOW)📋 What's available:$(NC)"
	@echo "• PostgreSQL Dev:  localhost:5432 (oilgas_dev)"
	@echo "• PostgreSQL Test: localhost:5433 (oilgas_test)"
	@echo "• Customer domain: Enhanced with analytics and multi-tenant support"
	@echo ""
	@echo "$(YELLOW)🚀 Ready for development:$(NC)"
	@echo "• Run 'make migrate-complete' to import your MDB data"
	@echo "• Run 'make dev' to start the development server"
	@echo "• Run 'make test-customer' to test customer domain"

# =============================================================================
# Simple MDB to PostgreSQL Migration
# =============================================================================

# Use your actual exported customers.csv file
import-customers: ## Import customers.csv directly to PostgreSQL
	@echo "$(YELLOW)📥 Importing customers.csv to PostgreSQL...$(NC)"
	@echo "Building simple importer..."
	@cd backend && go build -o ../simple-importer cmd/simple-importer/main.go
	@echo "Running import..."
	@DATABASE_URL="$(DEV_DATABASE_URL)" ./simple-importer customers.csv $(TENANT_ID)
	@echo "$(GREEN)✅ Customer import completed$(NC)"

import-customers-test: ## Import customers.csv to test database
	@echo "$(YELLOW)📥 Importing customers.csv to test database...$(NC)"
	@cd backend && go build -o ../simple-importer cmd/simple-importer/main.go
	@DATABASE_URL="$(TEST_DATABASE_URL)" ./simple-importer customers.csv $(TENANT_ID)
	@echo "$(GREEN)✅ Test customer import completed$(NC)"

# Quick setup for your actual data
quick-setup: ## Quick setup with your actual customers.csv
	@echo "$(GREEN)🚀 Quick setup with actual customer data...$(NC)"
	@$(MAKE) setup-docker
	@$(MAKE) wait-for-db
	@$(MAKE) migrate-simple
	@$(MAKE) import-customers
	@$(MAKE) verify-customers
	@echo "$(GREEN)✅ Quick setup completed!$(NC)"

migrate-simple: ## Run simplified customer migration
	@echo "$(YELLOW)📊 Running simplified customer migration...$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -f /docker-entrypoint-initdb.d/001_init_database.sql
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -f /docker-entrypoint-initdb.d/migrations/001_create_customers_simplified.sql
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -f /docker-entrypoint-initdb.d/001_init_database.sql
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -f /docker-entrypoint-initdb.d/migrations/001_create_customers_simplified.sql
	@echo "$(GREEN)✅ Simplified migrations completed$(NC)"

verify-customers: ## Verify customer data was imported correctly
	@echo "$(YELLOW)🔍 Verifying customer data...$(NC)"
	@echo "Customer count:"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT COUNT(*) as total_customers FROM store.customers;"
	@echo ""
	@echo "Sample customers:"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT customer_id, customer, billingstate, CASE WHEN color1 IS NOT NULL THEN 'Yes' ELSE 'No' END as has_colors FROM store.customers LIMIT 5;"
	@echo ""
	@echo "Color system summary:"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT COUNT(*) as customers_with_colors FROM store.customers WHERE color1 IS NOT NULL OR color2 IS NOT NULL;"
	@echo ""
	@echo "W-String summary:"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "SELECT COUNT(*) as customers_with_wstring FROM store.customers WHERE wscolor1 IS NOT NULL OR wscolor2 IS NOT NULL;"

# Analyze your actual CSV structure
analyze-csv: ## Analyze the structure of customers.csv
	@echo "$(YELLOW)🔍 Analyzing customers.csv structure...$(NC)"
	@if [ -f "customers.csv" ]; then \
		echo "File: customers.csv"; \
		echo "Rows: $$(wc -l < customers.csv)"; \
		echo ""; \
		echo "Headers:"; \
		head -1 customers.csv | tr ',' '\n' | nl; \
		echo ""; \
		echo "Sample data (first 3 rows):"; \
		head -3 customers.csv; \
		echo ""; \
		echo "Data types detected:"; \
		head -2 customers.csv | tail -1 | tr ',' '\n' | while read field; do \
			if [[ "$$field" =~ ^[0-9]+$$ ]]; then \
				echo "Integer: $$field"; \
			elif [[ "$$field" =~ ^[0-9]+\.[0-9]+$$ ]]; then \
				echo "Float: $$field"; \
			elif [ -n "$$field" ]; then \
				echo "String: $$field"; \
			else \
				echo "Empty: $$field"; \
			fi; \
		done; \
	else \
		echo "$(RED)❌ customers.csv not found$(NC)"; \
		echo "Please ensure your exported CSV file is named 'customers.csv' and is in the project root."; \
	fi

# Clean up for fresh start
reset-simple: ## Reset for fresh simple migration
	@echo "$(YELLOW)🧹 Resetting for fresh migration...$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "TRUNCATE store.customers RESTART IDENTITY CASCADE;" 2>/dev/null || true
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -c "TRUNCATE store.customers RESTART IDENTITY CASCADE;" 2>/dev/null || true
	@rm -f simple-importer
	@echo "$(GREEN)✅ Reset completed$(NC)"

# Show what's in your database
show-customers: ## Show current customers in database
	@echo "$(YELLOW)📋 Current customers in database:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\
		SELECT \
			customer_id, \
			custid as orig_id, \
			customer, \
			billingstate as state, \
			CASE WHEN email IS NOT NULL THEN '✓' ELSE '✗' END as email, \
			CASE WHEN color1 IS NOT NULL THEN '✓' ELSE '✗' END as colors, \
			CASE WHEN wscolor1 IS NOT NULL THEN '✓' ELSE '✗' END as wstring, \
			deleted \
		FROM store.customers \
		ORDER BY customer_id \
		LIMIT 10;"

# Export current data for backup
export-customers: ## Export current customer data to CSV
	@echo "$(YELLOW)📤 Exporting current customer data...$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\
		COPY ( \
			SELECT custid, customer, billingaddress, billingcity, billingstate, billingzipcode, \
				   contact, phone, fax, email, \
				   color1, color2, color3, color4, color5, \
				   loss1, loss2, loss3, loss4, loss5, \
				   wscolor1, wscolor2, wscolor3, wscolor4, wscolor5, \
				   wsloss1, wsloss2, wsloss3, wsloss4, wsloss5, \
				   deleted \
			FROM store.customers \
			ORDER BY customer_id \
		) TO STDOUT WITH CSV HEADER;" > exported_customers_$(shell date +%Y%m%d_%H%M%S).csv
	@echo "$(GREEN)✅ Exported to: exported_customers_$(shell date +%Y%m%d_%H%M%S).csv$(NC)"

# =============================================================================
# Help for simple migration
# =============================================================================

help-simple: ## Show help for simple migration commands
	@echo "$(CYAN)Simple Customer Migration Commands:$(NC)"
	@echo ""
	@echo "$(YELLOW)📋 Setup:$(NC)"
	@echo "  make quick-setup           # Complete setup with your customers.csv"
	@echo "  make analyze-csv           # Analyze your customers.csv structure"
	@echo ""
	@echo "$(YELLOW)📥 Import:$(NC)"
	@echo "  make import-customers      # Import customers.csv to development DB"
	@echo "  make import-customers-test # Import customers.csv to test DB"
	@echo ""
	@echo "$(YELLOW)🔍 Verify:$(NC)"
	@echo "  make verify-customers      # Check import was successful"
	@echo "  make show-customers        # Display current customers"
	@echo ""
	@echo "$(YELLOW)🛠️  Manage:$(NC)"
	@echo "  make reset-simple          # Reset for fresh migration"
	@echo "  make export-customers      # Export current data to CSV"
	@echo ""
	@echo "$(YELLOW)📁 Required Files:$(NC)"
	@echo "  • customers.csv            # Your exported MDB customer data"
	@echo "  • docker-compose.yml       # Docker configuration"
	@echo ""
	@echo "$(YELLOW)🎯 Quick Start:$(NC)"
	@echo "  1. Place your customers.csv in the project root"
	@echo "  2. Run: make quick-setup"
	@echo "  3. Verify: make show-customers"

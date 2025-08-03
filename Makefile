# Optimized Makefile for Customer Migration with petros-lb.mdb
.PHONY: help setup clean dev test

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
BLUE := \033[34m
NC := \033[0m

# Configuration
MDB_FILE := db_prep/petros-lb.mdb
TENANT_ID := local-dev
DEV_DATABASE_URL := postgresql://oilgas_user:oilgas_pass@localhost:5432/oilgas_dev
TEST_DATABASE_URL := postgresql://oilgas_test_user:oilgas_test_pass@localhost:5433/oilgas_test

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

list-tables: check-mdb ## List all tables in MDB file
	@echo "$(BLUE)📋 Tables in $(MDB_FILE):$(NC)"
	@mdb-tables "$(MDB_FILE)" | tr ' ' '\n' | sort | nl

# =============================================================================
# DATABASE SETUP
# =============================================================================

setup-db: check-deps ## Setup PostgreSQL databases
	@echo "$(YELLOW)🐳 Setting up PostgreSQL databases...$(NC)"
	@docker-compose down 2>/dev/null || true
	@docker-compose up -d postgres postgres-test
	@echo "$(YELLOW)⏳ Waiting for databases...$(NC)"
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if docker-compose exec -T postgres pg_isready -U oilgas_user -d oilgas_dev >/dev/null 2>&1 && \
		   docker-compose exec -T postgres-test pg_isready -U oilgas_test_user -d oilgas_test >/dev/null 2>&1; then \
			echo "$(GREEN)✅ Databases ready$(NC)"; \
			break; \
		fi; \
		echo "   Waiting... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "$(RED)❌ Database timeout$(NC)"; \
		exit 1; \
	fi

migrate-db: ## Run database migrations
	@echo "$(YELLOW)📊 Running database migrations...$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -f /docker-entrypoint-initdb.d/001_init_database.sql
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -f /docker-entrypoint-initdb.d/001_init_database.sql
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -f /docker-entrypoint-initdb.d/migrations/001_create_customers_standardized.sql
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -f /docker-entrypoint-initdb.d/migrations/001_create_customers_standardized.sql
	@echo "$(GREEN)✅ Database migrations completed$(NC)"

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

setup-customers: check-deps check-mdb ## Complete customer setup workflow
	@echo "$(GREEN)🚀 Complete Customer Setup Workflow$(NC)"
	@echo "====================================="
	@$(MAKE) analyze-mdb
	@$(MAKE) analyze-customers
	@$(MAKE) setup-db
	@$(MAKE) migrate-db
	@$(MAKE) clean-customers
	@$(MAKE) import-customers
	@echo ""
	@echo "$(GREEN)🎉 Customer setup completed successfully!$(NC)"
	@$(MAKE) verify-customers

quick-setup: ## Quick setup (skip analysis steps)
	@echo "$(GREEN)⚡ Quick Customer Setup$(NC)"
	@$(MAKE) setup-db
	@$(MAKE) migrate-db
	@echo "$(GREEN)✅ Quick setup completed$(NC)"

# =============================================================================
# VERIFICATION AND TESTING
# =============================================================================

verify-customers: ## Verify customer data was imported correctly
	@echo "$(YELLOW)🔍 Verifying customer data...$(NC)"
	@echo "$(BLUE)Development Database:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\
		SELECT \
			COUNT(*) as total_customers, \
			COUNT(CASE WHEN is_deleted = false THEN 1 END) as active_customers, \
			COUNT(CASE WHEN email_address IS NOT NULL THEN 1 END) as customers_with_email, \
			COUNT(CASE WHEN color_grade_1 IS NOT NULL THEN 1 END) as customers_with_colors \
		FROM store.customers;"
	@echo ""
	@echo "$(BLUE)Sample customers:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\
		SELECT customer_id, customer_name, billing_state, \
			CASE WHEN email_address IS NOT NULL THEN '✓' ELSE '✗' END as email, \
			CASE WHEN color_grade_1 IS NOT NULL THEN '✓' ELSE '✗' END as colors \
		FROM store.customers WHERE is_deleted = false ORDER BY customer_id LIMIT 5;"

check-duplicates: ## Check for potential customer duplicates
	@echo "$(YELLOW)🔍 Checking for customer duplicates...$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\
		SELECT * FROM detect_customer_duplicates('$(TENANT_ID)');"

show-customers: ## Show current customers in database
	@echo "$(BLUE)📋 Current customers:$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\
		SELECT customer_id, customer_name, billing_city, billing_state, \
			CASE WHEN is_deleted THEN 'Deleted' ELSE 'Active' END as status \
		FROM store.customers ORDER BY customer_id LIMIT 10;"

test-customer-domain: ## Run customer domain tests
	@echo "$(YELLOW)🧪 Running customer domain tests...$(NC)"
	@cd backend && go test ./internal/customer/... -v

# =============================================================================
# DATABASE MANAGEMENT
# =============================================================================

db-status: ## Show database status
	@echo "$(BLUE)📊 Database Status:$(NC)"
	@echo "Development Database:"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\
		SELECT \
			current_database() as database, \
			current_user as user, \
			version() as version;" 2>/dev/null || echo "$(RED)❌ Dev database not accessible$(NC)"
	@echo ""
	@echo "Test Database:"
	@docker-compose exec -T postgres-test psql -U oilgas_test_user -d oilgas_test -c "\
		SELECT current_database() as database, current_user as user;" 2>/dev/null || echo "$(RED)❌ Test database not accessible$(NC)"

db-reset: ## Reset databases (WARNING: destroys all data)
	@echo "$(RED)⚠️  WARNING: This will destroy all database data!$(NC)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
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
# CLEANUP
# =============================================================================

clean: ## Clean up generated files and containers
	@echo "$(YELLOW)🧹 Cleaning up...$(NC)"
	@docker-compose down 2>/dev/null || true
	@rm -f customer-cleaner customer-analyzer standardized-importer
	@rm -f database/data/exported/*.csv
	@rm -f database/data/clean/*.csv
	@rm -f database/logs/*.log
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

clean-data: ## Clean up data files only (keep containers)
	@echo "$(YELLOW)🧹 Cleaning data files...$(NC)"
	@rm -f customers.csv database/data/exported/*.csv database/data/clean/*.csv
	@rm -f database/logs/*.log
	@echo "$(GREEN)✅ Data files cleaned$(NC)"

repo-cleanup: ## Run repository cleanup and optimization
	@echo "$(YELLOW)🔧 Running repository cleanup...$(NC)"
	@chmod +x cleanup_repository.sh
	@./cleanup_repository.sh
	@echo "$(GREEN)✅ Repository optimized$(NC)"

# =============================================================================
# UTILITIES
# =============================================================================

logs: ## Show Docker container logs
	@echo "$(BLUE)📋 Container logs:$(NC)"
	@docker-compose logs postgres postgres-test

export-customers: ## Export current customers to CSV
	@echo "$(YELLOW)📤 Exporting customers...$(NC)"
	@docker-compose exec -T postgres psql -U oilgas_user -d oilgas_dev -c "\
		COPY ( \
			SELECT original_customer_id, customer_name, billing_address, billing_city, \
				   billing_state, billing_zip_code, contact_name, phone_number, \
				   email_address, is_deleted \
			FROM store.customers ORDER BY customer_id \
		) TO STDOUT WITH CSV HEADER;" > exported_customers_$(shell date +%Y%m%d_%H%M%S).csv
	@echo "$(GREEN)✅ Customers exported$(NC)"

# =============================================================================
# INFORMATION COMMANDS
# =============================================================================

info: ## Show current configuration
	@echo "$(BLUE)ℹ️  Current Configuration:$(NC)"
	@echo "MDB File: $(MDB_FILE)"
	@echo "Tenant ID: $(TENANT_ID)"
	@echo "Dev Database: $(DEV_DATABASE_URL)"
	@echo "Test Database: $(TEST_DATABASE_URL)"
	@echo ""
	@echo "$(BLUE)📁 Key Files:$(NC)"
	@ls -la $(MDB_FILE) 2>/dev/null || echo "❌ MDB file not found"
	@ls -la customers.csv 2>/dev/null || echo "ℹ️  No customers.csv (will be generated)"
	@ls -la database/data/clean/customers_cleaned.csv 2>/dev/null || echo "ℹ️  No cleaned data yet"

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

# =============================================================================
# WORKFLOW HELPERS
# =============================================================================

first-time: ## Complete first-time setup workflow
	@echo "$(GREEN)🎯 First-Time Setup for Oil & Gas Customer Migration$(NC)"
	@echo "======================================================"
	@echo ""
	@echo "$(YELLOW)This will:$(NC)"
	@echo "1. Check all dependencies"
	@echo "2. Verify your MDB file"
	@echo "3. Set up PostgreSQL databases"
	@echo "4. Analyze and import customer data"
	@echo "5. Verify everything works"
	@echo ""
	@read -p "Continue? (y/N): " confirm && [ "$confirm" = "y" ] || exit 1
	@$(MAKE) check-deps
	@$(MAKE) check-mdb
	@$(MAKE) setup-customers
	@echo ""
	@echo "$(GREEN)🎉 First-time setup completed successfully!$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "• Run 'make dev' to start development server"
	@echo "• Run 'make start-pgadmin' to access database GUI"
	@echo "• Check 'make help' for all available commands"

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
	@echo "   Ubuntu: sudo apt-get install mdb-tools"
	@echo ""
	@echo "3. Docker not running:"
	@echo "   Start Docker Desktop or Docker daemon"
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
	@echo ""
	@echo "$(BLUE)Current Status:$(NC)"
	@$(MAKE) status

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

help-advanced: ## Show advanced commands
	@echo "$(BLUE)🔧 Advanced Commands:$(NC)"
	@echo "======================"
	@echo ""
	@echo "$(YELLOW)Database Management:$(NC)"
	@echo "  make db-reset           # Reset all databases (destructive)"
	@echo "  make start-pgadmin      # Start database GUI"
	@echo "  make export-customers   # Export current data to CSV"
	@echo ""
	@echo "$(YELLOW)Data Processing:$(NC)"
	@echo "  make analyze-customers  # Analyze CSV structure"
	@echo "  make check-duplicates   # Check for duplicate customers"
	@echo "  make clean-customers    # Clean data with deduplication"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  make test-customer-domain # Run customer domain tests"
	@echo "  make build              # Build all tools"
	@echo "  make dev                # Start development server"
	@echo ""
	@echo "$(YELLOW)Troubleshooting:$(NC)"
	@echo "  make troubleshoot       # Show troubleshooting guide"
	@echo "  make status             # Show system status"
	@echo "  make logs               # Show container logs"

.DEFAULT_GOAL := help

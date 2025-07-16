# Oil & Gas Inventory System - Makefile
# Phase 1: MDB Migration → Phase 2: Local Development

.PHONY: help setup migrate seed test build clean dev

# Default environment
ENV ?= local

# Color definitions for output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

help: ## Show this help message
	@echo "$(CYAN)Oil & Gas Inventory System - Available Commands$(RESET)"
	@echo ""
	@echo "$(GREEN)Phase 1 Commands (MDB Migration):$(RESET)"
	@echo "  phase1              - Complete Phase 1 MDB migration"
	@echo "  phase1-status       - Check Phase 1 completion status"
	@echo ""
	@echo "$(GREEN)Phase 2 Commands (Local Development):$(RESET)"
	@echo "  setup               - Complete local development setup"
	@echo "  import-clean-data   - Import Phase 1 normalized data"
	@echo "  dev                 - Start development servers"
	@echo ""
	@echo "$(GREEN)Database Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E "(migrate|seed|reset|status)" | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s - %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)Development Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E "(dev|build|test|clean)" | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s - %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Usage Examples:$(RESET)"
	@echo "  make phase1         # Start with Phase 1 migration"
	@echo "  make setup          # After Phase 1, setup local environment"
	@echo "  make dev            # Start development servers"

# =============================================================================
# PHASE 1: MDB MIGRATION COMMANDS
# =============================================================================

phase1: ## Run complete Phase 1 MDB migration
	@echo "$(CYAN)🚀 Starting Phase 1: MDB Migration$(RESET)"
	@if [ ! -f "setup_phase1.sh" ]; then \
		echo "$(RED)❌ setup_phase1.sh not found$(RESET)"; \
		echo "Please ensure you have the Phase 1 setup script"; \
		exit 1; \
	fi
	@./setup_phase1.sh

phase1-status: ## Check Phase 1 completion status
	@echo "$(CYAN)📊 Phase 1 Status Check$(RESET)"
	@echo ""
	@echo "$(YELLOW)Required Phase 1 Files:$(RESET)"
	@if [ -f "database/analysis/mdb_column_analysis.txt" ]; then \
		echo "  ✅ Column analysis: database/analysis/mdb_column_analysis.txt"; \
	else \
		echo "  ❌ Missing: database/analysis/mdb_column_analysis.txt"; \
	fi
	@if [ -f "database/analysis/table_counts.txt" ]; then \
		echo "  ✅ Table counts: database/analysis/table_counts.txt"; \
	else \
		echo "  ❌ Missing: database/analysis/table_counts.txt"; \
	fi
	@if [ -f "database/analysis/table_list.txt" ]; then \
		echo "  ✅ Table list: database/analysis/table_list.txt"; \
	else \
		echo "  ❌ Missing: database/analysis/table_list.txt"; \
	fi
	@if [ -f "database/schema/mdb_schema.sql" ]; then \
		echo "  ✅ Schema: database/schema/mdb_schema.sql"; \
	else \
		echo "  ❌ Missing: database/schema/mdb_schema.sql"; \
	fi
	@if [ -d "database/data/clean" ] && [ "$$(ls -A database/data/clean 2>/dev/null)" ]; then \
		csv_count=$$(ls -1 database/data/clean/*.csv 2>/dev/null | wc -l); \
		echo "  ✅ Clean data: $$csv_count CSV files in database/data/clean/"; \
	else \
		echo "  ❌ Missing: CSV files in database/data/clean/"; \
	fi
	@echo ""
	@if [ -f "database/analysis/phase1_migration_report.txt" ]; then \
		echo "$(GREEN)📄 Phase 1 Report Available:$(RESET)"; \
		echo "  cat database/analysis/phase1_migration_report.txt"; \
	else \
		echo "$(RED)❌ Phase 1 not completed - run 'make phase1'$(RESET)"; \
	fi

# =============================================================================
# ENVIRONMENT SETUP
# =============================================================================

init-env: ## Initialize environment file from template
	@echo "$(CYAN)🔧 Initializing environment configuration$(RESET)"
	@if [ ! -f .env ]; then \
		if [ -f .env.local ]; then \
			cp .env.local .env; \
			echo "$(GREEN)✅ Created .env from .env.local template$(RESET)"; \
		elif [ -f .env.example ]; then \
			cp .env.example .env; \
			echo "$(GREEN)✅ Created .env from .env.example template$(RESET)"; \
			echo "$(YELLOW)⚠️  Please update .env with your actual values$(RESET)"; \
		else \
			echo "$(RED)❌ No environment template found (.env.local or .env.example)$(RESET)"; \
			exit 1; \
		fi \
	else \
		echo "$(GREEN)✅ .env file already exists$(RESET)"; \
	fi

check-env: ## Validate environment configuration
	@echo "$(CYAN)🔍 Checking environment configuration$(RESET)"
	@if [ ! -f .env ]; then \
		echo "$(RED)❌ .env file not found - run 'make init-env'$(RESET)"; \
		exit 1; \
	fi
	@if ! grep -q "DATABASE_URL" .env; then \
		echo "$(RED)❌ DATABASE_URL not set in .env$(RESET)"; \
		exit 1; \
	fi
	@if ! grep -q "APP_PORT" .env; then \
		echo "$(RED)❌ APP_PORT not set in .env$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Environment configuration looks good$(RESET)"

check-phase1: ## Verify Phase 1 is complete before Phase 2
	@echo "$(CYAN)🔍 Verifying Phase 1 completion for Phase 2$(RESET)"
	@missing_files=0; \
	if [ ! -f "database/analysis/mdb_column_analysis.txt" ]; then \
		echo "$(RED)❌ Missing: database/analysis/mdb_column_analysis.txt$(RESET)"; \
		missing_files=$$((missing_files + 1)); \
	fi; \
	if [ ! -d "database/data/clean" ] || [ ! "$$(ls -A database/data/clean 2>/dev/null)" ]; then \
		echo "$(RED)❌ Missing: Clean CSV files in database/data/clean/$(RESET)"; \
		missing_files=$$((missing_files + 1)); \
	fi; \
	if [ $$missing_files -gt 0 ]; then \
		echo "$(RED)❌ Phase 1 not complete - run 'make phase1' first$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Phase 1 complete - ready for Phase 2$(RESET)"

# =============================================================================
# PHASE 2: LOCAL DEVELOPMENT SETUP
# =============================================================================

setup: check-phase1 init-env check-env ## Complete Phase 2 development setup
	@echo "$(CYAN)🚀 Phase 2: Setting up Local Development Environment$(RESET)"
	@echo "$(YELLOW)Prerequisites: Phase 1 must be completed$(RESET)"
	@echo ""
	@echo "$(CYAN)📁 Setting up backend structure...$(RESET)"
	@if [ -f "scripts/phase2_backend_structure.sh" ]; then \
		./scripts/phase2_backend_structure.sh; \
	else \
		echo "$(RED)❌ Backend structure script not found at scripts/phase2_backend_structure.sh$(RESET)"; \
		exit 1; \
	fi
	@echo "$(CYAN)Starting Docker services...$(RESET)"
	docker-compose up -d postgres
	@echo "$(CYAN)⏳ Waiting for database to be ready...$(RESET)"
	@sleep 10
	@echo "$(CYAN)📦 Installing Go dependencies...$(RESET)"
	@if [ -d "backend" ]; then \
		cd backend && go mod tidy; \
		echo "$(GREEN)✅ Go dependencies installed$(RESET)"; \
	fi
	@if [ -d "frontend" ]; then \
		echo "$(CYAN)📦 Installing frontend dependencies...$(RESET)"; \
		cd frontend && npm install; \
		echo "$(GREEN)✅ Frontend dependencies installed$(RESET)"; \
	fi
	@echo "$(CYAN)🔄 Running database migrations...$(RESET)"
	$(MAKE) migrate ENV=$(ENV)
	@echo "$(CYAN)🌱 Seeding database...$(RESET)"
	$(MAKE) seed ENV=$(ENV)
	@echo ""
	@echo "$(GREEN)✅ Phase 2 setup complete!$(RESET)"
	@echo "$(YELLOW)Next steps:$(RESET)"
	@echo "  1. Import your Phase 1 data: make import-clean-data"
	@echo "  2. Start development: make dev"

# =============================================================================
# DATABASE OPERATIONS
# =============================================================================

migrate: ## Run database migrations
	@echo "$(CYAN)🔄 Running migrations for $(ENV) environment$(RESET)"
	@if [ ! -f "backend/migrator" ]; then \
		echo "$(YELLOW)⚠️  Building migrator first...$(RESET)"; \
		cd backend && go build -o migrator migrator.go; \
	fi
	cd backend && ./migrator migrate $(ENV)

seed: ## Seed database with data
	@echo "$(CYAN)🌱 Seeding database for $(ENV) environment$(RESET)"
	@if [ ! -f "backend/migrator" ]; then \
		echo "$(YELLOW)⚠️  Building migrator first...$(RESET)"; \
		cd backend && go build -o migrator migrator.go; \
	fi
	cd backend && ./migrator seed $(ENV)

import-clean-data: check-phase1 ## Import normalized CSV files from Phase 1
	@echo "$(CYAN)📊 Importing Phase 1 normalized data$(RESET)"
	@if [ ! -d "database/data/clean" ]; then \
		echo "$(RED)❌ Clean data directory not found$(RESET)"; \
		echo "Run 'make phase1' first to generate normalized CSV files"; \
		exit 1; \
	fi
	@csv_count=$$(ls -1 database/data/clean/*.csv 2>/dev/null | wc -l); \
	if [ $$csv_count -eq 0 ]; then \
		echo "$(RED)❌ No CSV files found in database/data/clean/$(RESET)"; \
		echo "Run 'make phase1' to generate normalized data"; \
		exit 1; \
	fi
	@echo "$(CYAN)Found $$csv_count CSV files to import$(RESET)"
	@echo "$(CYAN)🔄 Importing data...$(RESET)"
	@for csv_file in database/data/clean/*.csv; do \
		if [ -f "$$csv_file" ]; then \
			table_name=$$(basename "$$csv_file" .csv); \
			echo "  📄 Importing $$table_name..."; \
			docker-compose exec -T postgres psql -U postgres -d oilgas_inventory_local \
				-c "\\copy store.$$table_name FROM STDIN WITH (FORMAT CSV, HEADER true, NULL '')" \
				< "$$csv_file" || echo "  ⚠️  Warning: $$table_name import had issues"; \
		fi; \
	done
	@echo "$(GREEN)✅ Phase 1 data import complete$(RESET)"

status: ## Show migration status
	@echo "$(CYAN)📊 Migration status for $(ENV) environment$(RESET)"
	@if [ -f "backend/migrator" ]; then \
		cd backend && ./migrator status $(ENV); \
	else \
		echo "$(RED)❌ Migrator not found - run 'make setup' first$(RESET)"; \
	fi

reset: ## Reset database (WARNING: Destructive)
	@echo "$(RED)⚠️  Resetting database for $(ENV) environment$(RESET)"
	@echo "$(YELLOW)This will delete all data!$(RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		echo "$(CYAN)🔄 Resetting database...$(RESET)"; \
		docker-compose down -v; \
		docker-compose up -d postgres; \
		sleep 10; \
		$(MAKE) migrate seed ENV=$(ENV); \
		echo "$(GREEN)✅ Database reset complete$(RESET)"; \
	else \
		echo ""; \
		echo "$(YELLOW)Database reset cancelled$(RESET)"; \
	fi

# =============================================================================
# DEVELOPMENT SERVERS
# =============================================================================

dev-start: ## Start development environment (databases only)
	@echo "$(CYAN)🚀 Starting development environment$(RESET)"
	docker-compose up -d
	@echo "$(GREEN)✅ Development environment started$(RESET)"
	@echo ""
	@echo "$(CYAN)📋 Available services:$(RESET)"
	@echo "  🐘 PostgreSQL: localhost:5432"
	@echo "  🗄️  PgAdmin: http://localhost:8080"

dev-stop: ## Stop development environment
	@echo "$(CYAN)🛑 Stopping development environment$(RESET)"
	docker-compose down
	@echo "$(GREEN)✅ Development environment stopped$(RESET)"

dev: dev-start ## Start development servers (database + application)
	@echo "$(CYAN)🚀 Starting development servers$(RESET)"
	@echo ""
	@echo "$(YELLOW)Starting in parallel...$(RESET)"
	@echo "$(CYAN)Backend server: http://localhost:8000$(RESET)"
	@echo "$(CYAN)Frontend server: http://localhost:3000$(RESET)"
	@echo "$(CYAN)PgAdmin: http://localhost:8080$(RESET)"
	@echo ""
	@echo "$(YELLOW)Press Ctrl+C to stop all servers$(RESET)"
	@trap 'echo "$(CYAN)Stopping servers...$(RESET)"; kill 0' INT; \
	( \
		if [ -d "backend" ] && [ -f "backend/cmd/server/main.go" ]; then \
			echo "$(CYAN)Starting backend...$(RESET)"; \
			cd backend && go run cmd/server/main.go; \
		else \
			echo "$(YELLOW)⚠️  Backend not found - skipping$(RESET)"; \
			sleep infinity; \
		fi \
	) & \
	( \
		if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then \
			echo "$(CYAN)Starting frontend...$(RESET)"; \
			cd frontend && npm run dev; \
		else \
			echo "$(YELLOW)⚠️  Frontend not found - skipping$(RESET)"; \
			sleep infinity; \
		fi \
	) & \
	wait

dev-backend: ## Start backend development server only
	@echo "$(CYAN)🚀 Starting backend development server$(RESET)"
	@if [ -d "backend" ] && [ -f "backend/cmd/server/main.go" ]; then \
		cd backend && go run cmd/server/main.go; \
	else \
		echo "$(RED)❌ Backend not found at backend/cmd/server/main.go$(RESET)"; \
	fi

dev-frontend: ## Start frontend development server only
	@echo "$(CYAN)🚀 Starting frontend development server$(RESET)"
	@if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then \
		cd frontend && npm run dev; \
	else \
		echo "$(RED)❌ Frontend not found or package.json missing$(RESET)"; \
	fi

# =============================================================================
# BUILD OPERATIONS
# =============================================================================

build: ## Build all components
	@echo "$(CYAN)🔨 Building all components$(RESET)"
	$(MAKE) build-backend
	$(MAKE) build-frontend

build-backend: ## Build backend only
	@echo "$(CYAN)🔨 Building backend$(RESET)"
	@if [ -d "backend" ]; then \
		cd backend && go build -o migrator migrator.go; \
		if [ -f "cmd/server/main.go" ]; then \
			go build -o server cmd/server/main.go; \
		fi; \
		echo "$(GREEN)✅ Backend built successfully$(RESET)"; \
	else \
		echo "$(RED)❌ Backend directory not found$(RESET)"; \
	fi

build-frontend: ## Build frontend only
	@echo "$(CYAN)🔨 Building frontend$(RESET)"
	@if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then \
		cd frontend && npm run build; \
		echo "$(GREEN)✅ Frontend built successfully$(RESET)"; \
	else \
		echo "$(RED)❌ Frontend not found or package.json missing$(RESET)"; \
	fi

# =============================================================================
# TESTING
# =============================================================================

test: ## Run all tests
	@echo "$(CYAN)🧪 Running all tests$(RESET)"
	$(MAKE) test-unit
	$(MAKE) test-integration

test-unit: ## Run unit tests
	@echo "$(CYAN)🔬 Running unit tests$(RESET)"
	@if [ -d "backend" ]; then \
		cd backend && go test ./... -v -race -short; \
	else \
		echo "$(YELLOW)⚠️  Backend not found - skipping unit tests$(RESET)"; \
	fi

test-integration: ## Run integration tests
	@echo "$(CYAN)🔗 Running integration tests$(RESET)"
	@echo "$(YELLOW)⚠️  Requires test database setup$(RESET)"
	@if [ -d "backend/test/integration" ]; then \
		cd backend && go test ./test/integration/... -v; \
	else \
		echo "$(YELLOW)⚠️  Integration tests not found$(RESET)"; \
	fi

test-coverage: ## Run tests with coverage
	@echo "$(CYAN)📊 Running tests with coverage$(RESET)"
	@if [ -d "backend" ]; then \
		cd backend && go test ./... -coverprofile=coverage.out; \
		cd backend && go tool cover -html=coverage.out -o coverage.html; \
		echo "$(GREEN)📈 Coverage report: backend/coverage.html$(RESET)"; \
	fi

# =============================================================================
# UTILITY COMMANDS
# =============================================================================

clean: ## Clean up generated files
	@echo "$(CYAN)🧹 Cleaning up generated files$(RESET)"
	@rm -f backend/migrator backend/server 2>/dev/null || true
	@rm -rf frontend/dist backend/coverage.out backend/coverage.html 2>/dev/null || true
	@rm -rf backend/data-tools 2>/dev/null || true
	@echo "$(GREEN)✅ Cleanup complete$(RESET)"

clean-all: clean ## Clean everything including Docker volumes
	@echo "$(CYAN)🧹 Deep cleaning (including Docker volumes)$(RESET)"
	docker-compose down -v
	@echo "$(GREEN)✅ Deep cleanup complete$(RESET)"

logs: ## Show Docker service logs
	@echo "$(CYAN)📋 Docker service logs$(RESET)"
	docker-compose logs -f

db-shell: ## Connect to PostgreSQL shell
	@echo "$(CYAN)🐘 Connecting to PostgreSQL$(RESET)"
	docker-compose exec postgres psql -U postgres -d oilgas_inventory_local

# =============================================================================
# INFORMATION COMMANDS
# =============================================================================

info: ## Show project information and status
	@echo "$(CYAN)📋 Oil & Gas Inventory System - Project Information$(RESET)"
	@echo ""
	@echo "$(GREEN)Project Structure:$(RESET)"
	@echo "  📁 docs/           - Documentation for each phase"
	@echo "  📁 scripts/        - Utility scripts (Phase 1 migration, etc.)"
	@echo "  📁 database/       - Generated data and analysis"
	@echo "  📁 backend/        - Go backend application"
	@echo "  📁 frontend/       - Vue.js frontend application"
	@echo ""
	@echo "$(GREEN)Current Status:$(RESET)"
	@$(MAKE) phase1-status
	@echo ""
	@echo "$(GREEN)Available Documentation:$(RESET)"
	@if [ -f "docs/README_PHASE1.md" ]; then \
		echo "  📖 Phase 1 Guide: docs/README_PHASE1.md"; \
	fi
	@if [ -f "docs/README_PHASE2.md" ]; then \
		echo "  📖 Phase 2 Guide: docs/README_PHASE2.md"; \
	fi
	@echo ""
	@echo "$(YELLOW)Quick Start:$(RESET)"
	@echo "  1. make phase1     # Complete MDB migration"
	@echo "  2. make setup      # Setup local development"
	@echo "  3. make dev        # Start development servers"

# =============================================================================
# SHORTCUTS AND ALIASES
# =============================================================================

start: dev-start ## Alias for dev-start
stop: dev-stop ## Alias for dev-stop
serve: dev ## Alias for dev

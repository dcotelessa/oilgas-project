# Database Management
.PHONY: migrate migrate-local migrate-test seed seed-local seed-test status reset db-connect

# Environment variables
DB_LOCAL := $(DATABASE_URL)
DB_TEST := $(TEST_DATABASE_URL)

# Migration commands
migrate: migrate-local ## Run migrations on local database (default)

migrate-local: ## Run migrations on local database
	@echo "$(GREEN)Running migrations on local database...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set. Run 'make debug-env' to check environment.$(RESET)"; \
		exit 1; \
	fi
	@go run migrator.go migrate local

migrate-test: ## Run migrations on test database
	@echo "$(GREEN)Running migrations on test database...$(RESET)"
	@if [ -z "$(TEST_DATABASE_URL)" ]; then \
		echo "$(RED)❌ TEST_DATABASE_URL not set. Check .env.local$(RESET)"; \
		exit 1; \
	fi
	@DATABASE_URL="$(TEST_DATABASE_URL)" go run migrator.go migrate local

# Seeding commands
seed: seed-local ## Seed local database with sample data (default)

seed-local: migrate-local ## Seed local database with sample data
	@echo "$(GREEN)Seeding local database...$(RESET)"
	@go run migrator.go seed local

seed-test: migrate-test ## Seed test database with minimal test data
	@echo "$(GREEN)Seeding test database with minimal data...$(RESET)"
	@DATABASE_URL="$(TEST_DATABASE_URL)" go run migrator.go seed local

# Status and diagnostics
status: ## Show migration status for local database
	@echo "$(GREEN)Local Database Status:$(RESET)"
	@go run migrator.go status local
	@echo ""
	@echo "$(BLUE)Test Database Status:$(RESET)"
	@if [ -n "$(TEST_DATABASE_URL)" ]; then \
		DATABASE_URL="$(TEST_DATABASE_URL)" go run migrator.go status local; \
	else \
		echo "$(YELLOW)TEST_DATABASE_URL not configured$(RESET)"; \
	fi

# Reset commands (destructive)
reset: ## Reset local database (destructive)
	@echo "$(RED)⚠️  This will destroy all local database data!$(RESET)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ]
	@go run migrator.go reset local

reset-test: ## Reset test database (destructive)
	@echo "$(YELLOW)Resetting test database...$(RESET)"
	@DATABASE_URL="$(TEST_DATABASE_URL)" go run migrator.go reset local

# Database connections
db-connect: ## Connect to local database with psql
	@echo "$(GREEN)Connecting to local database...$(RESET)"
	@psql "$(DATABASE_URL)"

db-connect-test: ## Connect to test database with psql
	@echo "$(GREEN)Connecting to test database...$(RESET)"
	@psql "$(TEST_DATABASE_URL)"

# Import/Export (for your MDB data)
import-mdb-data: seed-local ## Import converted MDB data to local database
	@echo "$(GREEN)Importing converted MDB data...$(RESET)"
	@echo "$(BLUE)📋 This will import your normalized CSV data from Phase 1$(RESET)"
	@if [ -d "data/normalized" ]; then \
		echo "$(GREEN)Found normalized data directory$(RESET)"; \
		# Add your CSV import logic here when ready; \
	else \
		echo "$(YELLOW)⚠️  No normalized data found. Run Phase 1 MDB conversion first.$(RESET)"; \
	fi

# Test database management
test-db-setup: migrate-test seed-test ## Complete test database setup
	@echo "$(GREEN)✅ Test database ready for testing$(RESET)"

test-db-clean: reset-test ## Clean test database completely
	@echo "$(GREEN)✅ Test database cleaned$(RESET)"

# Health checks
db-health: ## Check health of both databases
	@echo "$(GREEN)Checking database health...$(RESET)"
	@echo "$(BLUE)Local Database:$(RESET)"
	@psql "$(DATABASE_URL)" -c "SELECT 'Local DB: Connected' as status;" 2>/dev/null || echo "❌ Local DB: Connection failed"
	@echo "$(BLUE)Test Database:$(RESET)"
	@if [ -n "$(TEST_DATABASE_URL)" ]; then \
		psql "$(TEST_DATABASE_URL)" -c "SELECT 'Test DB: Connected' as status;" 2>/dev/null || echo "❌ Test DB: Connection failed"; \
	else \
		echo "⚠️  Test DB: Not configured"; \
	fi

# Development workflow helpers
dev-setup: docker-up migrate-local seed-local ## Complete development setup
	@echo "$(GREEN)✅ Development environment ready$(RESET)"
	@echo "$(BLUE)Local database: localhost:5433$(RESET)"
	@echo "$(BLUE)Test database: localhost:5434$(RESET)"

test-setup: docker-up test-db-setup ## Setup for running tests
	@echo "$(GREEN)✅ Test environment ready$(RESET)"

# Show database information
db-info: ## Show database connection information
	@echo "$(BLUE)Database Configuration:$(RESET)"
	@echo "Local DB:  $(DATABASE_URL)"
	@echo "Test DB:   $(TEST_DATABASE_URL)"
	@echo ""
	@echo "$(BLUE)Quick Commands:$(RESET)"
	@echo "  make dev-setup     - Setup local development"
	@echo "  make test-setup    - Setup for testing"
	@echo "  make db-health     - Check database health"
	@echo "  make status        - Show migration status"

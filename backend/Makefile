# =============================================================================
# MAIN MAKEFILE - Oil & Gas Inventory System
# =============================================================================
# Project orchestrator with modular domain-specific includes

# Colors for output
GREEN  := \033[0;32m
YELLOW := \033[1;33m
BLUE   := \033[0;34m
RED    := \033[0;31m
RESET  := \033[0m

# Project configuration
PROJECT_NAME := oil-gas-inventory
API_PORT     := 8000
DB_NAME      := oilgas_inventory_local

# Environment detection
ENV ?= local
ifeq ($(ENV),production)
    API_PORT := 80
    DB_NAME  := oilgas_inventory_production
endif

# =============================================================================
# INCLUDE MODULAR MAKEFILES
# =============================================================================
include make/development.mk
include make/database.mk
include make/auth.mk
include make/events.mk
include make/api.mk
include make/monitoring.mk
include make/testing.mk

# =============================================================================
# MAIN ORCHESTRATION COMMANDS
# =============================================================================

.PHONY: setup dev test clean help

setup: db-setup auth-setup events-setup api-setup ## 🚀 Complete project setup
	@echo "$(GREEN)✅ Oil & Gas Inventory System setup complete!$(RESET)"
	@echo "$(YELLOW)Next steps:$(RESET)"
	@echo "  1. make dev              # Start development server"
	@echo "  2. make auth-create-admin # Create admin user"
	@echo "  3. make test             # Run test suite"

dev: db-start events-start api-dev ## 🛠️  Start development environment
	@echo "$(GREEN)🚀 Development environment running$(RESET)"
	@echo "$(BLUE)API: http://localhost:$(API_PORT)$(RESET)"
	@echo "$(BLUE)Health: http://localhost:$(API_PORT)/health$(RESET)"

test: test-unit test-integration test-tenant test-api ## 🧪 Run comprehensive test suite
	@echo "$(GREEN)✅ All tests completed$(RESET)"

clean: api-clean db-clean events-clean test-clean ## 🧹 Clean all artifacts
	@echo "$(GREEN)✅ Cleanup complete$(RESET)"

monitor: monitor-connections monitor-events monitor-api ## 📊 System monitoring dashboard
	@echo "$(GREEN)📊 Monitoring active$(RESET)"

# =============================================================================
# QUICK WORKFLOW COMMANDS
# =============================================================================

.PHONY: quick-start quick-test quick-deploy

quick-start: setup dev ## ⚡ One-command startup (setup + dev)

quick-test: ## ⚡ Fast test cycle (unit + API only)
	@echo "$(YELLOW)🧪 Running quick test cycle...$(RESET)"
	@$(MAKE) test-unit test-api

quick-deploy: test api-build ## ⚡ Build and validate for deployment
	@echo "$(GREEN)🚀 Ready for deployment$(RESET)"

# =============================================================================
# TENANT MANAGEMENT
# =============================================================================

.PHONY: tenant-create tenant-list tenant-import

tenant-create: ## 🏢 Create new tenant database
	@read -p "Tenant ID (e.g., longbeach): " tenant && \
	echo "$(YELLOW)Creating tenant: $$tenant$(RESET)" && \
	go run cmd/migrator/main.go tenant-create $$tenant

tenant-list: ## 📋 List all tenant databases
	@echo "$(BLUE)Active tenants:$(RESET)"
	@psql "$$DATABASE_URL" -c "\l" | grep oilgas_ | awk '{print $$1}' | sed 's/oilgas_//' || echo "No tenants found"

tenant-import: ## 📥 Import MDB data for tenant
	@read -p "MDB file path: " file && \
	read -p "Tenant ID: " tenant && \
	echo "$(YELLOW)Importing $$file for tenant $$tenant$(RESET)" && \
	go run cmd/tools/mdb-importer.go --file="$$file" --tenant="$$tenant"

# =============================================================================
# DEVELOPMENT UTILITIES
# =============================================================================

.PHONY: logs debug install-deps update-deps

logs: ## 📋 Show application logs
	@echo "$(BLUE)Recent application logs:$(RESET)"
	@tail -f logs/app.log 2>/dev/null || echo "No log file found"

debug: ## 🐛 Debug environment information
	@echo "$(BLUE)🐛 Debug Information$(RESET)"
	@echo "Environment: $(ENV)"
	@echo "API Port: $(API_PORT)"
	@echo "Database: $(DB_NAME)"
	@echo "DATABASE_URL: $$DATABASE_URL"
	@echo "Go version: $$(go version)"
	@echo "Available make targets:"
	@$(MAKE) help | head -10

install-deps: ## 📦 Install all dependencies
	@echo "$(YELLOW)📦 Installing dependencies...$(RESET)"
	@go mod download
	@go mod tidy

update-deps: ## 🔄 Update all dependencies
	@echo "$(YELLOW)🔄 Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy

# =============================================================================
# HELP SYSTEM
# =============================================================================

help: ## 📖 Show available commands
	@echo "$(BLUE)Oil & Gas Inventory System - Available Commands$(RESET)"
	@echo "================================================="
	@echo ""
	@echo "$(GREEN)🚀 MAIN COMMANDS:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## 🚀.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)🛠️  DEVELOPMENT:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## 🛠️.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BLUE)🧪 TESTING:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## 🧪.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "$(RED)🏢 TENANT MANAGEMENT:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## 🏢.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BLUE)📊 MONITORING:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## 📊.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)⚡ QUICK COMMANDS:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## ⚡.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "For detailed help on specific modules:"
	@echo "  make help-database    # Database operations"
	@echo "  make help-auth        # Authentication system"  
	@echo "  make help-events      # Event system"
	@echo "  make help-api         # API management"
	@echo "  make help-testing     # Testing framework"

commands: ## 📋 List all available commands (detailed)
	@echo "$(BLUE)📋 All Available Commands:$(RESET)"
	@echo "=========================="
	@$(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($1 !~ "^[#.]") {print $1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@

# =============================================================================
# END OF MAIN MAKEFILE
# =============================================================================

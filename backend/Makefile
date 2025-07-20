# Oil & Gas Inventory System - Phase 3 Makefile
# Tenant-aware authentication and core API

# Default environment
ENV ?= local

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

# Auto-load environment variables
ifneq (,$(wildcard ../.env.local))
    include ../.env.local
    export
endif
ifneq (,$(wildcard .env.local))
    include .env.local
    export
endif
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Include all module makefiles
include make/database.mk
include make/development.mk
include make/testing.mk
include make/auth.mk
include make/docker.mk

.PHONY: help setup clean quick-start

help: ## Show this help message
	@echo "$(GREEN)Oil & Gas Inventory System - Phase 3$(RESET)"
	@echo "$(YELLOW)Tenant-aware authentication and core API$(RESET)"
	@echo ""
	@echo "$(GREEN)Main Commands:$(RESET)"
	@grep -E "^[a-zA-Z_-]+:.*?## .*$$" $(MAKEFILE_LIST) | grep -v "make/" | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)Module Commands:$(RESET)"
	@echo "$(YELLOW)Database:$(RESET) migrate, seed, migrate-reset, db-status, backup-db"
	@echo "$(YELLOW)Development:$(RESET) dev, build, lint, format, install-deps"
	@echo "$(YELLOW)Testing:$(RESET) test, test-unit, test-integration, test-coverage, test-auth"
	@echo "$(YELLOW)Authentication:$(RESET) create-admin, create-tenant, list-tenants, list-users, validate-rls"
	@echo "$(YELLOW)Docker:$(RESET) docker-up, docker-down, docker-logs"

setup: ## Complete setup for development
	@echo "$(GREEN)Setting up Oil & Gas Inventory System - Phase 3...$(RESET)"
	@echo "$(YELLOW)Environment loaded: DATABASE_URL=$(DATABASE_URL)$(RESET)"
	$(MAKE) install-deps
	$(MAKE) docker-up
	@echo "$(YELLOW)Waiting for PostgreSQL to be ready...$(RESET)"
	sleep 5
	$(MAKE) migrate ENV=local
	$(MAKE) seed ENV=local
	@echo "$(GREEN)✅ Setup complete!$(RESET)"

quick-start: ## Quick start for new developers
	@echo "$(GREEN)🚀 Quick Start - Phase 3$(RESET)"
	$(MAKE) setup
	@echo "$(YELLOW)Creating default admin user...$(RESET)"
	@go run scripts/utilities/create_user.go --email=admin@oilgas.local --password=admin123 --role=admin --tenant=default --company="Admin Company" || true
	@echo "$(GREEN)✅ System ready!$(RESET)"
	@echo "$(GREEN)📋 Admin login:$(RESET)"
	@echo "   Email: admin@oilgas.local"
	@echo "   Password: admin123"
	@echo "   Tenant: default"

clean: ## Clean up build artifacts and logs
	@echo "$(GREEN)Cleaning up...$(RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf exports/
	@echo "$(GREEN)✅ Cleanup complete$(RESET)"

# Environment debugging
debug-env: ## Show current environment variables
	@echo "$(GREEN)Environment Debug:$(RESET)"
	@echo "DATABASE_URL: $(DATABASE_URL)"
	@echo "APP_ENV: $(APP_ENV)"
	@echo "APP_PORT: $(APP_PORT)"
	@echo "ENV: $(ENV)"

# Default target
.DEFAULT_GOAL := help

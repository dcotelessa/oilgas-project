#!/bin/bash
# scripts/generators/makefile_modules_generator.sh
# Generates modular Makefile structure for Phase 3

set -e
echo "📋 Generating modular Makefile structure..."

# Detect backend directory
BACKEND_DIR=""
if [ -d "backend" ] && [ -f "backend/go.mod" ]; then
    BACKEND_DIR="backend/"
elif [ -f "go.mod" ]; then
    BACKEND_DIR=""
else
    echo "❌ Error: Cannot find go.mod file"
    exit 1
fi

# Create make modules directory
mkdir -p "${BACKEND_DIR}make"

# Generate main Makefile
cat > "${BACKEND_DIR}Makefile" << 'EOF'
# Oil & Gas Inventory System - Phase 3 Makefile
# Modular structure following established patterns

ENV ?= local

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

# Include all module makefiles
include make/database.mk
include make/development.mk
include make/testing.mk
include make/auth.mk

.PHONY: help setup clean quick-start

help: ## Show this help message
	@echo "$(GREEN)Oil & Gas Inventory System - Phase 3$(RESET)"
	@echo "$(YELLOW)Tenant-aware authentication and core API$(RESET)"
	@echo ""
	@echo "$(GREEN)Main Commands:$(RESET)"
	@grep -E "^[a-zA-Z_-]+:.*?## .*$$" $(MAKEFILE_LIST) | grep -v "make/" | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

setup: ## Complete setup for development
	@echo "$(GREEN)Setting up Phase 3...$(RESET)"
	go mod tidy
	$(MAKE) migrate ENV=local
	$(MAKE) seed ENV=local
	@echo "$(GREEN)✅ Setup complete!$(RESET)"

quick-start: ## Quick start with demo admin user
	@echo "$(GREEN)🚀 Quick Start - Phase 3$(RESET)"
	$(MAKE) setup
	go run scripts/utilities/create_user.go --email=admin@oilgas.com --password=admin123 --role=admin --company="Oil Gas Admin" --tenant=default || true
	@echo "$(GREEN)✅ Ready! Use: make dev$(RESET)"

clean: ## Clean up build artifacts
	@echo "$(GREEN)Cleaning up...$(RESET)"
	rm -rf bin/ coverage.out coverage.html

.DEFAULT_GOAL := help
EOF

# Generate database module
cat > "${BACKEND_DIR}make/database.mk" << 'EOF'
# Database Operations Module
.PHONY: migrate seed migrate-reset db-status

migrate: ## Run database migrations
	@echo "$(GREEN)Running migrations for $(ENV)...$(RESET)"
	go run migrator.go migrate $(ENV)

seed: ## Seed database with test data
	@echo "$(GREEN)Seeding database for $(ENV)...$(RESET)"
	go run migrator.go seed $(ENV)

migrate-reset: ## Reset database (DESTRUCTIVE)
	@echo "$(RED)⚠️  Resetting database for $(ENV)...$(RESET)"
	@read -p "Are you sure? This will delete all data! (y/N): " confirm && [ "$$confirm" = "y" ]
	go run migrator.go reset $(ENV)

db-status: ## Show database status
	@echo "$(GREEN)Database status for $(ENV):$(RESET)"
	go run migrator.go status $(ENV)
EOF

# Generate development module
cat > "${BACKEND_DIR}make/development.mk" << 'EOF'
# Development Module
.PHONY: dev build lint format

dev: ## Start development server
	@echo "$(GREEN)Starting development server...$(RESET)"
	@echo "$(YELLOW)API: http://localhost:8000$(RESET)"
	go run cmd/server/main.go

build: ## Build production binary
	@echo "$(GREEN)Building production binary...$(RESET)"
	CGO_ENABLED=0 go build -o bin/server cmd/server/main.go

lint: ## Run code linting
	@echo "$(GREEN)Running code linting...$(RESET)"
	@if command -v golangci-lint &> /dev/null; then golangci-lint run ./...; fi

format: ## Format code
	@echo "$(GREEN)Formatting code...$(RESET)"
	go fmt ./...
EOF

# Generate testing module
cat > "${BACKEND_DIR}make/testing.mk" << 'EOF'
# Testing Module
.PHONY: test test-unit test-coverage

test: ## Run all tests
	@echo "$(GREEN)Running all tests...$(RESET)"
	go test -v ./...

test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(RESET)"
	go test -v -short ./internal/...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(RESET)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
EOF

# Generate auth module
cat > "${BACKEND_DIR}make/auth.mk" << 'EOF'
# Authentication Module
.PHONY: create-admin create-tenant list-tenants

create-admin: ## Create admin user
	@echo "$(GREEN)Creating admin user...$(RESET)"
	@read -p "Email: " email; \
	read -s -p "Password: " password; echo; \
	read -p "Company: " company; \
	read -p "Tenant slug [default]: " tenant; \
	tenant=$${tenant:-default}; \
	go run scripts/utilities/create_user.go --email="$$email" --password="$$password" --role=admin --company="$$company" --tenant="$$tenant"

create-tenant: ## Create new tenant
	@echo "$(GREEN)Creating new tenant...$(RESET)"
	@read -p "Tenant name: " name; \
	read -p "Tenant slug: " slug; \
	go run scripts/utilities/create_tenant.go --name="$$name" --slug="$$slug"

list-tenants: ## List all tenants
	@echo "$(GREEN)Active tenants:$(RESET)"
	@if [ -n "$$DATABASE_URL" ]; then \
		psql "$$DATABASE_URL" -c "SELECT name, slug, database_type FROM store.tenants ORDER BY created_at;" 2>/dev/null || echo "$(RED)Database not accessible$(RESET)"; \
	fi
EOF

echo "✅ Modular Makefile structure generated"
echo "   - Main orchestrator Makefile"
echo "   - Database operations module"
echo "   - Development and building module"
echo "   - Testing module with coverage"
echo "   - Authentication module"

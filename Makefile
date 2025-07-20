# Oil & Gas Inventory System - Root Makefile
# Orchestrates backend + frontend + infrastructure

.PHONY: help setup build test dev clean health
.DEFAULT_GOAL := help

# Environment detection
ENV ?= local
BACKEND_DIR := backend
FRONTEND_DIR := frontend

help: ## Show this help message
	@echo "Oil & Gas Inventory System"
	@echo "=========================="
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $1, $2}' $(MAKEFILE_LIST)

setup: ## Setup entire project (backend + infrastructure)
	@echo "🚀 Setting up Oil & Gas Inventory System..."
	@echo "📁 Setting up backend..."
	@cd $(BACKEND_DIR) && go mod tidy
	@echo "🐳 Starting infrastructure..."
	@docker-compose up -d postgres
	@echo "⏳ Waiting for database..."
	@sleep 3
	@echo "🔄 Running migrations..."
	@cd $(BACKEND_DIR) && go run migrator.go migrate $(ENV)
	@echo "🌱 Running seeds..."
	@cd $(BACKEND_DIR) && go run migrator.go seed $(ENV)
	@echo "✅ Project setup complete!"

build: ## Build backend
	@echo "🔨 Building backend..."
	@cd $(BACKEND_DIR) && go build -o ../bin/server cmd/server/main.go

test: ## Run backend tests
	@echo "🧪 Running backend tests..."
	@cd $(BACKEND_DIR) && go test ./...

dev: ## Start development environment
	@echo "🚀 Starting development environment..."
	@docker-compose up -d postgres
	@echo "⏳ Waiting for database..."
	@sleep 3
	@echo "🔄 Ensuring migrations are current..."
	@cd $(BACKEND_DIR) && go run migrator.go migrate $(ENV)
	@echo "🌟 Starting backend server..."
	@cd $(BACKEND_DIR) && go run cmd/server/main.go

clean: ## Clean all build artifacts
	@echo "🧹 Cleaning project..."
	@rm -rf bin/
	@docker-compose down

health: ## Check system health
	@echo "🔍 System health check..."
	@echo "🐳 Docker containers:"
	@docker-compose ps
	@echo ""
	@echo "🗄️ Database status:"
	@cd $(BACKEND_DIR) && go run migrator.go status $(ENV)
	@echo ""
	@echo "🌐 API health (if running):"
	@curl -s http://localhost:8000/health || echo "API not running"

# Database operations
db-status: ## Show database status
	@cd $(BACKEND_DIR) && go run migrator.go status $(ENV)

db-reset: ## Reset database (development only)
	@echo "⚠️ This will destroy all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; echo; \
	if [[ $REPLY =~ ^[Yy]$ ]]; then \
		cd $(BACKEND_DIR) && go run migrator.go reset $(ENV); \
		echo "Run 'make setup' to restore"; \
	fi

# Development utilities
logs: ## Show service logs
	@docker-compose logs -f

restart: ## Restart all services
	@docker-compose restart

# Phase 3 preparation
phase3-ready: ## Check Phase 3 readiness
	@echo "🔍 Checking Phase 3 readiness..."
	@./scripts/check_phase3_readiness.sh

# Quick demo
demo: ## Quick demo of system
	@echo "🎯 Oil & Gas Inventory System Demo"
	@echo "================================="
	@make health
	@echo ""
	@echo "📊 Sample Data:"
	@cd $(BACKEND_DIR) && docker-compose exec postgres psql -U postgres -d oil_gas_inventory -c "SELECT customer, billing_city, phone FROM store.customers LIMIT 3;"

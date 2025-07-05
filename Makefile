# Oil & Gas Inventory System - Development Commands

.PHONY: help setup migrate seed status clean build test dev-backend dev-frontend

# Default environment
ENV ?= local

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Complete development setup
	@echo "🚀 Setting up Oil & Gas Inventory System..."
	cp .env.local .env
	cp .env.local backend/.env.local
	cp .env backend/.env
	docker-compose up -d postgres
	@echo "⏳ Waiting for database to be ready..."
	sleep 5
	@echo "📦 Installing dependencies..."
	cd backend && go mod tidy
	cd frontend && npm install
	cd backend && go build -o migrator migrator.go
	$(MAKE) migrate seed ENV=$(ENV)
	@echo "✅ Setup complete! Run 'make dev' to start development servers"

migrate: ## Run database migrations
	@echo "🔄 Running migrations for $(ENV) environment..."
	cd backend && ./migrator migrate $(ENV)

seed: ## Seed database with data
	@echo "🌱 Seeding database for $(ENV) environment..."
	cd backend && ./migrator seed $(ENV)

status: ## Show migration status
	@echo "📊 Migration status for $(ENV) environment:"
	@if [ ! -f backend/.env.$(ENV) ]; then cp .env.$(ENV) backend/.env.$(ENV) 2>/dev/null || cp .env.local backend/.env.$(ENV); fi
	cd backend && ./migrator status $(ENV)

reset: ## Reset database (WARNING: Destructive)
	@echo "⚠️  Resetting database for $(ENV) environment..."
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		docker-compose up -d postgres; \
		sleep 5; \
		$(MAKE) migrate seed ENV=$(ENV); \
	fi

build: ## Build all components
	@echo "🔨 Building backend..."
	cd backend && go build -o migrator migrator.go
	cd backend && go build -o server cmd/server/main.go
	@echo "🔨 Building frontend..."
	cd frontend && npm run build

test: ## Run all tests
	@echo "🧪 Running backend tests..."
	cd backend && go test ./...
	@echo "🧪 Running frontend tests..."
	cd frontend && npm run test 2>/dev/null || echo "Add frontend tests"

clean: ## Clean up generated files
	rm -f backend/migrator backend/server
	rm -rf frontend/dist
	docker-compose down -v

# Development commands
dev-start: ## Start development environment (databases)
	docker-compose up -d
	@echo "✅ Development environment started"
	@echo "🐘 PostgreSQL: localhost:5432"
	@echo "🗄️  PgAdmin: http://localhost:8080"

dev-stop: ## Stop development environment
	docker-compose down

dev-backend: ## Start backend development server
	cd backend && go run cmd/server/main.go

dev-frontend: ## Start frontend development server
	@if [ ! -d "frontend/node_modules" ]; then \
		echo "📦 Installing frontend dependencies..."; \
		cd frontend && npm install; \
	fi
	cd frontend && npm run dev

dev: ## Start both backend and frontend (requires 2 terminals)
	@echo "🚀 Starting development servers..."
	@echo "📋 Run these commands in separate terminals:"
	@echo "   Terminal 1: make dev-backend"
	@echo "   Terminal 2: make dev-frontend"
	@echo ""
	@echo "🌐 URLs:"
	@echo "   Frontend: http://localhost:3000"
	@echo "   Backend:  http://localhost:8000"
	@echo "   PgAdmin:  http://localhost:8080"

# Installation commands
install-backend: ## Install backend dependencies
	cd backend && go mod tidy

install-frontend: ## Install frontend dependencies
	cd frontend && npm install

install: install-backend install-frontend ## Install all dependencies

# Production commands
deploy-check: ## Check deployment readiness
	@echo "🔍 Checking deployment readiness..."
	@if [ ! -f .env.prod ]; then echo "❌ .env.prod not found"; exit 1; fi
	@if grep -q "change_me" .env.prod; then echo "❌ Update passwords in .env.prod"; exit 1; fi
	@echo "✅ Deployment checks passed"

# Database commands
db-backup: ## Backup database
	@echo "📦 Creating database backup for $(ENV)..."
	@if [ "$(ENV)" = "local" ]; then \
		docker exec $$(docker-compose ps -q postgres) pg_dump -U postgres oilgas_inventory_local > backup_$(ENV)_$$(date +%Y%m%d_%H%M%S).sql; \
	else \
		echo "Configure production backup command"; \
	fi

# Quick shortcuts
start: setup ## Quick start (alias for setup)
stop: dev-stop ## Quick stop (alias for dev-stop)
restart: dev-stop dev-start ## Restart development environment

# Environment shortcuts
local: ## Run command in local environment
	@$(MAKE) $(CMD) ENV=local

dev-env: ## Run command in dev environment
	@$(MAKE) $(CMD) ENV=dev

prod: ## Run command in prod environment
	@$(MAKE) $(CMD) ENV=prod

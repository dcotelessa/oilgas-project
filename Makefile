# Oil & Gas Inventory System - Development Commands
# Updated for actual repository structure

.PHONY: help setup migrate seed test build clean dev

# Default environment
ENV ?= local

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

init-env: ## Initialize environment file from template
	@if [ ! -f .env ]; then \
		if [ -f .env.local ]; then \
			cp .env.local .env; \
			echo "✅ Created .env from .env.local template"; \
		elif [ -f .env.example ]; then \
			cp .env.example .env; \
			echo "✅ Created .env from .env.example template"; \
			echo "⚠️  Please update .env with your actual values"; \
		else \
			echo "❌ No environment template found (.env.local or .env.example)"; \
			exit 1; \
		fi \
	else \
		echo "✅ .env file already exists"; \
	fi

# Environment validation
check-env: ## Validate environment configuration
	@echo "🔍 Checking environment configuration..."
	@if [ ! -f .env ]; then echo "❌ .env file not found - run 'make init-env'"; exit 1; fi
	@if ! grep -q "DATABASE_URL" .env; then echo "❌ DATABASE_URL not set in .env"; exit 1; fi
	@if ! grep -q "APP_PORT" .env; then echo "❌ APP_PORT not set in .env"; exit 1; fi
	@echo "✅ Environment configuration looks good"

# Setup and development  
setup: init-env check-env ## Complete development setup
	@echo "🚀 Setting up Oil & Gas Inventory System..."
	@if [ ! -f .env ]; then cp .env.local .env; fi
	docker-compose up -d postgres
	@echo "⏳ Waiting for database to be ready..."
	sleep 5
	@echo "📦 Installing dependencies..."
	cd backend && go mod tidy
	cd frontend && npm install
	cd backend && go build -o migrator migrator.go
	$(MAKE) migrate seed ENV=$(ENV)
	@echo "✅ Setup complete! Run 'make dev' to start development servers"

# Database operations
migrate: ## Run database migrations
	@echo "🔄 Running migrations for $(ENV) environment..."
	cd backend && ./migrator migrate $(ENV)

seed: ## Seed database with data
	@echo "🌱 Seeding database for $(ENV) environment..."
	cd backend && ./migrator seed $(ENV)

status: ## Show migration status
	@echo "📊 Migration status for $(ENV) environment:"
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

# Testing commands for your actual repository structure
test: ## Run all tests
	@echo "🧪 Running all tests..."
	$(MAKE) test-unit
	$(MAKE) test-integration

test-unit: ## Run unit tests
	@echo "🔬 Running unit tests..."
	cd backend && go test ./test/unit/... -v -race

test-integration: ## Run integration tests
	@echo "🔗 Running integration tests..."
	@echo "⚠️  Requires test database setup"
	cd backend && go test ./test/integration/... -v
	cd backend && go test ./test -run TestIntegration -v

test-api: ## Run API endpoint tests
	@echo "🌐 Running API tests..."
	cd backend && go test ./test/api/... -v

test-repos: ## Test repository implementations
	@echo "🗄️  Testing repositories..."
	cd backend && go test ./internal/repository/... -v

test-services: ## Test service layer
	@echo "⚙️  Testing services..."
	cd backend && go test ./internal/services/... -v

test-handlers: ## Test HTTP handlers
	@echo "📡 Testing handlers..."
	cd backend && go test ./internal/handlers/... -v

test-validation: ## Test validation logic
	@echo "✅ Testing validation..."
	cd backend && go test ./pkg/validation/... -v

test-cache: ## Test cache functionality
	@echo "💾 Testing cache..."
	cd backend && go test ./pkg/cache/... -v

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	cd backend && go test ./... -coverprofile=coverage.out
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "📈 Coverage report generated: backend/coverage.html"

test-race: ## Run tests with race detection
	@echo "🏃 Running tests with race detection..."
	cd backend && go test ./... -race

test-short: ## Run only fast tests
	@echo "⚡ Running short tests..."
	cd backend && go test ./... -short

test-verbose: ## Run tests with verbose output
	@echo "📝 Running tests with verbose output..."
	cd backend && go test ./... -v

# Benchmarking
benchmark: ## Run performance benchmarks
	@echo "⚡ Running benchmarks..."
	cd backend && go test ./test/benchmark/... -bench=. -benchmem

benchmark-repo: ## Benchmark repository performance
	@echo "🗄️  Benchmarking repositories..."
	cd backend && go test ./test/benchmark/... -bench=BenchmarkRepository -benchmem

benchmark-cache: ## Benchmark cache performance
	@echo "💾 Benchmarking cache..."
	cd backend && go test ./pkg/cache/... -bench=. -benchmem

# Test database setup
test-db-setup: ## Setup test database
	@echo "🗄️  Setting up test database..."
	docker run --name oilgas-test-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=oilgas_inventory_test -p 5433:5432 -d postgres:15-alpine
	sleep 5
	@echo "✅ Test database ready on port 5433"

test-db-cleanup: ## Remove test database
	@echo "🧹 Cleaning up test database..."
	docker stop oilgas-test-db || true
	docker rm oilgas-test-db || true

# Code quality
lint: ## Run linter
	@echo "🔍 Running linter..."
	cd backend && golangci-lint run ./...

fmt: ## Format code
	@echo "🎨 Formatting code..."
	cd backend && go fmt ./...
	cd frontend && npm run format

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	cd backend && go vet ./...

# Build operations
build: ## Build all components
	@echo "🔨 Building backend..."
	cd backend && go build -o migrator migrator.go
	cd backend && go build -o server cmd/server/main.go
	@echo "🔨 Building frontend..."
	cd frontend && npm run build

build-backend: ## Build backend only
	@echo "🔨 Building backend..."
	cd backend && go build -o migrator migrator.go
	cd backend && go build -o server cmd/server/main.go

build-frontend: ## Build frontend only
	@echo "🔨 Building frontend..."
	cd frontend && npm run build

# Development servers
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

# Dependency management
deps-backend: ## Install backend dependencies
	cd backend && go mod tidy && go mod download

deps-frontend: ## Install frontend dependencies
	cd frontend && npm install

deps: deps-backend deps-frontend ## Install all dependencies

deps-update: ## Update dependencies
	cd backend && go get -u ./...
	cd frontend && npm update

# Testing automation
test-watch: ## Watch files and run tests automatically
	@echo "👀 Watching for changes and running tests..."
	cd backend && find . -name "*.go" | entr -r make test-unit

test-watch-integration: ## Watch and run integration tests
	@echo "👀 Watching for changes and running integration tests..."
	cd backend && find . -name "*.go" | entr -r make test-integration

# Continuous testing
test-ci: ## Run tests in CI mode
	@echo "🤖 Running CI tests..."
	$(MAKE) test-db-setup
	sleep 5
	$(MAKE) test-unit
	$(MAKE) test-integration
	$(MAKE) test-coverage
	$(MAKE) test-db-cleanup

# Test data management
test-seed: ## Seed test database with fake data
	@echo "🌱 Seeding test database..."
	ENV=test $(MAKE) seed

test-reset: ## Reset test database
	@echo "🔄 Resetting test database..."
	ENV=test $(MAKE) reset

# Repository-specific tests (matching your structure)
test-analytics: ## Test analytics repository
	cd backend && go test ./internal/repository/*analytics* -v

test-customer: ## Test customer repository
	cd backend && go test ./internal/repository/*customer* -v

test-grade: ## Test grade repository
	cd backend && go test ./internal/repository/*grade* -v

test-inventory: ## Test inventory repository
	cd backend && go test ./internal/repository/*inventory* -v

test-received: ## Test received repository
	cd backend && go test ./internal/repository/*received* -v

test-workflow: ## Test workflow state repository
	cd backend && go test ./internal/repository/*workflow* -v

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	cd backend && godoc -http=:6060 &
	@echo "📖 Documentation available at http://localhost:6060"

# Clean up
clean: ## Clean up generated files
	rm -f backend/migrator backend/server
	rm -rf frontend/dist backend/coverage.out backend/coverage.html
	docker-compose down -v

clean-cache: ## Clean Go module cache
	go clean -modcache

clean-all: clean clean-cache ## Clean everything

# Quick shortcuts for your repository structure
quick-test: test-unit ## Quick unit tests only
quick-build: build-backend ## Quick backend build
quick-start: dev-start migrate seed ## Quick environment start

# Production commands
deploy-check: ## Check deployment readiness
	@echo "🔍 Checking deployment readiness..."
	@if [ ! -f .env.prod ]; then echo "❌ .env.prod not found"; exit 1; fi
	@if grep -q "change_me" .env.prod; then echo "❌ Update passwords in .env.prod"; exit 1; fi
	$(MAKE) test-ci
	$(MAKE) build
	@echo "✅ Deployment checks passed"

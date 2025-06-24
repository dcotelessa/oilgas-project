#!/bin/bash

# Complete Project Organization Script v10
# Moves migration_output files to proper project structure and creates all environment files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[ORGANIZE]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
info() { echo -e "${BLUE}[INFO]${NC} $1"; }

# Configuration
MIGRATION_DIR="${MIGRATION_DIR:-./migration_output}"
PROJECT_ROOT="."
BACKEND_DIR="backend"
FRONTEND_DIR="frontend"
DATABASE_DIR="database"

# Check if migration_output exists
check_migration_output() {
    if [[ ! -d "$MIGRATION_DIR" ]]; then
        error "Migration output directory not found: $MIGRATION_DIR"
        echo "Run the migration script first: ./cf_migration_script.sh"
        exit 1
    fi
    
    log "Found migration output directory: $MIGRATION_DIR"
}

# Create project structure
create_project_structure() {
    log "Creating project directory structure..."
    
    # Backend structure
    mkdir -p "$BACKEND_DIR"/{cmd/server,internal/{handlers,services,repository,models},pkg}
    
    # Frontend structure  
    mkdir -p "$FRONTEND_DIR"/{src/{components,views,stores,utils},public}
    
    # Database structure
    mkdir -p "$DATABASE_DIR"/{schema,data,analysis}
    
    # Documentation
    mkdir -p docs
    
    log "✅ Project structure created"
}

# Move root-level files
move_root_files() {
    log "Moving root-level configuration files..."
    
    local root_files=(
        "docker-compose.yml"
        ".env.local"
        ".env.dev" 
        ".env.prod"
        ".env.example"
        "Makefile"
    )
    
    for file in "${root_files[@]}"; do
        if [[ -f "$MIGRATION_DIR/$file" ]]; then
            log "Moving $file to project root"
            mv "$MIGRATION_DIR/$file" "$PROJECT_ROOT/"
        else
            warn "File not found: $MIGRATION_DIR/$file"
        fi
    done
    
    log "✅ Root files moved"
}

# Move backend files
move_backend_files() {
    log "Moving backend files..."
    
    # Go files
    if [[ -f "$MIGRATION_DIR/migrator.go" ]]; then
        log "Moving migrator.go to backend/"
        mv "$MIGRATION_DIR/migrator.go" "$BACKEND_DIR/"
    fi
    
    if [[ -f "$MIGRATION_DIR/go.mod" ]]; then
        log "Moving go.mod to backend/"
        mv "$MIGRATION_DIR/go.mod" "$BACKEND_DIR/"
    fi
    
    # Migrations directory
    if [[ -d "$MIGRATION_DIR/migrations" ]]; then
        log "Moving migrations/ to backend/"
        mv "$MIGRATION_DIR/migrations" "$BACKEND_DIR/"
    fi
    
    # Seeds directory
    if [[ -d "$MIGRATION_DIR/seeds" ]]; then
        log "Moving seeds/ to backend/"
        mv "$MIGRATION_DIR/seeds" "$BACKEND_DIR/"
    fi
    
    # Cache package
    if [[ -d "$MIGRATION_DIR/pkg" ]]; then
        log "Moving pkg/ to backend/"
        mv "$MIGRATION_DIR/pkg"/* "$BACKEND_DIR/pkg/" 2>/dev/null || true
        rmdir "$MIGRATION_DIR/pkg" 2>/dev/null || true
    fi
    
    log "✅ Backend files moved"
}

# Move database files
move_database_files() {
    log "Moving database reference files..."
    
    # Schema files
    if [[ -d "$MIGRATION_DIR/schema" ]]; then
        log "Moving schema/ to database/"
        mv "$MIGRATION_DIR/schema"/* "$DATABASE_DIR/schema/" 2>/dev/null || true
        rmdir "$MIGRATION_DIR/schema" 2>/dev/null || true
    fi
    
    # Data files
    if [[ -d "$MIGRATION_DIR/data" ]]; then
        log "Moving data/ to database/"
        mv "$MIGRATION_DIR/data"/* "$DATABASE_DIR/data/" 2>/dev/null || true
        rmdir "$MIGRATION_DIR/data" 2>/dev/null || true
    fi
    
    # Analysis files
    if [[ -d "$MIGRATION_DIR/analysis" ]]; then
        log "Moving analysis/ to database/"
        mv "$MIGRATION_DIR/analysis"/* "$DATABASE_DIR/analysis/" 2>/dev/null || true
        rmdir "$MIGRATION_DIR/analysis" 2>/dev/null || true
    fi
    
    log "✅ Database files moved"
}

# Create environment files
create_env_files() {
    log "Creating environment configuration files..."
    
    # .env.local (development with fake data)
    cat > "$PROJECT_ROOT/.env.local" << 'EOF'
# Local Development Environment
APP_ENV=local
APP_PORT=8000
APP_DEBUG=true

# Database Configuration (Docker)
POSTGRES_DB=oilgas_inventory_local
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres123
DATABASE_URL=postgres://postgres:postgres123@localhost:5432/oilgas_inventory_local

# In-Memory Cache Configuration
CACHE_TTL=300s
CACHE_CLEANUP_INTERVAL=600s
CACHE_MAX_SIZE=1000

# PgAdmin Configuration
PGADMIN_EMAIL=admin@localhost.dev
PGADMIN_PASSWORD=admin123

# JWT Configuration
JWT_SECRET=local_jwt_secret_key_not_for_production
JWT_EXPIRES_IN=24h

# CORS Configuration
CORS_ORIGINS=http://localhost:3000,http://localhost:5173

# Logging
LOG_LEVEL=debug
LOG_FORMAT=text

# Feature Flags
ENABLE_METRICS=true
ENABLE_PROFILING=true
EOF

    # .env.dev (development/staging with real data)
    cat > "$PROJECT_ROOT/.env.dev" << 'EOF'
# Development/Staging Environment
APP_ENV=development
APP_PORT=8000
APP_DEBUG=true

# Database Configuration (Vultr or managed)
POSTGRES_DB=oilgas_inventory_dev
POSTGRES_USER=oilgas_user
POSTGRES_PASSWORD=secure_dev_password_change_me
DATABASE_URL=postgres://oilgas_user:secure_dev_password_change_me@dev-db-host:5432/oilgas_inventory_dev

# In-Memory Cache Configuration
CACHE_TTL=60s
CACHE_CLEANUP_INTERVAL=300s
CACHE_MAX_SIZE=5000

# JWT Configuration
JWT_SECRET=dev_jwt_secret_key_change_me_in_production
JWT_EXPIRES_IN=8h

# CORS Configuration
CORS_ORIGINS=https://dev.yourapp.com,https://staging.yourapp.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# External Services
VULTR_API_KEY=your_vultr_api_key
BACKUP_BUCKET=dev-oilgas-backups

# Feature Flags
ENABLE_METRICS=true
ENABLE_PROFILING=false
EOF

    # .env.prod (production)
    cat > "$PROJECT_ROOT/.env.prod" << 'EOF'
# Production Environment
APP_ENV=production
APP_PORT=8000
APP_DEBUG=false

# Database Configuration (Vultr Managed Database recommended)
POSTGRES_DB=oilgas_inventory_prod
POSTGRES_USER=oilgas_prod_user
POSTGRES_PASSWORD=very_secure_production_password_change_me
DATABASE_URL=postgres://oilgas_prod_user:very_secure_production_password_change_me@prod-db-host:5432/oilgas_inventory_prod

# In-Memory Cache Configuration (Production tuned)
CACHE_TTL=30s
CACHE_CLEANUP_INTERVAL=180s
CACHE_MAX_SIZE=10000

# JWT Configuration
JWT_SECRET=production_jwt_secret_key_must_be_very_secure
JWT_EXPIRES_IN=2h

# CORS Configuration
CORS_ORIGINS=https://yourapp.com,https://www.yourapp.com

# Logging
LOG_LEVEL=warn
LOG_FORMAT=json

# External Services
VULTR_API_KEY=your_production_vultr_api_key
BACKUP_BUCKET=prod-oilgas-backups

# Monitoring
SENTRY_DSN=your_sentry_dsn
METRICS_ENDPOINT=https://metrics.yourapp.com

# Feature Flags
ENABLE_METRICS=true
ENABLE_PROFILING=false
ENABLE_RATE_LIMITING=true
EOF

    # .env.example
    cat > "$PROJECT_ROOT/.env.example" << 'EOF'
# Example Environment Configuration
# Copy this file to .env.local and modify for your setup

# Database Configuration
POSTGRES_DB=oilgas_inventory
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password
DATABASE_URL=postgres://postgres:your_secure_password@localhost:5432/oilgas_inventory

# PgAdmin Configuration
PGADMIN_EMAIL=admin@yourdomain.com
PGADMIN_PASSWORD=your_admin_password

# Application Configuration
APP_ENV=development
APP_PORT=8000
JWT_SECRET=your_jwt_secret_key
CACHE_TTL=300s
CACHE_CLEANUP_INTERVAL=600s
CACHE_MAX_SIZE=1000

# External Services
VULTR_API_KEY=your_vultr_api_key
EOF

    log "✅ Environment configuration files created"
}

# Create frontend starter files
create_frontend_files() {
    log "Creating frontend starter files..."
    
    # Package.json
    cat > "$FRONTEND_DIR/package.json" << 'EOF'
{
  "name": "oilgas-inventory-frontend",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview",
    "type-check": "vue-tsc --noEmit"
  },
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.2.0",
    "pinia": "^2.1.0",
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "typescript": "^5.3.0",
    "vue-tsc": "^1.8.0",
    "vite": "^5.0.0",
    "@types/node": "^20.10.0"
  }
}
EOF

    # Vite config
    cat > "$FRONTEND_DIR/vite.config.ts" << 'EOF'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
    },
  },
})
EOF

    # TypeScript config
    cat > "$FRONTEND_DIR/tsconfig.json" << 'EOF'
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.vue"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
EOF

    # TypeScript Node config (for Vite config files)
    cat > "$FRONTEND_DIR/tsconfig.node.json" << 'EOF'
{
  "compilerOptions": {
    "composite": true,
    "skipLibCheck": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true
  },
  "include": ["vite.config.ts"]
}
EOF

    # Main Vue app
    cat > "$FRONTEND_DIR/src/App.vue" << 'EOF'
<template>
  <div id="app">
    <header>
      <h1>Oil & Gas Inventory System</h1>
    </header>
    
    <nav>
      <router-link to="/">Dashboard</router-link>
      <router-link to="/customers">Customers</router-link>
      <router-link to="/inventory">Inventory</router-link>
      <router-link to="/grades">Grades</router-link>
    </nav>
    
    <main>
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
// Vue 3 Composition API setup
</script>

<style scoped>
header {
  background: #2c3e50;
  color: white;
  padding: 1rem;
  text-align: center;
}

nav {
  background: #34495e;
  padding: 1rem;
}

nav a {
  color: white;
  text-decoration: none;
  margin-right: 1rem;
  padding: 0.5rem;
}

nav a:hover {
  background: #4a5f7a;
  border-radius: 4px;
}

main {
  padding: 2rem;
}
</style>
EOF

    # Main TypeScript file
    cat > "$FRONTEND_DIR/src/main.ts" << 'EOF'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'

// Import views
import Dashboard from './views/Dashboard.vue'
import Customers from './views/Customers.vue'
import Inventory from './views/Inventory.vue'
import Grades from './views/Grades.vue'

// Router setup
const routes = [
  { path: '/', component: Dashboard },
  { path: '/customers', component: Customers },
  { path: '/inventory', component: Inventory },
  { path: '/grades', component: Grades },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// App setup
const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
EOF

    # Create placeholder views
    local views=("Dashboard" "Customers" "Inventory" "Grades")
    for view in "${views[@]}"; do
        local view_lower=$(echo "$view" | tr '[:upper:]' '[:lower:]')
        cat > "$FRONTEND_DIR/src/views/${view}.vue" << EOF
<template>
  <div class="${view_lower}">
    <h2>${view}</h2>
    <p>Welcome to the ${view} page.</p>
    <!-- Add your ${view_lower} content here -->
  </div>
</template>

<script setup lang="ts">
// ${view} component logic
</script>

<style scoped>
.${view_lower} {
  padding: 1rem;
}
</style>
EOF
    done

    # Index.html
    cat > "$FRONTEND_DIR/index.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Oil & Gas Inventory System</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
EOF

    log "✅ Frontend starter files created"
}

# Create backend starter files
create_backend_files() {
    log "Creating backend starter files..."
    
    # Main server file
    cat > "$BACKEND_DIR/cmd/server/main.go" << 'EOF'
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from project root or local
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found")
		}
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	// Set Gin mode based on environment
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	r := gin.Default()

	// Configure trusted proxies (fix the warning)
	// In development, trust localhost and docker networks
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	// CORS middleware
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// In development, allow localhost origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173", // Vite dev server
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "oil-gas-inventory",
			"version": "1.0.0",
			"environment": os.Getenv("APP_ENV"),
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/grades", getGrades)
		api.GET("/customers", getCustomers)
		api.GET("/inventory", getInventory)
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("🌍 Environment: %s", os.Getenv("APP_ENV"))
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Fatal(r.Run(":" + port))
}

// Placeholder handlers
func getGrades(c *gin.Context) {
	grades := []string{"J55", "JZ55", "L80", "N80", "P105", "P110"}
	c.JSON(http.StatusOK, gin.H{"grades": grades})
}

func getCustomers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"customers": []interface{}{}})
}

func getInventory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"inventory": []interface{}{}})
}
EOF

    # Update go.mod for backend if it doesn't exist from migration
    if [[ ! -f "$BACKEND_DIR/go.mod" ]]; then
        cat > "$BACKEND_DIR/go.mod" << 'EOF'
module oilgas-inventory-backend

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/jackc/pgx/v5 v5.4.3
	github.com/joho/godotenv v1.4.0
)

require (
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
EOF
    fi

    log "✅ Backend starter files created"
}

# Create comprehensive Makefile
create_makefile() {
    log "Creating comprehensive Makefile..."
    
    cat > "$PROJECT_ROOT/Makefile" << 'EOF'
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
	@if [ ! -f backend/.env.$(ENV) ]; then cp .env.$(ENV) backend/.env.$(ENV) 2>/dev/null || cp .env.local backend/.env.$(ENV); fi
	cd backend && ./migrator migrate $(ENV)

seed: ## Seed database with data
	@echo "🌱 Seeding database for $(ENV) environment..."
	@if [ ! -f backend/.env.$(ENV) ]; then cp .env.$(ENV) backend/.env.$(ENV) 2>/dev/null || cp .env.local backend/.env.$(ENV); fi
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
EOF

    log "✅ Makefile created"
}

# Create Docker configuration
create_docker_config() {
    log "Creating Docker configuration..."
    
    # Main docker-compose.yml
    cat > "$PROJECT_ROOT/docker-compose.yml" << 'EOF'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-oilgas_inventory_local}
      POSTGRES_USER: ${POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-postgres123}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    networks:
      - app-network

  pgadmin:
    image: dpage/pgadmin4:latest
    environment:
      PGADMIN_DEFAULT_EMAIL: ${PGADMIN_EMAIL:-admin@localhost.dev}
      PGADMIN_DEFAULT_PASSWORD: ${PGADMIN_PASSWORD:-admin123}
    ports:
      - "8080:80"
    depends_on:
      - postgres
    networks:
      - app-network

volumes:
  postgres_data:

networks:
  app-network:
    driver: bridge
EOF

    log "✅ Docker configuration created"
}

# Create documentation
create_documentation() {
    log "Creating project documentation..."
    
    # Main README
    cat > "$PROJECT_ROOT/README.md" << 'EOF'
# Oil & Gas Inventory System

A modern inventory management system for the oil & gas industry, migrated from ColdFusion to a modern stack.

## ⚠️ Data Security Notice

This repository contains **ONLY** mock/fake data for development purposes. All sensitive customer information, production databases, and real business data are protected by .gitignore and should never be committed to version control.

## Tech Stack

- **Backend**: Go + Gin + PostgreSQL + In-Memory Caching
- **Frontend**: Vue.js 3 + TypeScript + Vite + Pinia
- **Database**: PostgreSQL 15
- **Infrastructure**: Docker + Docker Compose
- **Deployment**: Vultr VPS

## Project Structure

```
├── backend/                 # Go backend application
│   ├── cmd/server/         # Main application entry
│   ├── internal/           # Private application code
│   ├── pkg/               # Public packages (cache, utils)
│   ├── migrations/        # Database migrations
│   └── seeds/            # Database seed data (FAKE DATA ONLY)
├── frontend/              # Vue.js frontend application
│   ├── src/              # Source code
│   └── public/           # Static assets
├── database/             # Database reference files (schema only)
│   ├── schema/          # PostgreSQL schema
│   └── analysis/       # Migration analysis (no sensitive data)

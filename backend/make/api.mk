# =============================================================================
# API MODULE - make/api.mk
# =============================================================================
# API server management, endpoints testing, and documentation

.PHONY: api-setup api-dev api-build api-test api-docs

# =============================================================================
# API LIFECYCLE
# =============================================================================

api-setup: ## 🛠️  Setup API environment
	@echo "$(GREEN)🌐 Setting up API environment...$(RESET)"
	@mkdir -p logs/api tmp/uploads
	@echo "$(GREEN)✅ API environment ready$(RESET)"

api-dev: ## 🛠️  Start development API server
	@echo "$(GREEN)🚀 Starting API server...$(RESET)"
	@echo "$(BLUE)API: http://localhost:$(API_PORT)$(RESET)"
	@echo "$(BLUE)Health: http://localhost:$(API_PORT)/health$(RESET)"
	@echo "$(BLUE)Admin: http://localhost:$(API_PORT)/admin/health$(RESET)"
	@go run cmd/server/main.go

api-build: ## 🛠️  Build production API binary
	@echo "$(YELLOW)🔨 Building production API binary...$(RESET)"
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api-server cmd/server/main.go
	@echo "$(GREEN)✅ Production binary built: bin/api-server$(RESET)"

api-run-prod: api-build ## 🛠️  Run production binary
	@echo "$(GREEN)🚀 Starting production API server...$(RESET)"
	@./bin/api-server

# =============================================================================
# API TESTING
# =============================================================================

api-test: ## 🧪 Test API endpoints
	@echo "$(YELLOW)🧪 Testing API endpoints...$(RESET)"
	@go test -v ./internal/handlers/...
	@echo "$(GREEN)✅ API endpoint tests complete$(RESET)"

api-test-health: ## 🧪 Test health endpoint
	@echo "$(YELLOW)🧪 Testing health endpoint...$(RESET)"
	@curl -s http://localhost:$(API_PORT)/health | jq . || \
		echo "$(RED)❌ Health endpoint failed$(RESET)"

api-test-auth: ## 🧪 Test authentication endpoints
	@echo "$(YELLOW)🧪 Testing authentication endpoints...$(RESET)"
	@echo "Testing login endpoint..."
	@curl -X POST http://localhost:$(API_PORT)/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"test@example.com","password":"testpass"}' | jq .
	@echo ""

api-test-tenant: ## 🧪 Test tenant-aware endpoints
	@echo "$(YELLOW)🧪 Testing tenant endpoints...$(RESET)"
	@read -p "Tenant ID: " tenant && \
	read -p "Session ID (from login): " session && \
	echo "Testing customers endpoint..." && \
	curl -H "X-Tenant: $$tenant" -H "X-Session-ID: $$session" \
		"http://localhost:$(API_PORT)/api/v1/customers" | jq .

api-benchmark: ## 📊 Benchmark API performance
	@echo "$(YELLOW)📊 Benchmarking API performance...$(RESET)"
	@go test -bench=. -benchmem ./internal/handlers/...

# =============================================================================
# API DOCUMENTATION
# =============================================================================

api-docs: ## 📖 Generate API documentation
	@echo "$(YELLOW)📖 Generating API documentation...$(RESET)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs/swagger; \
		echo "$(GREEN)✅ Swagger docs generated at docs/swagger$(RESET)"; \
	else \
		echo "$(YELLOW)💡 Install swag: go install github.com/swaggo/swag/cmd/swag@latest$(RESET)"; \
		echo "$(BLUE)Generating basic docs...$(RESET)"; \
		go run cmd/tools/generate-docs.go; \
	fi

api-docs-serve: ## 📖 Serve API documentation
	@echo "$(BLUE)📖 Serving API documentation...$(RESET)"
	@echo "$(BLUE)Docs: http://localhost:8080$(RESET)"
	@if [ -d docs/swagger ]; then \
		cd docs/swagger && python3 -m http.server 8080; \
	else \
		echo "$(YELLOW)💡 Generate docs first: make api-docs$(RESET)"; \
	fi

# =============================================================================
# API UTILITIES
# =============================================================================

api-clean: ## 🛠️  Clean API artifacts
	@echo "$(YELLOW)🧹 Cleaning API artifacts...$(RESET)"
	@rm -rf bin/api-server
	@rm -rf logs/api/*.log
	@rm -rf tmp/uploads/*
	@echo "$(GREEN)✅ API cleanup complete$(RESET)"

help-api: ## 📖 Show API commands help
	@echo "$(BLUE)API Module Commands$(RESET)"
	@echo "==================="
	@echo ""
	@echo "$(GREEN)🛠️  LIFECYCLE:$(RESET)"
	@echo "  api-setup        - Setup API environment"
	@echo "  api-dev          - Start development server"
	@echo "  api-build        - Build production binary"
	@echo "  api-run-prod     - Run production binary"
	@echo ""
	@echo "$(YELLOW)🧪 TESTING:$(RESET)"
	@echo "  api-test         - Test API endpoints"
	@echo "  api-test-health  - Test health endpoint"
	@echo "  api-test-auth    - Test auth endpoints"
	@echo "  api-test-tenant  - Test tenant endpoints"
	@echo "  api-benchmark    - Benchmark performance"
	@echo ""
	@echo "$(BLUE)📖 DOCUMENTATION:$(RESET)"
	@echo "  api-docs         - Generate documentation"
	@echo "  api-docs-serve   - Serve documentation"

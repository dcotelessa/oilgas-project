# =============================================================================
# DEVELOPMENT MODULE - make/development.mk
# =============================================================================
# Development workflow and code quality commands

.PHONY: dev-setup dev-build dev-run dev-watch dev-lint dev-format dev-deps

# =============================================================================
# DEVELOPMENT ENVIRONMENT
# =============================================================================

dev-setup: install-deps ## 🛠️  Setup development environment
	@echo "$(GREEN)🛠️  Setting up development environment...$(RESET)"
	@mkdir -p logs tmp bin
	@echo "$(GREEN)✅ Development environment ready$(RESET)"

dev-build: ## 🛠️  Build development binary
	@echo "$(YELLOW)🔨 Building development binary...$(RESET)"
	@go build -race -o bin/server-dev cmd/server/main.go
	@echo "$(GREEN)✅ Development binary built: bin/server-dev$(RESET)"

dev-run: dev-build ## 🛠️  Run development server manually
	@echo "$(GREEN)🚀 Starting development server...$(RESET)"
	@echo "$(BLUE)API: http://localhost:$(API_PORT)$(RESET)"
	@./bin/server-dev

dev-watch: ## 🛠️  Run with auto-reload (requires air)
	@echo "$(YELLOW)🔄 Starting auto-reload development server...$(RESET)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(RED)Air not installed. Install with: go install github.com/cosmtrek/air@latest$(RESET)"; \
		echo "$(YELLOW)Falling back to manual run...$(RESET)"; \
		$(MAKE) dev-run; \
	fi

# =============================================================================
# CODE QUALITY
# =============================================================================

dev-lint: ## 🛠️  Run code linting
	@echo "$(YELLOW)🔍 Running code linting...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(GREEN)✅ Linting complete$(RESET)"; \
	else \
		echo "$(RED)golangci-lint not installed$(RESET)"; \
		echo "$(YELLOW)Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)"; \
		echo "$(BLUE)Running basic go vet instead...$(RESET)"; \
		go vet ./...; \
	fi

dev-format: ## 🛠️  Format code
	@echo "$(YELLOW)🎨 Formatting code...$(RESET)"
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
		echo "$(GREEN)✅ Code formatted with goimports$(RESET)"; \
	else \
		echo "$(GREEN)✅ Code formatted with go fmt$(RESET)"; \
		echo "$(BLUE)💡 Install goimports for better formatting: go install golang.org/x/tools/cmd/goimports@latest$(RESET)"; \
	fi

dev-deps: ## 🛠️  Analyze dependencies
	@echo "$(BLUE)📦 Dependency Analysis$(RESET)"
	@echo "Direct dependencies:"
	@go list -m all | grep -v "$(shell go list -m)" | head -10
	@echo ""
	@echo "Dependency count: $$(go list -m all | wc -l)"
	@echo "Module size: $$(du -sh go.mod go.sum 2>/dev/null || echo 'N/A')"

# =============================================================================
# DEVELOPMENT UTILITIES
# =============================================================================

dev-clean: ## 🛠️  Clean development artifacts
	@echo "$(YELLOW)🧹 Cleaning development artifacts...$(RESET)"
	@rm -rf bin/server-dev
	@rm -rf tmp/*
	@rm -rf logs/dev-*.log
	@echo "$(GREEN)✅ Development cleanup complete$(RESET)"

dev-reset: dev-clean dev-setup ## 🛠️  Reset development environment
	@echo "$(GREEN)🔄 Development environment reset$(RESET)"

dev-info: ## 🛠️  Show development environment info
	@echo "$(BLUE)🛠️  Development Environment Information$(RESET)"
	@echo "========================================"
	@echo "Go version: $$(go version)"
	@echo "GOPATH: $$GOPATH"
	@echo "GOROOT: $$GOROOT"
	@echo "Module: $$(go list -m)"
	@echo "Environment: $(ENV)"
	@echo "API Port: $(API_PORT)"
	@echo "Database: $(DB_NAME)"
	@echo ""
	@echo "Available tools:"
	@command -v air >/dev/null 2>&1 && echo "  ✅ air (auto-reload)" || echo "  ❌ air (auto-reload)"
	@command -v golangci-lint >/dev/null 2>&1 && echo "  ✅ golangci-lint" || echo "  ❌ golangci-lint"
	@command -v goimports >/dev/null 2>&1 && echo "  ✅ goimports" || echo "  ❌ goimports"

# =============================================================================
# DEVELOPMENT SHORTCUTS
# =============================================================================

dev: dev-watch ## 🛠️  Start development (alias for dev-watch)

restart: ## 🛠️  Quick restart development server
	@echo "$(YELLOW)🔄 Restarting development server...$(RESET)"
	@pkill -f "bin/server-dev" 2>/dev/null || true
	@sleep 1
	@$(MAKE) dev-run

# =============================================================================
# HELP
# =============================================================================

help-development: ## 📖 Show development commands help
	@echo "$(BLUE)Development Module Commands$(RESET)"
	@echo "============================="
	@echo ""
	@echo "$(GREEN)🛠️  DEVELOPMENT WORKFLOW:$(RESET)"
	@echo "  dev-setup     - Setup development environment"
	@echo "  dev           - Start development with auto-reload"
	@echo "  dev-build     - Build development binary"  
	@echo "  dev-run       - Run development server manually"
	@echo "  dev-watch     - Run with auto-reload (requires air)"
	@echo "  restart       - Quick restart development server"
	@echo ""
	@echo "$(YELLOW)🎨 CODE QUALITY:$(RESET)"
	@echo "  dev-lint      - Run code linting"
	@echo "  dev-format    - Format code"
	@echo "  dev-deps      - Analyze dependencies"
	@echo ""
	@echo "$(RED)🧹 CLEANUP:$(RESET)"
	@echo "  dev-clean     - Clean development artifacts"
	@echo "  dev-reset     - Reset development environment"
	@echo ""
	@echo "$(BLUE)ℹ️  INFORMATION:$(RESET)"
	@echo "  dev-info      - Show development environment info"

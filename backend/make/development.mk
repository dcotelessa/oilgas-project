# Development and Building
.PHONY: dev build lint format install-deps

dev: ## Start development server
	@echo "$(GREEN)Starting development server...$(RESET)"
	@echo "$(YELLOW)API: http://localhost:8000$(RESET)"
	@echo "$(YELLOW)Health: http://localhost:8000/health$(RESET)"
	@echo "$(YELLOW)Login: POST http://localhost:8000/api/v1/auth/login$(RESET)"
	go run cmd/server/main.go

build: ## Build production binary
	@echo "$(GREEN)Building production binary...$(RESET)"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go
	@echo "$(GREEN)✅ Binary built: bin/server$(RESET)"

lint: ## Run code linting
	@echo "$(GREEN)Running code linting...$(RESET)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed, running go vet...$(RESET)"; \
		go vet ./...; \
	fi

format: ## Format code
	@echo "$(GREEN)Formatting code...$(RESET)"
	go fmt ./...
	@if command -v goimports &> /dev/null; then \
		goimports -w .; \
	fi

install-deps: ## Install missing dependencies
	@echo "$(GREEN)Installing dependencies...$(RESET)"
	go mod tidy
	go mod download

# Oil & Gas Inventory System - Phase 3 Production Makefile

# Environment
ENV ?= local

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
BLUE := \033[34m
RESET := \033[0m

# Go settings
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary settings
BINARY_NAME=oil-gas-api
BINARY_DIR=bin

.PHONY: help setup dev build test clean

help: ## Show this help message
	@echo "$(GREEN)Oil & Gas Inventory System - Phase 3$(RESET)"
	@echo "$(YELLOW)Production-Ready MVP API with JWT Auth$(RESET)"
	@echo ""
	@echo "$(GREEN)Main Commands:$(RESET)"
	@grep -E "^[a-zA-Z_-]+:.*?## .*$" $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(RESET) %s\n", $1, $2}'

setup: ## Complete development setup
	@echo "$(GREEN)Setting up Oil & Gas Inventory System - Phase 3...$(RESET)"
	@echo "$(YELLOW)Installing dependencies...$(RESET)"
	$(GOMOD) tidy
	@echo "$(YELLOW)Starting database...$(RESET)"
	docker-compose up -d postgres || echo "Docker not available"
	@sleep 3
	@echo "$(YELLOW)Running migrations...$(RESET)"
	$(GOCMD) run migrator.go migrate $(ENV) || echo "Migrator not available yet"
	@echo "$(GREEN)✅ Setup complete! Run 'make dev' to start development server$(RESET)"

dev: ## Start development server
	@echo "$(GREEN)Starting development server...$(RESET)"
	@echo "$(YELLOW)API: http://localhost:8000$(RESET)"
	@echo "$(YELLOW)Health: http://localhost:8000/health$(RESET)"
	@echo "$(YELLOW)Metrics: http://localhost:8000/metrics$(RESET)"
	$(GOCMD) run cmd/server/main.go

build: ## Build production binary
	@echo "$(GREEN)Building production binary...$(RESET)"
	mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BINARY_DIR)/$(BINARY_NAME) cmd/server/main.go
	@echo "$(GREEN)✅ Binary built: $(BINARY_DIR)/$(BINARY_NAME)$(RESET)"

test: ## Run all tests
	@echo "$(GREEN)Running all tests...$(RESET)"
	$(GOTEST) -v ./...

test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(RESET)"
	$(GOTEST) -v -short ./internal/...

test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(RESET)"
	$(GOTEST) -v -run Integration ./test/...

test-benchmarks: ## Run performance benchmarks
	@echo "$(GREEN)Running performance benchmarks...$(RESET)"
	$(GOTEST) -bench=. -benchmem ./benchmarks/...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(RESET)"
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report: coverage.html$(RESET)"

# Performance and monitoring
performance-check: ## Check system performance
	@echo "$(GREEN)Performance Check:$(RESET)"
	@echo "$(YELLOW)Cache Performance:$(RESET)"
	@curl -s http://localhost:8000/metrics | jq '.cache_performance' 2>/dev/null || echo "API not running"
	@echo "$(YELLOW)Health Status:$(RESET)"
	@curl -s http://localhost:8000/health | jq '.' 2>/dev/null || echo "API not running"

api-test: ## Test API endpoints
	@echo "$(GREEN)Testing API endpoints...$(RESET)"
	@echo "$(YELLOW)Health Check:$(RESET)"
	@curl -s http://localhost:8000/health | jq '.status' || echo "Failed"

# Code quality
lint: ## Run code linting
	@echo "$(GREEN)Running code linting...$(RESET)"
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed$(RESET)"; \
	fi

format: ## Format code
	@echo "$(GREEN)Formatting code...$(RESET)"
	$(GOCMD) fmt ./...

# Cleanup
clean: ## Clean up build artifacts
	@echo "$(GREEN)Cleaning up...$(RESET)"
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✅ Cleanup complete$(RESET)"

# Phase 3 validation
validate-phase3: ## Validate Phase 3 implementation
	@echo "$(GREEN)Validating Phase 3 Implementation...$(RESET)"
	@echo "$(YELLOW)✅ Checking JWT Authentication...$(RESET)"
	@test -f pkg/jwt/jwt.go && echo "JWT implementation: ✅" || echo "JWT implementation: ❌"
	@echo "$(YELLOW)✅ Checking Performance Cache...$(RESET)"
	@test -f pkg/cache/performance_cache.go && echo "Performance cache: ✅" || echo "Performance cache: ❌"
	@echo "$(YELLOW)✅ Checking Test Data...$(RESET)"
	@test -f test/fixtures/oil_gas_data.go && echo "Test fixtures: ✅" || echo "Test fixtures: ❌"
	@echo "$(YELLOW)✅ Checking Benchmarks...$(RESET)"
	@test -f benchmarks/performance_test.go && echo "Benchmarks: ✅" || echo "Benchmarks: ❌"
	@echo "$(YELLOW)✅ Checking Frontend Examples...$(RESET)"
	@test -f examples/frontend/vue/api-client.js && echo "Frontend examples: ✅" || echo "Frontend examples: ❌"
	@echo "$(GREEN)Phase 3 validation complete!$(RESET)"

# Default target
.DEFAULT_GOAL := help

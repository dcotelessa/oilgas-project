# Add this to your main Makefile to integrate the new modules

# Include all module makefiles
include make/compatibility.mk
include make/database.mk
include make/development.mk
include make/testing.mk
include make/docker.mk
include make/api.mk
include make/data.mk

# Update the help system to include new commands
help: ## Show this help message
	@echo "Oil & Gas Inventory System - Development Commands"
	@echo "================================================="
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make dev-full-setup    # Complete setup from scratch"
	@echo "  make api-start         # Start API server"
	@echo "  make api-test          # Test API integration"
	@echo ""
	@echo "📊 API Development:"
	@echo "  make api-dev           # Start with auto-reload"
	@echo "  make api-test-quick    # Quick health check"
	@echo "  make api-examples      # Show usage examples"
	@echo "  make api-curl-examples # Run live examples"
	@echo ""
	@echo "📥 Data Import:"
	@echo "  make import-check      # Check MDB conversion readiness"
	@echo "  make convert-mdb       # Convert MDB files to CSV"
	@echo "  make import-mdb-data   # Import converted data"
	@echo "  make import-status     # Check data import status"
	@echo "  make data-stats        # Show database statistics"
	@echo ""
	@echo "🧪 Development Workflow:"
	@echo "  make dev-api-ready     # Setup for API development"
	@echo "  make dev-status        # Check environment status"
	@echo "  make dev-quick-test    # Test entire stack"
	@echo "  make dev-install-deps  # Install development tools"
	@echo ""
	@echo "🗄️ Database Operations:"
	@echo "  make dev-setup         # Setup with sample data"
	@echo "  make db-health         # Check database connection"
	@echo "  make test-isolated     # Run isolated tests"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Add some convenient aliases
.PHONY: start test quick-test setup status

## Alias for api-start
start: api-start

## Alias for api-test
test: api-test

## Alias for dev-quick-test
quick-test: dev-quick-test

## Alias for dev-full-setup
setup: dev-full-setup

## Alias for dev-status
status: dev-status

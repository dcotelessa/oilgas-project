# =============================================================================
# API Development Commands
# =============================================================================

.PHONY: api-start api-test api-test-quick api-dev api-logs api-examples api-curl-examples api-check-db

## Start API server in development mode
api-start: dev-ensure-db
	@echo "🚀 Starting API server..."
	@echo "📋 Health: http://localhost:8000/health"
	@echo "🔌 API: http://localhost:8000/api/v1"
	@echo "Press Ctrl+C to stop"
	@cd backend && go run cmd/server/main.go

## Test API integration with repository layer
api-test:
	@echo "🧪 Testing API integration..."
	@test -f scripts/test_api_integration.sh || (echo "❌ Test script missing" && exit 1)
	@chmod +x scripts/test_api_integration.sh
	@scripts/test_api_integration.sh

## Quick API health check
api-test-quick:
	@echo "⚡ Quick API test..."
	@curl -s http://localhost:8000/health | jq -r '"Status: " + .status + " | Service: " + .service' 2>/dev/null || echo "❌ API not responding (is it running?)"

## Start API in development mode with auto-reload
api-dev: dev-ensure-db
	@echo "🔄 Starting API with auto-reload..."
	@which air > /dev/null || (echo "💡 Install air: cd backend && go install github.com/cosmtrek/air@latest" && exit 1)
	@cd backend && air

## Show API usage examples
api-examples:
	@echo "🔍 API Usage Examples"
	@echo "===================="
	@echo "Health Check:"
	@echo "  curl http://localhost:8000/health"
	@echo ""
	@echo "Get All Customers:"
	@echo "  curl http://localhost:8000/api/v1/customers | jq"
	@echo ""
	@echo "Search Customers:"
	@echo "  curl 'http://localhost:8000/api/v1/customers/search?q=oil' | jq"
	@echo ""
	@echo "Get Customer by ID:"
	@echo "  curl http://localhost:8000/api/v1/customers/1 | jq"

## Check if database is ready for API
api-check-db:
	@echo "🔍 Checking database readiness for API..."
	@$(MAKE) db-health && echo "✅ Database accessible" || (echo "❌ Database not accessible" && exit 1)
	@echo "📊 Sample data check:"
	@$(MAKE) db-exec SQL="SELECT COUNT(*) as customers FROM store.customers;" 2>/dev/null || echo "❌ Cannot access customers table"

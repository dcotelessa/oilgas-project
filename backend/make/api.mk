# =============================================================================
# API Development Commands
# =============================================================================

.PHONY: api-start api-test api-test-quick api-dev api-logs api-examples api-curl-examples

## Start API server in development mode
api-start:
	@echo "🚀 Starting API server..."
	@echo "📋 Health: http://localhost:8000/health"
	@echo "🔌 API: http://localhost:8000/api/v1"
	@echo "Press Ctrl+C to stop"
	@go run cmd/server/main.go

## Test API integration with repository layer
api-test:
	@echo "🧪 Testing API integration..."
	@chmod +x scripts/test_api_integration.sh
	@scripts/test_api_integration.sh

## Quick API health check
api-test-quick:
	@echo "⚡ Quick API test..."
	@curl -s http://localhost:8000/health | jq -r '"Status: " + .status + " | Service: " + .service' 2>/dev/null || echo "❌ API not responding"
	@curl -s http://localhost:8000/api/v1/status | jq -r '"Message: " + .message' 2>/dev/null || echo "❌ API not responding"

## Start API in development mode with auto-reload
api-dev:
	@echo "🔄 Starting API with auto-reload..."
	@echo "💡 Install 'air' for auto-reload: go install github.com/cosmtrek/air@latest"
	@which air > /dev/null && air || $(MAKE) api-start

## Show API server logs (if running in docker)
api-logs:
	@docker logs -f oil-gas-inventory-api 2>/dev/null || echo "API not running in docker"

## Show API usage examples
api-examples:
	@echo "🔍 API Usage Examples"
	@echo "===================="
	@echo ""
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
	@echo ""
	@echo "Get All Inventory:"
	@echo "  curl http://localhost:8000/api/v1/inventory | jq"
	@echo ""
	@echo "Search Inventory:"
	@echo "  curl 'http://localhost:8000/api/v1/inventory/search?q=5' | jq"
	@echo ""
	@echo "Reference Data:"
	@echo "  curl http://localhost:8000/api/v1/grades | jq"
	@echo "  curl http://localhost:8000/api/v1/sizes | jq"

## Run live curl examples (requires API running)
api-curl-examples:
	@echo "🔗 Live API Examples"
	@echo "==================="
	@echo ""
	@echo "1. Health Check:"
	@curl -s http://localhost:8000/health | jq 2>/dev/null || echo "❌ API not responding"
	@echo ""
	@echo "2. Customer Count:"
	@curl -s http://localhost:8000/api/v1/customers | jq '.count' 2>/dev/null || echo "❌ API not responding"
	@echo ""
	@echo "3. Sample Customer:"
	@curl -s http://localhost:8000/api/v1/customers | jq '.customers[0] // "No customers found"' 2>/dev/null || echo "❌ API not responding"
	@echo ""
	@echo "4. Inventory Count:"
	@curl -s http://localhost:8000/api/v1/inventory | jq '.count' 2>/dev/null || echo "❌ API not responding"

## Check if database is ready for API
api-check-db:
	@echo "🔍 Checking database readiness for API..."
	@$(MAKE) db-health 2>/dev/null && echo "✅ Database accessible" || echo "❌ Database not accessible"
	@echo "Sample data check:"
	@$(MAKE) db-exec SQL="SELECT COUNT(*) as customers FROM store.customers;" 2>/dev/null || echo "❌ Cannot check customers table"
	@$(MAKE) db-exec SQL="SELECT COUNT(*) as inventory FROM store.inventory;" 2>/dev/null || echo "❌ Cannot check inventory table"

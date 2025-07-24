# =============================================================================
# Compatibility Layer - Missing Targets
# =============================================================================

.PHONY: dev-ensure-db dev-wait

## Ensure database is accessible (compatibility target)
dev-ensure-db:
	@echo "🔍 Checking database accessibility..."
	@$(MAKE) db-health 2>/dev/null || (echo "❌ Database not accessible - run: make docker-up && make dev-setup" && exit 1)

## Wait for services to be ready (compatibility target)  
dev-wait:
	@echo "⏳ Waiting for services..."
	@sleep 5
	@$(MAKE) db-health 2>/dev/null || echo "⚠️  Database not ready yet"

## Compatibility aliases
.PHONY: ensure-db wait-services

ensure-db: dev-ensure-db
wait-services: dev-wait

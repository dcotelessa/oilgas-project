# =============================================================================
# Compatibility Layer - Missing Targets and Dependencies
# Created by scripts/fix_makefile_conflicts.sh
# =============================================================================

.PHONY: dev-ensure-db dev-wait dev-db-reset

## Ensure database is accessible (compatibility target)
dev-ensure-db:
	@echo "🔍 Checking database accessibility..."
	@$(MAKE) db-health 2>/dev/null || (echo "❌ Database not accessible - run: make docker-up && make dev-setup" && exit 1)

## Wait for services to be ready (compatibility target)  
dev-wait:
	@echo "⏳ Waiting for services to be ready..."
	@sleep 3
	@echo "✅ Services should be ready"

## Reset development database (compatibility target)
dev-db-reset:
	@echo "🔄 Resetting development database..."
	@$(MAKE) test-db-reset 2>/dev/null || echo "⚠️  test-db-reset not available"

## Additional compatibility aliases
.PHONY: ensure-db wait-services db-reset

ensure-db: dev-ensure-db
wait-services: dev-wait
db-reset: dev-db-reset

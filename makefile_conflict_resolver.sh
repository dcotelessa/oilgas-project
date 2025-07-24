#!/bin/bash
# Resolve Makefile conflicts and missing targets

set -e

echo "🔧 Resolving Makefile Conflicts"
echo "==============================="

# Check for duplicate targets
echo ""
echo "1. Checking for duplicate targets..."
echo "======================================"

DUPLICATES=$(grep -h "^[a-zA-Z_-]*:" make/*.mk | sort | uniq -d | cut -d: -f1)

if [ -n "$DUPLICATES" ]; then
    echo "❌ Found duplicate targets:"
    for target in $DUPLICATES; do
        echo "   $target:"
        grep -n "^$target:" make/*.mk
    done
    echo ""
    echo "💡 Recommended resolution:"
    echo "   - Choose one file to own each target"
    echo "   - Rename duplicates with prefixes (test-db-setup vs data-db-setup)"
    echo "   - Use aliases in other files"
else
    echo "✅ No duplicate targets found"
fi

echo ""
echo "2. Checking for missing dependencies..."
echo "======================================="

# Common missing targets that might be referenced
COMMON_TARGETS=(
    "dev-ensure-db"
    "dev-setup"
    "dev-wait"
    "db-health"
    "db-exec"
    "test-unit"
    "docker-up"
    "docker-down"
)

MISSING=()
for target in "${COMMON_TARGETS[@]}"; do
    if ! grep -q "^$target:" make/*.mk 2>/dev/null; then
        MISSING+=("$target")
    fi
done

if [ ${#MISSING[@]} -gt 0 ]; then
    echo "❌ Missing targets referenced by new modules:"
    for target in "${MISSING[@]}"; do
        echo "   $target"
        # Show where it's referenced
        grep -n "$target" make/*.mk 2>/dev/null | head -2
    done
else
    echo "✅ All referenced targets found"
fi

echo ""
echo "3. Quick fixes for immediate testing..."
echo "======================================"

# Create minimal missing targets file
cat > make/compatibility.mk << 'EOF'
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
EOF

echo "✅ Created make/compatibility.mk with missing targets"

echo ""
echo "4. Testing fixes..."
echo "=================="

# Test if main Makefile loads
if make -n help > /dev/null 2>&1; then
    echo "✅ Main Makefile syntax OK"
else
    echo "❌ Main Makefile still has issues"
    echo "   Try: make -n help"
fi

# Test API target specifically
if make -n api-start > /dev/null 2>&1; then
    echo "✅ api-start target can be resolved"
else
    echo "❌ api-start still has dependency issues"
    echo "   Missing dependencies:"
    make -n api-start 2>&1 | grep "No rule to make target" | head -3
fi

echo ""
echo "5. Recommended resolution steps:"
echo "==============================="
echo ""
echo "Step 1: Add compatibility layer"
echo "  echo 'include make/compatibility.mk' >> Makefile"
echo ""
echo "Step 2: Test basic functionality"
echo "  make help"
echo "  make api-start"
echo ""
echo "Step 3: Resolve remaining conflicts"
echo "  - Review duplicate targets listed above"
echo "  - Choose which module owns each target"
echo "  - Rename or remove duplicates"
echo ""
echo "Step 4: Update your main Makefile include order"
echo "  # Suggested order:"
echo "  include make/compatibility.mk  # First - provides missing targets"
echo "  include make/database.mk"
echo "  include make/development.mk"
echo "  include make/testing.mk"
echo "  include make/docker.mk"
echo "  include make/api.mk           # After dependencies"
echo "  include make/data.mk          # Last - uses others"

echo ""
echo "🎯 Quick test commands:"
echo "======================"
echo "make help              # Test overall Makefile"
echo "make -n api-start      # Test API dependencies"
echo "make -n data-status    # Test data dependencies"
echo "make dev-ensure-db     # Test database check"

#!/bin/bash
# verify_migration.sh - Verify the new structure works

echo "🔍 Verifying migration..."

echo "1. Testing Go module..."
cd tools
if go mod tidy 2>/dev/null; then
    echo "✅ Go module OK"
else
    echo "❌ Go module issues"
    cd ..
    exit 1
fi
cd ..

echo "2. Testing Makefile commands..."
if make migration::setup >/dev/null 2>&1; then
    echo "✅ Migration setup OK"
else
    echo "❌ Migration setup failed"
fi

echo "3. Checking configuration files..."
for file in tools/config/*.json; do
    if python3 -m json.tool "$file" >/dev/null 2>&1; then
        echo "✅ $(basename "$file") valid"
    else
        echo "❌ $(basename "$file") invalid"
    fi
done

echo ""
echo "🎯 Migration verification complete!"
echo "If all checks passed, run: ./clean_old_structure.sh"

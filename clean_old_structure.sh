#!/bin/bash
# clean_old_structure.sh - Remove old structure after verification

echo "🗑️  Removing old mdb-conversion structure..."

if [ -d "tools/mdb-conversion" ]; then
    echo "Removing tools/mdb-conversion..."
    rm -rf tools/mdb-conversion
    echo "✅ Old structure removed"
else
    echo "❌ Old structure not found (already cleaned?)"
fi

# Clean up backup files
find tools -name "*.backup.*" -delete 2>/dev/null || true

echo "✅ Cleanup complete"

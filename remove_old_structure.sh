#!/bin/bash
# remove_old_structure.sh
# Removes the old tools/mdb-conversion structure after migration verification

echo "🗑️  Removing old tools structure..."

if [ -d "tools/mdb-conversion" ]; then
    echo "Removing tools/mdb-conversion..."
    rm -rf tools/mdb-conversion
    echo "✅ Old structure removed"
else
    echo "❌ Old structure not found"
fi

echo "✅ Cleanup complete"

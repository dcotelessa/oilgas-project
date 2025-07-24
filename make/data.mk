# =============================================================================
# Data Import Commands (MDB Processing)
# =============================================================================

.PHONY: data-check data-convert data-import data-status data-setup data-backup data-stats

## Check MDB files and conversion readiness
data-check:
	@echo "📂 Checking for MDB data files..."
	@test -d data/mdb && echo "✅ data/mdb directory exists" || echo "❌ Create: mkdir -p data/mdb"
	@test -d data/csv && echo "✅ data/csv directory exists" || echo "❌ Create: mkdir -p data/csv"
	@echo ""
	@echo "Expected structure:"
	@echo "  data/mdb/customers.mdb"
	@echo "  data/mdb/inventory.mdb"
	@echo "  data/csv/ (converted files)"

## Convert MDB files to CSV
data-convert:
	@echo "🔄 Converting MDB files to CSV..."
	@mkdir -p data/csv
	@echo "⚠️  MDB converter needs customization"
	@echo "   Available in Phase 3B implementation"

## Import converted data
data-import: dev-ensure-db
	@echo "📥 Importing MDB data..."
	@echo "⚠️  Import process needs customization"
	@echo "   Available in Phase 3B implementation"

## Show import status and data counts
data-status: dev-ensure-db
	@echo "📊 Data Import Status"
	@echo "====================="
	@$(MAKE) db-exec SQL="SELECT 'Customers' as table_name, count(*) as records FROM store.customers UNION ALL SELECT 'Inventory', count(*) FROM store.inventory UNION ALL SELECT 'Grades', count(*) FROM store.grade UNION ALL SELECT 'Sizes', count(*) FROM store.sizes;" 2>/dev/null || echo "❌ Cannot query tables"

## Create data directories
data-setup:
	@echo "📁 Creating data directories..."
	@mkdir -p data/mdb data/csv data/backup
	@echo "✅ Data directories created"

## Show data statistics
data-stats: dev-ensure-db
	@echo "📊 Database Statistics"
	@echo "======================"
	@$(MAKE) db-exec SQL="SELECT 'customers' as table_name, count(*) FROM store.customers UNION ALL SELECT 'inventory', count(*) FROM store.inventory;" 2>/dev/null || echo "❌ Cannot query tables"

## Legacy aliases for backward compatibility
.PHONY: import-check convert-mdb import-mdb-data import-status

import-check: data-check
convert-mdb: data-convert
import-mdb-data: data-import
import-status: data-status

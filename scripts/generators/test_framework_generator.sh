#!/bin/bash
# scripts/generators/test_framework_generator.sh
# Generates comprehensive test framework with tenant validation for Phase 3

set -e
echo "🧪 Generating comprehensive test framework..."

# Detect backend directory
BACKEND_DIR=""
if [ -d "backend" ] && [ -f "backend/go.mod" ]; then
    BACKEND_DIR="backend/"
elif [ -f "go.mod" ]; then
    BACKEND_DIR=""
else
    echo "❌ Error: Cannot find go.mod file"
    exit 1
fi

# Create directories
mkdir -p "${BACKEND_DIR}test/testdb"
mkdir -p "${BACKEND_DIR}test/auth"

# Generate test database utilities
cat > "${BACKEND_DIR}test/testdb/setup.go" << 'EOF'
package testdb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func SetupTestDB(t testing.TB) *pgxpool.Pool {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		testDatabaseURL = "postgres://postgres:password@localhost:5432/oil_gas_test?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(testDatabaseURL)
	require.NoError(t, err)

	config.MaxConns = 5
	config.MinConns = 1

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(t, err)

	err = pool.Ping(context.Background())
	require.NoError(t, err)

	cleanDatabase(t, pool)
	applyTestMigrations(t, pool)

	return pool
}

func cleanDatabase(t testing.TB, pool *pgxpool.Pool) {
	ctx := context.Background()
	
	queries := []string{
		"DELETE FROM auth.sessions",
		"DELETE FROM auth.users",
		"DELETE FROM store.received",
		"DELETE FROM store.inventory", 
		"DELETE FROM store.customers WHERE customer != 'Sample Oil Company'",
		"DELETE FROM store.tenants WHERE slug != 'default'",
	}
	
	for _, query := range queries {
		_, err := pool.Exec(ctx, query)
		if err != nil {
			t.Logf("Warning: cleanup query failed: %s - %v", query, err)
		}
	}
}

func applyTestMigrations(t testing.TB, pool *pgxpool.Pool) {
	ctx := context.Background()
	
	schema := `
		CREATE SCHEMA IF NOT EXISTS store;
		CREATE SCHEMA IF NOT EXISTS auth;
		
		CREATE TABLE IF NOT EXISTS store.tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) UNIQUE NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			database_type VARCHAR(20) DEFAULT 'shared',
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT NOW()
		);
		
		INSERT INTO store.tenants (name, slug, database_type, is_active) 
		VALUES ('Default Tenant', 'default', 'shared', true)
		ON CONFLICT (slug) DO NOTHING;
		
		CREATE TABLE IF NOT EXISTS store.customers (
			customer_id SERIAL PRIMARY KEY,
			tenant_id UUID NOT NULL REFERENCES store.tenants(id),
			customer VARCHAR(255) NOT NULL,
			billing_state VARCHAR(50),
			phone VARCHAR(50),
			email VARCHAR(255),
			deleted BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		);
		
		ALTER TABLE store.customers ENABLE ROW LEVEL SECURITY;
		
		DROP POLICY IF EXISTS test_tenant_isolation ON store.customers;
		CREATE POLICY test_tenant_isolation ON store.customers
		FOR ALL TO postgres
		USING (
			current_setting('app.user_role', true) IN ('admin', 'operator')
			OR tenant_id::text = current_setting('app.tenant_id', true)
		);
	`
	
	_, err := pool.Exec(ctx, schema)
	require.NoError(t, err)
}

func CreateTestTenant(t testing.TB, pool *pgxpool.Pool, name, slug string) string {
	ctx := context.Background()
	var tenantID string
	err := pool.QueryRow(ctx, `
		INSERT INTO store.tenants (name, slug, database_type, is_active)
		VALUES ($1, $2, 'shared', true)
		RETURNING id
	`, name, slug).Scan(&tenantID)
	require.NoError(t, err)
	return tenantID
}

func CreateTestCustomer(t testing.TB, pool *pgxpool.Pool, tenantID, name, state string) int {
	ctx := context.Background()
	var customerID int
	err := pool.QueryRow(ctx, `
		INSERT INTO store.customers (tenant_id, customer, billing_state, phone, email)
		VALUES ($1, $2, $3, '555-0123', $4)
		RETURNING customer_id
	`, tenantID, name, state, fmt.Sprintf("%s@test.com", strings.ToLower(strings.ReplaceAll(name, " ", "")))).Scan(&customerID)
	require.NoError(t, err)
	return customerID
}

func SetTenantContext(t testing.TB, pool *pgxpool.Pool, userRole, userCompany, tenantID string) {
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		SELECT 
			set_config('app.user_role', $1, true),
			set_config('app.user_company', $2, true),
			set_config('app.tenant_id', $3, true)
	`, userRole, userCompany, tenantID)
	require.NoError(t, err)
}

func ValidateTenantIsolation(t testing.TB, pool *pgxpool.Pool, tenantAID, tenantBID string) {
	ctx := context.Background()
	
	customerA := CreateTestCustomer(t, pool, tenantAID, "Tenant A Customer", "TX")
	customerB := CreateTestCustomer(t, pool, tenantBID, "Tenant B Customer", "CA")
	
	SetTenantContext(t, pool, "customer", "Company A", tenantAID)
	
	var visibleCustomers []string
	rows, err := pool.Query(ctx, "SELECT customer FROM store.customers ORDER BY customer")
	require.NoError(t, err)
	defer rows.Close()
	
	for rows.Next() {
		var customer string
		err := rows.Scan(&customer)
		require.NoError(t, err)
		visibleCustomers = append(visibleCustomers, customer)
	}
	
	found := false
	for _, customer := range visibleCustomers {
		if customer == "Tenant A Customer" {
			found = true
		}
		if customer == "Tenant B Customer" {
			t.Errorf("Tenant isolation failed: can see Tenant B's customer from Tenant A context")
		}
	}
	require.True(t, found, "Should be able to see own tenant's customers")
	
	t.Logf("✅ Tenant isolation validated - Tenant A sees %d customers: %v", len(visibleCustomers), visibleCustomers)
}
EOF

echo "✅ Comprehensive test framework generated"
echo "   - Test database setup utilities"
echo "   - Tenant isolation validation"
echo "   - Helper functions for test data"
echo "   - RLS testing capabilities"

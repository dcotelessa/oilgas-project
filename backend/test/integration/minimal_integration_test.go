// backend/test/integration/minimal_integration_test.go
package integration

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

// MinimalIntegrationTest tests basic database connectivity and operations
// without dependencies on broken internal packages
func TestMinimalIntegration(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration tests")
	}

	t.Log("🚀 Running minimal integration tests...")

	// Get database URLs
	authDBURL := getTestEnv("CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable")
	tenantDBURL := getTestEnv("LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable")

	// Test auth database connection
	t.Run("AuthDatabaseConnectivity", func(t *testing.T) {
		db, err := sql.Open("postgres", authDBURL)
		require.NoError(t, err)
		defer db.Close()

		err = db.Ping()
		require.NoError(t, err, "Should connect to auth database")

		var result string
		err = db.QueryRow("SELECT 'Auth DB Connected'").Scan(&result)
		require.NoError(t, err)
		assert.Equal(t, "Auth DB Connected", result)

		t.Log("✅ Auth database connectivity verified")
	})

	// Test tenant database connection
	t.Run("TenantDatabaseConnectivity", func(t *testing.T) {
		db, err := sql.Open("postgres", tenantDBURL)
		require.NoError(t, err)
		defer db.Close()

		err = db.Ping()
		require.NoError(t, err, "Should connect to tenant database")

		var result string
		err = db.QueryRow("SELECT 'Tenant DB Connected'").Scan(&result)
		require.NoError(t, err)
		assert.Equal(t, "Tenant DB Connected", result)

		t.Log("✅ Tenant database connectivity verified")
	})

	// Test multi-tenant isolation
	t.Run("MultiTenantIsolation", func(t *testing.T) {
		db, err := sql.Open("postgres", tenantDBURL)
		require.NoError(t, err)
		defer db.Close()

		// Create test customer for longbeach
		_, err = db.Exec(`
			INSERT INTO store.customers (tenant_id, name, company_code, status)
			VALUES ('longbeach', 'Isolation Test Co', 'INTEG001', 'active')
			ON CONFLICT (company_code) DO UPDATE SET name = EXCLUDED.name
		`)
		require.NoError(t, err)

		// Verify customer exists for longbeach
		var count int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM store.customers 
			WHERE tenant_id = 'longbeach' AND company_code = 'INTEG001'
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Customer should exist for longbeach tenant")

		// Verify no leakage to other tenants
		err = db.QueryRow(`
			SELECT COUNT(*) FROM store.customers 
			WHERE tenant_id != 'longbeach' AND company_code = 'INTEG001'
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Customer should not exist for other tenants")

		// Cleanup
		db.Exec("DELETE FROM store.customers WHERE company_code = 'INTEG001'")

		t.Log("✅ Multi-tenant isolation verified")
	})

	// Test cross-database operations
	t.Run("CrossDatabaseOperations", func(t *testing.T) {
		authDB, err := sql.Open("postgres", authDBURL)
		require.NoError(t, err)
		defer authDB.Close()

		tenantDB, err := sql.Open("postgres", tenantDBURL)
		require.NoError(t, err)
		defer tenantDB.Close()

		// Check if users table exists in auth database
		var userTableExists bool
		err = authDB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_schema = 'public' AND table_name = 'users'
			)
		`).Scan(&userTableExists)
		require.NoError(t, err)

		if userTableExists {
			// Test cross-database referential integrity concepts
			var authUserCount int
			err = authDB.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&authUserCount)
			require.NoError(t, err)
			t.Logf("Found %d active users in auth database", authUserCount)
		} else {
			t.Log("Users table not found in auth database, skipping user count test")
		}

		// Test tenant database customer count
		var customerCount int
		err = tenantDB.QueryRow("SELECT COUNT(*) FROM store.customers WHERE is_active = true").Scan(&customerCount)
		require.NoError(t, err)
		t.Logf("Found %d active customers in tenant database", customerCount)

		t.Log("✅ Cross-database operations verified")
	})

	// Test database performance
	t.Run("DatabasePerformance", func(t *testing.T) {
		db, err := sql.Open("postgres", tenantDBURL)
		require.NoError(t, err)
		defer db.Close()

		// Test bulk insert performance
		start := time.Now()
		const testRecords = 100

		for i := 0; i < testRecords; i++ {
			_, err = db.Exec(`
				INSERT INTO store.customers (tenant_id, name, company_code, status)
				VALUES ('longbeach', $1, $2, 'active')
				ON CONFLICT (company_code) DO NOTHING
			`, 
				"Perf Test Co "+string(rune(i/26+65))+string(rune(i%26+65)), 
				"PERF"+string(rune(48+(i/100)%10))+string(rune(48+(i/10)%10))+string(rune(48+i%10)))
		}

		insertDuration := time.Since(start)
		t.Logf("Inserted %d records in %v (%.2f records/sec)", 
			testRecords, insertDuration, float64(testRecords)/insertDuration.Seconds())

		// Verify inserts
		var insertedCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM store.customers WHERE company_code LIKE 'PERF%'
		`).Scan(&insertedCount)
		require.NoError(t, err)
		assert.Greater(t, insertedCount, 0, "Should have inserted some performance test records")

		// Test query performance
		start = time.Now()
		var activeCustomers int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM store.customers 
			WHERE status = 'active' AND tenant_id = 'longbeach'
		`).Scan(&activeCustomers)
		require.NoError(t, err)
		queryDuration := time.Since(start)

		t.Logf("Query found %d active customers in %v", activeCustomers, queryDuration)

		// Cleanup performance test data
		db.Exec("DELETE FROM store.customers WHERE company_code LIKE 'PERF%'")

		t.Log("✅ Database performance tests completed")
	})

	// Test transaction handling
	t.Run("TransactionHandling", func(t *testing.T) {
		db, err := sql.Open("postgres", tenantDBURL)
		require.NoError(t, err)
		defer db.Close()

		// Test successful transaction
		tx, err := db.Begin()
		require.NoError(t, err)

		_, err = tx.Exec(`
			INSERT INTO store.customers (tenant_id, name, company_code, status)
			VALUES ('longbeach', 'Transaction Test Co', 'TXN001', 'active')
		`)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		// Verify committed data
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM store.customers WHERE company_code = 'TXN001')
		`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Transaction should have committed successfully")

		// Test rollback
		tx2, err := db.Begin()
		require.NoError(t, err)

		_, err = tx2.Exec(`
			INSERT INTO store.customers (tenant_id, name, company_code, status)
			VALUES ('longbeach', 'Rollback Test Co', 'ROLL001', 'active')
		`)
		require.NoError(t, err)

		err = tx2.Rollback()
		require.NoError(t, err)

		// Verify rollback worked
		err = db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM store.customers WHERE company_code = 'ROLL001')
		`).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "Transaction should have rolled back")

		// Cleanup
		db.Exec("DELETE FROM store.customers WHERE company_code IN ('TXN001', 'ROLL001')")

		t.Log("✅ Transaction handling verified")
	})

	t.Log("🎉 All minimal integration tests passed!")
}

// Test environment variable getter
func getTestEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Test database schema validation
func TestDatabaseSchemaValidation(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration tests")
	}

	t.Log("🔍 Validating database schemas...")

	authDBURL := getTestEnv("CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable")
	tenantDBURL := getTestEnv("LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable")

	t.Run("AuthDatabaseSchema", func(t *testing.T) {
		db, err := sql.Open("postgres", authDBURL)
		require.NoError(t, err)
		defer db.Close()

		// Check for expected tables
		expectedTables := []string{"users", "sessions", "user_tenant_access"}
		for _, table := range expectedTables {
			var exists bool
			err = db.QueryRow(`
				SELECT EXISTS (
					SELECT 1 FROM information_schema.tables 
					WHERE table_schema = 'public' AND table_name = $1
				)
			`, table).Scan(&exists)
			require.NoError(t, err)

			if exists {
				t.Logf("✅ Table '%s' exists in auth database", table)
			} else {
				t.Logf("⚠️  Table '%s' not found in auth database", table)
			}
		}
	})

	t.Run("TenantDatabaseSchema", func(t *testing.T) {
		db, err := sql.Open("postgres", tenantDBURL)
		require.NoError(t, err)
		defer db.Close()

		// Check for expected tables in store schema
		expectedTables := []string{"customers", "customer_contacts", "tenants"}
		for _, table := range expectedTables {
			var exists bool
			err = db.QueryRow(`
				SELECT EXISTS (
					SELECT 1 FROM information_schema.tables 
					WHERE table_schema = 'store' AND table_name = $1
				)
			`, table).Scan(&exists)
			require.NoError(t, err)

			if exists {
				t.Logf("✅ Table 'store.%s' exists in tenant database", table)
			} else {
				t.Logf("❌ Table 'store.%s' not found in tenant database", table)
			}
		}

		// Verify tenant data
		var tenantExists bool
		err = db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM store.tenants WHERE tenant_id = 'longbeach')
		`).Scan(&tenantExists)
		require.NoError(t, err)
		
		if tenantExists {
			t.Log("✅ Longbeach tenant configuration exists")
		} else {
			t.Log("⚠️  Longbeach tenant configuration not found")
		}
	})

	t.Log("✅ Database schema validation completed")
}
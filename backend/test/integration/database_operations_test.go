// backend/test/integration/database_operations_test.go
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"

	"oilgas-backend/internal/shared/database"
)

type DatabaseOperationsTestSuite struct {
	suite.Suite
	dbManager   *database.DatabaseManager
	ctx         context.Context
	authDB      *sql.DB
	tenantDB    *sql.DB
}

func (suite *DatabaseOperationsTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Initialize database manager
	dbConfig := &database.Config{
		CentralDBURL: getEnvOrDefault("CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable"),
		TenantDBs: map[string]string{
			"longbeach": getEnvOrDefault("LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable"),
		},
		MaxOpenConns: 10,
		MaxIdleConns: 2,
		MaxLifetime:  time.Hour,
	}

	var err error
	suite.dbManager, err = database.NewDatabaseManager(dbConfig)
	suite.Require().NoError(err)

	// Get direct database connections
	suite.authDB, err = sql.Open("postgres", dbConfig.CentralDBURL)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.authDB.Ping())

	suite.tenantDB, err = sql.Open("postgres", dbConfig.TenantDBs["longbeach"])
	suite.Require().NoError(err)
	suite.Require().NoError(suite.tenantDB.Ping())

	suite.setupTestSchemas()
}

func (suite *DatabaseOperationsTestSuite) setupTestSchemas() {
	// Create test schemas
	suite.authDB.Exec("CREATE SCHEMA IF NOT EXISTS test_ops")
	suite.tenantDB.Exec("CREATE SCHEMA IF NOT EXISTS test_ops")

	// Create test tables for database operations testing
	suite.createOperationsTestTables()
}

func (suite *DatabaseOperationsTestSuite) createOperationsTestTables() {
	// Auth database test tables
	_, err := suite.authDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_ops.performance_test (
			id SERIAL PRIMARY KEY,
			data VARCHAR(255),
			number_field INTEGER,
			timestamp_field TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			json_field JSONB
		)
	`)
	suite.Require().NoError(err)

	// Tenant database test tables
	_, err = suite.tenantDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_ops.bulk_operations (
			id SERIAL PRIMARY KEY,
			tenant_id VARCHAR(100),
			name VARCHAR(255),
			status VARCHAR(50),
			data JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	suite.Require().NoError(err)

	// Create indexes for performance testing
	suite.authDB.Exec("CREATE INDEX IF NOT EXISTS idx_perf_test_data ON test_ops.performance_test(data)")
	suite.authDB.Exec("CREATE INDEX IF NOT EXISTS idx_perf_test_number ON test_ops.performance_test(number_field)")
	suite.tenantDB.Exec("CREATE INDEX IF NOT EXISTS idx_bulk_ops_tenant ON test_ops.bulk_operations(tenant_id)")
	suite.tenantDB.Exec("CREATE INDEX IF NOT EXISTS idx_bulk_ops_status ON test_ops.bulk_operations(status)")
}

func (suite *DatabaseOperationsTestSuite) TearDownSuite() {
	suite.cleanupTestData()
	if suite.authDB != nil {
		suite.authDB.Close()
	}
	if suite.tenantDB != nil {
		suite.tenantDB.Close()
	}
	if suite.dbManager != nil {
		suite.dbManager.Close()
	}
}

func (suite *DatabaseOperationsTestSuite) cleanupTestData() {
	if suite.authDB != nil {
		suite.authDB.Exec("DROP SCHEMA IF EXISTS test_ops CASCADE")
	}
	if suite.tenantDB != nil {
		suite.tenantDB.Exec("DROP SCHEMA IF EXISTS test_ops CASCADE")
	}
}

func (suite *DatabaseOperationsTestSuite) SetupTest() {
	// Clean test data before each test
	suite.authDB.Exec("TRUNCATE TABLE test_ops.performance_test RESTART IDENTITY")
	suite.tenantDB.Exec("TRUNCATE TABLE test_ops.bulk_operations RESTART IDENTITY")
}

// Test bulk insert operations
func (suite *DatabaseOperationsTestSuite) TestBulkInsertOperations() {
	suite.T().Log("Testing bulk insert operations...")

	const batchSize = 1000
	start := time.Now()

	// Prepare bulk insert data
	values := make([]string, 0, batchSize)
	args := make([]interface{}, 0, batchSize*4)
	
	for i := 0; i < batchSize; i++ {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		args = append(args, 
			"longbeach",                           // tenant_id
			fmt.Sprintf("Test Record %d", i+1),    // name
			"active",                              // status
			fmt.Sprintf(`{"index": %d}`, i),       // data (JSONB)
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO test_ops.bulk_operations (tenant_id, name, status, data)
		VALUES %s
	`, strings.Join(values, ","))

	_, err := suite.tenantDB.Exec(query, args...)
	suite.Require().NoError(err)

	insertDuration := time.Since(start)
	suite.T().Logf("Bulk insert of %d records took: %v", batchSize, insertDuration)

	// Verify all records were inserted
	var count int
	err = suite.tenantDB.QueryRow("SELECT COUNT(*) FROM test_ops.bulk_operations").Scan(&count)
	suite.Require().NoError(err)
	suite.Assert().Equal(batchSize, count)

	// Verify data integrity
	var sampleData struct {
		TenantID string
		Name     string
		Status   string
		Data     string
	}
	err = suite.tenantDB.QueryRow(`
		SELECT tenant_id, name, status, data::text
		FROM test_ops.bulk_operations
		WHERE id = 100
	`).Scan(&sampleData.TenantID, &sampleData.Name, &sampleData.Status, &sampleData.Data)
	suite.Require().NoError(err)
	suite.Assert().Equal("longbeach", sampleData.TenantID)
	suite.Assert().Equal("Test Record 100", sampleData.Name)
	suite.Assert().Contains(sampleData.Data, `"index": 99`) // 0-indexed
}

// Test bulk update operations
func (suite *DatabaseOperationsTestSuite) TestBulkUpdateOperations() {
	suite.T().Log("Testing bulk update operations...")

	// First, insert test data
	const numRecords = 500
	for i := 0; i < numRecords; i++ {
		_, err := suite.tenantDB.Exec(`
			INSERT INTO test_ops.bulk_operations (tenant_id, name, status, data)
			VALUES ('longbeach', $1, 'pending', '{"created": true}')
		`, fmt.Sprintf("Update Test %d", i+1))
		suite.Require().NoError(err)
	}

	start := time.Now()

	// Perform bulk update
	result, err := suite.tenantDB.Exec(`
		UPDATE test_ops.bulk_operations
		SET status = 'completed', data = data || '{"updated": true}'
		WHERE status = 'pending' AND tenant_id = 'longbeach'
	`)
	suite.Require().NoError(err)

	updateDuration := time.Since(start)
	suite.T().Logf("Bulk update took: %v", updateDuration)

	// Verify update results
	rowsAffected, err := result.RowsAffected()
	suite.Require().NoError(err)
	suite.Assert().Equal(int64(numRecords), rowsAffected)

	// Verify data integrity after update
	var count int
	err = suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations
		WHERE status = 'completed' AND data::text LIKE '%updated%'
	`).Scan(&count)
	suite.Require().NoError(err)
	suite.Assert().Equal(numRecords, count)
}

// Test transaction performance and rollback
func (suite *DatabaseOperationsTestSuite) TestTransactionPerformance() {
	suite.T().Log("Testing transaction performance and rollback...")

	// Test successful transaction
	start := time.Now()
	tx, err := suite.tenantDB.Begin()
	suite.Require().NoError(err)

	const numInserts = 100
	for i := 0; i < numInserts; i++ {
		_, err := tx.Exec(`
			INSERT INTO test_ops.bulk_operations (tenant_id, name, status)
			VALUES ('longbeach', $1, 'transaction_test')
		`, fmt.Sprintf("TX Test %d", i+1))
		suite.Require().NoError(err)
	}

	err = tx.Commit()
	suite.Require().NoError(err)
	commitDuration := time.Since(start)

	// Verify successful transaction
	var successCount int
	err = suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations WHERE status = 'transaction_test'
	`).Scan(&successCount)
	suite.Require().NoError(err)
	suite.Assert().Equal(numInserts, successCount)

	suite.T().Logf("Transaction with %d inserts and commit took: %v", numInserts, commitDuration)

	// Test transaction rollback
	start = time.Now()
	tx2, err := suite.tenantDB.Begin()
	suite.Require().NoError(err)

	// Insert some data
	for i := 0; i < numInserts; i++ {
		_, err := tx2.Exec(`
			INSERT INTO test_ops.bulk_operations (tenant_id, name, status)
			VALUES ('longbeach', $1, 'rollback_test')
		`, fmt.Sprintf("Rollback Test %d", i+1))
		suite.Require().NoError(err)
	}

	// Rollback the transaction
	err = tx2.Rollback()
	suite.Require().NoError(err)
	rollbackDuration := time.Since(start)

	// Verify rollback worked
	var rollbackCount int
	err = suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations WHERE status = 'rollback_test'
	`).Scan(&rollbackCount)
	suite.Require().NoError(err)
	suite.Assert().Equal(0, rollbackCount, "No records should exist after rollback")

	suite.T().Logf("Transaction with %d inserts and rollback took: %v", numInserts, rollbackDuration)
}

// Test connection pool behavior under load
func (suite *DatabaseOperationsTestSuite) TestConnectionPoolBehavior() {
	suite.T().Log("Testing connection pool behavior under load...")

	const numWorkers = 20
	const operationsPerWorker = 50
	
	// Channel to collect results
	results := make(chan error, numWorkers*operationsPerWorker)
	
	start := time.Now()

	// Spawn workers to stress test connection pool
	for worker := 0; worker < numWorkers; worker++ {
		go func(workerID int) {
			for op := 0; op < operationsPerWorker; op++ {
				// Get tenant database through manager (tests connection pooling)
				db, err := suite.dbManager.GetTenantDB("longbeach")
				if err != nil {
					results <- err
					continue
				}

				// Perform database operation
				_, err = db.Exec(`
					INSERT INTO test_ops.bulk_operations (tenant_id, name, status)
					VALUES ('longbeach', $1, 'pool_test')
				`, fmt.Sprintf("Worker %d Op %d", workerID, op))
				
				results <- err
			}
		}(worker)
	}

	// Collect all results
	var errors []error
	for i := 0; i < numWorkers*operationsPerWorker; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	loadTestDuration := time.Since(start)
	suite.T().Logf("Connection pool test with %d workers × %d ops took: %v", 
		numWorkers, operationsPerWorker, loadTestDuration)

	// Verify no errors occurred
	suite.Assert().Empty(errors, "No errors should occur during connection pool stress test")

	// Verify all operations completed
	var totalInserted int
	err := suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations WHERE status = 'pool_test'
	`).Scan(&totalInserted)
	suite.Require().NoError(err)
	suite.Assert().Equal(numWorkers*operationsPerWorker, totalInserted)
}

// Test JSON operations and indexing
func (suite *DatabaseOperationsTestSuite) TestJSONOperations() {
	suite.T().Log("Testing JSON operations and performance...")

	// Insert records with varying JSON structures
	testData := []struct {
		name string
		json string
	}{
		{"Customer Data", `{"type": "customer", "priority": "high", "tags": ["oil", "gas"]}`},
		{"Work Order", `{"type": "workorder", "priority": "medium", "equipment": {"id": 123, "type": "pump"}}`},
		{"Inventory", `{"type": "inventory", "priority": "low", "location": {"yard": "A", "rack": "R01"}}`},
		{"Complex Data", `{"type": "complex", "priority": "high", "nested": {"level1": {"level2": {"value": 42}}}}`},
	}

	for i, data := range testData {
		_, err := suite.tenantDB.Exec(`
			INSERT INTO test_ops.bulk_operations (tenant_id, name, status, data)
			VALUES ('longbeach', $1, 'json_test', $2::jsonb)
		`, data.name, data.json)
		suite.Require().NoError(err, "Failed to insert record %d", i)
	}

	// Test JSON path queries
	start := time.Now()
	var highPriorityCount int
	err := suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations
		WHERE data->>'priority' = 'high' AND status = 'json_test'
	`).Scan(&highPriorityCount)
	suite.Require().NoError(err)
	suite.Assert().Equal(2, highPriorityCount)

	jsonQueryDuration := time.Since(start)
	suite.T().Logf("JSON path query took: %v", jsonQueryDuration)

	// Test complex nested JSON query
	start = time.Now()
	var complexCount int
	err = suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations
		WHERE data->'nested'->'level1'->'level2'->>'value' = '42'
	`).Scan(&complexCount)
	suite.Require().NoError(err)
	suite.Assert().Equal(1, complexCount)

	nestedQueryDuration := time.Since(start)
	suite.T().Logf("Nested JSON query took: %v", nestedQueryDuration)

	// Test JSON aggregation
	start = time.Now()
	rows, err := suite.tenantDB.Query(`
		SELECT data->>'type' as type, COUNT(*) as count
		FROM test_ops.bulk_operations
		WHERE status = 'json_test'
		GROUP BY data->>'type'
		ORDER BY count DESC
	`)
	suite.Require().NoError(err)
	defer rows.Close()

	typeCount := make(map[string]int)
	for rows.Next() {
		var dataType string
		var count int
		err := rows.Scan(&dataType, &count)
		suite.Require().NoError(err)
		typeCount[dataType] = count
	}

	aggregationDuration := time.Since(start)
	suite.T().Logf("JSON aggregation query took: %v", aggregationDuration)

	// Verify aggregation results
	suite.Assert().Equal(1, typeCount["customer"])
	suite.Assert().Equal(1, typeCount["workorder"])
	suite.Assert().Equal(1, typeCount["inventory"])
	suite.Assert().Equal(1, typeCount["complex"])
}

// Test cross-database query performance
func (suite *DatabaseOperationsTestSuite) TestCrossDatabaseQueryPerformance() {
	suite.T().Log("Testing cross-database query performance...")

	// Insert test data in both databases
	const numRecords = 100

	// Auth database data
	start := time.Now()
	for i := 0; i < numRecords; i++ {
		_, err := suite.authDB.Exec(`
			INSERT INTO test_ops.performance_test (data, number_field, json_field)
			VALUES ($1, $2, $3)
		`, 
			fmt.Sprintf("Auth Record %d", i+1),
			i*10,
			fmt.Sprintf(`{"auth_id": %d}`, i+1),
		)
		suite.Require().NoError(err)
	}
	authInsertDuration := time.Since(start)

	// Tenant database data
	start = time.Now()
	for i := 0; i < numRecords; i++ {
		_, err := suite.tenantDB.Exec(`
			INSERT INTO test_ops.bulk_operations (tenant_id, name, status, data)
			VALUES ('longbeach', $1, 'cross_db_test', $2)
		`,
			fmt.Sprintf("Tenant Record %d", i+1),
			fmt.Sprintf(`{"tenant_id": %d}`, i+1),
		)
		suite.Require().NoError(err)
	}
	tenantInsertDuration := time.Since(start)

	suite.T().Logf("Auth DB insert (%d records): %v", numRecords, authInsertDuration)
	suite.T().Logf("Tenant DB insert (%d records): %v", numRecords, tenantInsertDuration)

	// Test individual database query performance
	start = time.Now()
	var authCount int
	err := suite.authDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.performance_test WHERE number_field >= 500
	`).Scan(&authCount)
	suite.Require().NoError(err)
	authQueryDuration := time.Since(start)

	start = time.Now()
	var tenantCount int
	err = suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations WHERE status = 'cross_db_test'
	`).Scan(&tenantCount)
	suite.Require().NoError(err)
	tenantQueryDuration := time.Since(start)

	suite.T().Logf("Auth DB query performance: %v (found %d records)", authQueryDuration, authCount)
	suite.T().Logf("Tenant DB query performance: %v (found %d records)", tenantQueryDuration, tenantCount)

	suite.Assert().Equal(numRecords, tenantCount)
	suite.Assert().Greater(authCount, 0)
}

// Test database backup and restore simulation
func (suite *DatabaseOperationsTestSuite) TestBackupRestoreSimulation() {
	suite.T().Log("Testing backup/restore simulation...")

	// Create some test data
	originalData := []string{"Record 1", "Record 2", "Record 3"}
	for _, name := range originalData {
		_, err := suite.tenantDB.Exec(`
			INSERT INTO test_ops.bulk_operations (tenant_id, name, status)
			VALUES ('longbeach', $1, 'backup_test')
		`, name)
		suite.Require().NoError(err)
	}

	// Simulate backup by exporting data
	rows, err := suite.tenantDB.Query(`
		SELECT name FROM test_ops.bulk_operations WHERE status = 'backup_test' ORDER BY id
	`)
	suite.Require().NoError(err)

	var backupData []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		suite.Require().NoError(err)
		backupData = append(backupData, name)
	}
	rows.Close()

	// Simulate data loss
	_, err = suite.tenantDB.Exec(`DELETE FROM test_ops.bulk_operations WHERE status = 'backup_test'`)
	suite.Require().NoError(err)

	// Verify data is gone
	var deletedCount int
	err = suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations WHERE status = 'backup_test'
	`).Scan(&deletedCount)
	suite.Require().NoError(err)
	suite.Assert().Equal(0, deletedCount, "All backup_test records should be deleted")

	// Simulate restore
	for _, name := range backupData {
		_, err := suite.tenantDB.Exec(`
			INSERT INTO test_ops.bulk_operations (tenant_id, name, status)
			VALUES ('longbeach', $1, 'backup_test')
		`, name)
		suite.Require().NoError(err)
	}

	// Verify restore
	var restoredCount int
	err = suite.tenantDB.QueryRow(`
		SELECT COUNT(*) FROM test_ops.bulk_operations WHERE status = 'backup_test'
	`).Scan(&restoredCount)
	suite.Require().NoError(err)
	suite.Assert().Equal(len(originalData), restoredCount, "All records should be restored")

	// Verify data integrity
	suite.Assert().Equal(originalData, backupData, "Backup data should match original data")
}

func TestDatabaseOperationsTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseOperationsTestSuite))
}
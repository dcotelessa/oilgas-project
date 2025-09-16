// backend/test/integration/test_fixtures.go
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"oilgas-backend/internal/shared/database"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	AuthDBURL     string
	LongbeachURL  string
	TestTimeout   time.Duration
	CleanupData   bool
	VerboseOutput bool
}

// NewTestConfig creates a new test configuration from environment variables
func NewTestConfig() *TestConfig {
	return &TestConfig{
		AuthDBURL:     getEnvOrDefault("TEST_CENTRAL_AUTH_DB_URL", getEnvOrDefault("CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable")),
		LongbeachURL:  getEnvOrDefault("TEST_LONGBEACH_DB_URL", getEnvOrDefault("LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable")),
		TestTimeout:   time.Duration(getEnvIntOrDefault("TEST_TIMEOUT_SECONDS", 300)) * time.Second,
		CleanupData:   getEnvBoolOrDefault("TEST_CLEANUP_DATA", true),
		VerboseOutput: getEnvBoolOrDefault("TEST_VERBOSE", false),
	}
}

// DatabaseFixture manages test database setup and cleanup
type DatabaseFixture struct {
	Config    *TestConfig
	DBManager *database.DatabaseManager
	AuthDB    *sql.DB
	TenantDB  *sql.DB
	ctx       context.Context
}

// NewDatabaseFixture creates a new database fixture for testing
func NewDatabaseFixture() (*DatabaseFixture, error) {
	config := NewTestConfig()
	
	dbConfig := &database.Config{
		CentralDBURL: config.AuthDBURL,
		TenantDBs: map[string]string{
			"longbeach": config.LongbeachURL,
		},
		MaxOpenConns: 10,
		MaxIdleConns: 2,
		MaxLifetime:  time.Hour,
	}

	dbManager, err := database.NewDatabaseManager(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database manager: %w", err)
	}

	authDB, err := sql.Open("postgres", config.AuthDBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth database: %w", err)
	}

	tenantDB, err := sql.Open("postgres", config.LongbeachURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	fixture := &DatabaseFixture{
		Config:    config,
		DBManager: dbManager,
		AuthDB:    authDB,
		TenantDB:  tenantDB,
		ctx:       context.Background(),
	}

	// Verify connections
	if err := fixture.VerifyConnections(); err != nil {
		fixture.Close()
		return nil, fmt.Errorf("database connection verification failed: %w", err)
	}

	return fixture, nil
}

// VerifyConnections checks that all database connections are working
func (f *DatabaseFixture) VerifyConnections() error {
	if err := f.AuthDB.Ping(); err != nil {
		return fmt.Errorf("auth database ping failed: %w", err)
	}

	if err := f.TenantDB.Ping(); err != nil {
		return fmt.Errorf("tenant database ping failed: %w", err)
	}

	// Test database manager connections
	tenantDB, err := f.DBManager.GetTenantDB("longbeach")
	if err != nil {
		return fmt.Errorf("failed to get tenant database through manager: %w", err)
	}

	if err := tenantDB.Ping(); err != nil {
		return fmt.Errorf("tenant database manager ping failed: %w", err)
	}

	if f.Config.VerboseOutput {
		log.Println("✅ All database connections verified")
	}

	return nil
}

// SetupTestSchemas creates test schemas and tables
func (f *DatabaseFixture) SetupTestSchemas() error {
	// Create test schemas
	schemas := []struct {
		db   *sql.DB
		name string
	}{
		{f.AuthDB, "test_integration"},
		{f.TenantDB, "test_integration"},
	}

	for _, schema := range schemas {
		_, err := schema.db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema.name))
		if err != nil {
			return fmt.Errorf("failed to create schema %s: %w", schema.name, err)
		}
	}

	// Create test tables
	if err := f.createAuthTestTables(); err != nil {
		return fmt.Errorf("failed to create auth test tables: %w", err)
	}

	if err := f.createTenantTestTables(); err != nil {
		return fmt.Errorf("failed to create tenant test tables: %w", err)
	}

	if f.Config.VerboseOutput {
		log.Println("✅ Test schemas and tables created")
	}

	return nil
}

func (f *DatabaseFixture) createAuthTestTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS test_integration.users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) UNIQUE,
			email VARCHAR(255) UNIQUE NOT NULL,
			full_name VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'OPERATOR',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS test_integration.user_tenant_access (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES test_integration.users(id) ON DELETE CASCADE,
			tenant_id VARCHAR(100) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'USER',
			permissions JSONB DEFAULT '[]'::jsonb,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, tenant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS test_integration.sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id INTEGER REFERENCES test_integration.users(id) ON DELETE CASCADE,
			tenant_id VARCHAR(100) NOT NULL,
			token_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			last_accessed TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
	}

	for _, table := range tables {
		if _, err := f.AuthDB.Exec(table); err != nil {
			return fmt.Errorf("failed to create auth table: %w", err)
		}
	}

	return nil
}

func (f *DatabaseFixture) createTenantTestTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS test_integration.customers (
			id SERIAL PRIMARY KEY,
			tenant_id VARCHAR(100) NOT NULL DEFAULT 'longbeach',
			name VARCHAR(255) NOT NULL,
			company_code VARCHAR(50) UNIQUE,
			status VARCHAR(20) DEFAULT 'active',
			tax_id VARCHAR(50),
			payment_terms VARCHAR(50) DEFAULT 'NET30',
			billing_street VARCHAR(255),
			billing_city VARCHAR(100),
			billing_state VARCHAR(10),
			billing_zip_code VARCHAR(20),
			billing_country VARCHAR(10) DEFAULT 'US',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS test_integration.customer_contacts (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER REFERENCES test_integration.customers(id) ON DELETE CASCADE,
			auth_user_id INTEGER NOT NULL,
			contact_type VARCHAR(50) DEFAULT 'PRIMARY',
			is_primary BOOLEAN DEFAULT false,
			is_active BOOLEAN DEFAULT true,
			full_name VARCHAR(255),
			email VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(customer_id, auth_user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS test_integration.test_data (
			id SERIAL PRIMARY KEY,
			tenant_id VARCHAR(100),
			name VARCHAR(255),
			data JSONB,
			number_field INTEGER,
			timestamp_field TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			status VARCHAR(50) DEFAULT 'active'
		)`,
	}

	for _, table := range tables {
		if _, err := f.TenantDB.Exec(table); err != nil {
			return fmt.Errorf("failed to create tenant table: %w", err)
		}
	}

	return nil
}

// CreateTestUser creates a test user in the auth database
func (f *DatabaseFixture) CreateTestUser(username, email, fullName, role string) (int, error) {
	var userID int
	err := f.AuthDB.QueryRow(`
		INSERT INTO test_integration.users (username, email, full_name, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, username, email, fullName, "hashed_password", role).Scan(&userID)
	
	if err != nil {
		return 0, fmt.Errorf("failed to create test user: %w", err)
	}

	if f.Config.VerboseOutput {
		log.Printf("Created test user: %s (ID: %d)", username, userID)
	}

	return userID, nil
}

// GrantTenantAccess grants a user access to a specific tenant
func (f *DatabaseFixture) GrantTenantAccess(userID int, tenantID, role string, permissions []string) error {
	permissionsJSON := "[]"
	if len(permissions) > 0 {
		permissionsJSON = fmt.Sprintf(`["%s"]`, fmt.Sprintf(`", "%s`, permissions))
	}

	_, err := f.AuthDB.Exec(`
		INSERT INTO test_integration.user_tenant_access (user_id, tenant_id, role, permissions)
		VALUES ($1, $2, $3, $4::jsonb)
		ON CONFLICT (user_id, tenant_id) DO UPDATE SET
		role = EXCLUDED.role,
		permissions = EXCLUDED.permissions,
		updated_at = NOW()
	`, userID, tenantID, role, permissionsJSON)

	if err != nil {
		return fmt.Errorf("failed to grant tenant access: %w", err)
	}

	if f.Config.VerboseOutput {
		log.Printf("Granted %s access to user %d for tenant %s", role, userID, tenantID)
	}

	return nil
}

// CreateTestCustomer creates a test customer in the tenant database
func (f *DatabaseFixture) CreateTestCustomer(name, companyCode, status string) (int, error) {
	var customerID int
	err := f.TenantDB.QueryRow(`
		INSERT INTO test_integration.customers (tenant_id, name, company_code, status)
		VALUES ('longbeach', $1, $2, $3)
		RETURNING id
	`, name, companyCode, status).Scan(&customerID)

	if err != nil {
		return 0, fmt.Errorf("failed to create test customer: %w", err)
	}

	if f.Config.VerboseOutput {
		log.Printf("Created test customer: %s (ID: %d)", name, customerID)
	}

	return customerID, nil
}

// LinkCustomerContact links a customer to an auth user
func (f *DatabaseFixture) LinkCustomerContact(customerID, authUserID int, contactType string, isPrimary bool) error {
	_, err := f.TenantDB.Exec(`
		INSERT INTO test_integration.customer_contacts (customer_id, auth_user_id, contact_type, is_primary)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (customer_id, auth_user_id) DO UPDATE SET
		contact_type = EXCLUDED.contact_type,
		is_primary = EXCLUDED.is_primary
	`, customerID, authUserID, contactType, isPrimary)

	if err != nil {
		return fmt.Errorf("failed to link customer contact: %w", err)
	}

	if f.Config.VerboseOutput {
		log.Printf("Linked customer %d to auth user %d", customerID, authUserID)
	}

	return nil
}

// SeedTestData creates a set of test data for integration tests
func (f *DatabaseFixture) SeedTestData() error {
	// Create test users
	users := []struct {
		username, email, fullName, role string
		tenants                         []string
	}{
		{"admin_user", "admin@test.com", "Admin User", "ADMIN", []string{"longbeach", "bakersfield"}},
		{"operator_user", "operator@test.com", "Operator User", "OPERATOR", []string{"longbeach"}},
		{"customer_contact", "customer@test.com", "Customer Contact", "CUSTOMER_CONTACT", []string{"longbeach"}},
	}

	for _, user := range users {
		userID, err := f.CreateTestUser(user.username, user.email, user.fullName, user.role)
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.username, err)
		}

		for _, tenant := range user.tenants {
			if err := f.GrantTenantAccess(userID, tenant, user.role, []string{"view", "edit"}); err != nil {
				return fmt.Errorf("failed to grant tenant access for user %s: %w", user.username, err)
			}
		}
	}

	// Create test customers
	customers := []struct {
		name, code, status string
	}{
		{"Test Oil Company", "TOC001", "active"},
		{"Sample Gas Corp", "SGC002", "active"},
		{"Demo Energy LLC", "DEL003", "inactive"},
	}

	for _, customer := range customers {
		customerID, err := f.CreateTestCustomer(customer.name, customer.code, customer.status)
		if err != nil {
			return fmt.Errorf("failed to create customer %s: %w", customer.name, err)
		}

		// Link first customer to customer contact user
		if customer.code == "TOC001" {
			if err := f.LinkCustomerContact(customerID, 3, "PRIMARY", true); err != nil { // User ID 3 is customer_contact
				return fmt.Errorf("failed to link customer contact: %w", err)
			}
		}
	}

	if f.Config.VerboseOutput {
		log.Println("✅ Test data seeded successfully")
	}

	return nil
}

// CleanupTestData removes all test data
func (f *DatabaseFixture) CleanupTestData() error {
	if !f.Config.CleanupData {
		if f.Config.VerboseOutput {
			log.Println("⏭️  Skipping test data cleanup (TEST_CLEANUP_DATA=false)")
		}
		return nil
	}

	// Clean tenant database first (due to foreign key constraints)
	tenantTables := []string{
		"test_integration.customer_contacts",
		"test_integration.customers",
		"test_integration.test_data",
	}

	for _, table := range tenantTables {
		_, err := f.TenantDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			// Don't fail if table doesn't exist
			if f.Config.VerboseOutput {
				log.Printf("Warning: Failed to truncate %s: %v", table, err)
			}
		}
	}

	// Clean auth database
	authTables := []string{
		"test_integration.sessions",
		"test_integration.user_tenant_access",
		"test_integration.users",
	}

	for _, table := range authTables {
		_, err := f.AuthDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			// Don't fail if table doesn't exist
			if f.Config.VerboseOutput {
				log.Printf("Warning: Failed to truncate %s: %v", table, err)
			}
		}
	}

	if f.Config.VerboseOutput {
		log.Println("✅ Test data cleanup completed")
	}

	return nil
}

// DropTestSchemas removes all test schemas and tables
func (f *DatabaseFixture) DropTestSchemas() error {
	if !f.Config.CleanupData {
		return nil
	}

	_, err := f.AuthDB.Exec("DROP SCHEMA IF EXISTS test_integration CASCADE")
	if err != nil {
		log.Printf("Warning: Failed to drop auth test schema: %v", err)
	}

	_, err = f.TenantDB.Exec("DROP SCHEMA IF EXISTS test_integration CASCADE")
	if err != nil {
		log.Printf("Warning: Failed to drop tenant test schema: %v", err)
	}

	if f.Config.VerboseOutput {
		log.Println("✅ Test schemas dropped")
	}

	return nil
}

// Close closes all database connections
func (f *DatabaseFixture) Close() error {
	var errors []error

	if f.AuthDB != nil {
		if err := f.AuthDB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close auth database: %w", err))
		}
	}

	if f.TenantDB != nil {
		if err := f.TenantDB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close tenant database: %w", err))
		}
	}

	if f.DBManager != nil {
		f.DBManager.Close()
	}

	if len(errors) > 0 {
		return fmt.Errorf("database close errors: %v", errors)
	}

	if f.Config.VerboseOutput {
		log.Println("✅ Database connections closed")
	}

	return nil
}

// Utility functions for environment variable handling

func getEnvIntOrDefault(key string, defaultValue int) int {
	if str := os.Getenv(key); str != "" {
		if val, err := fmt.Sscanf(str, "%d", &defaultValue); err == nil && val == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if str := os.Getenv(key); str != "" {
		switch str {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}
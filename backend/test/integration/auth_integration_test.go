// backend/test/integration/auth_integration_test.go
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"

	"oilgas-backend/internal/shared/database"
)

type AuthIntegrationTestSuite struct {
	suite.Suite
	dbManager   *database.DatabaseManager
	ctx         context.Context
	authDB      *sql.DB
	tenantDB    *sql.DB
}

func (suite *AuthIntegrationTestSuite) SetupSuite() {
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

	suite.setupTestData()
}

func (suite *AuthIntegrationTestSuite) setupTestData() {
	// Clean existing test data
	suite.cleanupTestData()

	// Ensure test schemas exist
	suite.authDB.Exec("CREATE SCHEMA IF NOT EXISTS test_auth")
	suite.tenantDB.Exec("CREATE SCHEMA IF NOT EXISTS test_tenant")

	// Create test tables for auth integration testing
	suite.createAuthTestTables()
	suite.createTenantTestTables()
}

func (suite *AuthIntegrationTestSuite) createAuthTestTables() {
	// Create test users table if users table doesn't exist or for testing
	_, err := suite.authDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_auth.users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) UNIQUE,
			email VARCHAR(255) UNIQUE NOT NULL,
			full_name VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'OPERATOR',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	suite.Require().NoError(err)

	// Create test tenant access table
	_, err = suite.authDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_auth.user_tenant_access (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES test_auth.users(id) ON DELETE CASCADE,
			tenant_id VARCHAR(100) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'USER',
			permissions JSONB DEFAULT '[]'::jsonb,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, tenant_id)
		)
	`)
	suite.Require().NoError(err)

	// Create test sessions table
	_, err = suite.authDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_auth.sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id INTEGER REFERENCES test_auth.users(id) ON DELETE CASCADE,
			tenant_id VARCHAR(100) NOT NULL,
			token_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			last_accessed TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	suite.Require().NoError(err)
}

func (suite *AuthIntegrationTestSuite) createTenantTestTables() {
	// Create test customer contacts table (linking to auth users)
	_, err := suite.tenantDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_tenant.customer_contacts (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER NOT NULL,
			auth_user_id INTEGER NOT NULL, -- References auth.users(id)
			contact_type VARCHAR(50) DEFAULT 'PRIMARY',
			is_primary BOOLEAN DEFAULT false,
			is_active BOOLEAN DEFAULT true,
			full_name VARCHAR(255),
			email VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(customer_id, auth_user_id)
		)
	`)
	suite.Require().NoError(err)
}

func (suite *AuthIntegrationTestSuite) TearDownSuite() {
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

func (suite *AuthIntegrationTestSuite) cleanupTestData() {
	if suite.authDB != nil {
		suite.authDB.Exec("DROP SCHEMA IF EXISTS test_auth CASCADE")
	}
	if suite.tenantDB != nil {
		suite.tenantDB.Exec("DROP SCHEMA IF EXISTS test_tenant CASCADE")
	}
}

func (suite *AuthIntegrationTestSuite) SetupTest() {
	// Clean test data before each test
	if suite.authDB != nil {
		suite.authDB.Exec("TRUNCATE TABLE test_auth.sessions, test_auth.user_tenant_access, test_auth.users RESTART IDENTITY CASCADE")
	}
	if suite.tenantDB != nil {
		suite.tenantDB.Exec("TRUNCATE TABLE test_tenant.customer_contacts RESTART IDENTITY")
	}
}

// Test user creation and authentication flow across databases
func (suite *AuthIntegrationTestSuite) TestUserAuthenticationFlow() {
	suite.T().Log("Testing user authentication flow across databases...")

	// 1. Create user in auth database
	var userID int
	err := suite.authDB.QueryRow(`
		INSERT INTO test_auth.users (username, email, full_name, password_hash, role)
		VALUES ('testuser', 'test@example.com', 'Test User', '$2a$10$hashedpassword', 'OPERATOR')
		RETURNING id
	`).Scan(&userID)
	suite.Require().NoError(err)
	suite.Assert().Greater(userID, 0)

	// 2. Grant tenant access
	_, err = suite.authDB.Exec(`
		INSERT INTO test_auth.user_tenant_access (user_id, tenant_id, role, permissions)
		VALUES ($1, 'longbeach', 'OPERATOR', '["view_customers", "create_workorders"]'::jsonb)
	`, userID)
	suite.Require().NoError(err)

	// 3. Create session
	var sessionID string
	err = suite.authDB.QueryRow(`
		INSERT INTO test_auth.sessions (user_id, tenant_id, token_hash, expires_at)
		VALUES ($1, 'longbeach', 'hashed_token', NOW() + INTERVAL '1 hour')
		RETURNING id
	`, userID).Scan(&sessionID)
	suite.Require().NoError(err)

	// 4. Verify authentication data integrity
	var userData struct {
		ID       int
		Email    string
		Role     string
		IsActive bool
	}
	err = suite.authDB.QueryRow(`
		SELECT u.id, u.email, u.role, u.is_active
		FROM test_auth.users u
		WHERE u.id = $1
	`, userID).Scan(&userData.ID, &userData.Email, &userData.Role, &userData.IsActive)
	suite.Require().NoError(err)
	suite.Assert().Equal(userID, userData.ID)
	suite.Assert().Equal("test@example.com", userData.Email)
	suite.Assert().Equal("OPERATOR", userData.Role)
	suite.Assert().True(userData.IsActive)

	// 5. Verify tenant access
	var tenantAccess struct {
		TenantID    string
		Role        string
		Permissions string
	}
	err = suite.authDB.QueryRow(`
		SELECT tenant_id, role, permissions::text
		FROM test_auth.user_tenant_access
		WHERE user_id = $1 AND tenant_id = 'longbeach'
	`, userID).Scan(&tenantAccess.TenantID, &tenantAccess.Role, &tenantAccess.Permissions)
	suite.Require().NoError(err)
	suite.Assert().Equal("longbeach", tenantAccess.TenantID)
	suite.Assert().Equal("OPERATOR", tenantAccess.Role)
	suite.Assert().Contains(tenantAccess.Permissions, "view_customers")

	// 6. Verify session validity
	var sessionValid bool
	err = suite.authDB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM test_auth.sessions
			WHERE user_id = $1 AND tenant_id = 'longbeach' AND expires_at > NOW()
		)
	`, userID).Scan(&sessionValid)
	suite.Require().NoError(err)
	suite.Assert().True(sessionValid, "Session should be valid")

	suite.T().Logf("Successfully created and validated user %d with longbeach access", userID)
}

// Test customer contact linking between auth and tenant databases
func (suite *AuthIntegrationTestSuite) TestCustomerContactLinking() {
	suite.T().Log("Testing customer contact linking across databases...")

	// 1. Create user in auth database
	var authUserID int
	err := suite.authDB.QueryRow(`
		INSERT INTO test_auth.users (username, email, full_name, password_hash, role)
		VALUES ('customer_contact', 'contact@customer.com', 'Customer Contact', 'hashed_pw', 'CUSTOMER_CONTACT')
		RETURNING id
	`).Scan(&authUserID)
	suite.Require().NoError(err)

	// 2. Create customer contact link in tenant database
	_, err = suite.tenantDB.Exec(`
		INSERT INTO test_tenant.customer_contacts (customer_id, auth_user_id, contact_type, is_primary, full_name, email)
		VALUES (101, $1, 'PRIMARY', true, 'Customer Contact', 'contact@customer.com')
	`, authUserID)
	suite.Require().NoError(err)

	// 3. Verify cross-database linking integrity
	var contactData struct {
		CustomerID   int
		AuthUserID   int
		ContactType  string
		IsPrimary    bool
		FullName     string
		Email        string
	}
	err = suite.tenantDB.QueryRow(`
		SELECT customer_id, auth_user_id, contact_type, is_primary, full_name, email
		FROM test_tenant.customer_contacts
		WHERE auth_user_id = $1
	`, authUserID).Scan(
		&contactData.CustomerID,
		&contactData.AuthUserID,
		&contactData.ContactType,
		&contactData.IsPrimary,
		&contactData.FullName,
		&contactData.Email,
	)
	suite.Require().NoError(err)
	suite.Assert().Equal(101, contactData.CustomerID)
	suite.Assert().Equal(authUserID, contactData.AuthUserID)
	suite.Assert().Equal("PRIMARY", contactData.ContactType)
	suite.Assert().True(contactData.IsPrimary)

	// 4. Verify auth user still exists
	var authUserExists bool
	err = suite.authDB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM test_auth.users WHERE id = $1)
	`, authUserID).Scan(&authUserExists)
	suite.Require().NoError(err)
	suite.Assert().True(authUserExists, "Auth user should still exist")

	suite.T().Logf("Successfully linked auth user %d to customer 101", authUserID)
}

// Test multi-tenant access validation
func (suite *AuthIntegrationTestSuite) TestMultiTenantAccessValidation() {
	suite.T().Log("Testing multi-tenant access validation...")

	// Create user with access to multiple tenants
	var userID int
	err := suite.authDB.QueryRow(`
		INSERT INTO test_auth.users (username, email, full_name, password_hash, role)
		VALUES ('multitenant_user', 'multi@example.com', 'Multi Tenant User', 'hashed_pw', 'ADMIN')
		RETURNING id
	`).Scan(&userID)
	suite.Require().NoError(err)

	// Grant access to multiple tenants
	tenants := []string{"longbeach", "bakersfield", "houston"}
	for _, tenant := range tenants {
		_, err = suite.authDB.Exec(`
			INSERT INTO test_auth.user_tenant_access (user_id, tenant_id, role, permissions)
			VALUES ($1, $2, 'ADMIN', '["full_access"]'::jsonb)
		`, userID, tenant)
		suite.Require().NoError(err)
	}

	// Verify user has access to all tenants
	var accessCount int
	err = suite.authDB.QueryRow(`
		SELECT COUNT(*) FROM test_auth.user_tenant_access WHERE user_id = $1
	`, userID).Scan(&accessCount)
	suite.Require().NoError(err)
	suite.Assert().Equal(len(tenants), accessCount)

	// Test access validation for each tenant
	for _, tenant := range tenants {
		var hasAccess bool
		err = suite.authDB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM test_auth.user_tenant_access
				WHERE user_id = $1 AND tenant_id = $2 AND is_active = true
			)
		`, userID, tenant).Scan(&hasAccess)
		suite.Require().NoError(err)
		suite.Assert().True(hasAccess, fmt.Sprintf("User should have access to tenant %s", tenant))
	}

	// Test access denial for non-existent tenant
	var hasInvalidAccess bool
	err = suite.authDB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM test_auth.user_tenant_access
			WHERE user_id = $1 AND tenant_id = 'nonexistent'
		)
	`, userID).Scan(&hasInvalidAccess)
	suite.Require().NoError(err)
	suite.Assert().False(hasInvalidAccess, "User should not have access to nonexistent tenant")
}

// Test session management across tenants
func (suite *AuthIntegrationTestSuite) TestCrossTenantSessionManagement() {
	suite.T().Log("Testing cross-tenant session management...")

	// Create user
	var userID int
	err := suite.authDB.QueryRow(`
		INSERT INTO test_auth.users (username, email, full_name, password_hash, role)
		VALUES ('session_user', 'session@example.com', 'Session User', 'hashed_pw', 'OPERATOR')
		RETURNING id
	`).Scan(&userID)
	suite.Require().NoError(err)

	// Grant access to multiple tenants
	tenants := []string{"longbeach", "bakersfield"}
	for _, tenant := range tenants {
		_, err = suite.authDB.Exec(`
			INSERT INTO test_auth.user_tenant_access (user_id, tenant_id, role)
			VALUES ($1, $2, 'OPERATOR')
		`, userID, tenant)
		suite.Require().NoError(err)
	}

	// Create sessions for different tenants
	sessionTokens := make(map[string]string)
	for _, tenant := range tenants {
		var sessionID string
		err = suite.authDB.QueryRow(`
			INSERT INTO test_auth.sessions (user_id, tenant_id, token_hash, expires_at)
			VALUES ($1, $2, $3, NOW() + INTERVAL '1 hour')
			RETURNING id
		`, userID, tenant, fmt.Sprintf("token_%s_%d", tenant, userID)).Scan(&sessionID)
		suite.Require().NoError(err)
		sessionTokens[tenant] = sessionID
	}

	// Verify each session is tenant-specific
	for _, tenant := range tenants {
		var sessionExists bool
		err = suite.authDB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM test_auth.sessions
				WHERE user_id = $1 AND tenant_id = $2 AND expires_at > NOW()
			)
		`, userID, tenant).Scan(&sessionExists)
		suite.Require().NoError(err)
		suite.Assert().True(sessionExists, fmt.Sprintf("Session should exist for tenant %s", tenant))
	}

	// Test session cleanup
	_, err = suite.authDB.Exec(`
		DELETE FROM test_auth.sessions WHERE user_id = $1 AND tenant_id = 'longbeach'
	`, userID)
	suite.Require().NoError(err)

	// Verify longbeach session is gone but bakersfield remains
	var longbeachSessionExists bool
	err = suite.authDB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM test_auth.sessions WHERE user_id = $1 AND tenant_id = 'longbeach')
	`, userID).Scan(&longbeachSessionExists)
	suite.Require().NoError(err)
	suite.Assert().False(longbeachSessionExists, "Longbeach session should be deleted")

	var bakersfieldSessionExists bool
	err = suite.authDB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM test_auth.sessions WHERE user_id = $1 AND tenant_id = 'bakersfield')
	`, userID).Scan(&bakersfieldSessionExists)
	suite.Require().NoError(err)
	suite.Assert().True(bakersfieldSessionExists, "Bakersfield session should still exist")
}

// Test data consistency between auth and tenant databases
func (suite *AuthIntegrationTestSuite) TestDataConsistency() {
	suite.T().Log("Testing data consistency between auth and tenant databases...")

	// Create user in auth database
	var userID int
	err := suite.authDB.QueryRow(`
		INSERT INTO test_auth.users (username, email, full_name, password_hash, role)
		VALUES ('consistency_user', 'consistency@example.com', 'Consistency User', 'hashed_pw', 'CUSTOMER_CONTACT')
		RETURNING id
	`).Scan(&userID)
	suite.Require().NoError(err)

	// Create customer contact in tenant database
	_, err = suite.tenantDB.Exec(`
		INSERT INTO test_tenant.customer_contacts (customer_id, auth_user_id, full_name, email)
		VALUES (202, $1, 'Consistency User', 'consistency@example.com')
	`, userID)
	suite.Require().NoError(err)

	// Update user email in auth database
	newEmail := "updated_consistency@example.com"
	_, err = suite.authDB.Exec(`
		UPDATE test_auth.users SET email = $1, updated_at = NOW() WHERE id = $2
	`, newEmail, userID)
	suite.Require().NoError(err)

	// Verify auth database has updated email
	var authEmail string
	err = suite.authDB.QueryRow(`
		SELECT email FROM test_auth.users WHERE id = $1
	`, userID).Scan(&authEmail)
	suite.Require().NoError(err)
	suite.Assert().Equal(newEmail, authEmail)

	// In a real application, you'd have triggers or application logic to sync
	// For testing, manually update tenant database to simulate sync
	_, err = suite.tenantDB.Exec(`
		UPDATE test_tenant.customer_contacts SET email = $1 WHERE auth_user_id = $2
	`, newEmail, userID)
	suite.Require().NoError(err)

	// Verify consistency across databases
	var tenantEmail string
	err = suite.tenantDB.QueryRow(`
		SELECT email FROM test_tenant.customer_contacts WHERE auth_user_id = $1
	`, userID).Scan(&tenantEmail)
	suite.Require().NoError(err)
	suite.Assert().Equal(newEmail, tenantEmail, "Email should be consistent across databases")

	suite.T().Log("Data consistency validated across databases")
}

// Test transaction rollback across databases (conceptual test)
func (suite *AuthIntegrationTestSuite) TestTransactionalIntegrity() {
	suite.T().Log("Testing transactional integrity concepts...")

	// Start transaction in auth database
	tx, err := suite.authDB.Begin()
	suite.Require().NoError(err)

	// Insert user within transaction
	var userID int
	err = tx.QueryRow(`
		INSERT INTO test_auth.users (username, email, full_name, password_hash, role)
		VALUES ('tx_user', 'tx@example.com', 'Transaction User', 'hashed_pw', 'OPERATOR')
		RETURNING id
	`).Scan(&userID)
	suite.Require().NoError(err)

	// Simulate error condition - rollback transaction
	err = tx.Rollback()
	suite.Require().NoError(err)

	// Verify user was not created due to rollback
	var userExists bool
	err = suite.authDB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM test_auth.users WHERE email = 'tx@example.com')
	`).Scan(&userExists)
	suite.Require().NoError(err)
	suite.Assert().False(userExists, "User should not exist after transaction rollback")

	// Test successful transaction
	tx2, err := suite.authDB.Begin()
	suite.Require().NoError(err)

	err = tx2.QueryRow(`
		INSERT INTO test_auth.users (username, email, full_name, password_hash, role)
		VALUES ('tx_user_success', 'tx_success@example.com', 'Transaction Success User', 'hashed_pw', 'OPERATOR')
		RETURNING id
	`).Scan(&userID)
	suite.Require().NoError(err)

	err = tx2.Commit()
	suite.Require().NoError(err)

	// Verify user was created after successful commit
	err = suite.authDB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM test_auth.users WHERE email = 'tx_success@example.com')
	`).Scan(&userExists)
	suite.Require().NoError(err)
	suite.Assert().True(userExists, "User should exist after successful transaction commit")

	suite.T().Log("Transactional integrity concepts validated")
}

func TestAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}
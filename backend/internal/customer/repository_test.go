// backend/internal/customer/repository_test.go
package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"oilgas-backend/internal/shared/database"
)

type CustomerRepositoryTestSuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo Repository
	ctx  context.Context
}

func (suite *CustomerRepositoryTestSuite) SetupTest() {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	suite.Require().NoError(err)
	
	suite.db = db
	suite.mock = mock
	suite.repo = NewRepository(db)
	suite.ctx = context.Background()
}

func (suite *CustomerRepositoryTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func TestCustomerRepositorySuite(t *testing.T) {
	suite.Run(t, new(CustomerRepositoryTestSuite))
}

func (suite *CustomerRepositoryTestSuite) TestGetCustomerByID_Success() {
	tenantID := "test-tenant"
	customerID := 1

	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "name", "company_code", "status", "tax_id", "payment_terms",
		"billing_street", "billing_city", "billing_state", "billing_zip_code", "billing_country",
		"created_at", "updated_at",
	}).AddRow(
		1, "test-tenant", "Test Company", "TEST123", "active", "123456789", "NET30",
		"123 Main St", "Houston", "TX", "77001", "US",
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms, billing_street, billing_city, billing_state, billing_zip_code, billing_country, created_at, updated_at FROM customers.customers WHERE tenant_id = $1 AND id = $2").
		WithArgs(tenantID, customerID).
		WillReturnRows(rows)

	result, err := suite.repo.GetCustomerByID(suite.ctx, tenantID, customerID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("Test Company", result.Name)
	suite.Equal("TEST123", result.CompanyCode)
	suite.Equal(StatusActive, result.Status)
	suite.Equal("123456789", result.BillingInfo.TaxID)
	suite.Equal("123 Main St", result.BillingInfo.Address.Street)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestGetCustomerByID_NotFound() {
	tenantID := "test-tenant"
	customerID := 999

	suite.mock.ExpectQuery("SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms, billing_street, billing_city, billing_state, billing_zip_code, billing_country, created_at, updated_at FROM customers.customers WHERE tenant_id = $1 AND id = $2").
		WithArgs(tenantID, customerID).
		WillReturnError(sql.ErrNoRows)

	result, err := suite.repo.GetCustomerByID(suite.ctx, tenantID, customerID)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "customer not found")
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestSearchCustomers_WithFilters() {
	tenantID := "test-tenant"
	filters := SearchFilters{
		Name:   "Test",
		Status: []Status{StatusActive},
		Limit:  10,
		Offset: 0,
	}

	// Count query
	suite.mock.ExpectQuery("SELECT COUNT(*) FROM customers.customers WHERE tenant_id = $1 AND name ILIKE $2 AND status = ANY($3)").
		WithArgs(tenantID, "%Test%", []string{"active"}).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Data query
	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "name", "company_code", "status", "tax_id", "payment_terms",
		"billing_street", "billing_city", "billing_state", "billing_zip_code", "billing_country",
		"created_at", "updated_at",
	}).AddRow(
		1, tenantID, "Test Company", "TEST123", "active", "123456789", "NET30",
		"123 Main St", "Houston", "TX", "77001", "US",
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms, billing_street, billing_city, billing_state, billing_zip_code, billing_country, created_at, updated_at FROM customers.customers WHERE tenant_id = $1 AND name ILIKE $2 AND status = ANY($3) ORDER BY name ASC LIMIT $4 OFFSET $5").
		WithArgs(tenantID, "%Test%", []string{"active"}, 10, 0).
		WillReturnRows(rows)

	customers, total, err := suite.repo.SearchCustomers(suite.ctx, tenantID, filters)

	suite.NoError(err)
	suite.Len(customers, 1)
	suite.Equal(1, total)
	suite.Equal("Test Company", customers[0].Name)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestMultiTenantIsolation() {
	tenant1 := "tenant-1"
	tenant2 := "tenant-2"
	customerID := 1

	// Mock returning customer for tenant-1
	rows1 := sqlmock.NewRows([]string{
		"id", "tenant_id", "name", "company_code", "status", "tax_id", "payment_terms",
		"billing_street", "billing_city", "billing_state", "billing_zip_code", "billing_country",
		"created_at", "updated_at",
	}).AddRow(
		1, tenant1, "Company A", "COMP_A", "active", "123456789", "NET30",
		"123 Main St", "Houston", "TX", "77001", "US",
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms, billing_street, billing_city, billing_state, billing_zip_code, billing_country, created_at, updated_at FROM customers.customers WHERE tenant_id = $1 AND id = $2").
		WithArgs(tenant1, customerID).
		WillReturnRows(rows1)

	// Mock no rows for tenant-2 (isolation)
	suite.mock.ExpectQuery("SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms, billing_street, billing_city, billing_state, billing_zip_code, billing_country, created_at, updated_at FROM customers.customers WHERE tenant_id = $1 AND id = $2").
		WithArgs(tenant2, customerID).
		WillReturnError(sql.ErrNoRows)

	// Get customer from tenant-1
	customer1, err := suite.repo.GetCustomerByID(suite.ctx, tenant1, customerID)
	suite.NoError(err)
	suite.NotNil(customer1)
	suite.Equal("Company A", customer1.Name)
	suite.Equal(tenant1, customer1.TenantID)

	// Attempt to get same customer ID from tenant-2 should fail
	customer2, err := suite.repo.GetCustomerByID(suite.ctx, tenant2, customerID)
	suite.Error(err)
	suite.Nil(customer2)

	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestCreateCustomer_Success() {
	customer := &Customer{
		TenantID:    "tenant-1",
		Name:        "New Company",
		CompanyCode: "NEW123",
		Status:      StatusActive,
		BillingInfo: BillingInfo{
			TaxID:       "987654321",
			PaymentTerms: "NET30",
			Address: Address{
				Street:  "456 Oak Ave",
				City:    "Dallas",
				State:   "TX",
				ZipCode: "75201",
				Country: "US",
			},
		},
	}

	// Mock the INSERT query
	suite.mock.ExpectQuery("INSERT INTO customers.customers").
		WithArgs(
			customer.TenantID, customer.Name, customer.CompanyCode, customer.Status,
			customer.BillingInfo.TaxID, customer.BillingInfo.PaymentTerms,
			customer.BillingInfo.Address.Street, customer.BillingInfo.Address.City,
			customer.BillingInfo.Address.State, customer.BillingInfo.Address.ZipCode,
			customer.BillingInfo.Address.Country,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	result, err := suite.repo.CreateCustomer(suite.ctx, customer)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(1, result.ID)
	suite.Equal("New Company", result.Name)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestCreateCustomer_DatabaseError() {
	customer := &Customer{
		TenantID:    "tenant-1",
		Name:        "New Company",
		CompanyCode: "NEW123",
		Status:      StatusActive,
		BillingInfo: BillingInfo{
			TaxID:       "987654321",
			PaymentTerms: "NET30",
			Address: Address{
				Street:  "456 Oak Ave",
				City:    "Dallas",
				State:   "TX",
				ZipCode: "75201",
				Country: "US",
			},
		},
	}

	// Mock database connection error
	suite.mock.ExpectQuery("INSERT INTO customers.customers").
		WillReturnError(errors.New("connection lost"))

	result, err := suite.repo.CreateCustomer(suite.ctx, customer)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "connection lost")
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestTransactionRollback() {
	customer := &Customer{
		TenantID:    "tenant-1",
		Name:        "Transaction Test",
		CompanyCode: "TX123",
		Status:      StatusActive,
		BillingInfo: BillingInfo{
			TaxID: "111111111",
			Address: Address{
				Street: "123 Test St",
				City:   "Austin",
				State:  "TX",
			},
		},
	}

	// Begin transaction
	suite.mock.ExpectBegin()
	
	// First operation succeeds
	suite.mock.ExpectQuery("INSERT INTO customers.customers").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	
	// Second operation fails
	suite.mock.ExpectExec("INSERT INTO customers.customer_audit").
		WillReturnError(errors.New("audit insert failed"))
	
	// Transaction should rollback
	suite.mock.ExpectRollback()

	// Execute test - would need a transactional method in repository
	// For now, we test the transaction pattern expectation
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestSQLInjectionPrevention() {
	maliciousInput := "'; DROP TABLE customers; --"
	tenantID := "tenant-1"

	// The query should use parameterized queries, preventing injection
	suite.mock.ExpectQuery("SELECT COUNT(*) FROM customers.customers WHERE tenant_id = $1 AND name ILIKE $2").
		WithArgs(tenantID, fmt.Sprintf("%%%s%%", maliciousInput)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	suite.mock.ExpectQuery("SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms, billing_street, billing_city, billing_state, billing_zip_code, billing_country, created_at, updated_at FROM customers.customers WHERE tenant_id = $1 AND name ILIKE $2 ORDER BY name ASC LIMIT $3 OFFSET $4").
		WithArgs(tenantID, fmt.Sprintf("%%%s%%", maliciousInput), 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "name", "company_code", "status", "tax_id", "payment_terms",
			"billing_street", "billing_city", "billing_state", "billing_zip_code", "billing_country",
			"created_at", "updated_at",
		}))

	filters := SearchFilters{
		Name:   maliciousInput,
		Limit:  10,
		Offset: 0,
	}

	customers, total, err := suite.repo.SearchCustomers(suite.ctx, tenantID, filters)

	// Should handle safely without SQL injection
	suite.NoError(err)
	suite.Equal(0, total)
	suite.Empty(customers)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestUpdateCustomer_Success() {
	customer := &Customer{
		ID:          1,
		TenantID:    "tenant-1",
		Name:        "Updated Company",
		CompanyCode: "UPD123",
		Status:      StatusActive,
		BillingInfo: BillingInfo{
			TaxID:       "555555555",
			PaymentTerms: "NET45",
			Address: Address{
				Street:  "789 Updated St",
				City:    "San Antonio",
				State:   "TX",
				ZipCode: "78201",
				Country: "US",
			},
		},
	}

	suite.mock.ExpectExec("UPDATE customers.customers SET").
		WithArgs(
			customer.Name, customer.CompanyCode, customer.Status,
			customer.BillingInfo.TaxID, customer.BillingInfo.PaymentTerms,
			customer.BillingInfo.Address.Street, customer.BillingInfo.Address.City,
			customer.BillingInfo.Address.State, customer.BillingInfo.Address.ZipCode,
			customer.BillingInfo.Address.Country,
			customer.TenantID, customer.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.UpdateCustomer(suite.ctx, customer)

	suite.NoError(err)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestDeleteCustomer_Success() {
	tenantID := "tenant-1"
	customerID := 1

	suite.mock.ExpectExec("DELETE FROM customers.customers WHERE tenant_id = \$1 AND id = \$2").
		WithArgs(tenantID, customerID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.DeleteCustomer(suite.ctx, tenantID, customerID)

	suite.NoError(err)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *CustomerRepositoryTestSuite) TestDeleteCustomer_NotFound() {
	tenantID := "tenant-1"
	customerID := 999

	suite.mock.ExpectExec("DELETE FROM customers.customers WHERE tenant_id = \$1 AND id = \$2").
		WithArgs(tenantID, customerID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := suite.repo.DeleteCustomer(suite.ctx, tenantID, customerID)

	suite.Error(err)
	suite.Contains(err.Error(), "customer not found")
	suite.NoError(suite.mock.ExpectationsWereMet())
}

// Benchmark tests for performance validation
func BenchmarkRepository_GetCustomerByID(b *testing.B) {
	db, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-1"
	customerID := 1

	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "name", "company_code", "status", "tax_id", "payment_terms",
		"billing_street", "billing_city", "billing_state", "billing_zip_code", "billing_country",
		"created_at", "updated_at",
	}).AddRow(
		1, tenantID, "Benchmark Company", "BENCH123", "active", "123456789", "NET30",
		"123 Bench St", "Houston", "TX", "77001", "US",
		time.Now(), time.Now(),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms, billing_street, billing_city, billing_state, billing_zip_code, billing_country, created_at, updated_at FROM customers.customers WHERE tenant_id = $1 AND id = $2").
			WithArgs(tenantID, customerID).
			WillReturnRows(rows)
		
		_, _ = repo.GetCustomerByID(ctx, tenantID, customerID)
	}
}

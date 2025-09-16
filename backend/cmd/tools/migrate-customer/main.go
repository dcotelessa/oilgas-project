// cmd/tools/migrate-customers/main.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"oilgas-backend/internal/customer"
	"oilgas-backend/internal/shared/database"
)

type AccessCustomer struct {
	ID              string `json:"id"`
	CompanyName     string `json:"company_name"`
	CompanyCode     string `json:"company_code"`
	BillingAddress  string `json:"billing_address"`
	City            string `json:"city"`
	State           string `json:"state"`
	ZipCode         string `json:"zip_code"`
	TaxID           string `json:"tax_id"`
	PaymentTerms    string `json:"payment_terms"`
	Status          string `json:"status"`
	PrimaryContact  string `json:"primary_contact"`
	ContactEmail    string `json:"contact_email"`
	ContactPhone    string `json:"contact_phone"`
}

func main() {
	var (
		tenant    = flag.String("tenant", "longbeach", "Tenant ID")
		file      = flag.String("file", "", "Path to customer export JSON file")
		dryRun    = flag.Bool("dry-run", true, "Perform dry run without writing to database")
		batchSize = flag.Int("batch-size", 10, "Number of customers to process in each batch")
	)
	flag.Parse()

	if *file == "" {
		log.Fatal("Please provide path to customer export file with --file")
	}

	// Read export file
	data, err := os.ReadFile(*file)
	if err != nil {
		log.Fatal("Failed to read export file:", err)
	}

	var accessCustomers []AccessCustomer
	if err := json.Unmarshal(data, &accessCustomers); err != nil {
		log.Fatal("Failed to parse export file:", err)
	}

	fmt.Printf("Found %d customers in export file\n", len(accessCustomers))

	if *dryRun {
		fmt.Println("🔍 DRY RUN - No data will be written to database")
		performDryRun(accessCustomers)
		return
	}

	// Database setup
	dbConfig := &database.Config{
		CentralDBURL: os.Getenv("CENTRAL_AUTH_DB_URL"),
		TenantDBs: map[string]string{
			*tenant: getTenantDBURL(*tenant),
		},
		MaxOpenConns: 10,
		MaxIdleConns: 2,
		MaxLifetime:  time.Hour,
	}

	dbManager, err := database.NewDatabaseManager(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to databases:", err)
	}
	defer dbManager.Close()

	// Run migration
	result := performMigration(context.Background(), dbManager, *tenant, accessCustomers, *batchSize)
	
	// Report results
	fmt.Printf("\n📊 Migration Results:\n")
	fmt.Printf("  Total: %d\n", result.Total)
	fmt.Printf("  Successful: %d\n", result.Successful)
	fmt.Printf("  Failed: %d\n", result.Failed)
	fmt.Printf("  Duration: %v\n", result.Duration)

	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  %s: %s\n", err.LegacyID, err.Error)
		}
	}

	if result.Failed == 0 {
		fmt.Println("\n✅ All customers migrated successfully!")
		
		// Perform data integrity validation
		fmt.Println("\n🔍 Validating data integrity...")
		integrityReport, err := ValidateDataIntegrity(context.Background(), dbManager, *tenant, result.Total)
		if err != nil {
			fmt.Printf("⚠️  Data integrity validation failed: %s\n", err)
		} else {
			printIntegrityReport(integrityReport)
		}
	}
}

func getTenantDBURL(tenant string) string {
	switch tenant {
	case "longbeach":
		return os.Getenv("LONGBEACH_DB_URL")
	case "bakersfield":
		return os.Getenv("BAKERSFIELD_DB_URL")
	default:
		log.Fatalf("Unknown tenant: %s", tenant)
		return ""
	}
}

func performDryRun(customers []AccessCustomer) {
	validator := &customer.ValidationUtils{}
	
	validCount := 0
	invalidCount := 0
	
	for i, ac := range customers {
		fmt.Printf("\n--- Customer %d ---\n", i+1)
		fmt.Printf("ID: %s\n", ac.ID)
		fmt.Printf("Name: %s\n", ac.CompanyName)
		fmt.Printf("Code: %s\n", ac.CompanyCode)
		
		// Validate company code
		normalizedCode := validator.NormalizeCompanyCode(ac.CompanyCode)
		if normalizedCode != ac.CompanyCode {
			fmt.Printf("  Code will be normalized: %s -> %s\n", ac.CompanyCode, normalizedCode)
		}
		
		// Validate tax ID
		if ac.TaxID != "" && !validator.ValidateTaxID(ac.TaxID) {
			fmt.Printf("  ⚠️  Invalid tax ID format: %s\n", ac.TaxID)
			invalidCount++
			continue
		}
		
		// Check required fields
		if ac.CompanyName == "" || ac.CompanyCode == "" {
			fmt.Printf("  ❌ Missing required fields\n")
			invalidCount++
			continue
		}
		
		fmt.Printf("  ✅ Valid\n")
		validCount++
	}
	
	fmt.Printf("\n📊 Dry Run Summary:\n")
	fmt.Printf("  Valid: %d\n", validCount)
	fmt.Printf("  Invalid: %d\n", invalidCount)
	fmt.Printf("  Total: %d\n", len(customers))
}

type MigrationResult struct {
	Total      int
	Successful int
	Failed     int
	Duration   time.Duration
	StartTime  time.Time
	Errors     []MigrationError
}

type MigrationError struct {
	LegacyID string
	Error    string
}

func performMigration(ctx context.Context, dbManager *database.DatabaseManager, tenant string, customers []AccessCustomer, batchSize int) *MigrationResult {
	startTime := time.Now()
	result := &MigrationResult{
		Total:     len(customers),
		StartTime: startTime,
	}
	
	db, err := dbManager.GetTenantDB(tenant)
	if err != nil {
		log.Fatal("Failed to get tenant database:", err)
	}
	
	// Initialize customer service for proper validation and insertion
	repo := customer.NewRepository(dbManager)
	service := customer.NewService(repo, nil, nil) // No auth service or cache for migration
	validator := &customer.ValidationUtils{}
	
	fmt.Printf("🚀 Starting migration to tenant: %s\n", tenant)
	fmt.Printf("📊 Total customers to migrate: %d\n", len(customers))
	
	for i, ac := range customers {
		// Convert to internal customer format
		internalCustomer := convertAccessCustomer(ac, tenant, validator)
		
		// Validate customer data
		if err := validateMigrationData(internalCustomer, ac); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, MigrationError{
				LegacyID: ac.ID,
				Error:    fmt.Sprintf("Validation failed: %s", err.Error()),
			})
			continue
		}
		
		// Use service to create customer (ensures proper multi-tenant handling)
		_, err := service.CreateCustomer(ctx, tenant, internalCustomer)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, MigrationError{
				LegacyID: ac.ID,
				Error:    fmt.Sprintf("Database insertion failed: %s", err.Error()),
			})
		} else {
			result.Successful++
		}
		
		// Progress reporting
		if (i+1)%batchSize == 0 {
			fmt.Printf("📈 Processed %d/%d customers... (Success: %d, Failed: %d)\n", 
				i+1, len(customers), result.Successful, result.Failed)
		}
	}
	
	result.Duration = time.Since(startTime)
	return result
}

func convertAccessCustomer(ac AccessCustomer, tenantID string, validator *customer.ValidationUtils) *customer.Customer {
	// Map status
	status := customer.StatusActive
	switch ac.Status {
	case "inactive", "disabled":
		status = customer.StatusInactive
	case "suspended", "hold":
		status = customer.StatusSuspended
	}
	
	// Parse address
	address := customer.Address{
		Street:  ac.BillingAddress,
		City:    ac.City,
		State:   ac.State,
		ZipCode: ac.ZipCode,
		Country: "US",
	}
	
	return &customer.Customer{
		TenantID:    tenantID,
		Name:        strings.TrimSpace(ac.CompanyName),
		CompanyCode: validator.NormalizeCompanyCode(ac.CompanyCode),
		Status:      status,
		BillingInfo: customer.BillingInfo{
			TaxID:        validator.FormatTaxID(ac.TaxID),
			PaymentTerms: normalizePaymentTerms(ac.PaymentTerms),
			Address:      address,
		},
		Contacts: []customer.Contact{
			{
				Name:        ac.PrimaryContact,
				Email:       ac.ContactEmail,
				Phone:       ac.ContactPhone,
				ContactType: customer.ContactTypePrimary,
				IsActive:    true,
			},
		},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func printIntegrityReport(report *DataIntegrityReport) {
	fmt.Printf("\n📋 Data Integrity Report:\n")
	fmt.Printf("  Original records: %d\n", report.TotalRecords)
	fmt.Printf("  Migrated records: %d\n", report.MigratedRecords)
	fmt.Printf("  Duplicate companies: %d\n", report.DuplicateCompanies)
	fmt.Printf("  Records with missing data: %d\n", report.MissingRequiredData)
	
	if len(report.DataQualityIssues) == 0 {
		fmt.Println("  ✅ No data quality issues detected")
	} else {
		fmt.Printf("  ⚠️  Data quality issues:\n")
		for _, issue := range report.DataQualityIssues {
			fmt.Printf("    • %s\n", issue)
		}
	}
}

func normalizePaymentTerms(terms string) string {
	terms = strings.TrimSpace(strings.ToUpper(terms))
	switch terms {
	case "NET 30", "N30", "30 DAYS":
		return "NET30"
	case "NET 15", "N15", "15 DAYS":
		return "NET15"
	case "COD", "CASH ON DELIVERY":
		return "COD"
	case "PREPAID", "PREPAYMENT":
		return "PREPAID"
	default:
		if terms == "" {
			return "NET30" // Default
		}
		return terms
	}
}

func validateMigrationData(customer *customer.Customer, original AccessCustomer) error {
	if customer.Name == "" {
		return fmt.Errorf("company name is required")
	}
	if customer.CompanyCode == "" {
		return fmt.Errorf("company code is required")
	}
	if len(customer.CompanyCode) > 20 {
		return fmt.Errorf("company code too long: %s", customer.CompanyCode)
	}
	return nil
}

type DataIntegrityReport struct {
	TotalRecords        int
	MigratedRecords     int
	DuplicateCompanies  int
	MissingRequiredData int
	DataQualityIssues   []string
}

func ValidateDataIntegrity(ctx context.Context, dbManager *database.DatabaseManager, tenant string, expectedCount int) (*DataIntegrityReport, error) {
	db, err := dbManager.GetTenantDB(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	report := &DataIntegrityReport{
		TotalRecords: expectedCount,
	}

	// Count migrated records
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM customers")
	if err := row.Scan(&report.MigratedRecords); err != nil {
		return nil, fmt.Errorf("failed to count migrated records: %w", err)
	}

	// Check for duplicates by company code
	row = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT company_code, COUNT(*) 
			FROM customers 
			GROUP BY company_code 
			HAVING COUNT(*) > 1
		) duplicates`)
	if err := row.Scan(&report.DuplicateCompanies); err != nil {
		return nil, fmt.Errorf("failed to check duplicates: %w", err)
	}

	// Check for missing required data
	row = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM customers 
		WHERE name IS NULL OR name = '' OR company_code IS NULL OR company_code = ''`)
	if err := row.Scan(&report.MissingRequiredData); err != nil {
		return nil, fmt.Errorf("failed to check missing data: %w", err)
	}

	// Collect data quality issues
	if report.MigratedRecords != expectedCount {
		report.DataQualityIssues = append(report.DataQualityIssues,
			fmt.Sprintf("Record count mismatch: expected %d, got %d", expectedCount, report.MigratedRecords))
	}

	if report.DuplicateCompanies > 0 {
		report.DataQualityIssues = append(report.DataQualityIssues,
			fmt.Sprintf("%d duplicate company codes found", report.DuplicateCompanies))
	}

	if report.MissingRequiredData > 0 {
		report.DataQualityIssues = append(report.DataQualityIssues,
			fmt.Sprintf("%d records with missing required data", report.MissingRequiredData))
	}

	return report, nil
}

// backend/cmd/simple-importer/main.go
// Direct CSV importer for your actual customers.csv structure
package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type CustomerRecord struct {
	CustID         *int    // custid from CSV
	Customer       string  // customer from CSV
	BillingAddress *string // billingaddress from CSV
	BillingCity    *string // billingcity from CSV
	BillingState   *string // billingstate from CSV
	BillingZipcode *string // billingzipcode from CSV
	Contact        *string // contact from CSV
	Phone          *string // phone from CSV
	Fax            *string // fax from CSV
	Email          *string // email from CSV
	
	// Color system - exact column names from your CSV
	Color1  *string // color1
	Color2  *string // color2
	Color3  *string // color3
	Color4  *string // color4
	Color5  *string // color5
	
	// Loss percentages
	Loss1   *string // loss1
	Loss2   *string // loss2
	Loss3   *string // loss3
	Loss4   *string // loss4
	Loss5   *string // loss5
	
	// W-String colors
	WSColor1 *string // wscolor1
	WSColor2 *string // wscolor2
	WSColor3 *string // wscolor3
	WSColor4 *string // wscolor4 - YOUR CSV HAS THIS!
	WSColor5 *string // wscolor5
	
	// W-String losses
	WSLoss1  *string // wsloss1
	WSLoss2  *string // wsloss2
	WSLoss3  *string // wsloss3
	WSLoss4  *string // wsloss4 - YOUR CSV HAS THIS!
	WSLoss5  *string // wsloss5
	
	Deleted    *int    // deleted from CSV
	TenantID   string  // Added for multi-tenant
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: simple-importer <customers.csv> [tenant_id]")
	}
	
	csvFile := os.Args[1]
	tenantID := "local-dev"
	if len(os.Args) > 2 {
		tenantID = os.Args[2]
	}
	
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://oilgas_user:oilgas_pass@localhost:5432/oilgas_dev"
	}
	
	log.Printf("Importing %s to database...", csvFile)
	log.Printf("Tenant ID: %s", tenantID)
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	
	// Set tenant context
	_, err = db.Exec("SELECT set_tenant_context($1)", tenantID)
	if err != nil {
		log.Fatal("Failed to set tenant context:", err)
	}
	
	// Import customers
	imported, skipped, err := importCustomers(db, csvFile, tenantID)
	if err != nil {
		log.Fatal("Import failed:", err)
	}
	
	log.Printf("Import completed: %d imported, %d skipped", imported, skipped)
}

func importCustomers(db *sql.DB, csvFile, tenantID string) (int, int, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, 0, err
	}
	
	if len(records) == 0 {
		return 0, 0, fmt.Errorf("empty CSV file")
	}
	
	// Map headers to indices - based on your actual CSV structure
	headers := records[0]
	columnMap := make(map[string]int)
	
	for i, header := range headers {
		// Normalize header names to match your CSV exactly
		normalized := strings.ToLower(strings.TrimSpace(header))
		columnMap[normalized] = i
	}
	
	log.Printf("Found %d columns: %v", len(headers), headers)
	log.Printf("Mapped columns: %v", columnMap)
	
	// Prepare insert statement - matching your exact CSV structure
	stmt, err := db.Prepare(`
		INSERT INTO store.customers (
			custid, customer, billingaddress, billingcity, billingstate, billingzipcode,
			contact, phone, fax, email,
			color1, color2, color3, color4, color5,
			loss1, loss2, loss3, loss4, loss5,
			wscolor1, wscolor2, wscolor3, wscolor4, wscolor5,
			wsloss1, wsloss2, wsloss3, wsloss4, wsloss5,
			deleted, tenant_id, imported_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25,
			$26, $27, $28, $29, $30,
			$31, $32, $33, $34
		) ON CONFLICT DO NOTHING`)
	
	if err != nil {
		return 0, 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	imported := 0
	skipped := 0
	now := time.Now()
	
	// Process each record
	for i := 1; i < len(records); i++ {
		record := records[i]
		
		// Extract customer data from CSV record
		customer := extractCustomer(record, columnMap, tenantID)
		
		// Skip if no customer name
		if customer.Customer == "" {
			log.Printf("Skipping record %d: no customer name", i)
			skipped++
			continue
		}
		
		// Execute insert
		_, err = stmt.Exec(
			customer.CustID,
			customer.Customer,
			customer.BillingAddress,
			customer.BillingCity,
			customer.BillingState,
			customer.BillingZipcode,
			customer.Contact,
			customer.Phone,
			customer.Fax,
			customer.Email,
			customer.Color1,
			customer.Color2,
			customer.Color3,
			customer.Color4,
			customer.Color5,
			customer.Loss1,
			customer.Loss2,
			customer.Loss3,
			customer.Loss4,
			customer.Loss5,
			customer.WSColor1,
			customer.WSColor2,
			customer.WSColor3,
			customer.WSColor4,
			customer.WSColor5,
			customer.WSLoss1,
			customer.WSLoss2,
			customer.WSLoss3,
			customer.WSLoss4,
			customer.WSLoss5,
			customer.Deleted,
			customer.TenantID,
			now,
			now,
		)
		
		if err != nil {
			log.Printf("Failed to insert record %d (%s): %v", i, customer.Customer, err)
			skipped++
		} else {
			imported++
			if imported%100 == 0 {
				log.Printf("Imported %d records...", imported)
			}
		}
	}
	
	return imported, skipped, nil
}

func extractCustomer(record []string, columnMap map[string]int, tenantID string) CustomerRecord {
	customer := CustomerRecord{
		TenantID: tenantID,
	}
	
	// Extract each field based on your CSV column names
	customer.CustID = getIntField(record, columnMap, "custid")
	customer.Customer = getStringField(record, columnMap, "customer")
	customer.BillingAddress = getStringPtrField(record, columnMap, "billingaddress")
	customer.BillingCity = getStringPtrField(record, columnMap, "billingcity")
	customer.BillingState = getStringPtrField(record, columnMap, "billingstate")
	customer.BillingZipcode = getStringPtrField(record, columnMap, "billingzipcode")
	customer.Contact = getStringPtrField(record, columnMap, "contact")
	customer.Phone = getStringPtrField(record, columnMap, "phone")
	customer.Fax = getStringPtrField(record, columnMap, "fax")
	customer.Email = getStringPtrField(record, columnMap, "email")
	
	// Color system
	customer.Color1 = getStringPtrField(record, columnMap, "color1")
	customer.Color2 = getStringPtrField(record, columnMap, "color2")
	customer.Color3 = getStringPtrField(record, columnMap, "color3")
	customer.Color4 = getStringPtrField(record, columnMap, "color4")
	customer.Color5 = getStringPtrField(record, columnMap, "color5")
	
	// Loss percentages
	customer.Loss1 = getStringPtrField(record, columnMap, "loss1")
	customer.Loss2 = getStringPtrField(record, columnMap, "loss2")
	customer.Loss3 = getStringPtrField(record, columnMap, "loss3")
	customer.Loss4 = getStringPtrField(record, columnMap, "loss4")
	customer.Loss5 = getStringPtrField(record, columnMap, "loss5")
	
	// W-String colors
	customer.WSColor1 = getStringPtrField(record, columnMap, "wscolor1")
	customer.WSColor2 = getStringPtrField(record, columnMap, "wscolor2")
	customer.WSColor3 = getStringPtrField(record, columnMap, "wscolor3")
	customer.WSColor4 = getStringPtrField(record, columnMap, "wscolor4") // Your CSV has this!
	customer.WSColor5 = getStringPtrField(record, columnMap, "wscolor5")
	
	// W-String losses
	customer.WSLoss1 = getStringPtrField(record, columnMap, "wsloss1")
	customer.WSLoss2 = getStringPtrField(record, columnMap, "wsloss2")
	customer.WSLoss3 = getStringPtrField(record, columnMap, "wsloss3")
	customer.WSLoss4 = getStringPtrField(record, columnMap, "wsloss4") // Your CSV has this!
	customer.WSLoss5 = getStringPtrField(record, columnMap, "wsloss5")
	
	customer.Deleted = getIntField(record, columnMap, "deleted")
	
	return customer
}

// Helper functions to extract fields safely
func getStringField(record []string, columnMap map[string]int, fieldName string) string {
	if idx, exists := columnMap[fieldName]; exists && idx < len(record) {
		value := strings.TrimSpace(record[idx])
		if value == "" || value == "NULL" || value == "null" {
			return ""
		}
		return value
	}
	return ""
}

func getStringPtrField(record []string, columnMap map[string]int, fieldName string) *string {
	value := getStringField(record, columnMap, fieldName)
	if value == "" {
		return nil
	}
	return &value
}

func getIntField(record []string, columnMap map[string]int, fieldName string) *int {
	value := getStringField(record, columnMap, fieldName)
	if value == "" {
		return nil
	}
	
	if intVal, err := strconv.Atoi(value); err == nil {
		return &intVal
	}
	
	return nil
}

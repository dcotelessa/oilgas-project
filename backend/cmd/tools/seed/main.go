// cmd/tools/seed/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"oilgas-backend/internal/customer"
	"oilgas-backend/internal/shared/database"
)

func main() {
	var (
		tenant    = flag.String("tenant", "longbeach", "Tenant ID")
		customers = flag.Int("customers", 0, "Number of test customers to create")
	)
	flag.Parse()

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

	ctx := context.Background()

	if *customers > 0 {
		if err := seedCustomers(ctx, dbManager, *tenant, *customers); err != nil {
			log.Fatal("Failed to seed customers:", err)
		}
	} else {
		fmt.Println("Usage: seed --customers=N [--tenant=TENANT]")
		fmt.Println("Example: seed --customers=50 --tenant=longbeach")
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

func seedCustomers(ctx context.Context, dbManager *database.DatabaseManager, tenant string, count int) error {
	db, err := dbManager.GetTenantDB(tenant)
	if err != nil {
		return err
	}

	companies := []string{
		"Chevron Corporation", "ExxonMobil", "ConocoPhillips", "EOG Resources",
		"Pioneer Natural Resources", "Devon Energy", "Marathon Petroleum",
		"Valero Energy", "Phillips 66", "Kinder Morgan", "Enterprise Products",
		"Plains All American", "ONEOK", "Enbridge", "TC Energy",
	}

	states := []string{"TX", "CA", "OK", "ND", "PA", "WV", "LA", "NM", "WY", "CO"}
	cities := []string{"Houston", "Dallas", "Oklahoma City", "Denver", "Midland", "Bakersfield", "Long Beach"}

	fmt.Printf("Creating %d test customers...\n", count)

	for i := 0; i < count; i++ {
		company := companies[rand.Intn(len(companies))]
		companyCode := fmt.Sprintf("TST%03d", i+1)
		
		query := `
			INSERT INTO store.customers (
				name, company_code, status, tax_id, payment_terms,
				billing_street, billing_city, billing_state, billing_zip_code, billing_country
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

		_, err := db.ExecContext(ctx, query,
			fmt.Sprintf("%s Test Division", company),
			companyCode,
			"active",
			fmt.Sprintf("%02d-%07d", rand.Intn(99), rand.Intn(9999999)),
			"NET30",
			fmt.Sprintf("%d Test St", rand.Intn(9999)+1),
			cities[rand.Intn(len(cities))],
			states[rand.Intn(len(states))],
			fmt.Sprintf("%05d", rand.Intn(99999)),
			"US",
		)

		if err != nil {
			return fmt.Errorf("failed to create customer %d: %w", i+1, err)
		}

		if (i+1)%10 == 0 {
			fmt.Printf("Created %d customers...\n", i+1)
		}
	}

	fmt.Printf("✅ Successfully created %d test customers\n", count)
	return nil
}

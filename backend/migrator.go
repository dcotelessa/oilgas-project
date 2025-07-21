package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Oil & Gas Inventory System - Database Migrator")
		fmt.Println("Usage:")
		fmt.Println("  migrator migrate [env]  - Run migrations")
		fmt.Println("  migrator seed [env]     - Seed database")
		fmt.Println("  migrator status [env]   - Show status")
		fmt.Println("  migrator reset [env]    - Reset database")
		os.Exit(1)
	}

	command := os.Args[1]
	env := "local"
	if len(os.Args) > 2 {
		env = os.Args[2]
	}

	// Get database URL with CONSISTENT naming enforced
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// ENFORCED: Use consistent database name in fallback
		databaseURL = "postgresql://postgres:postgres123@localhost:5432/oil_gas_inventory?sslmode=disable"
		fmt.Println("⚠️  Using fallback DATABASE_URL with consistent naming")
	}

	// SAFETY CHECK: Ensure we're using the consistent database name
	if !contains(databaseURL, "oil_gas_inventory") {
		fmt.Printf("🔧 Database URL correction needed. Current: %s\n", databaseURL)
		// Force consistent database name
		databaseURL = "postgresql://postgres:postgres123@localhost:5432/oil_gas_inventory?sslmode=disable"
		fmt.Printf("🔧 Corrected to: %s\n", databaseURL)
	}

	fmt.Printf("🔌 Connecting to database (env: %s)\n", env)
	fmt.Printf("🔗 Database URL: %s\n", databaseURL)

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connection successful")

	// Execute command
	switch command {
	case "migrate":
		runMigrations(db, env)
	case "seed":
		runSeeds(db, env)
	case "status":
		showStatus(db, env)
	case "reset":
		resetDatabase(db, env)
	default:
		log.Fatalf("❌ Unknown command: %s", command)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func runMigrations(db *sql.DB, env string) {
	fmt.Printf("🔄 Running migrations for environment: %s\n", env)

	// Step 1: Create schemas first (separately with error checking)
	fmt.Println("📁 Step 1: Creating schemas...")
	
	_, err := db.Exec("CREATE SCHEMA IF NOT EXISTS store")
	if err != nil {
		log.Fatalf("❌ Failed to create store schema: %v", err)
	}
	fmt.Println("✅ Store schema created")

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS migrations")
	if err != nil {
		log.Fatalf("❌ Failed to create migrations schema: %v", err)
	}
	fmt.Println("✅ Migrations schema created")

	// Step 2: Create migration tracking table
	fmt.Println("📋 Step 2: Creating migration tracking...")
	
	migrationTrackingSQL := `
	CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = db.Exec(migrationTrackingSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create migration tracking table: %v", err)
	}
	fmt.Println("✅ Migration tracking table created")

	// Step 3: Set search path
	fmt.Println("🛤️ Step 3: Setting search path...")
	
	_, err = db.Exec("SET search_path TO store, public")
	if err != nil {
		log.Fatalf("❌ Failed to set search path: %v", err)
	}
	fmt.Println("✅ Search path set to store, public")

	// Step 4: Create reference tables (no dependencies)
	fmt.Println("📊 Step 4: Creating reference tables...")
	
	gradeTableSQL := `
	CREATE TABLE IF NOT EXISTS store.grade (
		grade VARCHAR(10) PRIMARY KEY,
		description TEXT
	)`
	
	_, err = db.Exec(gradeTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create grade table: %v", err)
	}
	fmt.Println("✅ Grade table created")

	sizesTableSQL := `
	CREATE TABLE IF NOT EXISTS store.sizes (
		size_id SERIAL PRIMARY KEY,
		size VARCHAR(50) NOT NULL UNIQUE,
		description TEXT
	)`
	
	_, err = db.Exec(sizesTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create sizes table: %v", err)
	}
	fmt.Println("✅ Sizes table created")

	// Step 5: Create customers table
	fmt.Println("👥 Step 5: Creating customers table...")
	
	customersTableSQL := `
	CREATE TABLE IF NOT EXISTS store.customers (
		customer_id SERIAL PRIMARY KEY,
		customer VARCHAR(255) NOT NULL,
		billing_address TEXT,
		billing_city VARCHAR(100),
		billing_state VARCHAR(50),
		billing_zipcode VARCHAR(20),
		contact VARCHAR(255),
		phone VARCHAR(50),
		fax VARCHAR(50),
		email VARCHAR(255),
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = db.Exec(customersTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create customers table: %v", err)
	}
	fmt.Println("✅ Customers table created")

	// Step 6: Create inventory table (with foreign keys)
	fmt.Println("📦 Step 6: Creating inventory table...")
	
	inventoryTableSQL := `
	CREATE TABLE IF NOT EXISTS store.inventory (
		id SERIAL PRIMARY KEY,
		username VARCHAR(100),
		work_order VARCHAR(100),
		r_number VARCHAR(100),
		customer_id INTEGER REFERENCES store.customers(customer_id),
		customer VARCHAR(255),
		joints INTEGER,
		rack VARCHAR(50),
		size VARCHAR(50),
		weight DECIMAL(10,2),
		grade VARCHAR(10) REFERENCES store.grade(grade),
		connection VARCHAR(100),
		ctd VARCHAR(100),
		w_string VARCHAR(100),
		swgcc VARCHAR(100),
		color VARCHAR(50),
		customer_po VARCHAR(100),
		fletcher VARCHAR(100),
		date_in DATE,
		date_out DATE,
		well_in VARCHAR(255),
		lease_in VARCHAR(255),
		well_out VARCHAR(255),
		lease_out VARCHAR(255),
		trucking VARCHAR(100),
		trailer VARCHAR(100),
		location VARCHAR(100),
		notes TEXT,
		pcode VARCHAR(50),
		cn VARCHAR(50),
		ordered_by VARCHAR(100),
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = db.Exec(inventoryTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create inventory table: %v", err)
	}
	fmt.Println("✅ Inventory table created")

	// Step 7: Create received table (with foreign keys)
	fmt.Println("📨 Step 7: Creating received table...")
	
	receivedTableSQL := `
	CREATE TABLE IF NOT EXISTS store.received (
		id SERIAL PRIMARY KEY,
		work_order VARCHAR(100),
		customer_id INTEGER REFERENCES store.customers(customer_id),
		customer VARCHAR(255),
		joints INTEGER,
		rack VARCHAR(50),
		size_id INTEGER REFERENCES store.sizes(size_id),
		size VARCHAR(50),
		weight DECIMAL(10,2),
		grade VARCHAR(10) REFERENCES store.grade(grade),
		connection VARCHAR(100),
		ctd VARCHAR(100),
		w_string VARCHAR(100),
		well VARCHAR(255),
		lease VARCHAR(255),
		ordered_by VARCHAR(100),
		notes TEXT,
		customer_po VARCHAR(100),
		date_received DATE,
		background TEXT,
		norm VARCHAR(100),
		services TEXT,
		bill_to_id INTEGER,
		entered_by VARCHAR(100),
		when_entered TIMESTAMP,
		trucking VARCHAR(100),
		trailer VARCHAR(100),
		in_production BOOLEAN DEFAULT FALSE,
		inspected_date DATE,
		threading_date DATE,
		straighten_required BOOLEAN DEFAULT FALSE,
		excess_material BOOLEAN DEFAULT FALSE,
		complete BOOLEAN DEFAULT FALSE,
		inspected_by VARCHAR(100),
		updated_by VARCHAR(100),
		when_updated TIMESTAMP,
		deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err = db.Exec(receivedTableSQL)
	if err != nil {
		log.Fatalf("❌ Failed to create received table: %v", err)
	}
	fmt.Println("✅ Received table created")

	// Step 8: Create indexes
	fmt.Println("📈 Step 8: Creating performance indexes...")
	
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_date_in ON store.inventory(date_in)",
		"CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order)",
		"CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received)",
	}
	
	for i, indexSQL := range indexes {
		_, err = db.Exec(indexSQL)
		if err != nil {
			log.Fatalf("❌ Failed to create index %d: %v", i+1, err)
		}
	}
	fmt.Println("✅ Performance indexes created")

	// Step 9: Record migration
	fmt.Println("📝 Step 9: Recording migration...")
	
	_, err = db.Exec("INSERT INTO migrations.schema_migrations (version, name) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING", "001", "initial_oil_gas_schema_stepwise")
	if err != nil {
		log.Fatalf("❌ Failed to record migration: %v", err)
	}
	fmt.Println("✅ Migration recorded")

	fmt.Println("🎉 Migrations completed successfully!")
}

func runSeeds(db *sql.DB, env string) {
	fmt.Printf("🌱 Running seeds for environment: %s\n", env)

	// Set search path
	if _, err := db.Exec("SET search_path TO store, public;"); err != nil {
		log.Fatalf("❌ Failed to set search path: %v", err)
	}

	// Clear existing data (development only) with RESTART IDENTITY to reset sequences
	if env == "local" || env == "development" {
		fmt.Println("🧹 Clearing existing data and resetting sequences...")
		clearSQL := `
		TRUNCATE TABLE store.received CASCADE;
		TRUNCATE TABLE store.inventory CASCADE;
		TRUNCATE TABLE store.customers RESTART IDENTITY CASCADE;
		TRUNCATE TABLE store.sizes RESTART IDENTITY CASCADE;
		DELETE FROM store.grade;
		`
		if _, err := db.Exec(clearSQL); err != nil {
			log.Fatalf("❌ Failed to clear data: %v", err)
		}
		fmt.Println("✅ Data cleared, sequences reset to start at 1")
	}

	// Step 1: Insert reference data (no SERIAL dependencies)
	fmt.Println("📊 Inserting reference data...")
	referenceSQL := `
	-- Insert oil & gas industry grades (no SERIAL, uses explicit PKs)
	INSERT INTO store.grade (grade, description) VALUES 
	('J55', 'Standard grade steel casing - most common'),
	('L80', 'Higher strength grade for moderate environments'),
	('N80', 'Medium strength grade for standard applications'),
	('P105', 'High performance grade for demanding conditions'),
	('P110', 'Premium performance grade for extreme environments'),
	('Q125', 'Ultra-high strength grade for specialized applications'),
	('C75', 'Carbon steel grade for basic applications'),
	('C95', 'Higher carbon steel grade'),
	('T95', 'Tough grade for harsh environments');
	
	-- Insert common pipe sizes (uses SERIAL size_id, will start at 1)
	INSERT INTO store.sizes (size, description) VALUES 
	('4 1/2"', '4.5 inch diameter - small casing'),
	('5"', '5 inch diameter - intermediate casing'),
	('5 1/2"', '5.5 inch diameter - common production casing'),
	('7"', '7 inch diameter - intermediate casing'),
	('8 5/8"', '8.625 inch diameter - surface casing'),
	('9 5/8"', '9.625 inch diameter - surface casing'),
	('10 3/4"', '10.75 inch diameter - surface casing'),
	('13 3/8"', '13.375 inch diameter - surface casing'),
	('16"', '16 inch diameter - conductor casing'),
	('18 5/8"', '18.625 inch diameter - conductor casing'),
	('20"', '20 inch diameter - large conductor casing'),
	('24"', '24 inch diameter - extra large conductor'),
	('30"', '30 inch diameter - structural casing');
	`

	if _, err := db.Exec(referenceSQL); err != nil {
		log.Fatalf("❌ Failed to insert reference data: %v", err)
	}
	fmt.Println("✅ Reference data inserted")

	// Step 2: Insert customers and capture their actual SERIAL IDs
	fmt.Println("👥 Inserting customers and capturing SERIAL IDs...")
	
	type Customer struct {
		ID   int
		Name string
	}
	
	customers := make([]Customer, 0)
	
	// Insert customers one by one using RETURNING to capture actual customer_id
	customerData := [][]string{
		{"Permian Basin Energy", "1234 Oil Field Rd", "Midland", "TX", "79701", "John Smith", "432-555-0101", "operations@permianbasin.com"},
		{"Eagle Ford Solutions", "5678 Shale Ave", "San Antonio", "TX", "78201", "Sarah Johnson", "210-555-0201", "drilling@eagleford.com"},
		{"Bakken Industries", "9012 Prairie Blvd", "Williston", "ND", "58801", "Mike Wilson", "701-555-0301", "procurement@bakken.com"},
		{"Gulf Coast Drilling", "3456 Offshore Dr", "Houston", "TX", "77001", "Lisa Brown", "713-555-0401", "logistics@gulfcoast.com"},
		{"Marcellus Gas Co", "7890 Mountain View", "Pittsburgh", "PA", "15201", "Robert Davis", "412-555-0501", "operations@marcellus.com"},
	}
	
	for _, data := range customerData {
		var customerID int
		err := db.QueryRow(`
			INSERT INTO store.customers (customer, billing_address, billing_city, billing_state, billing_zipcode, contact, phone, email) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
			RETURNING customer_id`,
			data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7]).Scan(&customerID)
		
		if err != nil {
			log.Fatalf("❌ Failed to insert customer %s: %v", data[0], err)
		}
		
		customers = append(customers, Customer{ID: customerID, Name: data[0]})
		fmt.Printf("  ✅ %s (customer_id: %d)\n", data[0], customerID)
	}

	// Step 3: Query size_id values to avoid SERIAL assumptions
	fmt.Println("📏 Querying size IDs to avoid SERIAL assumptions...")
	sizeMap := make(map[string]int)
	rows, err := db.Query("SELECT size_id, size FROM store.sizes ORDER BY size_id")
	if err != nil {
		log.Fatalf("❌ Failed to query sizes: %v", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var sizeID int
		var size string
		if err := rows.Scan(&sizeID, &size); err != nil {
			log.Fatalf("❌ Failed to scan size: %v", err)
		}
		sizeMap[size] = sizeID
		fmt.Printf("  📏 %s = size_id %d\n", size, sizeID)
	}

	// Step 4: Insert inventory using captured customer_id values
	fmt.Println("📦 Inserting inventory using actual customer IDs...")
	
	inventoryData := []struct {
		workOrder string
		customerIdx int  // Index into customers array
		joints    int
		size      string
		weight    float64
		grade     string
		connection string
		dateIn    string
		wellIn    string
		leaseIn   string
		location  string
		notes     string
	}{
		{"WO-2024-001", 0, 100, "5 1/2\"", 2500.50, "L80", "BTC", "2024-01-15", "Well-PB-001", "Lease-PB-A", "Yard-A", "Standard production casing"},
		{"WO-2024-002", 1, 150, "7\"", 4200.75, "P110", "VAM TOP", "2024-01-16", "Well-EF-002", "Lease-EF-B", "Yard-B", "High pressure application"},
		{"WO-2024-003", 2, 75, "9 5/8\"", 6800.25, "N80", "LTC", "2024-01-17", "Well-BK-003", "Lease-BK-C", "Yard-C", "Surface casing"},
		{"WO-2024-004", 3, 200, "5 1/2\"", 5000.00, "J55", "STC", "2024-01-18", "Well-GC-004", "Lease-GC-D", "Yard-A", "Offshore application"},
	}
	
	for _, inv := range inventoryData {
		customer := customers[inv.customerIdx]
		_, err := db.Exec(`
			INSERT INTO store.inventory (work_order, customer_id, customer, joints, size, weight, grade, connection, date_in, well_in, lease_in, location, notes) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			inv.workOrder, customer.ID, customer.Name, inv.joints, inv.size, inv.weight, inv.grade, inv.connection, inv.dateIn, inv.wellIn, inv.leaseIn, inv.location, inv.notes)
		
		if err != nil {
			log.Fatalf("❌ Failed to insert inventory %s: %v", inv.workOrder, err)
		}
		fmt.Printf("  ✅ %s for %s (customer_id: %d)\n", inv.workOrder, customer.Name, customer.ID)
	}

	// Step 5: Insert received orders using actual customer_id and size_id values
	fmt.Println("📨 Inserting received orders using actual SERIAL values...")
	
	receivedData := []struct {
		workOrder string
		customerIdx int  // Index into customers array
		joints    int
		size      string
		weight    float64
		grade     string
		connection string
		dateReceived string
		well      string
		lease     string
		orderedBy string
		notes     string
	}{
		{"WO-2024-005", 0, 80, "7\"", 3200.00, "L80", "BTC", "2024-01-20", "Well-PB-005", "Lease-PB-E", "John Smith", "Expedited order"},
		{"WO-2024-006", 4, 120, "5 1/2\"", 3000.00, "P110", "VAM TOP", "2024-01-21", "Well-MG-006", "Lease-MG-F", "Robert Davis", "High pressure specs"},
		{"WO-2024-007", 1, 90, "8 5/8\"", 7200.00, "N80", "LTC", "2024-01-22", "Well-EF-007", "Lease-EF-G", "Sarah Johnson", "Surface casing rush"},
	}
	
	for _, rec := range receivedData {
		customer := customers[rec.customerIdx]
		sizeID, exists := sizeMap[rec.size]
		if !exists {
			log.Fatalf("❌ Size %s not found in sizes table", rec.size)
		}
		
		_, err := db.Exec(`
			INSERT INTO store.received (work_order, customer_id, customer, joints, size_id, size, weight, grade, connection, date_received, well, lease, ordered_by, notes, in_production, complete) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
			rec.workOrder, customer.ID, customer.Name, rec.joints, sizeID, rec.size, rec.weight, rec.grade, rec.connection, rec.dateReceived, rec.well, rec.lease, rec.orderedBy, rec.notes, false, false)
		
		if err != nil {
			log.Fatalf("❌ Failed to insert received order %s: %v", rec.workOrder, err)
		}
		fmt.Printf("  ✅ %s for %s (customer_id: %d, size_id: %d)\n", rec.workOrder, customer.Name, customer.ID, sizeID)
	}

	fmt.Println("✅ Seeding completed successfully - no SERIAL assumptions!")
}

func showStatus(db *sql.DB, env string) {
	fmt.Printf("📊 Database Status (env: %s)\n", env)
	fmt.Printf("============================\n")

	// Set search path
	if _, err := db.Exec("SET search_path TO store, public;"); err != nil {
		fmt.Printf("❌ Failed to set search path: %v\n", err)
		return
	}

	// Check each table
	tables := []string{"customers", "grade", "sizes", "inventory", "received"}
	
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM store.%s", table)
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", table, err)
		} else {
			fmt.Printf("✅ %s: %d records\n", table, count)
		}
	}

	// Check SERIAL sequence status
	fmt.Println("\n🔢 SERIAL Sequence Status:")
	sequences := []struct{
		table string
		sequence string
		column string
	}{
		{"customers", "customers_customer_id_seq", "customer_id"},
		{"sizes", "sizes_size_id_seq", "size_id"},
		{"inventory", "inventory_id_seq", "id"},
		{"received", "received_id_seq", "id"},
	}
	
	for _, seq := range sequences {
		var lastValue, nextValue sql.NullInt64
		err := db.QueryRow(fmt.Sprintf("SELECT last_value, (last_value + 1) as next_value FROM store.%s", seq.sequence)).Scan(&lastValue, &nextValue)
		if err != nil {
			fmt.Printf("  ⚠️  %s sequence: Error - %v\n", seq.table, err)
		} else {
			if lastValue.Valid {
				fmt.Printf("  📈 %s.%s: last=%d, next=%d\n", seq.table, seq.column, lastValue.Int64, nextValue.Int64)
			} else {
				fmt.Printf("  📈 %s.%s: not used yet, next=1\n", seq.table, seq.column)
			}
		}
	}

	// Test foreign key relationships with actual IDs
	fmt.Println("\n🔗 Foreign Key Validation:")
	
	var customerID int
	var customerName, city string
	var joints int
	err := db.QueryRow(`
		SELECT c.customer_id, c.customer, c.billing_city, i.joints 
		FROM store.customers c 
		JOIN store.inventory i ON c.customer_id = i.customer_id 
		LIMIT 1
	`).Scan(&customerID, &customerName, &city, &joints)
	
	if err != nil {
		fmt.Printf("❌ Customer-Inventory join failed: %v\n", err)
	} else {
		fmt.Printf("✅ Customer-Inventory join: ID %d - %s (%s) - %d joints\n", customerID, customerName, city, joints)
	}

	// Test size_id relationships
	var sizeID int
	var size string
	var receivedJoints int
	err = db.QueryRow(`
		SELECT s.size_id, s.size, r.joints
		FROM store.sizes s
		JOIN store.received r ON s.size_id = r.size_id
		LIMIT 1
	`).Scan(&sizeID, &size, &receivedJoints)
	
	if err != nil {
		fmt.Printf("❌ Size-Received join failed: %v\n", err)
	} else {
		fmt.Printf("✅ Size-Received join: size_id %d - %s - %d joints\n", sizeID, size, receivedJoints)
	}

	// Show actual customer and size ID mappings to verify no assumptions
	fmt.Println("\n📋 ID Verification (no hardcoded assumptions):")
	
	fmt.Println("  Customers:")
	customerRows, err := db.Query("SELECT customer_id, customer FROM store.customers ORDER BY customer_id")
	if err != nil {
		fmt.Printf("    ❌ Failed to query customers: %v\n", err)
	} else {
		defer customerRows.Close()
		for customerRows.Next() {
			var id int
			var name string
			if err := customerRows.Scan(&id, &name); err != nil {
				fmt.Printf("    ❌ Failed to scan customer: %v\n", err)
			} else {
				fmt.Printf("    📋 customer_id %d: %s\n", id, name)
			}
		}
	}

	fmt.Println("  Sizes:")
	sizeRows, err := db.Query("SELECT size_id, size FROM store.sizes ORDER BY size_id LIMIT 5")
	if err != nil {
		fmt.Printf("    ❌ Failed to query sizes: %v\n", err)
	} else {
		defer sizeRows.Close()
		for sizeRows.Next() {
			var id int
			var sizeName string
			if err := sizeRows.Scan(&id, &sizeName); err != nil {
				fmt.Printf("    ❌ Failed to scan size: %v\n", err)
			} else {
				fmt.Printf("    📏 size_id %d: %s\n", id, sizeName)
			}
		}
	}

	fmt.Println("\n✅ Status check complete - all SERIAL sequences properly handled")
}

func resetDatabase(db *sql.DB, env string) {
	if env == "production" || env == "prod" {
		log.Fatal("❌ Reset not allowed in production environment")
	}

	fmt.Printf("⚠️ Resetting database (env: %s)...\n", env)

	resetSQL := `
	DROP SCHEMA IF EXISTS store CASCADE;
	DROP SCHEMA IF EXISTS migrations CASCADE;
	`

	if _, err := db.Exec(resetSQL); err != nil {
		log.Fatalf("❌ Reset failed: %v", err)
	}

	fmt.Println("✅ Database reset complete")
	fmt.Println("Run 'go run migrator.go migrate' and 'go run migrator.go seed' to restore")
}

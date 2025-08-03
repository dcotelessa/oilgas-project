// backend/cmd/customer-cleaner/main.go
// Enhanced customer data cleaner with deduplication and naming standards
package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// CleaningRules defines our conversion standards for all future domains
type CleaningRules struct {
	// Column naming conventions
	ColumnMappings map[string]string
	
	// Deduplication settings
	DuplicateThreshold float64  // Similarity threshold (0.0 - 1.0)
	CompanyVariations  map[string]string // Known company name variations
	
	// Validation rules
	RequiredFields []string
	
	// Data transformation rules
	StateNormalizations map[string]string
	PhoneFormat        string
}

type CustomerRecord struct {
	// Standardized field names (no abbreviations)
	CustomerID         *int    `db:"customer_id"`       // custid -> customer_id
	CustomerName       string  `db:"customer_name"`     // customer -> customer_name
	BillingAddress     *string `db:"billing_address"`   // billingaddress -> billing_address
	BillingCity        *string `db:"billing_city"`      // billingcity -> billing_city  
	BillingState       *string `db:"billing_state"`     // billingstate -> billing_state
	BillingZipCode     *string `db:"billing_zip_code"`  // billingzipcode -> billing_zip_code
	ContactName        *string `db:"contact_name"`      // contact -> contact_name
	PhoneNumber        *string `db:"phone_number"`      // phone -> phone_number
	FaxNumber          *string `db:"fax_number"`        // fax -> fax_number
	EmailAddress       *string `db:"email_address"`     // email -> email_address
	
	// Color system - standardized names
	ColorGrade1        *string `db:"color_grade_1"`     // color1 -> color_grade_1
	ColorGrade2        *string `db:"color_grade_2"`     // color2 -> color_grade_2
	ColorGrade3        *string `db:"color_grade_3"`     // color3 -> color_grade_3
	ColorGrade4        *string `db:"color_grade_4"`     // color4 -> color_grade_4
	ColorGrade5        *string `db:"color_grade_5"`     // color5 -> color_grade_5
	
	// Loss percentages - standardized names
	WallLoss1          *string `db:"wall_loss_1"`       // loss1 -> wall_loss_1
	WallLoss2          *string `db:"wall_loss_2"`       // loss2 -> wall_loss_2
	WallLoss3          *string `db:"wall_loss_3"`       // loss3 -> wall_loss_3
	WallLoss4          *string `db:"wall_loss_4"`       // loss4 -> wall_loss_4
	WallLoss5          *string `db:"wall_loss_5"`       // loss5 -> wall_loss_5
	
	// W-String system - standardized names
	WStringColor1      *string `db:"wstring_color_1"`   // wscolor1 -> wstring_color_1
	WStringColor2      *string `db:"wstring_color_2"`   // wscolor2 -> wstring_color_2
	WStringColor3      *string `db:"wstring_color_3"`   // wscolor3 -> wstring_color_3
	WStringColor4      *string `db:"wstring_color_4"`   // wscolor4 -> wstring_color_4
	WStringColor5      *string `db:"wstring_color_5"`   // wscolor5 -> wstring_color_5
	
	// W-String losses - standardized names
	WStringLoss1       *string `db:"wstring_loss_1"`    // wsloss1 -> wstring_loss_1
	WStringLoss2       *string `db:"wstring_loss_2"`    // wsloss2 -> wstring_loss_2
	WStringLoss3       *string `db:"wstring_loss_3"`    // wsloss3 -> wstring_loss_3
	WStringLoss4       *string `db:"wstring_loss_4"`    // wsloss4 -> wstring_loss_4
	WStringLoss5       *string `db:"wstring_loss_5"`    // wsloss5 -> wstring_loss_5
	
	IsDeleted          bool    `db:"is_deleted"`        // deleted -> is_deleted
	TenantID           string  `db:"tenant_id"`
	
	// Computed fields for deduplication
	NormalizedName     string  `db:"-"`                 // For duplicate detection
	AddressHash        string  `db:"-"`                 // For address comparison
	DuplicateScore     float64 `db:"-"`                 // Similarity score
	PotentialDuplicate bool    `db:"-"`                 // Flagged as potential duplicate
}

type DuplicateGroup struct {
	Records []CustomerRecord
	Score   float64
	Reason  string
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: customer-cleaner <input.csv> <output.csv> [tenant_id]")
	}
	
	inputFile := os.Args[1]
	outputFile := os.Args[2]
	tenantID := "local-dev"
	if len(os.Args) > 3 {
		tenantID = os.Args[3]
	}
	
	log.Printf("🧹 Enhanced Customer Data Cleaner")
	log.Printf("Input: %s", inputFile)
	log.Printf("Output: %s", outputFile)
	log.Printf("Tenant: %s", tenantID)
	
	// Initialize cleaning rules
	rules := initializeCleaningRules()
	
	// Process the CSV
	cleaner := NewCustomerCleaner(rules)
	
	cleaned, duplicates, err := cleaner.ProcessFile(inputFile, tenantID)
	if err != nil {
		log.Fatal("Processing failed:", err)
	}
	
	// Report duplicates
	if len(duplicates) > 0 {
		log.Printf("⚠️  Found %d potential duplicate groups", len(duplicates))
		for i, group := range duplicates {
			log.Printf("Duplicate group %d (score: %.2f, reason: %s):", i+1, group.Score, group.Reason)
			for _, record := range group.Records {
				log.Printf("  - %s (%s)", record.CustomerName, getAddressSummary(record))
			}
		}
	}
	
	// Write cleaned data
	err = writeClearedData(outputFile, cleaned)
	if err != nil {
		log.Fatal("Failed to write output:", err)
	}
	
	log.Printf("✅ Processing complete")
	log.Printf("   Original records: %d", len(cleaned)+getTotalDuplicates(duplicates))
	log.Printf("   Clean records: %d", len(cleaned))
	log.Printf("   Potential duplicates: %d", getTotalDuplicates(duplicates))
	log.Printf("   Output: %s", outputFile)
}

type CustomerCleaner struct {
	rules            CleaningRules
	existingCustomers map[string]CustomerRecord // For cross-tenant duplicate detection
}

func NewCustomerCleaner(rules CleaningRules) *CustomerCleaner {
	return &CustomerCleaner{
		rules:            rules,
		existingCustomers: make(map[string]CustomerRecord),
	}
}

func initializeCleaningRules() CleaningRules {
	return CleaningRules{
		// Standard column mappings (no abbreviations rule)
		ColumnMappings: map[string]string{
			"custid":          "customer_id",
			"customer":        "customer_name",
			"billingaddress":  "billing_address",
			"billingcity":     "billing_city",
			"billingstate":    "billing_state",
			"billingzipcode":  "billing_zip_code",
			"contact":         "contact_name",
			"phone":           "phone_number",
			"fax":             "fax_number",
			"email":           "email_address",
			"color1":          "color_grade_1",
			"color2":          "color_grade_2",
			"color3":          "color_grade_3",
			"color4":          "color_grade_4",
			"color5":          "color_grade_5",
			"loss1":           "wall_loss_1",
			"loss2":           "wall_loss_2",
			"loss3":           "wall_loss_3",
			"loss4":           "wall_loss_4",
			"loss5":           "wall_loss_5",
			"wscolor1":        "wstring_color_1",
			"wscolor2":        "wstring_color_2",
			"wscolor3":        "wstring_color_3",
			"wscolor4":        "wstring_color_4",
			"wscolor5":        "wstring_color_5",
			"wsloss1":         "wstring_loss_1",
			"wsloss2":         "wstring_loss_2",
			"wsloss3":         "wstring_loss_3",
			"wsloss4":         "wstring_loss_4",
			"wsloss5":         "wstring_loss_5",
			"deleted":         "is_deleted",
		},
		
		// Duplicate detection threshold
		DuplicateThreshold: 0.85, // 85% similarity triggers duplicate flag
		
		// Known company name variations (for oil & gas industry)
		CompanyVariations: map[string]string{
			"chevron corp":           "Chevron Corporation",
			"chevron corporation":    "Chevron Corporation",
			"chevron":                "Chevron Corporation",
			"exxon mobil":           "ExxonMobil Corporation",
			"exxonmobil":            "ExxonMobil Corporation",
			"exxon":                 "ExxonMobil Corporation",
			"bp":                    "BP America Inc.",
			"bp america":            "BP America Inc.",
			"bp america inc":        "BP America Inc.",
			"shell":                 "Shell Oil Company",
			"shell oil":             "Shell Oil Company",
			"shell oil company":     "Shell Oil Company",
			"conocophillips":        "ConocoPhillips",
			"conoco phillips":       "ConocoPhillips",
			"conoco":                "ConocoPhillips",
			"marathon":              "Marathon Oil Corporation",
			"marathon oil":          "Marathon Oil Corporation",
			"marathon petroleum":    "Marathon Petroleum Corporation",
			"valero":                "Valero Energy Corporation",
			"valero energy":         "Valero Energy Corporation",
			"phillips 66":           "Phillips 66",
			"phillips66":            "Phillips 66",
			"kinder morgan":         "Kinder Morgan Inc.",
			"enterprise":            "Enterprise Products Partners",
			"plains":                "Plains All American Pipeline",
		},
		
		RequiredFields: []string{"customer_name"},
		
		// State normalizations
		StateNormalizations: map[string]string{
			"texas":           "TX",
			"oklahoma":        "OK",
			"louisiana":       "LA",
			"new mexico":      "NM",
			"north dakota":    "ND",
			"west virginia":   "WV",
			"pennsylvania":    "PA",
			"wyoming":         "WY",
			"colorado":        "CO",
			"utah":            "UT",
			"california":      "CA",
			"alaska":          "AK",
		},
		
		PhoneFormat: "(###) ###-####",
	}
}

func (c *CustomerCleaner) ProcessFile(inputFile, tenantID string) ([]CustomerRecord, []DuplicateGroup, error) {
	// Read CSV
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	
	if len(records) == 0 {
		return nil, nil, fmt.Errorf("empty CSV file")
	}
	
	// Map headers
	headers := records[0]
	columnMap := c.mapHeaders(headers)
	
	// Process records
	var customers []CustomerRecord
	for i := 1; i < len(records); i++ {
		customer := c.extractAndCleanRecord(records[i], columnMap, tenantID)
		if customer != nil {
			customers = append(customers, *customer)
		}
	}
	
	// Detect duplicates
	cleanCustomers, duplicates := c.detectDuplicates(customers)
	
	return cleanCustomers, duplicates, nil
}

func (c *CustomerCleaner) mapHeaders(headers []string) map[string]int {
	columnMap := make(map[string]int)
	
	for i, header := range headers {
		normalized := strings.ToLower(strings.TrimSpace(header))
		columnMap[normalized] = i
	}
	
	return columnMap
}

func (c *CustomerCleaner) extractAndCleanRecord(record []string, columnMap map[string]int, tenantID string) *CustomerRecord {
	customer := &CustomerRecord{
		TenantID: tenantID,
	}
	
	// Extract and clean each field with standardized naming
	customer.CustomerID = c.getIntField(record, columnMap, "custid")
	customer.CustomerName = c.cleanCustomerName(c.getStringField(record, columnMap, "customer"))
	
	// Skip if no customer name
	if customer.CustomerName == "" {
		return nil
	}
	
	customer.BillingAddress = c.getCleanStringField(record, columnMap, "billingaddress")
	customer.BillingCity = c.getCleanStringField(record, columnMap, "billingcity")
	customer.BillingState = c.cleanState(c.getStringField(record, columnMap, "billingstate"))
	customer.BillingZipCode = c.cleanZipCode(c.getStringField(record, columnMap, "billingzipcode"))
	customer.ContactName = c.cleanPersonName(c.getStringField(record, columnMap, "contact"))
	customer.PhoneNumber = c.cleanPhoneNumber(c.getStringField(record, columnMap, "phone"))
	customer.FaxNumber = c.cleanPhoneNumber(c.getStringField(record, columnMap, "fax"))
	customer.EmailAddress = c.cleanEmail(c.getStringField(record, columnMap, "email"))
	
	// Color system
	customer.ColorGrade1 = c.getCleanStringField(record, columnMap, "color1")
	customer.ColorGrade2 = c.getCleanStringField(record, columnMap, "color2")
	customer.ColorGrade3 = c.getCleanStringField(record, columnMap, "color3")
	customer.ColorGrade4 = c.getCleanStringField(record, columnMap, "color4")
	customer.ColorGrade5 = c.getCleanStringField(record, columnMap, "color5")
	
	// Loss percentages
	customer.WallLoss1 = c.cleanPercentage(c.getStringField(record, columnMap, "loss1"))
	customer.WallLoss2 = c.cleanPercentage(c.getStringField(record, columnMap, "loss2"))
	customer.WallLoss3 = c.cleanPercentage(c.getStringField(record, columnMap, "loss3"))
	customer.WallLoss4 = c.cleanPercentage(c.getStringField(record, columnMap, "loss4"))
	customer.WallLoss5 = c.cleanPercentage(c.getStringField(record, columnMap, "loss5"))
	
	// W-String colors
	customer.WStringColor1 = c.getCleanStringField(record, columnMap, "wscolor1")
	customer.WStringColor2 = c.getCleanStringField(record, columnMap, "wscolor2")
	customer.WStringColor3 = c.getCleanStringField(record, columnMap, "wscolor3")
	customer.WStringColor4 = c.getCleanStringField(record, columnMap, "wscolor4")
	customer.WStringColor5 = c.getCleanStringField(record, columnMap, "wscolor5")
	
	// W-String losses
	customer.WStringLoss1 = c.cleanPercentage(c.getStringField(record, columnMap, "wsloss1"))
	customer.WStringLoss2 = c.cleanPercentage(c.getStringField(record, columnMap, "wsloss2"))
	customer.WStringLoss3 = c.cleanPercentage(c.getStringField(record, columnMap, "wsloss3"))
	customer.WStringLoss4 = c.cleanPercentage(c.getStringField(record, columnMap, "wsloss4"))
	customer.WStringLoss5 = c.cleanPercentage(c.getStringField(record, columnMap, "wsloss5"))
	
	// Deleted flag
	deleted := c.getStringField(record, columnMap, "deleted")
	customer.IsDeleted = deleted == "1" || strings.ToLower(deleted) == "true"
	
	// Generate fields for duplicate detection
	customer.NormalizedName = c.normalizeCompanyName(customer.CustomerName)
	customer.AddressHash = c.generateAddressHash(customer)
	
	return customer
}

// Duplicate detection logic
func (c *CustomerCleaner) detectDuplicates(customers []CustomerRecord) ([]CustomerRecord, []DuplicateGroup) {
	var cleanCustomers []CustomerRecord
	var duplicateGroups []DuplicateGroup
	
	// Group by normalized name for efficiency
	nameGroups := make(map[string][]CustomerRecord)
	for _, customer := range customers {
		key := customer.NormalizedName
		nameGroups[key] = append(nameGroups[key], customer)
	}
	
	// Check each group for duplicates
	for _, group := range nameGroups {
		if len(group) == 1 {
			cleanCustomers = append(cleanCustomers, group[0])
			continue
		}
		
		// Multiple records with same normalized name - check for duplicates
		subGroups := c.analyzeGroup(group)
		
		for _, subGroup := range subGroups {
			if len(subGroup.Records) == 1 {
				cleanCustomers = append(cleanCustomers, subGroup.Records[0])
			} else {
				// Potential duplicates found
				duplicateGroups = append(duplicateGroups, subGroup)
				// Keep the first record, flag others as duplicates
				cleanCustomers = append(cleanCustomers, subGroup.Records[0])
			}
		}
	}
	
	return cleanCustomers, duplicateGroups
}

func (c *CustomerCleaner) analyzeGroup(records []CustomerRecord) []DuplicateGroup {
	var groups []DuplicateGroup
	processed := make(map[int]bool)
	
	for i, record1 := range records {
		if processed[i] {
			continue
		}
		
		group := DuplicateGroup{
			Records: []CustomerRecord{record1},
			Score:   1.0,
			Reason:  "exact_name_match",
		}
		processed[i] = true
		
		// Find similar records
		for j, record2 := range records {
			if i == j || processed[j] {
				continue
			}
			
			similarity := c.calculateSimilarity(record1, record2)
			if similarity >= c.rules.DuplicateThreshold {
				group.Records = append(group.Records, record2)
				group.Score = (group.Score + similarity) / 2 // Average similarity
				processed[j] = true
				
				// Determine reason
				if record1.AddressHash == record2.AddressHash {
					group.Reason = "same_name_and_address"
				} else if c.phoneMatch(record1, record2) {
					group.Reason = "same_name_and_phone"
				} else if c.emailMatch(record1, record2) {
					group.Reason = "same_name_and_email"
				}
			}
		}
		
		groups = append(groups, group)
	}
	
	return groups
}

func (c *CustomerCleaner) calculateSimilarity(r1, r2 CustomerRecord) float64 {
	score := 0.0
	factors := 0.0
	
	// Name similarity (most important)
	if r1.NormalizedName == r2.NormalizedName {
		score += 0.4
	}
	factors += 0.4
	
	// Address similarity
	if r1.AddressHash == r2.AddressHash {
		score += 0.3
	}
	factors += 0.3
	
	// Phone similarity
	if c.phoneMatch(r1, r2) {
		score += 0.2
	}
	factors += 0.2
	
	// Email similarity
	if c.emailMatch(r1, r2) {
		score += 0.1
	}
	factors += 0.1
	
	if factors == 0 {
		return 0
	}
	
	return score / factors
}

// Data cleaning methods
func (c *CustomerCleaner) cleanCustomerName(name string) string {
	if name == "" {
		return ""
	}
	
	// Normalize name
	normalized := strings.TrimSpace(name)
	normalized = strings.ReplaceAll(normalized, "  ", " ")
	
	// Check for known company variations
	lower := strings.ToLower(normalized)
	if standard, exists := c.rules.CompanyVariations[lower]; exists {
		return standard
	}
	
	// Standardize corporate suffixes
	suffixes := map[string]string{
		" corp":        " Corporation",
		" inc":         " Inc.",
		" llc":         " LLC",
		" lp":          " LP",
		" ltd":         " Ltd.",
		" co":          " Company",
		" company":     " Company",
		" corporation": " Corporation",
	}
	
	lowerName := strings.ToLower(normalized)
	for suffix, replacement := range suffixes {
		if strings.HasSuffix(lowerName, suffix) {
			return normalized[:len(normalized)-len(suffix)] + replacement
		}
	}
	
	return c.toTitleCase(normalized)
}

func (c *CustomerCleaner) normalizeCompanyName(name string) string {
	// Remove common words and punctuation for comparison
	normalized := strings.ToLower(name)
	
	// Remove corporate designations
	designations := []string{" corporation", " corp", " inc", " llc", " lp", " ltd", " company", " co"}
	for _, designation := range designations {
		normalized = strings.ReplaceAll(normalized, designation, "")
	}
	
	// Remove punctuation
	reg := regexp.MustCompile(`[^\w\s]`)
	normalized = reg.ReplaceAllString(normalized, "")
	
	// Remove extra spaces
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	
	return strings.TrimSpace(normalized)
}

func (c *CustomerCleaner) generateAddressHash(customer *CustomerRecord) string {
	if customer.BillingAddress == nil && customer.BillingCity == nil && customer.BillingState == nil {
		return ""
	}
	
	address := ""
	if customer.BillingAddress != nil {
		address += *customer.BillingAddress
	}
	if customer.BillingCity != nil {
		address += *customer.BillingCity
	}
	if customer.BillingState != nil {
		address += *customer.BillingState
	}
	
	// Normalize and hash
	normalized := strings.ToLower(strings.ReplaceAll(address, " ", ""))
	hash := fmt.Sprintf("%x", md5.Sum([]byte(normalized)))
	return hash[:8] // First 8 characters
}

func (c *CustomerCleaner) phoneMatch(r1, r2 CustomerRecord) bool {
	if r1.PhoneNumber == nil || r2.PhoneNumber == nil {
		return false
	}
	
	// Extract digits only
	phone1 := regexp.MustCompile(`\D`).ReplaceAllString(*r1.PhoneNumber, "")
	phone2 := regexp.MustCompile(`\D`).ReplaceAllString(*r2.PhoneNumber, "")
	
	return phone1 != "" && phone1 == phone2
}

func (c *CustomerCleaner) emailMatch(r1, r2 CustomerRecord) bool {
	if r1.EmailAddress == nil || r2.EmailAddress == nil {
		return false
	}
	return strings.ToLower(*r1.EmailAddress) == strings.ToLower(*r2.EmailAddress)
}

// Additional cleaning methods
func (c *CustomerCleaner) cleanState(state string) *string {
	if state == "" {
		return nil
	}
	
	state = strings.ToUpper(strings.TrimSpace(state))
	
	// Check if it's already a valid state code
	if len(state) == 2 {
		return &state
	}
	
	// Convert full name to code
	lower := strings.ToLower(state)
	if code, exists := c.rules.StateNormalizations[lower]; exists {
		return &code
	}
	
	return &state // Return as-is if we can't normalize
}

func (c *CustomerCleaner) cleanZipCode(zip string) *string {
	if zip == "" {
		return nil
	}
	
	// Remove non-alphanumeric characters except hyphens
	cleaned := regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(zip, "")
	
	if cleaned == "" {
		return nil
	}
	
	return &cleaned
}

func (c *CustomerCleaner) cleanPhoneNumber(phone string) *string {
	if phone == "" {
		return nil
	}
	
	// Extract digits only
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	
	// Format US phone numbers
	if len(digits) == 10 {
		formatted := fmt.Sprintf("(%s) %s-%s", digits[:3], digits[3:6], digits[6:])
		return &formatted
	} else if len(digits) == 11 && digits[0] == '1' {
		formatted := fmt.Sprintf("1 (%s) %s-%s", digits[1:4], digits[4:7], digits[7:])
		return &formatted
	}
	
	return &phone // Return original if not standard format
}

func (c *CustomerCleaner) cleanEmail(email string) *string {
	if email == "" {
		return nil
	}
	
	email = strings.ToLower(strings.TrimSpace(email))
	
	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(email) {
		return &email
	}
	
	return nil // Invalid email
}

func (c *CustomerCleaner) cleanPersonName(name string) *string {
	if name == "" {
		return nil
	}
	
	cleaned := c.toTitleCase(strings.TrimSpace(name))
	return &cleaned
}

func (c *CustomerCleaner) cleanPercentage(value string) *string {
	if value == "" {
		return nil
	}
	
	// Remove % sign and clean
	cleaned := strings.TrimSpace(strings.ReplaceAll(value, "%", ""))
	
	// Try to parse as float
	if percentage, err := strconv.ParseFloat(cleaned, 64); err == nil {
		if percentage >= 0 && percentage <= 100 {
			formatted := fmt.Sprintf("%.2f", percentage)
			return &formatted
		}
	}
	
	return nil // Invalid percentage
}

// Helper methods
func (c *CustomerCleaner) getStringField(record []string, columnMap map[string]int, fieldName string) string {
	if idx, exists := columnMap[fieldName]; exists && idx < len(record) {
		value := strings.TrimSpace(record[idx])
		if value == "" || value == "NULL" || value == "null" {
			return ""
		}
		return value
	}
	return ""
}

func (c *CustomerCleaner) getCleanStringField(record []string, columnMap map[string]int, fieldName string) *string {
	value := c.getStringField(record, columnMap, fieldName)
	if value == "" {
		return nil
	}
	return &value
}

func (c *CustomerCleaner) getIntField(record []string, columnMap map[string]int, fieldName string) *int {
	value := c.getStringField(record, columnMap, fieldName)
	if value == "" {
		return nil
	}
	
	if intVal, err := strconv.Atoi(value); err == nil {
		return &intVal
	}
	
	return nil
}

func (c *CustomerCleaner) toTitleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// Output functions
func getAddressSummary(customer CustomerRecord) string {
	parts := []string{}
	if customer.BillingCity != nil {
		parts = append(parts, *customer.BillingCity)
	}
	if customer.BillingState != nil {
		parts = append(parts, *customer.BillingState)
	}
	if len(parts) == 0 {
		return "No address"
	}
	return strings.Join(parts, ", ")
}

func getTotalDuplicates(duplicates []DuplicateGroup) int {
	total := 0
	for _, group := range duplicates {
		total += len(group.Records) - 1 // Don't count the kept record
	}
	return total
}

func writeClearedData(filename string, customers []CustomerRecord) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write header with standardized names (no abbreviations)
	header := []string{
		"customer_id", "customer_name", "billing_address", "billing_city", "billing_state", 
		"billing_zip_code", "contact_name", "phone_number", "fax_number", "email_address",
		"color_grade_1", "color_grade_2", "color_grade_3", "color_grade_4", "color_grade_5",
		"wall_loss_1", "wall_loss_2", "wall_loss_3", "wall_loss_4", "wall_loss_5",
		"wstring_color_1", "wstring_color_2", "wstring_color_3", "wstring_color_4", "wstring_color_5",
		"wstring_loss_1", "wstring_loss_2", "wstring_loss_3", "wstring_loss_4", "wstring_loss_5",
		"is_deleted", "tenant_id",
	}
	
	if err := writer.Write(header); err != nil {
		return err
	}
	
	// Write data
	for _, customer := range customers {
		record := []string{
			intToString(customer.CustomerID),
			customer.CustomerName,
			stringPtrToString(customer.BillingAddress),
			stringPtrToString(customer.BillingCity),
			stringPtrToString(customer.BillingState),
			stringPtrToString(customer.BillingZipCode),
			stringPtrToString(customer.ContactName),
			stringPtrToString(customer.PhoneNumber),
			stringPtrToString(customer.FaxNumber),
			stringPtrToString(customer.EmailAddress),
			stringPtrToString(customer.ColorGrade1),
			stringPtrToString(customer.ColorGrade2),
			stringPtrToString(customer.ColorGrade3),
			stringPtrToString(customer.ColorGrade4),
			stringPtrToString(customer.ColorGrade5),
			stringPtrToString(customer.WallLoss1),
			stringPtrToString(customer.WallLoss2),
			stringPtrToString(customer.WallLoss3),
			stringPtrToString(customer.WallLoss4),
			stringPtrToString(customer.WallLoss5),
			stringPtrToString(customer.WStringColor1),
			stringPtrToString(customer.WStringColor2),
			stringPtrToString(customer.WStringColor3),
			stringPtrToString(customer.WStringColor4),
			stringPtrToString(customer.WStringColor5),
			stringPtrToString(customer.WStringLoss1),
			stringPtrToString(customer.WStringLoss2),
			stringPtrToString(customer.WStringLoss3),
			stringPtrToString(customer.WStringLoss4),
			stringPtrToString(customer.WStringLoss5),
			boolToString(customer.IsDeleted),
			customer.TenantID,
		}
		
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	
	return nil
}

func intToString(i *int) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(*i)
}

func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

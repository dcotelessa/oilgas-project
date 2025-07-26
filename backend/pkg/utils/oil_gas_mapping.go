package utils

import (
	"regexp"
	"strings"
)

// Oil & Gas industry column mappings for PostgreSQL compatibility
// Extracted from proven Phase 1 logic
var OilGasColumnMapping = map[string]string{
	// Customer fields
	"custid":           "customer_id",
	"customerid":       "customer_id", 
	"custname":         "customer",
	"customername":     "customer",
	"customerpo":       "customer_po",
	"custpo":           "customer_po",
	// Work order fields
	"wkorder":          "work_order",
	"workorder":        "work_order",
	"wo":               "work_order",
	"rnumber":          "r_number",
	"rnum":             "r_number",
	// Date fields
	"datein":           "date_in",
	"dateout":          "date_out",
	"datereceived":     "date_received",
	"daterecvd":        "date_received",
	// Location fields
	"wellin":           "well_in",
	"wellout":          "well_out", 
	"leasein":          "lease_in",
	"leaseout":         "lease_out",
	// Address fields
	"billaddr":         "billing_address",
	"billcity":         "billing_city",
	"billstate":        "billing_state",
	"billzip":          "billing_zipcode",
	"billtoid":         "bill_to_id",
	// Contact fields
	"phoneno":          "phone",
	"phonenum":         "phone",
	"emailaddr":        "email",
	// Technical fields
	"wstring":          "w_string",
	"sizeid":           "size_id",
	"conntype":         "connection",
	"conn":             "connection",
	"locationcode":     "location",
	"loc":              "location",
	// Audit fields
	"orderedby":        "ordered_by",
	"enteredby":        "entered_by",
	"whenentered":      "when_entered",
	"when1":            "when_entered",
	"whenupdated":      "when_updated",
	"when2":            "when_updated",
	"updatedby":        "updated_by",
	"inspectedby":      "inspected_by",
	"inspecteddate":    "inspected_date",
	"inspected":        "inspected_date",
	"threadingdate":    "threading_date",
	"threading":        "threading_date",
	// Boolean fields
	"straightenreq":    "straighten_required",
	"straighten":       "straighten_required",
	"excessmat":        "excess_material",
	"excess":           "excess_material",
	"inproduction":     "in_production",
	"isdeleted":        "deleted",
	"createdat":        "created_at",
	"complete":         "complete",
}

// NormalizeOilGasColumn applies proven Phase 1 normalization logic
func NormalizeOilGasColumn(colName string) string {
	if colName == "" {
		return colName
	}

	// Basic normalization
	normalized := strings.ToLower(strings.TrimSpace(colName))
	normalized = strings.ReplaceAll(normalized, `"`, "")
	normalized = strings.ReplaceAll(normalized, `'`, "")

	// Convert non-alphanumeric to underscores
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	normalized = reg.ReplaceAllString(normalized, "_")
	
	// Collapse multiple underscores
	reg = regexp.MustCompile(`_+`)
	normalized = reg.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")

	// Apply industry-specific mappings
	if mapped, exists := OilGasColumnMapping[normalized]; exists {
		return mapped
	}

	// Convert common ID patterns
	if strings.HasSuffix(normalized, "id") && !strings.HasSuffix(normalized, "_id") {
		return normalized[:len(normalized)-2] + "_id"
	}

	return normalized
}

// ValidateOilGasColumnMapping returns statistics about mapping coverage
func ValidateOilGasColumnMapping(columns []string) (mapped, unmapped int, mappings map[string]string) {
	mappings = make(map[string]string)
	
	for _, col := range columns {
		normalized := NormalizeOilGasColumn(col)
		original := strings.ToLower(strings.TrimSpace(col))
		
		if original != normalized {
			mapped++
			mappings[col] = normalized
		} else {
			unmapped++
		}
	}
	
	return mapped, unmapped, mappings
}

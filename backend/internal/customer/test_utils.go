// backend/internal/customer/test_utils.go
// Shared test utilities to avoid duplicate declarations
package customer

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}

// Helper function for creating bool pointers
func boolPtr(b bool) *bool {
	return &b
}

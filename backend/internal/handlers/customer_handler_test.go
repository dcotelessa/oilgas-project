package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/handlers"
)

func TestCustomerHandler_GetCustomers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	handler := handlers.NewCustomerHandler()
	
	router := gin.New()
	router.GET("/customers", func(c *gin.Context) {
		c.Set("tenant_id", "testco")
		handler.GetCustomers(c)
	})
	
	req, _ := http.NewRequest("GET", "/customers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	
	// Should contain mock data
	body := w.Body.String()
	if !contains(body, "customers") {
		t.Error("Response should contain customers")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInner(s[1:len(s)-1], substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}


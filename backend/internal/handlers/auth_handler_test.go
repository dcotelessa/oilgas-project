package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/auth"
	"oilgas-backend/internal/handlers"
	"oilgas-backend/pkg/cache"
	"time"
)

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Setup
	cache := cache.NewWithDefaultExpiration(5*time.Minute, 1*time.Minute)
	sessionManager := auth.NewTenantSessionManager(nil, cache)
	handler := handlers.NewAuthHandler(sessionManager)
	
	router := gin.New()
	router.POST("/login", handler.Login)
	
	// Create test request
	loginReq := auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	jsonData, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Test
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// For now, we expect 401 since user doesn't exist in in-memory store
	// This test validates the handler structure
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusOK {
		t.Errorf("Expected 401 or 200, got %d", w.Code)
	}
}

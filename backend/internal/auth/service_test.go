package auth_test

import (
	"context"
	"testing"
	
	"oilgas-backend/internal/auth"
)

func TestService_CreateTenant(t *testing.T) {
	ctx := context.Background()
	service := auth.NewService()
	
	// Create tenant first
	tenant, err := service.CreateTenant(ctx, "Test Company", "testco")
	if err != nil {
		t.Fatalf("Expected no error creating tenant, got %v", err)
	}
	
	if tenant.Name != "Test Company" {
		t.Errorf("Expected name 'Test Company', got %s", tenant.Name)
	}
	
	if tenant.Slug != "testco" {
		t.Errorf("Expected slug 'testco', got %s", tenant.Slug)
	}
}

func TestService_CreateUser(t *testing.T) {
	ctx := context.Background()
	service := auth.NewService()
	
	// Create tenant first
	_, err := service.CreateTenant(ctx, "Test Company", "testco")
	if err != nil {
		t.Fatalf("Expected no error creating tenant, got %v", err)
	}
	
	// Create user
	req := &auth.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
		Company:  "Test Company",
		TenantID: "testco",
	}
	
	user, err := service.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error creating user, got %v", err)
	}
	
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
	
	if user.Role != "user" {
		t.Errorf("Expected role user, got %s", user.Role)
	}
	
	if user.TenantID != "testco" {
		t.Errorf("Expected tenant testco, got %s", user.TenantID)
	}
}

func TestService_CreateUser_Duplicate(t *testing.T) {
	ctx := context.Background()
	service := auth.NewService()
	
	// Create tenant
	_, err := service.CreateTenant(ctx, "Test Company", "testco")
	if err != nil {
		t.Fatalf("Expected no error creating tenant, got %v", err)
	}
	
	// Create first user
	req := &auth.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
		Company:  "Test Company",
		TenantID: "testco",
	}
	
	_, err = service.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error on first create, got %v", err)
	}
	
	// Try to create duplicate
	_, err = service.CreateUser(ctx, req)
	if err != auth.ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}

func TestService_Login(t *testing.T) {
	ctx := context.Background()
	service := auth.NewService()
	
	// Setup
	_, err := service.CreateTenant(ctx, "Test Company", "testco")
	if err != nil {
		t.Fatalf("Expected no error creating tenant, got %v", err)
	}
	
	userReq := &auth.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
		Company:  "Test Company",
		TenantID: "testco",
	}
	
	_, err = service.CreateUser(ctx, userReq)
	if err != nil {
		t.Fatalf("Expected no error creating user, got %v", err)
	}
	
	// Test login
	loginReq := &auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	response, err := service.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Expected no error on login, got %v", err)
	}
	
	if response.User.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", response.User.Email)
	}
	
	if response.SessionID == "" {
		t.Error("Expected non-empty session ID")
	}
	
	if response.Tenant.Slug != "testco" {
		t.Errorf("Expected tenant testco, got %s", response.Tenant.Slug)
	}
}

func TestService_Login_InvalidCredentials(t *testing.T) {
	ctx := context.Background()
	service := auth.NewService()
	
	loginReq := &auth.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}
	
	_, err := service.Login(ctx, loginReq)
	if err != auth.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_ValidateSession(t *testing.T) {
	ctx := context.Background()
	service := auth.NewService()
	
	// Setup tenant and user
	_, err := service.CreateTenant(ctx, "Test Company", "testco")
	if err != nil {
		t.Fatalf("Expected no error creating tenant, got %v", err)
	}
	
	userReq := &auth.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
		Company:  "Test Company",
		TenantID: "testco",
	}
	
	_, err = service.CreateUser(ctx, userReq)
	if err != nil {
		t.Fatalf("Expected no error creating user, got %v", err)
	}
	
	// Login to get session
	loginReq := &auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	response, err := service.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Expected no error on login, got %v", err)
	}
	
	// Validate session
	user, tenant, err := service.ValidateSession(ctx, response.SessionID)
	if err != nil {
		t.Fatalf("Expected no error validating session, got %v", err)
	}
	
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
	
	if tenant.Slug != "testco" {
		t.Errorf("Expected tenant testco, got %s", tenant.Slug)
	}
}

func TestService_ValidateSession_Invalid(t *testing.T) {
	ctx := context.Background()
	service := auth.NewService()
	
	_, _, err := service.ValidateSession(ctx, "invalid-session")
	if err != auth.ErrSessionExpired {
		t.Errorf("Expected ErrSessionExpired, got %v", err)
	}
}

func TestValidatePassword(t *testing.T) {
	service := auth.NewService()
	
	tests := []struct {
		password string
		valid    bool
	}{
		{"password123", true},
		{"short", false},
		{"", false},
		{"verylongpasswordthatshouldbefine", true},
	}
	
	for _, test := range tests {
		err := service.ValidatePassword(test.password)
		isValid := err == nil
		
		if isValid != test.valid {
			t.Errorf("Password %q: expected valid=%v, got valid=%v (err: %v)", 
				test.password, test.valid, isValid, err)
		}
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error hashing password, got %v", err)
	}
	
	if hash == "" {
		t.Error("Expected non-empty hash")
	}
	
	if hash == password {
		t.Error("Hash should not equal original password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"
	
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error hashing password, got %v", err)
	}
	
	// Test correct password
	if !auth.CheckPassword(password, hash) {
		t.Error("Expected password to match hash")
	}
	
	// Test wrong password  
	if auth.CheckPassword(wrongPassword, hash) {
		t.Error("Expected wrong password to not match hash")
	}
}

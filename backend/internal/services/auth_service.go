// backend/internal/services/auth_service.go
package services

import (
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
)

type AuthService struct {
	authRepo *repository.AuthRepository
}

func NewAuthService(authRepo *repository.AuthRepository) *AuthService {
	return &AuthService{
		authRepo: authRepo,
	}
}

func (s *AuthService) AuthenticateUser(email, password string) (*models.User, []models.Tenant, string, error) {
	// Get user by email
	user, err := s.authRepo.GetUserByEmail(email)
	if err != nil {
		return nil, nil, "", fmt.Errorf("invalid credentials")
	}
	
	// TODO: Validate password (bcrypt)
	// For now, just check if user exists
	
	// Note: GetUserTenants expects int but user.ID is UUID
	// For now, we'll use a simplified approach or update the repository interface
	// This is a TODO for when we implement proper user-tenant relationships
	
	// Mock tenants for now (until we update the repository to handle UUID)
	tenants := []models.Tenant{
		{
			ID:   uuid.New(),
			Name: "Default Tenant",
			Slug: "default",
			Active: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	// Create session (simplified)
	sessionID := fmt.Sprintf("session_%s_%d", user.ID.String(), time.Now().Unix())
	
	return user, tenants, sessionID, nil
}

func (s *AuthService) ValidateSession(sessionID string) (*models.User, error) {
	// TODO: Implement proper session validation
	// For now, return success if session exists
	if sessionID == "" {
		return nil, fmt.Errorf("invalid session")
	}
	
	// Mock user for demo (with proper UUID)
	mockUser := &models.User{
		ID:    uuid.New(),
		Email: "demo@example.com",
		Active: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	return mockUser, nil
}

// GetUserTenants - Updated to handle UUID properly
func (s *AuthService) GetUserTenants(userID uuid.UUID) ([]models.Tenant, error) {
	// TODO: Update repository method to accept UUID instead of int
	// For now, return empty slice or mock data
	return []models.Tenant{}, nil
}

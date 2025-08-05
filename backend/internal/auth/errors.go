// backend/internal/auth/errors.go
package auth

import "errors"

// Authentication errors
var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrUserInactive       = errors.New("user account is inactive")
)

// Authorization errors  
var (
	ErrPermissionDenied      = errors.New("permission denied")
	ErrTenantAccessDenied    = errors.New("tenant access denied")
	ErrYardAccessDenied      = errors.New("yard access denied")
	ErrEnterpriseAccessDenied = errors.New("enterprise access required")
	ErrNotCustomerContact    = errors.New("user is not a customer contact")
)

// Validation errors
var (
	ErrInvalidTenant      = errors.New("invalid tenant")
	ErrInvalidUserRole    = errors.New("invalid user role")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrCustomerNotFound   = errors.New("customer not found")
)

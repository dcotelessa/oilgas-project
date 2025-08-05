// backend/internal/auth/errors.go
package auth

import "errors"

// Authentication and authorization errors
var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserInactive        = errors.New("user account is inactive")
	ErrTokenExpired        = errors.New("token has expired")
	ErrTokenInvalid        = errors.New("token is invalid")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session has expired")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrYardAccessDenied    = errors.New("yard access denied")
	ErrTenantAccessDenied  = errors.New("tenant access denied")
	ErrEnterpriseAccessDenied = errors.New("enterprise access denied")
	ErrInvalidTenant       = errors.New("invalid tenant")
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrInvalidRole         = errors.New("invalid role")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrPasswordTooWeak     = errors.New("password is too weak")
	ErrEmailInvalid        = errors.New("email address is invalid")
	ErrUsernameInvalid     = errors.New("username is invalid")
)

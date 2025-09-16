// backend/internal/auth/errors.go
package auth

import "errors"

// Authentication errors
var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserExists          = errors.New("user already exists")
	ErrTenantAccessDenied  = errors.New("tenant access denied")
	ErrYardAccessDenied    = errors.New("yard access denied")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrSessionExpired      = errors.New("session expired")
	ErrInvalidSession      = errors.New("invalid session")
	ErrPermissionDenied    = errors.New("permission denied")
)

// backend/internal/repository/auth_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"oilgas-backend/internal/models"
)

type AuthRepository struct {
	authDB *sql.DB
}

func NewAuthRepository(authDB *sql.DB) *AuthRepository {
	return &AuthRepository{authDB: authDB}
}

func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, username, first_name, last_name, active, created_at 
		FROM users 
		WHERE email = $1 AND active = true`
	
	user := &models.User{}
	err := r.authDB.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Username, 
		&user.FirstName, &user.LastName, &user.Active, &user.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	
	return user, err
}

// Updated to accept UUID instead of int
func (r *AuthRepository) GetUserTenants(userID uuid.UUID) ([]models.Tenant, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.database_name, t.active, t.created_at
		FROM tenants t
		JOIN user_tenants ut ON t.id = ut.tenant_id
		WHERE ut.user_id = $1 AND ut.active = true AND t.active = true
		ORDER BY t.name`
	
	rows, err := r.authDB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tenants []models.Tenant
	for rows.Next() {
		var tenant models.Tenant
		err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, 
			&tenant.DatabaseName, &tenant.Active, &tenant.CreatedAt)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

func (r *AuthRepository) GetTenantBySlug(slug string) (*models.Tenant, error) {
	query := `SELECT id, name, slug, database_name, active, created_at 
			  FROM tenants WHERE slug = $1 AND active = true`
	
	tenant := &models.Tenant{}
	err := r.authDB.QueryRow(query, slug).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, 
		&tenant.DatabaseName, &tenant.Active, &tenant.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	
	return tenant, err
}

func (r *AuthRepository) ListTenants() ([]models.Tenant, error) {
	query := `
		SELECT id, name, slug, database_type, database_name, active, created_at, updated_at
		FROM tenants 
		WHERE active = true
		ORDER BY name`
	
	rows, err := r.authDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tenants []models.Tenant
	for rows.Next() {
		var tenant models.Tenant
		err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug,
			&tenant.DatabaseType, &tenant.DatabaseName, &tenant.Active,
			&tenant.CreatedAt, &tenant.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

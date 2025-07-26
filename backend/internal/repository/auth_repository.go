// backend/internal/repository/auth_repository.go
package repository

import (
	"database/sql"
	"fmt"
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

func (r *AuthRepository) GetUserTenants(userID int) ([]models.Tenant, error) {
	query := `
		SELECT t.id, t.code, t.name, t.database_name, t.active, t.created_at
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
		err := rows.Scan(&tenant.ID, &tenant.Code, &tenant.Name, 
			&tenant.DatabaseName, &tenant.Active, &tenant.CreatedAt)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

func (r *AuthRepository) GetTenantByCode(code string) (*models.Tenant, error) {
	query := `SELECT id, code, name, database_name, active, created_at 
			  FROM tenants WHERE code = $1 AND active = true`
	
	tenant := &models.Tenant{}
	err := r.authDB.QueryRow(query, code).Scan(
		&tenant.ID, &tenant.Code, &tenant.Name, 
		&tenant.DatabaseName, &tenant.Active, &tenant.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	
	return tenant, err
}

package repository

import (
    "database/sql"
    "fmt"
    
    "github.com/dcotelessa/oil-gas-inventory/internal/models"
)

type TenantRepository struct {
    db *sql.DB
}

func NewTenantRepository(db *sql.DB) *TenantRepository {
    return &TenantRepository{db: db}
}

func (r *TenantRepository) CreateTenant(tenant *models.Tenant) error {
    query := `
        INSERT INTO store.tenants (tenant_name, tenant_slug, description, contact_email, phone, address, settings)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING tenant_id, created_at, updated_at
    `
    err := r.db.QueryRow(
        query, 
        tenant.TenantName, 
        tenant.TenantSlug, 
        tenant.Description,
        tenant.ContactEmail,
        tenant.Phone,
        tenant.Address,
        tenant.Settings,
    ).Scan(&tenant.TenantID, &tenant.CreatedAt, &tenant.UpdatedAt)
    
    return err
}

func (r *TenantRepository) GetTenantByID(tenantID int) (*models.Tenant, error) {
    tenant := &models.Tenant{}
    query := `
        SELECT tenant_id, tenant_name, tenant_slug, description, contact_email, 
               phone, address, active, settings, created_at, updated_at
        FROM store.tenants 
        WHERE tenant_id = $1
    `
    err := r.db.QueryRow(query, tenantID).Scan(
        &tenant.TenantID,
        &tenant.TenantName,
        &tenant.TenantSlug,
        &tenant.Description,
        &tenant.ContactEmail,
        &tenant.Phone,
        &tenant.Address,
        &tenant.Active,
        &tenant.Settings,
        &tenant.CreatedAt,
        &tenant.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("tenant not found")
    }
    
    return tenant, err
}

func (r *TenantRepository) GetTenantBySlug(slug string) (*models.Tenant, error) {
    tenant := &models.Tenant{}
    query := `
        SELECT tenant_id, tenant_name, tenant_slug, description, contact_email,
               phone, address, active, settings, created_at, updated_at
        FROM store.tenants 
        WHERE tenant_slug = $1 AND active = true
    `
    err := r.db.QueryRow(query, slug).Scan(
        &tenant.TenantID,
        &tenant.TenantName,
        &tenant.TenantSlug,
        &tenant.Description,
        &tenant.ContactEmail,
        &tenant.Phone,
        &tenant.Address,
        &tenant.Active,
        &tenant.Settings,
        &tenant.CreatedAt,
        &tenant.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("tenant not found")
    }
    
    return tenant, err
}

func (r *TenantRepository) ListTenants() ([]*models.Tenant, error) {
    query := `
        SELECT tenant_id, tenant_name, tenant_slug, description, contact_email,
               phone, address, active, settings, created_at, updated_at
        FROM store.tenants 
        ORDER BY tenant_name
    `
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tenants []*models.Tenant
    for rows.Next() {
        tenant := &models.Tenant{}
        err := rows.Scan(
            &tenant.TenantID,
            &tenant.TenantName,
            &tenant.TenantSlug,
            &tenant.Description,
            &tenant.ContactEmail,
            &tenant.Phone,
            &tenant.Address,
            &tenant.Active,
            &tenant.Settings,
            &tenant.CreatedAt,
            &tenant.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        tenants = append(tenants, tenant)
    }

    return tenants, nil
}

func (r *TenantRepository) AssignCustomerToTenant(customerID, tenantID int, relationshipType string, assignedBy int) error {
    query := `
        INSERT INTO store.customer_tenant_assignments (customer_id, tenant_id, relationship_type, assigned_by)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (customer_id, tenant_id) 
        DO UPDATE SET 
            relationship_type = EXCLUDED.relationship_type,
            assigned_by = EXCLUDED.assigned_by,
            assigned_at = CURRENT_TIMESTAMP,
            active = true
    `
    _, err := r.db.Exec(query, customerID, tenantID, relationshipType, assignedBy)
    return err
}

func (r *TenantRepository) GetTenantCustomers(tenantID int) ([]*models.CustomerTenantAssignment, error) {
    query := `
        SELECT cta.id, cta.customer_id, cta.tenant_id, cta.relationship_type,
               cta.active, cta.assigned_at, cta.assigned_by, cta.notes
        FROM store.customer_tenant_assignments cta
        WHERE cta.tenant_id = $1 AND cta.active = true
        ORDER BY cta.assigned_at DESC
    `
    rows, err := r.db.Query(query, tenantID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var assignments []*models.CustomerTenantAssignment
    for rows.Next() {
        assignment := &models.CustomerTenantAssignment{}
        err := rows.Scan(
            &assignment.ID,
            &assignment.CustomerID,
            &assignment.TenantID,
            &assignment.RelationshipType,
            &assignment.Active,
            &assignment.AssignedAt,
            &assignment.AssignedBy,
            &assignment.Notes,
        )
        if err != nil {
            return nil, err
        }
        assignments = append(assignments, assignment)
    }

    return assignments, nil
}

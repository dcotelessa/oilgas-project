package tenant_test

import (
    "testing"
    
    "github.com/dcotelessa/oilgas-project/internal/models"
    "github.com/dcotelessa/oilgas-project/internal/repository"
    "github.com/dcotelessa/oilgas-project/internal/services"
)

func TestTenantService_CreateTenant(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    tenantRepo := repository.NewTenantRepository(db)
    tenantService := services.NewTenantService(tenantRepo)
    
    tests := []struct {
        name    string
        tenant  *models.Tenant
        wantErr bool
    }{
        {
            name: "Valid tenant",
            tenant: &models.Tenant{
                TenantName:  "Test Division",
                TenantSlug:  "test-division",
                Description: stringPtr("Test division for unit tests"),
            },
            wantErr: false,
        },
        {
            name: "Empty name",
            tenant: &models.Tenant{
                TenantName: "",
                TenantSlug: "empty-name",
            },
            wantErr: true,
        },
        {
            name: "Auto-generate slug",
            tenant: &models.Tenant{
                TenantName: "Auto Slug Test",
                // TenantSlug will be auto-generated
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tenantService.CreateTenant(tt.tenant)
            
            if tt.wantErr && err == nil {
                t.Errorf("Expected error, got nil")
            }
            
            if !tt.wantErr && err != nil {
                t.Errorf("Expected no error, got: %v", err)
            }
            
            if !tt.wantErr && err == nil {
                // Verify tenant was created
                if tt.tenant.TenantID == 0 {
                    t.Errorf("Expected tenant ID to be set")
                }
                
                if tt.tenant.TenantSlug == "" {
                    t.Errorf("Expected tenant slug to be set")
                }
            }
        })
    }
}

func TestTenantService_AssignCustomerToTenant(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    tenantRepo := repository.NewTenantRepository(db)
    tenantService := services.NewTenantService(tenantRepo)
    
    tests := []struct {
        name             string
        customerID       int
        tenantID         int
        relationshipType string
        assignedBy       int
        wantErr          bool
    }{
        {
            name:             "Valid assignment",
            customerID:       1,
            tenantID:         1,
            relationshipType: "primary",
            assignedBy:       1,
            wantErr:          false,
        },
        {
            name:             "Invalid relationship type",
            customerID:       1,
            tenantID:         1,
            relationshipType: "invalid",
            assignedBy:       1,
            wantErr:          true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tenantService.AssignCustomerToTenant(
                tt.customerID,
                tt.tenantID,
                tt.relationshipType,
                tt.assignedBy,
            )
            
            if tt.wantErr && err == nil {
                t.Errorf("Expected error, got nil")
            }
            
            if !tt.wantErr && err != nil {
                t.Errorf("Expected no error, got: %v", err)
            }
        })
    }
}

func stringPtr(s string) *string {
    return &s
}

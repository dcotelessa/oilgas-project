package services

import (
    "fmt"
    "regexp"
    "strings"
    
    "github.com/dcotelessa/oil-gas-inventory/internal/models"
    "github.com/dcotelessa/oil-gas-inventory/internal/repository"
)

type TenantService struct {
    tenantRepo *repository.TenantRepository
}

func NewTenantService(tenantRepo *repository.TenantRepository) *TenantService {
    return &TenantService{
        tenantRepo: tenantRepo,
    }
}

func (s *TenantService) CreateTenant(tenant *models.Tenant) error {
    // Validate tenant name
    if tenant.TenantName == "" {
        return fmt.Errorf("tenant name is required")
    }

    // Generate slug if not provided
    if tenant.TenantSlug == "" {
        tenant.TenantSlug = s.generateSlug(tenant.TenantName)
    }

    // Validate slug
    if err := s.validateSlug(tenant.TenantSlug); err != nil {
        return err
    }

    return s.tenantRepo.CreateTenant(tenant)
}

func (s *TenantService) GetTenant(tenantID int) (*models.Tenant, error) {
    return s.tenantRepo.GetTenantByID(tenantID)
}

func (s *TenantService) GetTenantBySlug(slug string) (*models.Tenant, error) {
    return s.tenantRepo.GetTenantBySlug(slug)
}

func (s *TenantService) ListTenants() ([]*models.Tenant, error) {
    return s.tenantRepo.ListTenants()
}

, slug)
    if !matched {
        return fmt.Errorf("tenant slug must contain only lowercase letters, numbers, and hyphens")
    }
    
    return nil
}

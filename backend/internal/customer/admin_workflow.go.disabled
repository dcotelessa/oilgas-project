// backend/internal/customer/admin_workflow.go
package customer

import (
    "context"
    "crypto/rand"
    "fmt"
    "strings"
    "time"
    
    "oilgas-backend/internal/auth"
)

// AdminWorkflowService handles admin operations for customer contact management
type AdminWorkflowService struct {
    customerRepo Repository
    authService  auth.Service
}

func NewAdminWorkflowService(customerRepo Repository, authService auth.Service) *AdminWorkflowService {
    return &AdminWorkflowService{
        customerRepo: customerRepo,
        authService:  authService,
    }
}

// RegisterCustomerContact allows admins to register new contacts for customers
func (aws *AdminWorkflowService) RegisterCustomerContact(ctx context.Context, tenantID string, req RegisterContactRequest) (*ContactRegistrationResponse, error) {
    // Validate admin permissions
    if err := aws.validateAdminAccess(ctx, tenantID); err != nil {
        return nil, fmt.Errorf("admin access validation failed: %w", err)
    }
    
    // Validate customer exists
    if err := aws.customerRepo.ValidateCustomerExists(ctx, tenantID, req.CustomerID); err != nil {
        return nil, fmt.Errorf("customer validation failed: %w", err)
    }
    
    // Generate temporary password if requested
    var tempPassword string
    if req.TemporaryPassword {
        var err error
        tempPassword, err = generateTemporaryPassword()
        if err != nil {
            return nil, fmt.Errorf("failed to generate temporary password: %w", err)
        }
    }
    
    // Create auth user for the contact
    authUser, err := aws.authService.CreateCustomerContact(ctx, auth.CreateCustomerContactRequest{
        Email:           req.Email,
        FullName:        req.FullName,
        TenantID:        tenantID,
        CustomerID:      req.CustomerID,
        Role:            auth.RoleCustomerContact,
        ContactType:     auth.ContactType(req.ContactType),
        TemporaryPassword: tempPassword,
        YardPermissions: req.YardPermissions,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create auth user: %w", err)
    }
    
    // Create customer-auth contact relationship
    contact := &CustomerAuthContact{
        CustomerID:      req.CustomerID,
        AuthUserID:      authUser.ID,
        ContactType:     req.ContactType,
        YardPermissions: req.YardPermissions,
        IsActive:        true,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }
    
    createdContact, err := aws.customerRepo.CreateCustomerContact(ctx, contact)
    if err != nil {
        // Rollback auth user creation if contact creation fails
        _ = aws.authService.DeleteUser(ctx, authUser.ID)
        return nil, fmt.Errorf("failed to create customer contact: %w", err)
    }
    
    return &ContactRegistrationResponse{
        ContactID:         createdContact.ID,
        UserID:            authUser.ID,
        Email:             req.Email,
        TemporaryPassword: tempPassword,
        Success:           true,
    }, nil
}

// BulkRegisterCustomerContacts allows admins to register multiple contacts at once
func (aws *AdminWorkflowService) BulkRegisterCustomerContacts(ctx context.Context, tenantID string, req BulkContactRegistrationRequest) ([]ContactRegistrationResponse, error) {
    // Validate admin permissions
    if err := aws.validateAdminAccess(ctx, tenantID); err != nil {
        return nil, fmt.Errorf("admin access validation failed: %w", err)
    }
    
    // Validate customer exists
    if err := aws.customerRepo.ValidateCustomerExists(ctx, tenantID, req.CustomerID); err != nil {
        return nil, fmt.Errorf("customer validation failed: %w", err)
    }
    
    // Validate request limits
    if len(req.Contacts) == 0 {
        return nil, fmt.Errorf("no contacts provided")
    }
    if len(req.Contacts) > 10 {
        return nil, fmt.Errorf("too many contacts in bulk request (max 10)")
    }
    
    responses := make([]ContactRegistrationResponse, len(req.Contacts))
    successCount := 0
    
    // Process each contact individually
    for i, contactReq := range req.Contacts {
        contactReq.CustomerID = req.CustomerID // Ensure consistency
        
        response, err := aws.RegisterCustomerContact(ctx, tenantID, contactReq)
        if err != nil {
            responses[i] = ContactRegistrationResponse{
                Email:   contactReq.Email,
                Success: false,
                Error:   err.Error(),
            }
        } else {
            responses[i] = *response
            successCount++
        }
    }
    
    return responses, nil
}

// UpdateContactPermissions allows admins to update yard permissions for contacts
func (aws *AdminWorkflowService) UpdateContactPermissions(ctx context.Context, tenantID string, customerID, userID int, permissions []string) error {
    // Validate admin permissions
    if err := aws.validateAdminAccess(ctx, tenantID); err != nil {
        return fmt.Errorf("admin access validation failed: %w", err)
    }
    
    // Update permissions in customer contact table
    if err := aws.customerRepo.UpdateContactPermissions(ctx, customerID, userID, permissions); err != nil {
        return fmt.Errorf("failed to update contact permissions: %w", err)
    }
    
    // Sync permissions to auth service
    if err := aws.authService.UpdateUserYardAccess(ctx, userID, tenantID, permissions); err != nil {
        return fmt.Errorf("failed to sync permissions to auth service: %w", err)
    }
    
    return nil
}

// ResetContactPassword allows admins to reset customer contact passwords
func (aws *AdminWorkflowService) ResetContactPassword(ctx context.Context, tenantID string, userID int) (string, error) {
    // Validate admin permissions
    if err := aws.validateAdminAccess(ctx, tenantID); err != nil {
        return "", fmt.Errorf("admin access validation failed: %w", err)
    }
    
    // Generate new temporary password
    tempPassword, err := generateTemporaryPassword()
    if err != nil {
        return "", fmt.Errorf("failed to generate temporary password: %w", err)
    }
    
    // Reset password in auth service
    if err := aws.authService.ResetUserPassword(ctx, userID, tempPassword, true); err != nil {
        return "", fmt.Errorf("failed to reset password: %w", err)
    }
    
    return tempPassword, nil
}

// DeactivateCustomerContact allows admins to deactivate customer contacts
func (aws *AdminWorkflowService) DeactivateCustomerContact(ctx context.Context, tenantID string, customerID, userID int) error {
    // Validate admin permissions
    if err := aws.validateAdminAccess(ctx, tenantID); err != nil {
        return fmt.Errorf("admin access validation failed: %w", err)
    }
    
    // Deactivate in customer contact table
    if err := aws.customerRepo.DeactivateCustomerContact(ctx, customerID, userID); err != nil {
        return fmt.Errorf("failed to deactivate customer contact: %w", err)
    }
    
    // Deactivate in auth service
    if err := aws.authService.DeactivateUser(ctx, userID); err != nil {
        return fmt.Errorf("failed to deactivate auth user: %w", err)
    }
    
    return nil
}

// validateAdminAccess checks if the current user has admin permissions
func (aws *AdminWorkflowService) validateAdminAccess(ctx context.Context, tenantID string) error {
    // Extract user from context (implementation depends on middleware)
    userID, ok := ctx.Value("user_id").(int)
    if !ok {
        return fmt.Errorf("user context not found")
    }
    
    // Check admin permissions through auth service
    hasPermission, err := aws.authService.CheckPermission(ctx, auth.UserPermissionCheck{
        UserID:     userID,
        TenantID:   tenantID,
        Permission: auth.PermissionManageCustomerContacts,
    })
    if err != nil {
        return fmt.Errorf("permission check failed: %w", err)
    }
    
    if !hasPermission {
        return fmt.Errorf("insufficient permissions for admin operation")
    }
    
    return nil
}

// generateTemporaryPassword creates a secure temporary password
func generateTemporaryPassword() (string, error) {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
    const length = 12
    
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    
    for i, b := range bytes {
        bytes[i] = charset[b%byte(len(charset))]
    }
    
    return string(bytes), nil
}

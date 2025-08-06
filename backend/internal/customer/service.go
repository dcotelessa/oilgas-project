// backend/internal/customer/service.go
type service struct {
    repo  Repository
    cache *cache.TenantAwareCache
}

func (s *service) GetCustomer(ctx context.Context, tenantID string, id int) (*Customer, error) {
    // Try cache first
    if customer, found := s.cache.GetCustomer(tenantID, id); found {
        return customer, nil
    }
    
    // Cache miss - fetch from database
    customer, err := s.repo.GetCustomerByID(ctx, tenantID, id)
    if err != nil {
        return nil, err
    }
    
    // Cache for future requests
    s.cache.CacheCustomer(tenantID, customer)
    
    return customer, nil
}

func (s *service) UpdateCustomer(ctx context.Context, customer *Customer) error {
    err := s.repo.UpdateCustomer(ctx, customer)
    if err != nil {
        return err
    }
    
    // Invalidate cache after update
    s.cache.InvalidateCustomer(customer.TenantID, customer.ID)
    
    return nil
}

// internal/services/customer_service.go
package services

import (
	"context"
	"fmt"
	"strings"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
)

type CustomerService struct {
	repo repository.CustomerRepository
}

func NewCustomerService(repo repository.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) GetAllCustomers(ctx context.Context) ([]models.Customer, error) {
	return s.repo.GetAll(ctx)
}

func (s *CustomerService) GetCustomerByID(ctx context.Context, id int) (*models.Customer, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid customer ID: %d", id)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *CustomerService) SearchCustomers(ctx context.Context, query string) ([]models.Customer, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}
	return s.repo.Search(ctx, query)
}

func (s *CustomerService) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return s.repo.Create(ctx, customer)
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, customer *models.Customer) error {
	if customer.CustomerID <= 0 {
		return fmt.Errorf("invalid customer ID: %d", customer.CustomerID)
	}
	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return s.repo.Update(ctx, customer)
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid customer ID: %d", id)
	}
	return s.repo.Delete(ctx, id)
}

func (s *CustomerService) validateCustomer(customer *models.Customer) error {
	if strings.TrimSpace(customer.Customer) == "" {
		return fmt.Errorf("customer name is required")
	}
	if len(customer.Customer) > 255 {
		return fmt.Errorf("customer name too long (max 255 characters)")
	}
	return nil
}

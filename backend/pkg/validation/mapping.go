// backend/pkg/validation/mapping.go
package validation

import (
	"time"

	"oilgas-backend/internal/models"
)

// ToInventoryModel converts InventoryValidation to models.InventoryItem
func (iv *InventoryValidation) ToInventoryModel() *models.InventoryItem {
	now := time.Now()
	
	return &models.InventoryItem{
		CustomerID: iv.CustomerID,
		Joints:     iv.Joints,
		Size:       NormalizeSize(iv.Size),
		Weight:     iv.Weight,
		Grade:      NormalizeGrade(iv.Grade),
		Connection: NormalizeConnection(iv.Connection),
		Color:      iv.Color,
		Location:   iv.Location,
		CTD:        false, // Default values
		WString:    false,
		Deleted:    false,
		DateIn:     &now,
		CreatedAt:  now,
	}
}

// FromInventoryModel converts models.InventoryItem to InventoryValidation
func FromInventoryModel(item *models.InventoryItem) *InventoryValidation {
	return &InventoryValidation{
		CustomerID: item.CustomerID,
		Joints:     item.Joints,
		Size:       item.Size,
		Weight:     item.Weight,
		Grade:      item.Grade,
		Connection: item.Connection,
		Color:      item.Color,
		Location:   item.Location,
	}
}

// ToCustomerModel converts CustomerValidation to models.Customer
func (cv *CustomerValidation) ToCustomerModel() *models.Customer {
	now := time.Now()
	
	return &models.Customer{
		Customer:       cv.Name,
		BillingAddress: cv.Address,
		BillingCity:    cv.City,
		BillingState:   cv.State,
		BillingZipcode: cv.Zipcode,
		Phone:          cv.Phone,
		Email:          cv.Email,
		Deleted:        false,
		CreatedAt:      now,
	}
}

// FromCustomerModel converts models.Customer to CustomerValidation
func FromCustomerModel(customer *models.Customer) *CustomerValidation {
	return &CustomerValidation{
		Name:    customer.Customer,
		Address: customer.BillingAddress,
		City:    customer.BillingCity,
		State:   customer.BillingState,
		Zipcode: customer.BillingZipcode,
		Phone:   customer.Phone,
		Email:   customer.Email,
	}
}

// UpdateCustomerValidation for additional validation fields not in basic CustomerValidation
type UpdateCustomerValidation struct {
	CustomerValidation
	Contact string `json:"contact,omitempty"`
	Fax     string `json:"fax,omitempty"`
}

// ToCustomerModel for UpdateCustomerValidation
func (ucv *UpdateCustomerValidation) ToCustomerModel() *models.Customer {
	base := ucv.CustomerValidation.ToCustomerModel()
	base.Contact = ucv.Contact
	base.Fax = ucv.Fax
	return base
}

// ExtendedInventoryValidation for additional fields not in basic validation
type ExtendedInventoryValidation struct {
	InventoryValidation
	Username   string `json:"username,omitempty"`
	WorkOrder  string `json:"work_order,omitempty"`
	RNumber    int    `json:"r_number,omitempty"`
	Rack       string `json:"rack,omitempty"`
	SWGCC      string `json:"swgcc,omitempty"`
	CustomerPO string `json:"customer_po,omitempty"`
	Fletcher   string `json:"fletcher,omitempty"`
	WellIn     string `json:"well_in,omitempty"`
	LeaseIn    string `json:"lease_in,omitempty"`
	WellOut    string `json:"well_out,omitempty"`
	LeaseOut   string `json:"lease_out,omitempty"`
	Trucking   string `json:"trucking,omitempty"`
	Trailer    string `json:"trailer,omitempty"`
	Notes      string `json:"notes,omitempty"`
	PCode      string `json:"pcode,omitempty"`
	CN         int    `json:"cn,omitempty"`
	OrderedBy  string `json:"ordered_by,omitempty"`
	CTD        bool   `json:"ctd,omitempty"`
	WString    bool   `json:"w_string,omitempty"`
}

// Validate extends basic validation for extended fields
func (eiv *ExtendedInventoryValidation) Validate() []ValidationError {
	// Start with basic validation
	errors := eiv.InventoryValidation.Validate()
	
	// Add extended validations
	if eiv.WorkOrder != "" && len(eiv.WorkOrder) > 50 {
		errors = append(errors, ValidationError{
			Field:   "work_order",
			Value:   eiv.WorkOrder,
			Message: "work order must be 50 characters or less",
		})
	}
	
	if eiv.CustomerPO != "" && len(eiv.CustomerPO) > 50 {
		errors = append(errors, ValidationError{
			Field:   "customer_po",
			Value:   eiv.CustomerPO,
			Message: "customer PO must be 50 characters or less",
		})
	}
	
	if eiv.Notes != "" && len(eiv.Notes) > 1000 {
		errors = append(errors, ValidationError{
			Field:   "notes",
			Value:   eiv.Notes,
			Message: "notes must be 1000 characters or less",
		})
	}
	
	return errors
}

// ToInventoryModel for ExtendedInventoryValidation
func (eiv *ExtendedInventoryValidation) ToInventoryModel() *models.InventoryItem {
	base := eiv.InventoryValidation.ToInventoryModel()
	
	// Add extended fields
	base.Username = eiv.Username
	base.WorkOrder = eiv.WorkOrder
	base.RNumber = eiv.RNumber
	base.Rack = eiv.Rack
	base.SWGCC = eiv.SWGCC
	base.CustomerPO = eiv.CustomerPO
	base.Fletcher = eiv.Fletcher
	base.WellIn = eiv.WellIn
	base.LeaseIn = eiv.LeaseIn
	base.WellOut = eiv.WellOut
	base.LeaseOut = eiv.LeaseOut
	base.Trucking = eiv.Trucking
	base.Trailer = eiv.Trailer
	base.Notes = eiv.Notes
	base.PCode = eiv.PCode
	base.CN = eiv.CN
	base.OrderedBy = eiv.OrderedBy
	base.CTD = eiv.CTD
	base.WString = eiv.WString
	
	return base
}

// FromExtendedInventoryModel converts models.InventoryItem to ExtendedInventoryValidation
func FromExtendedInventoryModel(item *models.InventoryItem) *ExtendedInventoryValidation {
	return &ExtendedInventoryValidation{
		InventoryValidation: InventoryValidation{
			CustomerID: item.CustomerID,
			Joints:     item.Joints,
			Size:       item.Size,
			Weight:     item.Weight,
			Grade:      item.Grade,
			Connection: item.Connection,
			Color:      item.Color,
			Location:   item.Location,
		},
		Username:   item.Username,
		WorkOrder:  item.WorkOrder,
		RNumber:    item.RNumber,
		Rack:       item.Rack,
		SWGCC:      item.SWGCC,
		CustomerPO: item.CustomerPO,
		Fletcher:   item.Fletcher,
		WellIn:     item.WellIn,
		LeaseIn:    item.LeaseIn,
		WellOut:    item.WellOut,
		LeaseOut:   item.LeaseOut,
		Trucking:   item.Trucking,
		Trailer:    item.Trailer,
		Notes:      item.Notes,
		PCode:      item.PCode,
		CN:         item.CN,
		OrderedBy:  item.OrderedBy,
		CTD:        item.CTD,
		WString:    item.WString,
	}
}

// Helper functions for common validation patterns

// ValidateWorkOrder validates work order format
func ValidateWorkOrder(workOrder string) error {
	if workOrder == "" {
		return nil // Optional field
	}
	
	if len(workOrder) > 50 {
		return ValidationError{
			Field:   "work_order",
			Value:   workOrder,
			Message: "work order must be 50 characters or less",
		}
	}
	
	return nil
}

// ValidateCustomerPO validates customer purchase order format
func ValidateCustomerPO(customerPO string) error {
	if customerPO == "" {
		return nil // Optional field
	}
	
	if len(customerPO) > 50 {
		return ValidationError{
			Field:   "customer_po",
			Value:   customerPO,
			Message: "customer PO must be 50 characters or less",
		}
	}
	
	return nil
}

// ValidateRack validates rack location format
func ValidateRack(rack string) error {
	if rack == "" {
		return nil // Optional field
	}
	
	if len(rack) > 50 {
		return ValidationError{
			Field:   "rack",
			Value:   rack,
			Message: "rack location must be 50 characters or less",
		}
	}
	
	return nil
}

// ValidateWellLease validates well and lease names
func ValidateWellLease(well, lease string) []ValidationError {
	var errors []ValidationError
	
	if well != "" && len(well) > 50 {
		errors = append(errors, ValidationError{
			Field:   "well",
			Value:   well,
			Message: "well name must be 50 characters or less",
		})
	}
	
	if lease != "" && len(lease) > 50 {
		errors = append(errors, ValidationError{
			Field:   "lease",
			Value:   lease,
			Message: "lease name must be 50 characters or less",
		})
	}
	
	return errors
}

// BatchInventoryValidation for bulk operations
type BatchInventoryValidation struct {
	Items []InventoryValidation `json:"items"`
}

// Validate validates all items in batch
func (biv *BatchInventoryValidation) Validate() map[int][]ValidationError {
	errors := make(map[int][]ValidationError)
	
	for i, item := range biv.Items {
		if itemErrors := item.Validate(); len(itemErrors) > 0 {
			errors[i] = itemErrors
		}
	}
	
	return errors
}

// ToInventoryModels converts batch to slice of models
func (biv *BatchInventoryValidation) ToInventoryModels() []*models.InventoryItem {
	var items []*models.InventoryItem
	
	for _, validation := range biv.Items {
		items = append(items, validation.ToInventoryModel())
	}
	
	return items
}

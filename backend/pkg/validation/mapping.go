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

// ToCustomerModel converts CustomerValidation to models.Customer (UPDATED for complete model)
func (cv *CustomerValidation) ToCustomerModel() *models.Customer {
	now := time.Now()
	
	return &models.Customer{
		Customer:       cv.CustomerName, // Fixed: was cv.Name, now cv.CustomerName
		BillingAddress: cv.Address,
		BillingCity:    cv.City,
		BillingState:   cv.State,
		BillingZipcode: cv.Zip,        // Fixed: was cv.Zipcode, now cv.Zip
		Contact:        cv.Contact,     // Added missing field
		Phone:          cv.Phone,
		Fax:            cv.Fax,         // Added missing field
		Email:          cv.Email,
		// Color fields
		Color1:         cv.Color1,
		Color2:         cv.Color2,
		Color3:         cv.Color3,
		Color4:         cv.Color4,
		Color5:         cv.Color5,
		// Loss fields
		Loss1:          cv.Loss1,
		Loss2:          cv.Loss2,
		Loss3:          cv.Loss3,
		Loss4:          cv.Loss4,
		Loss5:          cv.Loss5,
		// WS Color fields
		WSColor1:       cv.WSColor1,
		WSColor2:       cv.WSColor2,
		WSColor3:       cv.WSColor3,
		WSColor4:       cv.WSColor4,
		WSColor5:       cv.WSColor5,
		// WS Loss fields
		WSLoss1:        cv.WSLoss1,
		WSLoss2:        cv.WSLoss2,
		WSLoss3:        cv.WSLoss3,
		WSLoss4:        cv.WSLoss4,
		WSLoss5:        cv.WSLoss5,
		// System fields
		Deleted:        false,
		CreatedAt:      now,
	}
}

// FromCustomerModel converts models.Customer to CustomerValidation (UPDATED)
func FromCustomerModel(customer *models.Customer) *CustomerValidation {
	return &CustomerValidation{
		CustomerName: customer.Customer,       // Fixed: was Name, now CustomerName
		Address:      customer.BillingAddress,
		City:         customer.BillingCity,
		State:        customer.BillingState,
		Zip:          customer.BillingZipcode, // Fixed: was Zipcode, now Zip
		Contact:      customer.Contact,        // Added missing field
		Phone:        customer.Phone,
		Fax:          customer.Fax,            // Added missing field
		Email:        customer.Email,
		// Color fields
		Color1:       customer.Color1,
		Color2:       customer.Color2,
		Color3:       customer.Color3,
		Color4:       customer.Color4,
		Color5:       customer.Color5,
		// Loss fields
		Loss1:        customer.Loss1,
		Loss2:        customer.Loss2,
		Loss3:        customer.Loss3,
		Loss4:        customer.Loss4,
		Loss5:        customer.Loss5,
		// WS Color fields
		WSColor1:     customer.WSColor1,
		WSColor2:     customer.WSColor2,
		WSColor3:     customer.WSColor3,
		WSColor4:     customer.WSColor4,
		WSColor5:     customer.WSColor5,
		// WS Loss fields
		WSLoss1:      customer.WSLoss1,
		WSLoss2:      customer.WSLoss2,
		WSLoss3:      customer.WSLoss3,
		WSLoss4:      customer.WSLoss4,
		WSLoss5:      customer.WSLoss5,
	}
}

// REMOVED: UpdateCustomerValidation (no longer needed - use main CustomerValidation)

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

// NEW: Customer-specific helper functions

// ValidateCustomerColor validates customer color assignments
func ValidateCustomerColor(color string) error {
	if color == "" {
		return nil // Optional field
	}
	
	if !ValidateColorCode(color) {
		return ValidationError{
			Field:   "color",
			Value:   color,
			Message: "invalid color code",
		}
	}
	
	return nil
}

// ValidateCustomerLoss validates customer loss rates
func ValidateCustomerLoss(loss string) error {
	if loss == "" {
		return nil // Optional field
	}
	
	if !ValidateLossRate(loss) {
		return ValidationError{
			Field:   "loss",
			Value:   loss,
			Message: "invalid loss rate format",
		}
	}
	
	return nil
}

// BatchCustomerValidation for bulk customer operations
type BatchCustomerValidation struct {
	Customers []CustomerValidation `json:"customers"`
}

// Validate validates all customers in batch
func (bcv *BatchCustomerValidation) Validate() map[int]error {
	errors := make(map[int]error)
	
	for i, customer := range bcv.Customers {
		if err := customer.Validate(); err != nil {
			errors[i] = err
		}
	}
	
	return errors
}

// ToCustomerModels converts batch to slice of customer models
func (bcv *BatchCustomerValidation) ToCustomerModels() []*models.Customer {
	var customers []*models.Customer
	
	for _, validation := range bcv.Customers {
		customers = append(customers, validation.ToCustomerModel())
	}
	
	return customers
}

// Helper functions for common validation patterns (keeping existing ones)

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

// BatchInventoryValidation for bulk operations (keeping existing)
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

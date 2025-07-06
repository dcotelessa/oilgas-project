// backend/internal/models/customer_models.go
package models

import "time"

// Customer represents a customer in the oil & gas inventory system
type Customer struct {
	CustomerID      int       `json:"customer_id" db:"customer_id"`
	Customer        string    `json:"customer" db:"customer"`
	BillingAddress  string    `json:"billing_address" db:"billing_address"`
	BillingCity     string    `json:"billing_city" db:"billing_city"`
	BillingState    string    `json:"billing_state" db:"billing_state"`
	BillingZipcode  string    `json:"billing_zipcode" db:"billing_zipcode"`
	Contact         string    `json:"contact" db:"contact"`
	Phone           string    `json:"phone" db:"phone"`
	Fax             string    `json:"fax" db:"fax"`
	Email           string    `json:"email" db:"email"`
	
	// Color coding system (for pipe identification)
	Color1          string    `json:"color1" db:"color1"`
	Color2          string    `json:"color2" db:"color2"`
	Color3          string    `json:"color3" db:"color3"`
	Color4          string    `json:"color4" db:"color4"`
	Color5          string    `json:"color5" db:"color5"`
	
	// Loss tracking system
	Loss1           string    `json:"loss1" db:"loss1"`
	Loss2           string    `json:"loss2" db:"loss2"`
	Loss3           string    `json:"loss3" db:"loss3"`
	Loss4           string    `json:"loss4" db:"loss4"`
	Loss5           string    `json:"loss5" db:"loss5"`
	
	// Work String color coding
	WSColor1        string    `json:"ws_color1" db:"wscolor1"`
	WSColor2        string    `json:"ws_color2" db:"wscolor2"`
	WSColor3        string    `json:"ws_color3" db:"wscolor3"`
	WSColor4        string    `json:"ws_color4" db:"wscolor4"`
	WSColor5        string    `json:"ws_color5" db:"wscolor5"`
	
	// Work String loss tracking
	WSLoss1         string    `json:"ws_loss1" db:"wsloss1"`
	WSLoss2         string    `json:"ws_loss2" db:"wsloss2"`
	WSLoss3         string    `json:"ws_loss3" db:"wsloss3"`
	WSLoss4         string    `json:"ws_loss4" db:"wsloss4"`
	WSLoss5         string    `json:"ws_loss5" db:"wsloss5"`
	
	// System fields
	Deleted         bool      `json:"deleted" db:"deleted"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// IsActive returns true if customer is not deleted
func (c *Customer) IsActive() bool {
	return !c.Deleted
}

// GetDisplayName returns the customer name for display
func (c *Customer) GetDisplayName() string {
	if c.Customer != "" {
		return c.Customer
	}
	return "Unknown Customer"
}

// GetFullAddress returns the complete billing address
func (c *Customer) GetFullAddress() string {
	if c.BillingAddress == "" {
		return ""
	}
	
	address := c.BillingAddress
	if c.BillingCity != "" {
		address += ", " + c.BillingCity
	}
	if c.BillingState != "" {
		address += ", " + c.BillingState
	}
	if c.BillingZipcode != "" {
		address += " " + c.BillingZipcode
	}
	
	return address
}

// GetColors returns all non-empty color codes
func (c *Customer) GetColors() []string {
	colors := []string{}
	colorFields := []string{c.Color1, c.Color2, c.Color3, c.Color4, c.Color5}
	
	for _, color := range colorFields {
		if color != "" {
			colors = append(colors, color)
		}
	}
	
	return colors
}

// GetWSColors returns all non-empty work string color codes
func (c *Customer) GetWSColors() []string {
	colors := []string{}
	colorFields := []string{c.WSColor1, c.WSColor2, c.WSColor3, c.WSColor4, c.WSColor5}
	
	for _, color := range colorFields {
		if color != "" {
			colors = append(colors, color)
		}
	}
	
	return colors
}

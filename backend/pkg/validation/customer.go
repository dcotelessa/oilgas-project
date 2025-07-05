package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NormalizePhone cleans and formats phone numbers
func NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}
	
	// Remove all non-digit characters
	re := regexp.MustCompile(`[^\d]`)
	cleaned := re.ReplaceAllString(phone, "")
	
	// Format as (XXX) XXX-XXXX if 10 digits
	if len(cleaned) == 10 {
		return fmt.Sprintf("(%s) %s-%s", cleaned[:3], cleaned[3:6], cleaned[6:])
	}
	
	// Return as-is if not standard format
	return phone
}

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) bool {
	if email == "" {
		return true // Optional field
	}
	
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}

// CustomerValidation represents the validation struct for customers
type CustomerValidation struct {
	// Basic customer information
	CustomerName string `json:"customer_name" validate:"required,min=1,max=50"`
	Address      string `json:"address" validate:"max=50"`
	City         string `json:"city" validate:"max=50"`
	State        string `json:"state" validate:"max=50"`
	Zip          string `json:"zip" validate:"max=50"`
	Contact      string `json:"contact" validate:"max=50"`
	Phone        string `json:"phone" validate:"max=50"`
	Fax          string `json:"fax" validate:"max=50"`
	Email        string `json:"email" validate:"email,max=50"`
	
	// Color tracking fields (probably for pipe color coding)
	Color1 string `json:"color1" validate:"max=50"`
	Color2 string `json:"color2" validate:"max=50"`
	Color3 string `json:"color3" validate:"max=50"`
	Color4 string `json:"color4" validate:"max=50"`
	Color5 string `json:"color5" validate:"max=50"`
	
	// Loss tracking fields (probably for loss rates or percentages)
	Loss1 string `json:"loss1" validate:"max=50"`
	Loss2 string `json:"loss2" validate:"max=50"`
	Loss3 string `json:"loss3" validate:"max=50"`
	Loss4 string `json:"loss4" validate:"max=50"`
	Loss5 string `json:"loss5" validate:"max=50"`
	
	// WS (Work String?) Color tracking fields
	WSColor1 string `json:"ws_color1" validate:"max=50"`
	WSColor2 string `json:"ws_color2" validate:"max=50"`
	WSColor3 string `json:"ws_color3" validate:"max=50"`
	WSColor4 string `json:"ws_color4" validate:"max=50"`
	WSColor5 string `json:"ws_color5" validate:"max=50"`
	
	// WS Loss tracking fields
	WSLoss1 string `json:"ws_loss1" validate:"max=50"`
	WSLoss2 string `json:"ws_loss2" validate:"max=50"`
	WSLoss3 string `json:"ws_loss3" validate:"max=50"`
	WSLoss4 string `json:"ws_loss4" validate:"max=50"`
	WSLoss5 string `json:"ws_loss5" validate:"max=50"`
}

// ValidateColorCode validates oil & gas industry color codes
func ValidateColorCode(color string) bool {
	if color == "" {
		return true // Optional field
	}
	
	// Common oil & gas color codes
	validColors := []string{
		"RED", "BLUE", "GREEN", "YELLOW", "WHITE", "BLACK", "ORANGE", "PURPLE",
		"PINK", "BROWN", "GRAY", "SILVER", "GOLD", "CLEAR", "NATURAL",
	}
	
	upperColor := strings.ToUpper(strings.TrimSpace(color))
	for _, valid := range validColors {
		if upperColor == valid {
			return true
		}
	}
	
	return false
}

// ValidateLossRate validates loss percentage or rate
func ValidateLossRate(loss string) bool {
	if loss == "" {
		return true // Optional field
	}
	
	// Try to parse as percentage
	if strings.HasSuffix(loss, "%") {
		percentStr := strings.TrimSuffix(loss, "%")
		if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
			return percent >= 0 && percent <= 100
		}
	}
	
	// Try to parse as decimal
	if rate, err := strconv.ParseFloat(loss, 64); err == nil {
		return rate >= 0 && rate <= 1
	}
	
	return false
}

// Validate performs validation on CustomerValidation
func (cv *CustomerValidation) Validate() error {
	if strings.TrimSpace(cv.CustomerName) == "" {
		return fmt.Errorf("customer name is required")
	}
	
	if len(cv.CustomerName) > 50 {
		return fmt.Errorf("customer name cannot exceed 50 characters")
	}
	
	if cv.Email != "" && !ValidateEmail(cv.Email) {
		return fmt.Errorf("invalid email format")
	}
	
	if cv.State != "" && len(cv.State) > 2 {
		return fmt.Errorf("state should be a 2-character code")
	}
	
	// Validate color codes
	colors := []string{cv.Color1, cv.Color2, cv.Color3, cv.Color4, cv.Color5,
		cv.WSColor1, cv.WSColor2, cv.WSColor3, cv.WSColor4, cv.WSColor5}
	
	for i, color := range colors {
		if !ValidateColorCode(color) {
			return fmt.Errorf("invalid color code at position %d: %s", i+1, color)
		}
	}
	
	// Validate loss rates
	losses := []string{cv.Loss1, cv.Loss2, cv.Loss3, cv.Loss4, cv.Loss5,
		cv.WSLoss1, cv.WSLoss2, cv.WSLoss3, cv.WSLoss4, cv.WSLoss5}
	
	for i, loss := range losses {
		if !ValidateLossRate(loss) {
			return fmt.Errorf("invalid loss rate at position %d: %s", i+1, loss)
		}
	}
	
	return nil
}

// NormalizeCustomerData normalizes customer data for consistency
func (cv *CustomerValidation) NormalizeCustomerData() {
	cv.CustomerName = strings.TrimSpace(cv.CustomerName)
	cv.Address = strings.TrimSpace(cv.Address)
	cv.City = strings.TrimSpace(cv.City)
	cv.State = strings.ToUpper(strings.TrimSpace(cv.State))
	cv.Zip = strings.TrimSpace(cv.Zip)
	cv.Contact = strings.TrimSpace(cv.Contact)
	cv.Phone = NormalizePhone(cv.Phone)
	cv.Fax = NormalizePhone(cv.Fax)
	cv.Email = strings.ToLower(strings.TrimSpace(cv.Email))
	
	// Normalize color codes
	cv.Color1 = strings.ToUpper(strings.TrimSpace(cv.Color1))
	cv.Color2 = strings.ToUpper(strings.TrimSpace(cv.Color2))
	cv.Color3 = strings.ToUpper(strings.TrimSpace(cv.Color3))
	cv.Color4 = strings.ToUpper(strings.TrimSpace(cv.Color4))
	cv.Color5 = strings.ToUpper(strings.TrimSpace(cv.Color5))
	
	cv.WSColor1 = strings.ToUpper(strings.TrimSpace(cv.WSColor1))
	cv.WSColor2 = strings.ToUpper(strings.TrimSpace(cv.WSColor2))
	cv.WSColor3 = strings.ToUpper(strings.TrimSpace(cv.WSColor3))
	cv.WSColor4 = strings.ToUpper(strings.TrimSpace(cv.WSColor4))
	cv.WSColor5 = strings.ToUpper(strings.TrimSpace(cv.WSColor5))
	
	// Normalize loss rates (trim spaces)
	cv.Loss1 = strings.TrimSpace(cv.Loss1)
	cv.Loss2 = strings.TrimSpace(cv.Loss2)
	cv.Loss3 = strings.TrimSpace(cv.Loss3)
	cv.Loss4 = strings.TrimSpace(cv.Loss4)
	cv.Loss5 = strings.TrimSpace(cv.Loss5)
	
	cv.WSLoss1 = strings.TrimSpace(cv.WSLoss1)
	cv.WSLoss2 = strings.TrimSpace(cv.WSLoss2)
	cv.WSLoss3 = strings.TrimSpace(cv.WSLoss3)
	cv.WSLoss4 = strings.TrimSpace(cv.WSLoss4)
	cv.WSLoss5 = strings.TrimSpace(cv.WSLoss5)
}

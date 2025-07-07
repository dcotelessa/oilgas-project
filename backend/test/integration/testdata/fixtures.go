// backend/test/integration/testdata/fixtures.go
package integration

import (
	"time"
	"oilgas-backend/internal/models"
)

// TestFixtures provides reusable test data
type TestFixtures struct{}

func NewTestFixtures() *TestFixtures {
	return &TestFixtures{}
}

// StandardCustomers returns a set of standard test customers
func (f *TestFixtures) StandardCustomers() []*models.Customer {
	return []*models.Customer{
		{
			Customer:       "Alpha Oil & Gas Company",
			BillingAddress: "1234 Industry Blvd",
			BillingCity:    "Houston",
			BillingState:   "TX",
			BillingZipcode: "77001",
			Contact:        "John Smith",
			Phone:          "713-555-0001",
			Email:          "john@alphaoil.com",
		},
		{
			Customer:       "Beta Drilling Services",
			BillingAddress: "5678 Drilling Way",
			BillingCity:    "Midland",
			BillingState:   "TX",
			BillingZipcode: "79701",
			Contact:        "Sarah Johnson",
			Phone:          "432-555-0002",
			Email:          "sarah@betadrilling.com",
		},
		{
			Customer:       "Gamma Energy Solutions",
			BillingAddress: "9012 Energy Plaza",
			BillingCity:    "Oklahoma City",
			BillingState:   "OK",
			BillingZipcode: "73102",
			Contact:        "Mike Davis",
			Phone:          "405-555-0003",
			Email:          "mike@gammaenergy.com",
		},
	}
}

// StandardGrades returns common oil & gas pipe grades
func (f *TestFixtures) StandardGrades() []string {
	return []string{"J55", "JZ55", "L80", "N80", "P105", "P110", "Q125"}
}

// StandardSizes returns common pipe sizes
func (f *TestFixtures) StandardSizes() []string {
	return []string{
		"4 1/2\"",
		"5\"",
		"5 1/2\"",
		"7\"",
		"9 5/8\"",
		"13 3/8\"",
		"20\"",
	}
}

// StandardConnections returns common connection types
func (f *TestFixtures) StandardConnections() []string {
	return []string{"LTC", "BTC", "STC", "PREMIUM", "FLUSH"}
}

// ReceivedItemsForCustomer generates received items for a specific customer
func (f *TestFixtures) ReceivedItemsForCustomer(customer *models.Customer, count int) []*models.ReceivedItem {
	sizes := f.StandardSizes()
	grades := f.StandardGrades()
	connections := f.StandardConnections()
	
	items := make([]*models.ReceivedItem, count)
	
	for i := 0; i < count; i++ {
		items[i] = &models.ReceivedItem{
			WorkOrder:    fmt.Sprintf("WO-FIXTURE-%s-%03d", customer.Customer[:3], i+1),
			CustomerID:   customer.CustomerID,
			Customer:     customer.Customer,
			Joints:       50 + (i*25)%200, // Vary between 50-250
			Size:         sizes[i%len(sizes)],
			Weight:       fmt.Sprintf("%d", 15+(i*5)%30), // Vary weight
			Grade:        grades[i%len(grades)],
			Connection:   connections[i%len(connections)],
			Color:        []string{"RED", "BLUE", "GREEN", "YELLOW", "BLACK"}[i%5],
			Location:     fmt.Sprintf("Yard %c", 'A'+rune(i%4)),
			DateReceived: timePtr(time.Now().AddDate(0, 0, -(i+1))),
		}
	}
	
	return items
}

// InspectionItemsFromReceived creates inspection records from received items
func (f *TestFixtures) InspectionItemsFromReceived(received []*models.ReceivedItem, passRate float64) []*models.InspectionItem {
	inspections := make([]*models.InspectionItem, len(received))
	
	for i, item := range received {
		passedJoints := int(float64(item.Joints) * passRate)
		failedJoints := item.Joints - passedJoints
		
		inspections[i] = &models.InspectionItem{
			WorkOrder:      item.WorkOrder,
			CustomerID:     item.CustomerID,
			Customer:       item.Customer,
			Joints:         item.Joints,
			Size:           item.Size,
			Weight:         item.Weight,
			Grade:          item.Grade,
			Connection:     item.Connection,
			PassedJoints:   passedJoints,
			FailedJoints:   failedJoints,
			InspectionDate: timePtr(item.DateReceived.Add(24 * time.Hour)),
			Inspector:      []string{"John Doe", "Jane Smith", "Bob Wilson"}[i%3],
			Notes:          fmt.Sprintf("Inspection completed. %d joints failed quality check.", failedJoints),
		}
	}
	
	return inspections
}

// InventoryItemsFromInspections creates inventory from inspection results
func (f *TestFixtures) InventoryItemsFromInspections(inspections []*models.InspectionItem) []*models.InventoryItem {
	inventory := make([]*models.InventoryItem, len(inspections))
	
	for i, inspection := range inspections {
		inventory[i] = &models.InventoryItem{
			CustomerID: inspection.CustomerID,
			Customer:   inspection.Customer,
			Joints:     inspection.PassedJoints, // Only passed joints go to inventory
			Size:       inspection.Size,
			Weight:     inspection.Weight,
			Grade:      inspection.Grade,
			Connection: inspection.Connection,
			Color:      "PROCESSED", // Mark as processed
			Location:   "Main Storage",
			DateIn:     timePtr(inspection.InspectionDate.Add(48 * time.Hour)),
		}
	}
	
	return inventory
}


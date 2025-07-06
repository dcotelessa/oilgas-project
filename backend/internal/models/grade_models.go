// backend/internal/models/grade_models.go
package models

// GradeUsage represents usage statistics for a specific grade
type GradeUsage struct {
	Grade          string                 `json:"grade"`
	InventoryCount int                    `json:"inventory_count"`
	ReceivedCount  int                    `json:"received_count"`
	FletcherCount  int                    `json:"fletcher_count"`
	TotalJoints    int                    `json:"total_joints"`
	CustomerUsage  []CustomerGradeUsage   `json:"customer_usage"`
}

// CustomerGradeUsage represents how much a customer uses a specific grade
type CustomerGradeUsage struct {
	CustomerName string `json:"customer_name"`
	ItemCount    int    `json:"item_count"`
	TotalJoints  int    `json:"total_joints"`
}


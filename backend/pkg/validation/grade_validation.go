// backend/pkg/validation/grade_validation.go
package validation

import (
	"strings"

	"oilgas-backend/internal/models"
)

// GradeValidation validates grade creation and updates
type GradeValidation struct {
	Grade string `json:"grade" binding:"required"`
}

// Validate validates grade data
func (gv *GradeValidation) Validate() error {
	// Validate grade name
	if gv.Grade == "" {
		return ValidationError{
			Field:   "grade",
			Value:   gv.Grade,
			Message: "grade is required",
		}
	}

	// Normalize and validate grade
	normalizedGrade := strings.ToUpper(strings.TrimSpace(gv.Grade))
	if err := ValidateGrade(normalizedGrade); err != nil {
		return ValidationError{
			Field:   "grade",
			Value:   gv.Grade,
			Message: err.Error(),
		}
	}

	return nil
}

// ToGradeModel converts validation to grade model
func (gv *GradeValidation) ToGradeModel() *models.Grade {
	return &models.Grade{
		Grade: NormalizeGrade(gv.Grade),
	}
}

// FromGradeModel converts grade model to validation
func FromGradeModel(grade *models.Grade) *GradeValidation {
	return &GradeValidation{
		Grade: grade.Grade,
	}
}

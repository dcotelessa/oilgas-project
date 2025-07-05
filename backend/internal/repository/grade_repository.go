// backend/internal/repository/grade_repository.go
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/internal/models"
)

type gradeRepository struct {
	db *pgxpool.Pool
}

func NewGradeRepository(db *pgxpool.Pool) GradeRepository {
	return &gradeRepository{db: db}
}

func (r *gradeRepository) GetAll(ctx context.Context) ([]models.Grade, error) {
	query := `SELECT grade FROM store.grade ORDER BY grade`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}
	defer rows.Close()

	var grades []models.Grade
	for rows.Next() {
		var grade models.Grade
		if err := rows.Scan(&grade.Grade); err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}
		grades = append(grades, grade)
	}

	return grades, rows.Err()
}

func (r *gradeRepository) Create(ctx context.Context, grade *models.Grade) error {
	// Normalize grade name
	grade.Grade = strings.ToUpper(strings.TrimSpace(grade.Grade))
	
	query := `INSERT INTO store.grade (grade) VALUES ($1)`
	
	_, err := r.db.Exec(ctx, query, grade.Grade)
	if err != nil {
		// Check for duplicate
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return fmt.Errorf("grade '%s' already exists", grade.Grade)
		}
		return fmt.Errorf("failed to create grade: %w", err)
	}
	
	return nil
}

func (r *gradeRepository) Delete(ctx context.Context, gradeName string) error {
	gradeName = strings.ToUpper(strings.TrimSpace(gradeName))
	
	query := `DELETE FROM store.grade WHERE grade = $1`
	
	result, err := r.db.Exec(ctx, query, gradeName)
	if err != nil {
		return fmt.Errorf("failed to delete grade: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("grade '%s' not found", gradeName)
	}
	
	return nil
}

func (r *gradeRepository) IsInUse(ctx context.Context, gradeName string) (bool, error) {
	gradeName = strings.ToUpper(strings.TrimSpace(gradeName))
	
	// Check across all tables that might reference grades
	queries := []string{
		`SELECT EXISTS(SELECT 1 FROM store.inventory WHERE UPPER(grade) = $1 AND deleted = false)`,
		`SELECT EXISTS(SELECT 1 FROM store.received WHERE UPPER(grade) = $1 AND deleted = false)`,
		`SELECT EXISTS(SELECT 1 FROM store.fletcher WHERE UPPER(grade) = $1 AND deleted = false)`,
		`SELECT EXISTS(SELECT 1 FROM store.bakeout WHERE UPPER(grade) = $1)`,
	}
	
	for _, query := range queries {
		var inUse bool
		err := r.db.QueryRow(ctx, query, gradeName).Scan(&inUse)
		if err != nil {
			return false, fmt.Errorf("failed to check grade usage: %w", err)
		}
		if inUse {
			return true, nil
		}
	}
	
	return false, nil
}

func (r *gradeRepository) GetUsageStats(ctx context.Context, gradeName string) (*models.GradeUsage, error) {
	gradeName = strings.ToUpper(strings.TrimSpace(gradeName))
	
	usage := &models.GradeUsage{
		Grade: gradeName,
	}
	
	// Count in inventory
	err := r.db.QueryRow(ctx, 
		`SELECT COUNT(*), COALESCE(SUM(joints), 0) FROM store.inventory WHERE UPPER(grade) = $1 AND deleted = false AND joints > 0`,
		gradeName).Scan(&usage.InventoryCount, &usage.TotalJoints)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory usage: %w", err)
	}
	
	// Count in received
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM store.received WHERE UPPER(grade) = $1 AND deleted = false`,
		gradeName).Scan(&usage.ReceivedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get received usage: %w", err)
	}
	
	// Count in fletcher
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM store.fletcher WHERE UPPER(grade) = $1 AND deleted = false`,
		gradeName).Scan(&usage.FletcherCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get fletcher usage: %w", err)
	}
	
	// Get customer usage (top 5 customers using this grade)
	rows, err := r.db.Query(ctx, `
		SELECT customer, COUNT(*), SUM(joints)
		FROM store.inventory 
		WHERE UPPER(grade) = $1 AND deleted = false AND joints > 0
		GROUP BY customer
		ORDER BY COUNT(*) DESC
		LIMIT 5
	`, gradeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer usage: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var customerUsage models.CustomerGradeUsage
		err := rows.Scan(&customerUsage.CustomerName, &customerUsage.ItemCount, &customerUsage.TotalJoints)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer usage: %w", err)
		}
		usage.CustomerUsage = append(usage.CustomerUsage, customerUsage)
	}
	
	return usage, rows.Err()
}

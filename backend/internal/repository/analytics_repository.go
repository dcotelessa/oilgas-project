// backend/internal/repository/analytics_repository.go
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/internal/models"
)

// AnalyticsRepository handles dashboard and analytics queries
type AnalyticsRepository interface {
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)
	GetJobSummaries(ctx context.Context) ([]models.JobSummary, error)
	GetRecentActivity(ctx context.Context, limit int) ([]models.ActivityItem, error)
	GetCustomerActivity(ctx context.Context, days int, customerID *int) (*models.CustomerActivity, error)
	GetGradeAnalytics(ctx context.Context) (*models.GradeAnalytics, error)
	GetInventoryTrends(ctx context.Context, days int) (*models.InventoryTrends, error)
}

type analyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Get basic counts with single optimized query
	query := `
		WITH stats AS (
			SELECT 
				COUNT(DISTINCT CASE WHEN i.deleted = false AND i.joints > 0 THEN i.id END) as active_inventory,
				COUNT(DISTINCT CASE WHEN c.deleted = false THEN c.customer_id END) as total_customers,
				COUNT(DISTINCT CASE WHEN r.deleted = false AND r.complete = false THEN r.id END) as active_jobs,
				COALESCE(SUM(CASE WHEN i.deleted = false AND i.joints > 0 THEN i.joints END), 0) as total_joints
			FROM store.customers c
			LEFT JOIN store.inventory i ON c.customer_id = i.customer_id
			LEFT JOIN store.received r ON c.customer_id = r.customer_id
		)
		SELECT active_inventory, total_customers, active_jobs, total_joints FROM stats
	`

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.ActiveInventory, &stats.TotalCustomers, 
		&stats.ActiveJobs, &stats.TotalJoints,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	// Get top customers (most active)
	topCustomers, err := r.getTopCustomers(ctx, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get top customers: %w", err)
	}
	stats.TopCustomers = topCustomers

	// Get inventory by grade
	inventoryByGrade, err := r.getInventoryByGrade(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory by grade: %w", err)
	}
	stats.InventoryByGrade = inventoryByGrade

	stats.LastUpdated = time.Now()
	return stats, nil
}

func (r *analyticsRepository) getTopCustomers(ctx context.Context, limit int) ([]models.CustomerStats, error) {
	query := `
		SELECT 
			c.customer_id,
			c.customer,
			COUNT(DISTINCT CASE WHEN r.complete = false AND r.deleted = false THEN r.id END) as active_jobs,
			COALESCE(SUM(CASE WHEN i.joints > 0 AND i.deleted = false THEN i.joints END), 0) as total_joints
		FROM store.customers c
		LEFT JOIN store.received r ON c.customer_id = r.customer_id
		LEFT JOIN store.inventory i ON c.customer_id = i.customer_id
		WHERE c.deleted = false
		GROUP BY c.customer_id, c.customer
		HAVING COUNT(DISTINCT CASE WHEN r.complete = false AND r.deleted = false THEN r.id END) > 0
			OR COALESCE(SUM(CASE WHEN i.joints > 0 AND i.deleted = false THEN i.joints END), 0) > 0
		ORDER BY active_jobs DESC, total_joints DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []models.CustomerStats
	for rows.Next() {
		var customer models.CustomerStats
		err := rows.Scan(&customer.CustomerID, &customer.CustomerName, 
			&customer.ActiveJobs, &customer.TotalJoints)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, rows.Err()
}

func (r *analyticsRepository) getInventoryByGrade(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT grade, SUM(joints) as total_joints
		FROM store.inventory 
		WHERE deleted = false AND joints > 0 AND grade IS NOT NULL
		GROUP BY grade 
		ORDER BY total_joints DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inventoryByGrade := make(map[string]int)
	for rows.Next() {
		var grade string
		var joints int
		if err := rows.Scan(&grade, &joints); err != nil {
			return nil, err
		}
		inventoryByGrade[grade] = joints
	}

	return inventoryByGrade, rows.Err()
}

func (r *analyticsRepository) GetJobSummaries(ctx context.Context) ([]models.JobSummary, error) {
	// Simplified job state summary without complex workflow logic
	query := `
		WITH job_states AS (
			SELECT 
				CASE 
					WHEN complete = true THEN 'completed'
					WHEN in_production IS NOT NULL THEN 'in_production'
					WHEN inspected_date IS NOT NULL THEN 'inspected'
					ELSE 'received'
				END as status,
				joints,
				date_received
			FROM store.received 
			WHERE deleted = false
		)
		SELECT 
			status,
			COUNT(*) as count,
			COALESCE(SUM(joints), 0) as total_joints,
			ROUND(AVG(EXTRACT(DAYS FROM (NOW() - date_received))), 1) as avg_days
		FROM job_states
		GROUP BY status
		ORDER BY 
			CASE status
				WHEN 'received' THEN 1
				WHEN 'in_production' THEN 2
				WHEN 'inspected' THEN 3
				WHEN 'completed' THEN 4
			END
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get job summaries: %w", err)
	}
	defer rows.Close()

	var summaries []models.JobSummary
	for rows.Next() {
		var summary models.JobSummary
		var status string
		err := rows.Scan(&status, &summary.Count, &summary.TotalJoints, &summary.AvgDays)
		if err != nil {
			return nil, err
		}
		summary.Status = status
		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

func (r *analyticsRepository) GetRecentActivity(ctx context.Context, limit int) ([]models.ActivityItem, error) {
	query := `
		SELECT 
			'inventory' as type,
			i.id::text as reference_id,
			i.customer as title,
			CONCAT(i.joints, ' joints of ', i.size, ' ', i.grade) as description,
			i.customer_id,
			COALESCE(i.date_in, i.created_at) as activity_time
		FROM store.inventory i
		WHERE i.deleted = false
		UNION ALL
		SELECT 
			'received' as type,
			r.id::text as reference_id,
			r.customer as title,
			CONCAT('Received ', r.joints, ' joints of ', r.size, ' ', r.grade) as description,
			r.customer_id,
			COALESCE(r.date_received, r.created_at) as activity_time
		FROM store.received r
		WHERE r.deleted = false
		ORDER BY activity_time DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	defer rows.Close()

	var activities []models.ActivityItem
	for rows.Next() {
		var activity models.ActivityItem
		err := rows.Scan(&activity.Type, &activity.ReferenceID, &activity.Title,
			&activity.Description, &activity.CustomerID, &activity.ActivityTime)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

func (r *analyticsRepository) GetCustomerActivity(ctx context.Context, days int, customerID *int) (*models.CustomerActivity, error) {
	// Implementation for customer-specific activity analysis
	activity := &models.CustomerActivity{
		Days: days,
	}

	var whereClause string
	var args []interface{}
	
	if customerID != nil {
		whereClause = "AND customer_id = $2"
		args = []interface{}{days, *customerID}
	} else {
		args = []interface{}{days}
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT customer_id) as active_customers,
			COUNT(*) as total_activities,
			COALESCE(SUM(joints), 0) as total_joints
		FROM (
			SELECT customer_id, joints, date_in as activity_date
			FROM store.inventory 
			WHERE deleted = false AND date_in >= NOW() - INTERVAL '%d days' %s
			UNION ALL
			SELECT customer_id, joints, date_received as activity_date
			FROM store.received 
			WHERE deleted = false AND date_received >= NOW() - INTERVAL '%d days' %s
		) activities
	`, days, whereClause, days, whereClause)

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&activity.ActiveCustomers, &activity.TotalActivities, &activity.TotalJoints)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer activity: %w", err)
	}

	return activity, nil
}

func (r *analyticsRepository) GetGradeAnalytics(ctx context.Context) (*models.GradeAnalytics, error) {
	analytics := &models.GradeAnalytics{
		GradeDistribution: make(map[string]models.GradeStats),
	}

	query := `
		SELECT 
			grade,
			COUNT(*) as item_count,
			SUM(joints) as total_joints,
			AVG(joints) as avg_joints_per_item,
			COUNT(DISTINCT customer_id) as customer_count
		FROM store.inventory 
		WHERE deleted = false AND joints > 0 AND grade IS NOT NULL
		GROUP BY grade
		ORDER BY total_joints DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get grade analytics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var grade string
		var stats models.GradeStats
		err := rows.Scan(&grade, &stats.ItemCount, &stats.TotalJoints, 
			&stats.AvgJointsPerItem, &stats.CustomerCount)
		if err != nil {
			return nil, err
		}
		analytics.GradeDistribution[grade] = stats
	}

	return analytics, rows.Err()
}

func (r *analyticsRepository) GetInventoryTrends(ctx context.Context, days int) (*models.InventoryTrends, error) {
	// Daily inventory trends for the specified period
	query := `
		SELECT 
			DATE(date_in) as trend_date,
			COUNT(*) as items_added,
			SUM(joints) as joints_added,
			COUNT(DISTINCT customer_id) as customers_active
		FROM store.inventory 
		WHERE deleted = false 
			AND date_in >= NOW() - INTERVAL '%d days'
			AND date_in IS NOT NULL
		GROUP BY DATE(date_in)
		ORDER BY trend_date DESC
	`

	rows, err := r.db.Query(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory trends: %w", err)
	}
	defer rows.Close()

	trends := &models.InventoryTrends{
		Days:       days,
		DailyStats: []models.DailyInventoryStats{},
	}

	for rows.Next() {
		var daily models.DailyInventoryStats
		err := rows.Scan(&daily.Date, &daily.ItemsAdded, 
			&daily.JointsAdded, &daily.CustomersActive)
		if err != nil {
			return nil, err
		}
		trends.DailyStats = append(trends.DailyStats, daily)
	}

	return trends, rows.Err()
}

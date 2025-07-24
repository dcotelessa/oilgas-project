// scripts/utilities/system_health.go  
package main

import (
	"context"
	"fmt"
	"time"
)

func SystemHealthCheck() {
	fmt.Println("🏥 System Health Check...")
	
	pool := getDBConnection()
	defer pool.Close()
	
	ctx := context.Background()
	
	// Database connectivity
	start := time.Now()
	var result int
	err := pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ Database health check failed: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Database responding in %v\n", duration)
	
	// Connection pool stats
	stats := pool.Stat()
	fmt.Printf("📊 Connection Pool: %d/%d connections active\n", 
		stats.AcquiredConns(), stats.MaxConns())
	
	// Check core business data
	tables := []string{"customers", "inventory", "received"}
	for _, table := range tables {
		var count int
		err := pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM store.%s WHERE deleted = false", table)).Scan(&count)
		if err != nil {
			fmt.Printf("⚠️  Cannot count active records in %s: %v\n", table, err)
		} else {
			fmt.Printf("📋 Active records in %s: %d\n", table, count)
		}
	}
	
	// Check for recent activity (last 7 days)
	var recentCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM store.inventory 
		WHERE created_at > NOW() - INTERVAL '7 days'
	`).Scan(&recentCount)
	
	if err != nil {
		fmt.Printf("⚠️  Cannot check recent activity: %v\n", err)
	} else {
		fmt.Printf("📈 New inventory items (7 days): %d\n", recentCount)
	}
	
	fmt.Println("✅ System health check complete!")
}

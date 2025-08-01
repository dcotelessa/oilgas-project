// backend/cmd/server/monitoring_example.go
// Example of how to integrate monitoring into main.go
package main

import (
	"time"
	"oilgas-backend/internal/monitoring"
)

func setupMonitoring(pool *pgxpool.Pool) *monitoring.ConnectionMonitor {
	monitor := monitoring.NewConnectionMonitor()
	monitor.RegisterPool("main", pool)
	monitor.StartMonitoring(30 * time.Second)
	return monitor
}

// In main(): 
// monitor := setupMonitoring(pool)

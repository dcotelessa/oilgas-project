package monitoring

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ConnectionMonitor struct {
	pools map[string]*pgxpool.Pool
}

func NewConnectionMonitor() *ConnectionMonitor {
	return &ConnectionMonitor{
		pools: make(map[string]*pgxpool.Pool),
	}
}

func (m *ConnectionMonitor) RegisterPool(name string, pool *pgxpool.Pool) {
	m.pools[name] = pool
}

func (m *ConnectionMonitor) StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m.logPoolStats()
		}
	}()
}

func (m *ConnectionMonitor) logPoolStats() {
	for name, pool := range m.pools {
		stats := pool.Stat()
		log.Printf("POOL_STATS[%s]: total=%d idle=%d acquired=%d constructing=%d", 
			name, stats.TotalConns(), stats.IdleConns(), 
			stats.AcquiredConns(), stats.ConstructingConns())
	}
}

func (m *ConnectionMonitor) GetStats(name string) map[string]interface{} {
	if pool, exists := m.pools[name]; exists {
		stats := pool.Stat()
		return map[string]interface{}{
			"total_connections":       stats.TotalConns(),
			"idle_connections":        stats.IdleConns(),
			"acquired_connections":    stats.AcquiredConns(),
			"constructing_connections": stats.ConstructingConns(),
		}
	}
	return nil
}

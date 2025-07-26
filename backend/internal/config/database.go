package config

import (
	"database/sql"
	"fmt"
	"os"
)

type DatabaseConfig struct {
	AuthDatabaseURL      string
	TenantDatabaseBaseURL string
	MaxOpenConns         int
	MaxIdleConns         int
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		AuthDatabaseURL:      getEnv("AUTH_DATABASE_URL", "postgresql://user:password@localhost/oilgas_auth?sslmode=disable"),
		TenantDatabaseBaseURL: getEnv("TENANT_DATABASE_BASE_URL", "postgresql://user:password@localhost"),
		MaxOpenConns:         10,
		MaxIdleConns:         5,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

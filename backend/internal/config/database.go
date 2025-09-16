package config

import (
	"os"
)

type DatabaseConfig struct {
	CentralAuthURL   string
	LongBeachURL     string
	BakersfieldURL   string
	MaxOpenConns     int
	MaxIdleConns     int
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		CentralAuthURL:   getEnv("CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable"),
		LongBeachURL:     getEnv("LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable"),
		BakersfieldURL:   getEnv("BAKERSFIELD_DB_URL", ""), // Empty default - not implemented yet
		MaxOpenConns:     10,
		MaxIdleConns:     5,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

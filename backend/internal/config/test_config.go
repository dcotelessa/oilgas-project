// backend/internal/config/test_config.go
package config

type TestConfig struct {
    DatabaseConfig DatabaseConfig
    JWTSecret      string
}

func NewTestConfig() *TestConfig {
    return &TestConfig{
        DatabaseConfig: DatabaseConfig{
            AuthDBConnString: "postgresql://testuser:testpass@localhost:5432/auth_central?sslmode=disable",
            TenantDBConfigs: map[string]TenantDBConfig{
                "longbeach": {
                    ConnectionString: "postgresql://testuser:testpass@localhost:5433/location_longbeach?sslmode=disable",
                    Location:         "Long Beach",
                },
                "bakersfield": {
                    ConnectionString: "postgresql://testuser:testpass@localhost:5434/location_bakersfield?sslmode=disable", 
                    Location:         "Bakersfield",
                },
            },
            MaxOpenConns:    10,
            MaxIdleConns:    5,
            ConnMaxLifetime: time.Hour,
        },
        JWTSecret: "test-jwt-secret-key",
    }
}

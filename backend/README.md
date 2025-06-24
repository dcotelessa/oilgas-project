# Backend - Oil & Gas Inventory System

Go backend using Gin framework with PostgreSQL and in-memory caching.

## Structure

```
backend/
├── cmd/server/          # Main application
├── internal/           # Private packages
│   ├── handlers/      # HTTP handlers
│   ├── services/      # Business logic
│   ├── repository/    # Data access layer
│   └── models/        # Data models
├── pkg/               # Public packages
│   └── cache/        # In-memory cache
├── migrations/        # Database migrations
├── seeds/            # Seed data
└── migrator.go       # Migration tool
```

## Development

```bash
# Install dependencies
go mod tidy

# Run migrations
go run migrator.go migrate local

# Seed database
go run migrator.go seed local

# Start server
go run cmd/server/main.go
```

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/grades` - List oil & gas grades
- `GET /api/v1/customers` - List customers
- `GET /api/v1/inventory` - List inventory

## Cache Configuration

The system uses in-memory caching with configurable TTL:

```go
cache := cache.New(cache.Config{
    TTL:             5 * time.Minute,
    CleanupInterval: 10 * time.Minute,
    MaxSize:         1000,
})
```

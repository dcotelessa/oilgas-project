# Oil & Gas Inventory System - Backend

Go-based backend API for the Oil & Gas Inventory System.

## Quick Start

```bash
# Install dependencies
go mod tidy

# Run migrations
go run migrator.go migrate local

# Seed database
go run migrator.go seed local

# Start development server
go run cmd/server/main.go
```

## API Endpoints

- **Health**: `GET /health`
- **Status**: `GET /api/v1/status`
- **Customers**: `GET /api/v1/customers`
- **Inventory**: `GET /api/v1/inventory`
- **Received**: `GET /api/v1/received`

## Database Operations

```bash
# Show status
go run migrator.go status local

# Reset database (destructive)
go run migrator.go reset local
```

## Structure

```
backend/
├── cmd/server/          # Main application
├── internal/            # Private application code
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business logic
│   ├── repository/      # Data access
│   └── models/          # Data models
├── pkg/                 # Public packages
├── migrations/          # Database migrations
├── seeds/               # Database seed data
└── test/                # Tests
```

This backend integrates with Phase 1 normalized data and provides a foundation for Phase 2 development.

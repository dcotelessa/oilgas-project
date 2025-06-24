# Oil & Gas Inventory System

A modern inventory management system for the oil & gas industry, migrated from ColdFusion to a modern stack.

## ⚠️ Data Security Notice

This repository contains **ONLY** mock/fake data for development purposes. All sensitive customer information, production databases, and real business data are protected by .gitignore and should never be committed to version control.

## Tech Stack

- **Backend**: Go + Gin + PostgreSQL + In-Memory Caching
- **Frontend**: Vue.js 3 + TypeScript + Vite + Pinia
- **Database**: PostgreSQL 15
- **Infrastructure**: Docker + Docker Compose
- **Deployment**: Vultr VPS

## Project Structure

```
├── backend/                 # Go backend application
│   ├── cmd/server/         # Main application entry
│   ├── internal/           # Private application code
│   ├── pkg/               # Public packages (cache, utils)
│   ├── migrations/        # Database migrations
│   └── seeds/            # Database seed data (FAKE DATA ONLY)
├── frontend/              # Vue.js frontend application
│   ├── src/              # Source code
│   └── public/           # Static assets
├── database/             # Database reference files (schema only)
│   ├── schema/          # PostgreSQL schema
│   └── analysis/       # Migration analysis (no sensitive data)
└── docs/               # Project documentation
```

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15 (or use Docker)

### Development Setup

1. **Start the database:**
   ```bash
   make dev-start
   ```

2. **Run migrations:**
   ```bash
   make migrate
   ```

3. **Seed database with FAKE data:**
   ```bash
   make seed
   ```

4. **Start backend:**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

5. **Start frontend:**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

6. **Access the application:**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8000
   - PgAdmin: http://localhost:8080

## Available Commands

- `make help` - Show all available commands
- `make setup` - Complete development setup
- `make migrate` - Run database migrations
- `make seed` - Seed database with FAKE data
- `make status` - Show migration status
- `make reset` - Reset database (destructive)
- `make dev-start` - Start development services
- `make dev-stop` - Stop development services

## Environment Configuration

Copy `.env.local` to `.env` and adjust settings as needed:

```bash
cp .env.local .env
```

**Note:** Only `.env.local` with fake credentials is safe for version control. Real production credentials should never be committed.

## Data Security

### Safe for Version Control:
- ✅ Mock/fake customer data in `seeds/local_seeds.sql`
- ✅ Example environment files (`.env.example`, `.env.local`)
- ✅ Source code and documentation
- ✅ Database schema (structure only)

### Protected by .gitignore:
- ❌ Real customer data and business information
- ❌ Production credentials and API keys
- ❌ Database files (.mdb, backups, exports)
- ❌ Environment files with real credentials

## Migration Notes

This system was migrated from ColdFusion/Access to Go/PostgreSQL. The migration process:

1. **Schema Conversion**: Access → PostgreSQL with type mapping
2. **Data Migration**: CSV export with date/case normalization  
3. **Query Analysis**: ColdFusion queries analyzed for optimization
4. **Grade Validation**: Oil & gas industry grades (J55, JZ55, L80, N80, P105, P110)

## Deployment

### Local Development
Uses Docker Compose with local PostgreSQL and **FAKE** seed data.

### Production
Recommended setup on Vultr:
- Managed PostgreSQL database
- VPS with Docker deployment
- Environment-specific configurations
- **REAL** data imported separately (not from version control)

## Contributing

1. Follow Go and Vue.js best practices
2. Use conventional commits
3. Add tests for new features
4. Update documentation
5. **NEVER** commit real customer data or production credentials

## License

[Your License Here]

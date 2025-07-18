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

## Prerequisites

**Core Requirements:**
- **Go 1.21+** - Backend development ([Download](https://golang.org/dl/))
- **Node.js 18+** - Frontend development ([Download](https://nodejs.org/))
- **Docker & Docker Compose** - Database and deployment ([Download](https://docs.docker.com/get-docker/))
- **PostgreSQL 15** - Database engine (via Docker or local install)
- **PostgreSQL Client (psql)** - Database administration and debugging ⚠️ **ESSENTIAL**

**Development Tools:**
- **Git** - Version control
- **Make** - Build automation (pre-installed on most Unix systems)
- **curl** - API testing (optional but recommended)

### PostgreSQL Client Installation

The PostgreSQL client (`psql`) is **essential** for database administration, debugging, and development. Install it separately from Docker PostgreSQL.

#### macOS
```bash
# Using Homebrew (recommended)
brew install postgresql

# Verify installation
psql --version
```

#### Ubuntu/Debian
```bash
# Install PostgreSQL client only
sudo apt-get update
sudo apt-get install postgresql-client

# Verify installation
psql --version
```

#### CentOS/RHEL/Rocky Linux
```bash
# Install PostgreSQL client
sudo yum install postgresql
# OR for newer versions:
sudo dnf install postgresql

# Verify installation
psql --version
```

#### Windows
- **Download**: https://www.postgresql.org/download/windows/
- **Chocolatey**: `choco install postgresql`
- **Scoop**: `scoop install postgresql`

### Why PostgreSQL Client is Required

1. **Database Administration** - Manage PostgreSQL databases effectively
2. **Development Debugging** - Inspect database state and run queries interactively
3. **Migration Verification** - Check migration results and data integrity
4. **Production Deployment** - Run maintenance scripts in production environments
5. **Industry Standard** - Expected tool for any PostgreSQL development workflow

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
├── scripts/              # Setup and utility scripts
└── docs/               # Project documentation
```

## Quick Setup Verification

Verify all prerequisites are installed:

```bash
# Check Go
go version                        # Should show 1.21+

# Check Node.js  
node --version                    # Should show 18+
npm --version

# Check Docker
docker --version
docker-compose --version

# Check PostgreSQL client ⚠️ CRITICAL
psql --version                    # Should show PostgreSQL client

# Check other tools
git --version
make --version
```

## Development Setup

### 1. Environment Setup
```bash
# Clone repository
git clone <your-repo-url>
cd oil-gas-inventory

# Initialize environment files
make init-env
```

### 2. Database Setup
```bash
# Start PostgreSQL with Docker
make dev-start

# Wait for database to be ready (test with psql)
source .env.local
psql "$DATABASE_URL" -c "SELECT 1;" 

# Run migrations
make migrate ENV=local

# Seed with FAKE development data
make seed ENV=local
```

### 3. Application Setup
```bash
# Install all dependencies
make deps

# Start backend (terminal 1)
make dev-backend

# Start frontend (terminal 2)
make dev-frontend
```

### 4. Verify Setup
```bash
# Test database connection and data
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM store.customers;"

# Test backend API
curl http://localhost:8000/health

# Test frontend (open in browser)
open http://localhost:3000
```

## Available Commands

**Database Operations:**
- `make dev-start` - Start PostgreSQL database
- `make migrate ENV=local` - Run database migrations
- `make seed ENV=local` - Seed database with FAKE data
- `make status ENV=local` - Show migration status
- `make reset ENV=local` - Reset database (destructive)

**Development:**
- `make dev-backend` - Start backend development server
- `make dev-frontend` - Start frontend development server
- `make deps` - Install all dependencies
- `make build` - Build all components

**Testing:**
- `make test` - Run all tests
- `make test-unit` - Run unit tests only
- `make test-integration` - Run integration tests

**Utilities:**
- `make help` - Show all available commands
- `make clean` - Clean up generated files

## Database Operations

### Interactive Database Access
```bash
# Load environment
source .env.local

# Interactive PostgreSQL session
psql "$DATABASE_URL"

# Run single commands
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM store.customers;"

# List all tables in store schema
psql "$DATABASE_URL" -c "\dt store.*"

# Describe a specific table
psql "$DATABASE_URL" -c "\d store.customers"

# Run SQL files
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
```

### Development Database Commands
```bash
# Reset everything and start fresh
make reset ENV=local

# Check what data exists
psql "$DATABASE_URL" -c "
  SELECT 'customers' as table_name, COUNT(*) as count FROM store.customers
  UNION ALL
  SELECT 'inventory' as table_name, COUNT(*) as count FROM store.inventory
  UNION ALL  
  SELECT 'received' as table_name, COUNT(*) as count FROM store.received;
"

# Check migration history
psql "$DATABASE_URL" -c "SELECT * FROM migrations.schema_migrations ORDER BY executed_at;"
```

## Environment Configuration

### Local Development (.env.local)
```bash
# Copy template and customize
cp .env.local .env

# Key settings for local development
DATABASE_URL=postgres://postgres:postgres123@localhost:5432/oilgas_inventory_local
APP_ENV=local
APP_PORT=8000
```

**Note:** Only `.env.local` with fake credentials is safe for version control. Real production credentials should never be committed.

## Access Points

After successful setup:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8000  
- **PgAdmin**: http://localhost:8080
- **Database**: `psql "$DATABASE_URL"`

## Common Issues

### PostgreSQL Client Missing
**Problem**: `psql: command not found`
**Solution**: Install PostgreSQL client using the instructions above

### Database Connection Failed
**Problem**: `Failed to ping database`
**Solutions**:
1. Start database: `make dev-start`
2. Check Docker: `docker-compose ps`
3. Load environment: `source .env.local`

### Migration Issues
**Problem**: `relation does not exist`
**Solutions**:
1. Reset database: `make reset ENV=local`
2. Check migration status: `make status ENV=local`
3. Verify tables: `psql "$DATABASE_URL" -c "\dt store.*"`

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
- Managed PostgreSQL database (never use Docker PostgreSQL in production)
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

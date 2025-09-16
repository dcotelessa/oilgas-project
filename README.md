# Oil & Gas Multi-Tenant Inventory System

A comprehensive multi-tenant inventory management system for oil & gas operations with separate databases per location and centralized authentication.

## 🏗️ Architecture

### Database Structure
- **Central Auth Database** (`auth_central`): Cross-tenant users, permissions, sessions
- **Tenant Databases**: Separate database per location (Long Beach, Bakersfield, etc.)
  - `location_longbeach`: Customers, inventory, work orders, invoices
  - `location_bakersfield`: (Future) Same schema, separate data
  - `location_colorado`: (Future) Same schema, separate data

### Services
- **Customer Domain**: Customer management with tenant isolation
- **Auth Domain**: Multi-tenant authentication and authorization  
- **Inventory Domain**: (Next) Inventory tracking and management
- **Work Order Domain**: (Next) Service workflow management

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Make (optional, for convenience commands)

### Development Setup

1. **Clone and setup environment:**
```bash
git clone <your-repo>
cd oil-gas-inventory
cp .env.example .env
# Update .env with your preferred settings
```

2. **Start development environment:**
```bash
make dev-setup
```
This will:
- Start both databases (auth + Long Beach)
- Run database migrations
- Seed test data
- Verify everything is working

3. **Start the application:**
```bash
make app-run
```

4. **Access the system:**
- API: http://localhost:8080
- PgAdmin: http://localhost:8081 (optional)

### Manual Setup (if not using Make)

1. **Start databases:**
```bash
docker-compose up -d auth-db longbeach-db
```

2. **Run migrations:**
```bash
go run database/scripts/migrate.go auth up
go run database/scripts/migrate.go longbeach up
```

3. **Seed test data:**
```bash
go run cmd/tools/seed/main.go --tenant=longbeach --customers=10 --inventory=50
```

## 📊 Database Management

### Common Commands
```bash
# Database operations
make db-up          # Start databases
make db-down        # Stop databases  
make db-status      # Check database health
make db-reset       # Reset all data (DESTRUCTIVE)

# Migrations
make migrate-auth-up    # Run auth migrations
make migrate-lb-up      # Run Long Beach migrations
make migrate-status     # Check migration status

# Database access
make db-shell-auth      # Access auth database shell
make db-shell-longbeach # Access Long Beach database shell

# Backups
make backup-all        # Backup both databases
```

### Migration from Access Database

1. **Export your Access customers to JSON:**
```json
[
  {
    "id": "1",
    "company_name": "Chevron Corporation",
    "company_code": "CHEV001",
    "billing_address": "123 Oil St",
    "city": "Houston",
    "state": "TX",
    "zip_code": "77001",
    "tax_id": "12-3456789",
    "payment_terms": "NET30",
    "status": "active",
    "primary_contact": "John Smith",
    "contact_email": "john@chevron.com"
  }
]
```

2. **Run migration:**
```bash
make migrate-customers
# Or directly:
go run cmd/tools/migrate-customers/main.go --tenant=longbeach --file=customers.json
```

## 🔧 Development

### Project Structure
```
├── backend/
│   ├── internal/
│   │   ├── customer/      # Customer domain
│   │   ├── auth/          # Authentication domain
│   │   └── shared/        # Shared utilities
│   └── cmd/
│       ├── longbeach/     # Long Beach service
│       └── tools/         # Migration utilities
├── database/
│   ├── init/             # Initial schemas
│   ├── migrations/       # Version migrations  
│   └── scripts/          # Migration tools
└── scripts/              # Setup scripts
```

### Adding a New Tenant

1. **Add tenant to auth database:**
```sql
INSERT INTO tenants (id, name, location, database_name) 
VALUES ('bakersfield', 'Bakersfield Operations', 'Bakersfield, CA', 'location_bakersfield');
```

2. **Update docker-compose.yml:**
```yaml
bakersfield-db:
  image: postgres:15
  container_name: location_bakersfield_db
  environment:
    POSTGRES_DB: location_bakersfield
    # ... rest of config
```

3. **Create tenant-specific service:**
```bash
cp -r cmd/longbeach cmd/bakersfield
# Update tenant references in code
```

## 🧪 Testing

```bash
# Unit tests
make app-test

# Integration tests (requires databases)
make app-test-integration

# Test specific domain
go test ./internal/customer/...
```

## 📚 API Documentation

### Customer Management

**Create Customer:**
```bash
POST /api/v1/customers
{
  "name": "Test Oil Company",
  "company_code": "TEST001",
  "billing_info": {
    "tax_id": "12-3456789",
    "payment_terms": "NET30",
    "address": {
      "street": "123 Oil St",
      "city": "Houston",
      "state": "TX",
      "zip_code": "77001"
    }
  }
}
```

**Search Customers:**
```bash
GET /api/v1/customers?name=oil&status=active&limit=20&offset=0
```

**Register Customer Contact:**
```bash
POST /api/v1/customers/{id}/contacts
{
  "email": "contact@company.com",
  "first_name": "John",
  "last_name": "Smith",
  "role": "manager"
}
```

### Authentication
All API calls require authentication headers:
```bash
curl -H "Authorization: Bearer <jwt-token>" \
     -H "X-Tenant-ID: longbeach" \
     http://localhost:8080/api/v1/customers
```

## 🔒 Security

### Multi-Tenant Isolation
- **Database Level**: Completely separate databases per tenant
- **Application Level**: Tenant ID validation in all operations
- **User Level**: Role-based permissions per tenant

### Customer Contact Access
Customer contacts can only access their own company's data through the customer filter middleware.

## 🚧 Next Steps

### Phase 1: Current - Customer Foundation ✅
- [x] Multi-database architecture
- [x] Customer CRUD operations
- [x] Customer contact management
- [x] Migration from Access

### Phase 2: Work Order Service (Next)
- [ ] Work order domain implementation
- [ ] Inventory integration
- [ ] Service workflow management
- [ ] Invoice generation

### Phase 3: Multi-Location Expansion
- [ ] Bakersfield tenant setup
- [ ] Cross-location reporting
- [ ] Enterprise analytics

## 🆘 Troubleshooting

**Database Connection Issues:**
```bash
# Check database status
make db-status

# View database logs
make db-logs

# Reset databases if corrupted
make db-reset
```

**Migration Issues:**
```bash
# Check migration status
make migrate-status

# Manual migration rollback
go run database/scripts/migrate.go longbeach down
```

**Application Issues:**
```bash
# Check environment variables
cat .env

# Run with debug logging
LOG_LEVEL=debug make app-run
```

## 📞 Support

For issues with:
- **Database setup**: Check docker-compose logs
- **Migrations**: Verify database connectivity and permissions
- **Customer migration**: Validate JSON format and required fields
- **Authentication**: Verify JWT tokens and tenant access

# Oil & Gas Inventory Management System

A modern, multi-tenant inventory management system designed specifically for the oil & gas industry. Built with Go, PostgreSQL, and Vue.js.

## 🎯 Overview

This system replaces legacy ColdFusion applications with a modern, scalable architecture featuring:

- **Multi-tenant architecture** - Support for multiple divisions/locations
- **Row-level security** - Automatic data isolation between tenants
- **Real-time inventory tracking** - Pipes, casings, and equipment management
- **Customer relationship management** - Oil & gas company customer tracking
- **Modern tech stack** - Go backend, PostgreSQL database, Vue.js frontend

## 🏗️ Architecture

### Tech Stack
- **Backend**: Go 1.21+ with Gin framework
- **Database**: PostgreSQL 15+ with Row-Level Security (RLS)
- **Frontend**: Vue.js 3 (planned)
- **Authentication**: Session-based with tenant context
- **Deployment**: Docker-ready

### Database Schema
```
store.tenants                    # Internal divisions (West Texas, Gulf Coast, etc.)
store.customers                  # External oil & gas companies
store.users                      # System users with tenant assignments
store.inventory                  # Current inventory items
store.received                   # Work orders and incoming materials
store.customer_tenant_assignments # Customer-tenant relationships
store.user_tenant_roles          # User permissions within tenants
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Make (for build commands)

### Installation

1. **Clone and setup**
   ```bash
   git clone https://github.com/dcotelessa/oilgas-project
   cd oilgas-project
   ```

2. **Configure environment**
   ```bash
   # Create .env.local with your database settings
   cp .env.example .env.local
   # Edit .env.local with your database credentials
   ```

3. **Setup database and run migrations**
   ```bash
   # Complete setup (creates database, schema, and tenant architecture)
   ./scripts/setup.sh
   
   # Run tenant setup
   make setup
   ```

4. **Start development server**
   ```bash
   make dev
   ```

### Environment Configuration

Create `.env.local` with your database settings:

```bash
# Database Configuration
DATABASE_URL=postgresql://username:password@localhost:5433/oilgas_inventory_local
POSTGRES_DB=oilgas_inventory_local
POSTGRES_USER=username
POSTGRES_PASSWORD=password
POSTGRES_HOST=localhost
POSTGRES_PORT=5433

# Application Configuration
APP_ENV=local
APP_PORT=8000
```

## 📋 Available Commands

### Database Operations
```bash
make db-status              # Check database connection and tables
make setup                  # Run complete tenant setup
make migrate                # Run tenant migrations
make seed-data              # Seed default tenant data
```

### Development
```bash
make dev                    # Start development server
make env-info              # Show environment configuration
make help                  # Show all available commands
```

### Tenant Management
```bash
make create-tenant         # Create new tenant (interactive)
make list-tenants          # List all tenants
make tenant-status         # Show comprehensive tenant status
```

### Debugging
```bash
make debug-migrations      # Debug migration status
make ensure-basic-schema   # Verify basic tables exist
```

## 🏢 Multi-Tenant Architecture

### Tenant Concepts

**Tenants** = Internal divisions of your company
- West Texas Division
- Gulf Coast Division  
- Permian Basin Operations
- etc.

**Customers** = External oil & gas companies
- Chevron, ExxonMobil, Shell, etc.
- Can be served by multiple tenants

**Users** = Your employees
- Belong to specific tenants
- Have roles within those tenants (admin, manager, operator, viewer)

### Data Isolation

Row-Level Security (RLS) ensures:
- ✅ Users only see their tenant's data
- ✅ Admins can see all data or switch tenant context
- ✅ Customers are properly assigned to tenants
- ✅ No cross-tenant data leakage

### Example Usage

```bash
# Create a new division
make create-tenant
# Enter: "West Texas Division"

# Assign customers to tenants
make assign-customer-to-tenant
# Customer ID: 1, Tenant ID: 2, Type: primary

# Test tenant isolation
psql "$DATABASE_URL" -c "SET app.current_tenant_id = 1; SELECT COUNT(*) FROM store.inventory;"
```

## 🗄️ Database Schema Details

### Core Tables

#### `store.tenants`
Internal company divisions with settings and configuration.

#### `store.customers` 
External oil & gas companies that purchase services.

#### `store.users`
System users with tenant assignments and roles.

#### `store.inventory`
Current inventory items (pipes, casings, equipment) with tenant ownership.

#### `store.received`
Work orders and incoming materials with processing status.

### Relationship Tables

#### `store.customer_tenant_assignments`
Many-to-many relationship between customers and tenants.

#### `store.user_tenant_roles`
User permissions within specific tenants.

### Reference Data

#### `store.grade`
Oil & gas industry standard grades (J55, L80, P110, etc.).

#### `store.sizes`
Standard pipe sizes (5 1/2", 7", 9 5/8", etc.).

## 🔧 Development Setup

### Backend Structure
```
backend/
├── cmd/server/              # Main application entry point
├── internal/
│   ├── models/              # Data models and structs
│   ├── handlers/            # HTTP request handlers  
│   ├── middleware/          # Authentication, tenant context
│   ├── repository/          # Database access layer
│   └── services/            # Business logic
├── migrations/              # Database migration files
├── seeds/                   # Database seed data
└── test/                    # Test files
```

### Adding New Features

1. **Add database changes** to new migration files
2. **Create Go models** in `internal/models/`
3. **Add repository methods** in `internal/repository/`
4. **Implement business logic** in `internal/services/`
5. **Create API handlers** in `internal/handlers/`
6. **Add routes** to main server

### Testing

```bash
# Run all tests
make test

# Test tenant isolation
make test-tenant-isolation

# Run specific test suites
cd backend && go test ./internal/services/
```

## 🔐 Security Features

### Row-Level Security (RLS)
- Automatic data filtering by tenant
- Admin bypass capabilities
- Context-aware queries

### Authentication
- Session-based authentication
- Tenant context management
- Role-based permissions

### Data Protection
- SQL injection prevention
- Input validation
- Audit trails

## 📊 API Endpoints

### Health & Status
- `GET /health` - Service health check
- `GET /api/v1/status` - API status

### Tenant Management (Admin Only)
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants` - List tenants
- `GET /api/v1/tenants/:id` - Get tenant details

### Business Operations (Tenant-Aware)
- `GET /api/v1/customers` - List customers for current tenant
- `GET /api/v1/inventory` - List inventory for current tenant
- `GET /api/v1/received` - List work orders for current tenant

### Context Management
- `GET /api/v1/tenant/current` - Get current tenant context
- `POST /api/v1/tenant/switch` - Switch tenant (admin only)

## 🚀 Deployment

### Docker Setup
```bash
# Build image
docker build -t oil-gas-inventory .

# Run with PostgreSQL
docker-compose up -d
```

### Production Considerations
- Set `APP_ENV=production`
- Use connection pooling
- Enable PostgreSQL logging
- Set up backup procedures
- Configure SSL/TLS

## 📈 Performance

### Database Optimization
- Indexes on tenant_id columns
- Connection pooling (25 max, 10 idle)
- Query optimization for RLS

### Expected Performance
- Work order lookup: < 25ms
- Customer search: < 50ms  
- Material search: < 100ms
- State transitions: < 200ms

## 🔄 Migration History

### Phase 1 ✅
Legacy MDB data migration and normalization.

### Phase 2 ✅  
Go backend structure and basic API framework.

### Phase 3.5 ✅ (Current)
Multi-tenant architecture with Row-Level Security.

### Phase 4 (Planned)
Business logic implementation and workflow management.

### Phase 5 (Planned)
Vue.js frontend and customer portal.

## 🤝 Contributing

1. **Fork the repository**
2. **Create feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit changes** (`git commit -m 'Add amazing feature'`)
4. **Push to branch** (`git push origin feature/amazing-feature`)
5. **Open Pull Request**

### Development Guidelines
- Follow Go conventions and best practices
- Add tests for new functionality
- Update documentation for API changes
- Ensure tenant isolation for new features

## 📝 License

This project is proprietary software for internal use.

## 📞 Support

For questions or issues:
- Check existing issues in the repository
- Review the documentation in `docs/`
- Contact the development team

---

## 🎯 Current Status

✅ **Multi-tenant database architecture**  
✅ **Row-Level Security implementation**  
✅ **Basic API framework**  
✅ **Tenant management system**  
✅ **Customer-tenant relationships**  
✅ **Development environment setup**  

🔄 **In Progress**: Business logic implementation (Phase 4)  
📅 **Next**: Frontend development (Phase 5)

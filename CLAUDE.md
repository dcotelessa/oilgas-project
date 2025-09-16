# Claude Code Instructions - Oil & Gas Multi-Tenant Inventory System

## 🎯 Project Overview

**Business**: Multi-tenant oil & gas service provider managing inventory across multiple yard locations with comprehensive transport logistics tracking. Each tenant operates independently with enterprise-wide consolidation for reporting and work order management.

**Architecture**: Domain-driven design with separate databases per tenant location and centralized authentication.

## 📋 Core Development Principles

### 1. **Minimal Comments Philosophy**
- **File header**: Always include full pathname comment at top of file
- Only comment to describe **purpose** or **TODO** items  
- Only comment **non-obvious variables** or complex business logic
- Self-documenting function/variable names preferred
- No comments for obvious operations or standard patterns

```go
// ✅ GOOD - File header with full path
// backend/internal/customer/service.go

// ✅ GOOD - Describes business purpose
// generateWorkOrderNumber creates unique WO numbers per tenant per year
func generateWorkOrderNumber(tenantID string) string

// ❌ AVOID - Obvious from naming
// GetCustomerByID gets a customer by ID
func GetCustomerByID(id int) *Customer
```

### 2. **Cost-Effectiveness & Alternatives**
- **Single-instance deployment**: In-memory cache acceptable (not Redis)
- **Simple solutions first**: Prefer built-in Go features over external dependencies
- **Modular design**: Interface-based for easy swapping of implementations
- **Performance considerations**: Document alternatives for future scaling

```go
// ✅ GOOD - Interface allows swapping implementations
type CacheService interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{})
}

// Current: In-memory cache for single-instance
// Alternative: Redis for distributed caching when multi-instance needed
```

### 3. **File Organization Standards**
Break apart large files following domain boundaries:

```
backend/internal/{domain}/
├── models.go          # Domain entities and types
├── repository.go      # Data access layer
├── service.go         # Business logic
├── handlers.go        # HTTP handlers
├── utils.go           # Domain-specific utilities
├── cache.go           # Performance/caching layer
├── validation.go      # Input validation logic
├── middleware.go      # Domain-specific middleware
└── integration.go     # Cross-domain operations
```

### 4. **Testing Requirements**
- **Unit tests**: Focus on business logic reliability
- **Integration tests**: Database interactions
- **Performance tests**: Benchmarks for critical paths
- **Edge cases**: Error handling and cleanup scenarios
- Mock external dependencies for isolation

## 🏗️ System Architecture

### Database Structure
```
Central Auth Database (auth_central):
├── users                    # Cross-tenant users
├── tenants                  # Tenant registry  
├── user_tenant_access      # User-tenant permissions
├── customer_contacts       # Links auth users to tenant customers
├── sessions                # Cross-tenant sessions
└── permissions             # Role-based permissions

Tenant-Specific Databases:
├── location_longbeach_db    # Long Beach location
│   ├── customers           # Oil & gas companies
│   ├── inventory           # Pipes, casings, equipment
│   ├── workorders          # Service workflow
│   ├── invoices            # Billing
│   └── audit_trail         # Change tracking
├── location_bakersfield_db # Bakersfield location (same schema)
└── location_colorado_db    # Colorado location (future)
```

### Domain Architecture
- **Customer Domain**: Customer management with tenant isolation
- **Auth Domain**: Multi-tenant authentication and authorization
- **Inventory Domain**: Equipment tracking with transport logistics
- **Work Order Domain**: Service workflow management
- **Shared Domain**: Cross-cutting utilities and database management

## 🔧 Implementation Guidelines

### Domain Service Pattern
```go
type Service interface {
    // Core CRUD operations
    GetEntity(ctx context.Context, tenantID string, id int) (*Entity, error)
    CreateEntity(ctx context.Context, tenantID string, entity *Entity) error
    
    // Business operations
    ProcessBusinessLogic(ctx context.Context, params BusinessParams) error
}

type service struct {
    repo      Repository
    cache     CacheService
    validator ValidationService
}
```

### Multi-Tenant Database Access
```go
type DatabaseManager struct {
    centralDB *sql.DB
    tenantDBs map[string]*sql.DB
}

func (dm *DatabaseManager) GetTenantDB(tenantID string) (*sql.DB, error) {
    // Always validate tenant access before returning DB connection
}
```

### Error Handling Standards
```go
// Return structured errors with context
func (s *service) ProcessOperation(ctx context.Context, params *Params) error {
    if err := s.validateInput(params); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    if err := s.repo.ProcessData(ctx, params); err != nil {
        return fmt.Errorf("failed to process data: %w", err)
    }
    
    return nil
}
```

## 📊 Business Logic Requirements

### Customer Domain
- **Multi-tenant isolation**: Each tenant's customers completely separate
- **Contact management**: Link central auth users to tenant customers
- **Validation**: Company codes (alphanumeric, 2-10 chars), tax IDs, required fields
- **Analytics**: Work order counts, revenue totals, customer health scores

### Inventory Domain (Next Phase)
- **Item types**: Pipes, casings, equipment
- **Location tracking**: Yard, rack, level
- **Status management**: Available, allocated, in_service, returned, scrapped
- **Transport logistics**: Trucking company, trailer ID, delivery status
- **Condition tracking**: Excellent, good, fair, poor

### Work Order Domain (Next Phase)
- **Service workflow**: Pending → In Progress → Completed → Invoiced
- **Customer integration**: Link to tenant customers
- **Inventory integration**: Track item allocation and status changes
- **Invoice generation**: Service billing calculations

### Auth Domain
- **Cross-tenant users**: Enterprise users can access multiple locations
- **Customer contacts**: Restricted to their own company data
- **Role-based permissions**: Admin, Manager, Operator, Viewer, Customer Contact
- **Session management**: Secure JWT with proper expiration

## 🧪 Testing Standards

### Unit Test Structure
```go
func TestService_BusinessOperation(t *testing.T) {
    // Setup mocks
    mockRepo := &mockRepository{}
    mockCache := &mockCacheService{}
    service := NewService(mockRepo, mockCache)
    
    // Test cases
    testCases := []struct {
        name        string
        input       *Input
        setupMocks  func()
        expectError bool
        expected    *Expected
    }{
        {
            name: "successful operation",
            input: validInput,
            setupMocks: func() {
                mockRepo.On("Method", mock.Anything).Return(expectedResult, nil)
            },
            expectError: false,
            expected: expectedOutput,
        },
        {
            name: "validation failure",
            input: invalidInput,
            expectError: true,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Test Requirements
- Database transaction rollback for test isolation
- Test multi-tenant data isolation
- Verify cross-domain operations work correctly
- Performance benchmarks for critical operations

## 🚀 Current Implementation Status

### ✅ Completed
- **Multi-database architecture**: Central auth + tenant databases
- **Customer Domain**: CRUD operations, contact management, caching
- **Database migrations**: Versioned migrations with rollback capability
- **Development tools**: Makefile, Docker setup, migration utilities

### 🔄 In Progress
- **Code cleanup**: Apply principles to existing customer domain
- **Testing enhancement**: Comprehensive test coverage
- **Documentation**: API documentation and usage examples

### 📋 Next Priorities
1. **Work Order Domain**: Service workflow management
2. **Inventory Integration**: Connect work orders to inventory tracking
3. **Invoice Generation**: Service billing calculations
4. **Multi-location expansion**: Bakersfield and Colorado setups
5. **Frontend Development**: User interface for tenant operations
6. **CalOPPA Compliance**: Privacy policy and data handling requirements for California residents

## 🎯 Claude Code Tasks

### Immediate Cleanup Tasks
1. **Review customer domain files** for principle compliance:
   - Remove unnecessary comments
   - Ensure proper file organization
   - Add missing unit tests
   - Document cost-effective alternatives

2. **Standardize error handling** across all domains

3. **Implement comprehensive testing** for reliability

4. **Create utility functions** to reduce code duplication

### Code Quality Checks
- [ ] Minimal comments (purpose/TODO only)
- [ ] Cost-effective solutions documented
- [ ] Proper file organization by domain
- [ ] Comprehensive unit test coverage
- [ ] Multi-tenant isolation validated
- [ ] Performance considerations documented
- [ ] Interface-based design for flexibility

### Business Logic Validation
- [ ] Customer validation rules implemented correctly
- [ ] Multi-tenant data isolation enforced
- [ ] Auth integration working properly
- [ ] Database migration scripts tested
- [ ] Error handling provides useful context
- [ ] Caching strategy appropriate for single-instance deployment

## 📝 Development Workflow

1. **Before making changes**: Review existing tests and add missing coverage
2. **During development**: Follow minimal comment and file organization principles
3. **After changes**: Run full test suite and verify multi-tenant isolation
4. **Performance**: Benchmark critical operations and document alternatives
5. **Documentation**: Update API documentation if interfaces change

## 🔍 Code Review Checklist

- **Comments**: Only for business purpose or non-obvious logic
- **Organization**: Domain files properly separated
- **Testing**: Unit tests for all business logic
- **Performance**: Efficient database queries and caching
- **Security**: Multi-tenant isolation enforced
- **Error handling**: Structured errors with proper context
- **Interfaces**: Flexible design for future enhancements

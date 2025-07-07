# Integration Tests

Comprehensive integration test suite for the Oil & Gas Inventory System backend.

## Overview

This integration test suite covers:

- **Repository Integration**: Database operations and data consistency
- **Service Integration**: Business logic with caching and validation  
- **API Integration**: HTTP endpoints and request/response handling
- **Workflow Integration**: Complete business workflows from end-to-end
- **Performance Testing**: Concurrent operations and load testing
- **Error Handling**: Edge cases and failure scenarios

## Test Structure

```
test/integration/
├── full_integration_test.go      # Comprehensive workflow tests
├── service_integration_test.go   # Service layer tests
├── api_integration_test.go       # HTTP API tests  
├── runner_test.go               # Test orchestration
├── testdata/
│   └── fixtures.go              # Reusable test data
├── helpers.go                   # Common test utilities
├── docker-compose.test.yml      # Test database setup
├── .env.test                    # Test environment config
├── Makefile                     # Test automation
└── README.md                    # This file
```

## Running Tests

### Prerequisites

1. Docker and Docker Compose
2. Go 1.19+
3. Test database setup

### Quick Start

```bash
# Run all integration tests
make test-integration

# Run with coverage
make test-integration-coverage

# Run with race detection
make test-integration-race

# Full test cycle (setup DB, run tests, cleanup)
make test-integration-full
```

### Manual Setup

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Set environment
export TEST_DATABASE_URL="postgres://postgres:test123@localhost:5433/oilgas_inventory_test"

# Run migrations
go run ../../migrator.go migrate test

# Run tests
go test -v ./... -tags=integration

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

### Test Categories

```bash
# Repository tests only
go test -v -run TestRepositoryIntegration

# Service tests only  
go test -v -run TestServiceIntegration

# API tests only
go test -v -run TestAPIIntegration

# Complete workflow tests
go test -v -run TestCompleteWorkflow
```

## Test Data Management

### Fixtures

The test suite uses standardized fixtures for consistent data:

- **Customers**: 3 standard oil & gas companies
- **Grades**: J55, L80, N80, P110, etc.
- **Sizes**: Common pipe sizes (5 1/2", 7", 9 5/8")
- **Workflows**: Complete received → inspection → production cycles

### Data Isolation

Each test:
- Gets a clean database state
- Uses unique identifiers to prevent conflicts
- Cleans up after completion
- Runs in isolated transactions where possible

## Key Test Scenarios

### 1. Complete Workflow Integration
Tests the full business process:
```
Customer → Receive Inventory → Inspect → Move to Production → Analytics
```

### 2. Concurrent Operations
Validates system behavior under load:
- Multiple simultaneous API requests
- Concurrent database operations
- Cache consistency under pressure

### 3. Data Consistency
Ensures referential integrity:
- Customer references in all related tables
- Workflow state consistency
- Analytics data accuracy

### 4. Error Handling
Tests failure scenarios:
- Invalid data validation
- Duplicate key violations
- Network/database failures
- Malformed API requests

### 5. Performance Benchmarks
Establishes performance baselines:
- API response times
- Database query performance
- Cache hit/miss ratios

## Continuous Integration

The test suite is designed for CI/CD:

```bash
# CI-optimized test run
make test-ci
```

This includes:
- Parallel test execution
- Race condition detection
- Coverage reporting
- Fast feedback on failures

## Test Configuration

### Environment Variables

- `TEST_DATABASE_URL`: Test database connection
- `APP_ENV=test`: Enables test mode
- `TEST_TIMEOUT`: Maximum test duration
- `TEST_PARALLEL_WORKERS`: Concurrent test count

### Database Schema

Tests use the same schema as production but with:
- Isolated test database
- Faster cleanup procedures
- Additional test-specific indexes

## Troubleshooting

### Common Issues

1. **Database Connection Failures**
   ```bash
   # Check if test DB is running
   docker-compose -f docker-compose.test.yml ps
   
   # View logs
   docker-compose -f docker-compose.test.yml logs postgres-test
   ```

2. **Test Data Conflicts**
   ```bash
   # Clean test database
   go run ../../migrator.go clean test
   go run ../../migrator.go migrate test
   ```

3. **Slow Test Performance**
   ```bash
   # Run with profiling
   go test -v -cpuprofile=cpu.prof -memprofile=mem.prof
   ```

### Debug Mode

Enable verbose logging:
```bash
export LOG_LEVEL=debug
go test -v ./... -tags=integration
```

## Best Practices

1. **Test Isolation**: Each test should be independent
2. **Realistic Data**: Use production-like test data
3. **Error Coverage**: Test both success and failure paths
4. **Performance Awareness**: Monitor test execution time
5. **Clear Assertions**: Make test failures easy to diagnose

## Contributing

When adding new integration tests:

1. Follow the existing test structure
2. Use the fixture system for test data
3. Include both positive and negative test cases
4. Add performance considerations for new features
5. Update this README for significant changes



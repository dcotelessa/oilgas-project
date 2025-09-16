// backend/README_STARTUP.md
# 🚀 Quick Startup Guide

## After System Reset

1. **Environment Setup:**
```bash
# Copy environment file
cp .env.example .env
# Update .env if needed (defaults work for development)
```

2. **Start Development Environment:**
```bash
# Full setup (databases + migrations + seed data)
make dev-setup

# Or step by step:
make db-up        # Start databases
make migrate-up   # Run migrations  
make seed         # Add test data
```

3. **Start Application:**
```bash
make dev
```

4. **Verify Setup:**
```bash
# Check database status
make db-status

# Test API
curl -H "Content-Type: application/json" \
     -X POST http://localhost:8080/api/v1/auth/login \
     -d '{"email":"admin@oilgas.com","password":"admin123"}'
```

## Common Issues & Solutions

**Error: "undefined: auth.SessionManager"**
- ✅ Fixed with new auth handler structure

**Database connection issues:**
```bash
make db-down
make db-up
```

**Migration issues:**
```bash
make migrate-status  # Check current status
make migrate-up      # Run pending migrations
```

**Port conflicts:**
- Auth DB: 5432 (can change AUTH_DB_PORT in .env)
- LongBeach DB: 5433 (can change LONGBEACH_DB_PORT in .env)  
- App: 8080 (can change APP_PORT in .env)

## Available Commands

```bash
make dev          # Start development server
make test         # Run tests
make db-status    # Check database health
make db-shell-longbeach  # Access Long Beach DB
make db-shell-auth       # Access Auth DB
```

## Next Steps

1. ✅ Basic setup working
2. 🔄 Migrate your Access customers  
3. 🏗️ Build Work Order domain
4. 📊 Add inventory integration

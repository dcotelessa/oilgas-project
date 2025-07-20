#!/bin/bash
# scripts/generators/database_schema_generator.sh
# Generates tenant-aware database schema with RLS for Phase 3

set -e
echo "🗄️ Generating tenant-aware database schema..."

# Detect backend directory
BACKEND_DIR=""
if [ -d "backend" ] && [ -f "backend/go.mod" ]; then
    BACKEND_DIR="backend/"
elif [ -f "go.mod" ]; then
    BACKEND_DIR=""
else
    echo "❌ Error: Cannot find go.mod file"
    exit 1
fi

# Create directories
mkdir -p "${BACKEND_DIR}migrations"
mkdir -p "${BACKEND_DIR}seeds"

# Generate tenant-aware migration
cat > "${BACKEND_DIR}migrations/002_tenant_support.sql" << 'EOF'
-- Tenant-Aware Migration for Oil & Gas Inventory System
BEGIN;

-- Create tenant management table
CREATE TABLE IF NOT EXISTS store.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    database_type VARCHAR(20) DEFAULT 'shared' CHECK (database_type IN ('shared', 'schema', 'dedicated')),
    schema_name VARCHAR(100),
    database_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Add tenant_id to existing tables
ALTER TABLE store.customers ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES store.tenants(id);
ALTER TABLE store.inventory ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE store.received ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Create auth schema and tables
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'operator', 'customer_manager', 'customer_user', 'viewer')),
    company VARCHAR(255),
    tenant_id UUID NOT NULL REFERENCES store.tenants(id),
    email_verified BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(email, tenant_id)
);

CREATE TABLE IF NOT EXISTS auth.sessions (
    id VARCHAR(64) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES auth.users(id),
    tenant_id UUID NOT NULL REFERENCES store.tenants(id),
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    company VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Insert default tenant
INSERT INTO store.tenants (name, slug, database_type, is_active) 
VALUES ('Default Tenant', 'default', 'shared', true)
ON CONFLICT (slug) DO NOTHING;

-- Update existing data with default tenant
UPDATE store.customers SET tenant_id = (
    SELECT id FROM store.tenants WHERE slug = 'default'
) WHERE tenant_id IS NULL;

-- Enable Row-Level Security
ALTER TABLE store.customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE store.inventory ENABLE ROW LEVEL SECURITY;
ALTER TABLE store.received ENABLE ROW LEVEL SECURITY;

-- Create authenticated_users role for RLS
DO $$ BEGIN
    CREATE ROLE authenticated_users;
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- RLS Policies
CREATE POLICY tenant_isolation_customers ON store.customers
FOR ALL TO authenticated_users
USING (
    current_setting('app.user_role', true) IN ('admin', 'operator')
    OR tenant_id::text = current_setting('app.tenant_id', true)
);

CREATE POLICY tenant_isolation_inventory ON store.inventory
FOR ALL TO authenticated_users
USING (
    current_setting('app.user_role', true) IN ('admin', 'operator')
    OR tenant_id::text = current_setting('app.tenant_id', true)
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON store.customers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_id ON store.inventory(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email_tenant ON auth.users(email, tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON auth.sessions(expires_at) WHERE deleted_at IS NULL;

-- Migration tracking
INSERT INTO migrations.schema_migrations (version, name) 
VALUES ('002', 'Add tenant awareness and RLS')
ON CONFLICT (version) DO NOTHING;

COMMIT;
EOF

echo "✅ Tenant-aware database schema generated"
echo "   - Complete tenant management tables"
echo "   - Row-Level Security policies"
echo "   - Performance indexes"
echo "   - Auth schema with sessions"

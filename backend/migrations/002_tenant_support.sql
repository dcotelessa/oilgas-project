-- Tenant-Aware Migration for Oil & Gas Inventory System
-- Phase 3: Complete tenant isolation with Row-Level Security

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

-- Create auth schema and tables
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
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
    deleted_at TIMESTAMP
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

-- Add tenant_id to existing tables
ALTER TABLE store.customers ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES store.tenants(id);
ALTER TABLE store.inventory ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES store.tenants(id);
ALTER TABLE store.received ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES store.tenants(id);

-- Create default tenant for existing data
INSERT INTO store.tenants (name, slug, database_type, is_active) 
VALUES ('Default Tenant', 'default', 'shared', true)
ON CONFLICT (slug) DO NOTHING;

-- Update existing data with default tenant
UPDATE store.customers SET tenant_id = (
    SELECT id FROM store.tenants WHERE slug = 'default'
) WHERE tenant_id IS NULL;

UPDATE store.inventory SET tenant_id = (
    SELECT id FROM store.tenants WHERE slug = 'default'
) WHERE tenant_id IS NULL;

UPDATE store.received SET tenant_id = (
    SELECT id FROM store.tenants WHERE slug = 'default'
) WHERE tenant_id IS NULL;

-- Create authenticated_users role if it doesn't exist
DO $body$
BEGIN
    CREATE ROLE authenticated_users;
EXCEPTION
    WHEN duplicate_object THEN null;
END $body$;

-- Enable Row-Level Security
ALTER TABLE store.customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE store.inventory ENABLE ROW LEVEL SECURITY;
ALTER TABLE store.received ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist
DROP POLICY IF EXISTS tenant_isolation_customers ON store.customers;
DROP POLICY IF EXISTS tenant_isolation_inventory ON store.inventory;
DROP POLICY IF EXISTS tenant_isolation_received ON store.received;

-- Create comprehensive RLS policies
CREATE POLICY tenant_isolation_customers ON store.customers
FOR ALL TO authenticated_users, postgres
USING (
    current_setting('app.user_role', true) IN ('admin', 'operator')
    OR 
    tenant_id::text = current_setting('app.tenant_id', true)
);

CREATE POLICY tenant_isolation_inventory ON store.inventory
FOR ALL TO authenticated_users, postgres
USING (
    current_setting('app.user_role', true) IN ('admin', 'operator')
    OR 
    tenant_id::text = current_setting('app.tenant_id', true)
);

CREATE POLICY tenant_isolation_received ON store.received
FOR ALL TO authenticated_users, postgres
USING (
    current_setting('app.user_role', true) IN ('admin', 'operator')
    OR 
    tenant_id::text = current_setting('app.tenant_id', true)
);

-- Create performance indexes
CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON store.customers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_id ON store.inventory(tenant_id);
CREATE INDEX IF NOT EXISTS idx_received_tenant_id ON store.received(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON auth.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON auth.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON auth.sessions(expires_at);

COMMIT;

-- Migration 003: Add Tenant Architecture
BEGIN;

-- Create tenants table
CREATE TABLE IF NOT EXISTS store.tenants (
    tenant_id SERIAL PRIMARY KEY,
    tenant_name VARCHAR(255) NOT NULL UNIQUE,
    tenant_slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    contact_email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    active BOOLEAN DEFAULT TRUE,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add tenant_id to existing tables
ALTER TABLE store.users ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES store.tenants(tenant_id);
ALTER TABLE store.inventory ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES store.tenants(tenant_id);
ALTER TABLE store.received ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES store.tenants(tenant_id);

-- Customer-tenant relationship
CREATE TABLE IF NOT EXISTS store.customer_tenant_assignments (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES store.customers(customer_id) ON DELETE CASCADE,
    tenant_id INTEGER REFERENCES store.tenants(tenant_id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) DEFAULT 'primary',
    active BOOLEAN DEFAULT TRUE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by INTEGER REFERENCES store.users(user_id),
    notes TEXT,
    UNIQUE(customer_id, tenant_id)
);

-- User roles within tenants
CREATE TABLE IF NOT EXISTS store.user_tenant_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES store.users(user_id) ON DELETE CASCADE,
    tenant_id INTEGER REFERENCES store.tenants(tenant_id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by INTEGER REFERENCES store.users(user_id),
    active BOOLEAN DEFAULT TRUE,
    UNIQUE(user_id, tenant_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON store.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_id ON store.inventory(tenant_id);
CREATE INDEX IF NOT EXISTS idx_received_tenant_id ON store.received(tenant_id);

COMMIT;

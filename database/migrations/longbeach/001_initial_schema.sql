-- database/migrations/longbeach/001_initial_schema.sql
-- Migration: 001_initial_schema.sql
-- Description: Initial Long Beach tenant schema - Customer domain focused
-- Up Migration

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create store schema
CREATE SCHEMA IF NOT EXISTS store;

-- Tenants table
CREATE TABLE store.tenants (
    tenant_id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert Long Beach tenant
INSERT INTO store.tenants (tenant_id, name) VALUES ('longbeach', 'Long Beach Operations') ON CONFLICT DO NOTHING;

-- Customers table (matches Customer domain model exactly)
CREATE TABLE store.customers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL DEFAULT 'longbeach',
    name VARCHAR(255) NOT NULL,
    company_code VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    tax_id VARCHAR(50),
    payment_terms VARCHAR(50) DEFAULT 'NET30',
    billing_street VARCHAR(255),
    billing_city VARCHAR(100),
    billing_state VARCHAR(10),
    billing_zip_code VARCHAR(20),
    billing_country VARCHAR(10) DEFAULT 'US',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (tenant_id) REFERENCES store.tenants(tenant_id)
);

-- Customer contacts - links customers to central auth users
CREATE TABLE store.customer_contacts (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES store.customers(id) ON DELETE CASCADE,
    auth_user_id INTEGER NOT NULL, -- References auth.users(id) - enforced at app level
    contact_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY',
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    
    -- Cached contact info from auth database for performance
    full_name VARCHAR(255),
    email VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT chk_contact_type CHECK (contact_type IN (
        'PRIMARY', 'BILLING', 'SHIPPING', 'APPROVER'
    )),
    UNIQUE(customer_id, auth_user_id)
);

-- Customer domain audit trail
CREATE TABLE store.customer_audit (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values JSONB,
    new_values JSONB,
    changed_by_user_id INTEGER, -- References auth.users(id)
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Performance indexes optimized for Customer domain queries
CREATE INDEX idx_customers_tenant_active ON store.customers(tenant_id, is_active) WHERE is_active = true;
CREATE INDEX idx_customers_name_search ON store.customers USING gin(to_tsvector('english', name));
CREATE INDEX idx_customers_company_code ON store.customers(tenant_id, company_code) WHERE company_code IS NOT NULL;
CREATE INDEX idx_customers_status ON store.customers(status);

-- Customer contacts indexes
CREATE INDEX idx_customer_contacts_customer ON store.customer_contacts(customer_id) WHERE is_active = true;
CREATE INDEX idx_customer_contacts_auth_user ON store.customer_contacts(auth_user_id);
CREATE INDEX idx_customer_contacts_type ON store.customer_contacts(contact_type, is_active);
CREATE INDEX idx_customer_contacts_primary ON store.customer_contacts(is_primary) WHERE is_primary = true;

-- Customer audit indexes
CREATE INDEX idx_customer_audit_customer_date ON store.customer_audit(customer_id, created_at DESC);
CREATE INDEX idx_customer_audit_user ON store.customer_audit(changed_by_user_id);
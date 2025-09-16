-- database/init/bakersfield/01_bakersfield_schema.sql
-- Bakersfield Tenant Database Schema
-- This runs when bakersfield-db container starts

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create store schema
CREATE SCHEMA IF NOT EXISTS store;

-- Tenants table (single tenant for Bakersfield)
CREATE TABLE store.tenants (
    tenant_id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert Bakersfield tenant
INSERT INTO store.tenants (tenant_id, name) VALUES ('bakersfield', 'Bakersfield Operations') ON CONFLICT DO NOTHING;

-- Customers table (updated to match current domain model)
CREATE TABLE store.customers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL DEFAULT 'bakersfield',
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

-- NOTE: Inventory table removed - not part of Customer domain
-- Inventory will be implemented in separate Inventory domain
-- For now, work order items can reference inventory by external ID

-- NOTE: Work orders table removed - not part of current Customer domain focus
-- Work orders will be implemented in separate WorkOrder domain
-- Customer analytics can reference external work order data

-- Customer contacts bridge (for customer portal access)
CREATE TABLE store.customer_contacts (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    contact_type VARCHAR(50) DEFAULT 'primary' CHECK (contact_type IN ('primary', 'billing', 'shipping', 'emergency')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (customer_id) REFERENCES store.customers(id) ON DELETE CASCADE
);

-- Audit trail for customer changes
CREATE TABLE store.customer_audit (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER,
    action VARCHAR(20) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    changed_by_user_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_customers_tenant_id ON store.customers(tenant_id);
CREATE INDEX idx_customers_active ON store.customers(is_active) WHERE is_active = true;
CREATE INDEX idx_customers_company_code ON store.customers(company_code) WHERE company_code IS NOT NULL;
CREATE INDEX idx_customers_status ON store.customers(status);
CREATE INDEX idx_customers_name_search ON store.customers USING gin(to_tsvector('english', name));

CREATE INDEX idx_customer_contacts_customer ON store.customer_contacts(customer_id);
CREATE INDEX idx_customer_contacts_email ON store.customer_contacts(contact_email) WHERE contact_email IS NOT NULL;
CREATE INDEX idx_customer_contacts_active ON store.customer_contacts(is_active) WHERE is_active = true;

CREATE INDEX idx_customer_audit_customer_date ON store.customer_audit(customer_id, created_at DESC);
CREATE INDEX idx_customer_audit_user ON store.customer_audit(changed_by_user_id);
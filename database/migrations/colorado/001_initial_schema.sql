-- database/migrations/colorado/001_initial_schema.sql
-- Migration: 001_initial_schema.sql
-- Description: Initial Colorado tenant schema - Customer domain focused
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

-- Insert Colorado tenant
INSERT INTO store.tenants (tenant_id, name) VALUES ('colorado', 'Colorado Operations') ON CONFLICT DO NOTHING;

-- Customers table (matches Customer domain model exactly)
CREATE TABLE store.customers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL DEFAULT 'colorado',
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

-- Customer contacts for customer portal access
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
CREATE INDEX idx_customers_billing_state ON store.customers(billing_state) WHERE billing_state IS NOT NULL;
CREATE INDEX idx_customers_payment_terms ON store.customers(payment_terms);

CREATE INDEX idx_customer_contacts_customer ON store.customer_contacts(customer_id);
CREATE INDEX idx_customer_contacts_email ON store.customer_contacts(contact_email) WHERE contact_email IS NOT NULL;
CREATE INDEX idx_customer_contacts_active ON store.customer_contacts(is_active) WHERE is_active = true;

CREATE INDEX idx_customer_audit_customer_date ON store.customer_audit(customer_id, created_at DESC);
CREATE INDEX idx_customer_audit_user ON store.customer_audit(changed_by_user_id);
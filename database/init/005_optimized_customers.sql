-- database/init/005_optimized_customers.sql
-- Optimized customer and workorder tables

\echo 'Creating optimized customer and workorder tables...'

-- Optimized customers table (builds on your customers_standardized)
CREATE TABLE IF NOT EXISTS store.customers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL DEFAULT 'local-dev',
    name VARCHAR(255) NOT NULL,
    company_code VARCHAR(50),
    address TEXT,
    phone VARCHAR(20),
    email VARCHAR(255),
    billing_contact_name VARCHAR(255),
    billing_contact_email VARCHAR(255),
    billing_contact_phone VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Reference to your Access-imported data
    original_customer_id INTEGER REFERENCES store.customers_standardized(customer_id),
    
    CONSTRAINT fk_customers_tenant FOREIGN KEY (tenant_id) REFERENCES store.tenants(tenant_id)
);

-- Customer-Auth contacts junction table
CREATE TABLE IF NOT EXISTS store.customer_auth_contacts (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES store.customers(id) ON DELETE CASCADE,
    auth_user_id INTEGER NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    contact_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY',
    yard_permissions TEXT[] DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(customer_id, auth_user_id)
);

-- Work orders table (foundation for invoicing)
CREATE TABLE IF NOT EXISTS store.workorders (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    customer_id INTEGER NOT NULL REFERENCES store.customers(id),
    work_order_number VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    description TEXT NOT NULL,
    total_amount DECIMAL(10,2),
    created_by_user_id INTEGER NOT NULL REFERENCES auth.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_workorders_tenant FOREIGN KEY (tenant_id) REFERENCES store.tenants(tenant_id),
    UNIQUE(tenant_id, work_order_number)
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_customers_tenant_active ON store.customers(tenant_id, is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_customers_name_gin ON store.customers USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_workorders_customer_status ON store.workorders(customer_id, status);

-- Migration function to populate from customers_standardized
CREATE OR REPLACE FUNCTION migrate_customers_from_standardized() RETURNS void AS $$
BEGIN
    -- Only migrate if optimized table is empty and standardized has data
    IF (SELECT COUNT(*) FROM store.customers) = 0 AND 
       (SELECT COUNT(*) FROM store.customers_standardized WHERE NOT is_deleted) > 0 THEN
        
        INSERT INTO store.customers (
            tenant_id, name, company_code, address, phone, email,
            billing_contact_name, billing_contact_email, billing_contact_phone,
            is_active, created_at, updated_at, original_customer_id
        )
        SELECT 
            COALESCE(cs.tenant_id, 'local-dev'),
            cs.customer_name,
            cs.original_customer_id,
            CONCAT_WS(', ', 
                NULLIF(cs.billing_address, ''), 
                NULLIF(cs.billing_city, ''), 
                NULLIF(cs.billing_state, ''), 
                NULLIF(cs.billing_zip_code, '')
            ),
            cs.phone_number,
            cs.email_address,
            cs.contact_name,
            cs.email_address,
            cs.phone_number,
            NOT cs.is_deleted,
            cs.created_at,
            COALESCE(cs.updated_at, cs.created_at),
            cs.customer_id
        FROM store.customers_standardized cs
        WHERE NOT cs.is_deleted
        ORDER BY cs.customer_id;
        
        RAISE NOTICE 'Migrated % customers from standardized to optimized table', 
                     (SELECT COUNT(*) FROM store.customers);
    ELSE
        RAISE NOTICE 'Customer migration skipped - optimized table already has data or no standardized data found';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Run the migration
SELECT migrate_customers_from_standardized();

\echo 'Optimized customer and workorder tables created'

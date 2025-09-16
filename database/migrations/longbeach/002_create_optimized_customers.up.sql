-- 002_create_optimized_customers.up.sql
-- Create optimized customers table
CREATE TABLE store.customers (
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
    original_customer_id INTEGER REFERENCES store.customers_standardized(customer_id)
);

-- Add foreign key to tenants (if not already exists)
ALTER TABLE store.customers 
ADD CONSTRAINT fk_customers_tenant 
FOREIGN KEY (tenant_id) REFERENCES store.tenants(tenant_id);

-- Add foreign key from auth.users to customers
ALTER TABLE auth.users 
ADD CONSTRAINT fk_users_customer 
FOREIGN KEY (customer_id) REFERENCES store.customers(id);

-- Performance indexes
CREATE INDEX idx_customers_tenant_active ON store.customers(tenant_id, is_active) WHERE is_active = true;
CREATE INDEX idx_customers_name_search ON store.customers USING gin(to_tsvector('english', name));
CREATE INDEX idx_customers_company_code ON store.customers(tenant_id, company_code) WHERE company_code IS NOT NULL;

-- Migration function
CREATE OR REPLACE FUNCTION migrate_customers_from_standardized() RETURNS INTEGER AS $$
DECLARE
    migrated_count INTEGER := 0;
BEGIN
    -- Only migrate if optimized table is empty
    IF (SELECT COUNT(*) FROM store.customers) = 0 THEN
        INSERT INTO store.customers (
            tenant_id, name, company_code, address, phone, email,
            billing_contact_name, billing_contact_email, billing_contact_phone,
            is_active, original_customer_id, created_at, updated_at
        )
        SELECT 
            COALESCE(cs.tenant_id, 'local-dev'),
            cs.customer_name,
            cs.original_customer_id,
            NULLIF(TRIM(CONCAT_WS(', ', 
                NULLIF(cs.billing_address, ''),
                NULLIF(cs.billing_city, ''),
                NULLIF(cs.billing_state, ''),
                NULLIF(cs.billing_zip_code, '')
            )), ''),
            NULLIF(TRIM(cs.phone_number), ''),
            NULLIF(TRIM(cs.email_address), ''),
            NULLIF(TRIM(cs.contact_name), ''),
            NULLIF(TRIM(cs.email_address), ''),
            NULLIF(TRIM(cs.phone_number), ''),
            NOT COALESCE(cs.is_deleted, false),
            cs.customer_id,
            COALESCE(cs.created_at, NOW()),
            COALESCE(cs.updated_at, cs.created_at, NOW())
        FROM store.customers_standardized cs
        WHERE NOT COALESCE(cs.is_deleted, false)
        ORDER BY cs.customer_id;
        
        GET DIAGNOSTICS migrated_count = ROW_COUNT;
    END IF;
    
    RETURN migrated_count;
END;
$$ LANGUAGE plpgsql;

-- Run migration
SELECT migrate_customers_from_standardized() AS migrated_customers_count;

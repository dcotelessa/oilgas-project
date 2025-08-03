-- 003_create_customers_standardized.sql
-- Standardized customer table with proper multi-tenant support

\echo 'Creating customers table...'

CREATE TABLE IF NOT EXISTS store.customers_standardized (
    customer_id SERIAL PRIMARY KEY,
    original_customer_id VARCHAR(50),
    tenant_id VARCHAR(100) NOT NULL DEFAULT 'local-dev',
    
    -- Company Information
    customer_name VARCHAR(255) NOT NULL,
    
    -- Billing Address
    billing_address TEXT,
    billing_city VARCHAR(100),
    billing_state VARCHAR(50),
    billing_zip_code VARCHAR(20),
    
    -- Contact Information
    contact_name VARCHAR(255),
    phone_number VARCHAR(50),
    fax_number VARCHAR(50),
    email_address VARCHAR(255),
    
    -- Color Grades (1-5)
    color_grade_1 VARCHAR(50),
    color_grade_2 VARCHAR(50),
    color_grade_3 VARCHAR(50),
    color_grade_4 VARCHAR(50),
    color_grade_5 VARCHAR(50),
    
    -- Wall Losses (1-5)
    wall_loss_1 DECIMAL(10,4),
    wall_loss_2 DECIMAL(10,4),
    wall_loss_3 DECIMAL(10,4),
    wall_loss_4 DECIMAL(10,4),
    wall_loss_5 DECIMAL(10,4),
    
    -- W-String Colors (1-5)
    wstring_color_1 VARCHAR(50),
    wstring_color_2 VARCHAR(50),
    wstring_color_3 VARCHAR(50),
    wstring_color_4 VARCHAR(50),
    wstring_color_5 VARCHAR(50),
    
    -- W-String Losses (1-5)
    wstring_loss_1 DECIMAL(10,4),
    wstring_loss_2 DECIMAL(10,4),
    wstring_loss_3 DECIMAL(10,4),
    wstring_loss_4 DECIMAL(10,4),
    wstring_loss_5 DECIMAL(10,4),
    
    -- Metadata
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT fk_customers_tenant FOREIGN KEY (tenant_id) REFERENCES store.tenants(tenant_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON store.customers_standardized(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customers_original_id ON store.customers_standardized(original_customer_id);
CREATE INDEX IF NOT EXISTS idx_customers_name ON store.customers_standardized(customer_name);
CREATE INDEX IF NOT EXISTS idx_customers_state ON store.customers_standardized(billing_state);
CREATE INDEX IF NOT EXISTS idx_customers_active ON store.customers_standardized(tenant_id, is_deleted) WHERE is_deleted = FALSE;

-- Enable Row Level Security
ALTER TABLE store.customers_standardized ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for tenant isolation
DROP POLICY IF EXISTS tenant_isolation_policy ON store.customers_standardized;
CREATE POLICY tenant_isolation_policy ON store.customers_standardized
    FOR ALL 
    USING (
        tenant_id = current_setting('app.current_tenant_id', true)
        OR current_setting('app.current_tenant_id', true) IS NULL  -- Allow when no tenant set (admin operations)
    );

-- Create trigger for updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_customers_updated_at ON store.customers_standardized;
CREATE TRIGGER trigger_customers_updated_at
    BEFORE UPDATE ON store.customers_standardized
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Grant permissions to current database user (works for both dev and test)
GRANT SELECT, INSERT, UPDATE, DELETE ON store.customers_standardized TO PUBLIC;
GRANT USAGE, SELECT ON SEQUENCE store.customers_standardized_customer_id_seq TO PUBLIC;

\echo 'Customers table created successfully'

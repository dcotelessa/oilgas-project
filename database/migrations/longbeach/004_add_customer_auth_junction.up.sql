-- 004_add_customer_auth_junction.up.sql
-- Customer-Auth contacts junction table for admin contact management
CREATE TABLE store.customer_auth_contacts (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES store.customers(id) ON DELETE CASCADE,
    auth_user_id INTEGER NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    contact_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY',
    yard_permissions TEXT[] DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_by INTEGER REFERENCES auth.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT uq_customer_auth_user UNIQUE(customer_id, auth_user_id),
    CONSTRAINT chk_contact_type CHECK (contact_type IN (
        'PRIMARY', 'BILLING', 'SHIPPING', 'APPROVER'
    ))
);

-- Indexes for junction table
CREATE INDEX idx_customer_contacts_customer ON store.customer_auth_contacts(customer_id) WHERE is_active = true;
CREATE INDEX idx_customer_contacts_user ON store.customer_auth_contacts(auth_user_id) WHERE is_active = true;
CREATE INDEX idx_customer_contacts_type ON store.customer_auth_contacts(contact_type, is_active);

-- Password reset tokens table
CREATE TABLE auth.password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET
);

CREATE INDEX idx_password_reset_tokens_token ON auth.password_reset_tokens(token) WHERE used_at IS NULL;
CREATE INDEX idx_password_reset_tokens_user ON auth.password_reset_tokens(user_id);

-- Function to create customer contact
CREATE OR REPLACE FUNCTION create_customer_contact(
    p_customer_id INTEGER,
    p_email VARCHAR(255),
    p_full_name VARCHAR(255),
    p_password_hash VARCHAR(255),
    p_contact_type VARCHAR(50) DEFAULT 'PRIMARY',
    p_tenant_id VARCHAR(100) DEFAULT 'local-dev',
    p_created_by INTEGER DEFAULT NULL
) RETURNS INTEGER AS $$
DECLARE
    new_user_id INTEGER;
BEGIN
    -- Create auth user
    INSERT INTO auth.users (
        email, username, password_hash, full_name, role,
        customer_id, primary_tenant_id, tenant_access, is_active
    ) VALUES (
        p_email,
        LOWER(REPLACE(p_email, '@', '_')),
        p_password_hash,
        p_full_name,
        'CUSTOMER_CONTACT',
        p_customer_id,
        p_tenant_id,
        ARRAY[p_tenant_id],
        true
    ) RETURNING id INTO new_user_id;
    
    -- Create customer-auth relationship
    INSERT INTO store.customer_auth_contacts (
        customer_id, auth_user_id, contact_type, created_by
    ) VALUES (
        p_customer_id, new_user_id, p_contact_type, p_created_by
    );
    
    RETURN new_user_id;
END;
$$ LANGUAGE plpgsql;

-- View for customer contacts
CREATE VIEW store.customer_contact_details AS
SELECT 
    c.id as customer_id,
    c.tenant_id,
    c.name as customer_name,
    u.id as user_id,
    u.email,
    u.full_name,
    cac.contact_type,
    cac.yard_permissions,
    cac.is_active as contact_active,
    u.last_login_at,
    u.is_active as user_active
FROM store.customers c
JOIN store.customer_auth_contacts cac ON c.id = cac.customer_id
JOIN auth.users u ON cac.auth_user_id = u.id;

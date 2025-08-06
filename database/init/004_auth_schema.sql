-- database/init/004_auth_schema.sql
-- Auth schema for customer/auth/workorder domains (simplified for init)

\echo 'Creating auth schema for domains...'

-- Create schemas
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS audit;

-- Users table (integrates with your existing store.tenants)
CREATE TABLE IF NOT EXISTS auth.users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'CUSTOMER_CONTACT',
    tenant_access TEXT[] DEFAULT '{}', -- References store.tenants(tenant_id)
    primary_tenant_id VARCHAR(100) NOT NULL DEFAULT 'local-dev',
    customer_id INTEGER, -- Will reference store.customers
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_users_primary_tenant FOREIGN KEY (primary_tenant_id) REFERENCES store.tenants(tenant_id)
);

-- Sessions table
CREATE TABLE IF NOT EXISTS auth.sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(100) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_sessions_tenant FOREIGN KEY (tenant_id) REFERENCES store.tenants(tenant_id)
);

-- Events for audit trail
CREATE TABLE IF NOT EXISTS audit.events (
    id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100),
    entity_id VARCHAR(100),
    user_id INTEGER REFERENCES auth.users(id),
    event_data JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_events_tenant FOREIGN KEY (tenant_id) REFERENCES store.tenants(tenant_id)
);

-- Basic indexes
CREATE INDEX IF NOT EXISTS idx_auth_users_email ON auth.users(email);
CREATE INDEX IF NOT EXISTS idx_auth_users_tenant_access ON auth.users USING gin(tenant_access);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_id ON auth.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_type ON audit.events(tenant_id, event_type);

-- Insert system admin for testing
INSERT INTO auth.users (
    email, username, password_hash, full_name, role, 
    tenant_access, primary_tenant_id, is_active
) VALUES (
    'admin@system.local',
    'system_admin', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'System Administrator',
    'SYSTEM_ADMIN',
    '{"local-dev", "local-test"}',
    'local-dev',
    true
) ON CONFLICT (email) DO NOTHING;

\echo 'Auth schema created successfully'

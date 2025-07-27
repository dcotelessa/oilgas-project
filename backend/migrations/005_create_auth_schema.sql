-- migrations/003_create_auth_schema.sql
-- Authentication and authorization schema

-- Create auth schema
CREATE SCHEMA IF NOT EXISTS auth;

-- Create tenants table
CREATE TABLE IF NOT EXISTS auth.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    database_type VARCHAR(50) DEFAULT 'tenant',
    database_name VARCHAR(100),
    active BOOLEAN DEFAULT true,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create users table
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    company VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    last_login TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    password_changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_users_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug)
);

-- Create sessions table
CREATE TABLE IF NOT EXISTS auth.sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE,
    CONSTRAINT fk_sessions_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug)
);

-- Create user_tenant_roles table (for multi-tenant access)
CREATE TABLE IF NOT EXISTS auth.user_tenant_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    active BOOLEAN DEFAULT true,
    granted_by UUID,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_user_tenant_roles_user FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_tenant_roles_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug),
    CONSTRAINT fk_user_tenant_roles_granted_by FOREIGN KEY (granted_by) REFERENCES auth.users(id),
    CONSTRAINT unique_user_tenant UNIQUE (user_id, tenant_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON auth.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON auth.users(role);
CREATE INDEX IF NOT EXISTS idx_users_active ON auth.users(active);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON auth.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_tenant_id ON auth.sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON auth.sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_tenants_slug ON auth.tenants(slug);
CREATE INDEX IF NOT EXISTS idx_tenants_active ON auth.tenants(active);

CREATE INDEX IF NOT EXISTS idx_user_tenant_roles_user_id ON auth.user_tenant_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tenant_roles_tenant_id ON auth.user_tenant_roles(tenant_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON auth.tenants
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON auth.users
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Insert default system tenant
INSERT INTO auth.tenants (name, slug, database_type, database_name) 
VALUES ('System Administration', 'system', 'main', 'system')
ON CONFLICT (slug) DO NOTHING;

-- Create roles enum check constraints
ALTER TABLE auth.users ADD CONSTRAINT check_user_role 
    CHECK (role IN ('user', 'operator', 'manager', 'admin', 'super-admin'));

ALTER TABLE auth.user_tenant_roles ADD CONSTRAINT check_tenant_role 
    CHECK (role IN ('user', 'operator', 'manager', 'admin'));

-- Comments for documentation
COMMENT ON SCHEMA auth IS 'Authentication and authorization schema';
COMMENT ON TABLE auth.tenants IS 'Tenant organizations with database isolation';
COMMENT ON TABLE auth.users IS 'System users with tenant assignments';
COMMENT ON TABLE auth.sessions IS 'User sessions for authentication';
COMMENT ON TABLE auth.user_tenant_roles IS 'Multi-tenant role assignments';

COMMENT ON COLUMN auth.tenants.slug IS 'URL-safe tenant identifier';
COMMENT ON COLUMN auth.tenants.database_type IS 'Type: main, tenant, test';
COMMENT ON COLUMN auth.tenants.database_name IS 'Physical database name (e.g., oilgas_longbeach)';

COMMENT ON COLUMN auth.users.tenant_id IS 'Primary tenant assignment';
COMMENT ON COLUMN auth.users.failed_login_attempts IS 'Security: failed login counter';
COMMENT ON COLUMN auth.users.locked_until IS 'Security: account lock expiration';

COMMENT ON COLUMN auth.sessions.expires_at IS 'Session expiration time';
COMMENT ON COLUMN auth.sessions.last_activity IS 'Last activity for cleanup';

-- Grant necessary permissions
GRANT USAGE ON SCHEMA auth TO PUBLIC;
GRANT SELECT ON auth.tenants TO PUBLIC;
GRANT SELECT, INSERT, UPDATE ON auth.users TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.sessions TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.user_tenant_roles TO PUBLIC;

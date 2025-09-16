-- 001_create_auth_schema.up.sql
-- Central Auth Database Schema Migration

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table (cross-tenant users) - Updated to match domain model
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    
    -- Role and access control  
    role VARCHAR(50) DEFAULT 'OPERATOR' CHECK (role IN (
        'CUSTOMER_CONTACT', 'OPERATOR', 'MANAGER', 'ADMIN', 'ENTERPRISE_ADMIN', 'SYSTEM_ADMIN'
    )),
    access_level INTEGER DEFAULT 1,
    is_enterprise_user BOOLEAN DEFAULT false,
    
    -- Multi-tenant and customer relationship
    tenant_access JSONB DEFAULT '[]',
    primary_tenant_id VARCHAR(50),
    customer_id INTEGER,
    contact_type VARCHAR(50) CHECK (contact_type IN ('PRIMARY', 'BILLING', 'SHIPPING', 'APPROVER')),
    
    -- Session and activity tracking
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tenants registry
CREATE TABLE tenants (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    database_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add foreign key after tenants table is created
ALTER TABLE users ADD FOREIGN KEY (primary_tenant_id) REFERENCES tenants(id);

-- Sessions table (cross-tenant) - Enhanced for comprehensive auth
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INTEGER NOT NULL,
    tenant_id VARCHAR(50),
    token VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255),
    token_hash VARCHAR(255) NOT NULL,
    tenant_context JSONB,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    refresh_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL
);

-- User tenant access control
CREATE TABLE user_tenant_access (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    can_access BOOLEAN DEFAULT true,
    role VARCHAR(50) DEFAULT 'OPERATOR',
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    granted_by INTEGER,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id),
    UNIQUE(user_id, tenant_id)
);

-- Customer contacts (links central auth users to tenant customers) - Enhanced
CREATE TABLE customer_contacts (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    customer_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    contact_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY' CHECK (contact_type IN (
        'PRIMARY', 'BILLING', 'SHIPPING', 'APPROVER'
    )),
    yard_permissions TEXT[] DEFAULT '{}',
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(tenant_id, customer_id, user_id)
);

-- Permissions and roles
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE role_permissions (
    id SERIAL PRIMARY KEY,
    role VARCHAR(50) NOT NULL,
    permission_id INTEGER NOT NULL,
    tenant_id VARCHAR(50),
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(role, permission_id, tenant_id)
);

-- Event log for audit trail
CREATE TABLE auth_events (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    event_type VARCHAR(50) NOT NULL,
    tenant_id VARCHAR(50),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL
);

-- Password reset tokens for user management
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username) WHERE username IS NOT NULL;
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = true;
CREATE INDEX idx_users_customer ON users(customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX idx_users_tenant_access ON users USING gin(tenant_access);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_active ON sessions(is_active) WHERE is_active = true;
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
CREATE INDEX idx_sessions_token ON sessions(token_hash);
CREATE INDEX idx_sessions_refresh ON sessions(refresh_token) WHERE refresh_token IS NOT NULL;

CREATE INDEX idx_user_tenant_access_user ON user_tenant_access(user_id);
CREATE INDEX idx_user_tenant_access_tenant ON user_tenant_access(tenant_id);
CREATE INDEX idx_user_tenant_access_active ON user_tenant_access(can_access) WHERE can_access = true;

CREATE INDEX idx_customer_contacts_tenant_customer ON customer_contacts(tenant_id, customer_id);
CREATE INDEX idx_customer_contacts_user ON customer_contacts(user_id);
CREATE INDEX idx_customer_contacts_active ON customer_contacts(is_active) WHERE is_active = true;
CREATE INDEX idx_customer_contacts_primary ON customer_contacts(is_primary) WHERE is_primary = true;

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token) WHERE used_at IS NULL;
CREATE INDEX idx_password_reset_tokens_user ON password_reset_tokens(user_id);

CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX idx_role_permissions_role ON role_permissions(role);

CREATE INDEX idx_auth_events_user ON auth_events(user_id);
CREATE INDEX idx_auth_events_tenant ON auth_events(tenant_id);
CREATE INDEX idx_auth_events_type_created ON auth_events(event_type, created_at DESC);
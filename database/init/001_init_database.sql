-- database/init/001_init_database.sql
-- Initial database setup that runs when containers start

\echo 'Starting database initialization...'

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For better text search
CREATE EXTENSION IF NOT EXISTS "btree_gist"; -- For better indexing

-- Create schemas
CREATE SCHEMA IF NOT EXISTS store;
CREATE SCHEMA IF NOT EXISTS migrations;
CREATE SCHEMA IF NOT EXISTS analytics;

\echo 'Schemas created successfully'

-- Create migration tracking table
CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    execution_time_ms INTEGER DEFAULT 0
);

-- Create tenant configuration table
CREATE TABLE IF NOT EXISTS store.tenants (
    tenant_id VARCHAR(50) PRIMARY KEY,
    tenant_name VARCHAR(255) NOT NULL,
    company_name VARCHAR(255),
    database_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

\echo 'Core infrastructure tables created'

-- Insert default tenant for local development
INSERT INTO store.tenants (tenant_id, tenant_name, company_name, database_name, is_active)
VALUES 
    ('local-dev', 'Local Development', 'Development Company', 'oilgas_dev', true),
    ('local-test', 'Local Testing', 'Test Company', 'oilgas_test', true)
ON CONFLICT (tenant_id) DO NOTHING;

-- Enable Row Level Security globally (will be configured per table)
-- Note: Individual table policies will be created in migration files

\echo 'Database initialization completed successfully'

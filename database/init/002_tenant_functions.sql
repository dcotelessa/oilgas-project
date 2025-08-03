-- database/init/002_tenant_functions.sql
-- Tenant context management functions for multi-tenant architecture

\echo 'Creating tenant functions...'

CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id_param TEXT)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', tenant_id_param, false);
    PERFORM set_config('app.tenant_id', tenant_id_param, false);
    RAISE NOTICE 'Tenant context set to: %', tenant_id_param;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_current_tenant()
RETURNS TEXT AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id', true);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION clear_tenant_context()
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', NULL, false);
    PERFORM set_config('app.tenant_id', NULL, false);
    RAISE NOTICE 'Tenant context cleared';
END;
$$ LANGUAGE plpgsql;

-- Grant to current database user (works for both dev and test containers)
GRANT EXECUTE ON FUNCTION set_tenant_context(TEXT) TO PUBLIC;
GRANT EXECUTE ON FUNCTION get_current_tenant() TO PUBLIC;
GRANT EXECUTE ON FUNCTION clear_tenant_context() TO PUBLIC;

\echo 'Tenant functions created successfully'

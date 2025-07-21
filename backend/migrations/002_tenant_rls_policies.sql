-- Migration 004: Tenant Row-Level Security Policies
BEGIN;

-- Enable RLS on tenant-sensitive tables
ALTER TABLE store.inventory ENABLE ROW LEVEL SECURITY;
ALTER TABLE store.received ENABLE ROW LEVEL SECURITY;
ALTER TABLE store.customer_tenant_assignments ENABLE ROW LEVEL SECURITY;

-- Create function to get current tenant context
CREATE OR REPLACE FUNCTION get_current_tenant_id() 
RETURNS INTEGER AS $$
BEGIN
    BEGIN
        RETURN current_setting('app.current_tenant_id')::INTEGER;
    EXCEPTION
        WHEN OTHERS THEN
            RETURN (
                SELECT tenant_id 
                FROM store.users 
                WHERE user_id = current_setting('app.current_user_id')::INTEGER
            );
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to check if user is admin
CREATE OR REPLACE FUNCTION is_admin_user()
RETURNS BOOLEAN AS $$
BEGIN
    BEGIN
        RETURN current_setting('app.user_role') = 'admin';
    EXCEPTION
        WHEN OTHERS THEN
            RETURN FALSE;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- RLS policies
CREATE POLICY inventory_tenant_isolation ON store.inventory
    FOR ALL TO authenticated_users
    USING (
        tenant_id = get_current_tenant_id() 
        OR is_admin_user()
        OR tenant_id IS NULL
    );

CREATE POLICY received_tenant_isolation ON store.received
    FOR ALL TO authenticated_users
    USING (
        tenant_id = get_current_tenant_id()
        OR is_admin_user()
        OR tenant_id IS NULL
    );

COMMIT;

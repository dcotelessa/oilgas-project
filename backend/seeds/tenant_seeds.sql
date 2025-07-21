-- Tenant seed data
SET search_path TO store, public;

-- Insert default tenant
INSERT INTO store.tenants (tenant_name, tenant_slug, description, contact_email, phone) 
VALUES (
    'Petros',
    'petros',
    'Main division for Petros operations',
    'operations@petros.com',
    '432-555-0100'
) ON CONFLICT (tenant_slug) DO NOTHING;

-- Assign existing data to default tenant
DO $$
DECLARE
    petros_tenant_id INTEGER;
BEGIN
    SELECT tenant_id INTO petros_tenant_id 
    FROM store.tenants 
    WHERE tenant_slug = 'petros';

    -- Update existing data
    UPDATE store.users SET tenant_id = petros_tenant_id WHERE tenant_id IS NULL;
    UPDATE store.inventory SET tenant_id = petros_tenant_id WHERE tenant_id IS NULL;
    UPDATE store.received SET tenant_id = petros_tenant_id WHERE tenant_id IS NULL;

    -- Assign customers to default tenant
    INSERT INTO store.customer_tenant_assignments (customer_id, tenant_id, relationship_type)
    SELECT customer_id, petros_tenant_id, 'primary'
    FROM store.customers
    ON CONFLICT (customer_id, tenant_id) DO NOTHING;
END $$;

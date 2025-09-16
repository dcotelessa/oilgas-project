-- 005_add_audit_enhancements.up.sql
-- Enhanced audit logging with event types for work order tracking
ALTER TABLE audit.events 
ADD CONSTRAINT chk_event_type CHECK (event_type IN (
    -- User events
    'user.login', 'user.logout', 'user.created', 'user.password_changed',
    'user.deactivated', 'user.permission_granted',
    
    -- Customer events  
    'customer.created', 'customer.updated', 'customer.contact_added',
    'customer.contact_removed', 'customer.contact_permissions_updated',
    
    -- Work order events (for invoice audit trail)
    'workorder.created', 'workorder.status_changed', 'workorder.assigned',
    'workorder.started', 'workorder.completed', 'workorder.cancelled',
    'workorder.item_added', 'workorder.item_completed', 'workorder.item_removed',
    'workorder.approved', 'workorder.rejected', 'workorder.invoice_generated',
    
    -- Inventory events (for tracking where items go)
    'inventory.received', 'inventory.moved', 'inventory.inspected',
    'inventory.assigned_to_workorder', 'inventory.returned_from_workorder',
    
    -- System events
    'system.migration_completed', 'system.backup_created'
));

-- Enhanced audit logging function
CREATE OR REPLACE FUNCTION audit.log_event(
    p_event_type VARCHAR(100),
    p_tenant_id VARCHAR(100),
    p_entity_type VARCHAR(100) DEFAULT NULL,
    p_entity_id VARCHAR(100) DEFAULT NULL,
    p_user_id INTEGER DEFAULT NULL,
    p_event_data JSONB DEFAULT '{}',
    p_old_values JSONB DEFAULT NULL,
    p_new_values JSONB DEFAULT NULL
) RETURNS VARCHAR(255) AS $$
DECLARE
    event_id VARCHAR(255);
BEGIN
    event_id := gen_random_uuid()::text;
    
    INSERT INTO audit.events (
        id, event_type, tenant_id, entity_type, entity_id,
        user_id, event_data, old_values, new_values, created_at
    ) VALUES (
        event_id, p_event_type, p_tenant_id, p_entity_type, p_entity_id,
        p_user_id, p_event_data, p_old_values, p_new_values, NOW()
    );
    
    RETURN event_id;
END;
$$ LANGUAGE plpgsql;

-- Partitioning for audit events (performance for large datasets)
CREATE INDEX idx_audit_events_created_at_monthly ON audit.events(created_at, event_type);
CREATE INDEX idx_audit_events_workorder_trail ON audit.events(entity_type, entity_id, created_at) 
WHERE entity_type IN ('workorder', 'workorder_item');

-- Inventory tracking view (for "where did item go" queries)
CREATE VIEW audit.inventory_trail AS
SELECT 
    e.created_at,
    e.event_type,
    e.entity_id as inventory_item_id,
    e.event_data->>'workorder_id' as workorder_id,
    e.event_data->>'location' as location,
    e.event_data->>'status' as status,
    u.full_name as changed_by,
    e.tenant_id
FROM audit.events e
LEFT JOIN auth.users u ON e.user_id = u.id
WHERE e.entity_type = 'inventory'
ORDER BY e.created_at DESC;

-- Work order audit trail view
CREATE VIEW audit.workorder_trail AS
SELECT 
    wo.work_order_number,
    wo.customer_id,
    c.name as customer_name,
    e.event_type,
    e.created_at,
    e.event_data,
    e.old_values,
    e.new_values,
    u.full_name as changed_by
FROM audit.events e
JOIN store.workorders wo ON e.entity_id = wo.id::text
JOIN store.customers c ON wo.customer_id = c.id
LEFT JOIN auth.users u ON e.user_id = u.id
WHERE e.entity_type = 'workorder'
ORDER BY wo.id, e.created_at;

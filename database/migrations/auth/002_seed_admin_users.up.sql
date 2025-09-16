-- 002_seed_admin_users.up.sql
-- Seed initial admin users and permissions for multi-tenant system (Long Beach + Bakersfield)

-- Insert tenants (both Long Beach and Bakersfield)
INSERT INTO tenants (id, name, location, database_name, is_active) VALUES
('longbeach', 'Long Beach Operations', 'Long Beach, CA', 'location_longbeach', true),
('bakersfield', 'Bakersfield Operations', 'Bakersfield, CA', 'location_bakersfield', true),
('colorado', 'Colorado Operations', 'Colorado Springs, CO', 'location_colorado', false)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    location = EXCLUDED.location,
    database_name = EXCLUDED.database_name,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Insert comprehensive permissions system
INSERT INTO permissions (name, description, resource, action) VALUES
-- Customer permissions
('customer.read', 'View customer information', 'customer', 'read'),
('customer.write', 'Create and update customers', 'customer', 'write'),
('customer.delete', 'Delete customers', 'customer', 'delete'),
('customer.contacts.manage', 'Manage customer contact users', 'customer', 'contacts'),

-- Inventory permissions
('inventory.read', 'View inventory information', 'inventory', 'read'),
('inventory.write', 'Create and update inventory', 'inventory', 'write'),
('inventory.delete', 'Delete inventory items', 'inventory', 'delete'),
('inventory.transport', 'Manage transport logistics', 'inventory', 'transport'),

-- Work order permissions
('workorder.read', 'View work orders', 'workorder', 'read'),
('workorder.write', 'Create and update work orders', 'workorder', 'write'),
('workorder.approve', 'Approve work orders', 'workorder', 'approve'),
('workorder.assign', 'Assign work orders to users', 'workorder', 'assign'),

-- Invoice permissions
('invoice.read', 'View invoices', 'invoice', 'read'),
('invoice.write', 'Create and update invoices', 'invoice', 'write'),
('invoice.approve', 'Approve invoices for payment', 'invoice', 'approve'),

-- Analytics and reporting
('analytics.read', 'View analytics and reports', 'analytics', 'read'),
('analytics.export', 'Export data and reports', 'analytics', 'export'),

-- Admin permissions
('admin.users', 'Manage users and permissions', 'admin', 'users'),
('admin.tenants', 'Manage tenant configurations', 'admin', 'tenants'),
('admin.system', 'System administration access', 'admin', 'system'),

-- Cross-tenant permissions
('cross_tenant.view', 'View data across multiple tenants', 'cross_tenant', 'view'),
('cross_tenant.manage', 'Manage cross-tenant operations', 'cross_tenant', 'manage')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    resource = EXCLUDED.resource,
    action = EXCLUDED.action;

-- Assign permissions to roles based on business requirements
-- SYSTEM_ADMIN: All permissions
INSERT INTO role_permissions (role, permission_id) 
SELECT 'SYSTEM_ADMIN', id FROM permissions
ON CONFLICT (role, permission_id) DO NOTHING;

-- ENTERPRISE_ADMIN: All tenant permissions + cross-tenant
INSERT INTO role_permissions (role, permission_id) 
SELECT 'ENTERPRISE_ADMIN', id FROM permissions WHERE name NOT LIKE 'admin.system'
ON CONFLICT (role, permission_id) DO NOTHING;

-- ADMIN: All tenant-level permissions
INSERT INTO role_permissions (role, permission_id) 
SELECT 'ADMIN', id FROM permissions WHERE resource IN ('customer', 'inventory', 'workorder', 'invoice', 'analytics') OR name LIKE 'admin.users' OR name LIKE 'admin.tenants'
ON CONFLICT (role, permission_id) DO NOTHING;

-- MANAGER: Business operations without user management
INSERT INTO role_permissions (role, permission_id) 
SELECT 'MANAGER', id FROM permissions WHERE resource IN ('customer', 'inventory', 'workorder', 'invoice', 'analytics') AND action IN ('read', 'write', 'approve', 'export')
ON CONFLICT (role, permission_id) DO NOTHING;

-- OPERATOR: Daily operations (read/write but no delete/approve)
INSERT INTO role_permissions (role, permission_id) 
SELECT 'OPERATOR', id FROM permissions WHERE resource IN ('customer', 'inventory', 'workorder') AND action IN ('read', 'write')
ON CONFLICT (role, permission_id) DO NOTHING;

-- CUSTOMER_CONTACT: Limited to own work orders and customer data
INSERT INTO role_permissions (role, permission_id) 
SELECT 'CUSTOMER_CONTACT', id FROM permissions WHERE 
    (resource = 'workorder' AND action IN ('read', 'write')) OR
    (resource = 'customer' AND action = 'read') OR
    (resource = 'invoice' AND action = 'read')
ON CONFLICT (role, permission_id) DO NOTHING;

-- Create default system admin user with access to both tenants
INSERT INTO users (
    username, email, full_name, password_hash, role, 
    is_enterprise_user, primary_tenant_id, tenant_access, access_level, is_active
) VALUES (
    'sysadmin', 
    'admin@oilgas.com', 
    'System Administrator', 
    crypt('admin123', gen_salt('bf')), 
    'SYSTEM_ADMIN',
    true,
    'longbeach',
    '[
        {
            "tenant_id": "longbeach", 
            "role": "SYSTEM_ADMIN", 
            "permissions": ["VIEW_INVENTORY", "CREATE_WORK_ORDER", "APPROVE_WORK_ORDER", "MANAGE_TRANSPORT", "EXPORT_DATA", "USER_MANAGEMENT", "CROSS_TENANT_VIEW"], 
            "yard_access": [], 
            "can_read": true, 
            "can_write": true, 
            "can_delete": true, 
            "can_approve": true
        },
        {
            "tenant_id": "bakersfield", 
            "role": "SYSTEM_ADMIN", 
            "permissions": ["VIEW_INVENTORY", "CREATE_WORK_ORDER", "APPROVE_WORK_ORDER", "MANAGE_TRANSPORT", "EXPORT_DATA", "USER_MANAGEMENT", "CROSS_TENANT_VIEW"], 
            "yard_access": [], 
            "can_read": true, 
            "can_write": true, 
            "can_delete": true, 
            "can_approve": true
        }
    ]',
    10,
    true
) ON CONFLICT (email) DO UPDATE SET
    username = EXCLUDED.username,
    full_name = EXCLUDED.full_name,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    is_enterprise_user = EXCLUDED.is_enterprise_user,
    primary_tenant_id = EXCLUDED.primary_tenant_id,
    tenant_access = EXCLUDED.tenant_access,
    access_level = EXCLUDED.access_level,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Create Long Beach manager
INSERT INTO users (
    username, email, full_name, password_hash, role,
    is_enterprise_user, primary_tenant_id, tenant_access, access_level, is_active
) VALUES (
    'lb_manager',
    'manager@longbeach.oilgas.com',
    'Long Beach Manager',
    crypt('manager123', gen_salt('bf')),
    'MANAGER',
    false,
    'longbeach',
    '[{"tenant_id": "longbeach", "role": "MANAGER", "permissions": ["VIEW_INVENTORY", "CREATE_WORK_ORDER", "APPROVE_WORK_ORDER", "MANAGE_TRANSPORT", "EXPORT_DATA"], "yard_access": [], "can_read": true, "can_write": true, "can_delete": false, "can_approve": true}]',
    5,
    true
) ON CONFLICT (email) DO UPDATE SET
    username = EXCLUDED.username,
    full_name = EXCLUDED.full_name,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    is_enterprise_user = EXCLUDED.is_enterprise_user,
    primary_tenant_id = EXCLUDED.primary_tenant_id,
    tenant_access = EXCLUDED.tenant_access,
    access_level = EXCLUDED.access_level,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Create Bakersfield manager
INSERT INTO users (
    username, email, full_name, password_hash, role,
    is_enterprise_user, primary_tenant_id, tenant_access, access_level, is_active
) VALUES (
    'bk_manager',
    'manager@bakersfield.oilgas.com',
    'Bakersfield Manager',
    crypt('manager123', gen_salt('bf')),
    'MANAGER',
    false,
    'bakersfield',
    '[{"tenant_id": "bakersfield", "role": "MANAGER", "permissions": ["VIEW_INVENTORY", "CREATE_WORK_ORDER", "APPROVE_WORK_ORDER", "MANAGE_TRANSPORT", "EXPORT_DATA"], "yard_access": [], "can_read": true, "can_write": true, "can_delete": false, "can_approve": true}]',
    5,
    true
) ON CONFLICT (email) DO UPDATE SET
    username = EXCLUDED.username,
    full_name = EXCLUDED.full_name,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    is_enterprise_user = EXCLUDED.is_enterprise_user,
    primary_tenant_id = EXCLUDED.primary_tenant_id,
    tenant_access = EXCLUDED.tenant_access,
    access_level = EXCLUDED.access_level,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Grant system admin access to all tenants
INSERT INTO user_tenant_access (user_id, tenant_id, role, can_access)
SELECT u.id, t.id, 'SYSTEM_ADMIN', true 
FROM users u, tenants t 
WHERE u.username = 'sysadmin' AND t.is_active = true
ON CONFLICT (user_id, tenant_id) DO UPDATE SET
    role = EXCLUDED.role,
    can_access = EXCLUDED.can_access,
    granted_at = NOW();

-- Grant Long Beach manager access to Long Beach only
INSERT INTO user_tenant_access (user_id, tenant_id, role, can_access)
SELECT u.id, 'longbeach', 'MANAGER', true
FROM users u 
WHERE u.username = 'lb_manager'
ON CONFLICT (user_id, tenant_id) DO UPDATE SET
    role = EXCLUDED.role,
    can_access = EXCLUDED.can_access,
    granted_at = NOW();

-- Grant Bakersfield manager access to Bakersfield only
INSERT INTO user_tenant_access (user_id, tenant_id, role, can_access)
SELECT u.id, 'bakersfield', 'MANAGER', true
FROM users u 
WHERE u.username = 'bk_manager'
ON CONFLICT (user_id, tenant_id) DO UPDATE SET
    role = EXCLUDED.role,
    can_access = EXCLUDED.can_access,
    granted_at = NOW();

-- Log admin user creation events
INSERT INTO auth_events (user_id, event_type, tenant_id, details, ip_address)
SELECT u.id, 'USER_CREATED', 'longbeach', 
       jsonb_build_object('created_by', 'migration', 'role', 'SYSTEM_ADMIN', 'tenants', '["longbeach", "bakersfield"]'), 
       '127.0.0.1'::inet
FROM users u WHERE u.username = 'sysadmin';

INSERT INTO auth_events (user_id, event_type, tenant_id, details, ip_address)
SELECT u.id, 'USER_CREATED', 'longbeach', 
       jsonb_build_object('created_by', 'migration', 'role', 'MANAGER'), 
       '127.0.0.1'::inet
FROM users u WHERE u.username = 'lb_manager';

INSERT INTO auth_events (user_id, event_type, tenant_id, details, ip_address)
SELECT u.id, 'USER_CREATED', 'bakersfield', 
       jsonb_build_object('created_by', 'migration', 'role', 'MANAGER'), 
       '127.0.0.1'::inet
FROM users u WHERE u.username = 'bk_manager';
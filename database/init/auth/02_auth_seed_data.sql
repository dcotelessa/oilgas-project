# database/init/auth/02_auth_seed_data.sql
-- Seed data for Central Auth Database

-- Insert tenants
INSERT INTO tenants (id, name, location, database_name, is_active) VALUES
('longbeach', 'Long Beach Operations', 'Long Beach, CA', 'location_longbeach', true),
('bakersfield', 'Bakersfield Operations', 'Bakersfield, CA', 'location_bakersfield', false),
('colorado', 'Colorado Operations', 'Colorado Springs, CO', 'location_colorado', false);

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
('cross_tenant.manage', 'Manage cross-tenant operations', 'cross_tenant', 'manage');

-- Assign permissions to roles based on business requirements
-- SYSTEM_ADMIN: All permissions
INSERT INTO role_permissions (role, permission_id) 
SELECT 'SYSTEM_ADMIN', id FROM permissions;

-- ENTERPRISE_ADMIN: All tenant permissions + cross-tenant
INSERT INTO role_permissions (role, permission_id) 
SELECT 'ENTERPRISE_ADMIN', id FROM permissions WHERE name NOT LIKE 'admin.system';

-- ADMIN: All tenant-level permissions
INSERT INTO role_permissions (role, permission_id) 
SELECT 'ADMIN', id FROM permissions WHERE resource IN ('customer', 'inventory', 'workorder', 'invoice', 'analytics') OR name LIKE '%.users' OR name LIKE '%.tenants';

-- MANAGER: Business operations without user management
INSERT INTO role_permissions (role, permission_id) 
SELECT 'MANAGER', id FROM permissions WHERE resource IN ('customer', 'inventory', 'workorder', 'invoice', 'analytics') AND action IN ('read', 'write', 'approve', 'export');

-- OPERATOR: Daily operations (read/write but no delete/approve)
INSERT INTO role_permissions (role, permission_id) 
SELECT 'OPERATOR', id FROM permissions WHERE resource IN ('customer', 'inventory', 'workorder') AND action IN ('read', 'write');

-- CUSTOMER_CONTACT: Limited to own work orders and customer data
INSERT INTO role_permissions (role, permission_id) 
SELECT 'CUSTOMER_CONTACT', id FROM permissions WHERE 
    (resource = 'workorder' AND action IN ('read', 'write')) OR
    (resource = 'customer' AND action = 'read') OR
    (resource = 'invoice' AND action = 'read');

-- Create default system admin user
INSERT INTO users (
    username, email, full_name, password_hash, role, 
    is_enterprise_user, primary_tenant_id, tenant_access, access_level
) VALUES (
    'sysadmin', 
    'admin@oilgas.com', 
    'System Administrator', 
    crypt('admin123', gen_salt('bf')), 
    'SYSTEM_ADMIN',
    true,
    'longbeach',
    '[{"tenant_id": "longbeach", "role": "SYSTEM_ADMIN", "can_read": true, "can_write": true, "can_delete": true, "can_approve": true}]',
    10
);

-- Create Long Beach manager
INSERT INTO users (
    username, email, full_name, password_hash, role,
    is_enterprise_user, primary_tenant_id, tenant_access, access_level
) VALUES (
    'lb_manager',
    'manager@longbeach.oilgas.com',
    'Long Beach Manager',
    crypt('manager123', gen_salt('bf')),
    'MANAGER',
    false,
    'longbeach',
    '[{"tenant_id": "longbeach", "role": "MANAGER", "can_read": true, "can_write": true, "can_delete": false, "can_approve": true}]',
    5
);

-- Grant system admin access to all tenants
INSERT INTO user_tenant_access (user_id, tenant_id, role, can_access)
SELECT 1, id, 'SYSTEM_ADMIN', true FROM tenants;

-- Grant manager access to Long Beach only
INSERT INTO user_tenant_access (user_id, tenant_id, role, can_access)
VALUES (2, 'longbeach', 'MANAGER', true);


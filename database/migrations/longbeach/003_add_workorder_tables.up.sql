-- 003_add_workorder_tables.up.sql 
-- Create work orders table
CREATE TABLE store.workorders (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    customer_id INTEGER NOT NULL REFERENCES store.customers(id),
    work_order_number VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    service_type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) DEFAULT 'MEDIUM',
    description TEXT NOT NULL,
    instructions TEXT,
    
    -- Pricing and billing
    estimated_hours DECIMAL(8,2),
    actual_hours DECIMAL(8,2),
    hourly_rate DECIMAL(10,2),
    materials_cost DECIMAL(10,2),
    total_amount DECIMAL(10,2),
    
    -- Assignment and tracking
    assigned_to_user_id INTEGER REFERENCES auth.users(id),
    created_by_user_id INTEGER NOT NULL REFERENCES auth.users(id),
    
    -- Dates and timeline
    scheduled_date TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    due_date TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT fk_workorders_tenant FOREIGN KEY (tenant_id) REFERENCES store.tenants(tenant_id),
    CONSTRAINT uq_workorder_number UNIQUE(tenant_id, work_order_number),
    CONSTRAINT chk_status CHECK (status IN (
        'DRAFT', 'PENDING', 'APPROVED', 'IN_PROGRESS', 'COMPLETED', 
        'INVOICED', 'PAID', 'CANCELLED', 'ON_HOLD'
    )),
    CONSTRAINT chk_priority CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT'))
);

-- Work order items table
CREATE TABLE store.workorder_items (
    id SERIAL PRIMARY KEY,
    workorder_id INTEGER NOT NULL REFERENCES store.workorders(id) ON DELETE CASCADE,
    inventory_item_id INTEGER, -- Will reference inventory when created
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2),
    total_price DECIMAL(10,2),
    service_notes TEXT,
    is_completed BOOLEAN DEFAULT false,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT chk_quantity_positive CHECK (quantity > 0)
);

-- Work order history for audit trail
CREATE TABLE store.workorder_history (
    id SERIAL PRIMARY KEY,
    workorder_id INTEGER NOT NULL REFERENCES store.workorders(id) ON DELETE CASCADE,
    changed_by_user_id INTEGER NOT NULL REFERENCES auth.users(id),
    action VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Work order approvals table
CREATE TABLE store.workorder_approvals (
    id SERIAL PRIMARY KEY,
    workorder_id INTEGER NOT NULL REFERENCES store.workorders(id) ON DELETE CASCADE,
    approver_user_id INTEGER NOT NULL REFERENCES auth.users(id),
    approval_level INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    comments TEXT,
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_approval_status CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'))
);

-- Indexes for performance
CREATE INDEX idx_workorders_customer_status ON store.workorders(customer_id, status);
CREATE INDEX idx_workorders_tenant_status ON store.workorders(tenant_id, status) WHERE is_active = true;
CREATE INDEX idx_workorders_assigned_user ON store.workorders(assigned_to_user_id) WHERE assigned_to_user_id IS NOT NULL;
CREATE INDEX idx_workorders_created_by ON store.workorders(created_by_user_id);
CREATE INDEX idx_workorders_scheduled ON store.workorders(scheduled_date) WHERE scheduled_date IS NOT NULL;
CREATE INDEX idx_workorder_items_workorder ON store.workorder_items(workorder_id);
CREATE INDEX idx_workorder_history_workorder ON store.workorder_history(workorder_id, created_at DESC);
CREATE INDEX idx_workorder_approvals_pending ON store.workorder_approvals(approver_user_id, status) WHERE status = 'PENDING';

-- Work order number generation function
CREATE OR REPLACE FUNCTION generate_work_order_number(p_tenant_id VARCHAR(100)) RETURNS VARCHAR(100) AS $$
DECLARE
    next_number INTEGER;
    tenant_prefix VARCHAR(10);
BEGIN
    -- Get tenant prefix (first 3 chars of tenant_id, uppercase)
    tenant_prefix := UPPER(LEFT(p_tenant_id, 3));
    
    -- Get next number for this tenant
    SELECT COALESCE(MAX(
        CASE 
            WHEN work_order_number ~ ('^' || tenant_prefix || '-[0-9]+$')
            THEN CAST(SUBSTRING(work_order_number FROM LENGTH(tenant_prefix) + 2) AS INTEGER)
            ELSE 0
        END
    ), 0) + 1
    INTO next_number
    FROM store.workorders
    WHERE tenant_id = p_tenant_id;
    
    RETURN tenant_prefix || '-' || LPAD(next_number::TEXT, 6, '0');
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-generate work order numbers
CREATE OR REPLACE FUNCTION set_work_order_number() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.work_order_number IS NULL OR NEW.work_order_number = '' THEN
        NEW.work_order_number := generate_work_order_number(NEW.tenant_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_work_order_number
    BEFORE INSERT ON store.workorders
    FOR EACH ROW EXECUTE FUNCTION set_work_order_number();

-- Trigger for work order history
CREATE OR REPLACE FUNCTION log_workorder_changes() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        -- Log status changes
        IF OLD.status != NEW.status THEN
            INSERT INTO store.workorder_history (
                workorder_id, changed_by_user_id, action, old_value, new_value
            ) VALUES (
                NEW.id,
                COALESCE(NEW.assigned_to_user_id, NEW.created_by_user_id),
                'status_changed',
                OLD.status,
                NEW.status
            );
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_log_workorder_changes
    AFTER UPDATE ON store.workorders
    FOR EACH ROW EXECUTE FUNCTION log_workorder_changes();

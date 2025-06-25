-- Migration: 002_enhanced_indexes_clean_names
-- Description: Enhanced Database Indexes with clean column names

-- Customer search patterns
CREATE INDEX IF NOT EXISTS idx_customers_name_search 
ON store.customers USING gin(to_tsvector('english', customer));

CREATE INDEX IF NOT EXISTS idx_customers_active 
ON store.customers(customer, deleted) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_customers_contact_info 
ON store.customers(phone, email) WHERE deleted = false;

-- Inventory search and filtering (most critical for performance)
CREATE INDEX IF NOT EXISTS idx_inventory_customer_active 
ON store.inventory(customer_id, deleted, date_in DESC) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_inventory_grade_size 
ON store.inventory(grade, size, weight) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_inventory_location_rack 
ON store.inventory(location, rack) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_inventory_date_range 
ON store.inventory(date_in, date_out) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_inventory_work_order 
ON store.inventory(work_order) WHERE work_order IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_inventory_customer_po 
ON store.inventory(customer_po) WHERE customer_po IS NOT NULL;

-- Composite index for common inventory searches
CREATE INDEX IF NOT EXISTS idx_inventory_search_combo 
ON store.inventory(customer_id, grade, size, deleted, date_in DESC);

-- Full-text search for inventory notes (removed CONCURRENTLY)
CREATE INDEX IF NOT EXISTS idx_inventory_notes_search 
ON store.inventory USING gin(to_tsvector('english', notes)) 
WHERE notes IS NOT NULL;

-- Received items tracking
CREATE INDEX IF NOT EXISTS idx_received_customer_date 
ON store.received(customer_id, date_received DESC) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_received_work_order 
ON store.received(work_order) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_received_status 
ON store.received(complete, deleted, date_received DESC);

CREATE INDEX IF NOT EXISTS idx_received_well_lease 
ON store.received(well, lease) WHERE deleted = false;

-- Fletcher operations (threading/inspection)
CREATE INDEX IF NOT EXISTS idx_fletcher_customer_status 
ON store.fletcher(customer_id, complete, deleted);

CREATE INDEX IF NOT EXISTS idx_fletcher_date_range 
ON store.fletcher(date_in, date_out) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_fletcher_work_tracking 
ON store.fletcher(r_number, customer_id) WHERE deleted = false;

-- Bakeout tracking
CREATE INDEX IF NOT EXISTS idx_bakeout_customer_date 
ON store.bakeout(customer_id, date_in DESC);

CREATE INDEX IF NOT EXISTS idx_bakeout_grade_size 
ON store.bakeout(grade, size, joints);

-- Inspected items
CREATE INDEX IF NOT EXISTS idx_inspected_work_status 
ON store.inspected(work_order, complete, deleted);

CREATE INDEX IF NOT EXISTS idx_inspected_quality_metrics 
ON store.inspected(accept, reject, joints) WHERE deleted = false;

-- SWGC (Size, Weight, Grade, Connection) lookups
CREATE INDEX IF NOT EXISTS idx_swgc_customer_specs 
ON store.swgc(customer_id, size, weight, connection);

-- Users and access control
CREATE INDEX IF NOT EXISTS idx_users_username 
ON store.users(username) WHERE username IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_access_level 
ON store.users(access, username);

-- Temporary tables (if used for reporting)
CREATE INDEX IF NOT EXISTS idx_temp_work_order 
ON store.temp(work_order, username);

CREATE INDEX IF NOT EXISTS idx_tempinv_customer_work 
ON store.tempinv(customer_id, work_order);

-- Performance monitoring indexes
CREATE INDEX IF NOT EXISTS idx_inventory_performance_check 
ON store.inventory(created_at DESC) WHERE deleted = false;

-- Common WHERE clause patterns
CREATE INDEX IF NOT EXISTS idx_inventory_ctd_wstring 
ON store.inventory(ctd, w_string) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_inventory_color_tracking 
ON store.inventory(color, customer_id) WHERE deleted = false AND color IS NOT NULL;

-- Grade validation support
CREATE INDEX IF NOT EXISTS idx_grade_lookup 
ON store.grade(grade);

-- Customer color coding system
CREATE INDEX IF NOT EXISTS idx_customers_color_system 
ON store.customers(color1, color2, color3, color4, color5) WHERE deleted = false;

-- Analysis and reporting indexes
CREATE INDEX IF NOT EXISTS idx_inventory_monthly_reports 
ON store.inventory(date_trunc('month', date_in), customer_id, grade) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_received_monthly_reports 
ON store.received(date_trunc('month', date_received), customer_id) WHERE deleted = false;

-- Ensure we have statistics for the query planner
ANALYZE store.customers;
ANALYZE store.inventory; 
ANALYZE store.received;
ANALYZE store.fletcher;
ANALYZE store.bakeout;

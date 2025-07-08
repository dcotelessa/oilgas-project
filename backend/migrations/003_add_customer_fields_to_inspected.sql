-- Migration: 003_add_customer_fields_to_inspected
-- Description: Add customer tracking fields to inspected table for business logic

-- Add customer fields to inspected table (these are expected by the application logic)
ALTER TABLE store.inspected 
ADD COLUMN IF NOT EXISTS customer_id INTEGER REFERENCES store.customers(customer_id),
ADD COLUMN IF NOT EXISTS customer VARCHAR(50),
ADD COLUMN IF NOT EXISTS grade VARCHAR(50),
ADD COLUMN IF NOT EXISTS size VARCHAR(50),
ADD COLUMN IF NOT EXISTS weight VARCHAR(50),
ADD COLUMN IF NOT EXISTS connection VARCHAR(50),
ADD COLUMN IF NOT EXISTS inspector VARCHAR(50),
ADD COLUMN IF NOT EXISTS inspection_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS passed_joints INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS failed_joints INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS notes TEXT;

-- Add indexes for the new columns (building on 002's index strategy)
CREATE INDEX IF NOT EXISTS idx_inspected_customer_id ON store.inspected(customer_id);
CREATE INDEX IF NOT EXISTS idx_inspected_customer_grade ON store.inspected(customer_id, grade) WHERE customer_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_inspected_inspection_date ON store.inspected(inspection_date DESC) WHERE inspection_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_inspected_quality_tracking ON store.inspected(passed_joints, failed_joints, joints);

-- Update existing records to have valid customer references (if any exist)
-- This ensures referential integrity for existing data
UPDATE store.inspected 
SET customer_id = (
    SELECT r.customer_id 
    FROM store.received r 
    WHERE r.work_order = store.inspected.work_order 
    LIMIT 1
),
customer = (
    SELECT r.customer 
    FROM store.received r 
    WHERE r.work_order = store.inspected.work_order 
    LIMIT 1
),
grade = (
    SELECT r.grade 
    FROM store.received r 
    WHERE r.work_order = store.inspected.work_order 
    LIMIT 1
),
size = (
    SELECT r.size 
    FROM store.received r 
    WHERE r.work_order = store.inspected.work_order 
    LIMIT 1
)
WHERE customer_id IS NULL AND work_order IS NOT NULL;

-- Add constraints for data integrity
ALTER TABLE store.inspected 
ADD CONSTRAINT chk_passed_joints_valid 
CHECK (passed_joints >= 0),
ADD CONSTRAINT chk_failed_joints_valid 
CHECK (failed_joints >= 0),
ADD CONSTRAINT chk_joints_logic 
CHECK (passed_joints + failed_joints <= joints OR joints IS NULL);

-- Comments for documentation
COMMENT ON COLUMN store.inspected.customer_id IS 'Links inspection to customer for business tracking';
COMMENT ON COLUMN store.inspected.passed_joints IS 'Number of joints that passed inspection';
COMMENT ON COLUMN store.inspected.failed_joints IS 'Number of joints that failed inspection';
COMMENT ON COLUMN store.inspected.inspector IS 'Name of person who performed inspection';

-- Update statistics for query planner
ANALYZE store.inspected;

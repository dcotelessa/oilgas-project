-- 003_add_workorder_tables.down.sql
-- Drop work order related objects
DROP VIEW IF EXISTS audit.workorder_trail CASCADE;
DROP TRIGGER IF EXISTS trigger_log_workorder_changes ON store.workorders;
DROP TRIGGER IF EXISTS trigger_set_work_order_number ON store.workorders;
DROP FUNCTION IF EXISTS log_workorder_changes();
DROP FUNCTION IF EXISTS set_work_order_number();
DROP FUNCTION IF EXISTS generate_work_order_number(VARCHAR);
DROP TABLE IF EXISTS store.workorder_approvals CASCADE;
DROP TABLE IF EXISTS store.workorder_history CASCADE;
DROP TABLE IF EXISTS store.workorder_items CASCADE;
DROP TABLE IF EXISTS store.workorders CASCADE;

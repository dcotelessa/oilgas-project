-- 002_create_optimized_customers.down.sql
-- Remove customer optimization
DROP FUNCTION IF EXISTS migrate_customers_from_standardized();
ALTER TABLE auth.users DROP CONSTRAINT IF EXISTS fk_users_customer;
DROP TABLE IF EXISTS store.customers CASCADE;

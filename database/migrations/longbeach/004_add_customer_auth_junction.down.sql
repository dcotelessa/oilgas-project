-- 004_add_customer_auth_junction.down.sql
-- Drop customer-auth integration
DROP VIEW IF EXISTS store.customer_contact_details CASCADE;
DROP FUNCTION IF EXISTS create_customer_contact(INTEGER, VARCHAR, VARCHAR, VARCHAR, VARCHAR, VARCHAR, INTEGER);
DROP TABLE IF EXISTS auth.password_reset_tokens CASCADE;
DROP TABLE IF EXISTS store.customer_auth_contacts CASCADE;

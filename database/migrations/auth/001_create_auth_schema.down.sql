-- 001_create_auth_schema.down.sql
-- Drop auth schema and all related objects
DROP TABLE IF EXISTS auth.user_permissions CASCADE;
DROP TABLE IF EXISTS auth.sessions CASCADE;
DROP TABLE IF EXISTS auth.users CASCADE;
DROP TABLE IF EXISTS audit.events CASCADE;
DROP SCHEMA IF EXISTS auth CASCADE;
DROP SCHEMA IF EXISTS audit CASCADE;

-- Initial schema migration for Oil & Gas Inventory System
-- Generated from Phase 1 MDB migration

-- Create schema
CREATE SCHEMA IF NOT EXISTS store;
SET search_path TO store, public;

-- Create migrations table
CREATE SCHEMA IF NOT EXISTS migrations;
CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Note: Actual table schemas will be added based on MDB analysis
-- Run Phase 2 setup to complete the migration process

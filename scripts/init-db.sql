-- Initialize databases for Oil & Gas Inventory System

-- Create test database
CREATE DATABASE oil_gas_test;

-- Create schemas in main database
\c oil_gas_inventory;
CREATE SCHEMA IF NOT EXISTS store;
CREATE SCHEMA IF NOT EXISTS auth;

-- Create schemas in test database
\c oil_gas_test;
CREATE SCHEMA IF NOT EXISTS store;
CREATE SCHEMA IF NOT EXISTS auth;

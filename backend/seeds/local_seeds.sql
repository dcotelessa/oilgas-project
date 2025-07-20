-- Local development seed data
-- Oil & Gas Inventory System
-- Fixed: Proper foreign key order

SET search_path TO store, public;

-- Clear existing data (development only)
-- Note: Tables may not exist yet, so we use CASCADE and IF EXISTS
DROP TABLE IF EXISTS store.received CASCADE;
DROP TABLE IF EXISTS store.inventory CASCADE;
DROP TABLE IF EXISTS store.customers CASCADE;
DROP TABLE IF EXISTS store.sizes CASCADE; 
DROP TABLE IF EXISTS store.grade CASCADE;

-- Oil & gas industry standard grades (referenced by inventory)
CREATE TABLE IF NOT EXISTS store.grade (
    grade VARCHAR(10) PRIMARY KEY,
    description TEXT
);

INSERT INTO store.grade (grade, description) VALUES 
('J55', 'Standard grade steel casing - most common'),
('JZ55', 'Enhanced J55 grade with improved properties'),
('L80', 'Higher strength grade for moderate environments'),
('N80', 'Medium strength grade for standard applications'),
('P105', 'High performance grade for demanding conditions'),
('P110', 'Premium performance grade for extreme environments'),
('Q125', 'Ultra-high strength grade for specialized applications'),
('C75', 'Carbon steel grade for basic applications'),
('C95', 'Higher carbon steel grade'),
('T95', 'Tough grade for harsh environments');

-- Common pipe sizes in oil & gas industry
CREATE TABLE IF NOT EXISTS store.sizes (
    size_id SERIAL PRIMARY KEY,
    size VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

INSERT INTO store.sizes (size, description) VALUES 
('4 1/2"', '4.5 inch diameter - small casing'),
('5"', '5 inch diameter - intermediate casing'),
('5 1/2"', '5.5 inch diameter - common production casing'),
('7"', '7 inch diameter - intermediate casing'),
('8 5/8"', '8.625 inch diameter - surface casing'),
('9 5/8"', '9.625 inch diameter - surface casing'),
('10 3/4"', '10.75 inch diameter - surface casing'),
('13 3/8"', '13.375 inch diameter - surface casing'),
('16"', '16 inch diameter - conductor casing'),
('18 5/8"', '18.625 inch diameter - conductor casing'),
('20"', '20 inch diameter - large conductor casing'),
('24"', '24 inch diameter - extra large conductor'),
('30"', '30 inch diameter - structural casing');

-- Customers table (must come before inventory/received)
CREATE TABLE IF NOT EXISTS store.customers (
    customer_id SERIAL PRIMARY KEY,
    customer VARCHAR(255) NOT NULL,
    billing_address TEXT,
    billing_city VARCHAR(100),
    billing_state VARCHAR(50),
    billing_zipcode VARCHAR(20),
    contact VARCHAR(255),
    phone VARCHAR(50),
    fax VARCHAR(50),
    email VARCHAR(255),
    deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sample customers (oil & gas companies) - INSERT FIRST
INSERT INTO store.customers (customer, billing_address, billing_city, billing_state, billing_zipcode, contact, phone, fax, email) VALUES 
('Permian Basin Energy', '1234 Oil Field Rd', 'Midland', 'TX', '79701', 'John Smith', '432-555-0101', '432-555-0102', 'operations@permianbasin.com'),
('Eagle Ford Solutions', '5678 Shale Ave', 'San Antonio', 'TX', '78201', 'Sarah Johnson', '210-555-0201', '210-555-0202', 'drilling@eagleford.com'),
('Bakken Industries', '9012 Prairie Blvd', 'Williston', 'ND', '58801', 'Mike Wilson', '701-555-0301', '701-555-0302', 'procurement@bakken.com'),
('Gulf Coast Drilling', '3456 Offshore Dr', 'Houston', 'TX', '77001', 'Lisa Brown', '713-555-0401', '713-555-0402', 'logistics@gulfcoast.com'),
('Marcellus Gas Co', '7890 Mountain View', 'Pittsburgh', 'PA', '15201', 'Robert Davis', '412-555-0501', '412-555-0502', 'operations@marcellus.com');

-- Inventory table (references customers)
CREATE TABLE IF NOT EXISTS store.inventory (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100),
    work_order VARCHAR(100),
    r_number VARCHAR(100),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(255),
    joints INTEGER,
    rack VARCHAR(50),
    size VARCHAR(50),
    weight DECIMAL(10,2),
    grade VARCHAR(10) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd VARCHAR(100),
    w_string VARCHAR(100),
    swgcc VARCHAR(100),
    color VARCHAR(50),
    customer_po VARCHAR(100),
    fletcher VARCHAR(100),
    date_in DATE,
    date_out DATE,
    well_in VARCHAR(255),
    lease_in VARCHAR(255),
    well_out VARCHAR(255),
    lease_out VARCHAR(255),
    trucking VARCHAR(100),
    trailer VARCHAR(100),
    location VARCHAR(100),
    notes TEXT,
    pcode VARCHAR(50),
    cn VARCHAR(50),
    ordered_by VARCHAR(100),
    deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sample inventory data (NOW with valid customer_id references)
INSERT INTO store.inventory (work_order, customer_id, customer, joints, size, weight, grade, connection, date_in, well_in, lease_in, location, notes) VALUES 
('WO-2024-001', 1, 'Permian Basin Energy', 100, '5 1/2"', 2500.50, 'L80', 'BTC', '2024-01-15', 'Well-PB-001', 'Lease-PB-A', 'Yard-A', 'Standard production casing'),
('WO-2024-002', 2, 'Eagle Ford Solutions', 150, '7"', 4200.75, 'P110', 'VAM TOP', '2024-01-16', 'Well-EF-002', 'Lease-EF-B', 'Yard-B', 'High pressure application'),
('WO-2024-003', 3, 'Bakken Industries', 75, '9 5/8"', 6800.25, 'N80', 'LTC', '2024-01-17', 'Well-BK-003', 'Lease-BK-C', 'Yard-C', 'Surface casing'),
('WO-2024-004', 4, 'Gulf Coast Drilling', 200, '5 1/2"', 5000.00, 'J55', 'STC', '2024-01-18', 'Well-GC-004', 'Lease-GC-D', 'Yard-A', 'Offshore application');

-- Received table (also references customers and sizes)
CREATE TABLE IF NOT EXISTS store.received (
    id SERIAL PRIMARY KEY,
    work_order VARCHAR(100),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(255),
    joints INTEGER,
    rack VARCHAR(50),
    size_id INTEGER REFERENCES store.sizes(size_id),
    size VARCHAR(50),
    weight DECIMAL(10,2),
    grade VARCHAR(10) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd VARCHAR(100),
    w_string VARCHAR(100),
    well VARCHAR(255),
    lease VARCHAR(255),
    ordered_by VARCHAR(100),
    notes TEXT,
    customer_po VARCHAR(100),
    date_received DATE,
    background TEXT,
    norm VARCHAR(100),
    services TEXT,
    bill_to_id INTEGER,
    entered_by VARCHAR(100),
    when_entered TIMESTAMP,
    trucking VARCHAR(100),
    trailer VARCHAR(100),
    in_production BOOLEAN DEFAULT FALSE,
    inspected_date DATE,
    threading_date DATE,
    straighten_required BOOLEAN DEFAULT FALSE,
    excess_material BOOLEAN DEFAULT FALSE,
    complete BOOLEAN DEFAULT FALSE,
    inspected_by VARCHAR(100),
    updated_by VARCHAR(100),
    when_updated TIMESTAMP,
    deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sample received data (with valid customer_id and size_id references)
INSERT INTO store.received (work_order, customer_id, customer, joints, size_id, size, weight, grade, connection, date_received, well, lease, ordered_by, notes, in_production, complete) VALUES 
('WO-2024-005', 1, 'Permian Basin Energy', 80, 4, '7"', 3200.00, 'L80', 'BTC', '2024-01-20', 'Well-PB-005', 'Lease-PB-E', 'John Smith', 'Expedited order', false, false),
('WO-2024-006', 5, 'Marcellus Gas Co', 120, 3, '5 1/2"', 3000.00, 'P110', 'VAM TOP', '2024-01-21', 'Well-MG-006', 'Lease-MG-F', 'Robert Davis', 'High pressure specs', false, false),
('WO-2024-007', 2, 'Eagle Ford Solutions', 90, 5, '8 5/8"', 7200.00, 'N80', 'LTC', '2024-01-22', 'Well-EF-007', 'Lease-EF-G', 'Sarah Johnson', 'Surface casing rush', true, false);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
CREATE INDEX IF NOT EXISTS idx_inventory_date_in ON store.inventory(date_in);
CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order);
CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received);

-- Note: Phase 1 normalized CSV data will be imported later
-- Run 'make import-clean-data' after Phase 1 completion to import real data

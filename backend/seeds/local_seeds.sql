-- Local development seed data
-- Oil & Gas Inventory System

SET search_path TO store, public;

-- Clear existing data (development only)
TRUNCATE TABLE store.received, store.inventory CASCADE;
TRUNCATE TABLE store.customers CASCADE;
TRUNCATE TABLE store.sizes CASCADE; 
DELETE FROM store.grade;

-- Oil & gas industry standard grades
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

-- Sample customers (oil & gas companies)
INSERT INTO store.customers (customer, billing_address, billing_city, billing_state, billing_zipcode, contact, phone, fax, email) VALUES 
('Permian Basin Energy', '1234 Oil Field Rd', 'Midland', 'TX', '79701', 'John Smith', '432-555-0101', '432-555-0102', 'operations@permianbasin.com'),
('Eagle Ford Solutions', '5678 Shale Ave', 'San Antonio', 'TX', '78201', 'Sarah Johnson', '210-555-0201', '210-555-0202', 'drilling@eagleford.com'),
('Bakken Industries', '9012 Prairie Blvd', 'Williston', 'ND', '58801', 'Mike Wilson', '701-555-0301', '701-555-0302', 'procurement@bakken.com'),
('Gulf Coast Drilling', '3456 Offshore Dr', 'Houston', 'TX', '77001', 'Lisa Brown', '713-555-0401', '713-555-0402', 'logistics@gulfcoast.com'),
('Marcellus Gas Co', '7890 Mountain View', 'Pittsburgh', 'PA', '15201', 'Robert Davis', '412-555-0501', '412-555-0502', 'operations@marcellus.com');

-- Sample inventory data (will be replaced by Phase 1 import)
INSERT INTO store.inventory (work_order, customer_id, customer, joints, size, weight, grade, connection, date_in, well_in, lease_in, location, notes) VALUES 
('WO-2024-001', 1, 'Permian Basin Energy', 100, '5 1/2"', 2500.50, 'L80', 'BTC', '2024-01-15', 'Well-PB-001', 'Lease-PB-A', 'Yard-A', 'Standard production casing'),
('WO-2024-002', 2, 'Eagle Ford Solutions', 150, '7"', 4200.75, 'P110', 'VAM TOP', '2024-01-16', 'Well-EF-002', 'Lease-EF-B', 'Yard-B', 'High pressure application'),
('WO-2024-003', 3, 'Bakken Industries', 75, '9 5/8"', 6800.25, 'N80', 'LTC', '2024-01-17', 'Well-BK-003', 'Lease-BK-C', 'Yard-C', 'Surface casing'),
('WO-2024-004', 4, 'Gulf Coast Drilling', 200, '5 1/2"', 5000.00, 'J55', 'STC', '2024-01-18', 'Well-GC-004', 'Lease-GC-D', 'Yard-A', 'Offshore application');

-- Sample received data  
INSERT INTO store.received (work_order, customer_id, customer, joints, size, weight, grade, connection, date_received, well, lease, ordered_by, notes, in_production, complete) VALUES 
('WO-2024-005', 1, 'Permian Basin Energy', 80, '7"', 3200.00, 'L80', 'BTC', '2024-01-20', 'Well-PB-005', 'Lease-PB-E', 'John Smith', 'Expedited order', false, false),
('WO-2024-006', 5, 'Marcellus Gas Co', 120, '5 1/2"', 3000.00, 'P110', 'VAM TOP', '2024-01-21', 'Well-MG-006', 'Lease-MG-F', 'Robert Davis', 'High pressure specs', false, false),
('WO-2024-007', 2, 'Eagle Ford Solutions', 90, '8 5/8"', 7200.00, 'N80', 'LTC', '2024-01-22', 'Well-EF-007', 'Lease-EF-G', 'Sarah Johnson', 'Surface casing rush', true, false);

-- Note: Additional data will be imported from Phase 1 normalized CSV files
-- Run 'make import-clean-data' after Phase 1 completion to import real data

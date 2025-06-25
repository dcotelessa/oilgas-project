-- Local environment seed data with clean column names
-- FAKE/MOCK DATA ONLY - Safe for development and version control

SET search_path TO store, public;

-- Clear existing data first (in correct order to handle foreign keys)
TRUNCATE TABLE store.inventory, store.received, store.fletcher, store.bakeout CASCADE;
TRUNCATE TABLE store.customers CASCADE;
TRUNCATE TABLE store.users CASCADE;
TRUNCATE TABLE store.grade CASCADE;
TRUNCATE TABLE store.r_number, store.wk_number CASCADE;

-- Insert grade data (reference data - industry standard)
INSERT INTO store.grade (grade) VALUES 
('J55'), ('JZ55'), ('K55'), ('L80'), ('N80'), ('P105'), ('P110'), ('Q125'), ('T95'), ('C90'), ('C95'), ('S135')
ON CONFLICT (grade) DO NOTHING;

-- Insert FAKE customers (completely fictional companies)
INSERT INTO store.customers (
    customer, billing_address, billing_city, billing_state, billing_zipcode, 
    contact, phone, email, deleted
) VALUES 
('Demo Oil Services LLC', '1234 Fake Street', 'Houston', 'TX', '77001', 'John Demo', '555-0001', 'demo@example.com', false),
('Test Drilling Co', '5678 Sample Ave', 'Dallas', 'TX', '75001', 'Jane Test', '555-0002', 'test@example.com', false),
('Mock Energy Corp', '9101 Development Blvd', 'Austin', 'TX', '78701', 'Bob Mock', '555-0003', 'mock@example.com', false),
('Example Pipe & Supply', '1122 Placeholder Dr', 'San Antonio', 'TX', '78201', 'Alice Example', '555-0004', 'example@test.com', false),
('Sample Oilfield Services', '3344 Template Rd', 'Fort Worth', 'TX', '76101', 'Charlie Sample', '555-0005', 'sample@demo.com', false),
('Fictional Tubular Inc', '5566 Mock Lane', 'Midland', 'TX', '79701', 'David Fictional', '555-0006', 'david@fictional.com', false),
('Demo Casing Solutions', '7788 Test Boulevard', 'Odessa', 'TX', '79761', 'Emma Demo', '555-0007', 'emma@demosolutions.com', false),
('Sample Energy Partners', '9900 Example Circle', 'Lubbock', 'TX', '79401', 'Frank Sample', '555-0008', 'frank@sampleenergy.com', false);

-- Insert FAKE inventory records (realistic but fake data)
INSERT INTO store.inventory (
    username, work_order, r_number, customer_id, customer, joints, rack, 
    size, weight, grade, connection, ctd, w_string, color, customer_po, 
    location, deleted, date_in
) VALUES 
('demouser1', 'WO-DEMO-001', 1001, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 
 'Demo Oil Services LLC', 100, 'A1-DEMO', '5 1/2"', '20#', 'J55', 'LTC', 
 false, false, 'RED', 'PO-FAKE-001', 'Demo Yard North', false, '2024-01-15 08:00:00'),

('testuser2', 'WO-TEST-002', 1002, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Test Drilling Co'), 
 'Test Drilling Co', 150, 'B2-TEST', '7"', '26#', 'L80', 'BTC', 
 true, false, 'BLUE', 'PO-TEST-002', 'Test Yard South', false, '2024-01-20 09:30:00'),

('mockuser3', 'WO-MOCK-003', 1003, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Mock Energy Corp'), 
 'Mock Energy Corp', 75, 'C3-MOCK', '4 1/2"', '16.6#', 'N80', 'EUE', 
 false, true, 'GREEN', 'PO-MOCK-003', 'Mock Warehouse', false, '2024-02-01 10:15:00'),

('exampleuser', 'WO-EXAM-004', 1004, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Example Pipe & Supply'), 
 'Example Pipe & Supply', 200, 'D4-EXAM', '9 5/8"', '40#', 'P110', 'LTC', 
 true, false, 'YELLOW', 'PO-EXAM-004', 'Example Yard East', false, '2024-02-05 11:00:00'),

('sampleuser', 'WO-SAMP-005', 1005, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Sample Oilfield Services'), 
 'Sample Oilfield Services', 125, 'E5-SAMP', '6 5/8"', '24#', 'P105', 'BTC', 
 false, true, 'WHITE', 'PO-SAMP-005', 'Sample Yard West', false, '2024-02-10 07:30:00'),

('demouser1', 'WO-DEMO-006', 1006, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Fictional Tubular Inc'), 
 'Fictional Tubular Inc', 300, 'F6-FICT', '5"', '18#', 'JZ55', 'LTC', 
 false, false, 'ORANGE', 'PO-FICT-006', 'Fictional Yard', false, '2024-02-15 09:00:00'),

('testuser2', 'WO-DEMO-007', 1007, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Demo Casing Solutions'), 
 'Demo Casing Solutions', 80, 'G7-DEMO', '7 5/8"', '29.7#', 'L80', 'BTC', 
 true, true, 'PURPLE', 'PO-DEMO-007', 'Demo Casing Yard', false, '2024-02-20 14:30:00'),

('sampleuser', 'WO-SAMP-008', 1008, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Sample Energy Partners'), 
 'Sample Energy Partners', 175, 'H8-SAMP', '8 5/8"', '32#', 'N80', 'EUE', 
 false, false, 'BLACK', 'PO-SAMP-008', 'Sample Energy Yard', false, '2024-02-25 16:00:00');

-- Insert FAKE received records (incoming pipe tracking)
INSERT INTO store.received (
    work_order, customer_id, customer, joints, size, weight, grade, 
    connection, well, lease, date_received, entered_by, ordered_by, 
    complete, deleted
) VALUES 
('WO-DEMO-001', 
 (SELECT customer_id FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 
 'Demo Oil Services LLC', 100, '5 1/2"', '20#', 'J55', 'LTC', 
 'Demo Well #1', 'Fake Lease Alpha', '2024-01-15 08:00:00', 
 'demo_operator', 'Demo Foreman', false, false),

('WO-TEST-002', 
 (SELECT customer_id FROM store.customers WHERE customer = 'Test Drilling Co'), 
 'Test Drilling Co', 150, '7"', '26#', 'L80', 'BTC', 
 'Test Well #2', 'Sample Lease Beta', '2024-01-20 09:30:00', 
 'test_operator', 'Test Supervisor', true, false),

('WO-MOCK-003', 
 (SELECT customer_id FROM store.customers WHERE customer = 'Mock Energy Corp'), 
 'Mock Energy Corp', 75, '4 1/2"', '16.6#', 'N80', 'EUE', 
 'Mock Well #3', 'Example Lease Gamma', '2024-02-01 10:15:00', 
 'mock_operator', 'Mock Manager', false, false),

('WO-EXAM-004', 
 (SELECT customer_id FROM store.customers WHERE customer = 'Example Pipe & Supply'), 
 'Example Pipe & Supply', 200, '9 5/8"', '40#', 'P110', 'LTC', 
 'Example Well #4', 'Demo Lease Delta', '2024-02-05 11:00:00', 
 'example_operator', 'Example Supervisor', true, false),

('WO-SAMP-005', 
 (SELECT customer_id FROM store.customers WHERE customer = 'Sample Oilfield Services'), 
 'Sample Oilfield Services', 125, '6 5/8"', '24#', 'P105', 'BTC', 
 'Sample Well #5', 'Test Lease Epsilon', '2024-02-10 07:30:00', 
 'sample_operator', 'Sample Manager', false, false);

-- Insert FAKE fletcher records (threading/inspection)
INSERT INTO store.fletcher (
    username, fletcher, r_number, customer_id, customer, joints, size, 
    weight, grade, connection, well_in, lease_in, date_in, ordered_by, 
    complete, deleted
) VALUES 
('fletcher_demo', 'Demo Fletcher Station', 1001, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 
 'Demo Oil Services LLC', 100, '5 1/2"', '20#', 'J55', 'LTC', 
 'Demo Well #1', 'Fake Lease Alpha', '2024-01-16 07:00:00', 
 'Demo Foreman', true, false),

('fletcher_test', 'Test Threading Bay', 1002, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Test Drilling Co'), 
 'Test Drilling Co', 150, '7"', '26#', 'L80', 'BTC', 
 'Test Well #2', 'Sample Lease Beta', '2024-01-21 08:30:00', 
 'Test Supervisor', false, false),

('fletcher_mock', 'Mock Threading Station', 1003, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Mock Energy Corp'), 
 'Mock Energy Corp', 75, '4 1/2"', '16.6#', 'N80', 'EUE', 
 'Mock Well #3', 'Example Lease Gamma', '2024-02-02 09:00:00', 
 'Mock Manager', true, false);

-- Insert FAKE bakeout records
INSERT INTO store.bakeout (
    fletcher, joints, color, size, weight, grade, connection, ctd, 
    customer_id, accept, reject, pin, cplg, pc, date_in
) VALUES 
('Demo Fletcher Station', 100, 'RED', '5 1/2"', '20#', 'J55', 'LTC', false, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 
 95, 5, 2, 1, 2, '2024-01-17 10:00:00'),

('Test Threading Bay', 150, 'BLUE', '7"', '26#', 'L80', 'BTC', true, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Test Drilling Co'), 
 145, 5, 3, 1, 1, '2024-01-22 11:00:00'),

('Mock Threading Station', 75, 'GREEN', '4 1/2"', '16.6#', 'N80', 'EUE', false, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Mock Energy Corp'), 
 72, 3, 1, 1, 1, '2024-02-03 12:00:00');

-- Insert FAKE inspected records
INSERT INTO store.inspected (
    username, work_order, color, joints, accept, reject, pin, cplg, pc, 
    complete, rack, deleted
) VALUES 
('inspector_demo', 'WO-DEMO-001', 'RED', 100, 95, 5, 2, 1, 2, true, 'A1-DEMO', false),
('inspector_test', 'WO-TEST-002', 'BLUE', 150, 145, 5, 3, 1, 1, true, 'B2-TEST', false),
('inspector_mock', 'WO-MOCK-003', 'GREEN', 75, 72, 3, 1, 1, 1, false, 'C3-MOCK', false),
('inspector_example', 'WO-EXAM-004', 'YELLOW', 200, 195, 5, 2, 2, 1, true, 'D4-EXAM', false);

-- Insert FAKE users (development accounts only)
INSERT INTO store.users (username, password, access, full_name, email) VALUES 
('admin', '$2a$10$demo.hash.for.development.only', 1, 'Demo Administrator', 'admin@demo.local'),
('testuser', '$2a$10$test.hash.for.development.only', 2, 'Test User', 'user@test.local'),
('operator', '$2a$10$operator.hash.for.development.only', 3, 'Demo Operator', 'operator@demo.local'),
('inspector', '$2a$10$inspector.hash.for.development.only', 3, 'Demo Inspector', 'inspector@demo.local'),
('fletcher', '$2a$10$fletcher.hash.for.development.only', 3, 'Demo Fletcher', 'fletcher@demo.local')
ON CONFLICT (username) DO NOTHING;

-- Add some fake work/run numbers for testing
INSERT INTO store.r_number (r_number) VALUES (1001), (1002), (1003), (1004), (1005), (1006), (1007), (1008), (1009), (1010)
ON CONFLICT (r_number) DO NOTHING;

INSERT INTO store.wk_number (wk_number) VALUES (2001), (2002), (2003), (2004), (2005), (2006), (2007), (2008), (2009), (2010)
ON CONFLICT (wk_number) DO NOTHING;

-- Insert FAKE SWGC configurations (Size, Weight, Grade, Connection)
INSERT INTO store.swgc (
    size_id, customer_id, size, weight, connection, pcode_receive, pcode_inventory
) VALUES 
(1, (SELECT customer_id FROM store.customers WHERE customer = 'Demo Oil Services LLC'), '5 1/2"', '20#', 'LTC', 'REC-001', 'INV-001'),
(2, (SELECT customer_id FROM store.customers WHERE customer = 'Test Drilling Co'), '7"', '26#', 'BTC', 'REC-002', 'INV-002'),
(3, (SELECT customer_id FROM store.customers WHERE customer = 'Mock Energy Corp'), '4 1/2"', '16.6#', 'EUE', 'REC-003', 'INV-003'),
(4, (SELECT customer_id FROM store.customers WHERE customer = 'Example Pipe & Supply'), '9 5/8"', '40#', 'LTC', 'REC-004', 'INV-004'),
(5, (SELECT customer_id FROM store.customers WHERE customer = 'Sample Oilfield Services'), '6 5/8"', '24#', 'BTC', 'REC-005', 'INV-005');

-- Insert some FAKE temp processing records
INSERT INTO store.temp (
    username, work_order, color, joints, accept, reject, pin, cplg, pc, rack
) VALUES 
('temp_user1', 'WO-TEMP-001', 'RED', 50, 48, 2, 1, 0, 1, 'TEMP-A1'),
('temp_user2', 'WO-TEMP-002', 'BLUE', 75, 72, 3, 2, 1, 0, 'TEMP-B2'),
('temp_user3', 'WO-TEMP-003', 'GREEN', 25, 24, 1, 0, 0, 1, 'TEMP-C3');

-- Insert FAKE temp inventory records
INSERT INTO store.tempinv (
    username, work_order, customer_id, customer, joints, rack, size, weight, 
    grade, connection, ctd, w_string, color, customer_po, fletcher, 
    date_in, location
) VALUES 
('temp_inv1', 'WO-TEMP-INV-001', 
 (SELECT customer_id FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 
 'Demo Oil Services LLC', 30, 'TEMP-INV-A1', '5 1/2"', '20#', 'J55', 'LTC', 
 false, false, 'RED', 'PO-TEMP-001', 'Temp Fletcher', '2024-03-01 08:00:00', 'Temp Storage'),

('temp_inv2', 'WO-TEMP-INV-002', 
 (SELECT customer_id FROM store.customers WHERE customer = 'Test Drilling Co'), 
 'Test Drilling Co', 40, 'TEMP-INV-B2', '7"', '26#', 'L80', 'BTC', 
 true, false, 'BLUE', 'PO-TEMP-002', 'Temp Threading', '2024-03-02 09:00:00', 'Temp Processing');

-- Insert test record
INSERT INTO store.test (test) VALUES ('Demo Test Record'), ('Sample Test Data'), ('Mock Test Entry');

-- Add some realistic inventory with different statuses and locations
INSERT INTO store.inventory (
    username, work_order, r_number, customer_id, customer, joints, rack, 
    size, weight, grade, connection, ctd, w_string, color, customer_po, 
    location, deleted, date_in, notes
) VALUES 
-- More realistic inventory spread
('operator1', 'WO-2024-001', 2001, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 
 'Demo Oil Services LLC', 500, 'YARD-A-001', '5 1/2"', '20#', 'J55', 'LTC', 
 false, false, 'RED', 'PO-2024-001', 'North Yard', false, '2024-01-05 06:00:00',
 'Standard tubing for horizontal drilling'),

('operator2', 'WO-2024-002', 2002, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Test Drilling Co'), 
 'Test Drilling Co', 750, 'YARD-B-001', '7"', '26#', 'L80', 'BTC', 
 true, false, 'BLUE', 'PO-2024-002', 'South Yard', false, '2024-01-10 07:00:00',
 'Casing for deep well project'),

('operator3', 'WO-2024-003', 2003, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Fictional Tubular Inc'), 
 'Fictional Tubular Inc', 250, 'YARD-C-001', '9 5/8"', '40#', 'P110', 'VAM', 
 true, true, 'SILVER', 'PO-2024-003', 'Premium Storage', false, '2024-01-15 08:00:00',
 'Premium connections for high pressure well'),

('operator1', 'WO-2024-004', 2004, 
 (SELECT customer_id FROM store.customers WHERE customer = 'Sample Energy Partners'), 
 'Sample Energy Partners', 1000, 'YARD-D-001', '13 3/8"', '68#', 'K55', 'BTC', 
 false, false, 'GREEN', 'PO-2024-004', 'Large Pipe Storage', false, '2024-01-20 09:00:00',
 'Surface casing for multiple wells');

-- Update statistics to help query planner
ANALYZE;

-- Note: All data above is FICTIONAL and safe for development/testing
-- Real production data should only be in production_seeds.sql (not in version control)

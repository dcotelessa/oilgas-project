-- Local environment seed data
-- FAKE/MOCK DATA ONLY - Safe for development and version control

SET search_path TO store, public;

-- Clear existing data first (in correct order to handle foreign keys)
TRUNCATE TABLE store.inventory, store.received, store.fletcher CASCADE;
TRUNCATE TABLE store.customers CASCADE;
TRUNCATE TABLE store.users CASCADE;
TRUNCATE TABLE store.grade CASCADE;
TRUNCATE TABLE store.rnumber, store.wknumber CASCADE;

-- Insert grade data (reference data - industry standard)
INSERT INTO store.grade (grade) VALUES 
('J55'), ('JZ55'), ('L80'), ('N80'), ('P105'), ('P110')
ON CONFLICT (grade) DO NOTHING;

-- Insert FAKE customers (completely fictional companies)
INSERT INTO store.customers (customer, billingaddress, billingcity, billingstate, billingzipcode, contact, phone, email, deleted) VALUES 
('Demo Oil Services LLC', '1234 Fake Street', 'Houston', 'TX', '77001', 'John Demo', '555-0001', 'demo@example.com', false),
('Test Drilling Co', '5678 Sample Ave', 'Dallas', 'TX', '75001', 'Jane Test', '555-0002', 'test@example.com', false),
('Mock Energy Corp', '9101 Development Blvd', 'Austin', 'TX', '78701', 'Bob Mock', '555-0003', 'mock@example.com', false),
('Example Pipe & Supply', '1122 Placeholder Dr', 'San Antonio', 'TX', '78201', 'Alice Example', '555-0004', 'example@test.com', false),
('Sample Oilfield Services', '3344 Template Rd', 'Fort Worth', 'TX', '76101', 'Charlie Sample', '555-0005', 'sample@demo.com', false);

-- Get the actual customer IDs that were inserted
-- Insert FAKE inventory records (realistic but fake data)
INSERT INTO store.inventory (username, wkorder, rnumber, custid, customer, joints, rack, size, weight, grade, connection, ctd, wstring, color, customerpo, location, deleted) VALUES 
('demouser1', 'WO-DEMO-001', 1001, (SELECT custid FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 'Demo Oil Services LLC', 100, 'A1-DEMO', '5.5"', '20#', 'J55', 'LTC', false, false, 'RED', 'PO-FAKE-001', 'Demo Yard North', false),
('testuser2', 'WO-TEST-002', 1002, (SELECT custid FROM store.customers WHERE customer = 'Test Drilling Co'), 'Test Drilling Co', 150, 'B2-TEST', '7"', '26#', 'L80', 'BTC', true, false, 'BLUE', 'PO-TEST-002', 'Test Yard South', false),
('mockuser3', 'WO-MOCK-003', 1003, (SELECT custid FROM store.customers WHERE customer = 'Mock Energy Corp'), 'Mock Energy Corp', 75, 'C3-MOCK', '4.5"', '16.6#', 'N80', 'EUE', false, true, 'GREEN', 'PO-MOCK-003', 'Mock Warehouse', false),
('exampleuser', 'WO-EXAM-004', 1004, (SELECT custid FROM store.customers WHERE customer = 'Example Pipe & Supply'), 'Example Pipe & Supply', 200, 'D4-EXAM', '9 5/8"', '40#', 'P110', 'LTC', true, false, 'YELLOW', 'PO-EXAM-004', 'Example Yard East', false),
('sampleuser', 'WO-SAMP-005', 1005, (SELECT custid FROM store.customers WHERE customer = 'Sample Oilfield Services'), 'Sample Oilfield Services', 125, 'E5-SAMP', '6 5/8"', '24#', 'P105', 'BTC', false, true, 'WHITE', 'PO-SAMP-005', 'Sample Yard West', false);

-- Insert FAKE received records (incoming pipe tracking)
INSERT INTO store.received (wkorder, custid, customer, joints, size, weight, grade, connection, well, lease, daterecvd, enteredby, orderedby, complete, deleted) VALUES 
('WO-DEMO-001', (SELECT custid FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 'Demo Oil Services LLC', 100, '5.5"', '20#', 'J55', 'LTC', 'Demo Well #1', 'Fake Lease Alpha', '2024-01-15 08:00:00', 'demo_operator', 'Demo Foreman', false, false),
('WO-TEST-002', (SELECT custid FROM store.customers WHERE customer = 'Test Drilling Co'), 'Test Drilling Co', 150, '7"', '26#', 'L80', 'BTC', 'Test Well #2', 'Sample Lease Beta', '2024-01-20 09:30:00', 'test_operator', 'Test Supervisor', true, false),
('WO-MOCK-003', (SELECT custid FROM store.customers WHERE customer = 'Mock Energy Corp'), 'Mock Energy Corp', 75, '4.5"', '16.6#', 'N80', 'EUE', 'Mock Well #3', 'Example Lease Gamma', '2024-02-01 10:15:00', 'mock_operator', 'Mock Manager', false, false);

-- Insert FAKE fletcher records (threading/inspection)
INSERT INTO store.fletcher (username, fletcher, rnumber, custid, customer, joints, size, weight, grade, connection, wellin, leasein, datein, orderedby, complete, deleted) VALUES 
('fletcher_demo', 'Demo Fletcher Station', 1001, (SELECT custid FROM store.customers WHERE customer = 'Demo Oil Services LLC'), 'Demo Oil Services LLC', 100, '5.5"', '20#', 'J55', 'LTC', 'Demo Well #1', 'Fake Lease Alpha', '2024-01-16 07:00:00', 'Demo Foreman', true, false),
('fletcher_test', 'Test Threading Bay', 1002, (SELECT custid FROM store.customers WHERE customer = 'Test Drilling Co'), 'Test Drilling Co', 150, '7"', '26#', 'L80', 'BTC', 'Test Well #2', 'Sample Lease Beta', '2024-01-21 08:30:00', 'Test Supervisor', false, false);

-- Insert FAKE users (development accounts only) - let serial handle IDs
INSERT INTO store.users (username, password, access, fullname, email) VALUES 
('admin', 'demo123', 1, 'Demo Administrator', 'admin@demo.local'),
('testuser', 'test123', 2, 'Test User', 'user@test.local'),
('operator', 'op123', 3, 'Demo Operator', 'operator@demo.local')
ON CONFLICT (username) DO NOTHING;

-- Add some fake work/run numbers for testing
INSERT INTO store.rnumber (rnumber) VALUES (1001), (1002), (1003), (1004), (1005)
ON CONFLICT (rnumber) DO NOTHING;

INSERT INTO store.wknumber (wknumber) VALUES (2001), (2002), (2003), (2004), (2005)
ON CONFLICT (wknumber) DO NOTHING;

-- Note: All data above is FICTIONAL and safe for development/testing
-- Real production data should only be in production_seeds.sql (not in version control)

-- Local environment seed data with clean column names
-- FAKE/MOCK DATA ONLY - Safe for development and version control

SET search_path TO store, public;

-- Clear existing data first (in correct order to handle foreign keys)
TRUNCATE TABLE store.inventory, store.received, store.fletcher, store.bakeout CASCADE;
TRUNCATE TABLE store.inspected, store.temp, store.tempinv CASCADE;
TRUNCATE TABLE store.swgc CASCADE;
TRUNCATE TABLE store.customers CASCADE;
TRUNCATE TABLE store.users CASCADE;
TRUNCATE TABLE store.grade CASCADE;
TRUNCATE TABLE store.r_number, store.wk_number CASCADE;

-- Insert grade data (reference data - industry standard)
INSERT INTO store.grade (grade, description) VALUES 
('J55', 'API Grade J55 - Low carbon steel'),
('JZ55', 'API Grade JZ55 - Low carbon steel with improved properties'),
('K55', 'API Grade K55 - Medium carbon steel'),
('L80', 'API Grade L80 - Heat treated steel'),
('N80', 'API Grade N80 - Heat treated steel with higher strength'),
('P105', 'API Grade P105 - High strength steel'),
('P110', 'API Grade P110 - Very high strength steel'),
('Q125', 'API Grade Q125 - Ultra high strength steel'),
('T95', 'API Grade T95 - High strength seamless steel'),
('C90', 'API Grade C90 - Heat treated steel'),
('C95', 'API Grade C95 - Heat treated steel with improved properties'),
('S135', 'API Grade S135 - Super high strength steel')
ON CONFLICT (grade) DO NOTHING;

-- Insert reference numbers
INSERT INTO store.r_number (r_number, description) VALUES
(1001, 'Standard receipt number'),
(1002, 'Emergency receipt number'),
(1003, 'Bulk receipt number')
ON CONFLICT (r_number) DO NOTHING;

-- Insert work order numbers
INSERT INTO store.wk_number (wk_number, description) VALUES
(2001, 'Standard work order'),
(2002, 'Priority work order'),
(2003, 'Emergency work order')
ON CONFLICT (wk_number) DO NOTHING;

-- Insert FAKE users (development only)
INSERT INTO store.users (username, password_hash, full_name, email, role) VALUES 
('demouser', '$2a$10$dummy.hash.for.development.purposes.only', 'Demo User', 'demo@localhost.dev', 'admin'),
('testuser', '$2a$10$dummy.hash.for.development.purposes.only', 'Test User', 'test@localhost.dev', 'user'),
('operator1', '$2a$10$dummy.hash.for.development.purposes.only', 'Operator One', 'op1@localhost.dev', 'operator')
ON CONFLICT (username) DO NOTHING;

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
('demouser', 'WO-DEMO-001', 1001, 1, 'Demo Oil Services LLC', 100, 'A1', '5 1/2"', '20.00', 'J55', 'BTC', false, false, 'Red', 'PO-DEMO-001', 'Yard A', false, NOW() - INTERVAL '30 days'),
('testuser', 'WO-TEST-001', 1002, 2, 'Test Drilling Co', 75, 'B2', '7"', '26.00', 'L80', 'LTC', true, false, 'Blue', 'PO-TEST-001', 'Yard B', false, NOW() - INTERVAL '25 days'),
('operator1', 'WO-MOCK-001', 1003, 3, 'Mock Energy Corp', 150, 'C3', '9 5/8"', '40.00', 'N80', 'BTC', false, true, 'Green', 'PO-MOCK-001', 'Yard C', false, NOW() - INTERVAL '20 days'),
('demouser', 'WO-EXAMPLE-001', 1001, 4, 'Example Pipe & Supply', 50, 'D4', '4 1/2"', '12.75', 'P110', 'PH6', true, true, 'Yellow', 'PO-EX-001', 'Yard D', false, NOW() - INTERVAL '15 days'),
('testuser', 'WO-SAMPLE-001', 1002, 5, 'Sample Oilfield Services', 200, 'E5', '3 1/2"', '9.50', 'C90', 'EUE', false, false, 'Orange', 'PO-SAMP-001', 'Yard E', false, NOW() - INTERVAL '10 days');

-- Insert FAKE received records (realistic but fake data)
INSERT INTO store.received (
    work_order, customer_id, customer, joints, rack, size, weight, grade, 
    connection, well, lease, ordered_by, customer_po, date_received, 
    entered_by, when_entered, deleted
) VALUES 
('WO-REC-001', 1, 'Demo Oil Services LLC', 80, 'R1', '5 1/2"', '20.00', 'J55', 'BTC', 'Demo Well #1', 'Demo Lease A', 'John Demo', 'PO-REC-001', NOW() - INTERVAL '5 days', 'demouser', NOW() - INTERVAL '5 days', false),
('WO-REC-002', 2, 'Test Drilling Co', 60, 'R2', '7"', '26.00', 'L80', 'LTC', 'Test Well #2', 'Test Lease B', 'Jane Test', 'PO-REC-002', NOW() - INTERVAL '3 days', 'testuser', NOW() - INTERVAL '3 days', false),
('WO-REC-003', 3, 'Mock Energy Corp', 120, 'R3', '9 5/8"', '40.00', 'N80', 'BTC', 'Mock Well #3', 'Mock Lease C', 'Bob Mock', 'PO-REC-003', NOW() - INTERVAL '1 day', 'operator1', NOW() - INTERVAL '1 day', false);

-- Insert some FAKE SWGC combinations
INSERT INTO store.swgc (customer_id, size, weight, grade, connection, pcode_receive, pcode_inventory) VALUES
(1, '5 1/2"', '20.00', 'J55', 'BTC', 'RCV001', 'INV001'),
(2, '7"', '26.00', 'L80', 'LTC', 'RCV002', 'INV002'),
(3, '9 5/8"', '40.00', 'N80', 'BTC', 'RCV003', 'INV003'),
(4, '4 1/2"', '12.75', 'P110', 'PH6', 'RCV004', 'INV004'),
(5, '3 1/2"', '9.50', 'C90', 'EUE', 'RCV005', 'INV005');

-- Data integrity check
DO $$
DECLARE
    customer_count INTEGER;
    inventory_count INTEGER;
    received_count INTEGER;
    grade_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO customer_count FROM store.customers WHERE deleted = false;
    SELECT COUNT(*) INTO inventory_count FROM store.inventory WHERE deleted = false;
    SELECT COUNT(*) INTO received_count FROM store.received WHERE deleted = false;
    SELECT COUNT(*) INTO grade_count FROM store.grade;
    
    RAISE NOTICE '📊 Seed Data Summary:';
    RAISE NOTICE '   Customers: %', customer_count;
    RAISE NOTICE '   Inventory Items: %', inventory_count;
    RAISE NOTICE '   Received Items: %', received_count;
    RAISE NOTICE '   Oil & Gas Grades: %', grade_count;
END $$;

-- Update statistics for better query performance
ANALYZE;


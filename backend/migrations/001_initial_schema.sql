-- Initial schema for Oil & Gas Inventory System
-- Based on Access database conversion with cleaned column names

-- Create store schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS store;
SET search_path TO store, public;

-- Table: customers (main customer information)
CREATE TABLE IF NOT EXISTS store.customers (
    customer_id SERIAL PRIMARY KEY,
    customer VARCHAR(255) NOT NULL,
    billing_address TEXT,
    billing_city VARCHAR(100),
    billing_state VARCHAR(10),
    billing_zipcode VARCHAR(20),
    contact VARCHAR(255),
    phone VARCHAR(50),
    fax VARCHAR(50),
    email VARCHAR(255),
    color1 VARCHAR(50),
    color2 VARCHAR(50),
    color3 VARCHAR(50),
    color4 VARCHAR(50),
    color5 VARCHAR(50),
    loss1 DECIMAL(10,2),
    loss2 DECIMAL(10,2),
    loss3 DECIMAL(10,2),
    loss4 DECIMAL(10,2),
    loss5 DECIMAL(10,2),
    wscolor1 VARCHAR(50),
    wscolor2 VARCHAR(50),
    wscolor3 VARCHAR(50),
    wscolor4 VARCHAR(50),
    wscolor5 VARCHAR(50),
    wsloss1 DECIMAL(10,2),
    wsloss2 DECIMAL(10,2),
    wsloss3 DECIMAL(10,2),
    wsloss4 DECIMAL(10,2),
    wsloss5 DECIMAL(10,2),
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: grade (reference table for oil & gas grades)
CREATE TABLE IF NOT EXISTS store.grade (
    grade VARCHAR(50) PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: users (system users)
CREATE TABLE IF NOT EXISTS store.users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    email VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user',
    active BOOLEAN NOT NULL DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Table: r_number (reference numbers)
CREATE TABLE IF NOT EXISTS store.r_number (
    r_number INTEGER PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: wk_number (work order numbers)
CREATE TABLE IF NOT EXISTS store.wk_number (
    wk_number INTEGER PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: inventory (main inventory items)
CREATE TABLE IF NOT EXISTS store.inventory (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100),
    work_order VARCHAR(100),
    r_number INTEGER,
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(255),
    joints INTEGER,
    rack VARCHAR(100),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd BOOLEAN NOT NULL DEFAULT false,
    w_string BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(100),
    color VARCHAR(50),
    customer_po VARCHAR(100),
    fletcher BOOLEAN NOT NULL DEFAULT false,
    date_in TIMESTAMP,
    date_out TIMESTAMP,
    well_in VARCHAR(255),
    lease_in VARCHAR(255),
    well_out VARCHAR(255),
    lease_out VARCHAR(255),
    trucking VARCHAR(100),
    trailer VARCHAR(100),
    location VARCHAR(255),
    notes TEXT,
    pcode VARCHAR(50),
    cn INTEGER,
    ordered_by VARCHAR(100),
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: received (items received for processing)
CREATE TABLE IF NOT EXISTS store.received (
    id SERIAL PRIMARY KEY,
    work_order VARCHAR(100),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(255),
    joints INTEGER,
    rack VARCHAR(100),
    size_id INTEGER,
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd BOOLEAN NOT NULL DEFAULT false,
    w_string BOOLEAN NOT NULL DEFAULT false,
    well VARCHAR(255),
    lease VARCHAR(255),
    ordered_by VARCHAR(100),
    notes TEXT,
    customer_po VARCHAR(100),
    date_received TIMESTAMP,
    background VARCHAR(100),
    norm VARCHAR(100),
    services VARCHAR(100),
    bill_to_id VARCHAR(100),
    entered_by VARCHAR(100),
    when_entered TIMESTAMP,
    trucking VARCHAR(100),
    trailer VARCHAR(100),
    in_production TIMESTAMP,
    inspected_date TIMESTAMP,
    threading_date TIMESTAMP,
    straighten_required BOOLEAN NOT NULL DEFAULT false,
    excess_material BOOLEAN NOT NULL DEFAULT false,
    complete BOOLEAN NOT NULL DEFAULT false,
    inspected_by VARCHAR(100),
    updated_by VARCHAR(100),
    when_updated TIMESTAMP,
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: fletcher (fletcher operations)
CREATE TABLE IF NOT EXISTS store.fletcher (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100),
    work_order VARCHAR(100),
    joints INTEGER,
    color VARCHAR(50),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(100),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    accept INTEGER,
    reject INTEGER,
    pin INTEGER,
    cplg INTEGER,
    pc INTEGER,
    trucking VARCHAR(100),
    trailer VARCHAR(100),
    date_in TIMESTAMP,
    cn INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: bakeout (bakeout operations)
CREATE TABLE IF NOT EXISTS store.bakeout (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100),
    work_order VARCHAR(100),
    joints INTEGER,
    color VARCHAR(50),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(100),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    accept INTEGER,
    reject INTEGER,
    pin INTEGER,
    cplg INTEGER,
    pc INTEGER,
    trucking VARCHAR(100),
    trailer VARCHAR(100),
    date_in TIMESTAMP,
    cn INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: inspected (inspection results)
CREATE TABLE IF NOT EXISTS store.inspected (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100),
    work_order VARCHAR(100),
    color VARCHAR(50),
    joints INTEGER,
    accept INTEGER,
    reject INTEGER,
    pin INTEGER,
    cplg INTEGER,
    pc INTEGER,
    complete BOOLEAN NOT NULL DEFAULT false,
    rack VARCHAR(100),
    rep_pin INTEGER,
    rep_cplg INTEGER,
    rep_pc INTEGER,
    deleted BOOLEAN NOT NULL DEFAULT false,
    cn INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: temp (temporary processing)
CREATE TABLE IF NOT EXISTS store.temp (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100),
    work_order VARCHAR(100),
    color VARCHAR(50),
    joints INTEGER,
    accept INTEGER,
    reject INTEGER,
    pin INTEGER,
    cplg INTEGER,
    pc INTEGER,
    rack VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: tempinv (temporary inventory)
CREATE TABLE IF NOT EXISTS store.tempinv (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100),
    work_order VARCHAR(100),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(255),
    joints INTEGER,
    rack VARCHAR(100),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    ctd BOOLEAN NOT NULL DEFAULT false,
    w_string BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(100),
    color VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: swgc (size, weight, grade, connection combinations)
CREATE TABLE IF NOT EXISTS store.swgc (
    id SERIAL PRIMARY KEY,
    size_id INTEGER,
    customer_id INTEGER REFERENCES store.customers(customer_id),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50) REFERENCES store.grade(grade),
    connection VARCHAR(100),
    pcode_receive VARCHAR(50),
    pcode_inventory VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_customer ON store.customers(customer);
CREATE INDEX IF NOT EXISTS idx_customers_deleted ON store.customers(deleted);

CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
CREATE INDEX IF NOT EXISTS idx_inventory_grade ON store.inventory(grade);
CREATE INDEX IF NOT EXISTS idx_inventory_deleted ON store.inventory(deleted);
CREATE INDEX IF NOT EXISTS idx_inventory_date_in ON store.inventory(date_in);

CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
CREATE INDEX IF NOT EXISTS idx_received_work_order ON store.received(work_order);
CREATE INDEX IF NOT EXISTS idx_received_grade ON store.received(grade);
CREATE INDEX IF NOT EXISTS idx_received_deleted ON store.received(deleted);
CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received);

-- Set default search path
ALTER DATABASE $DATABASE_NAME SET search_path TO store, public;

-- Grant permissions (adjust as needed)
-- GRANT ALL PRIVILEGES ON SCHEMA store TO your_app_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA store TO your_app_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA store TO your_app_user;

-- Migration: 001_initial_schema_clean
-- Description: Create initial schema with clean, consistent naming
-- This creates the schema with proper snake_case naming from the start

-- Create schema
CREATE SCHEMA IF NOT EXISTS store;
SET search_path TO store, public;

-- Table: customers
CREATE TABLE IF NOT EXISTS store.customers (
    customer_id SERIAL PRIMARY KEY,
    customer VARCHAR(50),
    billing_address VARCHAR(50),
    billing_city VARCHAR(50),
    billing_state VARCHAR(50),
    billing_zipcode VARCHAR(50),
    contact VARCHAR(50),
    phone VARCHAR(50),
    fax VARCHAR(50),
    email VARCHAR(50),
    color1 VARCHAR(50),
    color2 VARCHAR(50),
    color3 VARCHAR(50),
    color4 VARCHAR(50),
    color5 VARCHAR(50),
    loss1 VARCHAR(50),
    loss2 VARCHAR(50),
    loss3 VARCHAR(50),
    loss4 VARCHAR(50),
    loss5 VARCHAR(50),
    wscolor1 VARCHAR(50),
    wscolor2 VARCHAR(50),
    wscolor3 VARCHAR(50),
    wscolor4 VARCHAR(50),
    wscolor5 VARCHAR(50),
    wsloss1 VARCHAR(50),
    wsloss2 VARCHAR(50),
    wsloss3 VARCHAR(50),
    wsloss4 VARCHAR(50),
    wsloss5 VARCHAR(50),
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: inventory
CREATE TABLE IF NOT EXISTS store.inventory (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    work_order VARCHAR(50),
    r_number INTEGER,
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(50),
    joints INTEGER,
    rack VARCHAR(50),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    w_string BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(50),
    color VARCHAR(50),
    customer_po VARCHAR(50),
    fletcher VARCHAR(50),
    date_in TIMESTAMP,
    date_out TIMESTAMP,
    well_in VARCHAR(50),
    lease_in VARCHAR(50),
    well_out VARCHAR(50),
    lease_out VARCHAR(50),
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    location VARCHAR(50),
    notes TEXT,
    pcode VARCHAR(50),
    cn INTEGER,
    ordered_by VARCHAR(50),
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: received
CREATE TABLE IF NOT EXISTS store.received (
    id SERIAL PRIMARY KEY,
    work_order VARCHAR(50),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(50),
    joints INTEGER,
    rack VARCHAR(50),
    size_id INTEGER,
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    w_string BOOLEAN NOT NULL DEFAULT false,
    well VARCHAR(50),
    lease VARCHAR(50),
    ordered_by VARCHAR(50),
    notes TEXT,
    customer_po VARCHAR(50),
    date_received TIMESTAMP,
    background VARCHAR(50),
    norm VARCHAR(50),
    services VARCHAR(50),
    bill_to_id VARCHAR(50),
    entered_by VARCHAR(50),
    when_entered TIMESTAMP,
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    in_production TIMESTAMP,
    inspected_date TIMESTAMP,
    threading_date TIMESTAMP,
    straighten_required BOOLEAN NOT NULL DEFAULT false,
    excess_material BOOLEAN NOT NULL DEFAULT false,
    complete BOOLEAN NOT NULL DEFAULT false,
    inspected_by VARCHAR(50),
    updated_by VARCHAR(50),
    when_updated TIMESTAMP,
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: fletcher
CREATE TABLE IF NOT EXISTS store.fletcher (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    fletcher VARCHAR(50),
    r_number INTEGER,
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(50),
    joints INTEGER,
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    w_string BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(50),
    color VARCHAR(50),
    customer_po VARCHAR(50),
    date_in TIMESTAMP,
    date_out TIMESTAMP,
    well_in VARCHAR(50),
    lease_in VARCHAR(50),
    well_out VARCHAR(50),
    lease_out VARCHAR(50),
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    location VARCHAR(50),
    notes TEXT,
    pcode VARCHAR(50),
    cn INTEGER,
    ordered_by VARCHAR(50),
    deleted BOOLEAN NOT NULL DEFAULT false,
    complete BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: bakeout
CREATE TABLE IF NOT EXISTS store.bakeout (
    id SERIAL PRIMARY KEY,
    fletcher VARCHAR(50),
    joints INTEGER,
    color VARCHAR(50),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(50),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    accept INTEGER,
    reject INTEGER,
    pin INTEGER,
    cplg INTEGER,
    pc INTEGER,
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    date_in TIMESTAMP,
    cn INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: inspected
CREATE TABLE IF NOT EXISTS store.inspected (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    work_order VARCHAR(50),
    color VARCHAR(50),
    joints INTEGER,
    accept INTEGER,
    reject INTEGER,
    pin INTEGER,
    cplg INTEGER,
    pc INTEGER,
    complete BOOLEAN NOT NULL DEFAULT false,
    rack VARCHAR(50),
    rep_pin INTEGER,
    rep_cplg INTEGER,
    rep_pc INTEGER,
    deleted BOOLEAN NOT NULL DEFAULT false,
    cn INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: grade
CREATE TABLE IF NOT EXISTS store.grade (
    grade VARCHAR(50) PRIMARY KEY
);

-- Table: swgc
CREATE TABLE IF NOT EXISTS store.swgc (
    size_id INTEGER,
    customer_id INTEGER REFERENCES store.customers(customer_id),
    size VARCHAR(50),
    weight VARCHAR(50),
    connection VARCHAR(50),
    pcode_receive VARCHAR(50),
    pcode_inventory VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: temp
CREATE TABLE IF NOT EXISTS store.temp (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    work_order VARCHAR(50),
    color VARCHAR(50),
    joints INTEGER,
    accept INTEGER,
    reject INTEGER,
    pin INTEGER,
    cplg INTEGER,
    pc INTEGER,
    rack VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: tempinv
CREATE TABLE IF NOT EXISTS store.tempinv (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    work_order VARCHAR(50),
    customer_id INTEGER REFERENCES store.customers(customer_id),
    customer VARCHAR(50),
    joints INTEGER,
    rack VARCHAR(50),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    w_string BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(50),
    color VARCHAR(50),
    customer_po VARCHAR(50),
    fletcher VARCHAR(50),
    date_in TIMESTAMP,
    date_out TIMESTAMP,
    well_in VARCHAR(50),
    lease_in VARCHAR(50),
    well_out VARCHAR(50),
    lease_out VARCHAR(50),
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    location VARCHAR(50),
    notes TEXT,
    pcode VARCHAR(50),
    cn INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: test
CREATE TABLE IF NOT EXISTS store.test (
    id SERIAL PRIMARY KEY,
    test VARCHAR(50)
);

-- Table: users
CREATE TABLE IF NOT EXISTS store.users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(12) UNIQUE,
    password VARCHAR(255), -- Increased for proper password hashing
    access INTEGER,
    full_name VARCHAR(50),
    email VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: r_number
CREATE TABLE IF NOT EXISTS store.r_number (
    r_number INTEGER PRIMARY KEY
);

-- Table: wk_number  
CREATE TABLE IF NOT EXISTS store.wk_number (
    wk_number INTEGER PRIMARY KEY
);

-- Insert standard oil & gas grades
INSERT INTO store.grade (grade) VALUES 
('J55'), ('JZ55'), ('K55'), ('L80'), ('N80'), 
('P105'), ('P110'), ('Q125'), ('T95'), ('C90'), ('C95'), ('S135')
ON CONFLICT (grade) DO NOTHING;

-- Basic indexes for common queries
CREATE INDEX IF NOT EXISTS idx_customers_customer ON store.customers(customer);
CREATE INDEX IF NOT EXISTS idx_customers_deleted ON store.customers(deleted) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_inventory_customer_id ON store.inventory(customer_id);
CREATE INDEX IF NOT EXISTS idx_inventory_grade ON store.inventory(grade);
CREATE INDEX IF NOT EXISTS idx_inventory_work_order ON store.inventory(work_order);
CREATE INDEX IF NOT EXISTS idx_inventory_deleted ON store.inventory(deleted) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_received_customer_id ON store.received(customer_id);
CREATE INDEX IF NOT EXISTS idx_received_date_received ON store.received(date_received);
CREATE INDEX IF NOT EXISTS idx_received_deleted ON store.received(deleted) WHERE deleted = false;

CREATE INDEX IF NOT EXISTS idx_fletcher_customer_id ON store.fletcher(customer_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON store.users(username);

-- Update statistics
ANALYZE;

-- PostgreSQL Schema for Oil & Gas Inventory System
-- Generated from ColdFusion/MDB migration

-- Create schema
CREATE SCHEMA IF NOT EXISTS store;

-- Set search path
SET search_path TO store, public;

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
    custid SERIAL PRIMARY KEY,
    accept INTEGER,
    reject INTEGER,
    pin INTEGER,
    cplg INTEGER,
    pc INTEGER,
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    datein TIMESTAMP,
    cn INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: customers
CREATE TABLE IF NOT EXISTS store.customers (
    custid SERIAL PRIMARY KEY,
    customer VARCHAR(50),
    billingaddress VARCHAR(50),
    billingcity VARCHAR(50),
    billingstate VARCHAR(50),
    billingzipcode VARCHAR(50),
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

-- Table: fletcher
CREATE TABLE IF NOT EXISTS store.fletcher (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    fletcher VARCHAR(50),
    rnumber INTEGER,
    custid SERIAL PRIMARY KEY,
    customer VARCHAR(50),
    joints INTEGER,
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    wstring BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(50),
    color VARCHAR(50),
    customerpo VARCHAR(50),
    datein TIMESTAMP,
    dateout TIMESTAMP,
    wellin VARCHAR(50),
    leasein VARCHAR(50),
    wellout VARCHAR(50),
    leaseout VARCHAR(50),
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    location VARCHAR(50),
    notes TEXT,
    pcode VARCHAR(50),
    cn INTEGER,
    orderedby VARCHAR(50),
    deleted BOOLEAN NOT NULL DEFAULT false,
    complete BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: grade
CREATE TABLE IF NOT EXISTS store.grade (
    grade VARCHAR(50)
);

-- Table: inspected
CREATE TABLE IF NOT EXISTS store.inspected (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    wkorder VARCHAR(50),
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

-- Table: inventory
CREATE TABLE IF NOT EXISTS store.inventory (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    wkorder VARCHAR(50),
    rnumber INTEGER,
    custid SERIAL PRIMARY KEY,
    customer VARCHAR(50),
    joints INTEGER,
    rack VARCHAR(50),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    wstring BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(50),
    color VARCHAR(50),
    customerpo VARCHAR(50),
    fletcher VARCHAR(50),
    datein TIMESTAMP,
    dateout TIMESTAMP,
    wellin VARCHAR(50),
    leasein VARCHAR(50),
    wellout VARCHAR(50),
    leaseout VARCHAR(50),
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    location VARCHAR(50),
    notes TEXT,
    pcode VARCHAR(50),
    cn INTEGER,
    orderedby VARCHAR(50),
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: received
CREATE TABLE IF NOT EXISTS store.received (
    id SERIAL PRIMARY KEY,
    wkorder VARCHAR(50),
    custid SERIAL PRIMARY KEY,
    customer VARCHAR(50),
    joints INTEGER,
    rack VARCHAR(50),
    sizeid INTEGER,
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    wstring BOOLEAN NOT NULL DEFAULT false,
    well VARCHAR(50),
    lease VARCHAR(50),
    orderedby VARCHAR(50),
    notes TEXT,
    customerpo VARCHAR(50),
    daterecvd TIMESTAMP,
    background VARCHAR(50),
    norm VARCHAR(50),
    services VARCHAR(50),
    billtoid VARCHAR(50),
    enteredby VARCHAR(50),
    when1 TIMESTAMP,
    trucking VARCHAR(50),
    trailer VARCHAR(50),
    inproduction TIMESTAMP,
    inspected TIMESTAMP,
    threading TIMESTAMP,
    straighten BOOLEAN NOT NULL DEFAULT false,
    excess BOOLEAN NOT NULL DEFAULT false,
    complete BOOLEAN NOT NULL DEFAULT false,
    inspectedby VARCHAR(50),
    updatedby VARCHAR(50),
    when2 TIMESTAMP,
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: rnumber
CREATE TABLE IF NOT EXISTS store.rnumber (
    rnumber INTEGER
);

-- Table: swgc
CREATE TABLE IF NOT EXISTS store.swgc (
    sizeid INTEGER,
    custid SERIAL PRIMARY KEY,
    size VARCHAR(50),
    weight VARCHAR(50),
    connection VARCHAR(50),
    pcodereceive VARCHAR(50),
    pcodeinventory VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: temp
CREATE TABLE IF NOT EXISTS store.temp (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    wkorder VARCHAR(50),
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
    wkorder VARCHAR(50),
    custid SERIAL PRIMARY KEY,
    customer VARCHAR(50),
    joints INTEGER,
    rack VARCHAR(50),
    size VARCHAR(50),
    weight VARCHAR(50),
    grade VARCHAR(50),
    connection VARCHAR(50),
    ctd BOOLEAN NOT NULL DEFAULT false,
    wstring BOOLEAN NOT NULL DEFAULT false,
    swgcc VARCHAR(50),
    color VARCHAR(50),
    customerpo VARCHAR(50),
    fletcher VARCHAR(50),
    datein TIMESTAMP,
    dateout TIMESTAMP,
    wellin VARCHAR(50),
    leasein VARCHAR(50),
    wellout VARCHAR(50),
    leaseout VARCHAR(50),
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
    userid SERIAL PRIMARY KEY,
    username VARCHAR(12),
    password VARCHAR(12),
    access INTEGER,
    fullname VARCHAR(50),
    email VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table: wknumber
CREATE TABLE IF NOT EXISTS store.wknumber (
    wknumber INTEGER
);


-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_customers_customer ON store.customers(customer);
CREATE INDEX IF NOT EXISTS idx_customers_custid ON store.customers(custid);
CREATE INDEX IF NOT EXISTS idx_inventory_custid ON store.inventory(custid);
CREATE INDEX IF NOT EXISTS idx_inventory_grade ON store.inventory(grade);
CREATE INDEX IF NOT EXISTS idx_inventory_wkorder ON store.inventory(wkorder);
CREATE INDEX IF NOT EXISTS idx_received_custid ON store.received(custid);
CREATE INDEX IF NOT EXISTS idx_received_daterecvd ON store.received(daterecvd);
CREATE INDEX IF NOT EXISTS idx_fletcher_custid ON store.fletcher(custid);
CREATE INDEX IF NOT EXISTS idx_grade_grade ON store.grade(grade);

-- Insert required oil & gas grades
INSERT INTO store.grade (grade) VALUES 
('J55'),
('JZ55'),
('L80'),
('N80'),
('P105'),
('P110')
ON CONFLICT (grade) DO NOTHING;

-- Validation: Ensure grade table contains required values
-- Expected grades: J55, JZ55, L80, N80, P105, P110


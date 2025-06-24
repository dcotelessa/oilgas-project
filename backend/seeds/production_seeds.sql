-- Production environment seed data
-- Real data from MDB migration

SET search_path TO store, public;

-- Data for bakeout
\echo 'Seeding bakeout...'
\copy store.bakeout FROM 'data/bakeout.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for customers
\echo 'Seeding customers...'
\copy store.customers FROM 'data/customers.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for fletcher
\echo 'Seeding fletcher...'
\copy store.fletcher FROM 'data/fletcher.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for grade
\echo 'Seeding grade...'
\copy store.grade FROM 'data/grade.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for inspected
\echo 'Seeding inspected...'
\copy store.inspected FROM 'data/inspected.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for inventory
\echo 'Seeding inventory...'
\copy store.inventory FROM 'data/inventory.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for received
\echo 'Seeding received...'
\copy store.received FROM 'data/received.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for rnumber
\echo 'Seeding rnumber...'
\copy store.rnumber FROM 'data/rnumber.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for swgc
\echo 'Seeding swgc...'
\copy store.swgc FROM 'data/swgc.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for temp
\echo 'Seeding temp...'
\copy store.temp FROM 'data/temp.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for tempinv
\echo 'Seeding tempinv...'
\copy store.tempinv FROM 'data/tempinv.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for test
\echo 'Seeding test...'
\copy store.test FROM 'data/test.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for users
\echo 'Seeding users...'
\copy store.users FROM 'data/users.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for wknumber
\echo 'Seeding wknumber...'
\copy store.wknumber FROM 'data/wknumber.csv' WITH (FORMAT CSV, HEADER true, NULL '');


-- Validate critical data
\echo 'Validating production seed data...'

-- Check grade table
DO $
DECLARE
    expected_grades TEXT[] := ARRAY['J55', 'JZ55', 'L80', 'N80', 'P105', 'P110'];
    missing_count INTEGER := 0;
BEGIN
    FOR i IN 1..array_length(expected_grades, 1) LOOP
        IF NOT EXISTS (SELECT 1 FROM store.grade WHERE UPPER(grade_name) = expected_grades[i]) THEN
            missing_count := missing_count + 1;
            RAISE WARNING 'Missing grade: %', expected_grades[i];
        END IF;
    END LOOP;
    
    IF missing_count = 0 THEN
        RAISE NOTICE '✅ All expected grades found';
    ELSE
        RAISE WARNING '❌ Missing % grades', missing_count;
    END IF;
END $;

-- Update statistics
ANALYZE;

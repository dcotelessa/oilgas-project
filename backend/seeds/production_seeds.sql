-- Production environment seed data template with clean column names
-- Real data from MDB migration (update paths to actual CSV files)

SET search_path TO store, public;

-- Clear existing data safely
TRUNCATE TABLE store.inventory, store.received, store.fletcher, store.bakeout CASCADE;
TRUNCATE TABLE store.inspected, store.temp, store.tempinv CASCADE;
TRUNCATE TABLE store.swgc CASCADE;
TRUNCATE TABLE store.customers CASCADE;
TRUNCATE TABLE store.users CASCADE;

-- Note: CSV files need to be updated to match new column names

-- Data for customers (update CSV headers to match new column names)
\echo 'Seeding customers...'
-- Expected columns: customer_id,customer,billing_address,billing_city,billing_state,billing_zipcode,contact,phone,fax,email,color1,color2,color3,color4,color5,loss1,loss2,loss3,loss4,loss5,wscolor1,wscolor2,wscolor3,wscolor4,wscolor5,wsloss1,wsloss2,wsloss3,wsloss4,wsloss5,deleted,created_at
\copy store.customers FROM 'data/customers_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for inventory (update CSV headers)
\echo 'Seeding inventory...'
-- Expected columns: id,username,work_order,r_number,customer_id,customer,joints,rack,size,weight,grade,connection,ctd,w_string,swgcc,color,customer_po,fletcher,date_in,date_out,well_in,lease_in,well_out,lease_out,trucking,trailer,location,notes,pcode,cn,ordered_by,deleted,created_at
\copy store.inventory FROM 'data/inventory_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for received (update CSV headers)
\echo 'Seeding received...'
-- Expected columns: id,work_order,customer_id,customer,joints,rack,size_id,size,weight,grade,connection,ctd,w_string,well,lease,ordered_by,notes,customer_po,date_received,background,norm,services,bill_to_id,entered_by,when_entered,trucking,trailer,in_production,inspected_date,threading_date,straighten_required,excess_material,complete,inspected_by,updated_by,when_updated,deleted,created_at
\copy store.received FROM 'data/received_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for fletcher (update CSV headers)  
\echo 'Seeding fletcher...'
-- Expected columns: id,username,fletcher,r_number,customer_id,customer,joints,size,weight,grade,connection,ctd,w_string,swgcc,color,customer_po,date_in,date_out,well_in,lease_in,well_out,lease_out,trucking,trailer,location,notes,pcode,cn,ordered_by,deleted,complete,created_at
\copy store.fletcher FROM 'data/fletcher_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for bakeout (update CSV headers)
\echo 'Seeding bakeout...'
-- Expected columns: id,fletcher,joints,color,size,weight,grade,connection,ctd,swgcc,customer_id,accept,reject,pin,cplg,pc,trucking,trailer,date_in,cn,created_at
\copy store.bakeout FROM 'data/bakeout_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for inspected (update CSV headers)
\echo 'Seeding inspected...'
-- Expected columns: id,username,work_order,color,joints,accept,reject,pin,cplg,pc,complete,rack,rep_pin,rep_cplg,rep_pc,deleted,cn,created_at
\copy store.inspected FROM 'data/inspected_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for grade (simple table)
\echo 'Seeding grade...'
\copy store.grade FROM 'data/grade_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for swgc (update CSV headers)
\echo 'Seeding swgc...'
-- Expected columns: size_id,customer_id,size,weight,connection,pcode_receive,pcode_inventory,created_at
\copy store.swgc FROM 'data/swgc_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for temp (update CSV headers)
\echo 'Seeding temp...'
-- Expected columns: id,username,work_order,color,joints,accept,reject,pin,cplg,pc,rack,created_at
\copy store.temp FROM 'data/temp_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for tempinv (update CSV headers)
\echo 'Seeding tempinv...'
-- Expected columns: id,username,work_order,customer_id,customer,joints,rack,size,weight,grade,connection,ctd,w_string,swgcc,color,customer_po,fletcher,date_in,date_out,well_in,lease_in,well_out,lease_out,trucking,trailer,location,notes,pcode,cn,created_at
\copy store.tempinv FROM 'data/tempinv_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for test (simple table)
\echo 'Seeding test...'
\copy store.test FROM 'data/test_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for users (update CSV headers) - Note: passwords should be properly hashed
\echo 'Seeding users...'
-- Expected columns: user_id,username,password,access,full_name,email,created_at
\copy store.users FROM 'data/users_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for r_number (simple table)
\echo 'Seeding r_number...'
\copy store.r_number FROM 'data/r_number_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Data for wk_number (simple table)
\echo 'Seeding wk_number...'
\copy store.wk_number FROM 'data/wk_number_clean.csv' WITH (FORMAT CSV, HEADER true, NULL '');

-- Validate critical data after import
\echo 'Validating production seed data...'

-- Check grade table has required values
DO $$
DECLARE
    expected_grades TEXT[] := ARRAY['J55', 'JZ55', 'K55', 'L80', 'N80', 'P105', 'P110', 'Q125', 'T95', 'C90', 'C95', 'S135'];
    missing_count INTEGER := 0;
    grade_name TEXT;
BEGIN
    FOREACH grade_name IN ARRAY expected_grades LOOP
        IF NOT EXISTS (SELECT 1 FROM store.grade WHERE UPPER(grade) = grade_name) THEN
            missing_count := missing_count + 1;
            RAISE WARNING 'Missing grade: %', grade_name;
        END IF;
    END LOOP;
    
    IF missing_count = 0 THEN
        RAISE NOTICE '✅ All expected grades found';
    ELSE
        RAISE WARNING '❌ Missing % grades', missing_count;
    END IF;
END $$;

-- Check for data integrity issues
DO $$
DECLARE
    orphaned_inventory INTEGER;
    orphaned_received INTEGER;
    invalid_dates INTEGER;
BEGIN
    -- Check for orphaned inventory records
    SELECT COUNT(*) INTO orphaned_inventory
    FROM store.inventory i
    LEFT JOIN store.customers c ON i.customer_id = c.customer_id
    WHERE c.customer_id IS NULL AND i.customer_id IS NOT NULL;
    
    -- Check for orphaned received records  
    SELECT COUNT(*) INTO orphaned_received
    FROM store.received r
    LEFT JOIN store.customers c ON r.customer_id = c.customer_id
    WHERE c.customer_id IS NULL AND r.customer_id IS NOT NULL;
    
    -- Check for invalid date ranges
    SELECT COUNT(*) INTO invalid_dates
    FROM store.inventory
    WHERE date_out IS NOT NULL AND date_in IS NOT NULL AND date_out < date_in;
    
    IF orphaned_inventory > 0 THEN
        RAISE WARNING '❌ Found % orphaned inventory records', orphaned_inventory;
    END IF;
    
    IF orphaned_received > 0 THEN
        RAISE WARNING '❌ Found % orphaned received records', orphaned_received;
    END IF;
    
    IF invalid_dates > 0 THEN
        RAISE WARNING '❌ Found % records with invalid date ranges', invalid_dates;
    END IF;
    
    IF orphaned_inventory = 0 AND orphaned_received = 0 AND invalid_dates = 0 THEN
        RAISE NOTICE '✅ Data integrity checks passed';
    END IF;
END $$;

-- Generate summary statistics
DO $$
DECLARE
    total_customers INTEGER;
    total_inventory INTEGER;
    total_received INTEGER;
    active_customers INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_customers FROM store.customers;
    SELECT COUNT(*) INTO active_customers FROM store.customers WHERE deleted = false;
    SELECT COUNT(*) INTO total_inventory FROM store.inventory WHERE deleted = false;
    SELECT COUNT(*) INTO total_received FROM store.received WHERE deleted = false;
    
    RAISE NOTICE '📊 Production Data Summary:';
    RAISE NOTICE '   Customers: % (% active)', total_customers, active_customers;
    RAISE NOTICE '   Inventory Items: %', total_inventory;
    RAISE NOTICE '   Received Items: %', total_received;
END $$;

-- Update statistics for query optimization
ANALYZE;

\echo '✅ Production seed data import completed';

-- Notes for production deployment:
-- 1. CSV files must be updated to use clean column names (customer_id instead of custid, etc.)
-- 2. User passwords should be properly hashed with bcrypt before import
-- 3. Ensure CSV files are placed in backend/data/ directory
-- 4. Run data validation queries after import
-- 5. Test critical queries with real data volumes
-- 6. Set up regular ANALYZE schedules for production
-- 7. Monitor index usage and query performance
-- 8. Consider partitioning large tables by date if needed

-- CSV Header Mapping Reference:
-- OLD → NEW column names for CSV conversion:
-- custid → customer_id
-- datein → date_in  
-- dateout → date_out
-- customerpo → customer_po
-- wkorder → work_order
-- rnumber → r_number
-- wellin → well_in
-- leasein → lease_in
-- wellout → well_out
-- leaseout → lease_out
-- billingaddress → billing_address
-- billingcity → billing_city
-- billingstate → billing_state
-- billingzipcode → billing_zipcode
-- userid → user_id
-- fullname → full_name
-- daterecvd → date_received
-- orderedby → ordered_by
-- enteredby → entered_by
-- inspectedby → inspected_by
-- updatedby → updated_by
-- And many more - see the model definitions for complete mapping
--

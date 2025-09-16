-- database/init/colorado/02_colorado_functions.sql
-- Colorado Database Functions and Triggers

-- NOTE: Work order functions removed - not part of Customer domain
-- These will be implemented in WorkOrder domain

-- NOTE: Invoice functions removed - not part of Customer domain

-- NOTE: Work order total functions removed

-- Customer audit trail function
CREATE OR REPLACE FUNCTION customer_audit_trigger()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO store.customer_audit (
        customer_id,
        action,
        old_values,
        new_values,
        changed_by_user_id,
        created_at
    ) VALUES (
        COALESCE(NEW.id, OLD.id),
        TG_OP,
        CASE WHEN TG_OP = 'DELETE' THEN row_to_json(OLD) ELSE NULL END,
        CASE WHEN TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN row_to_json(NEW) ELSE NULL END,
        NULL, -- Will be set by application context
        NOW()
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- NOTE: Work order audit functions removed

-- NOTE: Work order history functions removed

-- NOTE: Work order and invoice triggers removed

-- Customer domain audit triggers only
CREATE TRIGGER trg_customers_audit
    AFTER INSERT OR UPDATE OR DELETE ON store.customers
    FOR EACH ROW
    EXECUTE FUNCTION customer_audit_trigger();

-- Customer contacts audit
CREATE TRIGGER trg_customer_contacts_audit
    AFTER INSERT OR UPDATE OR DELETE ON store.customer_contacts
    FOR EACH ROW
    EXECUTE FUNCTION customer_audit_trigger();

-- Helper functions for Customer domain

-- Function to validate company code format
CREATE OR REPLACE FUNCTION validate_company_code(p_code VARCHAR(50))
RETURNS BOOLEAN AS $$
BEGIN
    RETURN p_code ~ '^[A-Za-z0-9]{2,10}$';
END;
$$ LANGUAGE plpgsql;

-- Function to ensure unique company codes within tenant
CREATE OR REPLACE FUNCTION check_unique_company_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.company_code IS NOT NULL THEN
        IF NOT validate_company_code(NEW.company_code) THEN
            RAISE EXCEPTION 'Company code must be alphanumeric, 2-10 characters';
        END IF;
        
        IF EXISTS (
            SELECT 1 FROM store.customers 
            WHERE tenant_id = NEW.tenant_id 
            AND company_code = NEW.company_code 
            AND id != COALESCE(NEW.id, -1)
        ) THEN
            RAISE EXCEPTION 'Company code already exists in tenant';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for company code validation
CREATE TRIGGER trg_validate_company_code
    BEFORE INSERT OR UPDATE ON store.customers
    FOR EACH ROW
    EXECUTE FUNCTION check_unique_company_code();
# Customer Database Schema

## Table: store.customers

Standardized customer table with no abbreviations:

- `customer_id` - Primary key (auto-generated)
- `original_customer_id` - Reference to MDB custid
- `customer_name` - Company name (required)
- `billing_address` - Full billing address
- `billing_city` - Billing city
- `billing_state` - 2-letter state code
- `billing_zip_code` - ZIP/postal code
- `contact_name` - Primary contact person
- `phone_number` - Formatted phone number
- `email_address` - Email address
- Color grades: `color_grade_1` through `color_grade_5`
- Wall losses: `wall_loss_1` through `wall_loss_5`
- W-String colors: `wstring_color_1` through `wstring_color_5`
- W-String losses: `wstring_loss_1` through `wstring_loss_5`
- `is_deleted` - Soft delete flag
- `tenant_id` - Multi-tenant isolation

## Features
- No abbreviations in column names
- Automatic duplicate detection
- Multi-tenant row-level security
- Data validation and cleaning

# Customer Migration Quick Guide

## Setup
1. Place `petros-lb.mdb` in `db_prep/` folder
2. Run: `make setup-customers`
3. Verify: `make verify-customers`

## Commands
- `make check-mdb` - Verify MDB file
- `make analyze-mdb` - Export customers from Access
- `make setup-customers` - Complete workflow
- `make verify-customers` - Check results

## Files
- `db_prep/petros-lb.mdb` - Your Access database
- `customers.csv` - Exported data (auto-generated)
- PostgreSQL database with standardized schema

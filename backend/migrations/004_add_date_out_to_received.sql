-- migrations/004_add_date_out_to_received.sql
-- Add date_out column to track shipping dates

-- Up migration
ALTER TABLE store.received ADD COLUMN date_out TIMESTAMP;

-- Add index for performance on date_out queries
CREATE INDEX idx_received_date_out ON store.received(date_out) WHERE date_out IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN store.received.date_out IS 'Timestamp when item was shipped/sent out';

-- Down migration (for rollback)
-- ALTER TABLE store.received DROP COLUMN date_out;
-- DROP INDEX IF EXISTS idx_received_date_out;

BEGIN;
ALTER TABLE tenants ADD COLUMN opening_time VARCHAR(5), ADD COLUMN closing_time VARCHAR(5);
ALTER TABLE tenants ADD CONSTRAINT tenants_opening_hours_valid CHECK (
 (opening_time IS NULL AND closing_time IS NULL) OR
 (opening_time IS NOT NULL AND closing_time IS NOT NULL AND opening_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'
 AND closing_time ~ '^(([01][0-9]|2[0-3]):[0-5][0-9]|24:00)$' AND opening_time < closing_time)
);
COMMIT;

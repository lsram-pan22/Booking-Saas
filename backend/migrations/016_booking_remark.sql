BEGIN;
ALTER TABLE bookings RENAME COLUMN cancellation_remark TO remark;
ALTER TABLE bookings RENAME CONSTRAINT bookings_cancellation_remark_check TO bookings_remark_check;
COMMIT;

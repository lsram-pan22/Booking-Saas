BEGIN;
ALTER TABLE bookings ADD COLUMN cancellation_remark TEXT NOT NULL DEFAULT '' CHECK (LENGTH(cancellation_remark) <= 1000);
ALTER TABLE booking_status_history ADD COLUMN remark TEXT NOT NULL DEFAULT '' CHECK (LENGTH(remark) <= 1000);
-- Keep historical cancellations intact; require a reason for new cancellations.
ALTER TABLE bookings ADD CONSTRAINT bookings_cancel_remark_required CHECK (status <> 'cancelled' OR LENGTH(TRIM(cancellation_remark)) > 0) NOT VALID;
ALTER TABLE booking_status_history ADD CONSTRAINT booking_history_cancel_remark_required CHECK (to_status <> 'cancelled' OR LENGTH(TRIM(remark)) > 0) NOT VALID;
COMMIT;

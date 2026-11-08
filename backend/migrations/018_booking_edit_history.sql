BEGIN;
CREATE TABLE booking_edit_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id),
    changed_by UUID NOT NULL REFERENCES users(id),
    before_data JSONB NOT NULL,
    after_data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX booking_edit_history_booking_idx ON booking_edit_history(booking_id,created_at);
COMMIT;

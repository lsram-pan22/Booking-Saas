BEGIN;
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    service_id UUID NOT NULL REFERENCES services(id),
    staff_member_id UUID NOT NULL REFERENCES tenant_members(id),
    customer_name VARCHAR(150) NOT NULL CHECK (LENGTH(TRIM(customer_name)) > 0),
    customer_phone VARCHAR(30) NOT NULL,
    note TEXT NOT NULL DEFAULT '' CHECK (LENGTH(note) <= 2000),
    service_name VARCHAR(150) NOT NULL,
    staff_name VARCHAR(150) NOT NULL,
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes BETWEEN 1 AND 1440),
    price_satang BIGINT NOT NULL CHECK (price_satang BETWEEN 0 AND 100000000),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','confirmed','completed','cancelled')),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at = starts_at + duration_minutes * INTERVAL '1 minute')
);
CREATE INDEX bookings_tenant_time_idx ON bookings(tenant_id,starts_at);
CREATE INDEX bookings_staff_time_idx ON bookings(tenant_id,staff_member_id,starts_at,ends_at) WHERE status <> 'cancelled';
CREATE TABLE booking_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id),
    from_status VARCHAR(20) CHECK (from_status IN ('pending','confirmed','completed','cancelled')),
    to_status VARCHAR(20) NOT NULL CHECK (to_status IN ('pending','confirmed','completed','cancelled')),
    changed_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX booking_status_history_booking_idx ON booking_status_history(booking_id,created_at);
COMMIT;

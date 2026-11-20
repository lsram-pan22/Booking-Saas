BEGIN;
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(150) NOT NULL CHECK (LENGTH(TRIM(name)) > 0),
    description TEXT NOT NULL DEFAULT '' CHECK (LENGTH(description) <= 2000),
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes BETWEEN 1 AND 1440),
    price_satang BIGINT NOT NULL CHECK (price_satang BETWEEN 0 AND 100000000),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX services_tenant_idx ON services(tenant_id,created_at,id);
COMMIT;

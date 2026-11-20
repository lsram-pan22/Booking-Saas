BEGIN;
-- Baseline for a new database. Later migrations add booking and security fields.
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(254) NOT NULL,
    email VARCHAR(254) NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(150) NOT NULL,
    phone VARCHAR(30),
    system_role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (system_role IN ('user','super_admin')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','suspended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX users_username_lower_unique ON users(LOWER(username));
CREATE UNIQUE INDEX users_email_lower_unique ON users(LOWER(email));

CREATE TABLE packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    monthly_price_satang BIGINT NOT NULL DEFAULT 0 CHECK (monthly_price_satang BETWEEN 0 AND 100000000),
    shop_limit INTEGER NOT NULL DEFAULT 1 CHECK (shop_limit BETWEEN 1 AND 100000),
    user_limit INTEGER NOT NULL DEFAULT 2 CHECK (user_limit BETWEEN 1 AND 100000),
    staff_management_level VARCHAR(20) NOT NULL DEFAULT 'basic' CHECK (staff_management_level IN ('basic','full')),
    report_level VARCHAR(20) NOT NULL DEFAULT 'basic' CHECK (report_level IN ('basic','advanced')),
    export_excel BOOLEAN NOT NULL DEFAULT FALSE,
    api_webhook BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
    sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order BETWEEN 0 AND 100000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    package_id UUID NOT NULL REFERENCES packages(id),
    package_code VARCHAR(50) NOT NULL,
    package_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','expired','cancelled','inactive')),
    monthly_price_satang BIGINT NOT NULL CHECK (monthly_price_satang BETWEEN 0 AND 100000000),
    shop_limit INTEGER NOT NULL CHECK (shop_limit BETWEEN 1 AND 100000),
    user_limit INTEGER NOT NULL CHECK (user_limit BETWEEN 1 AND 100000),
    staff_management_level VARCHAR(20) NOT NULL CHECK (staff_management_level IN ('basic','full')),
    report_level VARCHAR(20) NOT NULL CHECK (report_level IN ('basic','advanced')),
    export_excel BOOLEAN NOT NULL,
    api_webhook BOOLEAN NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at IS NULL OR ends_at > starts_at)
);

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL UNIQUE,
    business_type VARCHAR(20) NOT NULL CHECK (business_type IN ('salon','spa','clinic','fitness','other')),
    timezone VARCHAR(100) NOT NULL DEFAULT 'Asia/Bangkok',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('pending','active','inactive','suspended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tenant_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner','staff')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id,user_id)
);
CREATE UNIQUE INDEX tenant_members_one_staff_shop ON tenant_members(user_id) WHERE role='staff';
CREATE UNIQUE INDEX tenant_members_one_owner_per_shop ON tenant_members(tenant_id) WHERE role='owner';
CREATE INDEX tenant_members_user_idx ON tenant_members(user_id);

-- Bootstrap only the required Free plan; paid plans remain admin-configurable.
INSERT INTO packages(code,name,description,monthly_price_satang,shop_limit,user_limit)
VALUES ('free','Free','เริ่มต้นใช้งาน 1 ร้าน ผู้ใช้รวมเจ้าของร้าน 2 คน',0,1,2);
COMMIT;

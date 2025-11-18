-- Create TenantPreferences and UserPreferences tables
-- Migration 005: Preferences system for theming

-- 1. Create TenantPreferences table
CREATE TABLE IF NOT EXISTS "TenantPreferences" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    theme TEXT NOT NULL DEFAULT 'light',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_tenant_preferences_tenant FOREIGN KEY (tenant_id)
        REFERENCES "Tenant"(id) ON DELETE CASCADE,
    CONSTRAINT tenant_preferences_tenant_id_unique UNIQUE (tenant_id)
);

CREATE INDEX IF NOT EXISTS ix_TenantPreferences_tenant_id ON "TenantPreferences"(tenant_id);

-- 2. Create UserPreferences table
CREATE TABLE IF NOT EXISTS "UserPreferences" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    theme TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_preferences_user FOREIGN KEY (user_id)
        REFERENCES "User"(id) ON DELETE CASCADE,
    CONSTRAINT user_preferences_user_id_unique UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS ix_UserPreferences_user_id ON "UserPreferences"(user_id);

-- 3. Create default TenantPreferences for all existing tenants
INSERT INTO "TenantPreferences" (tenant_id, theme, created_at, updated_at)
SELECT id, 'light', NOW(), NOW()
FROM "Tenant"
ON CONFLICT (tenant_id) DO NOTHING;

-- 4. Create default UserPreferences for all existing users (theme=null to inherit)
INSERT INTO "UserPreferences" (user_id, theme, created_at, updated_at)
SELECT id, NULL, NOW(), NOW()
FROM "User"
ON CONFLICT (user_id) DO NOTHING;

-- 5. Drop Tenant.settings column
ALTER TABLE "Tenant" DROP COLUMN IF EXISTS settings;

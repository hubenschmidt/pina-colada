-- Create Tenant_Preferences and User_Preferences tables
-- Migration 005: Preferences system for theming

-- 1. Create Tenant_Preferences table
CREATE TABLE IF NOT EXISTS "Tenant_Preferences" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    theme TEXT NOT NULL DEFAULT 'light',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_tenant_preferences_tenant FOREIGN KEY (tenant_id)
        REFERENCES "Tenant"(id) ON DELETE CASCADE,
    CONSTRAINT tenant_preferences_tenant_id_unique UNIQUE (tenant_id)
);

CREATE INDEX IF NOT EXISTS ix_Tenant_Preferences_tenant_id ON "Tenant_Preferences"(tenant_id);

-- 2. Create User_Preferences table
CREATE TABLE IF NOT EXISTS "User_Preferences" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    theme TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_preferences_user FOREIGN KEY (user_id)
        REFERENCES "User"(id) ON DELETE CASCADE,
    CONSTRAINT user_preferences_user_id_unique UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS ix_User_Preferences_user_id ON "User_Preferences"(user_id);

-- 3. Create default Tenant_Preferences for all existing tenants
INSERT INTO "Tenant_Preferences" (tenant_id, theme, created_at, updated_at)
SELECT id, 'light', NOW(), NOW()
FROM "Tenant"
ON CONFLICT (tenant_id) DO NOTHING;

-- 4. Create default User_Preferences for all existing users (theme=null to inherit)
INSERT INTO "User_Preferences" (user_id, theme, created_at, updated_at)
SELECT id, NULL, NOW(), NOW()
FROM "User"
ON CONFLICT (user_id) DO NOTHING;

-- 5. Drop Tenant.settings column
ALTER TABLE "Tenant" DROP COLUMN IF EXISTS settings;

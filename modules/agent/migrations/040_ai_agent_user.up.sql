-- Migration 040: AI Agent User
-- Creates an AI agent system user per tenant to record AI actions in the system
-- The AI agent user has a special email format: ai-agent@{tenant_slug}.system

-- Add is_system_user flag to User table to identify system/AI users
ALTER TABLE "User" ADD COLUMN IF NOT EXISTS is_system_user BOOLEAN DEFAULT FALSE;

-- Create AI agent user for each existing tenant
INSERT INTO "User" (tenant_id, email, first_name, last_name, status, is_system_user, created_at, updated_at)
SELECT
    t.id,
    'ai-agent@' || t.slug || '.system',
    'AI',
    'Agent',
    'active',
    TRUE,
    NOW(),
    NOW()
FROM "Tenant" t
WHERE NOT EXISTS (
    SELECT 1 FROM "User" u
    WHERE u.tenant_id = t.id
    AND u.is_system_user = TRUE
);

-- Create index for quick lookup of system users
CREATE INDEX IF NOT EXISTS idx_user_is_system_user ON "User" (tenant_id, is_system_user) WHERE is_system_user = TRUE;

-- Show created AI agent users
SELECT u.id, u.email, u.first_name, u.last_name, t.name as tenant_name
FROM "User" u
JOIN "Tenant" t ON u.tenant_id = t.id
WHERE u.is_system_user = TRUE;

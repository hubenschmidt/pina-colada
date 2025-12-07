-- Add developer role for analytics access (global role with NULL tenant_id)
-- This is idempotent - safe to run multiple times
INSERT INTO "Role" (tenant_id, name, description)
SELECT NULL, 'developer', 'Developer access with analytics and debugging tools'
WHERE NOT EXISTS (
    SELECT 1 FROM "Role" WHERE name = 'developer' AND tenant_id IS NULL
);

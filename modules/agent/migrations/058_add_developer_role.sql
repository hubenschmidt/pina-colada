-- Ensure default tenant exists (required before creating tenant-scoped roles)
INSERT INTO "Tenant" (id, name, slug, plan, created_at, updated_at)
VALUES (1, 'PinaColada', 'pinacolada', 'free', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Add developer role for tenant 1
-- This is idempotent - safe to run multiple times
INSERT INTO "Role" (tenant_id, name, description)
SELECT 1, 'developer', 'Developer access with analytics and debugging tools'
WHERE NOT EXISTS (
    SELECT 1 FROM "Role" WHERE name = 'developer' AND tenant_id = 1
);

-- Assign admin and developer roles to whubenschmidt@gmail.com
INSERT INTO "User_Role" (user_id, role_id, created_at)
SELECT u.id, r.id, NOW()
FROM "User" u
CROSS JOIN "Role" r
WHERE u.email = 'whubenschmidt@gmail.com'
  AND r.name IN ('admin', 'developer')
  AND r.tenant_id = 1
ON CONFLICT (user_id, role_id) DO NOTHING;

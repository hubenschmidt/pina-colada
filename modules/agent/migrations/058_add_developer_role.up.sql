-- Ensure default tenant exists (required before creating tenant-scoped roles)
INSERT INTO "Tenant" (id, name, slug, plan, created_at, updated_at)
VALUES (1, 'PinaColada', 'pinacolada', 'free', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Create base roles for tenant 1
INSERT INTO "Role" (tenant_id, name, description)
SELECT 1, 'owner', 'Tenant owner with full access and billing control'
WHERE NOT EXISTS (SELECT 1 FROM "Role" WHERE name = 'owner' AND tenant_id = 1);

INSERT INTO "Role" (tenant_id, name, description)
SELECT 1, 'admin', 'Administrator with full access to all features'
WHERE NOT EXISTS (SELECT 1 FROM "Role" WHERE name = 'admin' AND tenant_id = 1);

INSERT INTO "Role" (tenant_id, name, description)
SELECT 1, 'member', 'Standard member with access to most features'
WHERE NOT EXISTS (SELECT 1 FROM "Role" WHERE name = 'member' AND tenant_id = 1);

INSERT INTO "Role" (tenant_id, name, description)
SELECT 1, 'viewer', 'Read-only access to view data'
WHERE NOT EXISTS (SELECT 1 FROM "Role" WHERE name = 'viewer' AND tenant_id = 1);

INSERT INTO "Role" (tenant_id, name, description)
SELECT 1, 'developer', 'Developer access with analytics and debugging tools'
WHERE NOT EXISTS (SELECT 1 FROM "Role" WHERE name = 'developer' AND tenant_id = 1);

-- Create bootstrap user
INSERT INTO "User" (tenant_id, email, first_name, last_name, status, created_at, updated_at)
VALUES (1, 'whubenschmidt@gmail.com', 'William', 'Hubenschmidt', 'active', NOW(), NOW())
ON CONFLICT (tenant_id, email) DO NOTHING;

-- Assign owner, admin, and developer roles to whubenschmidt@gmail.com
INSERT INTO "User_Role" (user_id, role_id, created_at)
SELECT u.id, r.id, NOW()
FROM "User" u
CROSS JOIN "Role" r
WHERE u.email = 'whubenschmidt@gmail.com'
  AND u.tenant_id = 1
  AND r.tenant_id = 1
  AND r.name IN ('owner', 'admin', 'developer')
ON CONFLICT (user_id, role_id) DO NOTHING;

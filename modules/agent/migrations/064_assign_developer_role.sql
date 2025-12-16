-- Assign developer role to whubenschmidt@gmail.com
-- This is idempotent - safe to run multiple times
INSERT INTO "User_Role" (user_id, role_id)
SELECT u.id, r.id
FROM "User" u
CROSS JOIN "Role" r
WHERE u.email = 'whubenschmidt@gmail.com'
  AND r.name = 'developer'
  AND r.tenant_id IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM "User_Role" ur
    WHERE ur.user_id = u.id AND ur.role_id = r.id
  );

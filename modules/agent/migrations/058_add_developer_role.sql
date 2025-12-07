-- Add developer role for analytics access
INSERT INTO "Role" (tenant_id, name, description)
VALUES (NULL, 'developer', 'Developer access with analytics and debugging tools');

-- Assign to William (user_id=1)
INSERT INTO "User_Role" (user_id, role_id)
SELECT 1, id FROM "Role" WHERE name = 'developer' AND tenant_id IS NULL;

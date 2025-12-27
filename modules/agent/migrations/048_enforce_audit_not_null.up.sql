-- Enforce NOT NULL on user audit columns (created_by, updated_by)
-- Use Jennifer Lev (existing prod user) to populate legacy records

-- Add updated_by to Comment table (was missing)
ALTER TABLE "Comment" ADD COLUMN IF NOT EXISTS updated_by INTEGER REFERENCES "User"(id);

-- Populate NULL audit columns using Jennifer Lev (jennifervlev@gmail.com) per tenant
-- Account
UPDATE "Account" a SET created_by = (
  SELECT id FROM "User" WHERE tenant_id = a.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Account" a SET updated_by = (
  SELECT id FROM "User" WHERE tenant_id = a.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Asset
UPDATE "Asset" a SET created_by = (
  SELECT id FROM "User" WHERE tenant_id = a.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Asset" a SET updated_by = (
  SELECT id FROM "User" WHERE tenant_id = a.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Contact (no tenant_id, get Jennifer from tenant 1)
UPDATE "Contact" SET created_by = (
  SELECT id FROM "User" WHERE email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Contact" SET updated_by = (
  SELECT id FROM "User" WHERE email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Deal
UPDATE "Deal" d SET created_by = (
  SELECT id FROM "User" WHERE tenant_id = d.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Deal" d SET updated_by = (
  SELECT id FROM "User" WHERE tenant_id = d.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Individual (no tenant_id directly, get from account)
UPDATE "Individual" i SET created_by = (
  SELECT u.id FROM "User" u
  JOIN "Account" a ON u.tenant_id = a.tenant_id
  WHERE a.id = i.account_id AND u.email = 'jennifervlev@gmail.com'
  LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Individual" i SET updated_by = (
  SELECT u.id FROM "User" u
  JOIN "Account" a ON u.tenant_id = a.tenant_id
  WHERE a.id = i.account_id AND u.email = 'jennifervlev@gmail.com'
  LIMIT 1
) WHERE updated_by IS NULL;

-- Lead
UPDATE "Lead" l SET created_by = (
  SELECT id FROM "User" WHERE tenant_id = l.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Lead" l SET updated_by = (
  SELECT id FROM "User" WHERE tenant_id = l.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Note
UPDATE "Note" n SET created_by = (
  SELECT id FROM "User" WHERE tenant_id = n.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Note" n SET updated_by = (
  SELECT id FROM "User" WHERE tenant_id = n.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Organization (no tenant_id directly, get from account)
UPDATE "Organization" o SET created_by = (
  SELECT u.id FROM "User" u
  JOIN "Account" a ON u.tenant_id = a.tenant_id
  WHERE a.id = o.account_id AND u.email = 'jennifervlev@gmail.com'
  LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Organization" o SET updated_by = (
  SELECT u.id FROM "User" u
  JOIN "Account" a ON u.tenant_id = a.tenant_id
  WHERE a.id = o.account_id AND u.email = 'jennifervlev@gmail.com'
  LIMIT 1
) WHERE updated_by IS NULL;

-- Project
UPDATE "Project" p SET created_by = (
  SELECT id FROM "User" WHERE tenant_id = p.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Project" p SET updated_by = (
  SELECT id FROM "User" WHERE tenant_id = p.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Task
UPDATE "Task" t SET created_by = (
  SELECT id FROM "User" WHERE tenant_id = t.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Task" t SET updated_by = (
  SELECT id FROM "User" WHERE tenant_id = t.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Comment
UPDATE "Comment" c SET created_by = (
  SELECT id FROM "User" WHERE tenant_id = c.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE created_by IS NULL;
UPDATE "Comment" c SET updated_by = (
  SELECT id FROM "User" WHERE tenant_id = c.tenant_id AND email = 'jennifervlev@gmail.com' LIMIT 1
) WHERE updated_by IS NULL;

-- Now add NOT NULL constraints
ALTER TABLE "Account" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Account" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Asset" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Asset" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Contact" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Contact" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Deal" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Deal" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Individual" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Individual" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Lead" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Lead" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Note" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Note" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Organization" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Organization" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Project" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Project" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Task" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Task" ALTER COLUMN updated_by SET NOT NULL;

ALTER TABLE "Comment" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Comment" ALTER COLUMN updated_by SET NOT NULL;

-- Add index for Comment audit columns
CREATE INDEX IF NOT EXISTS idx_comment_created_by ON "Comment"(created_by);

-- Move organization relationship from Job/Opportunity/Partnership up to Lead as Account
-- Lead → Account (which can be Organization or Individual)

-- 1. Add account_id to Lead
ALTER TABLE "Lead" ADD COLUMN IF NOT EXISTS account_id BIGINT REFERENCES "Account"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_lead_account_id ON "Lead"(account_id);

-- 2. Migrate existing organization_id from Job to Lead.account_id
-- Find the Account for each Organization and set it on the Lead
UPDATE "Lead" l
SET account_id = o.account_id
FROM "Job" j
JOIN "Organization" o ON j.organization_id = o.id
WHERE l.id = j.id AND l.account_id IS NULL;

-- 3. Migrate existing organization_id from Opportunity to Lead.account_id
UPDATE "Lead" l
SET account_id = o.account_id
FROM "Opportunity" op
JOIN "Organization" o ON op.organization_id = o.id
WHERE l.id = op.id AND l.account_id IS NULL;

-- 4. Migrate existing organization_id from Partnership to Lead.account_id
UPDATE "Lead" l
SET account_id = o.account_id
FROM "Partnership" p
JOIN "Organization" o ON p.organization_id = o.id
WHERE l.id = p.id AND l.account_id IS NULL;

-- 5. Drop organization_id from Job, Opportunity, Partnership
ALTER TABLE "Job" DROP COLUMN IF EXISTS organization_id;
ALTER TABLE "Opportunity" DROP COLUMN IF EXISTS organization_id;
ALTER TABLE "Partnership" DROP COLUMN IF EXISTS organization_id;

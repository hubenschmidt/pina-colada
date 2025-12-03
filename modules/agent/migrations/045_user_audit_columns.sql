-- Add user audit columns (created_by, updated_by) to core business models
-- These track which user created/modified each record for audit trails and AI workflows

-- Account
ALTER TABLE "Account" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Account" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Asset (Document inherits from Asset via joined table inheritance)
ALTER TABLE "Asset" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Asset" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Contact
ALTER TABLE "Contact" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Contact" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Deal
ALTER TABLE "Deal" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Deal" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Individual
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Lead (Job, Opportunity, Partnership inherit from Lead via joined table inheritance)
ALTER TABLE "Lead" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Lead" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Note (already has created_by, only add updated_by)
ALTER TABLE "Note" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Organization
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Project
ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Task
ALTER TABLE "Task" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;
ALTER TABLE "Task" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL;

-- Create indexes for efficient querying by creator/updater
CREATE INDEX IF NOT EXISTS idx_account_created_by ON "Account"(created_by);
CREATE INDEX IF NOT EXISTS idx_asset_created_by ON "Asset"(created_by);
CREATE INDEX IF NOT EXISTS idx_contact_created_by ON "Contact"(created_by);
CREATE INDEX IF NOT EXISTS idx_deal_created_by ON "Deal"(created_by);
CREATE INDEX IF NOT EXISTS idx_individual_created_by ON "Individual"(created_by);
CREATE INDEX IF NOT EXISTS idx_lead_created_by ON "Lead"(created_by);
CREATE INDEX IF NOT EXISTS idx_organization_created_by ON "Organization"(created_by);
CREATE INDEX IF NOT EXISTS idx_project_created_by ON "Project"(created_by);
CREATE INDEX IF NOT EXISTS idx_task_created_by ON "Task"(created_by);

-- Migration: Rename Job model to Lead
-- Renames Job table to Lead (existing data becomes lead data)
-- Updates all references, triggers, policies, indexes, and constraints

-- Step 1: Rename Job table to Lead
ALTER TABLE "Job" RENAME TO "Lead";

-- Step 2: Drop and recreate the status CHECK constraint with updated name
-- Drop both possible constraint names (old migration used job_status_check, SQLAlchemy uses check_status)
ALTER TABLE "Lead" DROP CONSTRAINT IF EXISTS job_status_check;
ALTER TABLE "Lead" DROP CONSTRAINT IF EXISTS check_status;
ALTER TABLE "Lead" ADD CONSTRAINT lead_status_check
    CHECK (status IN ('New', 'Qualified', 'Discovery', 'Proposal', 'Negotiation', 'Won', 'Lost', 'On Hold'));

-- Step 3: Update trigger name for renamed table
DROP TRIGGER IF EXISTS update_job_updated_at ON "Lead"; -- Old trigger name from initial schema
CREATE TRIGGER update_lead_updated_at
    BEFORE UPDATE ON "Lead"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Step 4: Update RLS policies for renamed table
ALTER TABLE "Lead" ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS "Enable read access for all users" ON "Lead";
CREATE POLICY "Enable read access for all users"
    ON "Lead" FOR SELECT
    USING (true);

DROP POLICY IF EXISTS "Enable insert for all users" ON "Lead";
CREATE POLICY "Enable insert for all users"
    ON "Lead" FOR INSERT
    WITH CHECK (true);

DROP POLICY IF EXISTS "Enable update for all users" ON "Lead";
CREATE POLICY "Enable update for all users"
    ON "Lead" FOR UPDATE
    USING (true);

DROP POLICY IF EXISTS "Enable delete for all users" ON "Lead";
CREATE POLICY "Enable delete for all users"
    ON "Lead" FOR DELETE
    USING (true);

-- Step 5: Update indexes for renamed table
-- Note: Keeping idx_company_job_title index name for backward compatibility, but it now indexes Lead table
DROP INDEX IF EXISTS idx_company_job_title;
CREATE INDEX IF NOT EXISTS idx_company_job_title ON "Lead"(company, job_title);
DROP INDEX IF EXISTS idx_date;
CREATE INDEX IF NOT EXISTS idx_date ON "Lead"(date DESC);
DROP INDEX IF EXISTS idx_status;
CREATE INDEX IF NOT EXISTS idx_status ON "Lead"(status);

-- Step 6: Add comments for documentation
COMMENT ON COLUMN "Lead".status IS 'CRM stage status: New, Qualified, Discovery, Proposal, Negotiation, Won, Lost, On Hold';
COMMENT ON COLUMN "Lead".source IS 'Source of lead entry: manual (user-entered) or agent (AI-generated lead)';


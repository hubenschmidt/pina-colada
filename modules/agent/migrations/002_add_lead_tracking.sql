-- Migration: Add lead tracking functionality
-- Adds support for job leads with qualification statuses and expanded status values
-- Uses proper LeadStatus table with foreign key relationship (not CHECK constraint)

-- Step 1: Drop and recreate the status CHECK constraint to include new values
-- Drop both possible constraint names (migration uses job_status_check, SQLAlchemy uses check_status)
ALTER TABLE "Job" DROP CONSTRAINT IF EXISTS job_status_check;
ALTER TABLE "Job" DROP CONSTRAINT IF EXISTS check_status;
ALTER TABLE "Job" ADD CONSTRAINT job_status_check
    CHECK (status IN ('lead', 'applied', 'interviewing', 'rejected', 'offer', 'accepted', 'do_not_apply'));

-- Step 2: Create LeadStatus table (must exist before foreign key reference)
CREATE TABLE IF NOT EXISTS "LeadStatus" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Step 3: Insert initial lead statuses
INSERT INTO "LeadStatus" (name, description) VALUES
    ('Qualifying', 'Initial lead from agent - needs qualification'),
    ('Cold', 'Cold lead - minimal prior engagement'),
    ('Warm', 'Warm lead - some prior engagement or interest'),
    ('Hot', 'Hot lead - high interest and engagement')
ON CONFLICT (name) DO NOTHING;

-- Step 4: Add foreign key column to Job table
ALTER TABLE "Job" ADD COLUMN IF NOT EXISTS lead_status_id UUID REFERENCES "LeadStatus"(id);

-- Step 5: Create index on lead_status_id for efficient filtering (only index non-null values)
CREATE INDEX IF NOT EXISTS idx_job_lead_status_id ON "Job"(lead_status_id) WHERE lead_status_id IS NOT NULL;

-- Step 6: Add comments for documentation
COMMENT ON TABLE "LeadStatus" IS 'Statuses of job leads for qualification tracking';
COMMENT ON COLUMN "Job".status IS 'Status of job application: lead (not yet applied), applied, interviewing, rejected, offer, accepted, do_not_apply';
COMMENT ON COLUMN "Job".lead_status_id IS 'Foreign key to LeadStatus table. NULL if not an active lead.';
COMMENT ON COLUMN "Job".source IS 'Source of job entry: manual (user-entered) or agent (AI-generated lead)';

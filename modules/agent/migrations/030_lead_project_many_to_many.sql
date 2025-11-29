-- Migration: Convert Lead-Project from one-to-many to many-to-many
-- Creates LeadProject junction table

-- Create junction table
CREATE TABLE IF NOT EXISTS "LeadProject" (
    lead_id BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (lead_id, project_id)
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_lead_project_lead_id ON "LeadProject"(lead_id);
CREATE INDEX IF NOT EXISTS idx_lead_project_project_id ON "LeadProject"(project_id);

-- Migrate existing project_id data to junction table
INSERT INTO "LeadProject" (lead_id, project_id)
SELECT id, project_id FROM "Lead" WHERE project_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- Drop the old project_id column and index from Lead
DROP INDEX IF EXISTS idx_lead_project_id;
ALTER TABLE "Lead" DROP COLUMN IF EXISTS project_id;

-- Migration: Add project_id to Saved_Report for project-specific reports
-- project_id = NULL means global report (visible regardless of selected project)
-- project_id = <id> means project-specific report (only visible when that project is selected)

ALTER TABLE "Saved_Report"
ADD COLUMN IF NOT EXISTS project_id BIGINT REFERENCES "Project"(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_saved_reports_project ON "Saved_Report"(project_id);

COMMENT ON COLUMN "Saved_Report".project_id IS 'NULL for global reports, set for project-specific reports';

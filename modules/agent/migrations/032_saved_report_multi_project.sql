-- Migration: Convert Saved_Report from single project_id to many-to-many with projects
-- This allows reports to be scoped to multiple projects or be global (no projects = global)

-- Create junction table
CREATE TABLE IF NOT EXISTS "Saved_Report_Project" (
    saved_report_id BIGINT NOT NULL REFERENCES "Saved_Report"(id) ON DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (saved_report_id, project_id)
);

CREATE INDEX IF NOT EXISTS idx_saved_report_project_report ON "Saved_Report_Project"(saved_report_id);
CREATE INDEX IF NOT EXISTS idx_saved_report_project_project ON "Saved_Report_Project"(project_id);

-- Migrate existing data from project_id column to junction table
INSERT INTO "Saved_Report_Project" (saved_report_id, project_id, created_at)
SELECT id, project_id, NOW()
FROM "Saved_Report"
WHERE project_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- Drop the old project_id column
ALTER TABLE "Saved_Report" DROP COLUMN IF EXISTS project_id;

-- Drop the old index
DROP INDEX IF EXISTS idx_saved_reports_project;

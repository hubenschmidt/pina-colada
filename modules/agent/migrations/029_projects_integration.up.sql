-- Migration 029: Projects Integration
-- Enhances Project table and links Deals/Leads to Projects

-- 1. Add current_status_id to Project table
ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS current_status_id BIGINT;
ALTER TABLE "Project" ADD CONSTRAINT fk_project_current_status
    FOREIGN KEY (current_status_id) REFERENCES "Status"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_project_current_status_id ON "Project"(current_status_id);

-- 2. Add project_id to Deal table
ALTER TABLE "Deal" ADD COLUMN IF NOT EXISTS project_id BIGINT;
ALTER TABLE "Deal" ADD CONSTRAINT fk_deal_project
    FOREIGN KEY (project_id) REFERENCES "Project"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_deal_project_id ON "Deal"(project_id);

-- 3. Add project_id to Lead table
ALTER TABLE "Lead" ADD COLUMN IF NOT EXISTS project_id BIGINT;
ALTER TABLE "Lead" ADD CONSTRAINT fk_lead_project
    FOREIGN KEY (project_id) REFERENCES "Project"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_lead_project_id ON "Lead"(project_id);

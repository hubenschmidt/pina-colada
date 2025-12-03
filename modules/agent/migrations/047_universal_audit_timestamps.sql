-- ==============================================
-- UNIVERSAL AUDIT TIMESTAMPS
-- Add created_at and updated_at to all tables missing them
-- ==============================================

-- Tables missing updated_at (already have created_at)
ALTER TABLE "Job_Status" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Lead_Status" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Deal_Status" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Task_Status" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Task_Priority" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE "Lead_Organization" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Lead_Individual" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Project_Deal" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Project_Lead" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Project_Contact" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Project_Organization" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Project_Individual" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE "Role" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "User_Role" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE "Technology" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Organization_Technology" ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Organization_Technology" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE "Funding_Round" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Company_Signal" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE "Industry" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Account_Industry" ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Account_Industry" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Account_Project" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE "Entity_Tag" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Contact_Individual" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Contact_Organization" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Lead_Project" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE "Saved_Report_Project" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Comment_Notification" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Tables missing both created_at and updated_at
ALTER TABLE "Tag" ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE "Tag" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Entity_Asset from 041_asset_refactor (has created_at only)
ALTER TABLE "Entity_Asset" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- ==============================================
-- BACKFILL: Set updated_at = created_at for existing records
-- (Records never updated should have updated_at equal to created_at)
-- ==============================================

-- Status config tables (had created_at)
UPDATE "Job_Status" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Lead_Status" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Deal_Status" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Task_Status" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Task_Priority" SET updated_at = created_at WHERE updated_at > created_at;

-- Junction tables (had created_at)
UPDATE "Lead_Organization" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Lead_Individual" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Project_Deal" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Project_Lead" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Project_Contact" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Project_Organization" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Project_Individual" SET updated_at = created_at WHERE updated_at > created_at;

-- Auth tables (had created_at)
UPDATE "Role" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "User_Role" SET updated_at = created_at WHERE updated_at > created_at;

-- Enrichment tables (had created_at)
UPDATE "Technology" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Organization_Technology" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Funding_Round" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Company_Signal" SET updated_at = created_at WHERE updated_at > created_at;

-- Reference tables (had created_at)
UPDATE "Industry" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Account_Industry" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Account_Project" SET updated_at = created_at WHERE updated_at > created_at;

-- Tag/Entity tables (Tag got both, but Entity_Tag had created_at)
UPDATE "Entity_Tag" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Entity_Asset" SET updated_at = created_at WHERE updated_at > created_at;

-- Contact junction tables (had created_at)
UPDATE "Contact_Individual" SET updated_at = created_at WHERE updated_at > created_at;
UPDATE "Contact_Organization" SET updated_at = created_at WHERE updated_at > created_at;

-- Lead/Project junction (had created_at)
UPDATE "Lead_Project" SET updated_at = created_at WHERE updated_at > created_at;

-- Saved reports (had created_at)
UPDATE "Saved_Report_Project" SET updated_at = created_at WHERE updated_at > created_at;

-- Notifications (had created_at)
UPDATE "Comment_Notification" SET updated_at = created_at WHERE updated_at > created_at;

-- ==============================================
-- CREATE TRIGGERS FOR AUTO-UPDATING updated_at
-- ==============================================

-- Status config tables
CREATE TRIGGER update_job_status_updated_at BEFORE UPDATE ON "Job_Status"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_lead_status_updated_at BEFORE UPDATE ON "Lead_Status"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_deal_status_updated_at BEFORE UPDATE ON "Deal_Status"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_task_status_updated_at BEFORE UPDATE ON "Task_Status"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_task_priority_updated_at BEFORE UPDATE ON "Task_Priority"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Junction tables
CREATE TRIGGER update_lead_organization_updated_at BEFORE UPDATE ON "Lead_Organization"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_lead_individual_updated_at BEFORE UPDATE ON "Lead_Individual"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_project_deal_updated_at BEFORE UPDATE ON "Project_Deal"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_project_lead_updated_at BEFORE UPDATE ON "Project_Lead"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_project_contact_updated_at BEFORE UPDATE ON "Project_Contact"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_project_organization_updated_at BEFORE UPDATE ON "Project_Organization"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_project_individual_updated_at BEFORE UPDATE ON "Project_Individual"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Auth tables
CREATE TRIGGER update_role_updated_at BEFORE UPDATE ON "Role"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_role_updated_at BEFORE UPDATE ON "User_Role"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Enrichment tables
CREATE TRIGGER update_technology_updated_at BEFORE UPDATE ON "Technology"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_organization_technology_updated_at BEFORE UPDATE ON "Organization_Technology"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_funding_round_updated_at BEFORE UPDATE ON "Funding_Round"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_company_signal_updated_at BEFORE UPDATE ON "Company_Signal"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Reference tables
CREATE TRIGGER update_industry_updated_at BEFORE UPDATE ON "Industry"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_account_industry_updated_at BEFORE UPDATE ON "Account_Industry"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_account_project_updated_at BEFORE UPDATE ON "Account_Project"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Tag/Entity tables
CREATE TRIGGER update_tag_updated_at BEFORE UPDATE ON "Tag"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entity_tag_updated_at BEFORE UPDATE ON "Entity_Tag"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entity_asset_updated_at BEFORE UPDATE ON "Entity_Asset"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Contact junction tables
CREATE TRIGGER update_contact_individual_updated_at BEFORE UPDATE ON "Contact_Individual"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_contact_organization_updated_at BEFORE UPDATE ON "Contact_Organization"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Lead/Project junction
CREATE TRIGGER update_lead_project_updated_at BEFORE UPDATE ON "Lead_Project"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Saved reports
CREATE TRIGGER update_saved_report_project_updated_at BEFORE UPDATE ON "Saved_Report_Project"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Notifications
CREATE TRIGGER update_comment_notification_updated_at BEFORE UPDATE ON "Comment_Notification"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

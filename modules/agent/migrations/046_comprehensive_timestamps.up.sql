-- Migration 046: Add missing updated_at/created_at columns to all tables
-- Ensures model-database schema parity for timestamp columns

-- Data_Provenance
ALTER TABLE "Data_Provenance" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_data_provenance_updated_at ON "Data_Provenance";
CREATE TRIGGER update_data_provenance_updated_at
    BEFORE UPDATE ON "Data_Provenance"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Role
ALTER TABLE "Role" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Role" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_role_updated_at ON "Role";
CREATE TRIGGER update_role_updated_at
    BEFORE UPDATE ON "Role"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- User_Role
ALTER TABLE "User_Role" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "User_Role" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_user_role_updated_at ON "User_Role";
CREATE TRIGGER update_user_role_updated_at
    BEFORE UPDATE ON "User_Role"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Organization_Technology
ALTER TABLE "Organization_Technology" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Organization_Technology" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_org_technology_updated_at ON "Organization_Technology";
CREATE TRIGGER update_org_technology_updated_at
    BEFORE UPDATE ON "Organization_Technology"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Funding_Round
ALTER TABLE "Funding_Round" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Funding_Round" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_funding_round_updated_at ON "Funding_Round";
CREATE TRIGGER update_funding_round_updated_at
    BEFORE UPDATE ON "Funding_Round"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Technology
ALTER TABLE "Technology" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Technology" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_technology_updated_at ON "Technology";
CREATE TRIGGER update_technology_updated_at
    BEFORE UPDATE ON "Technology"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Industry
ALTER TABLE "Industry" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Industry" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_industry_updated_at ON "Industry";
CREATE TRIGGER update_industry_updated_at
    BEFORE UPDATE ON "Industry"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Account_Industry
ALTER TABLE "Account_Industry" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Account_Industry" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_account_industry_updated_at ON "Account_Industry";
CREATE TRIGGER update_account_industry_updated_at
    BEFORE UPDATE ON "Account_Industry"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Tag
ALTER TABLE "Tag" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Tag" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_tag_updated_at ON "Tag";
CREATE TRIGGER update_tag_updated_at
    BEFORE UPDATE ON "Tag"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Entity_Tag
ALTER TABLE "Entity_Tag" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Entity_Tag" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_entity_tag_updated_at ON "Entity_Tag";
CREATE TRIGGER update_entity_tag_updated_at
    BEFORE UPDATE ON "Entity_Tag"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Company_Signal
ALTER TABLE "Company_Signal" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Company_Signal" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_company_signal_updated_at ON "Company_Signal";
CREATE TRIGGER update_company_signal_updated_at
    BEFORE UPDATE ON "Company_Signal"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Lead_Project
ALTER TABLE "Lead_Project" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Lead_Project" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_lead_project_updated_at ON "Lead_Project";
CREATE TRIGGER update_lead_project_updated_at
    BEFORE UPDATE ON "Lead_Project"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Entity_Asset
ALTER TABLE "Entity_Asset" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Entity_Asset" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_entity_asset_updated_at ON "Entity_Asset";
CREATE TRIGGER update_entity_asset_updated_at
    BEFORE UPDATE ON "Entity_Asset"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Account_Project
ALTER TABLE "Account_Project" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Account_Project" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_account_project_updated_at ON "Account_Project";
CREATE TRIGGER update_account_project_updated_at
    BEFORE UPDATE ON "Account_Project"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Saved_Report_Project
ALTER TABLE "Saved_Report_Project" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Saved_Report_Project" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_saved_report_project_updated_at ON "Saved_Report_Project";
CREATE TRIGGER update_saved_report_project_updated_at
    BEFORE UPDATE ON "Saved_Report_Project"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Employee_Count_Range
ALTER TABLE "Employee_Count_Range" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Employee_Count_Range" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_employee_count_range_updated_at ON "Employee_Count_Range";
CREATE TRIGGER update_employee_count_range_updated_at
    BEFORE UPDATE ON "Employee_Count_Range"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Funding_Stage
ALTER TABLE "Funding_Stage" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Funding_Stage" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_funding_stage_updated_at ON "Funding_Stage";
CREATE TRIGGER update_funding_stage_updated_at
    BEFORE UPDATE ON "Funding_Stage"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Revenue_Range
ALTER TABLE "Revenue_Range" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Revenue_Range" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_revenue_range_updated_at ON "Revenue_Range";
CREATE TRIGGER update_revenue_range_updated_at
    BEFORE UPDATE ON "Revenue_Range"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Salary_Range
ALTER TABLE "Salary_Range" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Salary_Range" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_salary_range_updated_at ON "Salary_Range";
CREATE TRIGGER update_salary_range_updated_at
    BEFORE UPDATE ON "Salary_Range"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Status
ALTER TABLE "Status" ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE "Status" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
DROP TRIGGER IF EXISTS update_status_updated_at ON "Status";
CREATE TRIGGER update_status_updated_at
    BEFORE UPDATE ON "Status"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

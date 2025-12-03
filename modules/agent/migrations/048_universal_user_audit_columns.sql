-- ==============================================
-- UNIVERSAL USER AUDIT COLUMNS
-- Add created_by and updated_by to all tables
-- ==============================================

-- Reference/Lookup tables
ALTER TABLE "Status" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Status" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Industry" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Industry" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Technology" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Technology" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Employee_Count_Range" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Employee_Count_Range" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Funding_Stage" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Funding_Stage" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Revenue_Range" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Revenue_Range" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Salary_Range" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Salary_Range" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Tag" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Tag" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Enrichment tables
ALTER TABLE "Organization_Technology" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Organization_Technology" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Funding_Round" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Funding_Round" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Company_Signal" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Company_Signal" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Data_Provenance" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Data_Provenance" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Junction tables
ALTER TABLE "Account_Industry" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Account_Industry" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Account_Project" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Account_Project" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Lead_Project" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Lead_Project" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Entity_Tag" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Entity_Tag" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Entity_Asset" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Entity_Asset" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Saved_Report_Project" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Saved_Report_Project" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Contact_Individual" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Contact_Individual" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Contact_Organization" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Contact_Organization" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Auth/System tables
ALTER TABLE "Role" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Role" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "User_Role" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "User_Role" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Tenant" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Tenant" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "User" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "User" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Tenant_Preferences" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Tenant_Preferences" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "User_Preferences" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "User_Preferences" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Comment/Notification tables
ALTER TABLE "Comment" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Comment_Notification" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Comment_Notification" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Saved reports - fix created_by FK (was Individual, should be User)
ALTER TABLE "Saved_Report" DROP CONSTRAINT IF EXISTS "Saved_Report_created_by_fkey";
ALTER TABLE "Saved_Report" ADD CONSTRAINT "Saved_Report_created_by_fkey" FOREIGN KEY (created_by) REFERENCES "User"(id);
ALTER TABLE "Saved_Report" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Document table inherits created_by/updated_by from Asset (joined table inheritance)

-- Lead child tables (joined table inheritance - each has own audit columns)
ALTER TABLE "Job" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Job" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Opportunity" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Opportunity" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Partnership" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Partnership" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Status config tables
ALTER TABLE "Job_Status" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Job_Status" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Lead_Status" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Lead_Status" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Deal_Status" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Deal_Status" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Task_Status" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Task_Status" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Task_Priority" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Task_Priority" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Legacy junction tables from initial schema
ALTER TABLE "Lead_Organization" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Lead_Organization" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Lead_Individual" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Lead_Individual" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Project_Deal" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Project_Deal" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Project_Lead" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Project_Lead" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Project_Contact" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Project_Contact" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Project_Organization" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Project_Organization" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Project_Individual" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Project_Individual" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- Activity and relationship tables
ALTER TABLE "Activity" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Activity" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Organization_Relationship" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Organization_Relationship" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

ALTER TABLE "Individual_Relationship" ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES "User"(id);
ALTER TABLE "Individual_Relationship" ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES "User"(id);

-- ==============================================
-- BACKFILL EXISTING DATA WITH JENNIFER LEV
-- Assigns Jennifer Lev as creator for legacy records
-- ==============================================

DO $$
DECLARE
    legacy_user_id BIGINT;
BEGIN
    -- Get Jennifer Lev's user ID
    SELECT id INTO legacy_user_id FROM "User"
    WHERE first_name = 'Jennifer' AND last_name = 'Lev' LIMIT 1;

    IF legacy_user_id IS NULL THEN
        RAISE NOTICE 'Jennifer Lev user not found, skipping backfill';
        RETURN;
    END IF;

    -- Core business tables (from user-audit-columns spec)
    UPDATE "Account" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Asset" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Contact" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Deal" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    -- Document inherits from Asset, no separate backfill needed
    UPDATE "Individual" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Job" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Lead" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Note" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Opportunity" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Organization" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Partnership" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Project" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Task" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Reference/Lookup tables
    UPDATE "Status" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Industry" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Technology" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Employee_Count_Range" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Funding_Stage" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Revenue_Range" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Salary_Range" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Tag" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Enrichment tables
    UPDATE "Organization_Technology" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Funding_Round" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Company_Signal" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Data_Provenance" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Junction tables
    UPDATE "Account_Industry" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Account_Project" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Lead_Project" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Entity_Tag" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Entity_Asset" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Saved_Report_Project" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Contact_Individual" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Contact_Organization" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Auth/System tables
    UPDATE "Role" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "User_Role" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Tenant" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "User" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Tenant_Preferences" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "User_Preferences" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Comment/Notification tables
    UPDATE "Comment_Notification" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Saved reports
    UPDATE "Saved_Report" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Status config tables
    UPDATE "Job_Status" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Lead_Status" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Deal_Status" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Task_Status" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Task_Priority" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Legacy junction tables
    UPDATE "Lead_Organization" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Lead_Individual" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Project_Deal" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Project_Lead" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Project_Contact" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Project_Organization" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Project_Individual" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    -- Activity and relationship tables
    UPDATE "Activity" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Organization_Relationship" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;
    UPDATE "Individual_Relationship" SET created_by = legacy_user_id, updated_by = legacy_user_id WHERE created_by IS NULL;

    RAISE NOTICE 'Backfilled created_by and updated_by with Jennifer Lev user ID: %', legacy_user_id;
END $$;

-- ==============================================
-- ADD NOT NULL CONSTRAINTS
-- After backfill, make audit columns required
-- ==============================================

-- Core business tables
ALTER TABLE "Account" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Account" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Asset" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Asset" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Contact" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Contact" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Deal" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Deal" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Individual" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Individual" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Job" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Job" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Lead" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Lead" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Note" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Note" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Opportunity" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Opportunity" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Organization" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Organization" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Partnership" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Partnership" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Project" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Project" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Task" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Task" ALTER COLUMN updated_by SET NOT NULL;

-- Reference/Lookup tables
ALTER TABLE "Status" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Status" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Industry" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Industry" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Technology" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Technology" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Employee_Count_Range" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Employee_Count_Range" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Funding_Stage" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Funding_Stage" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Revenue_Range" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Revenue_Range" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Salary_Range" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Salary_Range" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Tag" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Tag" ALTER COLUMN updated_by SET NOT NULL;

-- Enrichment tables
ALTER TABLE "Organization_Technology" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Organization_Technology" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Funding_Round" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Funding_Round" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Company_Signal" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Company_Signal" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Data_Provenance" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Data_Provenance" ALTER COLUMN updated_by SET NOT NULL;

-- Junction tables
ALTER TABLE "Account_Industry" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Account_Industry" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Account_Project" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Account_Project" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Lead_Project" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Lead_Project" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Entity_Tag" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Entity_Tag" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Entity_Asset" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Entity_Asset" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Saved_Report_Project" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Saved_Report_Project" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Contact_Individual" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Contact_Individual" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Contact_Organization" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Contact_Organization" ALTER COLUMN updated_by SET NOT NULL;

-- Auth/System tables
ALTER TABLE "Role" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Role" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "User_Role" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "User_Role" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Tenant" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Tenant" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "User" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "User" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Tenant_Preferences" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Tenant_Preferences" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "User_Preferences" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "User_Preferences" ALTER COLUMN updated_by SET NOT NULL;

-- Comment/Notification tables
ALTER TABLE "Comment" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Comment" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Comment_Notification" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Comment_Notification" ALTER COLUMN updated_by SET NOT NULL;

-- Saved reports
ALTER TABLE "Saved_Report" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Saved_Report" ALTER COLUMN updated_by SET NOT NULL;

-- Status config tables
ALTER TABLE "Job_Status" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Job_Status" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Lead_Status" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Lead_Status" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Deal_Status" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Deal_Status" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Task_Status" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Task_Status" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Task_Priority" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Task_Priority" ALTER COLUMN updated_by SET NOT NULL;

-- Legacy junction tables
ALTER TABLE "Lead_Organization" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Lead_Organization" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Lead_Individual" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Lead_Individual" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Project_Deal" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Project_Deal" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Project_Lead" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Project_Lead" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Project_Contact" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Project_Contact" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Project_Organization" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Project_Organization" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Project_Individual" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Project_Individual" ALTER COLUMN updated_by SET NOT NULL;

-- Activity and relationship tables
ALTER TABLE "Activity" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Activity" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Organization_Relationship" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Organization_Relationship" ALTER COLUMN updated_by SET NOT NULL;
ALTER TABLE "Individual_Relationship" ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE "Individual_Relationship" ALTER COLUMN updated_by SET NOT NULL;

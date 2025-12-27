-- Drop unused status configuration tables
-- These tables are not used - Status.category handles filtering instead

DROP TABLE IF EXISTS "Job_Status" CASCADE;
DROP TABLE IF EXISTS "Lead_Status" CASCADE;
DROP TABLE IF EXISTS "Deal_Status" CASCADE;
DROP TABLE IF EXISTS "Task_Status" CASCADE;
DROP TABLE IF EXISTS "Task_Priority" CASCADE;

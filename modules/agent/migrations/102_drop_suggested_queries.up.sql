-- Remove unused suggested_queries column from run log
ALTER TABLE "Automation_Run_Log" DROP COLUMN IF EXISTS suggested_queries;

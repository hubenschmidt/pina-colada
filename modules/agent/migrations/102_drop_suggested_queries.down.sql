-- Restore suggested_queries column
ALTER TABLE "Automation_Run_Log" ADD COLUMN IF NOT EXISTS suggested_queries TEXT;

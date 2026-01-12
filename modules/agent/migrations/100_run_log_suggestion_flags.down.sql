-- Remove suggestion tracking flags from run log
ALTER TABLE "Automation_Run_Log"
DROP COLUMN IF EXISTS query_updated,
DROP COLUMN IF EXISTS prompt_updated;

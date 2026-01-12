-- Add suggestion tracking flags to run log for auditability
ALTER TABLE "Automation_Run_Log"
ADD COLUMN IF NOT EXISTS query_updated BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS prompt_updated BOOLEAN NOT NULL DEFAULT false;

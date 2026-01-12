-- Add executed_system_prompt to Automation_Run_Log for audit trail
ALTER TABLE "Automation_Run_Log"
ADD COLUMN IF NOT EXISTS executed_system_prompt TEXT;

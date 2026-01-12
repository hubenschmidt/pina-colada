-- Remove executed_system_prompt from Automation_Run_Log
ALTER TABLE "Automation_Run_Log"
DROP COLUMN IF EXISTS executed_system_prompt;

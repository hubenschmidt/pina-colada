-- Remove character count column
ALTER TABLE "Automation_Run_Log"
DROP COLUMN IF EXISTS executed_system_prompt_charcount;

-- Add character count for executed system prompt (for audit debugging)
ALTER TABLE "Automation_Run_Log"
ADD COLUMN IF NOT EXISTS executed_system_prompt_charcount INT NOT NULL DEFAULT 0;

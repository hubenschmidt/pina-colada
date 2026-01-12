-- Remove prompt cooldown fields
ALTER TABLE "Automation_Config"
DROP COLUMN IF EXISTS prompt_cooldown_runs,
DROP COLUMN IF EXISTS prompt_cooldown_prospects;

-- Add prompt cooldown configuration to prevent excessive prompt updates
ALTER TABLE "Automation_Config"
ADD COLUMN IF NOT EXISTS prompt_cooldown_runs INT NOT NULL DEFAULT 5,
ADD COLUMN IF NOT EXISTS prompt_cooldown_prospects INT NOT NULL DEFAULT 50;

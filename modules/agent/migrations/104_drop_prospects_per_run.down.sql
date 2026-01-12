-- Restore prospects_per_run column
ALTER TABLE "Automation_Config" ADD COLUMN IF NOT EXISTS prospects_per_run INT NOT NULL DEFAULT 10;

-- Remove min_prospects_threshold column
ALTER TABLE "Automation_Config"
DROP COLUMN IF EXISTS min_prospects_threshold;

-- Remove analytics configuration from Automation_Config
ALTER TABLE "Automation_Config"
DROP COLUMN IF EXISTS use_analytics,
DROP COLUMN IF EXISTS analytics_model;

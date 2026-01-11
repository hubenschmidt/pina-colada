-- Add analytics configuration to Automation_Config
ALTER TABLE "Automation_Config"
ADD COLUMN IF NOT EXISTS use_analytics BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS analytics_model VARCHAR(50);

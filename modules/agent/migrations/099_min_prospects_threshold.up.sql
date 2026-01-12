-- Add min_prospects_threshold to trigger suggestions when prospects are too low
ALTER TABLE "Automation_Config"
ADD COLUMN IF NOT EXISTS min_prospects_threshold INT NOT NULL DEFAULT 5;

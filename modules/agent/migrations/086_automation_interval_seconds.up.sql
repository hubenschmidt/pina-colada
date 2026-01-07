-- Rename interval_minutes to interval_seconds and convert existing values
ALTER TABLE "Automation_Config" RENAME COLUMN interval_minutes TO interval_seconds;

-- Convert existing minute values to seconds (multiply by 60)
UPDATE "Automation_Config" SET interval_seconds = interval_seconds * 60;

-- Update default from 30 (minutes) to 1800 (seconds)
ALTER TABLE "Automation_Config" ALTER COLUMN interval_seconds SET DEFAULT 1800;

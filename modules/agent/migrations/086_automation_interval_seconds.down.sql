-- Convert seconds back to minutes (divide by 60)
UPDATE "Automation_Config" SET interval_seconds = interval_seconds / 60;

-- Rename back to interval_minutes
ALTER TABLE "Automation_Config" RENAME COLUMN interval_seconds TO interval_minutes;

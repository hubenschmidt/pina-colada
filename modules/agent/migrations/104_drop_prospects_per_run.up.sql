-- Remove unused prospects_per_run column (Serper determines result count)
ALTER TABLE "Automation_Config" DROP COLUMN IF EXISTS prospects_per_run;

ALTER TABLE "Automation_Config"
RENAME COLUMN prospects_per_run TO leads_per_run;

ALTER TABLE "Automation_Run_Log"
RENAME COLUMN prospects_found TO leads_found;

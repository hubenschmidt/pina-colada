ALTER TABLE "Automation_Config"
RENAME COLUMN leads_per_run TO prospects_per_run;

ALTER TABLE "Automation_Run_Log"
RENAME COLUMN leads_found TO prospects_found;

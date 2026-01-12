-- Rename Automation_Run_Log fields for clarity
ALTER TABLE "Automation_Run_Log"
RENAME COLUMN search_query TO executed_query;

ALTER TABLE "Automation_Run_Log"
RENAME COLUMN suggested_queries TO related_searches;

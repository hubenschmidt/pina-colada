-- Add location field to Automation_Config
-- LLM can only modify search terms, not location
ALTER TABLE "Automation_Config" ADD COLUMN location TEXT;

-- Add suggested_queries to run log to track LLM-suggested searches
ALTER TABLE "Automation_Run_Log" ADD COLUMN suggested_queries TEXT;

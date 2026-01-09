ALTER TABLE "Automation_Config" ADD COLUMN consecutive_zero_runs INT NOT NULL DEFAULT 0;
ALTER TABLE "Automation_Config" ADD COLUMN empty_proposal_limit INT NOT NULL DEFAULT 0;

DROP INDEX IF EXISTS idx_agent_proposal_automation_config;
ALTER TABLE "Agent_Proposal" DROP COLUMN IF EXISTS automation_config_id;

-- Add automation_config_id to Agent_Proposal for tracking proposals per crawler
ALTER TABLE "Agent_Proposal"
ADD COLUMN automation_config_id BIGINT REFERENCES "Automation_Config"(id) ON DELETE SET NULL;

-- Index for counting pending proposals by config
CREATE INDEX idx_agent_proposal_automation_config
ON "Agent_Proposal"(automation_config_id, status)
WHERE automation_config_id IS NOT NULL;

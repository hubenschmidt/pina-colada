-- Add source column to Agent_Proposal to track where proposals came from
ALTER TABLE "Agent_Proposal" ADD COLUMN IF NOT EXISTS source VARCHAR(50);

-- Create index for efficient source-based queries
CREATE INDEX IF NOT EXISTS idx_agent_proposal_source ON "Agent_Proposal"(source) WHERE source IS NOT NULL;

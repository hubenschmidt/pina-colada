-- Add fallback_chain column to Agent_Node_Config for model promotion
ALTER TABLE "Agent_Node_Config"
ADD COLUMN fallback_chain JSONB;

-- fallback_chain format: [{"model": "gpt-5.1", "timeout_seconds": 15}, {"model": "gpt-5.2", "timeout_seconds": 10}]
-- When null, uses default chains defined in code

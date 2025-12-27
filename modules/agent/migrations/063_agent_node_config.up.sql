-- Agent node model configuration per user
CREATE TABLE IF NOT EXISTS "Agent_Node_Config" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    node_name TEXT NOT NULL,
    model TEXT NOT NULL,
    provider TEXT NOT NULL DEFAULT 'openai',
    temperature DECIMAL(3,2),
    max_tokens INTEGER,
    top_p DECIMAL(3,2),
    top_k INTEGER,
    frequency_penalty DECIMAL(3,2),
    presence_penalty DECIMAL(3,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT agent_node_config_user_node_unique UNIQUE (user_id, node_name)
);

CREATE INDEX IF NOT EXISTS ix_agent_node_config_user_id ON "Agent_Node_Config"(user_id);

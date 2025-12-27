-- Agent config presets for saving/applying LLM parameter settings
CREATE TABLE IF NOT EXISTS "Agent_Config_Preset" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    temperature DECIMAL(3,2),
    max_tokens INTEGER,
    top_p DECIMAL(3,2),
    top_k INTEGER,
    frequency_penalty DECIMAL(3,2),
    presence_penalty DECIMAL(3,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT agent_config_preset_user_name_unique UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS ix_agent_config_preset_user_id ON "Agent_Config_Preset"(user_id);

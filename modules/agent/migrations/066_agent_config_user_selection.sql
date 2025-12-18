-- User's selected presets for agent config
CREATE TABLE IF NOT EXISTS "Agent_Config_User_Selection" (
    user_id BIGINT PRIMARY KEY REFERENCES "User"(id) ON DELETE CASCADE,
    selected_param_preset_id BIGINT REFERENCES "Agent_Config_Preset"(id) ON DELETE SET NULL,
    selected_cost_tier VARCHAR(20) DEFAULT 'premium',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

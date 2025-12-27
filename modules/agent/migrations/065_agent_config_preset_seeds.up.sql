-- Make user_id nullable for global presets
ALTER TABLE "Agent_Config_Preset"
    ALTER COLUMN user_id DROP NOT NULL;

-- Drop existing unique constraint and recreate to handle NULLs
ALTER TABLE "Agent_Config_Preset"
    DROP CONSTRAINT IF EXISTS agent_config_preset_user_name_unique;

-- Create unique index that handles NULL user_id (global presets)
CREATE UNIQUE INDEX agent_config_preset_user_name_unique
    ON "Agent_Config_Preset" (COALESCE(user_id, 0), name);

-- Insert global presets (user_id = NULL)
INSERT INTO "Agent_Config_Preset" (user_id, name, temperature, max_tokens, top_p, top_k, frequency_penalty, presence_penalty) VALUES
-- General purpose presets
(NULL, 'Creative', 1.20, NULL, 0.95, NULL, 0.30, 0.30),
(NULL, 'Precise', 0.30, NULL, 0.50, 40, 0.00, 0.00),
(NULL, 'Balanced', 0.70, NULL, 0.90, NULL, 0.00, 0.00),

-- Role-specific presets
(NULL, 'Deep Researcher', 0.40, 4096, 0.70, 50, 0.20, 0.10),
(NULL, 'Tech Recruiter', 0.60, 2048, 0.85, NULL, 0.10, 0.20),
(NULL, 'Sales Agent', 0.80, 2048, 0.90, NULL, 0.20, 0.30),
(NULL, 'Technical Writer', 0.35, 4096, 0.60, 40, 0.10, 0.00),
(NULL, 'Customer Support', 0.40, 1024, 0.70, NULL, 0.00, 0.10),
(NULL, 'Creative Writer', 1.10, 4096, 0.95, NULL, 0.40, 0.40),
(NULL, 'Data Analyst', 0.20, 2048, 0.40, 30, 0.00, 0.00);

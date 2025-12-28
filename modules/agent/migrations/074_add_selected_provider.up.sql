ALTER TABLE "Agent_Config_User_Selection"
ADD COLUMN IF NOT EXISTS selected_provider VARCHAR(50) DEFAULT 'openai';

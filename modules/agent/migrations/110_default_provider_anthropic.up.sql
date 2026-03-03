-- Switch all existing OpenAI provider configs to Anthropic
UPDATE "Agent_Node_Config" SET provider = 'anthropic', model = 'claude-sonnet-4-5-20250929' WHERE provider = 'openai';
UPDATE "Agent_Config_User_Selection" SET selected_provider = 'anthropic' WHERE selected_provider = 'openai';
ALTER TABLE "Agent_Node_Config" ALTER COLUMN provider SET DEFAULT 'anthropic';
ALTER TABLE "Agent_Config_User_Selection" ALTER COLUMN selected_provider SET DEFAULT 'anthropic';

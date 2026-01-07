ALTER TABLE "Automation_Config"
ADD COLUMN use_agent BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN agent_model VARCHAR(50);

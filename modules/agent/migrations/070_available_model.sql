-- Create Available_Model table for database-driven model configuration
CREATE TABLE "Available_Model" (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    default_timeout_seconds INT DEFAULT 20,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider, model_name)
);

CREATE INDEX idx_available_model_provider ON "Available_Model"(provider);
CREATE INDEX idx_available_model_active ON "Available_Model"(is_active);

-- Seed data: OpenAI models (longer timeouts: 30s/20s/15s)
INSERT INTO "Available_Model" (provider, model_name, display_name, sort_order, default_timeout_seconds) VALUES
('openai', 'gpt-5.2', 'GPT 5.2', 1, 15),
('openai', 'gpt-5.1', 'GPT 5.1', 2, 20),
('openai', 'gpt-5', 'GPT 5', 3, 30),
('openai', 'gpt-5-mini', 'GPT 5 Mini', 4, 15),
('openai', 'gpt-5-nano', 'GPT 5 Nano', 5, 10),
('openai', 'gpt-4.1', 'GPT 4.1', 6, 20),
('openai', 'gpt-4.1-mini', 'GPT 4.1 Mini', 7, 15),
('openai', 'gpt-4o', 'GPT 4o', 8, 20),
('openai', 'gpt-4o-mini', 'GPT 4o Mini', 9, 15),
('openai', 'o3', 'O3', 10, 45),
('openai', 'o4-mini', 'O4 Mini', 11, 20);

-- Seed data: Anthropic models
INSERT INTO "Available_Model" (provider, model_name, display_name, sort_order, default_timeout_seconds) VALUES
('anthropic', 'claude-sonnet-4-5-20250929', 'Claude Sonnet 4.5', 1, 20),
('anthropic', 'claude-opus-4-5-20251101', 'Claude Opus 4.5', 2, 45),
('anthropic', 'claude-haiku-4-5-20251001', 'Claude Haiku 4.5', 3, 10);

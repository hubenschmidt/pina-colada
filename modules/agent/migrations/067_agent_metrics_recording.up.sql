-- Recording sessions (manual start/stop by developers)
CREATE TABLE IF NOT EXISTS "Agent_Recording_Session" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    name VARCHAR(255),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    config_snapshot JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Individual metric data points (one per agent turn)
CREATE TABLE IF NOT EXISTS "Agent_Metric" (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES "Agent_Recording_Session"(id) ON DELETE CASCADE,
    conversation_id BIGINT REFERENCES "Conversation"(id) ON DELETE SET NULL,
    thread_id VARCHAR(255),

    -- Timing
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_ms INT NOT NULL,

    -- Tokens
    input_tokens INT NOT NULL,
    output_tokens INT NOT NULL,
    total_tokens INT NOT NULL,

    -- Cost (calculated from model pricing)
    estimated_cost_usd DECIMAL(10, 6),

    -- Model/Config used
    model VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    node_name VARCHAR(100),

    -- Full config snapshot for this turn
    config_snapshot JSONB,

    -- User input (for same prompt comparisons)
    user_message TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_metric_session ON "Agent_Metric"(session_id);
CREATE INDEX IF NOT EXISTS idx_agent_metric_started ON "Agent_Metric"(started_at);
CREATE INDEX IF NOT EXISTS idx_agent_recording_session_user ON "Agent_Recording_Session"(user_id);
CREATE INDEX IF NOT EXISTS idx_agent_recording_session_active ON "Agent_Recording_Session"(user_id, ended_at) WHERE ended_at IS NULL;

-- Usage analytics table for tracking token spend by node/tool
CREATE TABLE "Usage_Analytics" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id),
    user_id BIGINT NOT NULL REFERENCES "User"(id),
    conversation_id BIGINT REFERENCES "Conversation"(id),
    message_id BIGINT REFERENCES "Conversation_Message"(id),

    -- Token counts
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,

    -- Breakdown for developer analytics
    node_name TEXT,
    tool_name TEXT,
    model_name TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_usage_analytics_tenant_created ON "Usage_Analytics"(tenant_id, created_at);
CREATE INDEX idx_usage_analytics_user_created ON "Usage_Analytics"(user_id, created_at);
CREATE INDEX idx_usage_analytics_node ON "Usage_Analytics"(node_name) WHERE node_name IS NOT NULL;
CREATE INDEX idx_usage_analytics_tool ON "Usage_Analytics"(tool_name) WHERE tool_name IS NOT NULL;
CREATE INDEX idx_usage_analytics_model ON "Usage_Analytics"(model_name) WHERE model_name IS NOT NULL;

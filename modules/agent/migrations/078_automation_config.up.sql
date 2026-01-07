-- Automation configuration - supports multiple crawlers per user
CREATE TABLE "Automation_Config" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id),
    user_id BIGINT NOT NULL REFERENCES "User"(id),

    -- Crawler Identity
    name VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL DEFAULT 'job',
    enabled BOOLEAN NOT NULL DEFAULT false,

    -- Scheduling
    interval_minutes INT NOT NULL DEFAULT 30,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    run_count INT NOT NULL DEFAULT 0,

    -- Search Configuration
    leads_per_run INT NOT NULL DEFAULT 10,
    concurrent_searches INT NOT NULL DEFAULT 1,
    compilation_target INT NOT NULL DEFAULT 100,
    system_prompt TEXT,
    search_keywords JSONB,
    ats_mode BOOLEAN NOT NULL DEFAULT true,
    time_filter VARCHAR(20) DEFAULT 'week',

    -- Target Entities (what to match against)
    target_type VARCHAR(50),  -- individual, organization, job, etc.
    target_ids JSONB,         -- array of entity IDs
    source_document_ids JSONB,

    -- Digest Configuration
    digest_enabled BOOLEAN NOT NULL DEFAULT true,
    digest_emails TEXT,
    last_digest_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, user_id, name)
);

-- Run log for tracking automation execution history
CREATE TABLE "Automation_Run_Log" (
    id BIGSERIAL PRIMARY KEY,
    automation_config_id BIGINT NOT NULL REFERENCES "Automation_Config"(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    leads_found INT NOT NULL DEFAULT 0,
    proposals_created INT NOT NULL DEFAULT 0,
    error_message TEXT,
    search_query TEXT
);

CREATE INDEX idx_automation_config_enabled ON "Automation_Config"(enabled, next_run_at);
CREATE INDEX idx_automation_config_user ON "Automation_Config"(tenant_id, user_id);
CREATE INDEX idx_automation_run_log_config ON "Automation_Run_Log"(automation_config_id, started_at DESC);

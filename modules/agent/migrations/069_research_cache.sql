-- Research cache for web/search results to avoid redundant API calls
-- Scoped per-tenant for team collaboration
CREATE TABLE IF NOT EXISTS "Research_Cache" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    cache_key VARCHAR(255) NOT NULL,
    cache_type VARCHAR(50) NOT NULL,
    query_params JSONB,
    result_data JSONB NOT NULL,
    result_count INT NOT NULL DEFAULT 0,
    created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(tenant_id, cache_key)
);

CREATE INDEX IF NOT EXISTS idx_research_cache_tenant_key ON "Research_Cache"(tenant_id, cache_key);
CREATE INDEX IF NOT EXISTS idx_research_cache_tenant_type ON "Research_Cache"(tenant_id, cache_type);
CREATE INDEX IF NOT EXISTS idx_research_cache_expires ON "Research_Cache"(expires_at);

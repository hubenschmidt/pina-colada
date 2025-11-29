-- Migration: Create saved_reports table for custom report definitions

CREATE TABLE IF NOT EXISTS "SavedReport" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query_definition JSONB NOT NULL,
    created_by BIGINT REFERENCES "Individual"(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_saved_reports_tenant ON "SavedReport"(tenant_id);
CREATE INDEX idx_saved_reports_created_by ON "SavedReport"(created_by);

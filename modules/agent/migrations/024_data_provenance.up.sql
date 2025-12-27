-- ==============================================
-- DATA PROVENANCE (Field-level tracking)
-- ==============================================
CREATE TABLE IF NOT EXISTS "Data_Provenance" (
    id              BIGSERIAL PRIMARY KEY,
    entity_type     TEXT NOT NULL,            -- 'Organization', 'Individual'
    entity_id       BIGINT NOT NULL,
    field_name      TEXT NOT NULL,            -- 'revenue_range_id', 'seniority_level', etc.
    source          TEXT NOT NULL,            -- 'clearbit', 'apollo', 'linkedin', 'agent', 'manual'
    source_url      TEXT,
    confidence      NUMERIC(3,2),             -- 0.00 to 1.00
    verified_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    verified_by     BIGINT REFERENCES "User"(id) ON DELETE SET NULL,  -- NULL = AI agent
    raw_value       JSONB,                    -- Original value from source
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT confidence_range CHECK (confidence >= 0 AND confidence <= 1)
);

CREATE INDEX IF NOT EXISTS idx_provenance_entity ON "Data_Provenance"(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_provenance_field ON "Data_Provenance"(entity_type, entity_id, field_name);
CREATE INDEX IF NOT EXISTS idx_provenance_stale ON "Data_Provenance"(verified_at);

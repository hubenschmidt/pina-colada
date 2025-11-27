-- Create Note table for polymorphic notes across entities
CREATE TABLE IF NOT EXISTS "Note" (
    id              SERIAL PRIMARY KEY,
    tenant_id       INTEGER NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    entity_type     VARCHAR(50) NOT NULL,
    entity_id       INTEGER NOT NULL,
    content         TEXT NOT NULL,
    created_by      INTEGER REFERENCES "User"(id) ON DELETE SET NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for efficient lookups by entity
CREATE INDEX IF NOT EXISTS idx_note_entity ON "Note" (tenant_id, entity_type, entity_id);

-- Index for user's notes
CREATE INDEX IF NOT EXISTS idx_note_created_by ON "Note" (created_by);

-- Add comment describing valid entity_type values
COMMENT ON COLUMN "Note".entity_type IS 'Valid values: individual, organization, contact, job, lead, opportunity, partnership';

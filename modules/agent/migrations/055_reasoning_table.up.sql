-- Create Reasoning table for CRM schema registry (RAG)
-- Maps reasoning contexts to database tables for AI agents

CREATE TABLE "Reasoning" (
    id BIGSERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    table_name TEXT NOT NULL,
    description TEXT,
    schema_hint JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE(type, table_name)
);

-- Index for faster lookups by type
CREATE INDEX idx_reasoning_type ON "Reasoning"(type);

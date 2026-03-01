CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE "Embedding" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    source_type VARCHAR(20) NOT NULL,
    source_id BIGINT NOT NULL,
    config_id BIGINT,
    chunk_index INT NOT NULL DEFAULT 0,
    chunk_text TEXT NOT NULL,
    embedding vector(3072) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_embedding_source ON "Embedding" (source_type, source_id);
CREATE INDEX idx_embedding_config ON "Embedding" (config_id);

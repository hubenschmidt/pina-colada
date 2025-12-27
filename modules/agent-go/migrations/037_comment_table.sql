-- Create Comment table for polymorphic comments across entities
CREATE TABLE IF NOT EXISTS "Comment" (
    id              SERIAL PRIMARY KEY,
    tenant_id       INTEGER NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    commentable_type VARCHAR(50) NOT NULL,
    commentable_id  INTEGER NOT NULL,
    content         TEXT NOT NULL,
    created_by      INTEGER REFERENCES "User"(id) ON DELETE SET NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for efficient lookups by entity
CREATE INDEX IF NOT EXISTS idx_comment_entity ON "Comment" (tenant_id, commentable_type, commentable_id);

-- Index for user's comments
CREATE INDEX IF NOT EXISTS idx_comment_created_by ON "Comment" (created_by);

-- Add comment describing valid commentable_type values
COMMENT ON COLUMN "Comment".commentable_type IS 'Valid values: Account, Lead, Deal, Task';

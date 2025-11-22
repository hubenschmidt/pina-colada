-- Refactor Tag to be polymorphic (can tag any entity)

-- Drop the Asset-specific join table
DROP TABLE IF EXISTS "AssetTag";

-- Create polymorphic EntityTag join table
CREATE TABLE IF NOT EXISTS "EntityTag" (
    tag_id BIGINT NOT NULL REFERENCES "Tag"(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (tag_id, entity_type, entity_id)
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_entity_tag_entity ON "EntityTag"(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_tag_tag_id ON "EntityTag"(tag_id);

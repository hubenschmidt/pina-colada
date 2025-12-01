-- Create Asset, Tag, and Asset_Tag tables
-- Migration 006: Polymorphic content storage for AI agent context

-- 1. Create Asset table
CREATE TABLE IF NOT EXISTS "Asset" (
    id BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    filename VARCHAR(255),
    checksum VARCHAR(64),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_Asset_entity ON "Asset"(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS ix_Asset_checksum ON "Asset"(checksum);

-- 2. Create Tag table
CREATE TABLE IF NOT EXISTS "Tag" (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS ix_Tag_name ON "Tag"(name);

-- 3. Create Asset_Tag join table
CREATE TABLE IF NOT EXISTS "Asset_Tag" (
    asset_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY (asset_id, tag_id),
    CONSTRAINT fk_asset_tag_asset FOREIGN KEY (asset_id)
        REFERENCES "Asset"(id) ON DELETE CASCADE,
    CONSTRAINT fk_asset_tag_tag FOREIGN KEY (tag_id)
        REFERENCES "Tag"(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS ix_Asset_Tag_tag_id ON "Asset_Tag"(tag_id);

-- Migration 041: Refactor Asset for joined table inheritance
-- Asset becomes base class, Document extends it for file storage

-- 1. Drop old Asset table and related structures
DROP TABLE IF EXISTS "Entity_Tag" CASCADE;  -- Will recreate
DROP TABLE IF EXISTS "Asset" CASCADE;

-- 2. Recreate Tag table (tenant-scoped, if not already)
-- Tag already exists from migration 006, just ensure it has tenant_id
ALTER TABLE "Tag" ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES "Tenant"(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_tag_tenant ON "Tag"(tenant_id);

-- 3. Recreate Entity_Tag (polymorphic tagging)
CREATE TABLE IF NOT EXISTS "Entity_Tag" (
    tag_id BIGINT NOT NULL REFERENCES "Tag"(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (tag_id, entity_type, entity_id)
);
CREATE INDEX IF NOT EXISTS idx_entity_tag_entity ON "Entity_Tag"(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_tag_tag_id ON "Entity_Tag"(tag_id);

-- 4. Create new Asset base table (joined table inheritance)
CREATE TABLE "Asset" (
    id BIGSERIAL PRIMARY KEY,
    asset_type TEXT NOT NULL,  -- discriminator: 'document', 'image', etc.
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,  -- MIME type
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_asset_tenant ON "Asset"(tenant_id);
CREATE INDEX idx_asset_user ON "Asset"(user_id);
CREATE INDEX idx_asset_type ON "Asset"(asset_type);

-- 5. Create Document extension table
CREATE TABLE "Document" (
    id BIGINT PRIMARY KEY REFERENCES "Asset"(id) ON DELETE CASCADE,
    storage_path TEXT NOT NULL,  -- path in storage backend
    file_size BIGINT NOT NULL
);

-- 6. Create Entity_Asset polymorphic junction table
-- Links any asset to any entity (Project, Lead, Account, etc.)
CREATE TABLE "Entity_Asset" (
    asset_id BIGINT NOT NULL REFERENCES "Asset"(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,  -- 'Project', 'Lead', 'Account', etc.
    entity_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (asset_id, entity_type, entity_id)
);

CREATE INDEX idx_entityasset_entity ON "Entity_Asset"(entity_type, entity_id);
CREATE INDEX idx_entityasset_asset ON "Entity_Asset"(asset_id);

-- 7. Add updated_at trigger for Asset
CREATE OR REPLACE FUNCTION update_asset_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER asset_updated_at_trigger
    BEFORE UPDATE ON "Asset"
    FOR EACH ROW
    EXECUTE FUNCTION update_asset_updated_at();

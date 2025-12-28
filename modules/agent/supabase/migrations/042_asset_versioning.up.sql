-- Migration 042: Add versioning support to Asset table
-- Enables document versioning with version number, parent reference, and current version flag

-- 1. Add versioning columns to Asset table
ALTER TABLE "Asset" ADD COLUMN IF NOT EXISTS parent_id BIGINT REFERENCES "Asset"(id) ON DELETE CASCADE;
ALTER TABLE "Asset" ADD COLUMN IF NOT EXISTS version_number INTEGER NOT NULL DEFAULT 1;
ALTER TABLE "Asset" ADD COLUMN IF NOT EXISTS is_current_version BOOLEAN NOT NULL DEFAULT TRUE;

-- 2. Create index for version queries
CREATE INDEX IF NOT EXISTS idx_asset_parent ON "Asset"(parent_id);
CREATE INDEX IF NOT EXISTS idx_asset_current_version ON "Asset"(is_current_version) WHERE is_current_version = TRUE;

-- 3. Add constraint: version_number must be positive
ALTER TABLE "Asset" ADD CONSTRAINT asset_version_number_positive CHECK (version_number > 0);

-- 4. Update existing documents: ensure they have correct default values
UPDATE "Asset" SET version_number = 1, is_current_version = TRUE, parent_id = NULL
WHERE version_number IS NULL OR is_current_version IS NULL;

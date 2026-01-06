DROP INDEX IF EXISTS idx_entity_asset_linked_by;
ALTER TABLE "Entity_Asset" DROP COLUMN IF EXISTS linked_by_user_id;

-- Add linked_by_user_id to track who linked a document to an entity
ALTER TABLE "Entity_Asset" ADD COLUMN IF NOT EXISTS linked_by_user_id INTEGER REFERENCES "User"(id);

-- Create index for querying recent links by user
CREATE INDEX IF NOT EXISTS idx_entity_asset_linked_by ON "Entity_Asset"(linked_by_user_id, created_at DESC);

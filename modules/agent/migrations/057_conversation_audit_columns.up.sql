-- Add audit columns to Conversation table for sharing
ALTER TABLE "Conversation"
ADD COLUMN created_by_id BIGINT REFERENCES "User"(id),
ADD COLUMN updated_by_id BIGINT REFERENCES "User"(id);

-- Backfill existing rows with user_id
UPDATE "Conversation" SET created_by_id = user_id, updated_by_id = user_id;

-- Make columns NOT NULL after backfill
ALTER TABLE "Conversation"
ALTER COLUMN created_by_id SET NOT NULL,
ALTER COLUMN updated_by_id SET NOT NULL;

-- Index for querying by creator
CREATE INDEX idx_conversation_created_by ON "Conversation"(created_by_id);

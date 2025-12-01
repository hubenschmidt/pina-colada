-- Add parent_comment_id for threaded/nested comments (Reddit-style single thread replies)
ALTER TABLE "Comment" ADD COLUMN IF NOT EXISTS parent_comment_id INTEGER REFERENCES "Comment"(id) ON DELETE CASCADE;

-- Index for efficient lookup of replies
CREATE INDEX IF NOT EXISTS idx_comment_parent ON "Comment" (parent_comment_id);

-- Comment explaining the threading model
COMMENT ON COLUMN "Comment".parent_comment_id IS 'Self-referencing FK for threaded replies. NULL = top-level comment, non-NULL = reply to parent comment';

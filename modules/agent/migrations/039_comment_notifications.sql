-- Comment Notification System
-- Tracks notifications for comment activity (direct replies and thread participation)

CREATE TABLE IF NOT EXISTS "CommentNotification" (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    comment_id INTEGER NOT NULL REFERENCES "Comment"(id) ON DELETE CASCADE,
    notification_type VARCHAR(20) NOT NULL, -- 'direct_reply' | 'thread_activity'
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT unique_user_comment_notification UNIQUE (user_id, comment_id)
);

-- Index for fast unread count queries
CREATE INDEX IF NOT EXISTS idx_notification_user_unread ON "CommentNotification" (user_id, is_read) WHERE is_read = FALSE;

-- Index for cleanup of old notifications
CREATE INDEX IF NOT EXISTS idx_notification_created ON "CommentNotification" (created_at);

-- Index for tenant filtering
CREATE INDEX IF NOT EXISTS idx_notification_tenant ON "CommentNotification" (tenant_id);

COMMENT ON TABLE "CommentNotification" IS 'Tracks comment notifications for users - direct replies and thread activity';
COMMENT ON COLUMN "CommentNotification".notification_type IS 'Type of notification: direct_reply (reply to your comment) or thread_activity (activity on entity you commented on)';

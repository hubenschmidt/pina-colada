-- Add updated_at column to Comment_Notification table
ALTER TABLE "Comment_Notification"
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

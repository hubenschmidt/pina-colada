-- Add selected_project_id to User table for persisting UI project selection
ALTER TABLE "User"
ADD COLUMN selected_project_id BIGINT REFERENCES "Project"(id) ON DELETE SET NULL;

-- Index for faster lookups
CREATE INDEX idx_user_selected_project ON "User"(selected_project_id);

-- ==============================================
-- INDIVIDUAL: New contact intelligence columns
-- ==============================================
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS twitter_url TEXT;
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS github_url TEXT;
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS bio TEXT;
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS seniority_level TEXT;     -- 'C-Level', 'VP', 'Director', 'Manager', 'IC'
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS department TEXT;          -- 'Engineering', 'Sales', 'Marketing', etc.
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS is_decision_maker BOOLEAN DEFAULT FALSE;
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS reports_to_id BIGINT REFERENCES "Individual"(id) ON DELETE SET NULL;

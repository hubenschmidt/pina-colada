-- ==============================================
-- ORGANIZATION: New firmographic columns
-- ==============================================
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS revenue_range_id BIGINT REFERENCES "Revenue_Range"(id) ON DELETE SET NULL;
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS founding_year INTEGER;
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS headquarters_city TEXT;
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS headquarters_state TEXT;
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS headquarters_country TEXT DEFAULT 'USA';
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS company_type TEXT;      -- 'private', 'public', 'nonprofit', 'government'
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS linkedin_url TEXT;
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS crunchbase_url TEXT;

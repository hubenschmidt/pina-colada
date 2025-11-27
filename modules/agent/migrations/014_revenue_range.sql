-- ==============================
-- Migration: RevenueRange Table
-- ==============================
-- Generic money range brackets for leads
-- category: 'salary' for Jobs, 'deal_value' for Opportunities, etc.

-- Create RevenueRange table
CREATE TABLE IF NOT EXISTS "RevenueRange" (
  id              BIGSERIAL PRIMARY KEY,
  category        TEXT NOT NULL,
  label           TEXT NOT NULL,
  min_value       INTEGER,
  max_value       INTEGER,
  display_order   INTEGER NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unique constraint on category + label
CREATE UNIQUE INDEX IF NOT EXISTS idx_revenue_range_category_label
ON "RevenueRange"(category, label);

-- Index for filtering by category
CREATE INDEX IF NOT EXISTS idx_revenue_range_category
ON "RevenueRange"(category);

-- Add trigger for updated_at
DROP TRIGGER IF EXISTS update_revenue_range_updated_at ON "RevenueRange";
CREATE TRIGGER update_revenue_range_updated_at
    BEFORE UPDATE ON "RevenueRange"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add revenue_range_id to Job table
ALTER TABLE "Job"
ADD COLUMN IF NOT EXISTS revenue_range_id BIGINT REFERENCES "RevenueRange"(id) ON DELETE SET NULL;

-- Seed salary brackets ($20k increments)
INSERT INTO "RevenueRange" (category, label, min_value, max_value, display_order) VALUES
('salary', '$0 - $20k', 0, 20000, 0),
('salary', '$20k - $40k', 20000, 40000, 1),
('salary', '$40k - $60k', 40000, 60000, 2),
('salary', '$60k - $80k', 60000, 80000, 3),
('salary', '$80k - $100k', 80000, 100000, 4),
('salary', '$100k - $120k', 100000, 120000, 5),
('salary', '$120k - $140k', 120000, 140000, 6),
('salary', '$140k - $160k', 140000, 160000, 7),
('salary', '$160k - $180k', 160000, 180000, 8),
('salary', '$180k - $200k', 180000, 200000, 9),
('salary', '$200k+', 200000, NULL, 10)
ON CONFLICT (category, label) DO NOTHING;

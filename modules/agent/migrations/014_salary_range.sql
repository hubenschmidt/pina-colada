-- ==============================
-- Migration: SalaryRange Table
-- ==============================
-- Salary range brackets for job listings

-- Create SalaryRange table
CREATE TABLE IF NOT EXISTS "SalaryRange" (
  id              BIGSERIAL PRIMARY KEY,
  label           TEXT NOT NULL UNIQUE,
  min_value       INTEGER,
  max_value       INTEGER,
  display_order   INTEGER NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add trigger for updated_at
DROP TRIGGER IF EXISTS update_salary_range_updated_at ON "SalaryRange";
CREATE TRIGGER update_salary_range_updated_at
    BEFORE UPDATE ON "SalaryRange"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add salary_range_id to Job table
ALTER TABLE "Job"
ADD COLUMN IF NOT EXISTS salary_range_id BIGINT REFERENCES "SalaryRange"(id) ON DELETE SET NULL;

-- Seed salary brackets ($20k increments, starting at $100k)
INSERT INTO "SalaryRange" (label, min_value, max_value, display_order) VALUES
('$100,000 - $120,000', 100000, 120000, 0),
('$120,000 - $140,000', 120000, 140000, 1),
('$140,000 - $160,000', 140000, 160000, 2),
('$160,000 - $180,000', 160000, 180000, 3),
('$180,000 - $200,000', 180000, 200000, 4),
('$200,000 - $220,000', 200000, 220000, 5),
('$220,000 - $240,000', 220000, 240000, 6),
('$240,000 - $260,000', 240000, 260000, 7),
('$260,000 - $280,000', 260000, 280000, 8),
('$280,000 - $300,000', 280000, 300000, 9),
('$300,000+', 300000, NULL, 10)
ON CONFLICT (label) DO NOTHING;

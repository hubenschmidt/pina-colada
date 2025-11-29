-- Create EmployeeCountRange lookup table
CREATE TABLE IF NOT EXISTS "EmployeeCountRange" (
  id              BIGSERIAL PRIMARY KEY,
  label           TEXT NOT NULL UNIQUE,
  min_value       INTEGER,
  max_value       INTEGER,
  display_order   INTEGER NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed employee count brackets
INSERT INTO "EmployeeCountRange" (label, min_value, max_value, display_order) VALUES
('1-10', 1, 10, 0),
('11-50', 11, 50, 1),
('51-500', 51, 500, 2),
('501-1500', 501, 1500, 3),
('1500+', 1500, NULL, 4)
ON CONFLICT (label) DO NOTHING;

-- Add FK column to Organization
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS employee_count_range_id BIGINT REFERENCES "EmployeeCountRange"(id) ON DELETE SET NULL;

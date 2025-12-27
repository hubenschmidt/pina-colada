-- Create Funding_Stage lookup table
CREATE TABLE IF NOT EXISTS "Funding_Stage" (
  id              BIGSERIAL PRIMARY KEY,
  label           TEXT NOT NULL UNIQUE,
  display_order   INTEGER NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed funding stages
INSERT INTO "Funding_Stage" (label, display_order) VALUES
('Pre-seed', 0),
('Seed', 1),
('Series A', 2),
('Series B', 3),
('Series C', 4),
('Pre-IPO', 5),
('IPO', 6)
ON CONFLICT (label) DO NOTHING;

-- Add FK column to Organization
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS funding_stage_id BIGINT REFERENCES "Funding_Stage"(id) ON DELETE SET NULL;

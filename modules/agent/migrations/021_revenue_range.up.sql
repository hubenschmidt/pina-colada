-- ==============================================
-- REVENUE RANGE (Lookup table)
-- ==============================================
CREATE TABLE IF NOT EXISTS "Revenue_Range" (
    id              BIGSERIAL PRIMARY KEY,
    label           TEXT NOT NULL UNIQUE,
    min_value       BIGINT,                   -- USD, NULL = unbounded
    max_value       BIGINT,                   -- USD, NULL = unbounded
    display_order   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Trigger for updated_at
CREATE TRIGGER update_revenue_range_updated_at
    BEFORE UPDATE ON "Revenue_Range"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Seed data
INSERT INTO "Revenue_Range" (label, min_value, max_value, display_order) VALUES
('< $1M', 0, 1000000, 0),
('$1M - $10M', 1000000, 10000000, 1),
('$10M - $50M', 10000000, 50000000, 2),
('$50M - $100M', 50000000, 100000000, 3),
('$100M - $500M', 100000000, 500000000, 4),
('$500M - $1B', 500000000, 1000000000, 5),
('$1B+', 1000000000, NULL, 6)
ON CONFLICT (label) DO NOTHING;

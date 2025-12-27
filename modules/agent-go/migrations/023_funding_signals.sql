-- ==============================================
-- FUNDING ROUND (Historical funding events)
-- ==============================================
CREATE TABLE IF NOT EXISTS "Funding_Round" (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
    round_type      TEXT NOT NULL,            -- 'Pre-Seed', 'Seed', 'Series A', 'Series B', etc.
    amount          BIGINT,                   -- USD cents
    announced_date  DATE,
    lead_investor   TEXT,
    source_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_funding_round_org ON "Funding_Round"(organization_id);
CREATE INDEX IF NOT EXISTS idx_funding_round_date ON "Funding_Round"(announced_date DESC);

-- ==============================================
-- COMPANY SIGNAL (Intent signals: hiring, news, etc.)
-- ==============================================
CREATE TABLE IF NOT EXISTS "Company_Signal" (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
    signal_type     TEXT NOT NULL,            -- 'hiring', 'expansion', 'product_launch', 'partnership', 'leadership_change', 'news'
    headline        TEXT NOT NULL,
    description     TEXT,
    signal_date     DATE,
    source          TEXT,                     -- 'linkedin', 'news', 'crunchbase', 'agent'
    source_url      TEXT,
    sentiment       TEXT,                     -- 'positive', 'neutral', 'negative'
    relevance_score NUMERIC(3,2),             -- 0.00 to 1.00
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_company_signal_org ON "Company_Signal"(organization_id);
CREATE INDEX IF NOT EXISTS idx_company_signal_date ON "Company_Signal"(signal_date DESC);
CREATE INDEX IF NOT EXISTS idx_company_signal_type ON "Company_Signal"(signal_type);

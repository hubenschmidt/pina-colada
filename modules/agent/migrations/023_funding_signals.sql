-- ==============================================
-- FUNDING ROUND (Historical funding events)
-- ==============================================
CREATE TABLE IF NOT EXISTS "FundingRound" (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
    round_type      TEXT NOT NULL,            -- 'Pre-Seed', 'Seed', 'Series A', 'Series B', etc.
    amount          BIGINT,                   -- USD cents
    announced_date  DATE,
    lead_investor   TEXT,
    source_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_funding_round_org ON "FundingRound"(organization_id);
CREATE INDEX IF NOT EXISTS idx_funding_round_date ON "FundingRound"(announced_date DESC);

-- ==============================================
-- COMPANY SIGNAL (Intent signals: hiring, news, etc.)
-- ==============================================
CREATE TABLE IF NOT EXISTS "CompanySignal" (
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

CREATE INDEX IF NOT EXISTS idx_company_signal_org ON "CompanySignal"(organization_id);
CREATE INDEX IF NOT EXISTS idx_company_signal_date ON "CompanySignal"(signal_date DESC);
CREATE INDEX IF NOT EXISTS idx_company_signal_type ON "CompanySignal"(signal_type);

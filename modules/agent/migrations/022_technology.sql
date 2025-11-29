-- ==============================================
-- TECHNOLOGY (Lookup table for tech stack)
-- ==============================================
CREATE TABLE IF NOT EXISTS "Technology" (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    category        TEXT NOT NULL,            -- 'CRM', 'Marketing Automation', 'Cloud', 'Database', etc.
    vendor          TEXT,                     -- 'Salesforce', 'HubSpot', 'AWS', etc.
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(name, category)
);

CREATE INDEX IF NOT EXISTS idx_technology_category ON "Technology"(category);

-- ==============================================
-- ORGANIZATION_TECHNOLOGY (Junction table)
-- ==============================================
CREATE TABLE IF NOT EXISTS "OrganizationTechnology" (
    organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
    technology_id   BIGINT NOT NULL REFERENCES "Technology"(id) ON DELETE CASCADE,
    detected_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    source          TEXT,                     -- 'builtwith', 'wappalyzer', 'agent', 'manual'
    confidence      NUMERIC(3,2),

    PRIMARY KEY (organization_id, technology_id)
);

CREATE INDEX IF NOT EXISTS idx_org_tech_org ON "OrganizationTechnology"(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_tech_tech ON "OrganizationTechnology"(technology_id);

-- Create Industry table and Account_Industry join table for many-to-many

CREATE TABLE IF NOT EXISTS "Industry" (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS "Account_Industry" (
    account_id BIGINT NOT NULL REFERENCES "Account"(id) ON DELETE CASCADE,
    industry_id BIGINT NOT NULL REFERENCES "Industry"(id) ON DELETE CASCADE,
    PRIMARY KEY (account_id, industry_id)
);

-- Index for fast lookups by industry
CREATE INDEX IF NOT EXISTS idx_account_industry_industry_id ON "Account_Industry"(industry_id);

-- Seed common industries
INSERT INTO "Industry" (name) VALUES
    ('Software'),
    ('Cloud Infrastructure'),
    ('Data Analytics'),
    ('Cybersecurity'),
    ('Consulting'),
    ('Venture Capital'),
    ('AI/ML'),
    ('FinTech'),
    ('HealthTech'),
    ('E-commerce')
ON CONFLICT (name) DO NOTHING;

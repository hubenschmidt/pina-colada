-- Create Account_Relationship table for linking accounts (individuals/organizations) to each other
-- This replaces the separate Individual_Relationship and Organization_Relationship tables

CREATE TABLE IF NOT EXISTS "Account_Relationship" (
    id BIGSERIAL PRIMARY KEY,
    from_account_id BIGINT NOT NULL REFERENCES "Account"(id) ON DELETE CASCADE,
    to_account_id BIGINT NOT NULL REFERENCES "Account"(id) ON DELETE CASCADE,
    relationship_type TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL REFERENCES "User"(id),
    updated_by BIGINT NOT NULL REFERENCES "User"(id),
    CONSTRAINT account_relationship_no_self CHECK (from_account_id != to_account_id),
    CONSTRAINT account_relationship_unique UNIQUE (from_account_id, to_account_id)
);

-- Index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_account_relationship_from ON "Account_Relationship"(from_account_id);
CREATE INDEX IF NOT EXISTS idx_account_relationship_to ON "Account_Relationship"(to_account_id);

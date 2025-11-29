-- Migration: Add Account-Project association for reporting purposes
-- Accounts can be associated with projects but this doesn't affect scope/filtering

CREATE TABLE IF NOT EXISTS "AccountProject" (
    account_id BIGINT NOT NULL REFERENCES "Account"(id) ON DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (account_id, project_id)
);

CREATE INDEX IF NOT EXISTS idx_account_project_account ON "AccountProject"(account_id);
CREATE INDEX IF NOT EXISTS idx_account_project_project ON "AccountProject"(project_id);

COMMENT ON TABLE "AccountProject" IS 'Junction table for Account-Project association. Used for reporting purposes only, does not affect scope/filtering.';

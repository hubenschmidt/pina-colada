-- Replace Contact_Individual and Contact_Organization with Contact_Account
-- Contacts now link directly to Account (parent of Individual/Organization)

CREATE TABLE IF NOT EXISTS "Contact_Account" (
    id BIGSERIAL PRIMARY KEY,
    contact_id BIGINT NOT NULL REFERENCES "Contact"(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES "Account"(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT contact_account_unique UNIQUE (contact_id, account_id)
);

CREATE INDEX idx_contact_account_contact_id ON "Contact_Account"(contact_id);
CREATE INDEX idx_contact_account_account_id ON "Contact_Account"(account_id);

-- Drop old junction tables
DROP TABLE IF EXISTS "Contact_Individual";
DROP TABLE IF EXISTS "Contact_Organization";

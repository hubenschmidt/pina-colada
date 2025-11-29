-- Migration: Convert Contact to many-to-many relationships
-- Contacts can now link to multiple Individuals AND multiple Organizations

-- Create junction table for Contact-Individual relationships
CREATE TABLE IF NOT EXISTS "ContactIndividual" (
    id BIGSERIAL PRIMARY KEY,
    contact_id BIGINT NOT NULL REFERENCES "Contact"(id) ON DELETE CASCADE,
    individual_id BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    UNIQUE(contact_id, individual_id)
);

-- Create junction table for Contact-Organization relationships
CREATE TABLE IF NOT EXISTS "ContactOrganization" (
    id BIGSERIAL PRIMARY KEY,
    contact_id BIGINT NOT NULL REFERENCES "Contact"(id) ON DELETE CASCADE,
    organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    UNIQUE(contact_id, organization_id)
);

-- Migrate existing individual_id relationships to junction table
INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at)
SELECT id, individual_id, created_at
FROM "Contact"
WHERE individual_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- Migrate existing organization_id relationships to junction table
INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at)
SELECT id, organization_id, is_primary, created_at
FROM "Contact"
WHERE organization_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- Drop the old foreign key columns from Contact
ALTER TABLE "Contact" DROP COLUMN IF EXISTS individual_id;
ALTER TABLE "Contact" DROP COLUMN IF EXISTS organization_id;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_contact_individual_contact_id ON "ContactIndividual"(contact_id);
CREATE INDEX IF NOT EXISTS idx_contact_individual_individual_id ON "ContactIndividual"(individual_id);
CREATE INDEX IF NOT EXISTS idx_contact_organization_contact_id ON "ContactOrganization"(contact_id);
CREATE INDEX IF NOT EXISTS idx_contact_organization_organization_id ON "ContactOrganization"(organization_id);

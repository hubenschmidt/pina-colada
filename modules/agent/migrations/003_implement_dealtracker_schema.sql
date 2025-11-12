-- Migration: Implement DealTracker CRM Schema
-- Implements the schema described in dealtracker_schema.md
-- Uses UUIDs to match existing schema pattern (instead of BIGSERIAL)
-- Fixes constraint function to be DEFERRABLE for proper validation
--
-- Note: This creates NEW CRM tables (Deal, Lead, etc.) for the DealTracker system.
-- The "Lead" table from migration 002 contains existing lead data.
-- The new "Lead" table here is part of the CRM schema and will conflict with migration 002's "Lead".
-- You may need to rename one or migrate data before applying both migrations.

-- ==============================
-- 1) Core Tables
-- ==============================

-- Deal table
CREATE TABLE IF NOT EXISTS "Deal" (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name            TEXT NOT NULL,
  description     TEXT,
  owner_user_id   UUID,
  value_amount    NUMERIC(18,2),
  value_currency  TEXT DEFAULT 'USD',
  probability     NUMERIC(5,2),            -- 0..100 (pct)
  close_date      DATE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Lead table (new CRM schema - note: "Lead" from migration 002 contains existing lead data)
CREATE TABLE IF NOT EXISTS "Lead" (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  deal_id         UUID NOT NULL REFERENCES "Deal"(id) ON DELETE CASCADE,
  title           TEXT NOT NULL,           -- short label for the lead
  description     TEXT,
  source          TEXT,                    -- inbound/referral/event/campaign/etc.
  owner_user_id   UUID,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Lead_Status: strict one-to-one with Lead
-- Use lead_id as both PK and FK to enforce 1:1
-- Note: This replaces the old Lead_Status qualification table from migration 002
CREATE TABLE IF NOT EXISTS "Lead_Status" (
  lead_id         UUID PRIMARY KEY REFERENCES "Lead"(id) ON DELETE CASCADE,
  stage           TEXT NOT NULL,           -- e.g. New, Qualified, Proposal, Won, Lost
  reason          TEXT,                    -- optional; e.g. lost reason
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Handle old Lead_Status qualification table if it exists (from old migration 002)
-- The old table had id/name/description for qualification statuses
-- The new Lead_Status table uses lead_id/stage/reason for CRM stages
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'Lead_Status' 
        AND table_schema = current_schema()
        AND EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'Lead_Status' 
            AND column_name = 'id'
        )
    ) THEN
        -- Old qualification table exists (has 'id' column), rename it
        ALTER TABLE "Lead_Status" RENAME TO "Lead_Qualification_Status";
    END IF;
END $$;

-- Create Lead_Qualification_Status table if needed (for backward compatibility)
-- This stores qualification statuses: Qualifying, Cold, Warm, Hot
CREATE TABLE IF NOT EXISTS "Lead_Qualification_Status" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert initial qualification statuses
INSERT INTO "Lead_Qualification_Status" (name, description) VALUES
    ('Qualifying', 'Initial lead from agent - needs qualification'),
    ('Cold', 'Cold lead - minimal prior engagement'),
    ('Warm', 'Warm lead - some prior engagement or interest'),
    ('Hot', 'Hot lead - high interest and engagement')
ON CONFLICT (name) DO NOTHING;

-- Add lead_status_id column to Lead table if it doesn't exist (for qualification status)
-- This allows the old Lead table to reference qualification statuses
ALTER TABLE "Lead" ADD COLUMN IF NOT EXISTS lead_status_id UUID REFERENCES "Lead_Qualification_Status"(id);

-- Create index on lead_status_id for efficient filtering
CREATE INDEX IF NOT EXISTS idx_lead_lead_status_id ON "Lead"(lead_status_id) WHERE lead_status_id IS NOT NULL;

-- Organization table
CREATE TABLE IF NOT EXISTS "Organization" (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name            TEXT NOT NULL,
  website         TEXT,
  phone           TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (LOWER(name))
);

-- Individual table
CREATE TABLE IF NOT EXISTS "Individual" (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  given_name      TEXT NOT NULL,
  family_name     TEXT NOT NULL,
  email           TEXT,
  phone           TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Contact: must be an Individual; Organization is optional
CREATE TABLE IF NOT EXISTS "Contact" (
  id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  individual_id     UUID NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  organization_id   UUID REFERENCES "Organization"(id) ON DELETE SET NULL,
  title             TEXT,
  department        TEXT,
  email             TEXT,
  phone             TEXT,
  is_primary        BOOLEAN NOT NULL DEFAULT FALSE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ==============================
-- 2) Junction Tables (many-to-manys)
-- ==============================

-- Lead ↔ Contacts
CREATE TABLE IF NOT EXISTS "Lead_Contacts" (
  lead_id       UUID NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  contact_id    UUID NOT NULL REFERENCES "Contact"(id) ON DELETE CASCADE,
  role_on_lead  TEXT,                      -- Decision Maker, Influencer, Billing, etc.
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, contact_id)
);

-- Lead ↔ Organizations
CREATE TABLE IF NOT EXISTS "Lead_Organizations" (
  lead_id         UUID NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  relationship    TEXT,                    -- Prospect, Customer, Partner, Competitor
  is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, organization_id)
);

-- Lead ↔ Individuals (direct)
CREATE TABLE IF NOT EXISTS "Lead_Individuals" (
  lead_id       UUID NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  individual_id UUID NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  relationship  TEXT,                      -- Champion, Sponsor, Evaluator, etc.
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, individual_id)
);

-- ==============================
-- 3) Constraint: Lead must have ≥1 Individual or Organization
-- ==============================
-- FIXED: Made DEFERRABLE so junction table inserts can happen before validation
-- The constraint is checked at transaction end, allowing leads to be created
-- and then parties added in the same transaction

CREATE OR REPLACE FUNCTION enforce_lead_party_presence() RETURNS TRIGGER AS $$
DECLARE
  target_lead_id UUID;
  has_party BOOLEAN;
BEGIN
  -- Determine which lead_id to check based on the table and operation
  IF TG_TABLE_NAME = 'Lead' THEN
    -- For Lead table, check the lead being inserted/updated
    target_lead_id := NEW.id;
  ELSIF TG_TABLE_NAME = 'Lead_Individuals' THEN
    -- For junction tables, check the lead_id from NEW (insert) or OLD (delete)
    target_lead_id := COALESCE(NEW.lead_id, OLD.lead_id);
  ELSIF TG_TABLE_NAME = 'Lead_Organizations' THEN
    target_lead_id := COALESCE(NEW.lead_id, OLD.lead_id);
  ELSE
    RETURN COALESCE(NEW, OLD); -- Unknown table, skip check
  END IF;

  -- Check if lead has at least one party (checking current state in database)
  SELECT EXISTS (
    SELECT 1 FROM "Lead_Individuals" WHERE lead_id = target_lead_id
    UNION ALL
    SELECT 1 FROM "Lead_Organizations" WHERE lead_id = target_lead_id
  ) INTO has_party;

  -- Enforce constraint: lead must have at least one party
  IF NOT has_party THEN
    RAISE EXCEPTION 'Lead % must have at least one associated Individual or Organization', target_lead_id;
  END IF;

  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers (DEFERRABLE INITIALLY DEFERRED for proper validation)
DROP TRIGGER IF EXISTS trg_enforce_lead_party_on_leads ON "Lead";
CREATE CONSTRAINT TRIGGER trg_enforce_lead_party_on_leads
AFTER INSERT OR UPDATE ON "Lead"
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION enforce_lead_party_presence();

DROP TRIGGER IF EXISTS trg_enforce_lead_party_on_lead_individuals ON "Lead_Individuals";
CREATE CONSTRAINT TRIGGER trg_enforce_lead_party_on_lead_individuals
AFTER INSERT OR DELETE ON "Lead_Individuals"
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION enforce_lead_party_presence();

DROP TRIGGER IF EXISTS trg_enforce_lead_party_on_lead_organizations ON "Lead_Organizations";
CREATE CONSTRAINT TRIGGER trg_enforce_lead_party_on_lead_organizations
AFTER INSERT OR DELETE ON "Lead_Organizations"
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION enforce_lead_party_presence();

-- ==============================
-- 4) Helpful Indexes
-- ==============================
CREATE INDEX IF NOT EXISTS idx_leads_deal_id ON "Lead"(deal_id);
CREATE INDEX IF NOT EXISTS idx_lead_status_stage ON "Lead_Status"(stage);
CREATE INDEX IF NOT EXISTS idx_contacts_individual ON "Contact"(individual_id);
CREATE INDEX IF NOT EXISTS idx_contacts_org ON "Contact"(organization_id);
CREATE INDEX IF NOT EXISTS idx_lead_contacts_lead ON "Lead_Contacts"(lead_id);
CREATE INDEX IF NOT EXISTS idx_lead_orgs_lead ON "Lead_Organizations"(lead_id);
CREATE INDEX IF NOT EXISTS idx_lead_inds_lead ON "Lead_Individuals"(lead_id);

-- ==============================
-- 5) Auto-update triggers for updated_at
-- ==============================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers to all tables with updated_at column
DROP TRIGGER IF EXISTS update_deals_updated_at ON "Deal";
CREATE TRIGGER update_deals_updated_at
    BEFORE UPDATE ON "Deal"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_leads_updated_at ON "Lead";
CREATE TRIGGER update_leads_updated_at
    BEFORE UPDATE ON "Lead"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_lead_status_updated_at ON "Lead_Status";
CREATE TRIGGER update_lead_status_updated_at
    BEFORE UPDATE ON "Lead_Status"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organizations_updated_at ON "Organization";
CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON "Organization"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_individuals_updated_at ON "Individual";
CREATE TRIGGER update_individuals_updated_at
    BEFORE UPDATE ON "Individual"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_contacts_updated_at ON "Contact";
CREATE TRIGGER update_contacts_updated_at
    BEFORE UPDATE ON "Contact"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ==============================
-- 6) Picklist table for seed data
-- ==============================
CREATE TABLE IF NOT EXISTS "Picklist" (
  id BIGSERIAL PRIMARY KEY,
  category TEXT NOT NULL,   -- e.g., 'lead_contacts.role_on_lead', 'lead_individuals.relationship', 'lead_organizations.relationship', 'lead_status.stage'
  value TEXT NOT NULL,
  UNIQUE (category, value)
);

-- Insert seed data for picklists
INSERT INTO "Picklist" (category, value) VALUES
-- lead_status.stage
('lead_status.stage','New'),
('lead_status.stage','Qualified'),
('lead_status.stage','Discovery'),
('lead_status.stage','Proposal'),
('lead_status.stage','Negotiation'),
('lead_status.stage','Won'),
('lead_status.stage','Lost'),
('lead_status.stage','On Hold'),

-- lead_contacts.role_on_lead
('lead_contacts.role_on_lead','Decision Maker'),
('lead_contacts.role_on_lead','Economic Buyer'),
('lead_contacts.role_on_lead','Champion'),
('lead_contacts.role_on_lead','User'),
('lead_contacts.role_on_lead','Influencer'),
('lead_contacts.role_on_lead','Procurement'),
('lead_contacts.role_on_lead','Legal'),
('lead_contacts.role_on_lead','Billing'),

-- lead_organizations.relationship
('lead_organizations.relationship','Prospect'),
('lead_organizations.relationship','Customer'),
('lead_organizations.relationship','Partner'),
('lead_organizations.relationship','Competitor'),
('lead_organizations.relationship','Reseller'),

-- lead_individuals.relationship
('lead_individuals.relationship','Champion'),
('lead_individuals.relationship','Sponsor'),
('lead_individuals.relationship','Evaluator'),
('lead_individuals.relationship','Gatekeeper'),
('lead_individuals.relationship','Blocker')
ON CONFLICT DO NOTHING;


# DealTracker Data Model & Seeds
_Last generated: 2025-11-12T02:08:25_

This document describes a CRM-style **DealTracker** schema (renaming `Job` → `Deal` and `LeadStatus` → `Lead_Status`) and includes helpful seed data for stages and roles.

---


## Entity-Relationship Summary

- **Deal** (renamed from `Job`)
  - Has many **Leads**

- **Lead**
  - **One-to-one** with **Lead_Status**
  - Many **Contacts** (each Contact **must** be an **Individual**)
  - Many **Organizations**
  - Many **Individuals** (direct link, not just through Contact)
  - **Constraint:** A Lead must have **at least one** associated **Individual or Organization**

- **Organization**
  - Has many **Contacts**

- **Individual**
  - Represents a person
  - Contacts model an individual's role/context (often tied to an Organization)

- **Contact**
  - Must reference an **Individual** (required)
  - May reference an **Organization** (optional)
  - Represents the **person-in-a-role** (title/department, role-specific email/phone)

---

## Postgres DDL

```sql
-- ==============================
-- 1) Renames
-- ==============================
ALTER TABLE job RENAME TO deal;
-- If your old table was LeadStatus (or similar), rename to Lead_Status
-- (Use quotes if your identifier was mixed case)
ALTER TABLE leadstatus RENAME TO lead_status;

-- ==============================
-- 2) Core Tables
-- ==============================
CREATE TABLE deals (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  description     TEXT,
  owner_user_id   BIGINT,
  value_amount    NUMERIC(18,2),
  value_currency  TEXT DEFAULT 'USD',
  probability     NUMERIC(5,2),            -- 0..100 (pct)
  close_date      DATE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE leads (
  id              BIGSERIAL PRIMARY KEY,
  deal_id         BIGINT NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
  title           TEXT NOT NULL,           -- short label for the lead
  description     TEXT,
  source          TEXT,                    -- inbound/referral/event/campaign/etc.
  owner_user_id   BIGINT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Lead_Status: strict one-to-one with Lead
-- Use lead_id as both PK and FK to enforce 1:1
CREATE TABLE lead_status (
  lead_id         BIGINT PRIMARY KEY REFERENCES leads(id) ON DELETE CASCADE,
  stage           TEXT NOT NULL,           -- e.g. New, Qualified, Proposal, Won, Lost
  reason          TEXT,                    -- optional; e.g. lost reason
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organizations (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  website         TEXT,
  phone           TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (LOWER(name))
);

CREATE TABLE individuals (
  id              BIGSERIAL PRIMARY KEY,
  given_name      TEXT NOT NULL,
  family_name     TEXT NOT NULL,
  email           TEXT,
  phone           TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Contact: must be an Individual; Organization is optional
CREATE TABLE contacts (
  id                BIGSERIAL PRIMARY KEY,
  individual_id     BIGINT NOT NULL REFERENCES individuals(id) ON DELETE CASCADE,
  organization_id   BIGINT REFERENCES organizations(id) ON DELETE SET NULL,
  title             TEXT,
  department        TEXT,
  email             TEXT,
  phone             TEXT,
  is_primary        BOOLEAN NOT NULL DEFAULT FALSE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ==============================
-- 3) Junction Tables (many-to-manys)
-- ==============================

-- Lead ↔ Contacts
CREATE TABLE lead_contacts (
  lead_id       BIGINT NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
  contact_id    BIGINT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
  role_on_lead  TEXT,                      -- Decision Maker, Influencer, Billing, etc.
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, contact_id)
);

-- Lead ↔ Organizations
CREATE TABLE lead_organizations (
  lead_id         BIGINT NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  relationship    TEXT,                    -- Prospect, Customer, Partner, Competitor
  is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, organization_id)
);

-- Lead ↔ Individuals (direct)
CREATE TABLE lead_individuals (
  lead_id       BIGINT NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
  individual_id BIGINT NOT NULL REFERENCES individuals(id) ON DELETE CASCADE,
  relationship  TEXT,                      -- Champion, Sponsor, Evaluator, etc.
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, individual_id)
);

-- ==============================
-- 4) Constraint: Lead must have ≥1 Individual or Organization
-- ==============================

CREATE OR REPLACE FUNCTION enforce_lead_party_presence() RETURNS TRIGGER AS $$
DECLARE
  has_party BOOLEAN;
BEGIN
  IF TG_TABLE_NAME = 'leads' THEN
    SELECT EXISTS (
      SELECT 1 FROM lead_individuals WHERE lead_id = NEW.id
      UNION ALL
      SELECT 1 FROM lead_organizations WHERE lead_id = NEW.id
    ) INTO has_party;

    IF NOT has_party THEN
      RAISE EXCEPTION 'Lead % must have at least one associated Individual or Organization', NEW.id;
    END IF;

  ELSE
    -- For changes on junctions, check their lead
    SELECT EXISTS (
      SELECT 1 FROM lead_individuals WHERE lead_id = NEW.lead_id
      UNION ALL
      SELECT 1 FROM lead_organizations WHERE lead_id = NEW.lead_id
    ) INTO has_party;

    IF NOT has_party THEN
      RAISE EXCEPTION 'Lead % must have at least one associated Individual or Organization', NEW.lead_id;
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_enforce_lead_party_on_leads
AFTER INSERT OR UPDATE ON leads
FOR EACH ROW EXECUTE FUNCTION enforce_lead_party_presence();

CREATE TRIGGER trg_enforce_lead_party_on_lead_individuals
AFTER INSERT OR DELETE ON lead_individuals
FOR EACH ROW EXECUTE FUNCTION enforce_lead_party_presence();

CREATE TRIGGER trg_enforce_lead_party_on_lead_organizations
AFTER INSERT OR DELETE ON lead_organizations
FOR EACH ROW EXECUTE FUNCTION enforce_lead_party_presence();

-- ==============================
-- 5) Helpful Indexes
-- ==============================
CREATE INDEX idx_leads_deal_id ON leads(deal_id);
CREATE INDEX idx_lead_status_stage ON lead_status(stage);
CREATE INDEX idx_contacts_individual ON contacts(individual_id);
CREATE INDEX idx_contacts_org ON contacts(organization_id);
CREATE INDEX idx_lead_contacts_lead ON lead_contacts(lead_id);
CREATE INDEX idx_lead_orgs_lead ON lead_organizations(lead_id);
CREATE INDEX idx_lead_inds_lead ON lead_individuals(lead_id);
```

---

## Optional: Rails / ActiveRecord Mappings

```ruby
# deal.rb
class Deal < ApplicationRecord
  has_many :leads, dependent: :destroy
end

# lead.rb
class Lead < ApplicationRecord
  belongs_to :deal
  has_one  :lead_status, class_name: 'LeadStatus', foreign_key: :lead_id,
          dependent: :destroy, inverse_of: :lead

  has_many :lead_contacts, dependent: :destroy
  has_many :contacts, through: :lead_contacts

  has_many :lead_organizations, dependent: :destroy
  has_many :organizations, through: :lead_organizations

  has_many :lead_individuals, dependent: :destroy
  has_many :individuals, through: :lead_individuals

  validate :must_have_individual_or_organization

  private
  def must_have_individual_or_organization
    if organizations.empty? && individuals.empty?
      errors.add(:base, 'Lead must have at least one Individual or Organization')
    end
  end
end

# lead_status.rb
class LeadStatus < ApplicationRecord
  self.table_name = 'lead_status'
  belongs_to :lead, primary_key: :id, foreign_key: :lead_id, inverse_of: :lead_status
  validates :stage, presence: true
end

# organization.rb
class Organization < ApplicationRecord
  has_many :contacts, dependent: :nullify
  has_many :lead_organizations, dependent: :destroy
  has_many :leads, through: :lead_organizations
end

# individual.rb
class Individual < ApplicationRecord
  has_many :contacts, dependent: :destroy
  has_many :lead_individuals, dependent: :destroy
  has_many :leads, through: :lead_individuals
end

# contact.rb
class Contact < ApplicationRecord
  belongs_to :individual
  belongs_to :organization, optional: true

  has_many :lead_contacts, dependent: :destroy
  has_many :leads, through: :lead_contacts
end

# Junctions
class LeadContact < ApplicationRecord
  belongs_to :lead
  belongs_to :contact
end

class LeadOrganization < ApplicationRecord
  belongs_to :lead
  belongs_to :organization
end

class LeadIndividual < ApplicationRecord
  belongs_to :lead
  belongs_to :individual
end
```

> **Naming note:** You requested model `Lead_Status`. The DDL uses a table named `lead_status` with a strict 1:1 keyed by `lead_id`. In Rails, class `LeadStatus` maps to `lead_statuses` by default; keeping the table as `lead_status` via `self.table_name` (shown above) maintains your desired name. Choose the convention you prefer and apply consistently.

---

## Seed Roles & Stages

### SQL Seeds

```sql
-- Lead Stages
-- (Attach stages to existing leads later in your pipeline)
-- Example pattern:
-- INSERT INTO lead_status (lead_id, stage, reason) VALUES ({{lead_id}}, 'Discovery', NULL);

-- Enumerations (optional picklist approach)
CREATE TABLE IF NOT EXISTS picklists (
  id BIGSERIAL PRIMARY KEY,
  category TEXT NOT NULL,   -- e.g., 'lead_contacts.role_on_lead', 'lead_individuals.relationship', 'lead_organizations.relationship', 'lead_status.stage'
  value TEXT NOT NULL,
  UNIQUE (category, value)
);

INSERT INTO picklists (category, value) VALUES
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
```

### Rails Seeds (optional)

```ruby
# db/seeds.rb (append)

# Picklists
picklists = {
  'lead_status.stage' => %w[New Qualified Discovery Proposal Negotiation Won Lost On\ Hold],
  'lead_contacts.role_on_lead' => ['Decision Maker','Economic Buyer','Champion','User','Influencer','Procurement','Legal','Billing'],
  'lead_organizations.relationship' => ['Prospect','Customer','Partner','Competitor','Reseller'],
  'lead_individuals.relationship' => ['Champion','Sponsor','Evaluator','Gatekeeper','Blocker']
}

picklists.each do |category, values|
  values.each do |val|
    Picklist.find_or_create_by!(category: category, value: val)
  end
end

# Example creation flow for a Deal → Lead with parties and status
deal = Deal.create!(name: 'ACME Expansion', value_amount: 250000, value_currency: 'USD', probability: 35)

lead = deal.leads.create!(title: 'ACME Expansion FY25', source: 'Referral')

org = Organization.find_or_create_by!(name: 'ACME Corp', website: 'https://acme.example')
person = Individual.find_or_create_by!(given_name: 'Avery', family_name: 'Nguyen', email: 'avery.nguyen@acme.example')

contact = Contact.create!(individual: person, organization: org, title: 'Director of Ops', department: 'Operations', is_primary: true)

lead.lead_status = LeadStatus.create!(stage: 'Discovery')

lead.lead_organizations.create!(organization: org, relationship: 'Prospect', is_primary: true)
lead.lead_individuals.create!(individual: person, relationship: 'Champion', is_primary: true)
lead.lead_contacts.create!(contact: contact, role_on_lead: 'Decision Maker', is_primary: true)
```

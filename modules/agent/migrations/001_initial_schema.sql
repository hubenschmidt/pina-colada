-- ==============================
-- DealTracker Schema Migration
-- ==============================
-- Creates the complete DealTracker schema with:
-- - Status, Deal, Lead, Organization, Individual, Contact
-- - Job (using Joined Table Inheritance)
-- - Opportunity, Partnership
-- - Activity tracking and relationship management
--
-- BACKUP YOUR DATABASE BEFORE RUNNING THIS MIGRATION!
-- ==============================

-- ==============================
-- STEP 1: Create Helper Functions
-- ==============================

-- Create function for auto-updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ==============================
-- STEP 2: Create Core Tables
-- ==============================

-- Status table (central status/stage definitions)
CREATE TABLE IF NOT EXISTS "Status" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL UNIQUE,
  description     TEXT,
  category        TEXT,  -- 'job', 'lead', 'deal', 'task_status', 'task_priority'
  is_terminal     BOOLEAN DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Project table
CREATE TABLE IF NOT EXISTS "Project" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  description     TEXT,
  owner_user_id   BIGINT,
  status          TEXT,  -- Simple TEXT field, not using Status table
  start_date      DATE,
  end_date        DATE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Deal table
CREATE TABLE IF NOT EXISTS "Deal" (
  id                  BIGSERIAL PRIMARY KEY,
  name                TEXT NOT NULL,
  description         TEXT,
  owner_user_id       BIGINT,
  current_status_id   BIGINT REFERENCES "Status"(id) ON DELETE SET NULL,
  value_amount        NUMERIC(18,2),
  value_currency      TEXT DEFAULT 'USD',
  probability         NUMERIC(5,2),
  expected_close_date DATE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Organization table (for companies)
CREATE TABLE IF NOT EXISTS "Organization" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  website         TEXT,
  phone           TEXT,
  industry        TEXT,
  employee_count  INTEGER,
  description     TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create unique index on lowercase name
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_name_lower ON "Organization"(LOWER(name));

-- Individual table (for people)
CREATE TABLE IF NOT EXISTS "Individual" (
  id              BIGSERIAL PRIMARY KEY,
  first_name      TEXT NOT NULL,
  last_name       TEXT NOT NULL,
  email           TEXT,
  phone           TEXT,
  linkedin_url    TEXT,
  title           TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create unique index on lowercase email
CREATE UNIQUE INDEX IF NOT EXISTS idx_individual_email_lower ON "Individual"(LOWER(email));

-- Contact table
CREATE TABLE IF NOT EXISTS "Contact" (
  id                BIGSERIAL PRIMARY KEY,
  individual_id     BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  organization_id   BIGINT REFERENCES "Organization"(id) ON DELETE SET NULL,
  title             TEXT,
  department        TEXT,
  role              TEXT,
  email             TEXT,
  phone             TEXT,
  is_primary        BOOLEAN NOT NULL DEFAULT FALSE,
  notes             TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Lead table (base table for Joined Table Inheritance)
CREATE TABLE IF NOT EXISTS "Lead" (
  id              BIGSERIAL PRIMARY KEY,
  deal_id         BIGINT NOT NULL REFERENCES "Deal"(id) ON DELETE CASCADE,
  type            TEXT NOT NULL,  -- Discriminator: 'Job', 'Opportunity', 'Partnership', etc.
  title           TEXT NOT NULL,
  description     TEXT,
  source          TEXT,  -- inbound/referral/event/campaign/agent/manual/etc.
  current_status_id BIGINT REFERENCES "Status"(id) ON DELETE SET NULL,
  owner_user_id   BIGINT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Task table (polymorphic association)
CREATE TABLE IF NOT EXISTS "Task" (
  id                  BIGSERIAL PRIMARY KEY,
  taskable_type       TEXT,  -- 'Deal', 'Lead', 'Job', 'Project', 'Organization', 'Individual', 'Contact'
  taskable_id         BIGINT,
  title               TEXT NOT NULL,
  description         TEXT,
  current_status_id   BIGINT REFERENCES "Status"(id) ON DELETE SET NULL,
  priority_id         BIGINT REFERENCES "Status"(id) ON DELETE SET NULL,
  due_date            DATE,
  completed_at        TIMESTAMPTZ,
  assigned_to_user_id BIGINT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ==============================
-- STEP 3: Status Configuration Tables
-- ==============================

CREATE TABLE IF NOT EXISTS "Job_Status" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,
  is_default      BOOLEAN DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

CREATE TABLE IF NOT EXISTS "Lead_Status" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,
  is_default      BOOLEAN DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

CREATE TABLE IF NOT EXISTS "Deal_Status" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,
  is_default      BOOLEAN DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

CREATE TABLE IF NOT EXISTS "Task_Status" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,
  is_default      BOOLEAN DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

CREATE TABLE IF NOT EXISTS "Task_Priority" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,
  is_default      BOOLEAN DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

-- ==============================
-- STEP 4: Junction Tables
-- ==============================

CREATE TABLE IF NOT EXISTS "Lead_Contact" (
  lead_id       BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  contact_id    BIGINT NOT NULL REFERENCES "Contact"(id) ON DELETE CASCADE,
  role_on_lead  TEXT,
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, contact_id)
);

CREATE TABLE IF NOT EXISTS "Lead_Organization" (
  lead_id         BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  relationship    TEXT,
  is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, organization_id)
);

CREATE TABLE IF NOT EXISTS "Lead_Individual" (
  lead_id       BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  individual_id BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  relationship  TEXT,
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, individual_id)
);

CREATE TABLE IF NOT EXISTS "Project_Deal" (
  project_id    BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  deal_id       BIGINT NOT NULL REFERENCES "Deal"(id) ON DELETE CASCADE,
  relationship  TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, deal_id)
);

CREATE TABLE IF NOT EXISTS "Project_Lead" (
  project_id    BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  lead_id       BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  relationship  TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, lead_id)
);

CREATE TABLE IF NOT EXISTS "Project_Contact" (
  project_id    BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  contact_id    BIGINT NOT NULL REFERENCES "Contact"(id) ON DELETE CASCADE,
  role          TEXT,
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, contact_id)
);

CREATE TABLE IF NOT EXISTS "Project_Organization" (
  project_id      BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  relationship    TEXT,
  is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, organization_id)
);

CREATE TABLE IF NOT EXISTS "Project_Individual" (
  project_id    BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  individual_id BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  role          TEXT,
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, individual_id)
);

-- ==============================
-- STEP 5: Lead Subtypes (Joined Table Inheritance)
-- ==============================

-- Drop old Job table if it exists (production cleanup)
DROP TABLE IF EXISTS "Job" CASCADE;

-- Job table (extends Lead)
CREATE TABLE "Job" (
  id              BIGINT PRIMARY KEY REFERENCES "Lead"(id) ON DELETE CASCADE,
  organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  job_title       TEXT NOT NULL,
  job_url         TEXT,
  notes           TEXT,
  resume_date     TIMESTAMPTZ,
  salary_range    TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Opportunity table (extends Lead)
CREATE TABLE IF NOT EXISTS "Opportunity" (
  id                    BIGINT PRIMARY KEY REFERENCES "Lead"(id) ON DELETE CASCADE,
  organization_id       BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  opportunity_name      TEXT NOT NULL,
  estimated_value       NUMERIC(18,2),
  probability           NUMERIC(5,2),
  expected_close_date   DATE,
  notes                 TEXT,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Partnership table (extends Lead)
CREATE TABLE IF NOT EXISTS "Partnership" (
  id                BIGINT PRIMARY KEY REFERENCES "Lead"(id) ON DELETE CASCADE,
  organization_id   BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  partnership_type  TEXT,
  partnership_name  TEXT NOT NULL,
  start_date        DATE,
  end_date          DATE,
  notes             TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ==============================
-- STEP 6: Activity and Relationship Tables
-- ==============================

-- Activity table (polymorphic association to track interactions)
CREATE TABLE IF NOT EXISTS "Activity" (
  id                  BIGSERIAL PRIMARY KEY,
  activityable_type   TEXT,  -- 'Deal', 'Lead', 'Job', 'Opportunity', 'Partnership', etc.
  activityable_id     BIGINT,
  activity_type       TEXT NOT NULL,  -- 'Call', 'Email', 'Meeting', 'Note', etc.
  subject             TEXT NOT NULL,
  description         TEXT,
  activity_date       TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Organization_Relationship table (for tracking relationships between organizations)
CREATE TABLE IF NOT EXISTS "Organization_Relationship" (
  id                    BIGSERIAL PRIMARY KEY,
  from_organization_id  BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  to_organization_id    BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  relationship_type     TEXT NOT NULL,  -- 'Investor', 'Customer', 'Vendor', 'Partner', etc.
  notes                 TEXT,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT no_self_relationship CHECK (from_organization_id != to_organization_id)
);

-- Individual_Relationship table (for tracking relationships between individuals)
CREATE TABLE IF NOT EXISTS "Individual_Relationship" (
  id                  BIGSERIAL PRIMARY KEY,
  from_individual_id  BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  to_individual_id    BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  relationship_type   TEXT NOT NULL,  -- 'Colleague', 'Reports To', 'Mentor', etc.
  notes               TEXT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT no_self_relationship_ind CHECK (from_individual_id != to_individual_id)
);

-- ==============================
-- STEP 7: Create Triggers
-- ==============================

-- Apply triggers to all tables with updated_at
DROP TRIGGER IF EXISTS update_status_updated_at ON "Status";
CREATE TRIGGER update_status_updated_at
    BEFORE UPDATE ON "Status"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_project_updated_at ON "Project";
CREATE TRIGGER update_project_updated_at
    BEFORE UPDATE ON "Project"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_deal_updated_at ON "Deal";
CREATE TRIGGER update_deal_updated_at
    BEFORE UPDATE ON "Deal"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organization_updated_at ON "Organization";
CREATE TRIGGER update_organization_updated_at
    BEFORE UPDATE ON "Organization"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_individual_updated_at ON "Individual";
CREATE TRIGGER update_individual_updated_at
    BEFORE UPDATE ON "Individual"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_contact_updated_at ON "Contact";
CREATE TRIGGER update_contact_updated_at
    BEFORE UPDATE ON "Contact"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_lead_updated_at ON "Lead";
CREATE TRIGGER update_lead_updated_at
    BEFORE UPDATE ON "Lead"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_task_updated_at ON "Task";
CREATE TRIGGER update_task_updated_at
    BEFORE UPDATE ON "Task"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_job_updated_at ON "Job";
CREATE TRIGGER update_job_updated_at
    BEFORE UPDATE ON "Job"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_opportunity_updated_at ON "Opportunity";
CREATE TRIGGER update_opportunity_updated_at
    BEFORE UPDATE ON "Opportunity"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_partnership_updated_at ON "Partnership";
CREATE TRIGGER update_partnership_updated_at
    BEFORE UPDATE ON "Partnership"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_activity_updated_at ON "Activity";
CREATE TRIGGER update_activity_updated_at
    BEFORE UPDATE ON "Activity"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_org_relationship_updated_at ON "Organization_Relationship";
CREATE TRIGGER update_org_relationship_updated_at
    BEFORE UPDATE ON "Organization_Relationship"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_ind_relationship_updated_at ON "Individual_Relationship";
CREATE TRIGGER update_ind_relationship_updated_at
    BEFORE UPDATE ON "Individual_Relationship"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ==============================
-- STEP 8: Seed Status Data
-- ==============================

-- Job Statuses
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('Applied', 'job', false, 'Job application submitted'),
('Interviewing', 'job', false, 'In interview process'),
('Offer', 'job', false, 'Offer received'),
('Accepted', 'job', true, 'Job offer accepted'),
('Rejected', 'job', true, 'Application or candidacy rejected')
ON CONFLICT (name) DO NOTHING;

-- Lead Stages
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('New', 'lead', false, 'New lead created'),
('Qualified', 'lead', false, 'Lead has been qualified'),
('Nurturing', 'lead', false, 'Nurturing the relationship'),
('Discovery', 'lead', false, 'Discovery phase'),
('Proposal', 'lead', false, 'Proposal submitted'),
('Negotiation', 'lead', false, 'In negotiation'),
('Won', 'lead', true, 'Lead won'),
('Lost', 'lead', true, 'Lead lost'),
('On Hold', 'lead', false, 'Lead on hold')
ON CONFLICT (name) DO NOTHING;

-- Deal Stages
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('Prospecting', 'deal', false, 'Prospecting phase'),
('Qualification', 'deal', false, 'Qualification phase'),
('Negotiating', 'deal', false, 'Negotiating terms'),
('Closed Won', 'deal', true, 'Deal won'),
('Closed Lost', 'deal', true, 'Deal lost')
ON CONFLICT (name) DO NOTHING;

-- Task Statuses
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('To Do', 'task_status', false, 'Task not started'),
('In Progress', 'task_status', false, 'Task in progress'),
('Completed', 'task_status', true, 'Task completed'),
('Cancelled', 'task_status', true, 'Task cancelled')
ON CONFLICT (name) DO NOTHING;

-- Task Priorities
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('Low', 'task_priority', false, 'Low priority'),
('Medium', 'task_priority', false, 'Medium priority'),
('High', 'task_priority', false, 'High priority'),
('Urgent', 'task_priority', false, 'Urgent priority')
ON CONFLICT (name) DO NOTHING;

-- ==============================
-- STEP 9: Configure Status Workflows
-- ==============================

-- Configure Job workflow
INSERT INTO "Job_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'Applied' AND category = 'job'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Job_Status" (status_id, display_order, is_default)
SELECT id, 1, false FROM "Status" WHERE name = 'Interviewing' AND category = 'job'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Job_Status" (status_id, display_order, is_default)
SELECT id, 2, false FROM "Status" WHERE name = 'Offer' AND category = 'job'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Job_Status" (status_id, display_order, is_default)
SELECT id, 3, false FROM "Status" WHERE name = 'Accepted' AND category = 'job'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Job_Status" (status_id, display_order, is_default)
SELECT id, 4, false FROM "Status" WHERE name = 'Rejected' AND category = 'job'
ON CONFLICT (status_id) DO NOTHING;

-- Configure Lead workflow
INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'New' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 1, false FROM "Status" WHERE name = 'Qualified' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 2, false FROM "Status" WHERE name = 'Nurturing' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 3, false FROM "Status" WHERE name = 'Discovery' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 4, false FROM "Status" WHERE name = 'Proposal' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 5, false FROM "Status" WHERE name = 'Negotiation' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 6, false FROM "Status" WHERE name = 'Won' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 7, false FROM "Status" WHERE name = 'Lost' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 8, false FROM "Status" WHERE name = 'On Hold' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

-- Configure Deal workflow
INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'Prospecting' AND category = 'deal'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 1, false FROM "Status" WHERE name = 'Qualification' AND category = 'deal'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 2, false FROM "Status" WHERE name = 'Negotiating' AND category = 'deal'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 3, false FROM "Status" WHERE name = 'Closed Won' AND category = 'deal'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 4, false FROM "Status" WHERE name = 'Closed Lost' AND category = 'deal'
ON CONFLICT (status_id) DO NOTHING;

-- Configure Task statuses
INSERT INTO "Task_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'To Do' AND category = 'task_status'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Task_Status" (status_id, display_order, is_default)
SELECT id, 1, false FROM "Status" WHERE name = 'In Progress' AND category = 'task_status'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Task_Status" (status_id, display_order, is_default)
SELECT id, 2, false FROM "Status" WHERE name = 'Completed' AND category = 'task_status'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Task_Status" (status_id, display_order, is_default)
SELECT id, 3, false FROM "Status" WHERE name = 'Cancelled' AND category = 'task_status'
ON CONFLICT (status_id) DO NOTHING;

-- Configure Task priorities
INSERT INTO "Task_Priority" (status_id, display_order, is_default)
SELECT id, 0, false FROM "Status" WHERE name = 'Low' AND category = 'task_priority'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Task_Priority" (status_id, display_order, is_default)
SELECT id, 1, true FROM "Status" WHERE name = 'Medium' AND category = 'task_priority'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Task_Priority" (status_id, display_order, is_default)
SELECT id, 2, false FROM "Status" WHERE name = 'High' AND category = 'task_priority'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Task_Priority" (status_id, display_order, is_default)
SELECT id, 3, false FROM "Status" WHERE name = 'Urgent' AND category = 'task_priority'
ON CONFLICT (status_id) DO NOTHING;

-- ==============================
-- Migration Complete!
-- ==============================

-- Verification Queries:
-- SELECT COUNT(*) as total_jobs FROM "Job";
-- SELECT COUNT(*) as total_leads FROM "Lead" WHERE type = 'Job';
-- SELECT COUNT(*) as total_organizations FROM "Organization";
-- SELECT category, COUNT(*) FROM "Status" GROUP BY category ORDER BY category;

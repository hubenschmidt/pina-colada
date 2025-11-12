-- ==============================
-- DealTracker Migration 002
-- ==============================
-- Migrates from simple Job table to full DealTracker schema
--
-- WARNING: This is a complex migration that:
-- 1. Creates new tables (Status, Deal, Lead, Organization, etc.)
-- 2. Migrates existing Job data to new structure
-- 3. Transforms Job table to use Joined Table Inheritance
--
-- BACKUP YOUR DATABASE BEFORE RUNNING THIS MIGRATION!
--
-- Migration Steps:
-- 1. Create all new tables
-- 2. Seed status data and configure workflows
-- 3. Migrate data from old Job structure to new structure (if exists)
-- 4. Transform Job table to match new schema
-- 5. Create indexes
-- ==============================

-- ==============================
-- STEP 1: Create New Tables
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
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  description     TEXT,
  owner_user_id   BIGINT,
  current_status_id BIGINT REFERENCES "Status"(id) ON DELETE SET NULL,
  value_amount    NUMERIC(18,2),
  value_currency  TEXT DEFAULT 'USD',
  probability     NUMERIC(5,2),
  close_date      DATE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Organization table (for companies)
CREATE TABLE IF NOT EXISTS "Organization" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  website         TEXT,
  phone           TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_org_name UNIQUE (LOWER(name))
);

-- Individual table (for people)
CREATE TABLE IF NOT EXISTS "Individual" (
  id              BIGSERIAL PRIMARY KEY,
  given_name      TEXT NOT NULL,
  family_name     TEXT NOT NULL,
  email           TEXT,
  phone           TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Contact table
CREATE TABLE IF NOT EXISTS "Contact" (
  id                BIGSERIAL PRIMARY KEY,
  individual_id     BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  organization_id   BIGINT REFERENCES "Organization"(id) ON DELETE SET NULL,
  title             TEXT,
  department        TEXT,
  email             TEXT,
  phone             TEXT,
  is_primary        BOOLEAN NOT NULL DEFAULT FALSE,
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
  current_priority_id BIGINT REFERENCES "Status"(id) ON DELETE SET NULL,
  due_date            DATE,
  completed_at        TIMESTAMPTZ,
  assigned_to_user_id BIGINT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Status Configuration Tables
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

-- Junction Tables
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
-- STEP 2: Create Triggers
-- ==============================

-- update_updated_at_column() function already exists from 001_initial_schema.sql

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

-- ==============================
-- STEP 3: Seed Status Data
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
SELECT id, 2, false FROM "Status" WHERE name = 'Discovery' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 3, false FROM "Status" WHERE name = 'Proposal' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 4, false FROM "Status" WHERE name = 'Negotiation' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 5, false FROM "Status" WHERE name = 'Won' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 6, false FROM "Status" WHERE name = 'Lost' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 7, false FROM "Status" WHERE name = 'On Hold' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

-- Configure Deal workflow (reusing some Lead statuses)
INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'Prospecting' AND category = 'deal'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 1, false FROM "Status" WHERE name = 'Qualification' AND category = 'deal'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 2, false FROM "Status" WHERE name = 'Proposal' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 3, false FROM "Status" WHERE name = 'Negotiation' AND category = 'lead'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 4, false FROM "Status" WHERE name = 'Closed Won' AND category = 'deal'
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 5, false FROM "Status" WHERE name = 'Closed Lost' AND category = 'deal'
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
-- STEP 4: Migrate Existing Job Data (Production Only)
-- ==============================
-- This section only runs if the old Job table with UUID exists

DO $$
BEGIN
  -- Check if old Job table exists with UUID primary key
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'Job'
    AND column_name = 'id'
    AND data_type = 'uuid'
  ) THEN
    -- Create temporary mapping table
    CREATE TEMP TABLE job_id_mapping (
      old_uuid UUID,
      new_lead_id BIGINT
    );

    -- Migrate company names to Organization table
    INSERT INTO "Organization" (name, created_at, updated_at)
    SELECT DISTINCT
      company,
      MIN(created_at),
      MAX(updated_at)
    FROM "Job"
    GROUP BY company
    ON CONFLICT (LOWER(name)) DO NOTHING;

    -- Create default Deal for migration
    INSERT INTO "Deal" (name, description, current_status_id, created_at, updated_at)
    VALUES (
      'Job Search - Migrated Data',
      'Auto-created deal for migrating existing job applications',
      (SELECT id FROM "Status" WHERE name = 'Prospecting' AND category = 'deal' LIMIT 1),
      NOW(),
      NOW()
    )
    ON CONFLICT DO NOTHING;

    -- Migrate Job data
    DECLARE
      default_deal_id BIGINT;
      job_record RECORD;
      new_lead_id BIGINT;
      org_id BIGINT;
      mapped_status_id BIGINT;
    BEGIN
      SELECT id INTO default_deal_id FROM "Deal" WHERE name = 'Job Search - Migrated Data' LIMIT 1;

      FOR job_record IN SELECT * FROM "Job" ORDER BY created_at LOOP
        SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER(job_record.company) LIMIT 1;

        mapped_status_id := CASE job_record.status
          WHEN 'applied' THEN (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job' LIMIT 1)
          WHEN 'interviewing' THEN (SELECT id FROM "Status" WHERE name = 'Interviewing' AND category = 'job' LIMIT 1)
          WHEN 'offer' THEN (SELECT id FROM "Status" WHERE name = 'Offer' AND category = 'job' LIMIT 1)
          WHEN 'accepted' THEN (SELECT id FROM "Status" WHERE name = 'Accepted' AND category = 'job' LIMIT 1)
          WHEN 'rejected' THEN (SELECT id FROM "Status" WHERE name = 'Rejected' AND category = 'job' LIMIT 1)
          ELSE (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job' LIMIT 1)
        END;

        INSERT INTO "Lead" (
          deal_id, type, title, description, source, current_status_id, created_at, updated_at
        ) VALUES (
          default_deal_id, 'Job', job_record.company || ' - ' || job_record.job_title,
          job_record.notes, job_record.source, mapped_status_id,
          job_record.created_at, job_record.updated_at
        ) RETURNING id INTO new_lead_id;

        INSERT INTO job_id_mapping (old_uuid, new_lead_id) VALUES (job_record.id, new_lead_id);
      END LOOP;
    END;

    -- Rename old Job table
    ALTER TABLE "Job" RENAME TO "Job_old";

    -- Drop old indexes and triggers
    DROP INDEX IF EXISTS idx_company_job_title;
    DROP INDEX IF EXISTS idx_date;
    DROP INDEX IF EXISTS idx_status;
    DROP TRIGGER IF EXISTS update_job_updated_at ON "Job_old";

    RAISE NOTICE 'Migrated existing Job data to new schema';
  ELSE
    RAISE NOTICE 'No existing Job table found with UUID - skipping migration';
  END IF;
END $$;

-- ==============================
-- STEP 5: Create New Job Table
-- ==============================

-- Create new Job table with Joined Table Inheritance structure
CREATE TABLE IF NOT EXISTS "Job" (
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

-- Migrate data from Job_old if it exists
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'Job_old') THEN
    INSERT INTO "Job" (
      id, organization_id, job_title, job_url, notes, resume_date, salary_range, created_at, updated_at
    )
    SELECT
      m.new_lead_id, o.id, j.job_title, j.job_url, j.notes, j.resume, j.salary_range,
      j.created_at, j.updated_at
    FROM "Job_old" j
    JOIN job_id_mapping m ON j.id = m.old_uuid
    JOIN "Organization" o ON LOWER(o.name) = LOWER(j.company);

    RAISE NOTICE 'Migrated Job records to new table';
  END IF;
END $$;

-- Create trigger for new Job table
DROP TRIGGER IF EXISTS update_job_updated_at ON "Job";
CREATE TRIGGER update_job_updated_at
    BEFORE UPDATE ON "Job"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ==============================
-- STEP 6: Create Indexes
-- ==============================

CREATE INDEX IF NOT EXISTS idx_job_organization ON "Job"(organization_id);
CREATE INDEX IF NOT EXISTS idx_job_title ON "Job"(job_title);
CREATE INDEX IF NOT EXISTS idx_job_created ON "Job"(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lead_deal ON "Lead"(deal_id);
CREATE INDEX IF NOT EXISTS idx_lead_status ON "Lead"(current_status_id);
CREATE INDEX IF NOT EXISTS idx_lead_type ON "Lead"(type);
CREATE INDEX IF NOT EXISTS idx_lead_created ON "Lead"(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_deal_status ON "Deal"(current_status_id);
CREATE INDEX IF NOT EXISTS idx_task_taskable ON "Task"(taskable_type, taskable_id);
CREATE INDEX IF NOT EXISTS idx_task_status ON "Task"(current_status_id);
CREATE INDEX IF NOT EXISTS idx_task_priority ON "Task"(current_priority_id);
CREATE INDEX IF NOT EXISTS idx_task_due_date ON "Task"(due_date);
CREATE INDEX IF NOT EXISTS idx_organization_name ON "Organization"(LOWER(name));

-- ==============================
-- STEP 7: Cleanup (Optional)
-- ==============================

-- Keep Job_old table for safety
-- Uncomment to drop after verifying migration:
-- DROP TABLE IF EXISTS "Job_old" CASCADE;

-- ==============================
-- Migration Complete!
-- ==============================

-- Verification Queries:
-- SELECT COUNT(*) as total_jobs FROM "Job";
-- SELECT COUNT(*) as total_leads FROM "Lead" WHERE type = 'Job';
-- SELECT COUNT(*) as total_organizations FROM "Organization";
-- SELECT category, COUNT(*) FROM "Status" GROUP BY category ORDER BY category;

# DealTracker Data Model & Seeds
_Last updated: 2025-11-12_

This document describes a CRM-style **DealTracker** schema using capitalized table names (`Deal`, `Lead`, `Job_Status`, etc.) and includes helpful seed data for stages and roles.

**Key Design Pattern - Joined Table Inheritance:**
- **Lead** is the base table containing common attributes for all lead types (title, description, source, etc.)
- **Job** extends Lead via Joined Table Inheritance, adding job-specific columns (organization_id, job_title, etc.)
- Uses a discriminator column (`type`) in Lead table to identify the lead type ('Job', 'Opportunity', 'Partnership')
- Future lead types (Opportunity, Partnership, etc.) extend Lead following the same pattern
- SQLAlchemy and other ORMs have native support for this inheritance pattern
- Enables uniform CRM tracking (contacts, organizations, projects, status) across all lead types

---


## Entity-Relationship Summary

- **Status** (central status/stage definitions)
  - Contains all possible statuses across the system
  - Referenced by Deal, Lead, and Job through current_status_id
  - Configured per entity type via Job_Status, Lead_Status, Deal_Status tables
  - Supports ordered status workflows

- **Project**
  - Many **Deals** (many-to-many)
  - Many **Leads** (many-to-many)
  - Many **Contacts** (many-to-many)
  - Many **Individuals** (many-to-many)
  - Many **Organizations** (many-to-many)
  - Many **Tasks** (polymorphic)

- **Deal**
  - Has many **Leads**
  - References current **Status** (via current_status_id)
  - Available statuses configured via **Deal_Status** join table
  - Many **Projects** (many-to-many)
  - Many **Tasks** (polymorphic)

- **Lead** (base table for all lead types)
  - Belongs to one **Deal**
  - Has discriminator column (`type`) to identify subtype: 'Job', 'Opportunity', 'Partnership', etc.
  - References current **Status** (via current_status_id)
  - Available statuses configured via **Lead_Status** join table
  - Many **Contacts** (each Contact **must** be an **Individual**)
  - Many **Organizations**
  - Many **Individuals** (direct link, not just through Contact)
  - Many **Projects** (many-to-many)
  - Many **Tasks** (polymorphic)
  - **Constraint:** A Lead must have **at least one** associated **Individual or Organization**

- **Job** (extends Lead via Joined Table Inheritance)
  - Inherits all Lead attributes and relationships (deal, contacts, organizations, status, etc.)
  - References **Organization** (via organization_id) - the company for this job
  - Adds job-specific columns: job_title, job_url, notes, resume date, salary_range
  - Has foreign key (`id`) referencing Lead table (shared primary key pattern)
  - Represents job application opportunities in the CRM
  - Status inherited from Lead.current_status_id
  - Available statuses configured via **Job_Status** join table
  - Many **Tasks** (polymorphic, inherited from Lead)

- **Organization**
  - Has many **Contacts**
  - Many **Projects** (many-to-many)
  - Many **Tasks** (polymorphic)

- **Individual**
  - Represents a person
  - Contacts model an individual's role/context (often tied to an Organization)
  - Many **Projects** (many-to-many)
  - Many **Tasks** (polymorphic)

- **Contact**
  - Must reference an **Individual** (required)
  - May reference an **Organization** (optional)
  - Represents the **person-in-a-role** (title/department, role-specific email/phone)
  - Many **Projects** (many-to-many)
  - Many **Tasks** (polymorphic)

- **Task** (polymorphic association to any entity)
  - Can be associated with **any entity** via `taskable_type` and `taskable_id`
  - Supports: Project, Deal, Lead, Job, Organization, Individual, Contact
  - Tracks: title, description, status, priority, due_date, assigned_to, completion

---

## Postgres DDL

```sql
-- ==============================
-- 1) Migration Note: Pre-existing Job Table
-- ==============================
-- Note: The "Job" table already exists in production from migration 001_initial_schema.sql
-- Original structure:
--   - id (UUID PRIMARY KEY)
--   - company, job_title, date, status, job_url, notes, resume, salary_range, source
--   - created_at, updated_at
--
-- TARGET SCHEMA uses Joined Table Inheritance:
--   - Lead is the base table with common columns (id, deal_id, title, description, source, type)
--   - Job extends Lead, adding job-specific columns (organization_id FK, job_title, job_url, etc.)
--   - Job.id is both PK and FK to Lead.id
--   - Job.organization_id references Organization table (company data migrated to Organization)
--
-- Future lead types follow the same pattern:
--   - CREATE TABLE "Opportunity" (id FK to Lead, specific columns...)
--   - CREATE TABLE "Partnership" (id FK to Lead, specific columns...)
--
-- Migration path from current state to target:
--   1. Create Organization table
--   2. Migrate company names from Job.company to Organization table
--   3. Create Lead base table
--   4. Create Lead records for existing Jobs
--   5. Alter Job table:
--      - Add organization_id FK to Organization
--      - Add FK to Lead.id
--      - Remove company text column (data now in Organization)
--      - Remove status column (now in Lead.current_status_id via Status table)
--
-- IMPORTANT: The DDL below uses BIGSERIAL for illustration, but actual
-- migrations use UUID to match the existing Job table.

-- ==============================
-- 2) Core DealTracker Tables
-- ==============================

-- Project table
CREATE TABLE "Project" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  description     TEXT,
  owner_user_id   BIGINT,
  status          TEXT,                    -- Active, On Hold, Completed, Archived
  start_date      DATE,
  end_date        DATE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Status table (central status/stage definitions)
CREATE TABLE "Status" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL UNIQUE,    -- e.g., 'Applied', 'Interviewing', 'New', 'Qualified', 'Won', 'Lost'
  description     TEXT,
  category        TEXT,                     -- Optional grouping: 'job', 'lead', 'deal', 'task_status', 'task_priority'
  is_terminal     BOOLEAN DEFAULT FALSE,   -- Indicates final state (e.g., 'Won', 'Lost', 'Accepted', 'Rejected')
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Deal table
CREATE TABLE "Deal" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  description     TEXT,
  owner_user_id   BIGINT,
  current_status_id BIGINT REFERENCES "Status"(id) ON DELETE SET NULL,
  value_amount    NUMERIC(18,2),
  value_currency  TEXT DEFAULT 'USD',
  probability     NUMERIC(5,2),            -- 0..100 (pct)
  close_date      DATE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Lead table (base table for Joined Table Inheritance)
CREATE TABLE "Lead" (
  id              BIGSERIAL PRIMARY KEY,
  deal_id         BIGINT NOT NULL REFERENCES "Deal"(id) ON DELETE CASCADE,
  type            TEXT NOT NULL,           -- Discriminator: 'Job', 'Opportunity', 'Partnership', etc.
  title           TEXT NOT NULL,           -- Short label for the lead
  description     TEXT,
  source          TEXT,                    -- inbound/referral/event/campaign/agent/etc.
  current_status_id BIGINT REFERENCES "Status"(id) ON DELETE SET NULL,
  owner_user_id   BIGINT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Job table (extends Lead via Joined Table Inheritance)
CREATE TABLE "Job" (
  id              BIGINT PRIMARY KEY REFERENCES "Lead"(id) ON DELETE CASCADE,
  organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  job_title       TEXT NOT NULL,
  job_url         TEXT,
  notes           TEXT,
  resume_date     TIMESTAMP,               -- Date resume was submitted (renamed from 'resume')
  salary_range    TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Job_Status: defines which statuses are valid for Jobs and their order
CREATE TABLE "Job_Status" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,        -- 0, 1, 2, 3... determines status sequence
  is_default      BOOLEAN DEFAULT FALSE,   -- Marks the default initial status
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

-- Lead_Status: defines which statuses are valid for Leads and their order
CREATE TABLE "Lead_Status" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,        -- 0, 1, 2, 3... determines status sequence
  is_default      BOOLEAN DEFAULT FALSE,   -- Marks the default initial status
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

-- Deal_Status: defines which statuses are valid for Deals and their order
CREATE TABLE "Deal_Status" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,        -- 0, 1, 2, 3... determines status sequence
  is_default      BOOLEAN DEFAULT FALSE,   -- Marks the default initial status
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

-- Task_Status: defines which statuses are valid for Tasks and their order
CREATE TABLE "Task_Status" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,        -- 0, 1, 2, 3... determines status sequence
  is_default      BOOLEAN DEFAULT FALSE,   -- Marks the default initial status
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

-- Task_Priority: defines which priorities are valid for Tasks and their order
CREATE TABLE "Task_Priority" (
  id              BIGSERIAL PRIMARY KEY,
  status_id       BIGINT NOT NULL REFERENCES "Status"(id) ON DELETE CASCADE,
  display_order   INTEGER NOT NULL,        -- 0, 1, 2, 3... determines priority importance
  is_default      BOOLEAN DEFAULT FALSE,   -- Marks the default priority
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (status_id),
  UNIQUE (display_order)
);

CREATE TABLE "Organization" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  website         TEXT,
  phone           TEXT,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (LOWER(name))
);

CREATE TABLE "Individual" (
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
CREATE TABLE "Contact" (
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

-- Task table (polymorphic association to any entity)
CREATE TABLE "Task" (
  id                  BIGSERIAL PRIMARY KEY,
  taskable_type       TEXT,                    -- Polymorphic: 'Deal', 'Lead', 'Job', 'Project', 'Organization', 'Individual', 'Contact'
  taskable_id         BIGINT,                  -- Polymorphic: points to parent record ID
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

-- ==============================
-- 3) Junction Tables (many-to-manys)
-- ==============================

-- Lead ↔ Contacts
CREATE TABLE "Lead_Contact" (
  lead_id       BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  contact_id    BIGINT NOT NULL REFERENCES "Contact"(id) ON DELETE CASCADE,
  role_on_lead  TEXT,                      -- Decision Maker, Influencer, Billing, etc.
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, contact_id)
);

-- Lead ↔ Organizations
CREATE TABLE "Lead_Organization" (
  lead_id         BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  relationship    TEXT,                    -- Prospect, Customer, Partner, Competitor
  is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, organization_id)
);

-- Lead ↔ Individuals (direct)
CREATE TABLE "Lead_Individual" (
  lead_id       BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  individual_id BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  relationship  TEXT,                      -- Champion, Sponsor, Evaluator, etc.
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (lead_id, individual_id)
);

-- Project ↔ Deals
CREATE TABLE "Project_Deal" (
  project_id    BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  deal_id       BIGINT NOT NULL REFERENCES "Deal"(id) ON DELETE CASCADE,
  relationship  TEXT,                      -- Primary, Related, Dependent
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, deal_id)
);

-- Project ↔ Leads
CREATE TABLE "Project_Lead" (
  project_id    BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  lead_id       BIGINT NOT NULL REFERENCES "Lead"(id) ON DELETE CASCADE,
  relationship  TEXT,                      -- Primary, Related, Dependent
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, lead_id)
);

-- Project ↔ Contacts
CREATE TABLE "Project_Contact" (
  project_id    BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  contact_id    BIGINT NOT NULL REFERENCES "Contact"(id) ON DELETE CASCADE,
  role          TEXT,                      -- Project Manager, Stakeholder, Team Member
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, contact_id)
);

-- Project ↔ Organizations
CREATE TABLE "Project_Organization" (
  project_id      BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
  relationship    TEXT,                    -- Client, Partner, Vendor, Stakeholder
  is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, organization_id)
);

-- Project ↔ Individuals
CREATE TABLE "Project_Individual" (
  project_id    BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  individual_id BIGINT NOT NULL REFERENCES "Individual"(id) ON DELETE CASCADE,
  role          TEXT,                      -- Sponsor, Stakeholder, Contributor
  is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, individual_id)
);

-- ==============================
-- 4) Constraint: Lead must have ≥1 Individual or Organization
-- ==============================

-- KISS Approach: Application-level validation instead of database triggers
-- See dealtracker_models.py for @validates decorator implementation
-- This keeps the database schema simple while maintaining data integrity

```

---

## SQLAlchemy ORM Mappings

A complete SQLAlchemy 2.0+ implementation of this schema is available in **`dealtracker_models.py`**.

The implementation follows KISS/YAGNI principles while maintaining all features:
- **Joined Table Inheritance**: Lead (base) → Job (subclass) using polymorphic configuration
- **Association Objects**: All junction tables with extra columns (role, relationship, is_primary)
- **Type Safety**: Modern `Mapped[T]` annotations for excellent IDE support
- **All Relationships**: Properly configured `relationship()` with `back_populates`
- **Cascades**: Appropriate cascade rules for parent-child relationships
- **Simple Task Queries**: Helper function pattern instead of complex viewonly relationships
- **Flexible Status System**: All statuses/stages/priorities in database (no hardcoded constants)
- **Application Validation**: Python validators instead of complex database triggers

### Key Patterns

**Joined Table Inheritance (Lead → Job):**
```python
class Lead(Base):
    __tablename__ = "Lead"
    type: Mapped[str]  # Discriminator column
    
    __mapper_args__ = {
        "polymorphic_on": "type",
        "polymorphic_identity": "lead"
    }

class Job(Lead):
    __tablename__ = "Job"
    id: Mapped[int] = mapped_column(ForeignKey("Lead.id"), primary_key=True)
    organization_id: Mapped[int] = mapped_column(ForeignKey("Organization.id"))
    job_title: Mapped[str]
    organization: Mapped["Organization"] = relationship()

    __mapper_args__ = {"polymorphic_identity": "Job"}
```

**Association Objects (many-to-many with extra columns):**
```python
class LeadContact(Base):
    __tablename__ = "Lead_Contact"
    lead_id: Mapped[int] = mapped_column(ForeignKey("Lead.id"), primary_key=True)
    contact_id: Mapped[int] = mapped_column(ForeignKey("Contact.id"), primary_key=True)
    role_on_lead: Mapped[Optional[str]]  # Extra column
    is_primary: Mapped[bool] = mapped_column(default=False)

    lead: Mapped["Lead"] = relationship(back_populates="lead_contacts")
    contact: Mapped["Contact"] = relationship(back_populates="lead_contacts")
```

**Polymorphic Task Queries (simple helper function pattern):**
```python
def get_tasks_for_entity(session, entity_type: str, entity_id: int):
    """
    Query tasks for any entity type.

    Example:
        get_tasks_for_entity(session, 'Deal', deal.id)
        get_tasks_for_entity(session, 'Job', job.id)
    """
    return session.query(Task).filter_by(
        taskable_type=entity_type,
        taskable_id=entity_id
    ).all()
```

**Flexible Status Pattern:**

The schema uses a flexible status system with a central `Status` table and entity-specific join tables:

```python
class Status(Base):
    __tablename__ = "Status"
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(unique=True)
    category: Mapped[Optional[str]]  # 'job', 'lead', 'deal', 'task_status', 'task_priority'
    is_terminal: Mapped[bool] = mapped_column(default=False)

class JobStatus(Base):
    __tablename__ = "Job_Status"
    id: Mapped[int] = mapped_column(primary_key=True)
    status_id: Mapped[int] = mapped_column(ForeignKey("Status.id"))
    display_order: Mapped[int]  # 0, 1, 2, 3... defines workflow sequence
    is_default: Mapped[bool] = mapped_column(default=False)

    status: Mapped["Status"] = relationship()

# Deal and Lead use similar patterns with Deal_Status and Lead_Status
```

**Benefits:**
- **Flexible Workflows**: Add/remove statuses without schema changes
- **Ordered Progression**: `display_order` defines status sequences (0→1→2→3)
- **Reusable Statuses**: Same status ('Won', 'Lost') can be used across entity types
- **Type-Specific**: Each entity (Job, Lead, Deal) has its own configured status list
- **Default Status**: `is_default` marks initial status for new records

See **`dealtracker_models.py`** for the complete implementation with all 20+ models and relationships.

---

## Seed Roles & Stages

**Flexible Status System**: All statuses, stages, and priorities are stored in the `Status` table for maximum flexibility. This allows:
- Dynamic configuration without code changes
- Sharing statuses across entity types (e.g., 'Won', 'Lost')
- UI-driven workflow customization
- Ordered progressions via display_order
- Easy addition of new workflows

**Entity Types with Status Management:**
- **Job Statuses**: Applied → Interviewing → Offer → Accepted/Rejected
- **Lead Stages**: New → Qualified → Discovery → Proposal → Negotiation → Won/Lost
- **Deal Stages**: Prospecting → Qualification → Proposal → Negotiation → Closed Won/Lost
- **Task Statuses**: To Do → In Progress → Completed/Cancelled
- **Task Priorities**: Low → Medium → High → Urgent

**Other Enumerations** (stored as TEXT without Status table):
- Contact roles, organization relationships, individual relationships
- Project roles and relationships
- These are simple TEXT fields without workflow requirements

### Python Seeds (SQLAlchemy)

```python
from sqlalchemy.orm import Session
from dealtracker_models import *
from datetime import date, timedelta

# Assuming you have a session
session = Session(engine)

# ==============================
# 1. Seed Status table and configure workflows
# ==============================

# Create all statuses
job_statuses_data = [
    ('Applied', 'job', False),
    ('Interviewing', 'job', False),
    ('Offer', 'job', False),
    ('Accepted', 'job', True),
    ('Rejected', 'job', True),
]

lead_statuses_data = [
    ('New', 'lead', False),
    ('Qualified', 'lead', False),
    ('Discovery', 'lead', False),
    ('Proposal', 'lead', False),
    ('Negotiation', 'lead', False),
    ('Won', 'lead', True),
    ('Lost', 'lead', True),
    ('On Hold', 'lead', False),
]

deal_statuses_data = [
    ('Prospecting', 'deal', False),
    ('Qualification', 'deal', False),
    ('Proposal', 'deal', False),
    ('Negotiation', 'deal', False),
    ('Closed Won', 'deal', True),
    ('Closed Lost', 'deal', True),
]

task_statuses_data = [
    ('To Do', 'task_status', False),
    ('In Progress', 'task_status', False),
    ('Completed', 'task_status', True),
    ('Cancelled', 'task_status', True),
]

task_priorities_data = [
    ('Low', 'task_priority', False),
    ('Medium', 'task_priority', False),
    ('High', 'task_priority', False),
    ('Urgent', 'task_priority', False),
]

# Insert all statuses
all_statuses = {}
all_data = job_statuses_data + lead_statuses_data + deal_statuses_data + task_statuses_data + task_priorities_data
for name, category, is_terminal in all_data:
    if name not in all_statuses:  # Avoid duplicates (e.g., 'Proposal' in both lead and deal)
        status = Status(name=name, category=category, is_terminal=is_terminal)
        session.add(status)
        all_statuses[name] = status

session.flush()

# Configure Job workflow
for order, (name, _, _) in enumerate(job_statuses_data):
    job_status = JobStatus(
        status_id=all_statuses[name].id,
        display_order=order,
        is_default=(order == 0)  # First status is default
    )
    session.add(job_status)

# Configure Lead workflow
for order, (name, _, _) in enumerate(lead_statuses_data):
    lead_status = LeadStatus(
        status_id=all_statuses[name].id,
        display_order=order,
        is_default=(order == 0)
    )
    session.add(lead_status)

# Configure Deal workflow
for order, (name, _, _) in enumerate(deal_statuses_data):
    deal_status = DealStatus(
        status_id=all_statuses[name].id,
        display_order=order,
        is_default=(order == 0)
    )
    session.add(deal_status)

# Configure Task statuses
for order, (name, _, _) in enumerate(task_statuses_data):
    task_status = TaskStatus(
        status_id=all_statuses[name].id,
        display_order=order,
        is_default=(order == 0)
    )
    session.add(task_status)

# Configure Task priorities
for order, (name, _, _) in enumerate(task_priorities_data):
    task_priority = TaskPriority(
        status_id=all_statuses[name].id,
        display_order=order,
        is_default=(name == 'Medium')  # Medium is default priority
    )
    session.add(task_priority)

session.commit()

# ==============================
# 2. Example creation flow for a Deal → Job (Lead subclass) with parties and status
# ==============================

# Get default statuses
default_deal_status = session.query(Status).join(DealStatus).filter(DealStatus.is_default == True).first()
default_job_status = session.query(Status).join(JobStatus).filter(JobStatus.is_default == True).first()

deal = Deal(
    name='ACME Expansion',
    value_amount=250000.00,
    value_currency='USD',
    probability=35.00,
    current_status_id=default_deal_status.id  # Set to 'Prospecting'
)
session.add(deal)
session.flush()  # Get the deal.id

# Create organization and individual first (Job needs organization_id)
org = Organization(name='ACME Corp', website='https://acme.example')
session.add(org)

person = Individual(given_name='Avery', family_name='Nguyen', email='avery.nguyen@acme.example')
session.add(person)
session.flush()  # Get org.id and person.id

# Create a Job (which IS-A Lead via inheritance)
job = Job(
    deal_id=deal.id,
    title='ACME Corp - Senior Engineer',
    organization_id=org.id,  # FK to Organization (replaces company text field)
    job_title='Senior Engineer',
    source='agent',
    current_status_id=default_job_status.id  # Set to 'Applied' (inherited from Lead)
)
session.add(job)
session.flush()  # Get the job.id (which is also the lead.id)

# Create contact
contact = Contact(
    individual_id=person.id,
    organization_id=org.id,
    title='Director of Ops',
    department='Operations',
    is_primary=True
)
session.add(contact)
session.flush()

# Associate lead with organizations, individuals, and contacts
lead_org = LeadOrganization(
    lead_id=job.id,
    organization_id=org.id,
    relationship='Prospect',
    is_primary=True
)
session.add(lead_org)

lead_ind = LeadIndividual(
    lead_id=job.id,
    individual_id=person.id,
    relationship='Champion',
    is_primary=True
)
session.add(lead_ind)

lead_contact = LeadContact(
    lead_id=job.id,
    contact_id=contact.id,
    role_on_lead='Decision Maker',
    is_primary=True
)
session.add(lead_contact)

# Example: Create a Project and associate with Deals, Leads, and Parties
project = Project(
    name='ACME Digital Transformation',
    description='Multi-year digital transformation initiative',
    status='Active',
    start_date=date.today()
)
session.add(project)
session.flush()

# Associate the deal and lead with the project
project_deal = ProjectDeal(
    project_id=project.id,
    deal_id=deal.id,
    relationship='Primary'
)
session.add(project_deal)

project_lead = ProjectLead(
    project_id=project.id,
    lead_id=job.id,
    relationship='Primary'
)
session.add(project_lead)

# Associate organizations, individuals, and contacts with the project
project_org = ProjectOrganization(
    project_id=project.id,
    organization_id=org.id,
    relationship='Client',
    is_primary=True
)
session.add(project_org)

project_ind = ProjectIndividual(
    project_id=project.id,
    individual_id=person.id,
    role='Executive Sponsor',
    is_primary=True
)
session.add(project_ind)

project_contact = ProjectContact(
    project_id=project.id,
    contact_id=contact.id,
    role='Project Manager',
    is_primary=True
)
session.add(project_contact)

session.commit()

# ==============================
# Status Workflow Examples
# ==============================

# Query available statuses for Jobs in order
job_workflow = session.query(Status).join(JobStatus).order_by(JobStatus.display_order).all()
print(f"Job workflow: {[s.name for s in job_workflow]}")
# Output: ['Applied', 'Interviewing', 'Offer', 'Accepted', 'Rejected']

# Update job to next status
current_order = session.query(JobStatus.display_order).filter(
    JobStatus.status_id == job.current_status_id
).scalar()

next_status = session.query(Status).join(JobStatus).filter(
    JobStatus.display_order == current_order + 1
).first()

if next_status:
    job.current_status_id = next_status.id
    session.commit()
    print(f"Job status updated to: {next_status.name}")

# Query all Jobs with a specific status
interviewing_status = session.query(Status).filter_by(name='Interviewing').first()
interviewing_jobs = session.query(Job).join(Lead).filter(
    Lead.current_status_id == interviewing_status.id
).all()

# ==============================
# Task Examples (Polymorphic Associations)
# ==============================

# Get default task status and priorities
default_task_status = session.query(Status).join(TaskStatus).filter(TaskStatus.is_default == True).first()
default_task_priority = session.query(Status).join(TaskPriority).filter(TaskPriority.is_default == True).first()
high_priority = session.query(Status).filter_by(name='High', category='task_priority').first()
urgent_priority = session.query(Status).filter_by(name='Urgent', category='task_priority').first()
in_progress_status = session.query(Status).filter_by(name='In Progress', category='task_status').first()

# Create tasks associated with different entities

# Task associated with a Deal
deal_task = Task(
    taskable_type='Deal',
    taskable_id=deal.id,
    title='Follow up with decision maker',
    description='Schedule call to discuss pricing',
    current_status_id=default_task_status.id,  # 'To Do'
    current_priority_id=high_priority.id,       # 'High'
    due_date=date.today() + timedelta(days=7),
    assigned_to_user_id=1
)
session.add(deal_task)

# Task associated with a Job (Lead subclass)
job_task = Task(
    taskable_type='Job',  # Note: Use 'Job' not 'Lead' for Job-specific tasks
    taskable_id=job.id,
    title='Prepare for technical interview',
    description='Review system design patterns',
    current_status_id=in_progress_status.id,    # 'In Progress'
    current_priority_id=urgent_priority.id,      # 'Urgent'
    due_date=date.today() + timedelta(days=2)
)
session.add(job_task)

# Task associated with an Organization
org_task = Task(
    taskable_type='Organization',
    taskable_id=org.id,
    title='Research company background',
    description='Gather info on recent product launches',
    current_status_id=default_task_status.id,   # 'To Do'
    current_priority_id=default_task_priority.id # 'Medium'
)
session.add(org_task)

# Task associated with a Project
project_task = Task(
    taskable_type='Project',
    taskable_id=project.id,
    title='Project kickoff meeting',
    description='Align on scope and timeline with stakeholders',
    current_status_id=default_task_status.id,   # 'To Do'
    current_priority_id=high_priority.id,        # 'High'
    due_date=date.today() + timedelta(days=3)
)
session.add(project_task)

session.commit()

# ==============================
# Querying Tasks
# ==============================

# Using the simple helper function pattern (KISS approach)
from dealtracker_models import get_tasks_for_entity

# Query all tasks for a specific Deal
deal_tasks = get_tasks_for_entity(session, 'Deal', deal.id)

# Query all tasks for a specific Job
job_tasks = get_tasks_for_entity(session, 'Job', job.id)

# Query all tasks regardless of association
all_tasks = session.query(Task).all()

# Query tasks by status
todo_status = session.query(Status).filter_by(name='To Do', category='task_status').first()
pending_tasks = session.query(Task).filter(Task.current_status_id == todo_status.id).all()

# Query overdue tasks (not completed)
from datetime import datetime
completed_status = session.query(Status).filter_by(name='Completed', category='task_status').first()
overdue_tasks = session.query(Task).filter(
    Task.due_date < date.today(),
    Task.current_status_id != completed_status.id
).all()

# Query tasks by priority
high_priority = session.query(Status).filter_by(name='High', category='task_priority').first()
high_priority_tasks = session.query(Task).filter(Task.current_priority_id == high_priority.id).all()

# Get task status workflow
task_statuses = session.query(Status).join(TaskStatus).order_by(TaskStatus.display_order).all()
print(f"Task workflow: {[s.name for s in task_statuses]}")
# Output: ['To Do', 'In Progress', 'Completed', 'Cancelled']
```

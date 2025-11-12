# DealTracker Database Schema

## Overview

Multi-tenant CRM and job tracking system with support for deals, leads, organizations, and contacts.

## Entity Relationship Diagram

```
┌─────────────────┐
│     Tenant      │ (App customer/company)
│─────────────────│
│ id (PK)         │
│ name            │
│ slug (unique)   │
│ status          │
│ plan            │
│ settings (JSON) │
└────────┬────────┘
         │
         │ 1:N
         ├──────────────────────┬──────────────────────┬──────────────────────┐
         │                      │                      │                      │
         ▼                      ▼                      ▼                      ▼
┌────────────────┐    ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
│      User      │    │  Organization   │   │   Individual    │   │     Project     │
│────────────────│    │─────────────────│   │─────────────────│   │─────────────────│
│ id (PK)        │    │ id (PK)         │   │ id (PK)         │   │ id (PK)         │
│ tenant_id (FK) │    │ tenant_id (FK)  │   │ tenant_id (FK)  │   │ tenant_id (FK)  │
│ email          │    │ name (unique/t) │   │ first_name      │   │ name            │
│ first_name     │    │ website         │   │ last_name       │   │ description     │
│ last_name      │    │ phone           │   │ email (unique/t)│   │ owner_user_id   │
│ status         │    │ industry        │   │ phone           │   │ status          │
└────────┬───────┘    └────────┬────────┘   └────────┬────────┘   │ start_date      │
         │                     │                     │            │ end_date        │
         │ N:M                 │ 1:N                 │ 1:N        └─────────────────┘
         │                     │                     │
         ▼                     │                     │                      ▼
┌────────────────┐             │                     │            ┌─────────────────┐
│   UserRole     │             │                     │            │      Deal       │
│────────────────│             │                     │            │─────────────────│
│ user_id (FK)   │◄────────────┼─────────────────────┼────────────│ id (PK)         │
│ role_id (FK)   │             │                     │            │ tenant_id (FK)  │
└────────┬───────┘             │                     │            │ name            │
         │                     │                     │            │ owner_user_id   │
         │ N:1                 │                     │            │ current_status→ │
         ▼                     │                     │            │ value_amount    │
┌────────────────┐             │                     │            │ probability     │
│      Role      │             │                     │            └────────┬────────┘
│────────────────│             │                     │                     │
│ id (PK)        │             │                     │                     │ 1:N
│ name (unique)  │             │                     │                     ▼
│ description    │             │                     │            ┌─────────────────┐
└────────────────┘             │                     │            │      Lead       │
                               │                     │            │─────────────────│
                               │                     │            │ id (PK)         │
                               │                     │            │ deal_id (FK)    │
┌────────────────┐             │                     │            │ type (discrim.) │
│     Status     │             │                     │            │ title           │
│────────────────│             │                     │            │ current_status→ │
│ id (PK)        │─────────────┼─────────────────────┼────────────│ owner_user_id   │
│ name (unique)  │             │                     │            └────────┬────────┘
│ category       │             │                     │                     │
│ description    │             │                     │                     │ Joined Table
│ is_terminal    │             │                     │                     │ Inheritance
└────────────────┘             │                     │         ┌───────────┼──────────┐
                               │                     │         │           │          │
                               │                     │         ▼           ▼          ▼
                               │                     │  ┌──────────┐ ┌─────────┐ ┌────────────┐
                               │                     │  │   Job    │ │Opportun.│ │Partnership │
                               │                     │  │──────────│ │─────────│ │────────────│
                               │                     │  │id(PK/FK) │ │id(PK/FK)│ │id (PK/FK)  │
                               │                     │  │org_id(FK)│ │org_id(FK│ │org_id (FK) │
                               │                     │  │job_title │ │name     │ │name        │
                               └─────────────────────┼─▶│job_url   │ │est_value│ │type        │
                                                     │  │notes     │ │prob.    │ │start_date  │
                                                     │  └──────────┘ └─────────┘ └────────────┘
                                                     │
                                                     └──────────────────┐
                                                                        │
                                                                        ▼
                                                              ┌─────────────────┐
                                                              │    Contact      │
                                                              │─────────────────│
                                                              │ id (PK)         │
                                                              │ individual_id ─►│
                                                              │ organization_id►│
                                                              │ title           │
                                                              │ department      │
                                                              │ role            │
                                                              │ is_primary      │
                                                              └─────────────────┘
```

## Table Details

### Multi-Tenancy Tables

#### Tenant
App-level customer/company (multi-tenant isolation).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| name | TEXT | NOT NULL | Tenant display name |
| slug | TEXT | NOT NULL, UNIQUE | URL-safe identifier |
| status | TEXT | DEFAULT 'active' | active, suspended, trial, cancelled |
| plan | TEXT | DEFAULT 'free' | free, starter, professional, enterprise |
| settings | JSONB | DEFAULT '{}' | Tenant-specific configuration |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Relationships:**
- Has many: Users, Deals, Organizations, Individuals, Projects

#### User
Users within a tenant (multi-tenant users).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| tenant_id | BIGINT | FK → Tenant, NOT NULL | Parent tenant |
| email | TEXT | NOT NULL | User email |
| first_name | TEXT | | User first name |
| last_name | TEXT | | User last name |
| avatar_url | TEXT | | Profile picture URL |
| status | TEXT | DEFAULT 'active' | active, inactive, invited |
| last_login_at | TIMESTAMPTZ | | Last login timestamp |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Unique Constraints:**
- (tenant_id, email) - Email unique per tenant

**Indexes:**
- idx_user_tenant_id
- idx_user_email

**Relationships:**
- Belongs to: Tenant
- Has many: UserRoles

#### Role
System-defined roles.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| name | TEXT | NOT NULL, UNIQUE | Role name (owner, admin, member, viewer) |
| description | TEXT | | Role description |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |

**Default Roles:**
- **owner** - Tenant owner with full access and billing control
- **admin** - Administrator with full access to all features
- **member** - Standard member with access to most features
- **viewer** - Read-only access to view data

**Relationships:**
- Has many: UserRoles

#### UserRole
Junction table for User-Role many-to-many relationship.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| user_id | BIGINT | FK → User, NOT NULL | User reference |
| role_id | BIGINT | FK → Role, NOT NULL | Role reference |
| created_at | TIMESTAMPTZ | NOT NULL | Assignment timestamp |

**Primary Key:** (user_id, role_id)

**Relationships:**
- Belongs to: User, Role

---

### Core CRM Tables

#### Status
Centralized status/stage definitions for all trackable entities.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| name | TEXT | NOT NULL, UNIQUE | Status name |
| category | TEXT | | Status category (job, lead, deal, task_status, task_priority) |
| description | TEXT | | Status description |
| is_terminal | BOOLEAN | DEFAULT false | Whether this is a final state |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |

**Default Job Statuses:**
- Saved, Applied, Interviewing, Offer, Accepted, Rejected, Withdrawn

**Default Lead Statuses:**
- New, Contacted, Qualified, Proposal, Negotiation, Closed Won, Closed Lost

**Relationships:**
- Has many: Deals (current_status), Leads (current_status)

#### Organization
Companies and organizations (CRM entities).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| tenant_id | BIGINT | FK → Tenant | Parent tenant |
| name | TEXT | NOT NULL | Organization name |
| website | TEXT | | Company website |
| phone | TEXT | | Primary phone |
| industry | TEXT | | Industry/sector |
| employee_count | INTEGER | | Number of employees |
| description | TEXT | | Company description |
| notes | TEXT | | Internal notes |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Unique Constraints:**
- (tenant_id, LOWER(name)) - Case-insensitive name unique per tenant

**Indexes:**
- idx_organization_tenant_id
- idx_organization_name_lower_tenant

**Relationships:**
- Belongs to: Tenant
- Has many: Jobs, Opportunities, Partnerships, Contacts

#### Individual
People (separate from their company affiliations).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| tenant_id | BIGINT | FK → Tenant | Parent tenant |
| first_name | TEXT | NOT NULL | Person's first name |
| last_name | TEXT | NOT NULL | Person's last name |
| email | TEXT | | Primary email |
| phone | TEXT | | Primary phone |
| linkedin_url | TEXT | | LinkedIn profile URL |
| title | TEXT | | Job title |
| notes | TEXT | | Internal notes |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Unique Constraints:**
- (tenant_id, LOWER(email)) - Case-insensitive email unique per tenant

**Indexes:**
- idx_individual_tenant_id
- idx_individual_email_lower_tenant

**Relationships:**
- Belongs to: Tenant
- Has many: Contacts

#### Contact
Links an Individual to an Organization with role information.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| individual_id | BIGINT | FK → Individual, NOT NULL | Person reference |
| organization_id | BIGINT | FK → Organization | Company reference |
| title | TEXT | | Job title at this org |
| department | TEXT | | Department |
| role | TEXT | | Role/function |
| email | TEXT | | Work email at this org |
| phone | TEXT | | Work phone at this org |
| is_primary | BOOLEAN | DEFAULT false | Primary contact flag |
| notes | TEXT | | Internal notes |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Relationships:**
- Belongs to: Individual, Organization

#### Project
Projects for organizing work.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| tenant_id | BIGINT | FK → Tenant | Parent tenant |
| name | TEXT | NOT NULL | Project name |
| description | TEXT | | Project description |
| owner_user_id | BIGINT | | Project owner |
| status | TEXT | | Simple text status |
| start_date | DATE | | Start date |
| end_date | DATE | | End date |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Indexes:**
- idx_project_tenant_id

**Relationships:**
- Belongs to: Tenant

---

### Deal Tracking Tables

#### Deal
High-level deals that contain multiple leads.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| tenant_id | BIGINT | FK → Tenant | Parent tenant |
| name | TEXT | NOT NULL | Deal name |
| description | TEXT | | Deal description |
| owner_user_id | BIGINT | | Deal owner |
| current_status_id | BIGINT | FK → Status | Current status |
| value_amount | NUMERIC(18,2) | | Deal value |
| value_currency | TEXT | DEFAULT 'USD' | Currency code |
| probability | NUMERIC(5,2) | | Win probability (0-100) |
| expected_close_date | DATE | | Expected close date |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Indexes:**
- idx_deal_tenant_id

**Relationships:**
- Belongs to: Tenant, Status
- Has many: Leads

#### Lead
Base table for Joined Table Inheritance (Job, Opportunity, Partnership extend this).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-increment ID |
| deal_id | BIGINT | FK → Deal, NOT NULL | Parent deal |
| type | TEXT | NOT NULL | Discriminator: 'Job', 'Opportunity', 'Partnership' |
| title | TEXT | NOT NULL | Lead title |
| description | TEXT | | Lead description |
| source | TEXT | | Lead source |
| current_status_id | BIGINT | FK → Status | Current status |
| owner_user_id | BIGINT | | Lead owner |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Indexes:**
- idx_lead_deal_id
- idx_lead_type

**Relationships:**
- Belongs to: Deal, Status
- Has one: Job, Opportunity, or Partnership (polymorphic via type)

---

### Lead Subtypes (Joined Table Inheritance)

#### Job
Job applications (extends Lead).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PK, FK → Lead | Shared with Lead.id |
| organization_id | BIGINT | FK → Organization, NOT NULL | Employer |
| job_title | TEXT | NOT NULL | Job title |
| job_url | TEXT | | Job posting URL |
| notes | TEXT | | Application notes |
| resume_date | TIMESTAMPTZ | | When resume submitted |
| salary_range | TEXT | | Expected salary |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Relationships:**
- Belongs to: Lead (parent), Organization

**Notes:**
- Status is inherited through Lead.current_status_id
- Company is through organization_id reference

#### Opportunity
Business opportunities (extends Lead).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PK, FK → Lead | Shared with Lead.id |
| organization_id | BIGINT | FK → Organization, NOT NULL | Client organization |
| opportunity_name | TEXT | NOT NULL | Opportunity name |
| estimated_value | NUMERIC(18,2) | | Estimated deal value |
| probability | NUMERIC(5,2) | | Win probability |
| expected_close_date | DATE | | Expected close date |
| notes | TEXT | | Opportunity notes |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Relationships:**
- Belongs to: Lead (parent), Organization

#### Partnership
Partnership opportunities (extends Lead).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PK, FK → Lead | Shared with Lead.id |
| organization_id | BIGINT | FK → Organization, NOT NULL | Partner organization |
| partnership_type | TEXT | | Type of partnership |
| partnership_name | TEXT | NOT NULL | Partnership name |
| start_date | DATE | | Partnership start |
| end_date | DATE | | Partnership end |
| notes | TEXT | | Partnership notes |
| created_at | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Relationships:**
- Belongs to: Lead (parent), Organization

---

## Row Level Security (RLS)

All tenant-scoped tables have RLS enabled:
- Tenant, User, UserRole
- Deal, Organization, Individual, Project
- Lead, Job, Contact

**Status:** RLS policies are currently commented out pending authentication implementation.

**Future Policy Example:**
```sql
CREATE POLICY "Users can view their tenant's data" ON "Deal"
    FOR SELECT
    USING (tenant_id = (SELECT tenant_id FROM "User" WHERE id = auth.uid()));
```

---

## Key Design Patterns

### Multi-Tenancy
- **Tenant** = App customer/company (provides isolation)
- **Organization** = CRM entity (companies in their pipeline)
- Top-level tables (Deal, Organization, Individual, Project) include `tenant_id`
- Child tables inherit tenant context through relationships
- Unique constraints are scoped per tenant where applicable

### Joined Table Inheritance
- **Lead** is the base table with `type` discriminator
- **Job**, **Opportunity**, **Partnership** share the same `id` as their parent Lead
- Common fields (title, status, owner) stored in Lead
- Specific fields stored in subtype tables
- Foreign key from subtype to Lead ensures data integrity

### Status Management
- Centralized **Status** table with `category` field
- Replaces hard-coded status strings
- Allows flexible workflow customization per tenant (future)
- Terminal states marked with `is_terminal` flag

### Contact Model
- **Individual** and **Organization** are independent entities
- **Contact** junction table links them with role-specific information
- Supports people working at multiple companies
- Supports companies with multiple contacts

---

## Migration Files

1. **001_initial_schema.sql** - Legacy initial schema (being phased out)
2. **003_multitenancy.sql** - Multi-tenancy architecture
   - Tenant, User, Role, UserRole tables
   - Adds tenant_id to existing tables
   - Enables RLS (policies commented out)
   - Seeds default roles

---

## Indexes

### Tenant Isolation
- `idx_user_tenant_id` on User(tenant_id)
- `idx_deal_tenant_id` on Deal(tenant_id)
- `idx_organization_tenant_id` on Organization(tenant_id)
- `idx_individual_tenant_id` on Individual(tenant_id)
- `idx_project_tenant_id` on Project(tenant_id)

### Lookups
- `idx_user_email` on User(email)
- `idx_lead_deal_id` on Lead(deal_id)
- `idx_lead_type` on Lead(type)

### Unique Constraints
- `idx_organization_name_lower_tenant` on Organization(tenant_id, LOWER(name))
- `idx_individual_email_lower_tenant` on Individual(tenant_id, LOWER(email))

---

## Notes

- All `id` columns use BIGSERIAL (auto-incrementing BIGINT)
- All timestamps use TIMESTAMPTZ (timezone-aware)
- Nullable `tenant_id` allows gradual migration (will be NOT NULL in future)
- Numeric fields use NUMERIC for precision (e.g., currency amounts)
- Case-insensitive uniqueness via functional indexes with LOWER()

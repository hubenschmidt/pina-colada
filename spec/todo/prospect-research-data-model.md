# Feature: Prospect Research Data Model

## Overview

Enhance Organization and Individual entities to support AI-powered prospect research with comprehensive firmographic, technographic, intent signal, and contact intelligence data. All enriched data will include full provenance tracking (source, confidence, freshness) to support data quality and AI agent audit trails.

## Business Context

### Problem Statement
Current Account/Organization/Individual models lack the data richness needed for sophisticated prospect research. Sales teams cannot:
- Assess company fit (revenue, size, funding stage, tech stack)
- Identify decision-makers and org structure
- Track buying signals (hiring, news, growth indicators)
- Trust data freshness or sources

### Goals
1. **Enable ICP Scoring**: Firmographic and technographic data for Ideal Customer Profile matching
2. **Surface Intent Signals**: Capture buying indicators (funding, hiring, expansion)
3. **Identify Decision-Makers**: Seniority, department, and org chart visibility
4. **Trust the Data**: Field-level provenance for AI agent and manual enrichment

### Target Users
- **Sales Reps**: Research prospects before outreach
- **Sales Ops**: Build targeted lists, segment accounts
- **AI Agent**: Automated enrichment from external sources

---

## User Stories

### US-1: Firmographic Research
**As a** sales rep, **I want** to see company revenue, size, location, and founding year **so that** I can qualify accounts against our ICP.

### US-2: Tech Stack Discovery
**As a** sales rep, **I want** to see what technologies a company uses **so that** I can tailor my pitch to their stack.

### US-3: Intent Signal Monitoring
**As a** sales rep, **I want** to see recent funding rounds, hiring activity, and company news **so that** I can time my outreach to buying signals.

### US-4: Decision-Maker Identification
**As a** sales rep, **I want** to identify decision-makers by seniority and department **so that** I can target the right contacts.

### US-5: Data Provenance
**As a** sales ops manager, **I want** to see when data was last verified and from what source **so that** I can trust the information.

### US-6: AI Agent Enrichment
**As a** system, **I want** to record provenance when the AI agent enriches data **so that** users know what was auto-populated vs manually entered.

---

## Scenarios

### Scenario 1: View Company Firmographics
**Given** a user is viewing an Organization detail page
**When** the page loads
**Then** the user sees revenue range, employee count, founding year, HQ location, and company type

### Scenario 2: View Tech Stack
**Given** a user is viewing an Organization detail page
**When** they expand the "Technology" section
**Then** they see a list of technologies grouped by category (CRM, Cloud, etc.)

### Scenario 3: View Recent Signals
**Given** a user is viewing an Organization detail page
**When** they scroll to the "Signals" section
**Then** they see recent funding rounds, hiring activity, and news sorted by date

### Scenario 4: Filter by Decision-Maker
**Given** a user is on the Individuals list page
**When** they filter by `is_decision_maker = true` and `seniority_level = 'VP'`
**Then** they see only VP-level decision-makers

### Scenario 5: Check Data Freshness
**Given** a user hovers over a firmographic field (e.g., revenue range)
**When** the tooltip appears
**Then** it shows source, confidence %, and last verified date

---

## Database Schema

### New Tables

```sql
-- ==============================================
-- DATA PROVENANCE (Field-level tracking)
-- ==============================================
CREATE TABLE "DataProvenance" (
    id              BIGSERIAL PRIMARY KEY,
    entity_type     TEXT NOT NULL,            -- 'Organization', 'Individual'
    entity_id       BIGINT NOT NULL,
    field_name      TEXT NOT NULL,            -- 'revenue_range_id', 'seniority_level', etc.
    source          TEXT NOT NULL,            -- 'clearbit', 'apollo', 'linkedin', 'agent', 'manual'
    source_url      TEXT,
    confidence      NUMERIC(3,2),             -- 0.00 to 1.00
    verified_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    verified_by     BIGINT REFERENCES "User"(id) ON DELETE SET NULL,  -- NULL = AI agent
    raw_value       JSONB,                    -- Original value from source
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT confidence_range CHECK (confidence >= 0 AND confidence <= 1)
);

CREATE INDEX idx_provenance_entity ON "DataProvenance"(entity_type, entity_id);
CREATE INDEX idx_provenance_field ON "DataProvenance"(entity_type, entity_id, field_name);
CREATE INDEX idx_provenance_stale ON "DataProvenance"(verified_at);

-- ==============================================
-- REVENUE RANGE (Lookup table)
-- ==============================================
CREATE TABLE "RevenueRange" (
    id              BIGSERIAL PRIMARY KEY,
    label           TEXT NOT NULL UNIQUE,
    min_value       BIGINT,                   -- USD, NULL = unbounded
    max_value       BIGINT,                   -- USD, NULL = unbounded
    display_order   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Trigger for updated_at
CREATE TRIGGER update_revenue_range_updated_at
    BEFORE UPDATE ON "RevenueRange"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Seed data
INSERT INTO "RevenueRange" (label, min_value, max_value, display_order) VALUES
('< $1M', 0, 1000000, 0),
('$1M - $10M', 1000000, 10000000, 1),
('$10M - $50M', 10000000, 50000000, 2),
('$50M - $100M', 50000000, 100000000, 3),
('$100M - $500M', 100000000, 500000000, 4),
('$500M - $1B', 500000000, 1000000000, 5),
('$1B+', 1000000000, NULL, 6);

-- ==============================================
-- TECHNOLOGY (Lookup table for tech stack)
-- ==============================================
CREATE TABLE "Technology" (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    category        TEXT NOT NULL,            -- 'CRM', 'Marketing Automation', 'Cloud', 'Database', etc.
    vendor          TEXT,                     -- 'Salesforce', 'HubSpot', 'AWS', etc.
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(name, category)
);

CREATE INDEX idx_technology_category ON "Technology"(category);

-- ==============================================
-- ORGANIZATION_TECHNOLOGY (Junction table)
-- ==============================================
CREATE TABLE "OrganizationTechnology" (
    organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
    technology_id   BIGINT NOT NULL REFERENCES "Technology"(id) ON DELETE CASCADE,
    detected_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    source          TEXT,                     -- 'builtwith', 'wappalyzer', 'agent', 'manual'
    confidence      NUMERIC(3,2),

    PRIMARY KEY (organization_id, technology_id)
);

CREATE INDEX idx_org_tech_org ON "OrganizationTechnology"(organization_id);
CREATE INDEX idx_org_tech_tech ON "OrganizationTechnology"(technology_id);

-- ==============================================
-- FUNDING ROUND (Historical funding events)
-- ==============================================
CREATE TABLE "FundingRound" (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
    round_type      TEXT NOT NULL,            -- 'Pre-Seed', 'Seed', 'Series A', 'Series B', etc.
    amount          BIGINT,                   -- USD cents
    announced_date  DATE,
    lead_investor   TEXT,
    source_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_funding_round_org ON "FundingRound"(organization_id);
CREATE INDEX idx_funding_round_date ON "FundingRound"(announced_date DESC);

-- ==============================================
-- COMPANY SIGNAL (Intent signals: hiring, news, etc.)
-- ==============================================
CREATE TABLE "CompanySignal" (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES "Organization"(id) ON DELETE CASCADE,
    signal_type     TEXT NOT NULL,            -- 'hiring', 'expansion', 'product_launch', 'partnership', 'leadership_change', 'news'
    headline        TEXT NOT NULL,
    description     TEXT,
    signal_date     DATE,
    source          TEXT,                     -- 'linkedin', 'news', 'crunchbase', 'agent'
    source_url      TEXT,
    sentiment       TEXT,                     -- 'positive', 'neutral', 'negative'
    relevance_score NUMERIC(3,2),             -- 0.00 to 1.00
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_company_signal_org ON "CompanySignal"(organization_id);
CREATE INDEX idx_company_signal_date ON "CompanySignal"(signal_date DESC);
CREATE INDEX idx_company_signal_type ON "CompanySignal"(signal_type);
```

### Schema Modifications

```sql
-- ==============================================
-- ORGANIZATION: New firmographic columns
-- ==============================================
ALTER TABLE "Organization" ADD COLUMN revenue_range_id BIGINT REFERENCES "RevenueRange"(id) ON DELETE SET NULL;
ALTER TABLE "Organization" ADD COLUMN founding_year INTEGER;
ALTER TABLE "Organization" ADD COLUMN headquarters_city TEXT;
ALTER TABLE "Organization" ADD COLUMN headquarters_state TEXT;
ALTER TABLE "Organization" ADD COLUMN headquarters_country TEXT DEFAULT 'USA';
ALTER TABLE "Organization" ADD COLUMN company_type TEXT;      -- 'private', 'public', 'nonprofit', 'government'
ALTER TABLE "Organization" ADD COLUMN linkedin_url TEXT;
ALTER TABLE "Organization" ADD COLUMN crunchbase_url TEXT;

-- ==============================================
-- INDIVIDUAL: New contact intelligence columns
-- ==============================================
ALTER TABLE "Individual" ADD COLUMN twitter_url TEXT;
ALTER TABLE "Individual" ADD COLUMN github_url TEXT;
ALTER TABLE "Individual" ADD COLUMN bio TEXT;
ALTER TABLE "Individual" ADD COLUMN seniority_level TEXT;     -- 'C-Level', 'VP', 'Director', 'Manager', 'IC'
ALTER TABLE "Individual" ADD COLUMN department TEXT;          -- 'Engineering', 'Sales', 'Marketing', etc.
ALTER TABLE "Individual" ADD COLUMN is_decision_maker BOOLEAN DEFAULT FALSE;
ALTER TABLE "Individual" ADD COLUMN reports_to_id BIGINT REFERENCES "Individual"(id) ON DELETE SET NULL;
```

---

## API Endpoints

### Lookup Tables

```
GET /api/revenue-ranges
Response: { revenue_ranges: RevenueRange[] }

GET /api/technologies
Query: ?category=CRM
Response: { technologies: Technology[] }

POST /api/technologies
Body: { name: string, category: string, vendor?: string }
Response: { technology: Technology }
```

### Organization Firmographics

```
PATCH /api/organizations/:id
Body: {
  revenue_range_id?: number,
  founding_year?: number,
  headquarters_city?: string,
  headquarters_state?: string,
  headquarters_country?: string,
  company_type?: string,
  linkedin_url?: string,
  crunchbase_url?: string
}
Response: { organization: Organization }
```

### Tech Stack

```
GET /api/organizations/:id/technologies
Response: { technologies: OrganizationTechnology[] }

POST /api/organizations/:id/technologies
Body: { technology_id: number, source?: string, confidence?: number }
Response: { organization_technology: OrganizationTechnology }

DELETE /api/organizations/:id/technologies/:technology_id
Response: 204 No Content
```

### Funding Rounds

```
GET /api/organizations/:id/funding-rounds
Response: { funding_rounds: FundingRound[] }

POST /api/organizations/:id/funding-rounds
Body: { round_type: string, amount?: number, announced_date?: string, lead_investor?: string, source_url?: string }
Response: { funding_round: FundingRound }
```

### Company Signals

```
GET /api/organizations/:id/signals
Query: ?signal_type=hiring&limit=20
Response: { signals: CompanySignal[] }

POST /api/organizations/:id/signals
Body: { signal_type: string, headline: string, description?: string, signal_date?: string, source?: string, source_url?: string, sentiment?: string, relevance_score?: number }
Response: { signal: CompanySignal }
```

### Individual Contact Intelligence

```
PATCH /api/individuals/:id
Body: {
  twitter_url?: string,
  github_url?: string,
  bio?: string,
  seniority_level?: string,
  department?: string,
  is_decision_maker?: boolean,
  reports_to_id?: number
}
Response: { individual: Individual }

GET /api/individuals
Query: ?is_decision_maker=true&seniority_level=VP&department=Engineering
Response: { individuals: Individual[], pagination: ... }
```

### Data Provenance

```
GET /api/provenance/:entity_type/:entity_id
Query: ?field_name=revenue_range_id
Response: { provenance: DataProvenance[] }

POST /api/provenance
Body: { entity_type: string, entity_id: number, field_name: string, source: string, source_url?: string, confidence?: number, raw_value?: object }
Response: { provenance: DataProvenance }
```

---

## TypeScript Types

```typescript
// ==============================================
// Lookup Tables
// ==============================================

export interface RevenueRange {
  id: number;
  label: string;
  min_value: number | null;
  max_value: number | null;
  display_order: number;
}

export interface Technology {
  id: number;
  name: string;
  category: string;
  vendor: string | null;
}

// ==============================================
// Organization Extensions
// ==============================================

export interface OrganizationFirmographics {
  revenue_range_id: number | null;
  revenue_range?: RevenueRange;
  founding_year: number | null;
  headquarters_city: string | null;
  headquarters_state: string | null;
  headquarters_country: string | null;
  company_type: 'private' | 'public' | 'nonprofit' | 'government' | null;
  linkedin_url: string | null;
  crunchbase_url: string | null;
}

export interface OrganizationTechnology {
  organization_id: number;
  technology_id: number;
  technology?: Technology;
  detected_at: string;
  source: string | null;
  confidence: number | null;
}

export interface FundingRound {
  id: number;
  organization_id: number;
  round_type: string;
  amount: number | null;
  announced_date: string | null;
  lead_investor: string | null;
  source_url: string | null;
  created_at: string;
}

export interface CompanySignal {
  id: number;
  organization_id: number;
  signal_type: 'hiring' | 'expansion' | 'product_launch' | 'partnership' | 'leadership_change' | 'news';
  headline: string;
  description: string | null;
  signal_date: string | null;
  source: string | null;
  source_url: string | null;
  sentiment: 'positive' | 'neutral' | 'negative' | null;
  relevance_score: number | null;
  created_at: string;
}

// ==============================================
// Individual Extensions
// ==============================================

export interface IndividualContactIntelligence {
  twitter_url: string | null;
  github_url: string | null;
  bio: string | null;
  seniority_level: 'C-Level' | 'VP' | 'Director' | 'Manager' | 'IC' | null;
  department: string | null;
  is_decision_maker: boolean;
  reports_to_id: number | null;
  reports_to?: Individual;
}

// ==============================================
// Data Provenance
// ==============================================

export interface DataProvenance {
  id: number;
  entity_type: 'Organization' | 'Individual';
  entity_id: number;
  field_name: string;
  source: string;
  source_url: string | null;
  confidence: number | null;
  verified_at: string;
  verified_by: number | null;
  raw_value: Record<string, unknown> | null;
  created_at: string;
}

// ==============================================
// Extended Organization Type
// ==============================================

export interface OrganizationWithResearch extends Organization, OrganizationFirmographics {
  technologies?: OrganizationTechnology[];
  funding_rounds?: FundingRound[];
  signals?: CompanySignal[];
}

// ==============================================
// Extended Individual Type
// ==============================================

export interface IndividualWithResearch extends Individual, IndividualContactIntelligence {
  direct_reports?: Individual[];
}
```

---

## UI Mockup Notes

### Organization Detail Page

```
+--------------------------------------------------+
| [Logo] Acme Corp                    [Edit] [...]  |
+--------------------------------------------------+
| FIRMOGRAPHICS                                     |
| +--------------+  +--------------+  +----------+ |
| | Revenue      |  | Employees    |  | Founded  | |
| | $10M-$50M    |  | 51-200       |  | 2015     | |
| | [i] 85%      |  | [i] 90%      |  | [i] 95%  | |
| +--------------+  +--------------+  +----------+ |
|                                                   |
| Location: San Francisco, CA, USA                  |
| Type: Private                                     |
| LinkedIn | Crunchbase                             |
+--------------------------------------------------+
| TECH STACK                           [+ Add]     |
| +--------+  +--------+  +--------+  +--------+   |
| | HubSpot|  | AWS    |  | React  |  | Postgres|  |
| | CRM    |  | Cloud  |  | Frontend| | Database|  |
| +--------+  +--------+  +--------+  +--------+   |
+--------------------------------------------------+
| SIGNALS                              [View All]  |
| +----------------------------------------------+ |
| | [Hiring] +15 engineering roles   2024-11-20 | |
| | [Funding] Series B - $25M        2024-10-05 | |
| | [News] Expanded to Europe        2024-09-15 | |
| +----------------------------------------------+ |
+--------------------------------------------------+

[i] = Provenance tooltip showing source, confidence, verified_at
```

### Individual Detail Page

```
+--------------------------------------------------+
| [Avatar] Jane Smith               [Edit] [...]    |
| VP of Engineering @ Acme Corp                     |
+--------------------------------------------------+
| CONTACT INTELLIGENCE                              |
| Seniority: VP          Department: Engineering   |
| Decision Maker: Yes                               |
| Reports To: John CEO                              |
|                                                   |
| Email: jane@acme.com                              |
| Phone: +1 555-1234                                |
| LinkedIn | Twitter | GitHub                       |
+--------------------------------------------------+
| BIO                                               |
| 15+ years in engineering leadership. Previously  |
| at Google and Meta. Stanford CS.                 |
+--------------------------------------------------+
| DIRECT REPORTS                        [View All] |
| +----------------------------------------------+ |
| | Bob Dev - Senior Engineer                    | |
| | Alice Eng - Staff Engineer                   | |
| +----------------------------------------------+ |
+--------------------------------------------------+
```

### List Page Filters

```
Individuals List:
[Seniority ▼] [Department ▼] [Decision Maker ☑] [Search...]

Organizations List:
[Revenue ▼] [Employees ▼] [Tech Stack ▼] [Signals ▼] [Search...]
```

---

## Verification Checklist

### Functional Requirements
- [ ] Organizations display firmographic data (revenue, size, location, founding year)
- [ ] Organizations display tech stack with category grouping
- [ ] Organizations display funding history and company signals
- [ ] Individuals display contact intelligence (seniority, department, decision-maker flag)
- [ ] Individuals display org chart (reports_to relationship)
- [ ] Provenance tooltips show source, confidence, and verified_at for enriched fields
- [ ] List pages support filtering by new fields
- [ ] AI agent can create provenance records when enriching data

### Non-Functional Requirements
- [ ] Performance: Organization detail page loads in < 500ms
- [ ] Security: Provenance records respect tenant isolation
- [ ] Data integrity: Provenance verified_at updates when field changes

### Edge Cases
- [ ] Organization with no tech stack shows empty state
- [ ] Individual with no reports_to shows "Not set"
- [ ] Provenance for field with multiple sources shows most recent
- [ ] Deleting an organization cascades to tech stack, funding, signals

---

## Implementation Notes

### Migration Sequence
1. `021_revenue_range.sql` - RevenueRange lookup table with seed data
2. `022_technology.sql` - Technology + OrganizationTechnology tables
3. `023_funding_signals.sql` - FundingRound + CompanySignal tables
4. `024_data_provenance.sql` - DataProvenance polymorphic table
5. `025_organization_firmographics.sql` - New Organization columns
6. `026_individual_contact_intel.sql` - New Individual columns

### Files to Modify
- `modules/agent/migrations/` - New migration files
- `modules/agent/src/models/Organization.py` - Add firmographic fields and relationships
- `modules/agent/src/models/Individual.py` - Add contact intelligence fields
- `modules/agent/src/models/__init__.py` - Export new models
- `modules/client/types/types.ts` - Add TypeScript interfaces
- `modules/client/app/accounts/organizations/` - Update UI components
- `modules/client/app/accounts/individuals/` - Update UI components

### Dependencies
- Existing `update_updated_at_column()` trigger function (from 001_initial_schema.sql)
- Existing polymorphic patterns (EntityTag, Note)

### Out of Scope
- Automated data enrichment (AI agent tools) - separate feature
- Integration with external APIs (Clearbit, Apollo) - separate feature
- ICP scoring algorithm - separate feature
- Data export/import - separate feature

Type

Feature / Tech Debt

Priority

Medium

Description

Replace Google Sheets job application tracking with Supabase database to enable:

React UI for managing job applications directly in the website

Better data structure and querying capabilities

Real-time updates between backend agent and frontend UI

Maintained free tier with better scalability

User Story

As a job seeker using the agent
I want to view and manage my job applications in a UI on the website
So that I can track applications without leaving the site and the agent can filter duplicates more reliably

Current State

Job applications tracked in Google Sheet (1booMeM9GMAbrF6DgV9hsKo5_dxwYcD5kr049kAULAIA)

Python backend reads sheet via gspread API

Agent filters job search results based on sheet data

No UI for viewing/managing applications

Proposed Solution

Tech Stack

Database: Supabase (Postgres) - FREE tier

Backend: Supabase Python client (supabase-py)

Frontend: Supabase JavaScript client in React

Architecture

Agent (Python) → Supabase Python Client → Supabase DB
↓
React UI → Supabase JS Client → Supabase DB (real-time sync)

Acceptance Criteria

Backend (Python Agent)

[ ] Supabase database created with applied_jobs table

[ ] Python service replaces google_sheets.py with Supabase integration

[ ] Agent can read applied jobs from Supabase

[ ] Agent can write new applications to Supabase (optional enhancement)

[ ] Job search filtering works identically to Google Sheets version

[ ] Graceful fallback if Supabase is unavailable

[ ] Environment variables configured for Supabase credentials

POSSIBLE: see if migrations can first be stored in a /supabase_migrations folder

Frontend (React)

[ ] New React component for job tracker UI

[ ] Display list of applied jobs (company, title, date, status, notes)

[ ] Add new job application via form

[ ] Edit/delete existing applications

[ ] Real-time updates when agent adds jobs

[ ] Responsive design (mobile-friendly)

[ ] Protected route (optional: basic auth)

General

[ ] Migration script to import existing Google Sheets data

[ ] Documentation updated (replace GOOGLE_SHEETS_SETUP.md with SUPABASE_SETUP.md)

[ ] No regressions in job search filtering

[ ] Free tier limits acceptable for use case

Technical Implementation Tasks

Phase 1: Backend Integration (Python Agent)

1.1 Supabase Setup

[ ] Create Supabase account and project

[ ] Design applied_jobs table schema:

CREATE TABLE applied_jobs (
id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
company TEXT NOT NULL,
job_title TEXT NOT NULL,
application_date TIMESTAMP DEFAULT NOW(),
status TEXT DEFAULT 'applied',
job_url TEXT,
notes TEXT,
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_company_title ON applied_jobs(company, job_title);

Configure Row Level Security (RLS) policies

Get API keys (anon/public key, service role key)

1.2 Python Backend Changes

Add supabase to pyproject.toml dependencies

Create /modules/agent/src/agent/services/supabase_client.py:

AppliedJobsTracker class using Supabase

fetch_applied_jobs() method

is_job_applied() method

filter_jobs() method

add_applied_job() method (bonus)

Update /modules/agent/src/agent/tools/worker_tools.py:

Replace google_sheets import with supabase_client

Update job_search_with_filter() to use Supabase

Update .env.example:
SUPABASE_URL=""
SUPABASE_SERVICE_KEY=""

Remove Google Sheets dependencies from pyproject.toml (cleanup)

1.3 Migration Script

Create /modules/agent/scripts/migrate_sheets_to_supabase.py:

Read all rows from Google Sheet

Insert into Supabase applied_jobs table

Handle duplicates

Log migration results

Phase 2: Frontend Integration (React)

2.1 Client Setup

Install Supabase JS client: npm install @supabase/supabase-js

Create Supabase client config in /modules/client/lib/supabase.ts:
import { createClient } from '@supabase/supabase-js'

export const supabase = createClient(
process.env.NEXT_PUBLIC_SUPABASE_URL!,
process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
)

Add to .env.local:
NEXT_PUBLIC_SUPABASE_URL=""
NEXT_PUBLIC_SUPABASE_ANON_KEY=""

2.2 Job Tracker UI Component

Create /modules/client/app/jobs/page.tsx (new route)

Create /modules/client/components/JobTracker/JobTracker.tsx:

Fetch jobs on mount

Subscribe to real-time updates

Display jobs in table/card layout

Create /modules/client/components/JobTracker/JobForm.tsx:

Form for adding new applications

Validation (company, title required)

Create /modules/client/components/JobTracker/JobRow.tsx:

Display individual job

Edit/delete actions

Status badge (applied, interviewing, rejected, etc.)

Add link to job tracker in Header nav

2.3 Real-time Subscriptions

Subscribe to applied_jobs table changes

Auto-update UI when agent adds jobs

Handle connection status (online/offline indicator)

Phase 3: Testing & Documentation

3.1 Testing

Test Python backend reads applied jobs correctly

Test job filtering excludes Supabase jobs

Test React UI CRUD operations

Test real-time sync between agent and UI

Test migration script with production data

Test free tier limits (should be fine)

Load testing: simulate 100 job applications

3.2 Documentation

Create SUPABASE_SETUP.md with:

Account setup instructions

Table schema

RLS policy configuration

Environment variable setup

Migration instructions

Update main README with new job tracker features

Archive GOOGLE_SHEETS_SETUP.md

Add inline code comments

3.3 Deployment

Update Dockerfile to include Supabase dependencies

Add Supabase env vars to production environment

Run migration script in production

Deploy backend changes

Deploy frontend changes

Verify job search works in production

Database Schema

-- Table: applied_jobs
CREATE TABLE applied_jobs (
id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
company TEXT NOT NULL,
job_title TEXT NOT NULL,
application_date TIMESTAMP DEFAULT NOW(),
status TEXT DEFAULT 'applied' CHECK (status IN ('applied', 'interviewing', 'rejected', 'offer', 'accepted')),
job_url TEXT,
location TEXT,
salary_range TEXT,
notes TEXT,
source TEXT DEFAULT 'manual' CHECK (source IN ('manual', 'agent')),
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_company_title ON applied_jobs(company, job_title);
CREATE INDEX idx_application_date ON applied_jobs(application_date DESC);
CREATE INDEX idx_status ON applied_jobs(status);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
NEW.updated_at = NOW();
RETURN NEW;
END;

$$
LANGUAGE plpgsql;

  CREATE TRIGGER update_applied_jobs_updated_at
  BEFORE UPDATE ON applied_jobs
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

  Files to Create/Modify

  New Files

/modules/agent/src/agent/services/supabase_client.py

/modules/agent/scripts/migrate_sheets_to_supabase.py

/modules/agent/SUPABASE_SETUP.md

/modules/client/lib/supabase.ts

/modules/client/app/jobs/page.tsx

/modules/client/components/JobTracker/JobTracker.tsx

/modules/client/components/JobTracker/JobForm.tsx

/modules/client/components/JobTracker/JobRow.tsx

  Modified Files

/modules/agent/pyproject.toml (add supabase, remove gspread)

/modules/agent/src/agent/tools/worker_tools.py

/modules/agent/.env.example

/modules/client/.env.local.example

/modules/client/components/Header.tsx (add Jobs link)

/modules/client/package.json (add @supabase/supabase-js)

  Archived Files

/modules/agent/src/agent/services/google_sheets.py → delete or archive

/modules/agent/GOOGLE_SHEETS_SETUP.md → archive

  Dependencies

  Python

Remove:

gspread>=6.1.0

google-auth>=2.34.0

Add:

  supabase>=2.0.0

  JavaScript/TypeScript

  {
  "dependencies": {
    "@supabase/supabase-js": "^2.39.0"
  }
}

  Estimated Effort

Backend (Python): 4-6 hours

Frontend (React): 8-12 hours

Migration & Testing: 2-4 hours

Documentation: 2 hours

Total: ~16-24 hours (2-3 days)

  Risk Assessment

  Risks

Free tier limits - Unlikely but monitor usage

Real-time sync latency - Should be <1s, acceptable

Migration data loss - Mitigate with backup and dry-run

Breaking changes during migration - Deploy backend first, test, then frontend

  Mitigation

Keep Google Sheets as read-only backup during transition period

Feature flag for Supabase vs Google Sheets (fallback)

Comprehensive testing before production deployment

  Success Metrics

Agent filters jobs correctly (0 false positives/negatives)

React UI loads in <2s

Real-time updates appear in <5s

Migration completes with 100% data accuracy

Stays within Supabase free tier limits

  Future Enhancements (Out of Scope)

Agent automatically adds jobs when user applies via chat

Email notifications for job status changes

Analytics dashboard (applications over time, success rate)

Chrome extension to add jobs from LinkedIn/Indeed

Export to PDF resume tracker

  Notes

Supabase free tier: 500 MB storage, unlimited API requests (with rate limits)

RLS policies ensure data security even with public anon key

Consider adding basic authentication to /jobs route in future


$$

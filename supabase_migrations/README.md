# Supabase Migrations

This directory contains SQL migration files for the Supabase database schema.

## Migration Files

- **001_initial_schema.sql**: Initial database schema for job application tracking
  - Creates `applied_jobs` table with proper types and constraints
  - Adds indexes for query performance
  - Sets up automatic `updated_at` trigger
  - Enables Row Level Security (RLS) with permissive policies

## Running Migrations

### Automatic (Docker)

Migrations run automatically when the agent Docker container starts:

```bash
docker-compose up
```

See `DOCKER_MIGRATIONS.md` for details.

### Manual Options

#### Option A: Python Script

```bash
cd modules/agent
python scripts/apply_migrations.py
```

Requires:
- `SUPABASE_URL` environment variable
- `SUPABASE_SERVICE_KEY` environment variable
- `SUPABASE_DB_PASSWORD` environment variable (for psycopg2)
- `psycopg2-binary` installed

#### Option B: Supabase Dashboard

1. Navigate to **SQL Editor** in Supabase dashboard
2. Copy migration file contents
3. Paste and run

#### Option C: Supabase CLI

```bash
supabase login
supabase link --project-ref your-project-ref
supabase db push
```

## Migration Guidelines

### Naming Convention

Use sequential numbers with descriptive names:
- `001_initial_schema.sql`
- `002_add_user_preferences.sql`
- `003_create_notifications_table.sql`

### Writing Migrations

**Always use idempotent SQL:**

```sql
-- ✅ Good: Can run multiple times safely
CREATE TABLE IF NOT EXISTS my_table (...);
ALTER TABLE my_table ADD COLUMN IF NOT EXISTS new_field TEXT;
CREATE INDEX IF NOT EXISTS idx_name ON my_table(field);

-- ❌ Bad: Fails on second run
CREATE TABLE my_table (...);
ALTER TABLE my_table ADD COLUMN new_field TEXT;
```

**Include comments:**

```sql
-- Migration: Add email notifications
-- Date: 2024-11-10
-- Author: Team

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS email_notifications BOOLEAN DEFAULT true;
```

### Testing Migrations

Always test locally before deploying:

```bash
# Dry run (shows SQL without executing)
python modules/agent/scripts/apply_migrations.py --dry-run

# Apply migrations
python modules/agent/scripts/apply_migrations.py
```

## Rollback Strategy

We use **forward-only migrations** (no down migrations).

To rollback:
1. Create a new migration that reverses the change
2. Apply it like any other migration

Example:
```sql
-- 003_rollback_notifications.sql
ALTER TABLE users DROP COLUMN IF EXISTS email_notifications;
```

This approach:
- Matches git workflow (revert commits)
- Provides clear audit trail
- Avoids migration drift
- Simpler mental model

## Database Schema

### applied_jobs Table

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key (auto-generated) |
| company | TEXT | Company name (required) |
| job_title | TEXT | Job title (required) |
| application_date | TIMESTAMP | When applied (default: now) |
| status | TEXT | Application status (applied, interviewing, rejected, offer, accepted) |
| job_url | TEXT | Link to job posting |
| location | TEXT | Job location |
| salary_range | TEXT | Salary information |
| notes | TEXT | Additional notes |
| source | TEXT | manual or agent |
| created_at | TIMESTAMP | Record creation time |
| updated_at | TIMESTAMP | Last update time (auto-updated) |

### Indexes

- `idx_company_title`: Fast lookups by company + job title
- `idx_application_date`: Sorted by application date
- `idx_status`: Filter by status

### Row Level Security (RLS)

Currently uses permissive policies (suitable for single user).

For production with multiple users:
1. Implement Supabase Auth
2. Add user_id column to applied_jobs
3. Update RLS policies to restrict by user_id

## Troubleshooting

### Migration Script Not Found

Ensure you're running from the correct directory:
```bash
cd /home/hubenschmidt/pina-colada-co
python modules/agent/scripts/apply_migrations.py
```

### psycopg2 Import Error

Install the dependency:
```bash
pip install psycopg2-binary
```

### Connection Timeout

1. Check internet connection
2. Verify Supabase project is active (free tier pauses after inactivity)
3. Confirm credentials are correct

### SQL Syntax Errors

1. Test migration SQL in Supabase dashboard first
2. Check PostgreSQL version compatibility
3. Review migration file for typos

## Related Documentation

- `/SUPABASE_SETUP.md` - Complete setup guide
- `/DOCKER_MIGRATIONS.md` - Docker integration details
- `modules/agent/scripts/apply_migrations.py` - Migration script

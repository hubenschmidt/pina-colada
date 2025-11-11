# Supabase Migrations

This directory contains SQL migration files for the Supabase database schema.

## Migration Files

- `001_initial_schema.sql`: Creates the initial `applied_jobs` table with indexes, triggers, and RLS policies

## Running Migrations

### Option 1: Using Supabase Dashboard (Manual)

1. Log in to your Supabase project dashboard
2. Navigate to the SQL Editor
3. Copy the contents of each migration file (in order)
4. Execute the SQL in the editor

### Option 2: Using Supabase CLI (Recommended)

If you have the Supabase CLI installed:

```bash
# Initialize Supabase in your project (first time only)
supabase init

# Link to your remote project
supabase link --project-ref YOUR_PROJECT_REF

# Apply migrations
supabase db push
```

### Option 3: Using Python Migration Script

We've provided a Python script that can automatically apply migrations:

```bash
cd modules/agent
python scripts/apply_migrations.py
```

This script will:
- Read all SQL files from the `supabase_migrations` directory
- Execute them in order against your Supabase database
- Track which migrations have been applied

## Migration Naming Convention

Migrations should follow the pattern: `{number}_{description}.sql`

Examples:
- `001_initial_schema.sql`
- `002_add_contact_info.sql`
- `003_add_resume_link.sql`

## Creating New Migrations

1. Create a new file with the next sequential number
2. Write your SQL DDL statements
3. Test locally if possible
4. Apply to production using one of the methods above

## Important Notes

- Always backup your database before running migrations in production
- Migrations should be idempotent (safe to run multiple times)
- Use `IF NOT EXISTS` and `IF EXISTS` clauses where appropriate
- Test migrations on a staging environment first

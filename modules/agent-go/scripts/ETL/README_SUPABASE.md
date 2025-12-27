# Supabase Database Management

## Reset Database Schema (Test/Dev Only)

To completely reset the database and re-run all migrations:

### Step 1: Clear migration tracking

```sql
-- Clear Supabase migration tracking (so all migrations re-run)
DROP TABLE IF EXISTS supabase_migrations.schema_migrations;
DROP SCHEMA IF EXISTS supabase_migrations CASCADE;

-- Clear seeders tracking
DROP TABLE IF EXISTS public.schema_seeders;
```

### Step 2: Drop and recreate public schema

```sql
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;
```

### Step 3: Re-run migrations

Trigger the CI/CD pipeline by pushing to the appropriate branch, or run locally:

```bash
supabase db push --db-url "postgres://postgres.[PROJECT_ID]:[PASSWORD]@aws-1-us-east-1.pooler.supabase.com:5432/postgres"
```

## Table-Level Grants for Supabase Roles

After migrations create tables, grant permissions to Supabase API roles:

```sql
-- Source: https://stackoverflow.com/a/75869433
-- License: CC BY-SA 4.0

GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA "public" TO authenticated;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA "public" TO anon;
```

These grants allow the Supabase client API (`anon` key and `authenticated` users) to access tables.

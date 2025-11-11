# Database Migration Strategy

## Overview

This project uses a **hybrid approach** for database migrations:
- **Local Development**: Local PostgreSQL (via Docker Compose)
- **Production**: Supabase (managed PostgreSQL)

Migrations are applied using **Supabase CLI** in the CI/CD pipeline.

## Architecture

```
┌─────────────────┐         ┌──────────────────┐
│  Local Dev      │         │  Production      │
│                 │         │                  │
│  Docker Compose │         │  Supabase        │
│  └─ Postgres    │         │  └─ Managed DB   │
│                 │         │                  │
│  Migrations:    │         │  Migrations:     │
│  Manual/CLI     │         │  Azure Pipeline │
└─────────────────┘         └──────────────────┘
```

## Local Development

### Setup

1. **Start local Postgres** (already in `docker-compose.yml`):
   ```bash
   docker-compose up postgres
   ```

2. **Create database for agent** (already done via init script):
   ```bash
   # Database is created automatically via docker/postgres-init/01-create-agent-db.sql
   # Or manually:
   docker-compose exec postgres psql -U postgres -c "CREATE DATABASE pina_colada;"
   ```

3. **Run migrations manually**:
   ```bash
   # Option 1: Using psql
   docker-compose exec postgres psql -U postgres -d pina_colada -f /path/to/supabase_migrations/001_initial_schema.sql
   
   # Option 2: Using Supabase CLI (if you want to test against Supabase)
   supabase db push
   ```

### Configuration

In `modules/agent/.env` for local development:
```bash
# Use local Postgres (set USE_SUPABASE=false or don't set SUPABASE_URL)
USE_SUPABASE=false

# Or use Supabase for local dev (optional)
# USE_SUPABASE=true
# SUPABASE_URL="https://your-project.supabase.co"
# SUPABASE_SERVICE_KEY="your-key"
```

## Production Migrations

### Azure Pipeline Integration

Migrations run automatically in the Azure Pipeline **before deployment**:

1. **Install Supabase CLI** in pipeline
2. **Link to Supabase project** using access token
3. **Push migrations** using `supabase db push`
4. **Deploy application** (only if migrations succeed)

### Required Azure Pipeline Variables

Add these to your Azure DevOps variable group:

- `SUPABASE_ACCESS_TOKEN`: Your Supabase access token
  - Get from: https://supabase.com/dashboard/account/tokens
- `SUPABASE_PROJECT_REF`: Your project reference (e.g., `befhobypxfjhdzgirfml`)
  - Found in your Supabase project URL

### Pipeline Flow

```
Build → Push Image → Migrate (Supabase CLI) → Deploy
                         ↓
                    If migrations fail, deployment is blocked
```

## Migration Files

- **Location**: `/supabase_migrations/`
- **Naming**: `001_initial_schema.sql`, `002_add_feature.sql`, etc.
- **Format**: Standard PostgreSQL SQL
- **Idempotent**: Use `IF NOT EXISTS` / `IF EXISTS` clauses

## Benefits

✅ **No Docker IPv6 issues** - Migrations run from pipeline, not Docker  
✅ **Official Supabase tooling** - Uses supported `supabase` CLI  
✅ **Fast local development** - Local Postgres, no network dependency  
✅ **Safe production deploys** - Migrations run before deployment  
✅ **Version controlled** - All migrations in git  

## Troubleshooting

### Local Migrations Fail

- Check Postgres is running: `docker-compose ps postgres`
- Verify database exists: `docker-compose exec postgres psql -U postgres -l`
- Check connection: `docker-compose exec postgres psql -U postgres -d pina_colada`

### Production Migrations Fail

- Check Azure Pipeline logs for Supabase CLI errors
- Verify `SUPABASE_ACCESS_TOKEN` is valid
- Verify `SUPABASE_PROJECT_REF` matches your project
- Check Supabase Dashboard → Database → Migrations for status

### Migration Conflicts

- Supabase tracks applied migrations
- If a migration was applied manually, it won't run again
- To reset: Use Supabase Dashboard → Database → Reset (⚠️ destructive)

## Manual Migration (Emergency)

If pipeline migrations fail, you can run manually:

```bash
# Install CLI
npm install -g supabase

# Login
supabase login

# Link project
supabase link --project-ref YOUR_PROJECT_REF

# Push migrations
supabase db push
```

Or use Supabase Dashboard → SQL Editor to run SQL directly.


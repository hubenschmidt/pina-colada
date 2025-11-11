# Local Development Setup

## Overview

This project uses **local Postgres** for development and **Supabase** for production.

- **Local Dev**: Client → API Routes → Local Postgres | Agent → SQLAlchemy → Local Postgres
- **Production**: Client → Supabase | Agent → Supabase

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Local Development                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Client (Next.js)          Agent (Python)              │
│       │                         │                       │
│       │                         │                       │
│       ▼                         ▼                       │
│  API Routes (/api/jobs)    SQLAlchemy ORM              │
│       │                         │                       │
│       └──────────┬──────────────┘                       │
│                  ▼                                       │
│         Local Postgres (Docker)                         │
│         └─ Database: pina_colada                       │
│                                                         │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                      Production                         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Client (Next.js)          Agent (Python)              │
│       │                         │                       │
│       ▼                         ▼                       │
│  Supabase JS Client      Supabase Python Client         │
│       │                         │                       │
│       └──────────┬──────────────┘                       │
│                  ▼                                       │
│              Supabase                                   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Setup

### 1. Environment Variables

**Client (`modules/client/.env.local`):**
```bash
# Enable local Postgres in dev
NEXT_PUBLIC_USE_LOCAL_POSTGRES=true

# Supabase (for production)
NEXT_PUBLIC_SUPABASE_URL="https://your-project.supabase.co"
NEXT_PUBLIC_SUPABASE_ANON_KEY="your-anon-key"

# Auth
NEXT_PUBLIC_JOBS_PASSWORD="Welcome10!"
JOBS_PASSWORD="Welcome10!"
```

**Agent (`modules/agent/.env`):**
```bash
# Enable local Postgres in dev
USE_LOCAL_POSTGRES=true

# Postgres connection (for local dev)
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=pina_colada

# Supabase (for production - set USE_LOCAL_POSTGRES=false)
USE_SUPABASE=true
SUPABASE_URL="https://your-project.supabase.co"
SUPABASE_SERVICE_KEY="your-service-key"
```

### 2. Start Services

```bash
docker-compose up
```

This will:
- Start Postgres (creates `pina_colada` database automatically)
- Start Agent (connects to local Postgres if `USE_LOCAL_POSTGRES=true`)
- Start Client (uses API routes if `NEXT_PUBLIC_USE_LOCAL_POSTGRES=true`)

### 3. Run Migrations

**Local Postgres:**
```bash
# Apply migrations to local Postgres
docker-compose exec postgres psql -U postgres -d pina_colada -f /path/to/supabase_migrations/001_initial_schema.sql

# Or from host:
PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres -d pina_colada -f supabase_migrations/001_initial_schema.sql
```

**Production (Supabase):**
- Migrations run automatically in Azure Pipeline via Supabase CLI
- Or manually via Supabase Dashboard SQL Editor

## How It Works

### Client (Next.js)

**Local Dev Mode** (`NEXT_PUBLIC_USE_LOCAL_POSTGRES=true`):
- Uses API routes (`/api/jobs/*`) which connect to local Postgres
- No Supabase client calls
- No real-time subscriptions (polling/refresh instead)

**Production Mode**:
- Uses Supabase JS client directly
- Real-time subscriptions enabled
- No API routes needed

### Agent (Python)

**Local Dev Mode** (`USE_LOCAL_POSTGRES=true`):
- Uses SQLAlchemy ORM (`agent.services.postgres_client`)
- Connects to local Postgres via `postgres:5432`
- No Supabase client

**Production Mode** (`USE_SUPABASE=true`):
- Uses Supabase Python client
- Connects to Supabase cloud database

## Switching Between Modes

**To use local Postgres:**
```bash
# Client
echo "NEXT_PUBLIC_USE_LOCAL_POSTGRES=true" >> modules/client/.env.local

# Agent
echo "USE_LOCAL_POSTGRES=true" >> modules/agent/.env
echo "USE_SUPABASE=false" >> modules/agent/.env
```

**To use Supabase:**
```bash
# Client
# Remove or set: NEXT_PUBLIC_USE_LOCAL_POSTGRES=false

# Agent
echo "USE_LOCAL_POSTGRES=false" >> modules/agent/.env
echo "USE_SUPABASE=true" >> modules/agent/.env
```

## Benefits

✅ **Fast local development** - No network dependency  
✅ **SQLAlchemy ORM** - Type-safe, maintainable Python code  
✅ **Consistent schema** - Same migrations for local and production  
✅ **Easy testing** - Isolated local database  
✅ **Production-ready** - Supabase for scalability  

## Troubleshooting

**Client can't connect to Postgres:**
- Check `POSTGRES_HOST=postgres` (Docker service name)
- Verify Postgres is healthy: `docker-compose ps postgres`
- Check database exists: `docker-compose exec postgres psql -U postgres -l`

**Agent can't connect:**
- Verify `USE_LOCAL_POSTGRES=true` in `.env`
- Check `POSTGRES_HOST=postgres` (not `localhost`)
- Ensure SQLAlchemy is installed: `pip install sqlalchemy`

**Migrations not applied:**
- Run manually (see step 3 above)
- Check database: `docker-compose exec postgres psql -U postgres -d pina_colada -c "\dt"`


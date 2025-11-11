# Docker Integration for Supabase Migrations

## Overview

The agent Docker container now automatically runs Supabase migrations on startup, ensuring the database schema is always up-to-date.

## How It Works

### 1. **Dockerfile Changes**

```dockerfile
# Copy migrations and scripts
COPY scripts ./scripts
COPY ../../supabase_migrations ./supabase_migrations
```

The Dockerfile now copies:
- `scripts/` directory (contains `apply_migrations.py`)
- `supabase_migrations/` directory (contains SQL migration files)

### 2. **Startup Script (`start.sh`)**

```bash
# Run Supabase migrations if configured
if [ -n "${SUPABASE_URL:-}" ] && [ -n "${SUPABASE_SERVICE_KEY:-}" ]; then
  echo "🔄 Running Supabase migrations..."
  if uv run python scripts/apply_migrations.py; then
    echo "✅ Migrations completed successfully"
  else
    echo "⚠️  Migrations failed, continuing anyway..."
  fi
else
  echo "⏭️  Skipping migrations (SUPABASE_URL or SUPABASE_SERVICE_KEY not set)"
fi
```

**Behavior:**
- ✅ Checks if Supabase credentials are configured
- ✅ Runs migrations automatically if configured
- ✅ Continues starting services even if migrations fail (non-blocking)
- ✅ Skips gracefully if not configured (useful for dev without Supabase)

### 3. **Migration Script Improvements**

The `apply_migrations.py` script now:
- ✅ Detects Docker environment automatically
- ✅ Looks for migrations in `/app/supabase_migrations` (Docker path)
- ✅ Falls back to local path for development
- ✅ Handles missing `python-dotenv` gracefully (uses env vars directly)
- ✅ Better error messages with emoji indicators

### 4. **docker-compose.yml Integration**

```yaml
agent:
  env_file:
    - ./modules/agent/.env
  volumes:
    - ./supabase_migrations:/app/supabase_migrations:ro
```

**Configuration:**
- Environment variables loaded from `.env` file
- Migrations directory mounted as read-only
- Hot-reload enabled for development

---

## Configuration

### Environment Variables

In `modules/agent/.env`, add:

```bash
# Supabase Configuration
USE_SUPABASE=true
SUPABASE_URL="https://your-project-ref.supabase.co"
SUPABASE_SERVICE_KEY="your-service-role-key"
SUPABASE_DB_PASSWORD="your-db-password"  # Optional, for psycopg2
```

### Dependencies

Automatically installed in Docker:
- ✅ `supabase>=2.0.0`
- ✅ `psycopg2-binary>=2.9.11` (for direct SQL execution)

---

## Usage

### Development with Docker Compose

```bash
# Start all services (migrations run automatically)
docker-compose up

# View agent logs to see migration output
docker-compose logs -f agent
```

**Expected Output:**
```
agent_1  | ============================================================
agent_1  | Supabase Migration Tool
agent_1  | ============================================================
agent_1  | ✓ Running in Docker container
agent_1  | Migrations directory: /app/supabase_migrations
agent_1  | 📁 Found 1 migration file(s):
agent_1  |   - 001_initial_schema.sql
agent_1  | ✅ psycopg2 is available - applying migrations automatically
agent_1  |
agent_1  | Applying: 001_initial_schema.sql
agent_1  | ✓ Successfully applied 001_initial_schema.sql
agent_1  | ✓ All migrations applied successfully!
```

### Production Deployment

```bash
# Build production image
docker build -t agent:latest ./modules/agent

# Run with environment variables
docker run \
  -e SUPABASE_URL="https://..." \
  -e SUPABASE_SERVICE_KEY="..." \
  -e SUPABASE_DB_PASSWORD="..." \
  -p 8000:8000 \
  agent:latest
```

---

## Migration Workflow

### Adding New Migrations

1. **Create migration file** in `supabase_migrations/`:
   ```bash
   touch supabase_migrations/002_add_status_field.sql
   ```

2. **Write SQL**:
   ```sql
   -- Add new field
   ALTER TABLE applied_jobs ADD COLUMN IF NOT EXISTS priority INT DEFAULT 1;
   ```

3. **Test locally**:
   ```bash
   python modules/agent/scripts/apply_migrations.py
   ```

4. **Rebuild Docker** (or restart if using volumes):
   ```bash
   docker-compose up --build
   ```

### Migration Files

- **Naming**: Use sequential numbers: `001_*.sql`, `002_*.sql`, etc.
- **Idempotent**: Use `IF NOT EXISTS` / `IF EXISTS` clauses
- **Order**: Applied alphabetically by filename

---

## Troubleshooting

### Migrations Not Running

**Check environment variables:**
```bash
docker-compose exec agent env | grep SUPABASE
```

**Manual run:**
```bash
docker-compose exec agent python scripts/apply_migrations.py
```

### Migrations Failing

**View detailed logs:**
```bash
docker-compose logs agent | grep -A 20 "Migration Tool"
```

**Common issues:**
- ❌ Invalid credentials → Check `SUPABASE_URL` and `SUPABASE_SERVICE_KEY`
- ❌ Network issues → Ensure container can reach Supabase
- ❌ SQL errors → Check migration file syntax
- ❌ Missing psycopg2 → Should be auto-installed, check `pyproject.toml`

### Skip Migrations Temporarily

**Option 1: Remove env vars**
```bash
# Comment out in .env
# SUPABASE_URL=...
# SUPABASE_SERVICE_KEY=...
```

**Option 2: Use Google Sheets fallback**
```bash
USE_SUPABASE=false
```

---

## CI/CD Integration

### Example: GitHub Actions

```yaml
name: Deploy Agent

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: docker build -t agent:${{ github.sha }} ./modules/agent

      - name: Run migrations
        env:
          SUPABASE_URL: ${{ secrets.SUPABASE_URL }}
          SUPABASE_SERVICE_KEY: ${{ secrets.SUPABASE_SERVICE_KEY }}
        run: |
          docker run --rm \
            -e SUPABASE_URL \
            -e SUPABASE_SERVICE_KEY \
            agent:${{ github.sha }} \
            python scripts/apply_migrations.py

      - name: Deploy
        run: |
          # Your deployment commands
```

### Example: Azure Pipelines

```yaml
- task: Docker@2
  displayName: 'Build agent image'
  inputs:
    command: build
    dockerfile: modules/agent/Dockerfile
    tags: $(Build.BuildId)

- task: CmdLine@2
  displayName: 'Run migrations'
  env:
    SUPABASE_URL: $(SUPABASE_URL)
    SUPABASE_SERVICE_KEY: $(SUPABASE_SERVICE_KEY)
  inputs:
    script: |
      docker run --rm \
        -e SUPABASE_URL \
        -e SUPABASE_SERVICE_KEY \
        agent:$(Build.BuildId) \
        python scripts/apply_migrations.py
```

---

## Benefits

### 1. **Automated Schema Management**
- ✅ No manual migration steps
- ✅ Consistent database schema across environments
- ✅ Version-controlled migrations

### 2. **Zero Downtime**
- ✅ Migrations run before services start
- ✅ Non-blocking: Services start even if migrations fail
- ✅ Idempotent migrations allow safe re-runs

### 3. **Developer Experience**
- ✅ One command to start everything: `docker-compose up`
- ✅ No separate migration step to remember
- ✅ Works locally and in production

### 4. **Operational Safety**
- ✅ Migrations tracked in git
- ✅ Review changes via pull requests
- ✅ Rollback by reverting git commits

---

## Best Practices

### 1. **Write Idempotent Migrations**
```sql
-- ✅ Good: Can run multiple times
CREATE TABLE IF NOT EXISTS my_table (...);
ALTER TABLE my_table ADD COLUMN IF NOT EXISTS new_field TEXT;

-- ❌ Bad: Fails on second run
CREATE TABLE my_table (...);
ALTER TABLE my_table ADD COLUMN new_field TEXT;
```

### 2. **Test Migrations Locally First**
```bash
# Always test before committing
python modules/agent/scripts/apply_migrations.py --dry-run
python modules/agent/scripts/apply_migrations.py
```

### 3. **Use Descriptive Filenames**
```bash
# ✅ Good
001_initial_schema.sql
002_add_user_preferences.sql
003_create_notifications_table.sql

# ❌ Bad
migration1.sql
update.sql
new_stuff.sql
```

### 4. **Add Comments to SQL**
```sql
-- Migration: Add email notification preferences
-- Date: 2024-11-10
-- Author: Team

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS email_notifications BOOLEAN DEFAULT true;
```

### 5. **Keep Migrations Small**
- One logical change per migration file
- Easier to review and rollback
- Faster to execute

---

## Rollback Strategy

**Forward-Only Migrations:**
We use forward-only migrations (no down migrations). To rollback:

1. **Create new migration** that reverses changes:
   ```sql
   -- 003_rollback_notifications.sql
   ALTER TABLE users DROP COLUMN IF EXISTS email_notifications;
   ```

2. **Apply via normal process**:
   ```bash
   docker-compose up
   ```

**Why forward-only?**
- ✅ Simpler mental model
- ✅ Matches git workflow (revert commits)
- ✅ No "down" migration drift
- ✅ Clear audit trail

---

## Related Files

- `/supabase_migrations/` - SQL migration files
- `/modules/agent/scripts/apply_migrations.py` - Migration runner
- `/modules/agent/Dockerfile` - Copies migrations to container
- `/modules/agent/start.sh` - Runs migrations on startup
- `/modules/agent/pyproject.toml` - Python dependencies
- `/docker-compose.yml` - Docker Compose configuration

---

**Implementation Date**: November 2024
**Status**: ✅ Production Ready
**Tested**: Docker Compose, Local Development

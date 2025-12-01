# CICD Five-Tier Deployment Strategy Spec

## Overview

Five deployment tiers with progressive stability guarantees, tag-based promotion, and multi-branch support for testing environments.

---

## Tier Summary

| Tier | Environment | Reset Mode | Tag Pattern | Multi-Branch | Data Source |
|------|-------------|------------|-------------|--------------|-------------|
| 1 | Local | ON | N/A | N/A | Seeders |
| 2 | Development | ON | `v*-alpha` | No | Seeders |
| 3 | Test | OFF | `v*-beta` | Yes | Stable test data |
| 4 | Staging | OFF | `v*-rc` | Yes | Nightly prod sync |
| 5 | Production | OFF | `v*` (no suffix) | No | Live data |

---

## Tier Details

### Tier 1: Local (`local`)
**Purpose:** Developer workstation environment

- **Reset Mode:** ON - migrations and seeders run on `docker compose up --build`
- **Trigger:** Manual (developer action)
- **Database:** Local Postgres container, destroyed/rebuilt on compose rebuild
- **Branch:** Any local branch

### Tier 2: Development (`development`)
**Purpose:** Shared web environment for integration testing of latest changes

- **Reset Mode:** ON - clean slate on every deployment
- **Trigger:** Push tag matching `v*-alpha` (e.g., `v1.0.0-alpha`, `v1.0.0-alpha.2`)
- **Database:** Dropped and rebuilt from migrations + seeders each deploy
- **URL:** `dev.pina-colada.app`
- **Use Case:** Quick verification that code works outside local environment

### Tier 3: Test (`test`)
**Purpose:** Stable test data for QA and automated test suites

- **Reset Mode:** OFF - data persists across deployments
- **Trigger:** Push tag matching `v*-beta` (e.g., `v1.0.0-beta`, `v1.0.0-beta.3`)
- **Database:** Migrations run (additive only), seeders skipped
- **Multi-Branch:** Yes, subdomain routing
  - Default: `test.pina-colada.app`
  - Branch-specific: `{branch-slug}.test.pina-colada.app`
- **Use Case:** Multiple feature branches tested against same stable dataset

### Tier 4: Staging (`staging`)
**Purpose:** Production-like environment with real data patterns

- **Reset Mode:** OFF - data persists, synced nightly from prod
- **Trigger:** Push tag matching `v*-rc` (e.g., `v1.0.0-rc.1`)
- **Database:**
  - Nightly job syncs from production (sanitized PII)
  - Migrations run on deploy
- **Multi-Branch:** Yes, subdomain routing
  - Default: `staging.pina-colada.app`
  - Branch-specific: `{branch-slug}.staging.pina-colada.app`
- **Use Case:** Final validation with production-scale data before release

### Tier 5: Production (`production`)
**Purpose:** Live customer-facing environment

- **Reset Mode:** OFF - data is sacred
- **Trigger:** Push tag matching `v*` with no suffix (e.g., `v1.0.0`, `v1.2.3`)
- **Database:** Migrations only (must be backwards-compatible)
- **URL:** `app.pina-colada.app`
- **Safeguards:**
  - Requires passing all previous tiers
  - Automatic rollback on health check failure

---

## Tag Promotion Flow

```
v1.0.0-alpha → v1.0.0-beta → v1.0.0-rc.1 → v1.0.0
    ↓              ↓              ↓            ↓
   Dev           Test         Staging       Prod
```

### Tagging Commands
```bash
# Deploy to dev (clean slate)
git tag v1.0.0-alpha && git push origin v1.0.0-alpha

# Promote to test (stable data)
git tag v1.0.0-beta && git push origin v1.0.0-beta

# Promote to staging (prod data copy)
git tag v1.0.0-rc.1 && git push origin v1.0.0-rc.1

# Release to production
git tag v1.0.0 && git push origin v1.0.0
```

---

## Multi-Branch Routing (Test/Staging)

### Architecture
- DNS: Wildcard `*.test.pina-colada.app` → Load balancer
- Reverse proxy (nginx/traefik) routes based on subdomain
- Each branch deployment runs as separate container/service

### Branch Slug Convention
- Branch: `feature/user-auth-123` → Subdomain: `feature-user-auth-123.test.pina-colada.app`
- Branch: `fix/login-bug` → Subdomain: `fix-login-bug.test.pina-colada.app`
- Sanitization: lowercase, replace `/` and `_` with `-`, truncate to 63 chars

### Shared Database
- All branch deployments share the same database
- Requires backwards-compatible migrations
- Feature flags for branch-specific behavior if needed

---

## Database Reset Behavior

### Reset ON (Local, Development)
```sql
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
-- Run all migrations
-- Run all seeders
```

### Reset OFF (Test, Staging, Production)
```sql
-- Run pending migrations only (tracked in schema_migrations table)
-- No schema drop, no seeders
```

---

## Nightly Staging Sync

### Process
1. Create sanitized snapshot of production database
2. Restore to staging database (replacing existing data)
3. Run any pending migrations

### Sanitization Rules
- Hash/anonymize PII (emails, names, addresses)
- Preserve data relationships and volume
- Keep non-sensitive business data intact

### Schedule
- Time: 2:00 AM UTC (low-traffic window)
- Retention: Keep last 7 snapshots

---

## Pipeline Files

| Environment | Pipeline File | Trigger |
|-------------|---------------|---------|
| Development | `azure-pipelines-dev.yml` | `v*-alpha` tag |
| Test | `azure-pipelines-test.yml` | `v*-beta` tag |
| Staging | `azure-pipelines-staging.yml` | `v*-rc*` tag |
| Production | `azure-pipelines-prod.yml` | `v*` (no suffix) tag |

---

## Environment Variables

Each tier requires `ENVIRONMENT` variable set appropriately:

| Tier | ENVIRONMENT Value |
|------|-------------------|
| Local | `local` |
| Development | `development` |
| Test | `test` |
| Staging | `staging` |
| Production | `production` |

This enables conditional pipeline logic and environment-specific behavior.

---

## Implementation Checklist

1. [ ] Create `azure-pipelines-dev.yml` with reset mode ON
2. [ ] Update `azure-pipelines-test.yml` to remove reset, add multi-branch routing
3. [ ] Create `azure-pipelines-staging.yml` with nightly sync job
4. [ ] Create `azure-pipelines-prod.yml` with safeguards
5. [ ] Configure DNS wildcards for test/staging subdomains
6. [ ] Set up reverse proxy routing rules
7. [ ] Create staging data sanitization script
8. [ ] Configure nightly sync scheduled job

# Feature: Test Deployment Stage

## Overview
Deploy test environments for agent and client to `test.api.pinacolada.co` and `test.pinacolada.co` via Azure DevOps pipelines triggered by the `develop` branch.

## User Story
As a developer, I want a test deployment environment so that I can validate changes before deploying to production.

---

## Scenarios

### Scenario 1: Test Deployment Triggered

**Given** code is merged to the `develop` branch
**When** Azure DevOps pipeline runs
**Then** Docker images are built with `test` tag, pushed to GHCR, and DigitalOcean App is redeployed

### Scenario 2: Test Database Migrations

**Given** new migrations exist in `modules/agent/migrations/`
**When** agent test pipeline runs
**Then** migrations are applied to `pina-colada-test` Supabase project

---

## Verification Checklist

### Functional Requirements
- [ ] Agent pipeline triggers on `develop` branch changes to `**/agent/**`
- [ ] Client pipeline triggers on `develop` branch changes to `**/client/**`
- [ ] Images pushed with `test` tag to GHCR
- [ ] Migrations run against test Supabase project
- [ ] DigitalOcean deployment triggered successfully

### Non-Functional Requirements
- [ ] Pipeline completes in < 10 minutes
- [ ] Test environment isolated from production data

### Edge Cases
- [ ] PR to develop runs build/migrate but does not push or deploy
- [ ] Failed migrations block push stage

---

## Implementation Notes

### Files Created
- `modules/agent/azure-pipelines-test.yml` - Agent test pipeline
- `modules/client/azure-pipelines-test.yml` - Client test pipeline

### Manual Prerequisites

#### DigitalOcean App Platform
Add two services to existing App:
- `pina-colada-agent-test` → `test.api.pinacolada.co` (pulls `ghcr.io/hubenschmidt/pina-colada-agent:test`)
- `pina-colada-client-test` → `test.pinacolada.co` (pulls `ghcr.io/hubenschmidt/pina-colada-client:test`)

#### Azure DevOps Variable Groups
Create `pina-colada-supabase-test` with:
- `SUPABASE_ACCESS_TOKEN`
- `SUPABASE_PROJECT_ID` (from `pina-colada-test` project)
- `SUPABASE_DB_PASSWORD`

### Key Differences from Prod

| Aspect | Prod | Test |
|--------|------|------|
| Branch | `master` | `develop` |
| Image tag | `latest` | `test` |
| Supabase group | `pina-colada-supabase-prod` | `pina-colada-supabase-test` |
| Domain | api.pinacolada.co | test.api.pinacolada.co |

### Out of Scope
- Separate DigitalOcean App for full isolation (using same App for simplicity)
- Automated tests in pipeline

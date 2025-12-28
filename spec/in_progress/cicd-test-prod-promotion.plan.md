# Feature: CI/CD Test-to-Prod Promotion Pipeline

## Overview
Implement a CI/CD pipeline that builds a single Docker image, deploys to test environment for validation, requires manual approval, then promotes the same image to production. Includes automatic rollback capability when PRs are declined.

## User Story
As a developer, I want PRs to master to automatically deploy to test for validation before production, so that I can verify changes in a test environment before they reach users.

---

## Architecture

### Pipeline Flow

```
PR to master:    Build → Push → Deploy Test → [Approval] → Deploy Prod
Merge to master: Build → Push → Deploy Prod (skip test, already validated)
Develop branch:  Build → Push → Deploy Test (via separate test pipeline)
```

### Infrastructure
- **Container Registry**: GitHub Container Registry (GHCR)
- **Hosting**: Render (both test and prod)
- **CI/CD**: Azure DevOps Pipelines
- **Database**: Supabase (separate test and prod instances)

---

## Scenarios

### Scenario 1: PR Validation Flow

**Given** a developer opens a PR to master
**When** the pipeline runs
**Then**
- Image is built and pushed to GHCR
- Test database migrations run
- Image deploys to test environment on Render
- Pipeline pauses for manual approval
- After approval, prod migrations run and image deploys to prod

### Scenario 2: Direct Master Merge

**Given** a PR has been approved and merged to master
**When** the pipeline runs on the merge commit
**Then** image builds, pushes, and deploys directly to prod (skipping test stages)

### Scenario 3: PR Declined - Rollback

**Given** a PR deployed to test (and possibly prod after approval)
**When** the PR is abandoned/declined
**Then** rollback pipeline triggers to restore previous deployment

### Scenario 4: Develop Branch Deployment

**Given** a developer pushes to the develop branch
**When** the test pipeline runs
**Then** image builds and deploys to test environment only

---

## Verification Checklist

### Functional Requirements
- [x] Single image built once, deployed to both environments
- [x] PR to master triggers test deployment first
- [x] Manual approval gate before prod deployment
- [x] Merge to master skips test stages
- [x] Develop branch deploys to test via separate pipeline
- [x] Version tracking before each deployment for rollback
- [ ] Automatic rollback on PR decline (service hook pending)

### Infrastructure Setup
- [ ] Create `pina-colada-prod` environment in Azure DevOps with approval check
- [ ] Create 4 rollback pipelines in Azure DevOps
- [ ] Configure service hooks for automatic rollback (optional)

### Variable Groups Required

#### `pina-colada-render-test`
- [ ] `RENDER_API_KEY_TEST`
- [ ] `RENDER_SERVICE_ID_AGENT_TEST`
- [ ] `RENDER_SERVICE_ID_CLIENT_TEST`
- [ ] `RENDER_DEPLOY_HOOK_AGENT_TEST`
- [ ] `RENDER_DEPLOY_HOOK_CLIENT_TEST`

#### `pina-colada-render-prod`
- [ ] `RENDER_API_KEY_PROD`
- [ ] `RENDER_SERVICE_ID_AGENT_PROD`
- [ ] `RENDER_SERVICE_ID_CLIENT_PROD`
- [ ] `RENDER_DEPLOY_HOOK_AGENT` (prod)
- [ ] `RENDER_DEPLOY_HOOK_CLIENT` (prod)

---

## Implementation Notes

### Files Modified/Created

| File | Description |
|------|-------------|
| `render.yaml` | Added `pina-colada-agent-test` and `pina-colada-client-test` services |
| `modules/agent/azure-pipelines-prod.yml` | PR flow: test deploy → approval → prod deploy |
| `modules/client/azure-pipelines-prod.yml` | Same pattern (no migrations) |
| `modules/agent/azure-pipelines-test.yml` | Updated from DigitalOcean to Render |
| `modules/client/azure-pipelines-test.yml` | Updated from DigitalOcean to Render |
| `modules/agent/azure-pipelines-rollback-test.yml` | **New** - rollback test environment |
| `modules/agent/azure-pipelines-rollback-prod.yml` | **New** - rollback prod (requires approval) |
| `modules/client/azure-pipelines-rollback-test.yml` | **New** - rollback test environment |
| `modules/client/azure-pipelines-rollback-prod.yml` | **New** - rollback prod (requires approval) |

### Pipeline Stages (Prod Pipeline)

```yaml
stages:
  - Build           # Always runs
  - Push            # Always runs
  - MigrateTest     # PR only
  - DeployTest      # PR only
  - Approval        # PR only (uses pina-colada-prod environment)
  - MigrateProd     # After approval OR on master merge
  - DeployProd      # After MigrateProd succeeds
```

### Service Hook Setup (Optional - for auto rollback)

1. **Azure DevOps → Project Settings → Service hooks → New**
2. Service: Web Hooks or Pipelines
3. Trigger: Pull request updated
4. Filters:
   - Target branch: `master`
   - Status changed to: `Abandoned`
5. Action: Queue rollback pipeline

### Dependencies
- Render API key with deploy permissions
- Azure DevOps environment with approval gates
- GHCR credentials configured in Azure DevOps

### Out of Scope
- Blue/green deployments
- Canary releases
- Automated integration tests in pipeline
- Multi-region deployments

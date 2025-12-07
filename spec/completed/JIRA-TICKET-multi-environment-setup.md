# Multi-Environment Setup: Test and Production on Digital Ocean App Platform

## Summary
Set up test and production environments on Digital Ocean App Platform with proper domain routing, separate Supabase databases, and CI/CD pipeline configuration.

## Current State
- Single production environment deployed from `master` branch
- Production domain: `pinacolada.co` (API should be at `api.pinacolada.co`)
- Single Supabase production database
- Azure Pipelines: `azure-pipelines-prod.yml` triggers on `master` branch
- No test environment exists

## Desired State
- **Test Environment:**
  - Domain: `test.pina-colada.co` (frontend) and `test.api.pinacolada.co` (API)
  - Deploys from `develop` branch
  - Separate Supabase test database (same Supabase project, different database)
  - Runs migrations automatically
  - Runs seeders if `RUN_SEEDERS=true`

- **Production Environment:**
  - Domain: `pinacolada.co` (frontend) and `api.pinacolada.co` (API)
  - Deploys from `master` branch
  - Production Supabase database
  - Runs migrations automatically
  - Does NOT run seeders

## Implementation Steps

### 1. GitHub Branch Strategy
- [ ] Create `develop` branch from `master`
- [ ] Set `develop` as default branch for feature work
- [ ] Update branch protection rules:
  - `develop`: Require PR reviews, allow force push for hotfixes
  - `master`: Require PR reviews, no force push

### 2. Supabase Database Setup
- [ ] In Supabase Dashboard, create a new database for test environment
  - Project: Same Supabase project as production
  - Database name: `pina_colada_test` (or similar)
  - Note: Supabase uses separate projects per environment typically, but user requested same project
  - **Alternative**: Create separate Supabase project for test (recommended for better isolation)
- [ ] Create Azure DevOps variable group: `pina-colada-supabase-test`
  - `SUPABASE_ACCESS_TOKEN` (can reuse same token)
  - `SUPABASE_PROJECT_ID` (test project/database ID)
  - `SUPABASE_DB_PASSWORD` (test database password)
  - `SUPABASE_URL` (test project URL if separate project, or same URL if same project)

### 3. Digital Ocean App Platform Configuration

#### Option A: Single App with Multiple Routes (Recommended)
Digital Ocean App Platform supports multiple routes/domains on a single app spec.

- [ ] In Digital Ocean App Platform dashboard:
  - Navigate to your app
  - Go to Settings > Domains
  - Add domains:
    - `test.pina-colada.co` → Route to test component
    - `test.api.pinacolada.co` → Route to test API component
    - `api.pinacolada.co` → Route to production API component (if not already set)
  - Configure routing rules:
    - Use environment variables to determine which component serves which domain
    - Or use separate components/services within the same app spec

#### Option B: Separate App Specs (If routing doesn't work)
- [ ] Create new App Platform app for test environment
- [ ] Configure domains:
  - `test.pina-colada.co`
  - `test.api.pinacolada.co`
- [ ] Set environment variables:
  - `ENVIRONMENT=test`
  - `RUN_SEEDERS=true`
  - Test Supabase credentials

**Note**: Digital Ocean App Platform uses a single app spec YAML file. To support multiple environments:
- Use environment variables to differentiate behavior
- Or use separate apps (more expensive but cleaner separation)

### 4. Domain DNS Configuration
- [ ] Add DNS records for test domains:
  - `test.pina-colada.co` → CNAME to Digital Ocean app
  - `test.api.pinacolada.co` → CNAME to Digital Ocean app
- [ ] Verify `api.pinacolada.co` is correctly configured for production
- [ ] Update SSL certificates in Digital Ocean (automatic if using DO-managed domains)

### 5. Azure Pipelines Configuration

#### Create `azure-pipelines-test.yml`
- [ ] Create new pipeline file: `modules/agent/azure-pipelines-test.yml`
- [ ] Configure triggers:
  - Branch: `develop`
  - Paths: `**/agent/**`
- [ ] Stages:
  1. **Build**: Build Docker image, tag as `test` variant
  2. **Migrate**: Run migrations against test Supabase database
  3. **Push**: Push image to GHCR with `-test` suffix
  4. **Deploy**: Trigger Digital Ocean test app deployment

#### Update `azure-pipelines-prod.yml`
- [ ] Ensure production pipeline only triggers on `master`
- [ ] Verify migrations run against production Supabase database
- [ ] Ensure `RUN_SEEDERS` is NOT set in production environment variables

### 6. Environment Variables Configuration

#### Test Environment (Digital Ocean App Platform)
- [ ] Set environment variables:
  ```
  ENVIRONMENT=test
  RUN_SEEDERS=true
  SUPABASE_URL=<test_supabase_url>
  SUPABASE_SERVICE_KEY=<test_service_key>
  SUPABASE_DB_PASSWORD=<test_db_password>
  SUPABASE_PROJECT_ID=<test_project_id>
  ```

#### Production Environment (Digital Ocean App Platform)
- [ ] Verify environment variables:
  ```
  ENVIRONMENT=production
  RUN_SEEDERS=<not set or false>
  SUPABASE_URL=<prod_supabase_url>
  SUPABASE_SERVICE_KEY=<prod_service_key>
  SUPABASE_DB_PASSWORD=<prod_db_password>
  SUPABASE_PROJECT_ID=<prod_project_id>
  ```

### 7. Code Changes Required

#### Update Seeder Script
- [x] Already updated: `scripts/run_seeders.py` checks `RUN_SEEDERS=true` (exists and equals "true")

#### Update Start Scripts
- [ ] Verify `start-prod.sh` doesn't run seeders (should not call `run_seeders.py`)
- [ ] Ensure test environment calls seeders (via `RUN_SEEDERS=true` env var)

### 8. Azure DevOps Variable Groups

#### Create Variable Groups
- [ ] `pina-colada-digital-ocean-test`:
  - `DO_API_TOKEN` (can reuse same token)
  - `DO_APP_ID` (test app ID if using separate apps)
- [ ] `pina-colada-supabase-test`:
  - `SUPABASE_ACCESS_TOKEN`
  - `SUPABASE_PROJECT_ID`
  - `SUPABASE_DB_PASSWORD`
  - `SUPABASE_URL`

### 9. Testing Checklist
- [ ] Test deployment from `develop` branch triggers test pipeline
- [ ] Test deployment from `master` branch triggers production pipeline
- [ ] Verify test environment accessible at `test.api.pinacolada.co`
- [ ] Verify production environment accessible at `api.pinacolada.co`
- [ ] Verify migrations run in both environments
- [ ] Verify seeders run in test but NOT in production
- [ ] Verify test and production databases are separate

### 10. Documentation Updates
- [ ] Update README with environment setup instructions
- [ ] Document branch strategy (develop → test, master → prod)
- [ ] Document how to add new environments in the future

## Questions to Resolve

1. **Supabase Database Strategy**: 
   - Same Supabase project with separate databases? (Requires manual database creation)
   - Or separate Supabase projects? (Recommended, cleaner isolation)

2. **Digital Ocean App Platform Routing**:
   - Can single app spec handle multiple domains with environment-based routing?
   - Or do we need separate apps? (Check DO documentation for routing capabilities)

3. **Domain Configuration**:
   - Confirm `api.pinacolada.co` is currently configured correctly
   - Verify DNS provider and current DNS records

## Estimated Effort
- Supabase setup: 1-2 hours
- Digital Ocean configuration: 2-3 hours
- Azure Pipelines setup: 2-3 hours
- Testing and verification: 2-3 hours
- **Total: 7-11 hours**

## Dependencies
- Access to Digital Ocean App Platform dashboard
- Access to Supabase Dashboard
- Access to Azure DevOps project
- DNS access for domain configuration

## Acceptance Criteria
- [ ] Test environment accessible at `test.api.pinacolada.co`
- [ ] Production environment accessible at `api.pinacolada.co`
- [ ] Deployments from `develop` go to test
- [ ] Deployments from `master` go to production
- [ ] Migrations run in both environments
- [ ] Seeders run only in test environment
- [ ] Test and production databases are separate


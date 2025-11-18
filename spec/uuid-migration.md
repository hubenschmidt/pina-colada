# UUID Migration Plan

## Phase 1: Update Migration Files (5 files)

Change all `BIGSERIAL` → `UUID DEFAULT gen_random_uuid()` and `BIGINT` FKs → `UUID`

- `001_initial_schema.sql` - Core tables (Tenant, User, Role, etc.)
- `002_multitenancy.sql` - Multi-tenancy updates
- `003_auth0_integration.sql` - Auth0 integration
- `004_role_tenant_id.sql` - Role tenant updates
- `005_create_preferences_tables.sql` - Preferences tables

## Phase 2: Update SQLAlchemy Models (18 files)

Replace `Column(BigInteger, ...)` → `Column(UUID(as_uuid=True), ..., server_default=text("gen_random_uuid()"))`

- User.py, Tenant.py, Organization.py, Contact.py, Individual.py, Deal.py, Lead.py, Opportunity.py, Partnership.py, Role.py, UserRole.py, Task.py, Job.py, Status.py, TenantPreferences.py, UserPreferences.py, Activity.py, Organization_Relationship.py

## Phase 3: Update Seeder Files (3 files)

Change `BIGINT` variable declarations → `UUID` and ensure `SELECT id INTO` works with UUIDs

- `001_initial_seed.sql` - Replace ~50 BIGINT declarations with UUID (lines 39-51, 113-150, 295-304, 430-433, 514-518)
- `002_jobs_seed.sql` - Replace 6 BIGINT declarations with UUID (lines 9-15)
- `003_job_statuses_seed.sql` - No ID variables, no changes needed

## Phase 4: Dev Testing

- Drop local dev database
- Run migrations 001-005 (now with UUIDs)
- Run all 3 seeders
- Verify API endpoints work with UUIDs

## Phase 5: Prod Deployment (later - you handle)

- Export current prod data
- Provide export to me for UUID conversion
- Drop prod DB in Supabase
- Run migrations 001-005
- Run converted prod seeder

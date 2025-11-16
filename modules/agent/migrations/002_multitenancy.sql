-- ==============================
-- Multi-Tenancy Migration
-- ==============================
-- Adds tenant isolation with users and roles
--
-- Naming convention:
-- - Tenant = The app customer/company (multi-tenant isolation)
-- - Organization = CRM entity (companies in their pipeline)
-- ==============================

-- ==============================
-- STEP 1: Create Tenant Tables
-- ==============================

-- Tenant table (the app customer/company)
CREATE TABLE IF NOT EXISTS "Tenant" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  slug            TEXT NOT NULL UNIQUE,  -- URL-safe identifier
  status          TEXT DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'trial', 'cancelled')),
  plan            TEXT DEFAULT 'free' CHECK (plan IN ('free', 'starter', 'professional', 'enterprise')),
  settings        JSONB DEFAULT '{}'::jsonb,  -- Tenant-specific settings
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- User table (belongs to one Tenant)
CREATE TABLE IF NOT EXISTS "User" (
  id              BIGSERIAL PRIMARY KEY,
  tenant_id       BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
  email           TEXT NOT NULL,
  first_name      TEXT,
  last_name       TEXT,
  avatar_url      TEXT,
  status          TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'invited')),
  last_login_at   TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(tenant_id, email)  -- Email unique per tenant
);

-- Create index for user lookups
CREATE INDEX IF NOT EXISTS idx_user_tenant_id ON "User"(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_email ON "User"(email);

-- Role table (system-defined roles)
CREATE TABLE IF NOT EXISTS "Role" (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL UNIQUE,  -- 'owner', 'admin', 'member', 'viewer'
  description     TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- UserRole junction table (users can have multiple roles)
CREATE TABLE IF NOT EXISTS "UserRole" (
  user_id         BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
  role_id         BIGINT NOT NULL REFERENCES "Role"(id) ON DELETE CASCADE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, role_id)
);

-- ==============================
-- STEP 2: Add tenant_id to existing tables
-- ==============================

-- Add tenant_id to Deal (top-level entity)
ALTER TABLE "Deal" ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES "Tenant"(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_deal_tenant_id ON "Deal"(tenant_id);

-- Add tenant_id to Organization (companies in their CRM)
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES "Tenant"(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_organization_tenant_id ON "Organization"(tenant_id);

-- Add tenant_id to Individual (people in their CRM)
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES "Tenant"(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_individual_tenant_id ON "Individual"(tenant_id);

-- Add tenant_id to Project
ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES "Tenant"(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_project_tenant_id ON "Project"(tenant_id);

-- Lead, Job, Opportunity, Partnership, Contact inherit tenant_id through relationships
-- (Lead via Deal, Job via Lead, Contact via Organization/Individual)

-- ==============================
-- STEP 3: Update Triggers
-- ==============================

DROP TRIGGER IF EXISTS update_tenant_updated_at ON "Tenant";
CREATE TRIGGER update_tenant_updated_at
    BEFORE UPDATE ON "Tenant"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_user_updated_at ON "User";
CREATE TRIGGER update_user_updated_at
    BEFORE UPDATE ON "User"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ==============================
-- STEP 4: Seed Default Roles
-- ==============================

INSERT INTO "Role" (name, description) VALUES
('owner', 'Tenant owner with full access and billing control'),
('admin', 'Administrator with full access to all features'),
('member', 'Standard member with access to most features'),
('viewer', 'Read-only access to view data')
ON CONFLICT (name) DO NOTHING;

-- ==============================
-- STEP 5: Data Integrity Constraints
-- ==============================

-- Once multi-tenancy is fully implemented, uncomment these to enforce tenant_id:
-- ALTER TABLE "Deal" ALTER COLUMN tenant_id SET NOT NULL;
-- ALTER TABLE "Organization" ALTER COLUMN tenant_id SET NOT NULL;
-- ALTER TABLE "Individual" ALTER COLUMN tenant_id SET NOT NULL;
-- ALTER TABLE "Project" ALTER COLUMN tenant_id SET NOT NULL;

-- Update unique constraint for Organization to be scoped per tenant
DROP INDEX IF EXISTS idx_organization_name_lower;
CREATE UNIQUE INDEX idx_organization_name_lower_tenant ON "Organization"(tenant_id, LOWER(name));

-- Update unique constraint for Individual email to be scoped per tenant
DROP INDEX IF EXISTS idx_individual_email_lower;
CREATE UNIQUE INDEX idx_individual_email_lower_tenant ON "Individual"(tenant_id, LOWER(email));

-- ==============================
-- STEP 6: Row Level Security (RLS) Preparation
-- ==============================
-- Note: These are prepared for when you implement authentication
-- You'll need to set up Supabase Auth and configure these policies

-- Enable RLS on tenant-scoped tables
ALTER TABLE "Tenant" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "User" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "UserRole" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "Deal" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "Organization" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "Individual" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "Project" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "Lead" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "Job" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "Contact" ENABLE ROW LEVEL SECURITY;

-- Example RLS policies (commented out - implement when auth is ready)
/*
-- Users can only see their own tenant's data
CREATE POLICY "Users can view their tenant's deals" ON "Deal"
    FOR SELECT
    USING (tenant_id = (SELECT tenant_id FROM "User" WHERE id = auth.uid()));

CREATE POLICY "Users can create deals in their tenant" ON "Deal"
    FOR INSERT
    WITH CHECK (tenant_id = (SELECT tenant_id FROM "User" WHERE id = auth.uid()));

CREATE POLICY "Users can update their tenant's deals" ON "Deal"
    FOR UPDATE
    USING (tenant_id = (SELECT tenant_id FROM "User" WHERE id = auth.uid()));

CREATE POLICY "Users can delete their tenant's deals" ON "Deal"
    FOR DELETE
    USING (tenant_id = (SELECT tenant_id FROM "User" WHERE id = auth.uid()));
*/

-- ==============================
-- Migration Complete!
-- ==============================

-- Verification Queries:
-- SELECT * FROM "Tenant";
-- SELECT * FROM "User";
-- SELECT * FROM "Role";
-- SELECT COUNT(*) as deals_needing_tenant FROM "Deal" WHERE tenant_id IS NULL;

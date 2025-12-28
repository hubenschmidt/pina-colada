-- Migration 007: Account table as central tenant owner
-- Creates Account table with tenant relationship
-- Organization and Individual link to Account (tenant via Account)
-- User links to Individual for ownership

-- 1. Create Account table with tenant_id
CREATE TABLE IF NOT EXISTS "Account" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES "Tenant"(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_Account_name ON "Account"(name);
CREATE INDEX IF NOT EXISTS idx_account_tenant_id ON "Account"(tenant_id);

-- 2. Add account_id to Organization (remove tenant_id - now via Account)
ALTER TABLE "Organization" ADD COLUMN IF NOT EXISTS account_id BIGINT;
ALTER TABLE "Organization" ADD CONSTRAINT fk_organization_account
    FOREIGN KEY (account_id) REFERENCES "Account"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS ix_Organization_account_id ON "Organization"(account_id);

-- Drop tenant_id from Organization (now via Account)
DROP INDEX IF EXISTS idx_organization_tenant_id;
DROP INDEX IF EXISTS idx_organization_name_lower_tenant;
ALTER TABLE "Organization" DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_name_lower ON "Organization"((LOWER(name)));

-- 3. Add account_id to Individual (remove tenant_id - now via Account)
ALTER TABLE "Individual" ADD COLUMN IF NOT EXISTS account_id BIGINT;
ALTER TABLE "Individual" ADD CONSTRAINT fk_individual_account
    FOREIGN KEY (account_id) REFERENCES "Account"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS ix_Individual_account_id ON "Individual"(account_id);

-- Drop tenant_id from Individual (now via Account)
DROP INDEX IF EXISTS idx_individual_tenant_id;
DROP INDEX IF EXISTS idx_individual_email_lower_tenant;
ALTER TABLE "Individual" DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS idx_individual_email_lower ON "Individual"((LOWER(email)));

-- 4. Add individual_id to User table
ALTER TABLE "User" ADD COLUMN IF NOT EXISTS individual_id BIGINT;
ALTER TABLE "User" ADD CONSTRAINT fk_user_individual
    FOREIGN KEY (individual_id) REFERENCES "Individual"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS ix_User_individual_id ON "User"(individual_id);

-- 5. Add owner_individual_id to Deal, drop owner_user_id
ALTER TABLE "Deal" ADD COLUMN IF NOT EXISTS owner_individual_id BIGINT;
ALTER TABLE "Deal" ADD CONSTRAINT fk_deal_owner_individual
    FOREIGN KEY (owner_individual_id) REFERENCES "Individual"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS ix_Deal_owner_individual_id ON "Deal"(owner_individual_id);
ALTER TABLE "Deal" DROP COLUMN IF EXISTS owner_user_id;

-- 6. Add owner_individual_id to Lead, drop owner_user_id
ALTER TABLE "Lead" ADD COLUMN IF NOT EXISTS owner_individual_id BIGINT;
ALTER TABLE "Lead" ADD CONSTRAINT fk_lead_owner_individual
    FOREIGN KEY (owner_individual_id) REFERENCES "Individual"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS ix_Lead_owner_individual_id ON "Lead"(owner_individual_id);
ALTER TABLE "Lead" DROP COLUMN IF EXISTS owner_user_id;

-- 7. Add assigned_to_individual_id to Task, drop assigned_to_user_id
ALTER TABLE "Task" ADD COLUMN IF NOT EXISTS assigned_to_individual_id BIGINT;
ALTER TABLE "Task" ADD CONSTRAINT fk_task_assigned_to_individual
    FOREIGN KEY (assigned_to_individual_id) REFERENCES "Individual"(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS ix_Task_assigned_to_individual_id ON "Task"(assigned_to_individual_id);
ALTER TABLE "Task" DROP COLUMN IF EXISTS assigned_to_user_id;

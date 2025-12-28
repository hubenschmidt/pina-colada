-- Add tenant_id to Role table for multi-tenancy
ALTER TABLE "Role" ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE "Role" ADD CONSTRAINT fk_role_tenant
    FOREIGN KEY (tenant_id) REFERENCES "Tenant"(id) ON DELETE CASCADE;

-- Remove unique constraint on name since roles are now tenant-scoped
ALTER TABLE "Role" DROP CONSTRAINT IF EXISTS "Role_name_key";

-- Add unique constraint for (tenant_id, name)
ALTER TABLE "Role" ADD CONSTRAINT role_tenant_name_unique
    UNIQUE (tenant_id, name);

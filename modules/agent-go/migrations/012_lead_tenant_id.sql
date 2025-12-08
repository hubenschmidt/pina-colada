-- Add tenant_id to Lead for multitenancy

ALTER TABLE "Lead" ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES "Tenant"(id) ON DELETE CASCADE;

-- Index for fast tenant filtering
CREATE INDEX IF NOT EXISTS idx_lead_tenant_id ON "Lead"(tenant_id);

-- Update existing leads to use the tenant from their deal's...
-- Actually, Deal doesn't have tenant_id either. Set from the seeded tenant.
UPDATE "Lead" SET tenant_id = (SELECT id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1) WHERE tenant_id IS NULL;

-- Agent Proposal System: approval queue for agent-proposed CUD operations

-- 1. Create Agent_Proposal table (approval queue)
CREATE TABLE IF NOT EXISTS "Agent_Proposal" (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    proposed_by_id  BIGINT NOT NULL REFERENCES "User"(id),
    entity_type     VARCHAR(50) NOT NULL,
    entity_id       BIGINT,
    operation       VARCHAR(10) NOT NULL,
    payload         JSONB NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    validation_errors JSONB,
    reviewed_by_id  BIGINT REFERENCES "User"(id),
    reviewed_at     TIMESTAMPTZ,
    executed_at     TIMESTAMPTZ,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_proposal_tenant_status ON "Agent_Proposal"(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_agent_proposal_entity ON "Agent_Proposal"(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_agent_proposal_proposed_by ON "Agent_Proposal"(proposed_by_id);

COMMENT ON COLUMN "Agent_Proposal".entity_id IS 'NULL for CREATE operations';
COMMENT ON COLUMN "Agent_Proposal".operation IS 'create, update, or delete';
COMMENT ON COLUMN "Agent_Proposal".status IS 'pending, approved, rejected, executed, or failed';
COMMENT ON COLUMN "Agent_Proposal".validation_errors IS 'Structured validation errors from payload validation, null if valid';

-- 2. Create Entity_Approval_Config table (per-entity approval settings)
CREATE TABLE IF NOT EXISTS "Entity_Approval_Config" (
    id                BIGSERIAL PRIMARY KEY,
    tenant_id         BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    entity_type       VARCHAR(50) NOT NULL,
    requires_approval BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, entity_type)
);

-- 3. Create system agent user (tenant 1)
INSERT INTO "User" (email, is_system_user, status, tenant_id, created_at, updated_at)
VALUES ('agent@system.local', true, 'active', 1, NOW(), NOW())
ON CONFLICT (tenant_id, email) DO NOTHING;

-- 4. Create system-agent role for tenant 1
INSERT INTO "Role" (name, tenant_id, description, created_at, updated_at)
SELECT 'system-agent', 1, 'System agent with read-only entity access', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM "Role" WHERE name = 'system-agent' AND tenant_id = 1);

-- 5. Grant READ ONLY permissions to system-agent (no create/update/delete)
INSERT INTO "Role_Permission" (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM "Role" r
CROSS JOIN "Permission" p
WHERE r.name = 'system-agent' AND r.tenant_id = 1
  AND p.action = 'read'
ON CONFLICT DO NOTHING;

-- 6. Assign system-agent role to agent user
INSERT INTO "User_Role" (user_id, role_id, created_at)
SELECT u.id, r.id, NOW()
FROM "User" u
CROSS JOIN "Role" r
WHERE u.email = 'agent@system.local' AND u.tenant_id = 1
  AND r.name = 'system-agent' AND r.tenant_id = 1
ON CONFLICT DO NOTHING;

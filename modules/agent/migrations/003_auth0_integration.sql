-- Auth0 Integration Migration
-- Adds Auth0 subject mapping to User table and creates invitation system

-- Add Auth0 subject mapping to User table
ALTER TABLE "User" ADD COLUMN IF NOT EXISTS auth0_sub TEXT UNIQUE;
CREATE INDEX IF NOT EXISTS idx_user_auth0_sub ON "User"(auth0_sub);

-- Make tenant_id nullable for new users who haven't selected a tenant yet
ALTER TABLE "User" ALTER COLUMN tenant_id DROP NOT NULL;

-- Add invitation system for tenant onboarding
CREATE TABLE IF NOT EXISTS "TenantInvitation" (
  id              BIGSERIAL PRIMARY KEY,
  tenant_id       BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
  email           TEXT NOT NULL,
  role_id         BIGINT NOT NULL REFERENCES "Role"(id),
  token           TEXT NOT NULL UNIQUE,
  expires_at      TIMESTAMPTZ NOT NULL,
  accepted_at     TIMESTAMPTZ,
  invited_by      BIGINT REFERENCES "User"(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invitation_token ON "TenantInvitation"(token);
CREATE INDEX IF NOT EXISTS idx_invitation_email ON "TenantInvitation"(email);

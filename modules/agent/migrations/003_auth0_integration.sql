-- Auth0 Integration Migration
-- Adds Auth0 subject mapping to User table and creates invitation system

-- Add Auth0 subject mapping to User table
ALTER TABLE "User" ADD COLUMN IF NOT EXISTS auth0_sub TEXT UNIQUE;
CREATE INDEX IF NOT EXISTS idx_user_auth0_sub ON "User"(auth0_sub);

-- Make tenant_id nullable for new users who haven't selected a tenant yet
ALTER TABLE "User" ALTER COLUMN tenant_id DROP NOT NULL;

-- Add industry, website, employee_count columns to Tenant table

ALTER TABLE "Tenant" ADD COLUMN IF NOT EXISTS industry TEXT;
ALTER TABLE "Tenant" ADD COLUMN IF NOT EXISTS website TEXT;
ALTER TABLE "Tenant" ADD COLUMN IF NOT EXISTS employee_count INTEGER;

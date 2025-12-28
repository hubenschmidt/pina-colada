-- Migration: Make Contact a standalone entity
-- Add first_name, last_name fields
-- Make individual_id nullable
-- Copy existing names from linked Individuals

-- Add name fields to Contact
ALTER TABLE "Contact" ADD COLUMN IF NOT EXISTS first_name TEXT;
ALTER TABLE "Contact" ADD COLUMN IF NOT EXISTS last_name TEXT;

-- Migrate existing: copy names from linked Individual
UPDATE "Contact" c
SET first_name = i.first_name, last_name = i.last_name
FROM "Individual" i
WHERE c.individual_id = i.id
  AND c.first_name IS NULL;

-- Make individual_id nullable
ALTER TABLE "Contact" ALTER COLUMN individual_id DROP NOT NULL;

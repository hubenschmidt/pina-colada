-- Migration: Make Contact.first_name and last_name NOT NULL
-- First, update any existing NULL values to empty string
UPDATE "Contact" SET first_name = '' WHERE first_name IS NULL;
UPDATE "Contact" SET last_name = '' WHERE last_name IS NULL;

-- Then add the NOT NULL constraints
ALTER TABLE "Contact" ALTER COLUMN first_name SET NOT NULL;
ALTER TABLE "Contact" ALTER COLUMN last_name SET NOT NULL;

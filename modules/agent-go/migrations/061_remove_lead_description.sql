-- Migration: Remove duplicate description column from Lead table
-- The description field is redundant because:
-- - Jobs use description from Job table
-- - Opportunities use description from Opportunity table
-- - Partnerships use description from Partnership table

ALTER TABLE "Lead" DROP COLUMN IF EXISTS description;

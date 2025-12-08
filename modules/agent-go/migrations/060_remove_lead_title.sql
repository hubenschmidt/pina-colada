-- Migration: Remove duplicate title column from Lead table
-- The title field is redundant because:
-- - Jobs use job_title from Job table
-- - Opportunities use opportunity_name from Opportunity table
-- - Partnerships use partnership_name from Partnership table

ALTER TABLE "Lead" DROP COLUMN IF EXISTS title;

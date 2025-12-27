-- ==============================
-- Migration: Rename notes to description in Job, Opportunity, Partnership tables
-- ==============================
-- The "notes" field in these lead subtypes should be called "description"
-- to better reflect its purpose as a description of the job/opportunity/partnership

-- Rename column in Job table
ALTER TABLE "Job" RENAME COLUMN notes TO description;

-- Rename column in Opportunity table
ALTER TABLE "Opportunity" RENAME COLUMN notes TO description;

-- Rename column in Partnership table
ALTER TABLE "Partnership" RENAME COLUMN notes TO description;

-- Remove date_posted_confidence column from Job table
ALTER TABLE "Job" DROP COLUMN IF EXISTS date_posted_confidence;

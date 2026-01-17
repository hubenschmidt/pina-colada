-- Add date_posted_confidence column to Job table
ALTER TABLE "Job" ADD COLUMN IF NOT EXISTS date_posted_confidence VARCHAR(10);

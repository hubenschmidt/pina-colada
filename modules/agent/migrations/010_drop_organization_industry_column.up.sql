-- Drop redundant industry TEXT column from Organization (now using many-to-many)

ALTER TABLE "Organization" DROP COLUMN IF EXISTS industry;

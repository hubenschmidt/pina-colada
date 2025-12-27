-- Migration: Rename Individual.notes to Individual.description
-- This avoids confusion with the Notes entity which stores separate notes records

ALTER TABLE "Individual" RENAME COLUMN notes TO description;

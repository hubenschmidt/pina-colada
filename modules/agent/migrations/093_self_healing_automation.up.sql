-- Self-healing automation: prompt suggestions and simplified single search query
-- Add prompt suggestion fields
ALTER TABLE "Automation_Config"
ADD COLUMN suggested_prompt TEXT,
ADD COLUMN use_suggested_prompt BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN suggestion_threshold INTEGER NOT NULL DEFAULT 50;

-- Add single search query fields (replacing search_slots array)
-- excluded: terms to exclude from search (e.g. "-junior -intern"), applied at search time
ALTER TABLE "Automation_Config"
ADD COLUMN search_query TEXT,
ADD COLUMN suggested_query TEXT,
ADD COLUMN use_suggested_query BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN excluded TEXT;

-- Migrate data from search_slots array to single search_query
UPDATE "Automation_Config"
SET search_query = search_slots->>0
WHERE search_slots IS NOT NULL AND search_slots->>0 IS NOT NULL;

-- Drop old multi-slot columns
ALTER TABLE "Automation_Config"
DROP COLUMN IF EXISTS concurrent_searches,
DROP COLUMN IF EXISTS search_slots;

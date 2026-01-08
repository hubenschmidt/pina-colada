-- Convert search_slots from [][]string to []string format
-- Old: [["senior software engineer"], ["full stack developer"]]
-- New: ["senior software engineer", "full stack developer"]

UPDATE "Automation_Config"
SET search_slots = (
    SELECT jsonb_agg(
        CASE
            WHEN jsonb_typeof(slot) = 'array' THEN slot->>0
            ELSE slot::text
        END
    )
    FROM jsonb_array_elements(search_slots) AS slot
)
WHERE search_slots IS NOT NULL
  AND jsonb_typeof(search_slots) = 'array'
  AND EXISTS (
    SELECT 1 FROM jsonb_array_elements(search_slots) AS s
    WHERE jsonb_typeof(s) = 'array'
  );

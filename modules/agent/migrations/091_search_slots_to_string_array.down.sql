-- Revert search_slots from []string to [][]string format
-- Note: This wraps each string in an array

UPDATE "Automation_Config"
SET search_slots = (
    SELECT jsonb_agg(jsonb_build_array(slot))
    FROM jsonb_array_elements_text(search_slots) AS slot
)
WHERE search_slots IS NOT NULL
  AND jsonb_typeof(search_slots) = 'array'
  AND NOT EXISTS (
    SELECT 1 FROM jsonb_array_elements(search_slots) AS s
    WHERE jsonb_typeof(s) = 'array'
  );

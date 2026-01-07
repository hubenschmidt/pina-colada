-- search_keywords is deprecated; search_slots now stores keywords directly as [][]string
ALTER TABLE "Automation_Config"
DROP COLUMN search_keywords;

ALTER TABLE "Automation_Config"
    DROP COLUMN IF EXISTS vector_prefilter_enabled,
    DROP COLUMN IF EXISTS vector_similarity_threshold,
    DROP COLUMN IF EXISTS vector_max_results;

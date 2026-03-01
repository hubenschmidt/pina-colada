ALTER TABLE "Automation_Config"
    ADD COLUMN vector_prefilter_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN vector_similarity_threshold FLOAT NOT NULL DEFAULT 0.3,
    ADD COLUMN vector_max_results INT NOT NULL DEFAULT 10;

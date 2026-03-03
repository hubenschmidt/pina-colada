-- Switch embedding dimension from OpenAI 3072 to Voyage AI 1024
-- Existing embeddings are incompatible and must be re-generated
TRUNCATE TABLE "Embedding";
ALTER TABLE "Embedding" ALTER COLUMN embedding TYPE vector(1024);

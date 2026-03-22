-- Switch embedding dimension from Voyage AI 1024 back to OpenAI 3072
-- Existing embeddings are incompatible and must be re-generated
TRUNCATE TABLE "Embedding";
ALTER TABLE "Embedding" ALTER COLUMN embedding TYPE vector(3072);

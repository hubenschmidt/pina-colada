-- Add summary and compressed columns to Document table
-- Summary is JSONB for AI-generated summary with metadata
-- Compressed is TEXT with whitespace-trimmed content for efficient full-document scanning

ALTER TABLE "Document"
ADD COLUMN summary JSONB,
ADD COLUMN compressed TEXT;

COMMENT ON COLUMN "Document".summary IS 'AI-generated summary: {text: string, keywords: string[], generated_at: timestamp}';
COMMENT ON COLUMN "Document".compressed IS 'Whitespace-trimmed content for efficient full-document scanning';

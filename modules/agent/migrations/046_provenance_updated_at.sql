-- Add updated_at column to Data_Provenance table
ALTER TABLE "Data_Provenance"
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

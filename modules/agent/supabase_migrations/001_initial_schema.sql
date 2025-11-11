-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the Job table (capitalized) with final schema
CREATE TABLE IF NOT EXISTS "Job" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company TEXT NOT NULL,
    job_title TEXT NOT NULL,
    date TIMESTAMP DEFAULT NOW(),
    status TEXT DEFAULT 'applied' CHECK (status IN ('applied', 'interviewing', 'rejected', 'offer', 'accepted')),
    job_url TEXT,
    notes TEXT,
    resume TIMESTAMP,
    salary_range TEXT,
    source TEXT DEFAULT 'manual' CHECK (source IN ('manual', 'agent')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_company_job_title ON "Job"(company, job_title);
CREATE INDEX IF NOT EXISTS idx_date ON "Job"(date DESC);
CREATE INDEX IF NOT EXISTS idx_status ON "Job"(status);

-- Create function for auto-updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
DROP TRIGGER IF EXISTS update_job_updated_at ON "Job";
CREATE TRIGGER update_job_updated_at
    BEFORE UPDATE ON "Job"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security (RLS) policies
-- Note: Adjust these policies based on your authentication setup
ALTER TABLE "Job" ENABLE ROW LEVEL SECURITY;

-- Allow all operations for authenticated users (adjust as needed)
-- Drop policies if they exist, then create them
DROP POLICY IF EXISTS "Enable read access for all users" ON "Job";
CREATE POLICY "Enable read access for all users"
    ON "Job" FOR SELECT
    USING (true);

DROP POLICY IF EXISTS "Enable insert for all users" ON "Job";
CREATE POLICY "Enable insert for all users"
    ON "Job" FOR INSERT
    WITH CHECK (true);

DROP POLICY IF EXISTS "Enable update for all users" ON "Job";
CREATE POLICY "Enable update for all users"
    ON "Job" FOR UPDATE
    USING (true);

DROP POLICY IF EXISTS "Enable delete for all users" ON "Job";
CREATE POLICY "Enable delete for all users"
    ON "Job" FOR DELETE
    USING (true);

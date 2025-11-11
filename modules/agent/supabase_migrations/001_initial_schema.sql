-- Initial schema for job application tracking
-- Migration: 001_initial_schema.sql
-- Description: Creates applied_jobs table with indexes and triggers

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the applied_jobs table
CREATE TABLE IF NOT EXISTS applied_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company TEXT NOT NULL,
    job_title TEXT NOT NULL,
    application_date TIMESTAMP DEFAULT NOW(),
    status TEXT DEFAULT 'applied' CHECK (status IN ('applied', 'interviewing', 'rejected', 'offer', 'accepted')),
    job_url TEXT,
    location TEXT,
    salary_range TEXT,
    notes TEXT,
    source TEXT DEFAULT 'manual' CHECK (source IN ('manual', 'agent')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_company_title ON applied_jobs(company, job_title);
CREATE INDEX IF NOT EXISTS idx_application_date ON applied_jobs(application_date DESC);
CREATE INDEX IF NOT EXISTS idx_status ON applied_jobs(status);

-- Create function for auto-updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
DROP TRIGGER IF EXISTS update_applied_jobs_updated_at ON applied_jobs;
CREATE TRIGGER update_applied_jobs_updated_at
    BEFORE UPDATE ON applied_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security (RLS) policies
-- Note: Adjust these policies based on your authentication setup
ALTER TABLE applied_jobs ENABLE ROW LEVEL SECURITY;

-- Allow all operations for authenticated users (adjust as needed)
CREATE POLICY IF NOT EXISTS "Enable read access for all users"
    ON applied_jobs FOR SELECT
    USING (true);

CREATE POLICY IF NOT EXISTS "Enable insert for all users"
    ON applied_jobs FOR INSERT
    WITH CHECK (true);

CREATE POLICY IF NOT EXISTS "Enable update for all users"
    ON applied_jobs FOR UPDATE
    USING (true);

CREATE POLICY IF NOT EXISTS "Enable delete for all users"
    ON applied_jobs FOR DELETE
    USING (true);

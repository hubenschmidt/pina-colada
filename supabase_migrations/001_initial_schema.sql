-- Migration: Initial schema for job application tracking
-- Date: 2024-11-10
-- Description: Creates applied_jobs table with indexes, triggers, and RLS policies

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create applied_jobs table
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

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_company_title ON applied_jobs(company, job_title);
CREATE INDEX IF NOT EXISTS idx_application_date ON applied_jobs(application_date DESC);
CREATE INDEX IF NOT EXISTS idx_status ON applied_jobs(status);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_applied_jobs_updated_at
    BEFORE UPDATE ON applied_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Enable Row Level Security (RLS)
ALTER TABLE applied_jobs ENABLE ROW LEVEL SECURITY;

-- Create permissive policies for now (single user)
-- For production with multiple users, restrict to authenticated users only
CREATE POLICY "Enable read access for all users" ON applied_jobs
    FOR SELECT USING (true);

CREATE POLICY "Enable insert access for all users" ON applied_jobs
    FOR INSERT WITH CHECK (true);

CREATE POLICY "Enable update access for all users" ON applied_jobs
    FOR UPDATE USING (true);

CREATE POLICY "Enable delete access for all users" ON applied_jobs
    FOR DELETE USING (true);

-- Grant access to authenticated and anon roles
GRANT ALL ON applied_jobs TO authenticated;
GRANT ALL ON applied_jobs TO anon;

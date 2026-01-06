-- Create schema_seeders table to track which seeders have been applied
CREATE TABLE IF NOT EXISTS schema_seeders (
    seeder_name TEXT PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW()
);

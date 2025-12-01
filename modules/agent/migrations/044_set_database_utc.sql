-- ==============================
-- Set Database Timezone to UTC
-- ==============================
-- Ensures all NOW() calls return UTC timestamps
-- Fresh prod DB - no data conversion needed

-- Set database-level timezone to UTC
ALTER DATABASE postgres SET timezone TO 'UTC';

-- Update the trigger function to explicitly use UTC
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW() AT TIME ZONE 'UTC';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add timezone preference to UserPreferences if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'UserPreferences' AND column_name = 'timezone'
    ) THEN
        ALTER TABLE "UserPreferences" ADD COLUMN timezone TEXT DEFAULT 'America/New_York';
    END IF;
END $$;

-- Verify timezone setting (run manually after migration):
-- SELECT current_setting('TIMEZONE');
-- Expected output: UTC

-- ==============================
-- Seeder: Additional Job Statuses
-- ==============================
-- Adds missing job status options to match frontend expectations

DO $$
BEGIN
  -- Insert "Lead" status if it doesn't exist
  INSERT INTO "Status" (name, description, category, is_terminal, created_at, updated_at)
  VALUES ('Lead', 'Job lead not yet applied to', 'job', FALSE, NOW(), NOW())
  ON CONFLICT (name) DO NOTHING;

  -- Insert "Do Not Apply" status if it doesn't exist
  INSERT INTO "Status" (name, description, category, is_terminal, created_at, updated_at)
  VALUES ('Do Not Apply', 'Job decided not to apply for', 'job', TRUE, NOW(), NOW())
  ON CONFLICT (name) DO NOTHING;

  RAISE NOTICE '✓ Additional job statuses seeded';
END $$;

-- Return count of all job statuses
SELECT COUNT(*) as total_job_statuses FROM "Status" WHERE category = 'job';

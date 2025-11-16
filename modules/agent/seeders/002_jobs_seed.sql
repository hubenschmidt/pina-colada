-- ==============================
-- Seeder: Sample Job Applications
-- ==============================
-- Creates sample organizations, deal, and job applications for local development
-- Run AFTER 001_initial_seed.sql

DO $$
DECLARE
  default_deal_id BIGINT;
  default_deal_status_id BIGINT;
  new_lead_id BIGINT;
  org_id BIGINT;
  job_status_applied_id BIGINT;
  job_status_interviewing_id BIGINT;
  job_status_rejected_id BIGINT;
BEGIN
  -- Get status IDs
  SELECT id INTO default_deal_status_id FROM "Status" WHERE name = 'Prospecting' AND category = 'deal' LIMIT 1;
  SELECT id INTO job_status_applied_id FROM "Status" WHERE name = 'Applied' AND category = 'job' LIMIT 1;
  SELECT id INTO job_status_interviewing_id FROM "Status" WHERE name = 'Interviewing' AND category = 'job' LIMIT 1;
  SELECT id INTO job_status_rejected_id FROM "Status" WHERE name = 'Rejected' AND category = 'job' LIMIT 1;

  -- Create or get default deal for job search
  SELECT id INTO default_deal_id FROM "Deal" WHERE name = 'Job Search 2025' LIMIT 1;

  IF default_deal_id IS NULL THEN
    INSERT INTO "Deal" (name, description, current_status_id, created_at, updated_at)
    VALUES (
      'Job Search 2025',
      'Job search applications for 2025',
      default_deal_status_id,
      NOW(),
      NOW()
    )
    RETURNING id INTO default_deal_id;
    RAISE NOTICE 'Created Deal: Job Search 2025 (id=%)', default_deal_id;
  END IF;

  -- Create sample organizations for jobs (only if they don't exist)
  INSERT INTO "Organization" (tenant_id, name, website, industry, created_at, updated_at)
  VALUES
    (NULL, 'Acme Corp', 'https://acme.example.com', 'Technology', NOW(), NOW()),
    (NULL, 'TechStartup Inc', 'https://techstartup.example.com', 'Software', NOW(), NOW()),
    (NULL, 'DataSystems Ltd', 'https://datasystems.example.com', 'Data Analytics', NOW(), NOW()),
    (NULL, 'CloudWorks', 'https://cloudworks.example.com', 'Cloud Computing', NOW(), NOW()),
    (NULL, 'AI Innovations', 'https://aiinnovations.example.com', 'Artificial Intelligence', NOW(), NOW())
  ON CONFLICT (tenant_id, (LOWER(name))) DO NOTHING;

  -- Job 1: Acme Corp - Senior Full Stack Engineer (Applied)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Acme Corp') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
    VALUES (
      default_deal_id,
      'Job',
      'Acme Corp - Senior Full Stack Engineer',
      'Building next-generation e-commerce platform',
      'manual',
      job_status_applied_id,
      NOW() - INTERVAL '5 days',
      NOW() - INTERVAL '5 days'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, organization_id, job_title, job_url, salary_range, created_at, updated_at)
    VALUES (
      new_lead_id,
      org_id,
      'Senior Full Stack Engineer',
      'https://acme.example.com/careers/senior-fullstack',
      '$120k - $160k',
      NOW() - INTERVAL '5 days',
      NOW() - INTERVAL '5 days'
    );
  END IF;

  -- Job 2: TechStartup Inc - Software Engineer (Interviewing)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('TechStartup Inc') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
    VALUES (
      default_deal_id,
      'Job',
      'TechStartup Inc - Software Engineer',
      'Early-stage startup building developer tools',
      'referral',
      job_status_interviewing_id,
      NOW() - INTERVAL '10 days',
      NOW() - INTERVAL '2 days'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, organization_id, job_title, job_url, salary_range, resume_date, created_at, updated_at)
    VALUES (
      new_lead_id,
      org_id,
      'Software Engineer',
      'https://techstartup.example.com/jobs/swe',
      '$100k - $140k',
      NOW() - INTERVAL '8 days',
      NOW() - INTERVAL '10 days',
      NOW() - INTERVAL '2 days'
    );
  END IF;

  -- Job 3: DataSystems Ltd - Backend Engineer (Applied)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataSystems Ltd') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
    VALUES (
      default_deal_id,
      'Job',
      'DataSystems Ltd - Backend Engineer',
      'Work on high-performance data processing systems',
      'manual',
      job_status_applied_id,
      NOW() - INTERVAL '3 days',
      NOW() - INTERVAL '3 days'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, organization_id, job_title, job_url, salary_range, created_at, updated_at)
    VALUES (
      new_lead_id,
      org_id,
      'Backend Engineer',
      'https://datasystems.example.com/careers/backend',
      '$110k - $150k',
      NOW() - INTERVAL '3 days',
      NOW() - INTERVAL '3 days'
    );
  END IF;

  -- Job 4: CloudWorks - DevOps Engineer (Rejected)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudWorks') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
    VALUES (
      default_deal_id,
      'Job',
      'CloudWorks - DevOps Engineer',
      'Infrastructure and deployment automation',
      'manual',
      job_status_rejected_id,
      NOW() - INTERVAL '15 days',
      NOW() - INTERVAL '7 days'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, organization_id, job_title, job_url, salary_range, created_at, updated_at)
    VALUES (
      new_lead_id,
      org_id,
      'DevOps Engineer',
      'https://cloudworks.example.com/jobs/devops',
      '$115k - $155k',
      NOW() - INTERVAL '15 days',
      NOW() - INTERVAL '7 days'
    );
  END IF;

  -- Job 5: AI Innovations - ML Engineer (Applied)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('AI Innovations') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
    VALUES (
      default_deal_id,
      'Job',
      'AI Innovations - ML Engineer',
      'Building cutting-edge machine learning models',
      'manual',
      job_status_applied_id,
      NOW() - INTERVAL '1 day',
      NOW() - INTERVAL '1 day'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, organization_id, job_title, job_url, salary_range, resume_date, created_at, updated_at)
    VALUES (
      new_lead_id,
      org_id,
      'ML Engineer',
      'https://aiinnovations.example.com/careers/ml',
      '$130k - $180k',
      NOW() - INTERVAL '1 day',
      NOW() - INTERVAL '1 day',
      NOW() - INTERVAL '1 day'
    );
  END IF;

  RAISE NOTICE '✓ Sample job applications seeded successfully';
END $$;

-- Return count of jobs created
SELECT COUNT(*) as jobs_seeded FROM "Job"
WHERE organization_id IN (
  SELECT id FROM "Organization"
  WHERE LOWER(name) IN (
    LOWER('Acme Corp'),
    LOWER('TechStartup Inc'),
    LOWER('DataSystems Ltd'),
    LOWER('CloudWorks'),
    LOWER('AI Innovations')
  )
);

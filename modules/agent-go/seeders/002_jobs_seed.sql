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
  account_id BIGINT;
  v_tenant_id BIGINT;
  v_user_id BIGINT;
  job_status_applied_id BIGINT;
  job_status_interviewing_id BIGINT;
  job_status_rejected_id BIGINT;
BEGIN
  -- Get tenant ID and bootstrap user
  SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;
  SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

  -- Get status IDs
  SELECT id INTO default_deal_status_id FROM "Status" WHERE name = 'Prospecting' AND category = 'deal' LIMIT 1;
  SELECT id INTO job_status_applied_id FROM "Status" WHERE name = 'Applied' AND category = 'job' LIMIT 1;
  SELECT id INTO job_status_interviewing_id FROM "Status" WHERE name = 'Interviewing' AND category = 'job' LIMIT 1;
  SELECT id INTO job_status_rejected_id FROM "Status" WHERE name = 'Rejected' AND category = 'job' LIMIT 1;

  -- Create or get default deal for job search
  SELECT id INTO default_deal_id FROM "Deal" WHERE name = 'Job Search 2025' LIMIT 1;

  IF default_deal_id IS NULL THEN
    INSERT INTO "Deal" (tenant_id, name, description, current_status_id, created_by, updated_by, created_at, updated_at)
    VALUES (
      v_tenant_id,
      'Job Search 2025',
      'Job search applications for 2025',
      default_deal_status_id,
      v_user_id,
      v_user_id,
      NOW(),
      NOW()
    )
    RETURNING id INTO default_deal_id;
    RAISE NOTICE 'Created Deal: Job Search 2025 (id=%)', default_deal_id;
  END IF;

  -- Create sample organizations for jobs (with Accounts)
  -- Acme Corp
  INSERT INTO "Account" (tenant_id, name, created_by, updated_by, created_at, updated_at) VALUES (v_tenant_id, 'Acme Corp', v_user_id, v_user_id, NOW(), NOW()) RETURNING id INTO account_id;
  INSERT INTO "Organization" (account_id, name, website, created_by, updated_by, created_at, updated_at)
  VALUES (account_id, 'Acme Corp', 'https://acme.example.com', v_user_id, v_user_id, NOW(), NOW())
  ON CONFLICT ((LOWER(name))) DO NOTHING;

  -- TechStartup Inc
  INSERT INTO "Account" (tenant_id, name, created_by, updated_by, created_at, updated_at) VALUES (v_tenant_id, 'TechStartup Inc', v_user_id, v_user_id, NOW(), NOW()) RETURNING id INTO account_id;
  INSERT INTO "Organization" (account_id, name, website, created_by, updated_by, created_at, updated_at)
  VALUES (account_id, 'TechStartup Inc', 'https://techstartup.example.com', v_user_id, v_user_id, NOW(), NOW())
  ON CONFLICT ((LOWER(name))) DO NOTHING;

  -- DataSystems Ltd
  INSERT INTO "Account" (tenant_id, name, created_by, updated_by, created_at, updated_at) VALUES (v_tenant_id, 'DataSystems Ltd', v_user_id, v_user_id, NOW(), NOW()) RETURNING id INTO account_id;
  INSERT INTO "Organization" (account_id, name, website, created_by, updated_by, created_at, updated_at)
  VALUES (account_id, 'DataSystems Ltd', 'https://datasystems.example.com', v_user_id, v_user_id, NOW(), NOW())
  ON CONFLICT ((LOWER(name))) DO NOTHING;

  -- CloudWorks
  INSERT INTO "Account" (tenant_id, name, created_by, updated_by, created_at, updated_at) VALUES (v_tenant_id, 'CloudWorks', v_user_id, v_user_id, NOW(), NOW()) RETURNING id INTO account_id;
  INSERT INTO "Organization" (account_id, name, website, created_by, updated_by, created_at, updated_at)
  VALUES (account_id, 'CloudWorks', 'https://cloudworks.example.com', v_user_id, v_user_id, NOW(), NOW())
  ON CONFLICT ((LOWER(name))) DO NOTHING;

  -- AI Innovations
  INSERT INTO "Account" (tenant_id, name, created_by, updated_by, created_at, updated_at) VALUES (v_tenant_id, 'AI Innovations', v_user_id, v_user_id, NOW(), NOW()) RETURNING id INTO account_id;
  INSERT INTO "Organization" (account_id, name, website, created_by, updated_by, created_at, updated_at)
  VALUES (account_id, 'AI Innovations', 'https://aiinnovations.example.com', v_user_id, v_user_id, NOW(), NOW())
  ON CONFLICT ((LOWER(name))) DO NOTHING;

  -- Job 1: Acme Corp - Senior Full Stack Engineer (Applied)
  SELECT o.id, o.account_id INTO org_id, account_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('Acme Corp') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (tenant_id, account_id, deal_id, type, title, description, source, current_status_id, created_by, updated_by, created_at, updated_at)
    VALUES (
      v_tenant_id,
      account_id,
      default_deal_id,
      'Job',
      'Acme Corp - Senior Full Stack Engineer',
      'Building next-generation e-commerce platform',
      'manual',
      job_status_applied_id,
      v_user_id,
      v_user_id,
      NOW() - INTERVAL '5 days',
      NOW() - INTERVAL '5 days'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, job_title, job_url, salary_range, created_at, updated_at)
    VALUES (
      new_lead_id,
      'Senior Full Stack Engineer',
      'https://acme.example.com/careers/senior-fullstack',
      '$120k - $160k',
      NOW() - INTERVAL '5 days',
      NOW() - INTERVAL '5 days'
    );
  END IF;

  -- Job 2: TechStartup Inc - Software Engineer (Interviewing)
  SELECT o.id, o.account_id INTO org_id, account_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('TechStartup Inc') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (tenant_id, account_id, deal_id, type, title, description, source, current_status_id, created_by, updated_by, created_at, updated_at)
    VALUES (
      v_tenant_id,
      account_id,
      default_deal_id,
      'Job',
      'TechStartup Inc - Software Engineer',
      'Early-stage startup building developer tools',
      'referral',
      job_status_interviewing_id,
      v_user_id,
      v_user_id,
      NOW() - INTERVAL '10 days',
      NOW() - INTERVAL '2 days'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, job_title, job_url, salary_range, resume_date, created_at, updated_at)
    VALUES (
      new_lead_id,
      'Software Engineer',
      'https://techstartup.example.com/jobs/swe',
      '$100k - $140k',
      NOW() - INTERVAL '8 days',
      NOW() - INTERVAL '10 days',
      NOW() - INTERVAL '2 days'
    );
  END IF;

  -- Job 3: DataSystems Ltd - Backend Engineer (Applied)
  SELECT o.id, o.account_id INTO org_id, account_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('DataSystems Ltd') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (tenant_id, account_id, deal_id, type, title, description, source, current_status_id, created_by, updated_by, created_at, updated_at)
    VALUES (
      v_tenant_id,
      account_id,
      default_deal_id,
      'Job',
      'DataSystems Ltd - Backend Engineer',
      'Work on high-performance data processing systems',
      'manual',
      job_status_applied_id,
      v_user_id,
      v_user_id,
      NOW() - INTERVAL '3 days',
      NOW() - INTERVAL '3 days'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, job_title, job_url, salary_range, created_at, updated_at)
    VALUES (
      new_lead_id,
      'Backend Engineer',
      'https://datasystems.example.com/careers/backend',
      '$110k - $150k',
      NOW() - INTERVAL '3 days',
      NOW() - INTERVAL '3 days'
    );
  END IF;

  -- Job 4: CloudWorks - DevOps Engineer (Rejected)
  SELECT o.id, o.account_id INTO org_id, account_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('CloudWorks') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (tenant_id, account_id, deal_id, type, title, description, source, current_status_id, created_by, updated_by, created_at, updated_at)
    VALUES (
      v_tenant_id,
      account_id,
      default_deal_id,
      'Job',
      'CloudWorks - DevOps Engineer',
      'Infrastructure and deployment automation',
      'manual',
      job_status_rejected_id,
      v_user_id,
      v_user_id,
      NOW() - INTERVAL '15 days',
      NOW() - INTERVAL '7 days'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, job_title, job_url, salary_range, created_at, updated_at)
    VALUES (
      new_lead_id,
      'DevOps Engineer',
      'https://cloudworks.example.com/jobs/devops',
      '$115k - $155k',
      NOW() - INTERVAL '15 days',
      NOW() - INTERVAL '7 days'
    );
  END IF;

  -- Job 5: AI Innovations - ML Engineer (Applied)
  SELECT o.id, o.account_id INTO org_id, account_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('AI Innovations') LIMIT 1;
  IF org_id IS NOT NULL THEN
    INSERT INTO "Lead" (tenant_id, account_id, deal_id, type, title, description, source, current_status_id, created_by, updated_by, created_at, updated_at)
    VALUES (
      v_tenant_id,
      account_id,
      default_deal_id,
      'Job',
      'AI Innovations - ML Engineer',
      'Building cutting-edge machine learning models',
      'manual',
      job_status_applied_id,
      v_user_id,
      v_user_id,
      NOW() - INTERVAL '1 day',
      NOW() - INTERVAL '1 day'
    )
    RETURNING id INTO new_lead_id;

    INSERT INTO "Job" (id, job_title, job_url, salary_range, resume_date, created_at, updated_at)
    VALUES (
      new_lead_id,
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
SELECT COUNT(*) as jobs_seeded FROM "Job" j
JOIN "Lead" l ON j.id = l.id
JOIN "Account" a ON l.account_id = a.id
JOIN "Organization" o ON a.id = o.account_id
WHERE LOWER(o.name) IN (
  LOWER('Acme Corp'),
  LOWER('TechStartup Inc'),
  LOWER('DataSystems Ltd'),
  LOWER('CloudWorks'),
  LOWER('AI Innovations')
);

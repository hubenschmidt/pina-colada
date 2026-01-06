-- ==============================
-- Seeder: Bulk Job Records (1000)
-- ==============================
-- Creates 1000 job records for testing batch operations and agent capabilities
-- Run AFTER 002_jobs_seed.sql

DO $$
DECLARE
  v_tenant_id BIGINT;
  v_user_id BIGINT;
  v_deal_id BIGINT;
  v_status_id BIGINT;
  v_lead_id BIGINT;
  v_counter INT := 0;
  v_titles TEXT[] := ARRAY[
    'Software Engineer', 'Senior Software Engineer', 'Staff Engineer', 'Principal Engineer',
    'Frontend Developer', 'Backend Developer', 'Full Stack Developer', 'DevOps Engineer',
    'Site Reliability Engineer', 'Platform Engineer', 'Infrastructure Engineer',
    'Data Engineer', 'ML Engineer', 'AI Engineer', 'Data Scientist',
    'Engineering Manager', 'Tech Lead', 'Architect', 'Solutions Architect',
    'Product Manager', 'Technical Product Manager', 'Program Manager',
    'QA Engineer', 'SDET', 'Security Engineer', 'Cloud Engineer'
  ];
  v_levels TEXT[] := ARRAY['Junior', 'Mid-Level', 'Senior', 'Staff', 'Principal', 'Lead'];
  v_companies TEXT[] := ARRAY[
    'TechCorp', 'DataFlow', 'CloudBase', 'AILabs', 'DevTools', 'ScaleUp',
    'ByteWorks', 'CodeHive', 'NetSphere', 'QuantumSoft', 'CyberSec', 'AppDynamics',
    'StreamData', 'EdgeCloud', 'BlockTech', 'AutomateAI', 'VirtualOps', 'SmartSystems'
  ];
  v_title TEXT;
  v_company TEXT;
  v_level TEXT;
  v_salary_min INT;
  v_salary_max INT;
BEGIN
  -- Get tenant ID and bootstrap user
  SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;
  SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;
  SELECT id INTO v_deal_id FROM "Deal" WHERE name = 'Job Search 2025' LIMIT 1;
  SELECT id INTO v_status_id FROM "Status" WHERE name = 'Applied' AND category = 'job' LIMIT 1;

  IF v_tenant_id IS NULL OR v_user_id IS NULL OR v_deal_id IS NULL OR v_status_id IS NULL THEN
    RAISE EXCEPTION 'Required records not found. Run 001_initial_seed.sql and 002_jobs_seed.sql first.';
  END IF;

  -- Generate 1000 jobs
  FOR i IN 1..1000 LOOP
    -- Pick random title, level, company
    v_title := v_titles[1 + floor(random() * array_length(v_titles, 1))::int];
    v_level := v_levels[1 + floor(random() * array_length(v_levels, 1))::int];
    v_company := v_companies[1 + floor(random() * array_length(v_companies, 1))::int];

    -- Generate salary range based on level
    v_salary_min := CASE v_level
      WHEN 'Junior' THEN 70000 + floor(random() * 20000)::int
      WHEN 'Mid-Level' THEN 90000 + floor(random() * 30000)::int
      WHEN 'Senior' THEN 130000 + floor(random() * 40000)::int
      WHEN 'Staff' THEN 160000 + floor(random() * 50000)::int
      WHEN 'Principal' THEN 200000 + floor(random() * 60000)::int
      WHEN 'Lead' THEN 180000 + floor(random() * 50000)::int
      ELSE 100000
    END;
    v_salary_max := v_salary_min + 30000 + floor(random() * 20000)::int;

    -- Create Lead record
    INSERT INTO "Lead" (
      tenant_id, account_id, deal_id, type, source,
      current_status_id, created_by, updated_by, created_at, updated_at
    )
    VALUES (
      v_tenant_id,
      NULL,  -- No account (optional now)
      v_deal_id,
      'Job',
      'seed',
      v_status_id,
      v_user_id,
      v_user_id,
      NOW() - (random() * INTERVAL '90 days'),
      NOW() - (random() * INTERVAL '30 days')
    )
    RETURNING id INTO v_lead_id;

    -- Create Job record
    INSERT INTO "Job" (id, job_title, job_url, salary_range, description, created_at, updated_at)
    VALUES (
      v_lead_id,
      v_level || ' ' || v_title,
      'https://' || lower(v_company) || '.example.com/careers/' || i,
      '$' || (v_salary_min / 1000)::text || 'k - $' || (v_salary_max / 1000)::text || 'k',
      'Role at ' || v_company || '. Looking for a ' || v_level || ' ' || v_title || ' to join our team.',
      NOW() - (random() * INTERVAL '90 days'),
      NOW() - (random() * INTERVAL '30 days')
    );

    v_counter := v_counter + 1;

    -- Log progress every 100 records
    IF v_counter % 100 = 0 THEN
      RAISE NOTICE 'Created % jobs...', v_counter;
    END IF;
  END LOOP;

  RAISE NOTICE '✓ Created % bulk job records', v_counter;
END $$;

-- Return count
SELECT COUNT(*) as total_jobs FROM "Job";

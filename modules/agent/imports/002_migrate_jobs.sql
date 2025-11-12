-- ==============================
-- Production Data Migration: Migrate Job Records
-- ==============================
-- Run this AFTER 001_import_orgs.sql
-- This migrates all 188 job records from old format to new DealTracker schema
-- ==============================

DO $$
DECLARE
  default_deal_id BIGINT;
  default_deal_status_id BIGINT;
  new_lead_id BIGINT;
  org_id BIGINT;
  mapped_status_id BIGINT;
BEGIN
  -- Get default deal status for "Prospecting"
  SELECT id INTO default_deal_status_id
  FROM "Status"
  WHERE name = 'Prospecting' AND category = 'deal'
  LIMIT 1;

  -- Get or create the default deal
  SELECT id INTO default_deal_id
  FROM "Deal"
  WHERE name = 'Job Search 2025'
  LIMIT 1;

  IF default_deal_id IS NULL THEN
    INSERT INTO "Deal" (name, description, current_status_id, created_at, updated_at)
    VALUES (
      'Job Search 2025',
      'Job search applications for 2025',
      default_deal_status_id,
      '2025-11-11 11:22:46',
      NOW()
    )
    RETURNING id INTO default_deal_id;
    RAISE NOTICE 'Created Deal: Job Search 2025 (id=%)', default_deal_id;
  ELSE
    RAISE NOTICE 'Using existing Deal: Job Search 2025 (id=%)', default_deal_id;
  END IF;

  -- Job 1: Poly AI - Senior Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Poly AI');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Poly AI - Senior Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Full Stack Engineer', 'https://job-boards.eu.greenhouse.io/polyai/jobs/4669020101', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 2: Crustdata - Founding Solutions Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Crustdata');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Crustdata - Founding Solutions Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Founding Solutions Engineer', 'https://www.ycombinator.com/companies/crustdata/jobs/Bdu5Rez-founding-solutions-engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 3: city of new york - IT Automation and Monitoring Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('city of new york');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'city of new york - IT Automation and Monitoring Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'IT Automation and Monitoring Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 4: zeta global - solutions engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('zeta global');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'zeta global - solutions engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'solutions engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 5: linda werner - technical solutions consultant
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('linda werner');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'linda werner - technical solutions consultant', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'technical solutions consultant', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 6: kofas northamerica insurance co. - tech solutions engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('kofas northamerica insurance co.');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'kofas northamerica insurance co. - tech solutions engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'tech solutions engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 7: Peloton - Software Engineer II
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Peloton');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Peloton - Software Engineer II', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer II', 'https://careers.onepeloton.com/en/all-jobs/7360532/software-engineer-ii/?gh_jid=7360532&utm_source=startup.jobs&utm_medium=organic#apply-now', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 8: Peloton - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Peloton');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Peloton - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://careers.onepeloton.com/en/all-jobs/7360403/software-engineer/?gh_jid=7360403&utm_source=startup.jobs&utm_medium=organic#apply-now', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 9: Consumer Edge - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Consumer Edge');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Consumer Edge - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://job-boards.greenhouse.io/consumeredge/jobs/5688362004', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 10: ZocDoc - Senior Software Engineer, Online Booking
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('ZocDoc');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'ZocDoc - Senior Software Engineer, Online Booking', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Online Booking', 'https://startup.jobs/senior-software-engineer-online-booking-zocdoc-7414445', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 11: Quizlet - Senior Software Engineer, Consumer Experience
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Quizlet');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Quizlet - Senior Software Engineer, Consumer Experience', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Consumer Experience', 'https://startup.jobs/sr-software-engineer-consumer-experience-quizlet-7414960', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 12: BitGo - Senior Software Engineer, Wallet Core
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('BitGo');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'BitGo - Senior Software Engineer, Wallet Core', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Wallet Core', 'https://job-boards.greenhouse.io/bitgo/jobs/8231309002?utm_source=startup.jobs&utm_medium=organic#application_form', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 13: Capitalize - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Capitalize');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Capitalize - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://www.hicapitalize.com/careers/opportunity/?gh_jid=5690067004&utm_source=startup.jobs&utm_medium=organic#application_form', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 14: Vestwell - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Vestwell');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Vestwell - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://startup.jobs/senior-software-engineer-vestwell-2-7419853', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 15: Vesto - founding engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Vesto');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Vesto - founding engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'founding engineer', 'https://getvesto.notion.site/Founding-Full-stack-Engineer-Vesto-4d579a18be674538878556b4c6c417af', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 16: Western & Southern Financial - web engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Western & Southern Financial');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Western & Southern Financial - web engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'web engineer', 'https://www.indeed.com/viewjob?jk=9433960d325f0eb9&from=shareddesktop_copy', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 17: Bronx Defenders - Solutions Architect
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Bronx Defenders');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Bronx Defenders - Solutions Architect', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Solutions Architect', 'https://www.idealist.org/en/nonprofit-job/973347ef524747a6a3a449ae7703c850-solutions-architect-the-bronx-defenders-bronx#apply', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 18: Hudson Valley Credit Union - Solution Architect
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Hudson Valley Credit Union');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Hudson Valley Credit Union - Solution Architect', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Solution Architect', 'https://careers.hvcu.org/jobs/5104?lang=en-us', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 19: Hudson Valley Credit Union - Sr Cloud Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Hudson Valley Credit Union');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Hudson Valley Credit Union - Sr Cloud Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Sr Cloud Engineer', 'https://careers-hvcu.icims.com/jobs/5076/sr-cloud-engineer/job?mode=submit_apply', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 20: IBM - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('IBM');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'IBM - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://ibmglobal.avature.net/en_US/careers/JobDetailApplied?jobId=33412&qtvc=059fae09d7a632343173d9226a89954915999ac31c625c4e9212583bb337a762', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 21: Pryon - Software Engineer, Fullstack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Pryon');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Pryon - Software Engineer, Fullstack', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer, Fullstack', 'https://jobs.lever.co/pryon/15412b82-8bf6-4b25-820d-9f129828d796?lever-source=Otta', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 22: Nourish - Senior or Staff Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Nourish');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Nourish - Senior or Staff Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior or Staff Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 23: Adonis.io - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Adonis.io');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Adonis.io - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 24: Harmonic - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Harmonic');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Harmonic - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 25: Asana - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Asana');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Asana - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 26: Cooperidge Consulting - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Cooperidge Consulting');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Cooperidge Consulting - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 27: Coast - Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Coast');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Coast - Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 28: April - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('April');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'April - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 29: Crosby - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Crosby');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Crosby - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 30: Better - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Better');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Better - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 31: Headspace - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Headspace');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Headspace - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 32: Harvey - Staff Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Harvey');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Harvey - Staff Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Staff Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 33: Traversal - Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Traversal');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Traversal - Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 34: Brex - Software Enginer II
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Brex');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Brex - Software Enginer II', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Enginer II', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 35: Together AI - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Together AI');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Together AI - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 36: Neem - Founding Senior Full-Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Neem');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Neem - Founding Senior Full-Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Founding Senior Full-Stack Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 37: ShelfCycle - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('ShelfCycle');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'ShelfCycle - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://shelfcycle.com/careers/senior-software-engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 38: On Me - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('On Me');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'On Me - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 39: Cassidy - Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Cassidy');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Cassidy - Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 40: keru.ai - Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('keru.ai');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'keru.ai - Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 41: Norm.AI - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Norm.AI');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Norm.AI - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 42: Vestwell - Senior Software Engineer (AI)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Vestwell');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Vestwell - Senior Software Engineer (AI)', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer (AI)', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 43: Harold - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Harold');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Harold - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 44: Farther - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Farther');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Farther - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 45: Parabola - Senior Software Engineer, Full Stack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Parabola');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Parabola - Senior Software Engineer, Full Stack', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Full Stack', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 46: Brisk Teaching - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Brisk Teaching');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Brisk Teaching - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 47: sonar hiring - senior software engineer fullstack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('sonar hiring');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Interviewing' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'sonar hiring - senior software engineer fullstack', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 17:37:27')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'senior software engineer fullstack', '2025-11-11 11:22:46', '2025-11-11 17:37:27');

  -- Job 48: Stepful - Senior Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Stepful');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Stepful - Senior Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Full Stack Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 49: Miro - Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Miro');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Miro - Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Engineer', 'https://miro.com/careers/vacancy/8013376002/', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 50: Affirma Group - Mid Level App Developer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Affirma Group');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Affirma Group - Mid Level App Developer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Mid Level App Developer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  RAISE NOTICE 'Migrated 50 jobs...';

  -- Job 51: Airgoods - Software Engineer - Full Stack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Airgoods');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Airgoods - Software Engineer - Full Stack', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer - Full Stack', 'https://jobs.ashbyhq.com/airgoods/036e4ed9-9c41-4aaa-9219-3851fb02f8e3', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 52: Method Financial - Technical Integration Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Method Financial');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Method Financial - Technical Integration Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Technical Integration Engineer', 'https://jobs.ashbyhq.com/method/5dde1ef0-4df6-4209-a60c-f966d446320b', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 53: Clarion - Senior Fullstack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Clarion');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Clarion - Senior Fullstack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Fullstack Engineer', 'https://jobs.ashbyhq.com/clarion/acc28862-14e1-43bd-bfc8-449af1fa283a', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 54: Peregrine - Senior Fullstack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Peregrine');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Peregrine - Senior Fullstack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Fullstack Engineer', 'https://job-boards.greenhouse.io/peregrinetechnologies/jobs/4580135005?gh_src=gwf63pot5us', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 55: Jacobsen Consulting Applications - Product Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Jacobsen Consulting Applications');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Jacobsen Consulting Applications - Product Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Product Engineer', 'https://www.idealist.org/en/consultant-job/e3d9c6b05c5c427c82f419397db8fa9f-product-engineer-jacobson-consulting-applications-new-york', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 56: DataKind - Sr Director, Engineering
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataKind');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'DataKind - Sr Director, Engineering', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Sr Director, Engineering', 'https://job-boards.greenhouse.io/datakindinc/jobs/6586200003?utm_medium=referral&utm_source=idealist', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 57: Friends From The City - Technical Lead, Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Friends From The City');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Friends From The City - Technical Lead, Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Technical Lead, Full Stack Engineer', 'https://apply.workable.com/friends-from-the-city/j/B442D71A20/', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 58: ContentSquare - Backend Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('ContentSquare');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'ContentSquare - Backend Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Backend Software Engineer', 'https://jobs.lever.co/contentsquare/048f9fc4-0b8f-4da2-9388-51751c9721d7', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 59: B&H - Senior Web Project Manager
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('B&H');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'B&H - Senior Web Project Manager', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Web Project Manager', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 60: Braze - Senior Software Engineer - Fullstack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Braze');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Braze - Senior Software Engineer - Fullstack', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer - Fullstack', 'https://job-boards.greenhouse.io/braze/jobs/7365359?gh_jid=7365359', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 61: Semgrep - Staff Full Stack Engineer, Software Composition Analysis
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Semgrep');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Semgrep - Staff Full Stack Engineer, Software Composition Analysis', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Staff Full Stack Engineer, Software Composition Analysis', 'https://job-boards.greenhouse.io/semgrep/jobs/4961381007', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 62: Replica - Software Engineer (Execution Focused)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Replica');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Replica - Software Engineer (Execution Focused)', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer (Execution Focused)', 'https://replicainc.applytojob.com/apply/confirm/7njkiRsLA2', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 63: Bumble - Lead Software Engineer (BFF)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Bumble');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Bumble - Lead Software Engineer (BFF)', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Lead Software Engineer (BFF)', 'https://hiring.cafe/viewjob/86wick766qbckbdc', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 64: DeepL - Senior Software Engineer API
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DeepL');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'DeepL - Senior Software Engineer API', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer API', 'https://jobs.ashbyhq.com/deepl/f4cdb92c-524a-425a-b04b-720d21c6ba86', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 65: Replit - Premium Support Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Replit');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Replit - Premium Support Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Premium Support Engineer', 'https://jobs.ashbyhq.com/replit/7dae8c96-d8ff-42be-a099-8f370c1db3db', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 66: Navan - Staff Software Engineer, Security, Risk & Fraud
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Navan');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Navan - Staff Software Engineer, Security, Risk & Fraud', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Staff Software Engineer, Security, Risk & Fraud', 'https://navan.com/careers/openings/7379198?gh_jid=7379198', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 67: Okta - Senior Software Engineer Sessions (Auth0)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Okta');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Okta - Senior Software Engineer Sessions (Auth0)', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer Sessions (Auth0)', 'https://www.okta.com/company/careers/engineering/senior-software-engineer-sessions-auth0-7302752/', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 68: Melio - Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Melio');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Melio - Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Engineer', 'https://job-boards.greenhouse.io/melio/jobs/7174896003', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 69: Xogene, LLC - AI Software Developer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Xogene, LLC');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Xogene, LLC - AI Software Developer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'AI Software Developer', 'https://www.xogene.com/careers#job-2263505', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 70: Patreon - Senior Backend Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Patreon');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Patreon - Senior Backend Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Backend Engineer', 'https://jobs.ashbyhq.com/patreon/392fa521-141c-4b1d-b5e6-01b1d60be3e2', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 71: New York Post - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('New York Post');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'New York Post - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://dowjones.wd1.myworkdayjobs.com/New_York_Post_Careers/job/NYC---1211-Ave-of-the-Americas/Software-Engineer_Job_Req_50029', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 72: Inkeep - Member of the Technical Staff, TypeScript Engineer (Principal)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Inkeep');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Inkeep - Member of the Technical Staff, TypeScript Engineer (Principal)', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Member of the Technical Staff, TypeScript Engineer (Principal)', 'https://jobs.ashbyhq.com/inkeep/7bf64445-f811-4e05-8ad1-06549299597f', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 73: Amplify Education Inc. - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Amplify Education Inc.');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Amplify Education Inc. - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://amplify.wd1.myworkdayjobs.com/en-US/Amplify_Careers/details/Software-Engineer_Req_12320', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 74: Siro - Full Stack Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Siro');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Siro - Full Stack Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Software Engineer', 'https://jobs.ashbyhq.com/siro/3eb43837-5422-4c93-88ae-efb3fbb6f37c', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 75: Arc - Software Engineer, Fullstack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Arc');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Arc - Software Engineer, Fullstack', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer, Fullstack', 'https://jobs.ashbyhq.com/joinarc/7a975960-40f5-4ae4-b96b-6d9046dc2d1d', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 76: Courier Health - Senior Software Engineer, Backend
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Courier Health');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Courier Health - Senior Software Engineer, Backend', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Backend', 'https://job-boards.greenhouse.io/courierhealth/jobs/4737873007', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 77: Tarte - eCommerce Integration Specialist
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Tarte');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Tarte - eCommerce Integration Specialist', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'eCommerce Integration Specialist', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 78: Regrello - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Regrello');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Regrello - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://jobs.lever.co/regrello/3ec9f439-0700-453a-b33b-568c9ef15571', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 79: Rokt - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Rokt');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Rokt - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://apply.workable.com/rokt/j/59FE73F3A9/', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 80: Christine Valmy School - CRM Developer/Consultant
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Christine Valmy School');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Christine Valmy School - CRM Developer/Consultant', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'CRM Developer/Consultant', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 81: Baseten - Senior Software Engineer, Core Product
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Baseten');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Baseten - Senior Software Engineer, Core Product', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Core Product', 'https://jobs.ashbyhq.com/baseten/ebbcec9f-e147-4bef-a181-0b89045c1ec1', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 82: Scale GP - Senior Software Engineer, Full-Stack - Enterprise Gen AI
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Scale GP');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Scale GP - Senior Software Engineer, Full-Stack - Enterprise Gen AI', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Full-Stack - Enterprise Gen AI', 'https://job-boards.greenhouse.io/scaleai/jobs/4529529005#app', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 83: Viam - Backend Software Engineer, Netcore
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Viam');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Viam - Backend Software Engineer, Netcore', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Backend Software Engineer, Netcore', 'https://job-boards.greenhouse.io/viamrobotics/jobs/5587594004', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 84: Hearth - Senior Software Engineer, Backend
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Hearth');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Interviewing' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Hearth - Senior Software Engineer, Backend', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 17:39:18')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Backend', 'https://ats.rippling.com/hearth-careers/jobs/83471fcd-a0ec-4d7e-a994-2e1cffbc9f15', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 17:39:18');

  -- Job 85: DataDog - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataDog');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Rejected' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'DataDog - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 23:02:42')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://careers.datadoghq.com/detail/3851935/?gh_jid=3851935', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 23:02:42');

  -- Job 86: LogRocket - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('LogRocket');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'LogRocket - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://jobs.lever.co/logrocket/0aa8b60f-1a74-4e53-b23a-8d5dc7c497a8', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 87: Tennr - Backend Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Tennr');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Tennr - Backend Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Backend Software Engineer', 'https://jobs.ashbyhq.com/tennr/9aae43bf-3303-468e-aae0-038b7fb395f3', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 88: Peloton - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Peloton');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Peloton - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://portfoliojobs.tcv.com/companies/peloton-3/jobs/55997056-senior-software-engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 89: Paramount - Lead Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Paramount');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Paramount - Lead Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Lead Software Engineer', 'https://careers.paramount.com/job/New-York-Lead-Software-Engineer-NY-10036/1306410600/', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 90: Modern Life - Software Engineer - Full Stack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Modern Life');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Modern Life - Software Engineer - Full Stack', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer - Full Stack', 'https://jobs.lever.co/modernlife/48ab78f8-a264-4960-9fb1-b2e8f8a52b3a', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 91: Kale - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Kale');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Kale - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 92: Crosby - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Crosby');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Crosby - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 93: Applied Labs - Founding Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Applied Labs');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Applied Labs - Founding Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Founding Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 94: Miter - Software Engineer, INtegrations
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Miter');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Miter - Software Engineer, INtegrations', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer, INtegrations', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 95: Etsy - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Etsy');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Etsy - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 96: Adonis - Fullstack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Adonis');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Adonis - Fullstack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Fullstack Engineer', 'https://www.adonis.io/job?gh_jid=4247324007', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 97: GlossGenius - Software Engineer - All levels
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('GlossGenius');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'GlossGenius - Software Engineer - All levels', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer - All levels', 'https://job-boards.greenhouse.io/glossgenius/jobs/6681936003', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 98: Graphite - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Graphite');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Graphite - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://jobs.ashbyhq.com/graphite/81900333-3f22-442c-9b0d-b8c52a928ff3', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 99: Graphite - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Graphite');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Graphite - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://jobs.ashbyhq.com/graphite/5379d006-abe5-4fbf-a459-4928245da6bb', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 100: Tavus - Solutions Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Tavus');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Tavus - Solutions Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Solutions Engineer', 'https://jobs.ashbyhq.com/tavus/6aefde62-2113-40c2-ac36-14da660bd6a3', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  RAISE NOTICE 'Migrated 100 jobs...';

  -- Job 101: ShiftSmart - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('ShiftSmart');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'ShiftSmart - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://shiftsmart.com/apply?ashby_jid=d7e843aa-0f1d-4c8f-a430-ffb0320aa02b#jobs', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 102: Persona - Senior Software Engineer, Product
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Persona');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Persona - Senior Software Engineer, Product', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Product', 'https://jobs.ashbyhq.com/persona/3859c7d2-071f-42a0-b585-258ab3854242?locationId=8963f399-144b-4d2a-b755-9b7279944cb4', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 103: Versana - Full-Stack Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Versana');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Versana - Full-Stack Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full-Stack Software Engineer', 'https://versana.io/career-opportunities/', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 104: Torch - Lead Software Engineer, Product
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Torch');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Torch - Lead Software Engineer, Product', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Lead Software Engineer, Product', 'https://jobs.lever.co/torchdental/0d558379-9450-4c54-b282-4b3cc91f7fe4', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 105: Capitalize - Integrations Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Capitalize');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Capitalize - Integrations Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Integrations Engineer', 'https://www.hicapitalize.com/careers/opportunity/?gh_jid=5669079004', '0025-10-30 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 106: Superblocks - Solutions Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Superblocks');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Superblocks - Solutions Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Solutions Engineer', 'https://jobs.gem.com/superblocks/4024365005', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 107: Meridian - Founding Backend Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Meridian');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Meridian - Founding Backend Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Founding Backend Engineer', 'https://jobs.ashbyhq.com/meridian/5bc5c0de-080f-4796-b3a9-6f5139abd671', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 108: Axion Ray - Senior Software Engineer, CoreApp
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Axion Ray');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Axion Ray - Senior Software Engineer, CoreApp', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, CoreApp', 'https://jobs.ashbyhq.com/axionray/9b7060af-0dcd-4b67-bbd1-05b4ffe94917', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 109: Axion Ray - Software Engineer, Fullstack, CoreApp
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Axion Ray');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Axion Ray - Software Engineer, Fullstack, CoreApp', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer, Fullstack, CoreApp', 'https://jobs.ashbyhq.com/axionray/0f5a3255-339b-4532-9fef-a33809477553', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 110: Sigma Computing - Senior Software Engineer - Backend
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Sigma Computing');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Sigma Computing - Senior Software Engineer - Backend', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer - Backend', 'https://job-boards.greenhouse.io/sigmacomputing/jobs/7523613003', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 111: Hudson River Trading - Experienced Software Engineer, Systems
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Hudson River Trading');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Hudson River Trading - Experienced Software Engineer, Systems', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Experienced Software Engineer, Systems', 'https://www.hudsonrivertrading.com/hrt-job/experienced-software-engineer-systems-5/?gh_src=ca07bf8d1us', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 112: American Equity Investment Life Insurance Co. - Senior Application Developer, Optimizely Expert
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('American Equity Investment Life Insurance Co.');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'American Equity Investment Life Insurance Co. - Senior Application Developer, Optimizely Expert', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Application Developer, Optimizely Expert', 'https://www.american-equity.com/about/careers/openings?gh_jid=4800008007&gh_src=4pb0emxe7us', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 113: SquareSpace - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('SquareSpace');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'SquareSpace - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://www.squarespace.com/careers/jobs/7243239', '2025-11-05 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 114: Particle Health - API/Backend Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Particle Health');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Particle Health - API/Backend Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'API/Backend Software Engineer', 'https://www.particlehealth.com/careers-jobs?gh_jid=5596514004', '2025-11-11 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 115: Lithos - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Lithos');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Lithos - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://apply.workable.com/lithos/j/5E5DAFE1E5/', '2025-11-11 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 116: Mirage - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Mirage');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Mirage - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://jobs.ashbyhq.com/mirage/203b7270-d65c-4cbe-b790-bdcbcdce385b', '2025-11-11 00:00:00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 117: Monday.com - Solution Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Monday.com');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Monday.com - Solution Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:24:50')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Solution Engineer', 'https://monday.com/careers/23.637', '2025-11-11 11:22:46', '2025-11-11 11:24:50');

  -- Job 118: Rockstar Games - Associate Lead Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Rockstar Games');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Rockstar Games - Associate Lead Software Engineer', 'N', 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:25:13')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, notes, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Associate Lead Software Engineer', 'https://job-boards.greenhouse.io/rockstargames/jobs/7491702003', 'N', '2025-11-11 11:22:46', '2025-11-11 11:25:13');

  -- Job 119: MTA - Application Developer Analyst 1-5
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('MTA');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'MTA - Application Developer Analyst 1-5', 'N', 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:25:22')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, notes, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Application Developer Analyst 1-5', 'https://careers.mta.org/jobs/16890214-application-developer-analyst-1-5-java-tcu', 'N', '2025-11-11 11:22:46', '2025-11-11 11:25:22');

  -- Job 120: Quantiphi - Software Developer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Quantiphi');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Quantiphi - Software Developer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:25:31')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Developer', 'https://hiring.cafe/viewjob/9zzkju7jdrhbaf6h', '2025-11-11 11:22:46', '2025-11-11 11:25:31');

  -- Job 121: Radical AI - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Radical AI');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Radical AI - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 17:35:45')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://jobs.lever.co/RadicalAI/2885cd8e-f208-4bbc-86c2-d7b4d890fba0', '2025-11-05 00:00:00', 'not disclosed', '2025-11-11 11:22:46', '2025-11-11 17:35:45');

  -- Job 122: Workmate - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Workmate');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Workmate - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://jobs.ashbyhq.com/workmatelabs/51a523a0-a2d1-425d-8367-e7e07255ba25', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 123: Pave - Senior Software Engineer, Backend
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Pave');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Pave - Senior Software Engineer, Backend', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Backend', 'https://job-boards.greenhouse.io/paveakatroveinformationtechnologies/jobs/4625023005', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 124: MLabs - Senior Backend Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('MLabs');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'MLabs - Senior Backend Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Backend Engineer', 'https://apply.workable.com/mlabs/j/CDBDA34508/', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 125: Parker - Senior Software Engineer - Backend
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Parker');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Parker - Senior Software Engineer - Backend', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer - Backend', 'https://jobs.ashbyhq.com/parker/bebf55d0-79a0-4a01-8d4d-0d264fd0caed', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 126: Unify - Software Engineer - AI
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Unify');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Unify - Software Engineer - AI', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer - AI', 'https://jobs.ashbyhq.com/unify/704a93f1-74a6-482f-acd7-f0d226d83055', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 127: Findigs - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Findigs');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Findigs - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://jobs.lever.co/findigs/3b39616d-223f-4326-a500-f804c3e18dcf', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 128: Didero - Lead Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Didero');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Didero - Lead Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Lead Software Engineer', 'https://www.didero.ai/careers?ashby_jid=7b892898-c809-440e-901b-27cafb060f00', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 129: OnMed - Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('OnMed');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'OnMed - Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Engineer', 'https://apply.workable.com/onmed/j/EA5E475685/', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 130: Spotify - Staff Backend Enginer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Spotify');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Spotify - Staff Backend Enginer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Staff Backend Enginer', 'https://www.lifeatspotify.com/jobs/staff-backend-engineer-home-foundations-personalization', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 131: Materialize - Staff Full Stack Engineer, Console
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Materialize');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Materialize - Staff Full Stack Engineer, Console', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Staff Full Stack Engineer, Console', 'https://job-boards.greenhouse.io/materialize/jobs/5693572004', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 132: Cassidy - Support Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Cassidy');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Cassidy - Support Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Support Engineer', 'https://jobs.ashbyhq.com/cassidy/83dd49cc-f346-48de-b5e9-902631182a11', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 133: ZocDoc - Senior Software Engineer, Appointment Management
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('ZocDoc');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'ZocDoc - Senior Software Engineer, Appointment Management', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Appointment Management', 'https://hiring.cafe/viewjob/8jxsvy5b2fq717sq', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 134: Elite - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Elite');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Elite - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://job-boards.greenhouse.io/elitetechnology/jobs/4971130008', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 135: Capgemini - Azure Data Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Capgemini');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Capgemini - Azure Data Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Azure Data Engineer', 'https://hiring.cafe/viewjob/8a9h7yehvu5zv7lz', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 136: West Monroe Partners - Cloud Engineer, Cloud & Infrastructure
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('West Monroe Partners');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'West Monroe Partners - Cloud Engineer, Cloud & Infrastructure', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Cloud Engineer, Cloud & Infrastructure', 'https://www.westmonroe.com/careers/job-details-experienced-professionals?gh_jid=5692754004', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 137: The Weather Company - Technical Program Manager, Consumer Data
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('The Weather Company');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'The Weather Company - Technical Program Manager, Consumer Data', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Technical Program Manager, Consumer Data', 'https://job-boards.greenhouse.io/theweathercompany/jobs/4962180007', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 138: Permit Flow - Fullstack Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Permit Flow');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Permit Flow - Fullstack Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Fullstack Software Engineer', 'https://jobs.ashbyhq.com/permitflow/b3505afd-5cf1-4a2e-b875-1775ce346919?departmentId=d33195eb-8978-4439-abc6-5a8a072de808', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 139: Privy - Staff Fullstack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Privy');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Privy - Staff Fullstack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Staff Fullstack Engineer', 'https://stripe.com/jobs/listing/staff-fullstack-engineer-privy/7091959?utm_source=Tech%3ANYC+job+board&utm_medium=getro.com&gh_src=Tech%3ANYC+job+board', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 140: MongoDB - Senior Software Engineer, Deployments
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('MongoDB');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'MongoDB - Senior Software Engineer, Deployments', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Deployments', 'https://www.mongodb.com/careers/jobs/7364854', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 141: Ramp - Backend Engineer Procure to Pay
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Ramp');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Ramp - Backend Engineer Procure to Pay', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Backend Engineer Procure to Pay', 'https://jobs.ashbyhq.com/ramp/2a4968ae-220c-471b-b890-a011de570bbb?utm_source=Tech%3ANYC+job+board&utm_medium=getro.com&gh_src=Tech%3ANYC+job+board', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 142: Reddit - Backend Engineer, Reddit Pro for Publishers
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Reddit');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Reddit - Backend Engineer, Reddit Pro for Publishers', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Backend Engineer, Reddit Pro for Publishers', 'https://job-boards.greenhouse.io/reddit/jobs/7371086', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 143: Lyft - Senior Software Engineer, AI
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Lyft');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Lyft - Senior Software Engineer, AI', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, AI', 'https://app.careerpuck.com/job-board/lyft/job/8221138002?gh_jid=8221138002', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 144: Northwell Health - Lead Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Northwell Health');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Northwell Health - Lead Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Lead Software Engineer', 'https://eppr.fa.us2.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_2/my-profile/preview/168930', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 145: Octane - Staff Software Engineer, AI
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Octane');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Octane - Staff Software Engineer, AI', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Staff Software Engineer, AI', 'https://octane.co/o/who-we-are/careers/jobs-open/?gh_jid=7416486003', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 146: Rokt - Solutions Consultant
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Rokt');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Rokt - Solutions Consultant', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Solutions Consultant', 'https://apply.workable.com/rokt/j/59744FEFFA/', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 147: KYD Labs - Founding Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('KYD Labs');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'KYD Labs - Founding Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Founding Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 148: Strativ Group - Staff Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Strativ Group');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Strativ Group - Staff Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Staff Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 149: Glanceable - Senior Full Stack Egnineer`
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Glanceable');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Glanceable - Senior Full Stack Egnineer`', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Full Stack Egnineer`', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 150: probono net - Data Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('probono net');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'probono net - Data Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Data Engineer', 'https://www.idealist.org/en/nonprofit-job/16ad2d35b130426b94a5f09a56340f41-data-engineer-pro-bono-net-new-york#apply', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  RAISE NOTICE 'Migrated 150 jobs...';

  -- Job 151: Stash - Web Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Stash');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Stash - Web Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Web Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 152: Mias Bakery - Project Based Computer Programmer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Mias Bakery');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Mias Bakery - Project Based Computer Programmer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Project Based Computer Programmer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 153: Veritext - Product Development Specialist
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Veritext');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Veritext - Product Development Specialist', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Product Development Specialist', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 154: Neotecra, Inc - Solutions Architect
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Neotecra, Inc');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Neotecra, Inc - Solutions Architect', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Solutions Architect', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 155: Kintegral Asset Management - Python Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Kintegral Asset Management');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Kintegral Asset Management - Python Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Python Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 156: Simons Foundation - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Simons Foundation');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Simons Foundation - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 157: Braze - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Braze');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Braze - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:22:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://job-boards.greenhouse.io/braze/jobs/7365359', '2025-11-11 11:22:46', '2025-11-11 11:22:46');

  -- Job 158: NYU Langone Health - Software Architect - Education IT & Analytics
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('NYU Langone Health');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'NYU Langone Health - Software Architect - Education IT & Analytics', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:24:31')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Architect - Education IT & Analytics', 'https://jobs.silkroad.com/NYULangone/NYULHCareers/jobs/114034/?source=Indeed.com', '2025-11-11 11:22:46', '2025-11-11 11:24:31');

  -- Job 159: Polimorphic - Full Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Polimorphic');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Polimorphic - Full Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:25:49')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Engineer', 'https://app.trinethire.com/companies/381666-polimorphic/jobs/105062-full-stack-engineer', '2025-11-11 11:22:46', '2025-11-11 11:25:49');

  -- Job 160: Uphold - Senior Backend Engineer (Digital Payments)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Uphold');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Interviewing' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Uphold - Senior Backend Engineer (Digital Payments)', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 17:37:17')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Backend Engineer (Digital Payments)', 'https://uphold.bamboohr.com/careers/748', '2025-11-11 11:22:46', '2025-11-11 17:37:17');

  -- Job 161: Adobe - Senior Full Stack Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Adobe');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Adobe - Senior Full Stack Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 17:40:17')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Full Stack Software Engineer', 'https://adobe.wd5.myworkdayjobs.com/external_experienced/job/San-Jose/Senior-Full-Stack-Software-Engineer---Web-Applications_R162000', '2025-10-30 00:00:00', '2025-11-11 11:22:46', '2025-11-11 17:40:17');

  -- Job 162: Mastercard - Lead Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Mastercard');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Mastercard - Lead Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:25:38')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Lead Software Engineer', 'https://careers.mastercard.com/us/en/job/MASRUSR262392EXTERNALENUS/Lead-Software-Engineer?utm_source=Appcast_IndeedOrganic&utm_medium=phenom-feeds&_ccid=1759263775191oiz3fsss4&utm_source=Indeed&utm_medium=organic&utm_campaign=Indeed&ittk=CTGGR1ZQYT', '2025-11-11 11:22:46', '2025-11-11 11:25:38');

  -- Job 163: Kargo - Senior Javascript Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Kargo');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Kargo - Senior Javascript Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 11:25:44')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Javascript Software Engineer', 'https://www.kargo.com/careers?gh_jid=4621786007', '2025-11-11 11:22:46', '2025-11-11 11:25:44');

  -- Job 164: Evolution IQ - Solutions Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Evolution IQ');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Rejected' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Evolution IQ - Solutions Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 17:36:40')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Solutions Engineer', 'https://job-boards.greenhouse.io/evolutioniq/jobs/5685993004', '2025-11-11 11:22:46', '2025-11-11 17:36:40');

  -- Job 165: Substack - Software Engineer (Payments)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Substack');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Rejected' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Substack - Software Engineer (Payments)', NULL, 'manual', mapped_status_id, '2025-11-11 11:22:46', '2025-11-11 17:36:54')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer (Payments)', 'https://jobs.ashbyhq.com/substack/3033346a-5126-47d4-8f66-d89c996efc61', '2025-11-03 00:00:00', '2025-11-11 11:22:46', '2025-11-11 17:36:54');

  -- Job 166: Okta - Senior Software Engineer, Core Identity (Auth0)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Okta');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Okta - Senior Software Engineer, Core Identity (Auth0)', NULL, 'manual', mapped_status_id, '2025-11-11 19:25:09', '2025-11-11 19:25:09')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Core Identity (Auth0)', 'https://www.okta.com/company/careers/engineering/senior-software-engineer-core-identity-auth0-7229375/', '2025-11-05 00:00:00', '142-214K', '2025-11-11 19:25:09', '2025-11-11 19:25:09');

  -- Job 167: Sage - Software Engineer, Data & Integrations
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Sage');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Sage - Software Engineer, Data & Integrations', NULL, 'manual', mapped_status_id, '2025-11-11 19:30:34', '2025-11-11 19:30:34')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer, Data & Integrations', 'https://job-boards.greenhouse.io/sage49/jobs/5453231004', '2025-11-11 00:00:00', '$160-190k', '2025-11-11 19:30:34', '2025-11-11 19:30:34');

  -- Job 168: Firsthand - Senior Software Engineer, Web Fullstack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Firsthand');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Firsthand - Senior Software Engineer, Web Fullstack', NULL, 'manual', mapped_status_id, '2025-11-11 19:37:33', '2025-11-11 19:37:33')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer, Web Fullstack', 'https://jobs.ashbyhq.com/firsthand/4fc07b55-33c6-40e9-8c91-bbd5bd479f2e', '2025-11-11 00:00:00', '150-190K', '2025-11-11 19:37:33', '2025-11-11 19:37:33');

  -- Job 169: WalkMe - Senior Software Engineer Backend, R&D
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('WalkMe');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'WalkMe - Senior Software Engineer Backend, R&D', NULL, 'manual', mapped_status_id, '2025-11-11 19:44:16', '2025-11-11 19:46:26')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer Backend, R&D', 'https://www.walkme.com/jobs/senior-software-engineer-backend/', '2025-11-11 00:00:00', '2025-11-11 19:44:16', '2025-11-11 19:46:26');

  -- Job 170: Notion - Software Engineer, Product Infrastructure
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Notion');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Notion - Software Engineer, Product Infrastructure', 'DEI heavy', 'manual', mapped_status_id, '2025-11-11 19:53:48', '2025-11-11 19:53:48')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, notes, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer, Product Infrastructure', 'https://jobs.ashbyhq.com/notion/d41b635b-c17b-4efd-89fd-fdb2ddb62e9a', 'DEI heavy', '2025-11-11 00:00:00', '209-240K', '2025-11-11 19:53:48', '2025-11-11 19:53:48');

  -- Job 171: Titan - Principal Software Engineer - Product/AI
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Titan');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Titan - Principal Software Engineer - Product/AI', NULL, 'manual', mapped_status_id, '2025-11-11 20:02:29', '2025-11-11 20:02:29')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Principal Software Engineer - Product/AI', 'https://jobs.ashbyhq.com/titan/0e65d04d-c32d-442d-9b99-d41e2611e359', '2025-11-11 00:00:00', '2025-11-11 20:02:29', '2025-11-11 20:02:29');

  -- Job 172: Fortify - Fullstack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Fortify');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Fortify - Fullstack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 20:09:56', '2025-11-11 20:09:56')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Fullstack Engineer', 'https://jobs.ashbyhq.com/pear/1bf7fd8a-a787-4325-8b65-7dcf5dc159e4', '2025-11-11 00:00:00', '120-155K', '2025-11-11 20:09:56', '2025-11-11 20:09:56');

  -- Job 173: Bridge Health AI - Full Stack Product Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Bridge Health AI');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Bridge Health AI - Full Stack Product Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 20:12:43', '2025-11-11 20:12:43')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Full Stack Product Engineer', 'https://jobs.ashbyhq.com/pear/523c697f-557b-4cdf-a4a9-6b3bac93f379', '2025-11-11 00:00:00', '115-160K', '2025-11-11 20:12:43', '2025-11-11 20:12:43');

  -- Job 174: Vendelux - Senior Fullstack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Vendelux');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Vendelux - Senior Fullstack Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 20:26:47', '2025-11-11 20:26:47')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Fullstack Engineer', 'https://jobs.ashbyhq.com/vendelux/39c512c2-75b9-44a4-ba54-3156d44a593e', '2025-11-11 00:00:00', '100-200K', '2025-11-11 20:26:47', '2025-11-11 20:26:47');

  -- Job 175: EliseAI - Senior Software Engineer (Full Stack)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('EliseAI');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'EliseAI - Senior Software Engineer (Full Stack)', NULL, 'manual', mapped_status_id, '2025-11-11 21:27:03', '2025-11-11 21:27:03')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer (Full Stack)', 'https://jobs.ashbyhq.com/eliseai/bd740e67-b145-46f0-b1f0-0a13610365dc', '2025-11-11 00:00:00', '230-270K', '2025-11-11 21:27:03', '2025-11-11 21:27:03');

  -- Job 176: Normal Computing - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Normal Computing');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Normal Computing - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 22:22:38', '2025-11-11 22:22:38')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://jobs.ashbyhq.com/Normal%20Computing%20AI/6228a7ca-2a62-47ce-99dd-08bed646db94', '2025-11-11 00:00:00', '150-250K', '2025-11-11 22:22:38', '2025-11-11 22:22:38');

  -- Job 177: Nominal - Software Engineer Fullstack
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Nominal');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Nominal - Software Engineer Fullstack', NULL, 'manual', mapped_status_id, '2025-11-11 22:33:44', '2025-11-11 22:33:44')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer Fullstack', 'https://jobs.lever.co/nominal/c6f158d0-ef1d-484f-81cb-b5d29c34270e', '2025-11-11 00:00:00', '130-230K', '2025-11-11 22:33:44', '2025-11-11 22:33:44');

  -- Job 178: TechTree - Backend Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('TechTree');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'TechTree - Backend Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-11 23:12:09', '2025-11-11 23:12:09')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Backend Software Engineer', 'https://jobs.techtree.dev/job/b9226df7-3830-466c-89ef-d14fc77ad551', '2025-11-11 00:00:00', '150-250K', '2025-11-11 23:12:09', '2025-11-11 23:12:09');

  -- Job 179: Valon - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Valon');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Valon - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-12 00:32:26', '2025-11-12 00:32:26')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://jobs.ashbyhq.com/valon/6052f210-29f1-4ef4-93cc-48029969eaf7', '2025-11-11 00:00:00', '180-230K', '2025-11-12 00:32:26', '2025-11-12 00:32:26');

  -- Job 180: Clay - Software Engineer, Integrations
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Clay');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Clay - Software Engineer, Integrations', NULL, 'manual', mapped_status_id, '2025-11-12 00:43:44', '2025-11-12 00:43:44')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer, Integrations', 'https://jobs.ashbyhq.com/claylabs/263fa713-5d67-4acf-8dd3-65ebc00bd6e6?utm_source=remotive.com&ref=remotive.com', '2025-11-11 00:00:00', '2025-11-12 00:43:44', '2025-11-12 00:43:44');

  -- Job 181: Altana - Senior Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Altana');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Altana - Senior Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-12 00:57:11', '2025-11-12 00:57:11')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Software Engineer', 'https://job-boards.greenhouse.io/altanaai/jobs/7491156003?gh_src=13ef1a8f3us', '2025-11-11 00:00:00', '2025-11-12 00:57:11', '2025-11-12 00:57:11');

  -- Job 182: T Rowe Price - Senior AI Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('T Rowe Price');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'T Rowe Price - Senior AI Software Engineer', 'Workday app', 'manual', mapped_status_id, '2025-11-12 02:10:44', '2025-11-12 02:10:44')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, notes, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior AI Software Engineer', 'https://www.builtinnyc.com/job/senior-ai-software-engineer-t-rowe-labs-baltimore-md-or-nyc/7659558', 'Workday app', '2025-11-11 00:00:00', '148-253K', '2025-11-12 02:10:44', '2025-11-12 02:10:44');

  -- Job 183: Accrue - Technical Product Manager / Solutions Engineer - Enterprise Integrations
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Accrue');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Accrue - Technical Product Manager / Solutions Engineer - Enterprise Integrations', NULL, 'manual', mapped_status_id, '2025-11-12 04:04:13', '2025-11-12 04:04:13')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Technical Product Manager / Solutions Engineer - Enterprise Integrations', 'https://job-boards.greenhouse.io/accrue/jobs/4941447008', '2025-11-11 00:00:00', '180-220', '2025-11-12 04:04:13', '2025-11-12 04:04:13');

  -- Job 184: Auctor - Software Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Auctor');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Auctor - Software Engineer', NULL, 'manual', mapped_status_id, '2025-11-12 04:08:39', '2025-11-12 04:08:39')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer', 'https://jobs.ashbyhq.com/auctor/45cd780b-30bd-4887-b1e8-0b4858aa8e63', '2025-11-11 00:00:00', '175-290', '2025-11-12 04:08:39', '2025-11-12 04:08:39');

  -- Job 185: Ridgeline - Senior Technical Consultant
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Ridgeline');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Ridgeline - Senior Technical Consultant', NULL, 'manual', mapped_status_id, '2025-11-12 04:11:26', '2025-11-12 04:11:26')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Technical Consultant', 'https://job-boards.greenhouse.io/ridgeline/jobs/7523160003?gh_jid=7523160003', '2025-11-11 00:00:00', '141-168K', '2025-11-12 04:11:26', '2025-11-12 04:11:26');

  -- Job 186: Norm AI - Software Engineer, Platform (Senior/Staff)
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Norm AI');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Norm AI - Software Engineer, Platform (Senior/Staff)', NULL, 'manual', mapped_status_id, '2025-11-12 04:45:34', '2025-11-12 04:45:34')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Software Engineer, Platform (Senior/Staff)', 'https://jobs.ashbyhq.com/norm-ai/7447df01-aad9-4827-8a25-e43d194e9c53?utm_source=7ow8wN5o3g', '2025-11-11 00:00:00', '2025-11-12 04:45:34', '2025-11-12 04:45:34');

  -- Job 187: Current - Senior Full-Stack Engineer
  SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('Current');
  mapped_status_id := (SELECT id FROM "Status" WHERE name = 'Applied' AND category = 'job');
  INSERT INTO "Lead" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)
  VALUES (default_deal_id, 'Job', 'Current - Senior Full-Stack Engineer', NULL, 'manual', mapped_status_id, '2025-11-12 04:49:18', '2025-11-12 04:49:46')
  RETURNING id INTO new_lead_id;
  INSERT INTO "Job" (id, organization_id, job_title, job_url, resume_date, salary_range, created_at, updated_at)
  VALUES (new_lead_id, org_id, 'Senior Full-Stack Engineer', 'https://current.com/careers/open-positions/?id=7305170&gh_jid=7305170', '2025-11-11 00:00:00', '150-240K', '2025-11-12 04:49:18', '2025-11-12 04:49:46');

  RAISE NOTICE 'Migration complete! Migrated 187 jobs';
END $$;

-- Verification queries
SELECT COUNT(*) as total_jobs FROM "Job";
SELECT COUNT(*) as total_leads_jobs FROM "Lead" WHERE type = 'Job';
SELECT
  s.name as status,
  COUNT(*) as count
FROM "Lead" l
JOIN "Status" s ON l.current_status_id = s.id
WHERE l.type = 'Job'
GROUP BY s.name
ORDER BY count DESC;

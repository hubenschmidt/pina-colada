-- Seeder: Prospect Research Data
-- Seeds data for migrations 021-026: Technologies, Funding Rounds, Company Signals,
-- Organization firmographics, and Individual contact intelligence
--
-- IMPORTANT: This runs AFTER 004_contacts_relationships_seed.sql

-- ==============================
-- STEP 1: Seed Technologies
-- ==============================
INSERT INTO "Technology" (name, category, vendor) VALUES
    -- CRM
    ('Salesforce', 'CRM', 'Salesforce'),
    ('HubSpot CRM', 'CRM', 'HubSpot'),
    ('Pipedrive', 'CRM', 'Pipedrive'),
    -- Marketing Automation
    ('HubSpot Marketing', 'Marketing Automation', 'HubSpot'),
    ('Marketo', 'Marketing Automation', 'Adobe'),
    ('Mailchimp', 'Marketing Automation', 'Intuit'),
    -- Cloud Infrastructure
    ('AWS', 'Cloud Infrastructure', 'Amazon'),
    ('Google Cloud', 'Cloud Infrastructure', 'Google'),
    ('Azure', 'Cloud Infrastructure', 'Microsoft'),
    ('Cloudflare', 'Cloud Infrastructure', 'Cloudflare'),
    -- Database
    ('PostgreSQL', 'Database', 'PostgreSQL'),
    ('MongoDB', 'Database', 'MongoDB'),
    ('Redis', 'Database', 'Redis'),
    ('Snowflake', 'Database', 'Snowflake'),
    -- Analytics
    ('Google Analytics', 'Analytics', 'Google'),
    ('Mixpanel', 'Analytics', 'Mixpanel'),
    ('Amplitude', 'Analytics', 'Amplitude'),
    ('Tableau', 'Analytics', 'Salesforce'),
    -- DevOps
    ('GitHub', 'DevOps', 'Microsoft'),
    ('GitLab', 'DevOps', 'GitLab'),
    ('Docker', 'DevOps', 'Docker'),
    ('Kubernetes', 'DevOps', 'CNCF'),
    ('Terraform', 'DevOps', 'HashiCorp'),
    -- Communication
    ('Slack', 'Communication', 'Salesforce'),
    ('Zoom', 'Communication', 'Zoom'),
    ('Microsoft Teams', 'Communication', 'Microsoft'),
    -- Security
    ('Okta', 'Security', 'Okta'),
    ('Auth0', 'Security', 'Okta'),
    ('CrowdStrike', 'Security', 'CrowdStrike'),
    -- Payments
    ('Stripe', 'Payments', 'Stripe'),
    ('Braintree', 'Payments', 'PayPal')
ON CONFLICT (name, category) DO NOTHING;

-- ==============================
-- STEP 2: Update Organization Firmographics
-- ==============================
DO $$
DECLARE
    v_org_id BIGINT;
    v_revenue_range_id BIGINT;
    v_employee_count_id BIGINT;
    v_funding_stage_id BIGINT;
BEGIN
    -- TechVentures Inc - VC firm (51-500 employees, established)
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$100M - $500M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '51-500' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Series C' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2015,
            headquarters_city = 'San Francisco',
            headquarters_state = 'CA',
            headquarters_country = 'USA',
            company_type = 'private',
            linkedin_url = 'https://linkedin.com/company/techventures',
            crunchbase_url = 'https://crunchbase.com/organization/techventures'
        WHERE id = v_org_id;
    END IF;

    -- CloudScale Systems - Enterprise SaaS (501-1500 employees, Series B)
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$50M - $100M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '501-1500' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Series B' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2018,
            headquarters_city = 'Seattle',
            headquarters_state = 'WA',
            headquarters_country = 'USA',
            company_type = 'private',
            linkedin_url = 'https://linkedin.com/company/cloudscale-systems',
            crunchbase_url = 'https://crunchbase.com/organization/cloudscale-systems'
        WHERE id = v_org_id;
    END IF;

    -- DataFlow Analytics - Growth stage startup (51-500 employees, Series A)
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$10M - $50M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '51-500' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Series A' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2020,
            headquarters_city = 'Austin',
            headquarters_state = 'TX',
            headquarters_country = 'USA',
            company_type = 'private',
            linkedin_url = 'https://linkedin.com/company/dataflow-analytics',
            crunchbase_url = 'https://crunchbase.com/organization/dataflow-analytics'
        WHERE id = v_org_id;
    END IF;

    -- SecureNet Solutions - Established security company (51-500 employees, Series B)
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$50M - $100M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '51-500' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Series B' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2016,
            headquarters_city = 'San Jose',
            headquarters_state = 'CA',
            headquarters_country = 'USA',
            company_type = 'private',
            linkedin_url = 'https://linkedin.com/company/securenet-solutions',
            crunchbase_url = 'https://crunchbase.com/organization/securenet-solutions'
        WHERE id = v_org_id;
    END IF;

    -- InnovateLab - Early stage consulting (11-50 employees, Seed)
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$1M - $10M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '11-50' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Seed' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2021,
            headquarters_city = 'Oakland',
            headquarters_state = 'CA',
            headquarters_country = 'USA',
            company_type = 'private',
            linkedin_url = 'https://linkedin.com/company/innovatelab'
        WHERE id = v_org_id;
    END IF;

    -- Also seed some orgs from 001_initial_seed.sql
    -- Acme Corp - mid-size, Series A
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('Acme Corp') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$10M - $50M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '51-500' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Series A' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2019,
            headquarters_city = 'Denver',
            headquarters_state = 'CO',
            headquarters_country = 'USA',
            company_type = 'private'
        WHERE id = v_org_id;
    END IF;

    -- TechStartup Inc - early stage
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('TechStartup Inc') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$1M - $10M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '11-50' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Seed' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2022,
            headquarters_city = 'Boston',
            headquarters_state = 'MA',
            headquarters_country = 'USA',
            company_type = 'private'
        WHERE id = v_org_id;
    END IF;

    -- DataSystems Ltd - established
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataSystems Ltd') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$50M - $100M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '501-1500' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Series C' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2012,
            headquarters_city = 'Chicago',
            headquarters_state = 'IL',
            headquarters_country = 'USA',
            company_type = 'private'
        WHERE id = v_org_id;
    END IF;

    -- CloudWorks - growth
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudWorks') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$10M - $50M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '51-500' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Series B' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2017,
            headquarters_city = 'Portland',
            headquarters_state = 'OR',
            headquarters_country = 'USA',
            company_type = 'private'
        WHERE id = v_org_id;
    END IF;

    -- AI Innovations - early stage AI startup
    SELECT id INTO v_org_id FROM "Organization" WHERE LOWER(name) = LOWER('AI Innovations') LIMIT 1;
    SELECT id INTO v_revenue_range_id FROM "RevenueRange" WHERE label = '$1M - $10M' LIMIT 1;
    SELECT id INTO v_employee_count_id FROM "EmployeeCountRange" WHERE label = '1-10' LIMIT 1;
    SELECT id INTO v_funding_stage_id FROM "FundingStage" WHERE label = 'Pre-seed' LIMIT 1;
    IF v_org_id IS NOT NULL THEN
        UPDATE "Organization" SET
            revenue_range_id = v_revenue_range_id,
            employee_count_range_id = v_employee_count_id,
            funding_stage_id = v_funding_stage_id,
            founding_year = 2023,
            headquarters_city = 'San Francisco',
            headquarters_state = 'CA',
            headquarters_country = 'USA',
            company_type = 'private'
        WHERE id = v_org_id;
    END IF;

    RAISE NOTICE 'Organization firmographics updated successfully';
END $$;

-- ==============================
-- STEP 3: Add Tech Stacks to Organizations
-- ==============================
DO $$
DECLARE
    org_id BIGINT;
    tech_id BIGINT;
BEGIN
    -- CloudScale Systems tech stack
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    IF org_id IS NOT NULL THEN
        SELECT id INTO tech_id FROM "Technology" WHERE name = 'AWS' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.95) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Kubernetes' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.90) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'PostgreSQL' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.85) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Terraform' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.80) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Slack' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.95) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'GitHub' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.90) ON CONFLICT DO NOTHING;
    END IF;

    -- DataFlow Analytics tech stack
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    IF org_id IS NOT NULL THEN
        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Google Cloud' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.95) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Snowflake' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.90) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Amplitude' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.85) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'HubSpot CRM' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.90) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Stripe' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.95) ON CONFLICT DO NOTHING;
    END IF;

    -- SecureNet Solutions tech stack
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    IF org_id IS NOT NULL THEN
        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Azure' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.95) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Okta' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.95) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'CrowdStrike' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'agent', 0.85) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Salesforce' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.90) ON CONFLICT DO NOTHING;
    END IF;

    -- TechVentures Inc tech stack (lighter, they're a VC)
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    IF org_id IS NOT NULL THEN
        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Salesforce' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.95) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Slack' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.95) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Zoom' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.90) ON CONFLICT DO NOTHING;
    END IF;

    -- InnovateLab tech stack
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;
    IF org_id IS NOT NULL THEN
        SELECT id INTO tech_id FROM "Technology" WHERE name = 'HubSpot CRM' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.90) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Slack' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.95) ON CONFLICT DO NOTHING;

        SELECT id INTO tech_id FROM "Technology" WHERE name = 'Google Analytics' LIMIT 1;
        INSERT INTO "OrganizationTechnology" (organization_id, technology_id, source, confidence) VALUES (org_id, tech_id, 'builtwith', 0.95) ON CONFLICT DO NOTHING;
    END IF;

    RAISE NOTICE 'Organization tech stacks added successfully';
END $$;

-- ==============================
-- STEP 4: Add Funding Rounds
-- ==============================
DO $$
DECLARE
    org_id BIGINT;
BEGIN
    -- CloudScale Systems funding history
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    IF org_id IS NOT NULL THEN
        INSERT INTO "FundingRound" (organization_id, round_type, amount, announced_date, lead_investor, source_url)
        VALUES
            (org_id, 'Seed', 300000000, '2018-06-15', 'Y Combinator', 'https://techcrunch.com/2018/06/15/cloudscale-seed'),
            (org_id, 'Series A', 1500000000, '2020-03-20', 'Andreessen Horowitz', 'https://techcrunch.com/2020/03/20/cloudscale-series-a'),
            (org_id, 'Series B', 5000000000, '2022-09-10', 'Sequoia Capital', 'https://techcrunch.com/2022/09/10/cloudscale-series-b');
    END IF;

    -- DataFlow Analytics funding history
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    IF org_id IS NOT NULL THEN
        INSERT INTO "FundingRound" (organization_id, round_type, amount, announced_date, lead_investor, source_url)
        VALUES
            (org_id, 'Pre-Seed', 50000000, '2020-02-01', 'Angel Investors', NULL),
            (org_id, 'Seed', 250000000, '2021-05-15', 'First Round Capital', 'https://techcrunch.com/2021/05/15/dataflow-seed'),
            (org_id, 'Series A', 1200000000, '2023-08-20', 'Accel', 'https://techcrunch.com/2023/08/20/dataflow-series-a');
    END IF;

    -- SecureNet Solutions funding history
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    IF org_id IS NOT NULL THEN
        INSERT INTO "FundingRound" (organization_id, round_type, amount, announced_date, lead_investor, source_url)
        VALUES
            (org_id, 'Seed', 400000000, '2016-09-01', 'Cybersecurity Angels', NULL),
            (org_id, 'Series A', 2000000000, '2018-11-15', 'Insight Partners', 'https://techcrunch.com/2018/11/15/securenet-series-a'),
            (org_id, 'Series B', 4500000000, '2021-04-10', 'Greylock Partners', 'https://techcrunch.com/2021/04/10/securenet-series-b');
    END IF;

    -- InnovateLab (bootstrapped + small seed)
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;
    IF org_id IS NOT NULL THEN
        INSERT INTO "FundingRound" (organization_id, round_type, amount, announced_date, lead_investor, source_url)
        VALUES
            (org_id, 'Seed', 100000000, '2022-01-15', 'TechVentures Inc', NULL);
    END IF;

    RAISE NOTICE 'Funding rounds added successfully';
END $$;

-- ==============================
-- STEP 5: Add Company Signals
-- ==============================
DO $$
DECLARE
    org_id BIGINT;
BEGIN
    -- CloudScale Systems signals
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    IF org_id IS NOT NULL THEN
        INSERT INTO "CompanySignal" (organization_id, signal_type, headline, description, signal_date, source, source_url, sentiment, relevance_score)
        VALUES
            (org_id, 'hiring', 'CloudScale expanding engineering team by 50%', 'CloudScale Systems announced plans to hire 100 new engineers across their Seattle and Austin offices.', '2024-11-15', 'linkedin', 'https://linkedin.com/posts/cloudscale-hiring', 'positive', 0.85),
            (org_id, 'expansion', 'CloudScale opens new Austin office', 'CloudScale Systems opened a new engineering hub in Austin, TX to tap into local tech talent.', '2024-10-01', 'news', 'https://techcrunch.com/2024/10/01/cloudscale-austin', 'positive', 0.75),
            (org_id, 'product_launch', 'CloudScale launches AI-powered analytics', 'New AI features added to CloudScale platform for predictive infrastructure scaling.', '2024-09-20', 'news', 'https://techcrunch.com/2024/09/20/cloudscale-ai', 'positive', 0.90);
    END IF;

    -- DataFlow Analytics signals
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    IF org_id IS NOT NULL THEN
        INSERT INTO "CompanySignal" (organization_id, signal_type, headline, description, signal_date, source, source_url, sentiment, relevance_score)
        VALUES
            (org_id, 'hiring', 'DataFlow hiring senior data scientists', 'DataFlow Analytics seeking experienced data scientists for their ML team.', '2024-11-10', 'linkedin', 'https://linkedin.com/jobs/dataflow', 'positive', 0.80),
            (org_id, 'partnership', 'DataFlow partners with Snowflake', 'Strategic partnership announced to improve data integration capabilities.', '2024-08-15', 'news', 'https://prnewswire.com/dataflow-snowflake', 'positive', 0.85),
            (org_id, 'product_launch', 'DataFlow launches real-time dashboards', 'New real-time dashboard feature enables instant data visualization.', '2024-07-01', 'news', NULL, 'positive', 0.70);
    END IF;

    -- SecureNet Solutions signals
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    IF org_id IS NOT NULL THEN
        INSERT INTO "CompanySignal" (organization_id, signal_type, headline, description, signal_date, source, source_url, sentiment, relevance_score)
        VALUES
            (org_id, 'leadership_change', 'SecureNet appoints new CRO', 'Former Palo Alto Networks executive joins as Chief Revenue Officer.', '2024-10-15', 'linkedin', 'https://linkedin.com/posts/securenet-cro', 'positive', 0.75),
            (org_id, 'hiring', 'SecureNet building out sales team', 'Aggressive enterprise sales hiring in Q4 2024.', '2024-11-01', 'linkedin', NULL, 'positive', 0.70),
            (org_id, 'news', 'SecureNet featured in Gartner Magic Quadrant', 'Named a challenger in the 2024 endpoint security Magic Quadrant.', '2024-09-01', 'news', 'https://gartner.com/magic-quadrant-2024', 'positive', 0.95);
    END IF;

    -- TechVentures Inc signals
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    IF org_id IS NOT NULL THEN
        INSERT INTO "CompanySignal" (organization_id, signal_type, headline, description, signal_date, source, source_url, sentiment, relevance_score)
        VALUES
            (org_id, 'news', 'TechVentures closes $500M Fund III', 'New fund focused on early-stage AI and infrastructure startups.', '2024-06-01', 'news', 'https://techcrunch.com/2024/06/01/techventures-fund-iii', 'positive', 0.80);
    END IF;

    RAISE NOTICE 'Company signals added successfully';
END $$;

-- ==============================
-- STEP 6: Update Individual Contact Intelligence
-- ==============================
DO $$
DECLARE
    ind_id BIGINT;
BEGIN
    -- Alex Thompson - Independent consultant
    SELECT id INTO ind_id FROM "Individual" WHERE LOWER(email) = LOWER('alex.thompson@gmail.com') LIMIT 1;
    IF ind_id IS NOT NULL THEN
        UPDATE "Individual" SET
            twitter_url = 'https://twitter.com/alexthompson',
            github_url = 'https://github.com/alexthompson',
            bio = 'Former VP of Product at Stripe. Now advising early-stage startups on product strategy and go-to-market.',
            seniority_level = 'VP',
            department = 'Product',
            is_decision_maker = TRUE
        WHERE id = ind_id;
    END IF;

    -- Maria Garcia - Freelance designer
    SELECT id INTO ind_id FROM "Individual" WHERE LOWER(email) = LOWER('maria.garcia@designstudio.com') LIMIT 1;
    IF ind_id IS NOT NULL THEN
        UPDATE "Individual" SET
            twitter_url = 'https://twitter.com/mariagarcia_ux',
            github_url = NULL,
            bio = 'Award-winning UX designer specializing in B2B SaaS products. Previously at Figma and Notion.',
            seniority_level = 'IC',
            department = 'Design',
            is_decision_maker = FALSE
        WHERE id = ind_id;
    END IF;

    -- James Wilson - Mentor/Advisor
    SELECT id INTO ind_id FROM "Individual" WHERE LOWER(email) = LOWER('james.wilson@advisors.co') LIMIT 1;
    IF ind_id IS NOT NULL THEN
        UPDATE "Individual" SET
            twitter_url = 'https://twitter.com/jameswilson_vc',
            github_url = NULL,
            bio = 'Former CTO at 2 successful exits. Board member and advisor to 15+ startups. Focus on enterprise software and developer tools.',
            seniority_level = 'C-Level',
            department = 'Executive',
            is_decision_maker = TRUE
        WHERE id = ind_id;
    END IF;

    RAISE NOTICE 'Individual contact intelligence updated successfully';
END $$;

-- ==============================
-- STEP 7: Add Industries to Accounts (for orgs from 001_initial_seed.sql)
-- ==============================
DO $$
DECLARE
    v_account_id BIGINT;
    v_industry_id BIGINT;
BEGIN
    -- Acme Corp - Software
    SELECT a.id INTO v_account_id FROM "Account" a
    JOIN "Organization" o ON o.account_id = a.id
    WHERE LOWER(o.name) = LOWER('Acme Corp') LIMIT 1;
    SELECT id INTO v_industry_id FROM "Industry" WHERE name = 'Software' LIMIT 1;
    IF v_account_id IS NOT NULL AND v_industry_id IS NOT NULL THEN
        INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (v_account_id, v_industry_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- TechStartup Inc - AI/ML
    SELECT a.id INTO v_account_id FROM "Account" a
    JOIN "Organization" o ON o.account_id = a.id
    WHERE LOWER(o.name) = LOWER('TechStartup Inc') LIMIT 1;
    SELECT id INTO v_industry_id FROM "Industry" WHERE name = 'AI/ML' LIMIT 1;
    IF v_account_id IS NOT NULL AND v_industry_id IS NOT NULL THEN
        INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (v_account_id, v_industry_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- DataSystems Ltd - Data Analytics
    SELECT a.id INTO v_account_id FROM "Account" a
    JOIN "Organization" o ON o.account_id = a.id
    WHERE LOWER(o.name) = LOWER('DataSystems Ltd') LIMIT 1;
    SELECT id INTO v_industry_id FROM "Industry" WHERE name = 'Data Analytics' LIMIT 1;
    IF v_account_id IS NOT NULL AND v_industry_id IS NOT NULL THEN
        INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (v_account_id, v_industry_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- CloudWorks - Cloud Infrastructure
    SELECT a.id INTO v_account_id FROM "Account" a
    JOIN "Organization" o ON o.account_id = a.id
    WHERE LOWER(o.name) = LOWER('CloudWorks') LIMIT 1;
    SELECT id INTO v_industry_id FROM "Industry" WHERE name = 'Cloud Infrastructure' LIMIT 1;
    IF v_account_id IS NOT NULL AND v_industry_id IS NOT NULL THEN
        INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (v_account_id, v_industry_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- AI Innovations - AI/ML
    SELECT a.id INTO v_account_id FROM "Account" a
    JOIN "Organization" o ON o.account_id = a.id
    WHERE LOWER(o.name) = LOWER('AI Innovations') LIMIT 1;
    SELECT id INTO v_industry_id FROM "Industry" WHERE name = 'AI/ML' LIMIT 1;
    IF v_account_id IS NOT NULL AND v_industry_id IS NOT NULL THEN
        INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (v_account_id, v_industry_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- Also add industry for Individual accounts
    -- Alex Thompson - Consulting
    SELECT a.id INTO v_account_id FROM "Account" a
    JOIN "Individual" i ON i.account_id = a.id
    WHERE LOWER(i.email) = LOWER('alex.thompson@gmail.com') LIMIT 1;
    SELECT id INTO v_industry_id FROM "Industry" WHERE name = 'Consulting' LIMIT 1;
    IF v_account_id IS NOT NULL AND v_industry_id IS NOT NULL THEN
        INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (v_account_id, v_industry_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- Maria Garcia - Software (design for software)
    SELECT a.id INTO v_account_id FROM "Account" a
    JOIN "Individual" i ON i.account_id = a.id
    WHERE LOWER(i.email) = LOWER('maria.garcia@designstudio.com') LIMIT 1;
    SELECT id INTO v_industry_id FROM "Industry" WHERE name = 'Software' LIMIT 1;
    IF v_account_id IS NOT NULL AND v_industry_id IS NOT NULL THEN
        INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (v_account_id, v_industry_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- James Wilson - Venture Capital
    SELECT a.id INTO v_account_id FROM "Account" a
    JOIN "Individual" i ON i.account_id = a.id
    WHERE LOWER(i.email) = LOWER('james.wilson@advisors.co') LIMIT 1;
    SELECT id INTO v_industry_id FROM "Industry" WHERE name = 'Venture Capital' LIMIT 1;
    IF v_account_id IS NOT NULL AND v_industry_id IS NOT NULL THEN
        INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (v_account_id, v_industry_id)
        ON CONFLICT DO NOTHING;
    END IF;

    RAISE NOTICE 'Account industries added successfully';
END $$;

-- ==============================
-- Return counts for seeder verification
-- ==============================
SELECT
  (SELECT COUNT(*) FROM "Technology") as technologies,
  (SELECT COUNT(*) FROM "OrganizationTechnology") as org_tech_links,
  (SELECT COUNT(*) FROM "FundingRound") as funding_rounds,
  (SELECT COUNT(*) FROM "CompanySignal") as company_signals,
  (SELECT COUNT(*) FROM "Organization" WHERE revenue_range_id IS NOT NULL) as orgs_with_revenue,
  (SELECT COUNT(*) FROM "Organization" WHERE employee_count_range_id IS NOT NULL) as orgs_with_employees,
  (SELECT COUNT(*) FROM "Organization" WHERE funding_stage_id IS NOT NULL) as orgs_with_funding_stage,
  (SELECT COUNT(*) FROM "Individual" WHERE seniority_level IS NOT NULL) as individuals_with_seniority,
  (SELECT COUNT(*) FROM "Account_Industry") as account_industries;

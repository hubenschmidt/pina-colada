-- Seeder: DealTracker sample data (NEW SCHEMA)
-- This seeder demonstrates the full DealTracker schema with sample data
-- Includes: Organizations, Individuals, Contacts, Deals, Leads, Tasks, Activities
--
-- IMPORTANT: This runs AFTER migrations/002_dealtracker.sql
-- This seeds data using the new DealTracker schema structure

-- ==============================
-- STEP 0: Create Default Tenant and User
-- ==============================
DO $$
DECLARE
    v_tenant_id BIGINT;
    v_account_id BIGINT;
    v_individual_id BIGINT;
    v_user_id BIGINT;
    v_role_id BIGINT;
BEGIN
    -- Create Tenant
    INSERT INTO "Tenant" (name, slug, plan, industry, website, employee_count, created_at, updated_at)
    VALUES ('PinaColada', 'pinacolada', 'free', 'Software', 'https://pinacolada.co', 1, NOW(), NOW())
    ON CONFLICT (slug) DO UPDATE SET name = 'PinaColada', industry = 'Software', website = 'https://pinacolada.co', employee_count = 1
    RETURNING id INTO v_tenant_id;

    -- Ensure we have tenant_id (in case ON CONFLICT didn't return it)
    IF v_tenant_id IS NULL THEN
        SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;
    END IF;

    RAISE NOTICE 'Tenant ID: %', v_tenant_id;

    -- Create Account for the user's Individual (with tenant_id)
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at)
    VALUES (v_tenant_id, 'William Hubenschmidt', NOW(), NOW())
    RETURNING id INTO v_account_id;

    -- Create or update Individual for the user
    SELECT id INTO v_individual_id FROM "Individual" WHERE LOWER(email) = LOWER('whubenschmidt@gmail.com') LIMIT 1;

    IF v_individual_id IS NULL THEN
        -- Individual doesn't exist, create new
        INSERT INTO "Individual" (account_id, first_name, last_name, email, created_at, updated_at)
        VALUES (v_account_id, 'William', 'Hubenschmidt', 'whubenschmidt@gmail.com', NOW(), NOW())
        RETURNING id INTO v_individual_id;
    ELSE
        -- Individual exists, update with account
        UPDATE "Individual"
        SET account_id = v_account_id,
            first_name = 'William',
            last_name = 'Hubenschmidt',
            updated_at = NOW()
        WHERE id = v_individual_id;
    END IF;

    -- Create or update User
    -- Use email index to find any existing user (regardless of tenant)
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    IF v_user_id IS NULL THEN
        -- User doesn't exist, create new
        INSERT INTO "User" (tenant_id, individual_id, email, first_name, last_name, status, created_at, updated_at)
        VALUES (v_tenant_id, v_individual_id, 'whubenschmidt@gmail.com', 'William', 'Hubenschmidt', 'active', NOW(), NOW())
        RETURNING id INTO v_user_id;
    ELSE
        -- User exists, update with tenant and individual
        UPDATE "User"
        SET tenant_id = v_tenant_id,
            individual_id = v_individual_id,
            first_name = 'William',
            last_name = 'Hubenschmidt',
            updated_at = NOW()
        WHERE id = v_user_id;
    END IF;

    -- Ensure tenant_id is set (safety check)
    IF v_tenant_id IS NOT NULL THEN
        UPDATE "User" SET tenant_id = v_tenant_id WHERE id = v_user_id AND tenant_id IS NULL;
    END IF;

    -- Create Owner Role
    INSERT INTO "Role" (tenant_id, name, description)
    VALUES (v_tenant_id, 'owner', 'Full access to all resources')
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_role_id;

    -- If role already exists, get its id
    IF v_role_id IS NULL THEN
        SELECT id INTO v_role_id FROM "Role" WHERE tenant_id = v_tenant_id AND name = 'owner';
    END IF;

    -- Create UserRole
    INSERT INTO "UserRole" (user_id, role_id, created_at)
    VALUES (v_user_id, v_role_id, NOW())
    ON CONFLICT DO NOTHING;

    -- Create default Organization for the tenant
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at)
    VALUES (v_tenant_id, 'PinaColada', NOW(), NOW())
    RETURNING id INTO v_account_id;

    INSERT INTO "Organization" (account_id, name, created_at, updated_at)
    VALUES (v_account_id, 'PinaColada', NOW(), NOW())
    ON CONFLICT ((LOWER(name))) DO NOTHING;

    RAISE NOTICE 'Default tenant and user created successfully. Tenant ID: %', v_tenant_id;
END $$;

-- ==============================
-- STEP 1: Create Sample Organizations (with Accounts and Industries)
-- ==============================
DO $$
DECLARE
    account_id BIGINT;
    org_id BIGINT;
    v_tenant_id BIGINT;
    industry_id BIGINT;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';

    -- TechVentures Inc
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'TechVentures Inc', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at)
    VALUES (account_id, 'TechVentures Inc', 'https://techventures.example.com', 50, 'Early-stage technology investor', NOW(), NOW())
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Venture Capital';
    INSERT INTO "Organization_Industry" (organization_id, industry_id) VALUES (org_id, industry_id) ON CONFLICT DO NOTHING;

    -- CloudScale Systems
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'CloudScale Systems', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at)
    VALUES (account_id, 'CloudScale Systems', 'https://cloudscale.example.com', 200, 'Enterprise cloud solutions provider', NOW(), NOW())
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Cloud Infrastructure';
    INSERT INTO "Organization_Industry" (organization_id, industry_id) VALUES (org_id, industry_id) ON CONFLICT DO NOTHING;

    -- DataFlow Analytics
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'DataFlow Analytics', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at)
    VALUES (account_id, 'DataFlow Analytics', 'https://dataflow.example.com', 75, 'Real-time data analytics platform', NOW(), NOW())
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Data Analytics';
    INSERT INTO "Organization_Industry" (organization_id, industry_id) VALUES (org_id, industry_id) ON CONFLICT DO NOTHING;

    -- SecureNet Solutions
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'SecureNet Solutions', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at)
    VALUES (account_id, 'SecureNet Solutions', 'https://securenet.example.com', 150, 'Enterprise security software', NOW(), NOW())
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Cybersecurity';
    INSERT INTO "Organization_Industry" (organization_id, industry_id) VALUES (org_id, industry_id) ON CONFLICT DO NOTHING;

    -- InnovateLab
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'InnovateLab', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at)
    VALUES (account_id, 'InnovateLab', 'https://innovatelab.example.com', 30, 'Innovation consulting for startups', NOW(), NOW())
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Consulting';
    INSERT INTO "Organization_Industry" (organization_id, industry_id) VALUES (org_id, industry_id) ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Organizations with Accounts and Industries created successfully';
END $$;

-- ==============================
-- STEP 2: Create Sample Individuals (with Accounts)
-- ==============================
DO $$
DECLARE
    account_id BIGINT;
    v_tenant_id BIGINT;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';

    -- Sarah Chen
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'Sarah Chen', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, notes, created_at, updated_at)
    VALUES (account_id, 'Sarah', 'Chen', 'sarah.chen@techventures.example.com', '+1-415-555-0101', 'https://linkedin.com/in/sarachen', 'Partner', 'Focus on SaaS and AI investments', NOW(), NOW())
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    -- Michael Rodriguez
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'Michael Rodriguez', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, notes, created_at, updated_at)
    VALUES (account_id, 'Michael', 'Rodriguez', 'michael.r@cloudscale.example.com', '+1-415-555-0102', 'https://linkedin.com/in/michaelrodriguez', 'VP of Engineering', 'Leads cloud infrastructure team', NOW(), NOW())
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    -- Emily Johnson
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'Emily Johnson', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, notes, created_at, updated_at)
    VALUES (account_id, 'Emily', 'Johnson', 'emily.j@dataflow.example.com', '+1-650-555-0103', 'https://linkedin.com/in/emilyjohnson', 'Head of Product', 'Former Google PM', NOW(), NOW())
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    -- David Kim
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'David Kim', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, notes, created_at, updated_at)
    VALUES (account_id, 'David', 'Kim', 'david.kim@securenet.example.com', '+1-408-555-0104', 'https://linkedin.com/in/davidkim', 'CTO', 'Security expert with 15+ years experience', NOW(), NOW())
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    -- Jessica Williams
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'Jessica Williams', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, notes, created_at, updated_at)
    VALUES (account_id, 'Jessica', 'Williams', 'jessica.w@innovatelab.example.com', '+1-510-555-0105', 'https://linkedin.com/in/jessicawilliams', 'Principal Consultant', 'Specializes in go-to-market strategy', NOW(), NOW())
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    -- Robert Taylor
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'Robert Taylor', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, notes, created_at, updated_at)
    VALUES (account_id, 'Robert', 'Taylor', 'robert.t@techventures.example.com', '+1-415-555-0106', 'https://linkedin.com/in/roberttaylor', 'Associate', 'Focuses on early-stage deals', NOW(), NOW())
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    -- Amanda Brown
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at) VALUES (v_tenant_id, 'Amanda Brown', NOW(), NOW()) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, notes, created_at, updated_at)
    VALUES (account_id, 'Amanda', 'Brown', 'amanda.b@cloudscale.example.com', '+1-415-555-0107', 'https://linkedin.com/in/amandabrown', 'Senior Software Engineer', 'Backend infrastructure specialist', NOW(), NOW())
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    RAISE NOTICE 'Individuals with Accounts created successfully';
END $$;

-- ==============================
-- STEP 3: Create Contact Relationships
-- ==============================
DO $$
DECLARE
    org_techventures_id BIGINT;
    org_cloudscale_id BIGINT;
    org_dataflow_id BIGINT;
    org_securenet_id BIGINT;
    org_innovatelab_id BIGINT;

    ind_sarah_id BIGINT;
    ind_michael_id BIGINT;
    ind_emily_id BIGINT;
    ind_david_id BIGINT;
    ind_jessica_id BIGINT;
    ind_robert_id BIGINT;
    ind_amanda_id BIGINT;
BEGIN
    -- Get Organization IDs
    SELECT id INTO org_techventures_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT id INTO org_cloudscale_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT id INTO org_dataflow_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT id INTO org_securenet_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    SELECT id INTO org_innovatelab_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;

    -- Get Individual IDs
    SELECT id INTO ind_sarah_id FROM "Individual" WHERE LOWER(email) = LOWER('sarah.chen@techventures.example.com') LIMIT 1;
    SELECT id INTO ind_michael_id FROM "Individual" WHERE LOWER(email) = LOWER('michael.r@cloudscale.example.com') LIMIT 1;
    SELECT id INTO ind_emily_id FROM "Individual" WHERE LOWER(email) = LOWER('emily.j@dataflow.example.com') LIMIT 1;
    SELECT id INTO ind_david_id FROM "Individual" WHERE LOWER(email) = LOWER('david.kim@securenet.example.com') LIMIT 1;
    SELECT id INTO ind_jessica_id FROM "Individual" WHERE LOWER(email) = LOWER('jessica.w@innovatelab.example.com') LIMIT 1;
    SELECT id INTO ind_robert_id FROM "Individual" WHERE LOWER(email) = LOWER('robert.t@techventures.example.com') LIMIT 1;
    SELECT id INTO ind_amanda_id FROM "Individual" WHERE LOWER(email) = LOWER('amanda.b@cloudscale.example.com') LIMIT 1;

    -- Create Contact relationships
    IF org_techventures_id IS NOT NULL AND ind_sarah_id IS NOT NULL THEN
        INSERT INTO "Contact" (organization_id, individual_id, role, is_primary, notes, created_at, updated_at)
        VALUES (org_techventures_id, ind_sarah_id, 'Decision Maker', TRUE, 'Main contact for investment discussions', NOW(), NOW());
    END IF;

    IF org_techventures_id IS NOT NULL AND ind_robert_id IS NOT NULL THEN
        INSERT INTO "Contact" (organization_id, individual_id, role, is_primary, notes, created_at, updated_at)
        VALUES (org_techventures_id, ind_robert_id, 'Influencer', FALSE, 'Can provide warm introductions', NOW(), NOW());
    END IF;

    IF org_cloudscale_id IS NOT NULL AND ind_michael_id IS NOT NULL THEN
        INSERT INTO "Contact" (organization_id, individual_id, role, is_primary, notes, created_at, updated_at)
        VALUES (org_cloudscale_id, ind_michael_id, 'Decision Maker', TRUE, 'Hiring manager for engineering roles', NOW(), NOW());
    END IF;

    IF org_cloudscale_id IS NOT NULL AND ind_amanda_id IS NOT NULL THEN
        INSERT INTO "Contact" (organization_id, individual_id, role, is_primary, notes, created_at, updated_at)
        VALUES (org_cloudscale_id, ind_amanda_id, 'Technical Contact', FALSE, 'Can provide technical insight', NOW(), NOW());
    END IF;

    IF org_dataflow_id IS NOT NULL AND ind_emily_id IS NOT NULL THEN
        INSERT INTO "Contact" (organization_id, individual_id, role, is_primary, notes, created_at, updated_at)
        VALUES (org_dataflow_id, ind_emily_id, 'Decision Maker', TRUE, 'Leads product hiring', NOW(), NOW());
    END IF;

    IF org_securenet_id IS NOT NULL AND ind_david_id IS NOT NULL THEN
        INSERT INTO "Contact" (organization_id, individual_id, role, is_primary, notes, created_at, updated_at)
        VALUES (org_securenet_id, ind_david_id, 'Decision Maker', TRUE, 'Technical leadership contact', NOW(), NOW());
    END IF;

    IF org_innovatelab_id IS NOT NULL AND ind_jessica_id IS NOT NULL THEN
        INSERT INTO "Contact" (organization_id, individual_id, role, is_primary, notes, created_at, updated_at)
        VALUES (org_innovatelab_id, ind_jessica_id, 'Champion', TRUE, 'Strong advocate, met at conference', NOW(), NOW());
    END IF;

    RAISE NOTICE 'Contact relationships created successfully';
END $$;

-- ==============================
-- STEP 4: Create Sample Deals
-- ==============================
DO $$
DECLARE
    status_prospecting_id BIGINT;
    status_negotiating_id BIGINT;
    status_closed_won_id BIGINT;
BEGIN
    -- Get Status IDs
    SELECT id INTO status_prospecting_id FROM "Status" WHERE name = 'Prospecting' AND category = 'deal' LIMIT 1;
    SELECT id INTO status_negotiating_id FROM "Status" WHERE name = 'Negotiating' AND category = 'deal' LIMIT 1;
    SELECT id INTO status_closed_won_id FROM "Status" WHERE name = 'Closed Won' AND category = 'deal' LIMIT 1;

    -- Create Deals
    INSERT INTO "Deal" (name, description, current_status_id, expected_close_date, created_at, updated_at)
    VALUES
        ('Consulting Engagement - TechVentures', 'Product strategy consulting for portfolio companies', status_negotiating_id, NOW() + INTERVAL '30 days', NOW(), NOW()),
        ('Partnership - CloudScale', 'Technical partnership for cloud infrastructure', status_prospecting_id, NOW() + INTERVAL '60 days', NOW(), NOW()),
        ('Business Development - SecureNet', 'Exploring collaboration opportunities', status_prospecting_id, NOW() + INTERVAL '90 days', NOW(), NOW());

    RAISE NOTICE 'Deals created successfully';
END $$;

-- ==============================
-- STEP 5: Create Sample Leads (Opportunities & Partnerships)
-- ==============================
DO $$
DECLARE
    deal_techventures_id BIGINT;
    deal_cloudscale_id BIGINT;
    deal_securenet_id BIGINT;

    org_techventures_id BIGINT;
    org_cloudscale_id BIGINT;
    org_dataflow_id BIGINT;
    org_securenet_id BIGINT;
    org_innovatelab_id BIGINT;

    acc_techventures_id BIGINT;
    acc_cloudscale_id BIGINT;
    acc_dataflow_id BIGINT;
    acc_securenet_id BIGINT;

    status_qualified_id BIGINT;
    status_proposal_id BIGINT;
    status_nurturing_id BIGINT;

    lead_id BIGINT;
BEGIN
    -- Get Deal IDs
    SELECT id INTO deal_techventures_id FROM "Deal" WHERE name = 'Consulting Engagement - TechVentures' LIMIT 1;
    SELECT id INTO deal_cloudscale_id FROM "Deal" WHERE name = 'Partnership - CloudScale' LIMIT 1;
    SELECT id INTO deal_securenet_id FROM "Deal" WHERE name = 'Business Development - SecureNet' LIMIT 1;

    -- Get Organization IDs and their Account IDs
    SELECT o.id, o.account_id INTO org_techventures_id, acc_techventures_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT o.id, o.account_id INTO org_cloudscale_id, acc_cloudscale_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT o.id, o.account_id INTO org_dataflow_id, acc_dataflow_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT o.id, o.account_id INTO org_securenet_id, acc_securenet_id FROM "Organization" o WHERE LOWER(o.name) = LOWER('SecureNet Solutions') LIMIT 1;
    SELECT id INTO org_innovatelab_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;

    -- Get Status IDs
    SELECT id INTO status_qualified_id FROM "Status" WHERE name = 'Qualified' AND category = 'lead' LIMIT 1;
    SELECT id INTO status_proposal_id FROM "Status" WHERE name = 'Proposal' AND category = 'lead' LIMIT 1;
    SELECT id INTO status_nurturing_id FROM "Status" WHERE name = 'Nurturing' AND category = 'lead' LIMIT 1;

    -- Create Opportunity Lead for TechVentures
    IF deal_techventures_id IS NOT NULL AND acc_techventures_id IS NOT NULL THEN
        INSERT INTO "Lead" (deal_id, account_id, type, title, description, source, current_status_id, created_at, updated_at)
        VALUES (
            deal_techventures_id,
            acc_techventures_id,
            'Opportunity',
            'TechVentures - Product Strategy Consulting',
            'Opportunity to provide product strategy consulting for their portfolio companies. Potential for ongoing engagement.',
            'referral',
            status_proposal_id,
            NOW(),
            NOW()
        )
        RETURNING id INTO lead_id;

        -- Create Opportunity record
        INSERT INTO "Opportunity" (id, opportunity_name, estimated_value, probability, expected_close_date, notes, created_at, updated_at)
        VALUES (
            lead_id,
            'Product Strategy Consulting - Q1 2026',
            75000.00,
            0.70,
            NOW() + INTERVAL '30 days',
            'Met Sarah at conference. She expressed interest in consulting services for portfolio companies.',
            NOW(),
            NOW()
        );
    END IF;

    -- Create Partnership Lead for CloudScale
    IF deal_cloudscale_id IS NOT NULL AND acc_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Lead" (deal_id, account_id, type, title, description, source, current_status_id, created_at, updated_at)
        VALUES (
            deal_cloudscale_id,
            acc_cloudscale_id,
            'Partnership',
            'CloudScale - Technical Partnership',
            'Technical partnership to integrate services. They have strong infrastructure we could leverage.',
            'outbound',
            status_qualified_id,
            NOW(),
            NOW()
        )
        RETURNING id INTO lead_id;

        -- Create Partnership record
        INSERT INTO "Partnership" (id, partnership_type, partnership_name, start_date, notes, created_at, updated_at)
        VALUES (
            lead_id,
            'Technical',
            'CloudScale Infrastructure Integration',
            NOW() + INTERVAL '60 days',
            'Initial conversation with Michael Rodriguez. Need to schedule technical deep dive.',
            NOW(),
            NOW()
        );
    END IF;

    -- Create Opportunity Lead for DataFlow
    IF deal_cloudscale_id IS NOT NULL AND acc_dataflow_id IS NOT NULL THEN
        INSERT INTO "Lead" (deal_id, account_id, type, title, description, source, current_status_id, created_at, updated_at)
        VALUES (
            deal_cloudscale_id,
            acc_dataflow_id,
            'Opportunity',
            'DataFlow - Analytics Consulting',
            'Consulting opportunity for data analytics strategy. Growing company with budget.',
            'inbound',
            status_nurturing_id,
            NOW(),
            NOW()
        )
        RETURNING id INTO lead_id;

        -- Create Opportunity record
        INSERT INTO "Opportunity" (id, opportunity_name, estimated_value, probability, expected_close_date, notes, created_at, updated_at)
        VALUES (
            lead_id,
            'Data Analytics Strategy - Q2 2026',
            50000.00,
            0.40,
            NOW() + INTERVAL '90 days',
            'Emily Johnson reached out via LinkedIn. Early stage conversation.',
            NOW(),
            NOW()
        );
    END IF;

    -- Create Partnership Lead for SecureNet
    IF deal_securenet_id IS NOT NULL AND acc_securenet_id IS NOT NULL THEN
        INSERT INTO "Lead" (deal_id, account_id, type, title, description, source, current_status_id, created_at, updated_at)
        VALUES (
            deal_securenet_id,
            acc_securenet_id,
            'Partnership',
            'SecureNet - Security Integration Partnership',
            'Partnership to integrate security capabilities. Potential for joint go-to-market.',
            'referral',
            status_qualified_id,
            NOW(),
            NOW()
        )
        RETURNING id INTO lead_id;

        -- Create Partnership record
        INSERT INTO "Partnership" (id, partnership_type, partnership_name, start_date, notes, created_at, updated_at)
        VALUES (
            lead_id,
            'Strategic',
            'SecureNet Security Integration',
            NOW() + INTERVAL '90 days',
            'Referred by mutual connection. David Kim is interested in exploring.',
            NOW(),
            NOW()
        );
    END IF;

    RAISE NOTICE 'Leads created successfully';
END $$;

-- ==============================
-- STEP 6: Create Sample Tasks
-- ==============================
DO $$
DECLARE
    lead_techventures_id BIGINT;
    lead_cloudscale_id BIGINT;
    lead_dataflow_id BIGINT;
    lead_securenet_id BIGINT;

    status_todo_id BIGINT;
    status_in_progress_id BIGINT;
    priority_high_id BIGINT;
    priority_medium_id BIGINT;
BEGIN
    -- Get Lead IDs
    SELECT l.id INTO lead_techventures_id
    FROM "Lead" l
    WHERE l.title = 'TechVentures - Product Strategy Consulting'
    LIMIT 1;

    SELECT l.id INTO lead_cloudscale_id
    FROM "Lead" l
    WHERE l.title = 'CloudScale - Technical Partnership'
    LIMIT 1;

    SELECT l.id INTO lead_dataflow_id
    FROM "Lead" l
    WHERE l.title = 'DataFlow - Analytics Consulting'
    LIMIT 1;

    SELECT l.id INTO lead_securenet_id
    FROM "Lead" l
    WHERE l.title = 'SecureNet - Security Integration Partnership'
    LIMIT 1;

    -- Get Status IDs
    SELECT id INTO status_todo_id FROM "Status" WHERE name = 'To Do' AND category = 'task_status' LIMIT 1;
    SELECT id INTO status_in_progress_id FROM "Status" WHERE name = 'In Progress' AND category = 'task_status' LIMIT 1;
    SELECT id INTO priority_high_id FROM "Status" WHERE name = 'High' AND category = 'task_priority' LIMIT 1;
    SELECT id INTO priority_medium_id FROM "Status" WHERE name = 'Medium' AND category = 'task_priority' LIMIT 1;

    -- Create Tasks for TechVentures Opportunity
    IF lead_techventures_id IS NOT NULL THEN
        INSERT INTO "Task" (title, description, taskable_type, taskable_id, current_status_id, priority_id, due_date, created_at, updated_at)
        VALUES
            (
                'Prepare proposal for TechVentures',
                'Create detailed proposal outlining consulting services, timeline, and pricing',
                'Lead',
                lead_techventures_id,
                status_in_progress_id,
                priority_high_id,
                NOW() + INTERVAL '7 days',
                NOW(),
                NOW()
            ),
            (
                'Schedule call with Sarah Chen',
                'Follow-up call to discuss proposal and answer questions',
                'Lead',
                lead_techventures_id,
                status_todo_id,
                priority_high_id,
                NOW() + INTERVAL '10 days',
                NOW(),
                NOW()
            );
    END IF;

    -- Create Tasks for CloudScale Partnership
    IF lead_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Task" (title, description, taskable_type, taskable_id, current_status_id, priority_id, due_date, created_at, updated_at)
        VALUES
            (
                'Technical deep dive with CloudScale',
                'Schedule meeting with engineering team to discuss integration architecture',
                'Lead',
                lead_cloudscale_id,
                status_todo_id,
                priority_medium_id,
                NOW() + INTERVAL '14 days',
                NOW(),
                NOW()
            ),
            (
                'Draft partnership agreement',
                'Work with legal to draft initial partnership terms',
                'Lead',
                lead_cloudscale_id,
                status_todo_id,
                priority_medium_id,
                NOW() + INTERVAL '30 days',
                NOW(),
                NOW()
            );
    END IF;

    -- Create Tasks for DataFlow Opportunity
    IF lead_dataflow_id IS NOT NULL THEN
        INSERT INTO "Task" (title, description, taskable_type, taskable_id, current_status_id, priority_id, due_date, created_at, updated_at)
        VALUES
            (
                'Research DataFlow business model',
                'Understand their current analytics stack and pain points',
                'Lead',
                lead_dataflow_id,
                status_todo_id,
                priority_medium_id,
                NOW() + INTERVAL '5 days',
                NOW(),
                NOW()
            );
    END IF;

    -- Create Tasks for SecureNet Partnership
    IF lead_securenet_id IS NOT NULL THEN
        INSERT INTO "Task" (title, description, taskable_type, taskable_id, current_status_id, priority_id, due_date, created_at, updated_at)
        VALUES
            (
                'Initial call with David Kim',
                'Exploratory conversation to understand SecureNet''s partnership goals',
                'Lead',
                lead_securenet_id,
                status_todo_id,
                priority_high_id,
                NOW() + INTERVAL '7 days',
                NOW(),
                NOW()
            );
    END IF;

    RAISE NOTICE 'Tasks created successfully';
END $$;

-- ==============================
-- STEP 7: Create Sample Activities
-- ==============================
DO $$
DECLARE
    lead_techventures_id BIGINT;
    lead_cloudscale_id BIGINT;
BEGIN
    -- Get Lead IDs
    SELECT l.id INTO lead_techventures_id
    FROM "Lead" l
    WHERE l.title = 'TechVentures - Product Strategy Consulting'
    LIMIT 1;

    SELECT l.id INTO lead_cloudscale_id
    FROM "Lead" l
    WHERE l.title = 'CloudScale - Technical Partnership'
    LIMIT 1;

    -- Create Activities for TechVentures
    IF lead_techventures_id IS NOT NULL THEN
        INSERT INTO "Activity" (activity_type, subject, description, activity_date, activityable_type, activityable_id, created_at, updated_at)
        VALUES
            (
                'Meeting',
                'Initial meeting with Sarah Chen at TechConf 2025',
                'Met Sarah at TechConf. Discussed their need for product strategy consulting for portfolio companies. She seemed very interested and asked for a follow-up.',
                NOW() - INTERVAL '14 days',
                'Lead',
                lead_techventures_id,
                NOW() - INTERVAL '14 days',
                NOW() - INTERVAL '14 days'
            ),
            (
                'Email',
                'Follow-up email to Sarah',
                'Sent follow-up email with case studies and proposed engagement structure. Received positive response.',
                NOW() - INTERVAL '7 days',
                'Lead',
                lead_techventures_id,
                NOW() - INTERVAL '7 days',
                NOW() - INTERVAL '7 days'
            ),
            (
                'Call',
                'Phone call with Sarah and Robert',
                'Great call discussing scope of work. They want a formal proposal. Robert will champion internally.',
                NOW() - INTERVAL '2 days',
                'Lead',
                lead_techventures_id,
                NOW() - INTERVAL '2 days',
                NOW() - INTERVAL '2 days'
            );
    END IF;

    -- Create Activities for CloudScale
    IF lead_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Activity" (activity_type, subject, description, activity_date, activityable_type, activityable_id, created_at, updated_at)
        VALUES
            (
                'Email',
                'Cold outreach to Michael Rodriguez',
                'Sent initial outreach email about potential partnership opportunity.',
                NOW() - INTERVAL '10 days',
                'Lead',
                lead_cloudscale_id,
                NOW() - INTERVAL '10 days',
                NOW() - INTERVAL '10 days'
            ),
            (
                'Call',
                'Exploratory call with Michael',
                'Good first conversation. Michael is interested but wants to involve engineering team before proceeding.',
                NOW() - INTERVAL '5 days',
                'Lead',
                lead_cloudscale_id,
                NOW() - INTERVAL '5 days',
                NOW() - INTERVAL '5 days'
            );
    END IF;

    RAISE NOTICE 'Activities created successfully';
END $$;

-- ==============================
-- STEP 8: Create Organization Relationships
-- ==============================
DO $$
DECLARE
    org_techventures_id BIGINT;
    org_innovatelab_id BIGINT;
    org_cloudscale_id BIGINT;
    org_dataflow_id BIGINT;
BEGIN
    -- Get Organization IDs
    SELECT id INTO org_techventures_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT id INTO org_innovatelab_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;
    SELECT id INTO org_cloudscale_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT id INTO org_dataflow_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;

    -- TechVentures invested in InnovateLab
    IF org_techventures_id IS NOT NULL AND org_innovatelab_id IS NOT NULL THEN
        INSERT INTO "Organization_Relationship" (from_organization_id, to_organization_id, relationship_type, notes, created_at, updated_at)
        VALUES (
            org_techventures_id,
            org_innovatelab_id,
            'Investor',
            'TechVentures led Series A round in 2024',
            NOW(),
            NOW()
        );
    END IF;

    -- CloudScale is a customer of DataFlow
    IF org_cloudscale_id IS NOT NULL AND org_dataflow_id IS NOT NULL THEN
        INSERT INTO "Organization_Relationship" (from_organization_id, to_organization_id, relationship_type, notes, created_at, updated_at)
        VALUES (
            org_cloudscale_id,
            org_dataflow_id,
            'Customer',
            'CloudScale uses DataFlow for internal analytics',
            NOW(),
            NOW()
        );
    END IF;

    RAISE NOTICE 'Organization relationships created successfully';
END $$;

-- ==============================
-- Return counts for seeder verification
-- ==============================
SELECT
  (SELECT COUNT(*) FROM "Organization") as organizations,
  (SELECT COUNT(*) FROM "Individual") as individuals,
  (SELECT COUNT(*) FROM "Contact") as contacts,
  (SELECT COUNT(*) FROM "Deal") as deals,
  (SELECT COUNT(*) FROM "Lead" WHERE type IN ('Opportunity', 'Partnership')) as leads,
  (SELECT COUNT(*) FROM "Task") as tasks,
  (SELECT COUNT(*) FROM "Activity") as activities;

-- ==============================
-- Verification Queries (commented out)
-- ==============================
-- Uncomment to verify seed data:

-- SELECT COUNT(*) as total_organizations FROM "Organization";
-- SELECT COUNT(*) as total_individuals FROM "Individual";
-- SELECT COUNT(*) as total_contacts FROM "Contact";
-- SELECT COUNT(*) as total_deals FROM "Deal";
-- SELECT COUNT(*) as total_leads FROM "Lead";
-- SELECT COUNT(*) as total_opportunities FROM "Opportunity";
-- SELECT COUNT(*) as total_partnerships FROM "Partnership";
-- SELECT COUNT(*) as total_tasks FROM "Task";
-- SELECT COUNT(*) as total_activities FROM "Activity";

-- View full lead details with relationships:
-- SELECT
--     l.title as lead_title,
--     l.type as lead_type,
--     d.name as deal_name,
--     s.name as current_status,
--     o.name as organization_name,
--     COUNT(DISTINCT t.id) as task_count,
--     COUNT(DISTINCT a.id) as activity_count
-- FROM "Lead" l
-- LEFT JOIN "Deal" d ON l.deal_id = d.id
-- LEFT JOIN "Status" s ON l.current_status_id = s.id
-- LEFT JOIN "Opportunity" opp ON opp.id = l.id AND l.type = 'Opportunity'
-- LEFT JOIN "Partnership" p ON p.id = l.id AND l.type = 'Partnership'
-- LEFT JOIN "Organization" o ON o.id = COALESCE(opp.organization_id, p.organization_id)
-- LEFT JOIN "Task" t ON t.taskable_type = 'Lead' AND t.taskable_id = l.id
-- LEFT JOIN "Activity" a ON a.activityable_type = 'Lead' AND a.activityable_id = l.id
-- WHERE l.type IN ('Opportunity', 'Partnership')
-- GROUP BY l.id, l.title, l.type, d.name, s.name, o.name;

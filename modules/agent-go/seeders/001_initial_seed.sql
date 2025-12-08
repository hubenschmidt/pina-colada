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
    v_bootstrap_user_id BIGINT;
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

    -- STEP 0a: Create bootstrap User first (without individual_id) to use for created_by/updated_by
    SELECT id INTO v_bootstrap_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    IF v_bootstrap_user_id IS NULL THEN
        INSERT INTO "User" (tenant_id, email, first_name, last_name, status, created_at, updated_at)
        VALUES (v_tenant_id, 'whubenschmidt@gmail.com', 'William', 'Hubenschmidt', 'active', NOW(), NOW())
        RETURNING id INTO v_bootstrap_user_id;
    ELSE
        UPDATE "User" SET tenant_id = v_tenant_id WHERE id = v_bootstrap_user_id AND tenant_id IS NULL;
    END IF;

    -- Create Account for the user's Individual (with tenant_id and audit columns)
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by)
    VALUES (v_tenant_id, 'William Hubenschmidt', NOW(), NOW(), v_bootstrap_user_id, v_bootstrap_user_id)
    RETURNING id INTO v_account_id;

    -- Create or update Individual for the user
    SELECT id INTO v_individual_id FROM "Individual" WHERE LOWER(email) = LOWER('whubenschmidt@gmail.com') LIMIT 1;

    IF v_individual_id IS NULL THEN
        -- Individual doesn't exist, create new
        INSERT INTO "Individual" (account_id, first_name, last_name, email, created_at, updated_at, created_by, updated_by)
        VALUES (v_account_id, 'William', 'Hubenschmidt', 'whubenschmidt@gmail.com', NOW(), NOW(), v_bootstrap_user_id, v_bootstrap_user_id)
        RETURNING id INTO v_individual_id;
    ELSE
        -- Individual exists, update with account
        UPDATE "Individual"
        SET account_id = v_account_id,
            first_name = 'William',
            last_name = 'Hubenschmidt',
            updated_at = NOW(),
            updated_by = v_bootstrap_user_id
        WHERE id = v_individual_id;
    END IF;

    -- Now link the User to the Individual
    UPDATE "User" SET individual_id = v_individual_id WHERE id = v_bootstrap_user_id;
    v_user_id := v_bootstrap_user_id;

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
    INSERT INTO "User_Role" (user_id, role_id, created_at)
    VALUES (v_user_id, v_role_id, NOW())
    ON CONFLICT DO NOTHING;

    -- Create Developer Role (global)
    INSERT INTO "Role" (tenant_id, name, description)
    VALUES (NULL, 'developer', 'Developer access with analytics and debugging tools')
    ON CONFLICT DO NOTHING;

    -- Assign developer role to William
    SELECT id INTO v_role_id FROM "Role" WHERE name = 'developer' AND tenant_id IS NULL;
    IF v_role_id IS NOT NULL THEN
        INSERT INTO "User_Role" (user_id, role_id, created_at)
        VALUES (v_user_id, v_role_id, NOW())
        ON CONFLICT DO NOTHING;
    END IF;

    -- Set William's theme preference to dark
    INSERT INTO "User_Preferences" (user_id, theme, timezone, created_at, updated_at)
    VALUES (v_user_id, 'dark', 'America/New_York', NOW(), NOW())
    ON CONFLICT (user_id) DO UPDATE SET theme = 'dark', updated_at = NOW();

    -- Create default Organization for the tenant
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by)
    VALUES (v_tenant_id, 'PinaColada', NOW(), NOW(), v_bootstrap_user_id, v_bootstrap_user_id)
    RETURNING id INTO v_account_id;

    INSERT INTO "Organization" (account_id, name, created_at, updated_at, created_by, updated_by)
    VALUES (v_account_id, 'PinaColada', NOW(), NOW(), v_bootstrap_user_id, v_bootstrap_user_id)
    ON CONFLICT ((LOWER(name))) DO NOTHING;

    RAISE NOTICE 'Default tenant and user created successfully. Tenant ID: %, User ID: %', v_tenant_id, v_user_id;

    -- ==============================
    -- Create Second User: Jennifer Lev
    -- ==============================
    -- Create Account for Jennifer
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by)
    VALUES (v_tenant_id, 'Jennifer Lev', NOW(), NOW(), v_bootstrap_user_id, v_bootstrap_user_id)
    RETURNING id INTO v_account_id;

    -- Create Individual for Jennifer
    SELECT id INTO v_individual_id FROM "Individual" WHERE LOWER(email) = LOWER('jennifervlev@gmail.com') LIMIT 1;

    IF v_individual_id IS NULL THEN
        INSERT INTO "Individual" (account_id, first_name, last_name, email, created_at, updated_at, created_by, updated_by)
        VALUES (v_account_id, 'Jennifer', 'Lev', 'jennifervlev@gmail.com', NOW(), NOW(), v_bootstrap_user_id, v_bootstrap_user_id)
        RETURNING id INTO v_individual_id;
    ELSE
        UPDATE "Individual"
        SET account_id = v_account_id,
            first_name = 'Jennifer',
            last_name = 'Lev',
            updated_at = NOW(),
            updated_by = v_bootstrap_user_id
        WHERE id = v_individual_id;
    END IF;

    -- Create or update User for Jennifer
    SELECT id INTO v_user_id FROM "User" WHERE email = 'jennifervlev@gmail.com' LIMIT 1;

    IF v_user_id IS NULL THEN
        INSERT INTO "User" (tenant_id, individual_id, email, first_name, last_name, status, created_at, updated_at)
        VALUES (v_tenant_id, v_individual_id, 'jennifervlev@gmail.com', 'Jennifer', 'Lev', 'active', NOW(), NOW())
        RETURNING id INTO v_user_id;
    ELSE
        UPDATE "User"
        SET tenant_id = v_tenant_id,
            individual_id = v_individual_id,
            first_name = 'Jennifer',
            last_name = 'Lev',
            updated_at = NOW()
        WHERE id = v_user_id;
    END IF;

    -- Create member role if it doesn't exist
    SELECT id INTO v_role_id FROM "Role" WHERE tenant_id = v_tenant_id AND name = 'member' LIMIT 1;
    IF v_role_id IS NULL THEN
        INSERT INTO "Role" (tenant_id, name, description)
        VALUES (v_tenant_id, 'member', 'Standard team member access')
        RETURNING id INTO v_role_id;
    END IF;

    -- Assign member role to Jennifer
    INSERT INTO "User_Role" (user_id, role_id, created_at)
    VALUES (v_user_id, v_role_id, NOW())
    ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Second user Jennifer Lev created successfully. User ID: %', v_user_id;
END $$;

-- ==============================
-- STEP 1: Create Sample Organizations (with Accounts and Industries)
-- ==============================
DO $$
DECLARE
    account_id BIGINT;
    org_id BIGINT;
    v_tenant_id BIGINT;
    v_user_id BIGINT;
    industry_id BIGINT;
BEGIN
    -- Get the tenant ID and bootstrap user
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    -- TechVentures Inc
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by) VALUES (v_tenant_id, 'TechVentures Inc', NOW(), NOW(), v_user_id, v_user_id) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at, created_by, updated_by)
    VALUES (account_id, 'TechVentures Inc', 'https://techventures.example.com', 50, 'Early-stage technology investor', NOW(), NOW(), v_user_id, v_user_id)
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Venture Capital';
    INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (account_id, industry_id) ON CONFLICT DO NOTHING;

    -- CloudScale Systems
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by) VALUES (v_tenant_id, 'CloudScale Systems', NOW(), NOW(), v_user_id, v_user_id) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at, created_by, updated_by)
    VALUES (account_id, 'CloudScale Systems', 'https://cloudscale.example.com', 200, 'Enterprise cloud solutions provider', NOW(), NOW(), v_user_id, v_user_id)
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Cloud Infrastructure';
    INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (account_id, industry_id) ON CONFLICT DO NOTHING;

    -- DataFlow Analytics
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by) VALUES (v_tenant_id, 'DataFlow Analytics', NOW(), NOW(), v_user_id, v_user_id) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at, created_by, updated_by)
    VALUES (account_id, 'DataFlow Analytics', 'https://dataflow.example.com', 75, 'Real-time data analytics platform', NOW(), NOW(), v_user_id, v_user_id)
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Data Analytics';
    INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (account_id, industry_id) ON CONFLICT DO NOTHING;

    -- SecureNet Solutions
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by) VALUES (v_tenant_id, 'SecureNet Solutions', NOW(), NOW(), v_user_id, v_user_id) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at, created_by, updated_by)
    VALUES (account_id, 'SecureNet Solutions', 'https://securenet.example.com', 150, 'Enterprise security software', NOW(), NOW(), v_user_id, v_user_id)
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Cybersecurity';
    INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (account_id, industry_id) ON CONFLICT DO NOTHING;

    -- InnovateLab
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by) VALUES (v_tenant_id, 'InnovateLab', NOW(), NOW(), v_user_id, v_user_id) RETURNING id INTO account_id;
    INSERT INTO "Organization" (account_id, name, website, employee_count, description, created_at, updated_at, created_by, updated_by)
    VALUES (account_id, 'InnovateLab', 'https://innovatelab.example.com', 30, 'Innovation consulting for startups', NOW(), NOW(), v_user_id, v_user_id)
    ON CONFLICT ((LOWER(name))) DO NOTHING;
    SELECT id INTO org_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab');
    SELECT id INTO industry_id FROM "Industry" WHERE name = 'Consulting';
    INSERT INTO "Account_Industry" (account_id, industry_id) VALUES (account_id, industry_id) ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Organizations with Accounts and Industries created successfully';
END $$;

-- ==============================
-- STEP 2: Create Sample Individuals (with Accounts)
-- These are people tracked independently, NOT organization contacts
-- ==============================
DO $$
DECLARE
    account_id BIGINT;
    v_tenant_id BIGINT;
    v_user_id BIGINT;
BEGIN
    -- Get the tenant ID and bootstrap user
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    -- Alex Thompson - Independent consultant
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by) VALUES (v_tenant_id, 'Alex Thompson', NOW(), NOW(), v_user_id, v_user_id) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, description, created_at, updated_at, created_by, updated_by)
    VALUES (account_id, 'Alex', 'Thompson', 'alex.thompson@gmail.com', '+1-415-555-0201', 'https://linkedin.com/in/alexthompson', 'Independent Consultant', 'Met at TechCrunch Disrupt 2024. Interested in advisory roles.', NOW(), NOW(), v_user_id, v_user_id)
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    -- Maria Garcia - Freelance designer
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by) VALUES (v_tenant_id, 'Maria Garcia', NOW(), NOW(), v_user_id, v_user_id) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, description, created_at, updated_at, created_by, updated_by)
    VALUES (account_id, 'Maria', 'Garcia', 'maria.garcia@designstudio.com', '+1-650-555-0202', 'https://linkedin.com/in/mariagarcia', 'UX Designer', 'Freelance designer, available for contract work.', NOW(), NOW(), v_user_id, v_user_id)
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    -- James Wilson - Mentor/Advisor
    INSERT INTO "Account" (tenant_id, name, created_at, updated_at, created_by, updated_by) VALUES (v_tenant_id, 'James Wilson', NOW(), NOW(), v_user_id, v_user_id) RETURNING id INTO account_id;
    INSERT INTO "Individual" (account_id, first_name, last_name, email, phone, linkedin_url, title, description, created_at, updated_at, created_by, updated_by)
    VALUES (account_id, 'James', 'Wilson', 'james.wilson@advisors.co', '+1-510-555-0203', 'https://linkedin.com/in/jameswilson', 'Startup Advisor', 'Former CTO, now advising early-stage startups.', NOW(), NOW(), v_user_id, v_user_id)
    ON CONFLICT ((LOWER(email))) DO NOTHING;

    RAISE NOTICE 'Individuals with Accounts created successfully';
END $$;

-- ==============================
-- STEP 3: Create Contacts (linked to Accounts via Contact_Account)
-- Note: Contacts at organizations do NOT create Individual records.
-- Individuals are only created explicitly when needed.
-- ==============================
DO $$
DECLARE
    acc_techventures_id BIGINT;
    acc_cloudscale_id BIGINT;
    acc_dataflow_id BIGINT;
    acc_securenet_id BIGINT;
    acc_innovatelab_id BIGINT;
    contact_id BIGINT;
    v_user_id BIGINT;
BEGIN
    -- Get bootstrap user
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;
    -- Get Account IDs for Organizations
    SELECT account_id INTO acc_techventures_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT account_id INTO acc_cloudscale_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT account_id INTO acc_dataflow_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT account_id INTO acc_securenet_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    SELECT account_id INTO acc_innovatelab_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;

    -- Sarah Chen at TechVentures
    IF acc_techventures_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Sarah', 'Chen', 'Partner', 'Investments', 'Decision Maker', 'sarah.chen@techventures.example.com', '+1-415-555-0101', TRUE, 'Main contact for investment discussions', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_techventures_id, TRUE, NOW());
    END IF;

    -- Robert Taylor at TechVentures
    IF acc_techventures_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Robert', 'Taylor', 'Associate', 'Investments', 'Influencer', 'robert.t@techventures.example.com', '+1-415-555-0106', FALSE, 'Can provide warm introductions', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_techventures_id, FALSE, NOW());
    END IF;

    -- Michael Rodriguez at CloudScale
    IF acc_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Michael', 'Rodriguez', 'VP of Engineering', 'Engineering', 'Decision Maker', 'michael.r@cloudscale.example.com', '+1-415-555-0102', TRUE, 'Hiring manager for engineering roles', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_cloudscale_id, TRUE, NOW());
    END IF;

    -- Amanda Brown at CloudScale
    IF acc_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Amanda', 'Brown', 'Senior Software Engineer', 'Engineering', 'Technical Contact', 'amanda.b@cloudscale.example.com', '+1-415-555-0107', FALSE, 'Can provide technical insight', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_cloudscale_id, FALSE, NOW());
    END IF;

    -- Emily Johnson at DataFlow
    IF acc_dataflow_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Emily', 'Johnson', 'Head of Product', 'Product', 'Decision Maker', 'emily.j@dataflow.example.com', '+1-650-555-0103', TRUE, 'Leads product hiring', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_dataflow_id, TRUE, NOW());
    END IF;

    -- David Kim at SecureNet
    IF acc_securenet_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('David', 'Kim', 'CTO', 'Technology', 'Decision Maker', 'david.kim@securenet.example.com', '+1-408-555-0104', TRUE, 'Technical leadership contact', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_securenet_id, TRUE, NOW());
    END IF;

    -- Jessica Williams at InnovateLab
    IF acc_innovatelab_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Jessica', 'Williams', 'Principal Consultant', 'Consulting', 'Champion', 'jessica.w@innovatelab.example.com', '+1-510-555-0105', TRUE, 'Strong advocate, met at conference', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_innovatelab_id, TRUE, NOW());
    END IF;

    -- Generic/department contacts

    -- Reception at TechVentures
    IF acc_techventures_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Front', 'Desk', 'Reception', 'Administration', 'Gatekeeper', 'info@techventures.example.com', '+1-415-555-0100', FALSE, 'General inquiries contact', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_techventures_id, FALSE, NOW());
    END IF;

    -- Legal department contact at CloudScale
    IF acc_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Legal', 'Department', 'General Counsel', 'Legal', 'Legal Contact', 'legal@cloudscale.example.com', '+1-415-555-0200', FALSE, 'For contracts and agreements', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_cloudscale_id, FALSE, NOW());
    END IF;

    -- Support team at DataFlow
    IF acc_dataflow_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at, created_by, updated_by)
        VALUES ('Support', 'Team', 'Customer Support', 'Support', 'Support Contact', 'support@dataflow.example.com', '+1-650-555-0200', FALSE, 'Technical support inquiries', NOW(), NOW(), v_user_id, v_user_id)
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_dataflow_id, FALSE, NOW());
    END IF;

    RAISE NOTICE 'Contacts linked to Accounts created successfully';
END $$;

-- ==============================
-- STEP 4: Create Sample Deals
-- ==============================
DO $$
DECLARE
    status_prospecting_id BIGINT;
    status_negotiating_id BIGINT;
    status_closed_won_id BIGINT;
    v_user_id BIGINT;
    v_tenant_id BIGINT;
BEGIN
    -- Get bootstrap user and tenant
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;
    -- Get Status IDs
    SELECT id INTO status_prospecting_id FROM "Status" WHERE name = 'Prospecting' AND category = 'deal' LIMIT 1;
    SELECT id INTO status_negotiating_id FROM "Status" WHERE name = 'Negotiating' AND category = 'deal' LIMIT 1;
    SELECT id INTO status_closed_won_id FROM "Status" WHERE name = 'Closed Won' AND category = 'deal' LIMIT 1;

    -- Create Deals
    INSERT INTO "Deal" (tenant_id, name, description, current_status_id, expected_close_date, created_at, updated_at, created_by, updated_by)
    VALUES
        (v_tenant_id, 'Consulting Engagement - TechVentures', 'Product strategy consulting for portfolio companies', status_negotiating_id, NOW() + INTERVAL '30 days', NOW(), NOW(), v_user_id, v_user_id),
        (v_tenant_id, 'Partnership - CloudScale', 'Technical partnership for cloud infrastructure', status_prospecting_id, NOW() + INTERVAL '60 days', NOW(), NOW(), v_user_id, v_user_id),
        (v_tenant_id, 'Business Development - SecureNet', 'Exploring collaboration opportunities', status_prospecting_id, NOW() + INTERVAL '90 days', NOW(), NOW(), v_user_id, v_user_id);

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
    v_user_id BIGINT;
    v_tenant_id BIGINT;
BEGIN
    -- Get bootstrap user and tenant
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;
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
        INSERT INTO "Lead" (tenant_id, deal_id, account_id, type, description, source, current_status_id, created_at, updated_at, created_by, updated_by)
        VALUES (
            v_tenant_id,
            deal_techventures_id,
            acc_techventures_id,
            'Opportunity',
            'Opportunity to provide product strategy consulting for their portfolio companies. Potential for ongoing engagement.',
            'referral',
            status_proposal_id,
            NOW(),
            NOW(),
            v_user_id,
            v_user_id
        )
        RETURNING id INTO lead_id;

        -- Create Opportunity record
        INSERT INTO "Opportunity" (id, opportunity_name, estimated_value, probability, expected_close_date, description, created_at, updated_at)
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
        INSERT INTO "Lead" (tenant_id, deal_id, account_id, type, description, source, current_status_id, created_at, updated_at, created_by, updated_by)
        VALUES (
            v_tenant_id,
            deal_cloudscale_id,
            acc_cloudscale_id,
            'Partnership',
            'Technical partnership to integrate services. They have strong infrastructure we could leverage.',
            'outbound',
            status_qualified_id,
            NOW(),
            NOW(),
            v_user_id,
            v_user_id
        )
        RETURNING id INTO lead_id;

        -- Create Partnership record
        INSERT INTO "Partnership" (id, partnership_type, partnership_name, start_date, description, created_at, updated_at)
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
        INSERT INTO "Lead" (tenant_id, deal_id, account_id, type, description, source, current_status_id, created_at, updated_at, created_by, updated_by)
        VALUES (
            v_tenant_id,
            deal_cloudscale_id,
            acc_dataflow_id,
            'Opportunity',
            'Consulting opportunity for data analytics strategy. Growing company with budget.',
            'inbound',
            status_nurturing_id,
            NOW(),
            NOW(),
            v_user_id,
            v_user_id
        )
        RETURNING id INTO lead_id;

        -- Create Opportunity record
        INSERT INTO "Opportunity" (id, opportunity_name, estimated_value, probability, expected_close_date, description, created_at, updated_at)
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
        INSERT INTO "Lead" (tenant_id, deal_id, account_id, type, description, source, current_status_id, created_at, updated_at, created_by, updated_by)
        VALUES (
            v_tenant_id,
            deal_securenet_id,
            acc_securenet_id,
            'Partnership',
            'Partnership to integrate security capabilities. Potential for joint go-to-market.',
            'referral',
            status_qualified_id,
            NOW(),
            NOW(),
            v_user_id,
            v_user_id
        )
        RETURNING id INTO lead_id;

        -- Create Partnership record
        INSERT INTO "Partnership" (id, partnership_type, partnership_name, start_date, description, created_at, updated_at)
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
    v_user_id BIGINT;
    v_tenant_id BIGINT;
BEGIN
    -- Get bootstrap user and tenant
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;
    -- Get Lead IDs via Opportunity/Partnership names
    SELECT l.id INTO lead_techventures_id
    FROM "Lead" l
    JOIN "Opportunity" o ON o.id = l.id
    WHERE o.opportunity_name = 'Product Strategy Consulting - Q1 2026'
    LIMIT 1;

    SELECT l.id INTO lead_cloudscale_id
    FROM "Lead" l
    JOIN "Partnership" p ON p.id = l.id
    WHERE p.partnership_name = 'CloudScale Infrastructure Integration'
    LIMIT 1;

    SELECT l.id INTO lead_dataflow_id
    FROM "Lead" l
    JOIN "Opportunity" o ON o.id = l.id
    WHERE o.opportunity_name = 'Data Analytics Strategy - Q2 2026'
    LIMIT 1;

    SELECT l.id INTO lead_securenet_id
    FROM "Lead" l
    JOIN "Partnership" p ON p.id = l.id
    WHERE p.partnership_name = 'SecureNet Security Integration'
    LIMIT 1;

    -- Get Status IDs
    SELECT id INTO status_todo_id FROM "Status" WHERE name = 'To Do' AND category = 'task_status' LIMIT 1;
    SELECT id INTO status_in_progress_id FROM "Status" WHERE name = 'In Progress' AND category = 'task_status' LIMIT 1;
    SELECT id INTO priority_high_id FROM "Status" WHERE name = 'High' AND category = 'task_priority' LIMIT 1;
    SELECT id INTO priority_medium_id FROM "Status" WHERE name = 'Medium' AND category = 'task_priority' LIMIT 1;

    -- Create Tasks for TechVentures Opportunity
    IF lead_techventures_id IS NOT NULL THEN
        INSERT INTO "Task" (tenant_id, title, description, taskable_type, taskable_id, current_status_id, priority_id, due_date, created_at, updated_at, created_by, updated_by)
        VALUES
            (
                v_tenant_id,
                'Prepare proposal for TechVentures',
                'Create detailed proposal outlining consulting services, timeline, and pricing',
                'Lead',
                lead_techventures_id,
                status_in_progress_id,
                priority_high_id,
                NOW() + INTERVAL '7 days',
                NOW(),
                NOW(),
                v_user_id,
                v_user_id
            ),
            (
                v_tenant_id,
                'Schedule call with Sarah Chen',
                'Follow-up call to discuss proposal and answer questions',
                'Lead',
                lead_techventures_id,
                status_todo_id,
                priority_high_id,
                NOW() + INTERVAL '10 days',
                NOW(),
                NOW(),
                v_user_id,
                v_user_id
            );
    END IF;

    -- Create Tasks for CloudScale Partnership
    IF lead_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Task" (tenant_id, title, description, taskable_type, taskable_id, current_status_id, priority_id, due_date, created_at, updated_at, created_by, updated_by)
        VALUES
            (
                v_tenant_id,
                'Technical deep dive with CloudScale',
                'Schedule meeting with engineering team to discuss integration architecture',
                'Lead',
                lead_cloudscale_id,
                status_todo_id,
                priority_medium_id,
                NOW() + INTERVAL '14 days',
                NOW(),
                NOW(),
                v_user_id,
                v_user_id
            ),
            (
                v_tenant_id,
                'Draft partnership agreement',
                'Work with legal to draft initial partnership terms',
                'Lead',
                lead_cloudscale_id,
                status_todo_id,
                priority_medium_id,
                NOW() + INTERVAL '30 days',
                NOW(),
                NOW(),
                v_user_id,
                v_user_id
            );
    END IF;

    -- Create Tasks for DataFlow Opportunity
    IF lead_dataflow_id IS NOT NULL THEN
        INSERT INTO "Task" (tenant_id, title, description, taskable_type, taskable_id, current_status_id, priority_id, due_date, created_at, updated_at, created_by, updated_by)
        VALUES
            (
                v_tenant_id,
                'Research DataFlow business model',
                'Understand their current analytics stack and pain points',
                'Lead',
                lead_dataflow_id,
                status_todo_id,
                priority_medium_id,
                NOW() + INTERVAL '5 days',
                NOW(),
                NOW(),
                v_user_id,
                v_user_id
            );
    END IF;

    -- Create Tasks for SecureNet Partnership
    IF lead_securenet_id IS NOT NULL THEN
        INSERT INTO "Task" (tenant_id, title, description, taskable_type, taskable_id, current_status_id, priority_id, due_date, created_at, updated_at, created_by, updated_by)
        VALUES
            (
                v_tenant_id,
                'Initial call with David Kim',
                'Exploratory conversation to understand SecureNet''s partnership goals',
                'Lead',
                lead_securenet_id,
                status_todo_id,
                priority_high_id,
                NOW() + INTERVAL '7 days',
                NOW(),
                NOW(),
                v_user_id,
                v_user_id
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
    -- Get Lead IDs via Opportunity/Partnership names
    SELECT l.id INTO lead_techventures_id
    FROM "Lead" l
    JOIN "Opportunity" o ON o.id = l.id
    WHERE o.opportunity_name = 'Product Strategy Consulting - Q1 2026'
    LIMIT 1;

    SELECT l.id INTO lead_cloudscale_id
    FROM "Lead" l
    JOIN "Partnership" p ON p.id = l.id
    WHERE p.partnership_name = 'CloudScale Infrastructure Integration'
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

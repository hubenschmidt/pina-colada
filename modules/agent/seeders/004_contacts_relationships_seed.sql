-- Seeder: Additional Contacts and Relationships for Accounts
-- This seeder adds contacts linked to Accounts (which contain Organizations and Individuals)
--
-- IMPORTANT: This runs AFTER 001_initial_seed.sql

-- ==============================
-- STEP 1: Add Contacts to Individual Accounts
-- ==============================
DO $$
DECLARE
    acc_alex_id BIGINT;
    acc_maria_id BIGINT;
    acc_james_id BIGINT;
    contact_id BIGINT;
    v_user_id BIGINT;
BEGIN
    -- Get bootstrap user for audit columns
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    -- Get Account IDs for Individuals (via Individual's account_id)
    SELECT account_id INTO acc_alex_id FROM "Individual" WHERE LOWER(email) = LOWER('alex.thompson@gmail.com') LIMIT 1;
    SELECT account_id INTO acc_maria_id FROM "Individual" WHERE LOWER(email) = LOWER('maria.garcia@designstudio.com') LIMIT 1;
    SELECT account_id INTO acc_james_id FROM "Individual" WHERE LOWER(email) = LOWER('james.wilson@advisors.co') LIMIT 1;

    -- Alex Thompson's contacts
    IF acc_alex_id IS NOT NULL THEN
        -- His assistant
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Lisa', 'Park', 'Executive Assistant', NULL, 'Assistant', 'lisa.park@alexconsulting.com', '+1-415-555-0301', TRUE, 'Primary contact for scheduling meetings with Alex', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_alex_id, TRUE, NOW());

        -- A referral partner
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Tom', 'Mitchell', 'Business Development', NULL, 'Referral Partner', 'tom.m@referralnetwork.com', '+1-415-555-0302', FALSE, 'Refers consulting clients to Alex', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_alex_id, FALSE, NOW());
    END IF;

    -- Maria Garcia's contacts
    IF acc_maria_id IS NOT NULL THEN
        -- Her agent
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Rachel', 'Adams', 'Talent Agent', NULL, 'Agent', 'rachel@creativetalent.com', '+1-650-555-0303', TRUE, 'Handles contract negotiations for Maria', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_maria_id, TRUE, NOW());
    END IF;

    -- James Wilson's contacts
    IF acc_james_id IS NOT NULL THEN
        -- His co-advisor
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Patricia', 'Lee', 'Startup Advisor', NULL, 'Co-Advisor', 'patricia@advisors.co', '+1-510-555-0304', FALSE, 'Co-advises with James on fintech startups', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_james_id, FALSE, NOW());

        -- His administrative contact
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Kevin', 'Nguyen', 'Office Manager', NULL, 'Administrative', 'kevin@advisors.co', '+1-510-555-0305', TRUE, 'Manages James office schedule and logistics', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_james_id, TRUE, NOW());
    END IF;

    RAISE NOTICE 'Individual account contacts created successfully';
END $$;

-- ==============================
-- STEP 2: Add More Contacts to Organization Accounts
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
    -- Get bootstrap user for audit columns
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    -- Get Account IDs for Organizations (via Organization's account_id)
    SELECT account_id INTO acc_techventures_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT account_id INTO acc_cloudscale_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT account_id INTO acc_dataflow_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT account_id INTO acc_securenet_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    SELECT account_id INTO acc_innovatelab_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;

    -- Additional TechVentures contacts
    IF acc_techventures_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Brian', 'Cooper', 'Managing Director', 'Investments', 'Executive', 'brian.c@techventures.example.com', '+1-415-555-0110', FALSE, 'Oversees all portfolio companies', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_techventures_id, FALSE, NOW());

        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Nina', 'Patel', 'Operations Manager', 'Operations', 'Operations', 'nina.p@techventures.example.com', '+1-415-555-0111', FALSE, 'Handles logistics and event planning', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_techventures_id, FALSE, NOW());
    END IF;

    -- Additional DataFlow contacts
    IF acc_dataflow_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Marcus', 'Chen', 'Lead Data Scientist', 'Engineering', 'Technical Contact', 'marcus.c@dataflow.example.com', '+1-650-555-0210', FALSE, 'Technical expert for data pipeline questions', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_dataflow_id, FALSE, NOW());

        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Stephanie', 'Wright', 'Sales Director', 'Sales', 'Sales Contact', 'stephanie.w@dataflow.example.com', '+1-650-555-0211', FALSE, 'For partnership and sales discussions', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_dataflow_id, FALSE, NOW());
    END IF;

    -- Additional SecureNet contacts
    IF acc_securenet_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Andrew', 'Martinez', 'Head of Sales', 'Sales', 'Decision Maker', 'andrew.m@securenet.example.com', '+1-408-555-0210', FALSE, 'Handles enterprise sales', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_securenet_id, FALSE, NOW());
    END IF;

    -- Additional InnovateLab contacts
    IF acc_innovatelab_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Diana', 'Ross', 'Senior Consultant', 'Consulting', 'Technical Contact', 'diana.r@innovatelab.example.com', '+1-510-555-0210', FALSE, 'Specialist in product-market fit', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_innovatelab_id, FALSE, NOW());

        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Chris', 'Anderson', 'Founder & CEO', 'Executive', 'Decision Maker', 'chris.a@innovatelab.example.com', '+1-510-555-0211', FALSE, 'Final decision maker for large engagements', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_innovatelab_id, FALSE, NOW());
    END IF;

    RAISE NOTICE 'Additional organization contacts created successfully';
END $$;

-- ==============================
-- STEP 3: Create Contacts Linked to Multiple Accounts
-- Some contacts work across multiple accounts
-- ==============================
DO $$
DECLARE
    acc_cloudscale_id BIGINT;
    acc_dataflow_id BIGINT;
    acc_securenet_id BIGINT;
    acc_techventures_id BIGINT;
    acc_innovatelab_id BIGINT;
    acc_james_id BIGINT;
    acc_alex_id BIGINT;
    acc_maria_id BIGINT;
    contact_id BIGINT;
    v_user_id BIGINT;
BEGIN
    -- Get bootstrap user for audit columns
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    -- Get Organization Account IDs
    SELECT account_id INTO acc_cloudscale_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT account_id INTO acc_dataflow_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT account_id INTO acc_securenet_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    SELECT account_id INTO acc_techventures_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT account_id INTO acc_innovatelab_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;

    -- Get Individual Account IDs
    SELECT account_id INTO acc_james_id FROM "Individual" WHERE LOWER(email) = LOWER('james.wilson@advisors.co') LIMIT 1;
    SELECT account_id INTO acc_alex_id FROM "Individual" WHERE LOWER(email) = LOWER('alex.thompson@gmail.com') LIMIT 1;
    SELECT account_id INTO acc_maria_id FROM "Individual" WHERE LOWER(email) = LOWER('maria.garcia@designstudio.com') LIMIT 1;

    -- James Wilson advises TechVentures (Contact linked to both accounts)
    IF acc_james_id IS NOT NULL AND acc_techventures_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('James', 'Wilson', 'Venture Partner', 'Investments', 'Advisor', 'james.wilson@advisors.co', '+1-510-555-0203', FALSE, 'Advisor to TechVentures, helps evaluate portfolio companies', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_james_id, FALSE, NOW());
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_techventures_id, FALSE, NOW());
    END IF;

    -- James Wilson is on InnovateLab's advisory board
    IF acc_james_id IS NOT NULL AND acc_innovatelab_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('James', 'Wilson', 'Board Advisor', 'Advisory Board', 'Board Member', 'james.wilson@advisors.co', '+1-510-555-0203', FALSE, 'Member of InnovateLab advisory board since 2023', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_james_id, FALSE, NOW());
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_innovatelab_id, FALSE, NOW());
    END IF;

    -- Alex Thompson consults for CloudScale
    IF acc_alex_id IS NOT NULL AND acc_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Alex', 'Thompson', 'Strategy Consultant', 'Executive', 'Consultant', 'alex.thompson@gmail.com', '+1-415-555-0201', FALSE, 'External consultant for cloud strategy', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_alex_id, FALSE, NOW());
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_cloudscale_id, FALSE, NOW());
    END IF;

    -- Maria Garcia does design work for DataFlow
    IF acc_maria_id IS NOT NULL AND acc_dataflow_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Maria', 'Garcia', 'Contract Designer', 'Product', 'Contractor', 'maria.garcia@designstudio.com', '+1-650-555-0202', FALSE, 'Contracted for dashboard redesign project Q1 2025', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_maria_id, FALSE, NOW());
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_dataflow_id, FALSE, NOW());
    END IF;

    -- Shared IT consultant who works with multiple companies
    IF acc_cloudscale_id IS NOT NULL AND acc_securenet_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Ryan', 'Foster', 'IT Consultant', 'External', 'Consultant', 'ryan.foster@itconsulting.com', '+1-408-555-0400', FALSE, 'Provides IT consulting to multiple companies in our network', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_cloudscale_id, FALSE, NOW());
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_securenet_id, FALSE, NOW());
    END IF;

    -- PR agency contact who represents both DataFlow and CloudScale
    IF acc_cloudscale_id IS NOT NULL AND acc_dataflow_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_by, updated_by, created_at, updated_at)
        VALUES ('Jennifer', 'Blake', 'Account Director', 'PR', 'Agency Contact', 'jennifer.b@techpr.com', '+1-415-555-0401', FALSE, 'PR agency contact, handles both CloudScale and DataFlow accounts', v_user_id, v_user_id, NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_cloudscale_id, FALSE, NOW());
        INSERT INTO "Contact_Account" (contact_id, account_id, is_primary, created_at) VALUES (contact_id, acc_dataflow_id, FALSE, NOW());
    END IF;

    RAISE NOTICE 'Multi-account contacts created successfully';
END $$;

-- ==============================
-- Return counts for seeder verification
-- ==============================
SELECT
  (SELECT COUNT(*) FROM "Contact") as total_contacts,
  (SELECT COUNT(*) FROM "Contact_Account") as account_contact_links,
  (SELECT COUNT(DISTINCT contact_id) FROM "Contact_Account" ca
   WHERE EXISTS (SELECT 1 FROM "Contact_Account" ca2 WHERE ca2.contact_id = ca.contact_id AND ca2.account_id != ca.account_id)) as multi_account_contacts;

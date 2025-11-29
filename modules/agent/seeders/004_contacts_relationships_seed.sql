-- Seeder: Additional Contacts and Relationships for Organizations and Individuals
-- This seeder adds more contacts to organizations and individuals, and creates
-- cross-entity relationships (Org-to-Individual, Individual-to-Org)
--
-- IMPORTANT: This runs AFTER 001_initial_seed.sql

-- ==============================
-- STEP 1: Add Contacts to Individuals
-- Individuals can have contacts (e.g., assistants, referrers, collaborators)
-- ==============================
DO $$
DECLARE
    ind_alex_id BIGINT;
    ind_maria_id BIGINT;
    ind_james_id BIGINT;
    contact_id BIGINT;
BEGIN
    -- Get Individual IDs
    SELECT id INTO ind_alex_id FROM "Individual" WHERE LOWER(email) = LOWER('alex.thompson@gmail.com') LIMIT 1;
    SELECT id INTO ind_maria_id FROM "Individual" WHERE LOWER(email) = LOWER('maria.garcia@designstudio.com') LIMIT 1;
    SELECT id INTO ind_james_id FROM "Individual" WHERE LOWER(email) = LOWER('james.wilson@advisors.co') LIMIT 1;

    -- Alex Thompson's contacts
    IF ind_alex_id IS NOT NULL THEN
        -- His assistant
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Lisa', 'Park', 'Executive Assistant', NULL, 'Assistant', 'lisa.park@alexconsulting.com', '+1-415-555-0301', TRUE, 'Primary contact for scheduling meetings with Alex', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_alex_id, NOW());

        -- A referral partner
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Tom', 'Mitchell', 'Business Development', NULL, 'Referral Partner', 'tom.m@referralnetwork.com', '+1-415-555-0302', FALSE, 'Refers consulting clients to Alex', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_alex_id, NOW());
    END IF;

    -- Maria Garcia's contacts
    IF ind_maria_id IS NOT NULL THEN
        -- Her agent
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Rachel', 'Adams', 'Talent Agent', NULL, 'Agent', 'rachel@creativetalent.com', '+1-650-555-0303', TRUE, 'Handles contract negotiations for Maria', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_maria_id, NOW());
    END IF;

    -- James Wilson's contacts
    IF ind_james_id IS NOT NULL THEN
        -- His co-advisor
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Patricia', 'Lee', 'Startup Advisor', NULL, 'Co-Advisor', 'patricia@advisors.co', '+1-510-555-0304', FALSE, 'Co-advises with James on fintech startups', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_james_id, NOW());

        -- His administrative contact
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Kevin', 'Nguyen', 'Office Manager', NULL, 'Administrative', 'kevin@advisors.co', '+1-510-555-0305', TRUE, 'Manages James office schedule and logistics', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_james_id, NOW());
    END IF;

    RAISE NOTICE 'Individual contacts created successfully';
END $$;

-- ==============================
-- STEP 2: Add More Contacts to Organizations
-- ==============================
DO $$
DECLARE
    org_techventures_id BIGINT;
    org_cloudscale_id BIGINT;
    org_dataflow_id BIGINT;
    org_securenet_id BIGINT;
    org_innovatelab_id BIGINT;
    contact_id BIGINT;
BEGIN
    -- Get Organization IDs
    SELECT id INTO org_techventures_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT id INTO org_cloudscale_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT id INTO org_dataflow_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT id INTO org_securenet_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;
    SELECT id INTO org_innovatelab_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;

    -- Additional TechVentures contacts
    IF org_techventures_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Brian', 'Cooper', 'Managing Director', 'Investments', 'Executive', 'brian.c@techventures.example.com', '+1-415-555-0110', FALSE, 'Oversees all portfolio companies', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_techventures_id, FALSE, NOW());

        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Nina', 'Patel', 'Operations Manager', 'Operations', 'Operations', 'nina.p@techventures.example.com', '+1-415-555-0111', FALSE, 'Handles logistics and event planning', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_techventures_id, FALSE, NOW());
    END IF;

    -- Additional DataFlow contacts
    IF org_dataflow_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Marcus', 'Chen', 'Lead Data Scientist', 'Engineering', 'Technical Contact', 'marcus.c@dataflow.example.com', '+1-650-555-0210', FALSE, 'Technical expert for data pipeline questions', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_dataflow_id, FALSE, NOW());

        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Stephanie', 'Wright', 'Sales Director', 'Sales', 'Sales Contact', 'stephanie.w@dataflow.example.com', '+1-650-555-0211', FALSE, 'For partnership and sales discussions', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_dataflow_id, FALSE, NOW());
    END IF;

    -- Additional SecureNet contacts
    IF org_securenet_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Andrew', 'Martinez', 'Head of Sales', 'Sales', 'Decision Maker', 'andrew.m@securenet.example.com', '+1-408-555-0210', FALSE, 'Handles enterprise sales', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_securenet_id, FALSE, NOW());
    END IF;

    -- Additional InnovateLab contacts
    IF org_innovatelab_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Diana', 'Ross', 'Senior Consultant', 'Consulting', 'Technical Contact', 'diana.r@innovatelab.example.com', '+1-510-555-0210', FALSE, 'Specialist in product-market fit', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_innovatelab_id, FALSE, NOW());

        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Chris', 'Anderson', 'Founder & CEO', 'Executive', 'Decision Maker', 'chris.a@innovatelab.example.com', '+1-510-555-0211', FALSE, 'Final decision maker for large engagements', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_innovatelab_id, FALSE, NOW());
    END IF;

    RAISE NOTICE 'Additional organization contacts created successfully';
END $$;

-- ==============================
-- STEP 3: Create Organization-Individual Relationships
-- Link Individuals to Organizations (e.g., advisors, consultants, board members)
-- ==============================
DO $$
DECLARE
    org_techventures_id BIGINT;
    org_cloudscale_id BIGINT;
    org_dataflow_id BIGINT;
    org_innovatelab_id BIGINT;
    ind_alex_id BIGINT;
    ind_james_id BIGINT;
    ind_maria_id BIGINT;
    contact_id BIGINT;
BEGIN
    -- Get Organization IDs
    SELECT id INTO org_techventures_id FROM "Organization" WHERE LOWER(name) = LOWER('TechVentures Inc') LIMIT 1;
    SELECT id INTO org_cloudscale_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT id INTO org_dataflow_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT id INTO org_innovatelab_id FROM "Organization" WHERE LOWER(name) = LOWER('InnovateLab') LIMIT 1;

    -- Get Individual IDs
    SELECT id INTO ind_alex_id FROM "Individual" WHERE LOWER(email) = LOWER('alex.thompson@gmail.com') LIMIT 1;
    SELECT id INTO ind_james_id FROM "Individual" WHERE LOWER(email) = LOWER('james.wilson@advisors.co') LIMIT 1;
    SELECT id INTO ind_maria_id FROM "Individual" WHERE LOWER(email) = LOWER('maria.garcia@designstudio.com') LIMIT 1;

    -- James Wilson advises TechVentures (as a Contact linked to both)
    IF ind_james_id IS NOT NULL AND org_techventures_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('James', 'Wilson', 'Venture Partner', 'Investments', 'Advisor', 'james.wilson@advisors.co', '+1-510-555-0203', FALSE, 'Advisor to TechVentures, helps evaluate portfolio companies', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_james_id, NOW());
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_techventures_id, FALSE, NOW());
    END IF;

    -- Alex Thompson consults for CloudScale (as a Contact linked to both)
    IF ind_alex_id IS NOT NULL AND org_cloudscale_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Alex', 'Thompson', 'Strategy Consultant', 'Executive', 'Consultant', 'alex.thompson@gmail.com', '+1-415-555-0201', FALSE, 'External consultant for cloud strategy', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_alex_id, NOW());
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_cloudscale_id, FALSE, NOW());
    END IF;

    -- James Wilson is on InnovateLab's advisory board
    IF ind_james_id IS NOT NULL AND org_innovatelab_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('James', 'Wilson', 'Board Advisor', 'Advisory Board', 'Board Member', 'james.wilson@advisors.co', '+1-510-555-0203', FALSE, 'Member of InnovateLab advisory board since 2023', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_james_id, NOW());
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_innovatelab_id, FALSE, NOW());
    END IF;

    -- Maria Garcia does design work for DataFlow
    IF ind_maria_id IS NOT NULL AND org_dataflow_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Maria', 'Garcia', 'Contract Designer', 'Product', 'Contractor', 'maria.garcia@designstudio.com', '+1-650-555-0202', FALSE, 'Contracted for dashboard redesign project Q1 2025', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactIndividual" (contact_id, individual_id, created_at) VALUES (contact_id, ind_maria_id, NOW());
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_dataflow_id, FALSE, NOW());
    END IF;

    RAISE NOTICE 'Organization-Individual relationships created successfully';
END $$;

-- ==============================
-- STEP 4: Create Shared Contacts (Multi-Org)
-- Some contacts work across multiple organizations
-- ==============================
DO $$
DECLARE
    org_cloudscale_id BIGINT;
    org_dataflow_id BIGINT;
    org_securenet_id BIGINT;
    contact_id BIGINT;
BEGIN
    -- Get Organization IDs
    SELECT id INTO org_cloudscale_id FROM "Organization" WHERE LOWER(name) = LOWER('CloudScale Systems') LIMIT 1;
    SELECT id INTO org_dataflow_id FROM "Organization" WHERE LOWER(name) = LOWER('DataFlow Analytics') LIMIT 1;
    SELECT id INTO org_securenet_id FROM "Organization" WHERE LOWER(name) = LOWER('SecureNet Solutions') LIMIT 1;

    -- Shared IT consultant who works with multiple companies
    IF org_cloudscale_id IS NOT NULL AND org_securenet_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Ryan', 'Foster', 'IT Consultant', 'External', 'Consultant', 'ryan.foster@itconsulting.com', '+1-408-555-0400', FALSE, 'Provides IT consulting to multiple companies in our network', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_cloudscale_id, FALSE, NOW());
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_securenet_id, FALSE, NOW());
    END IF;

    -- PR agency contact who represents both DataFlow and CloudScale
    IF org_cloudscale_id IS NOT NULL AND org_dataflow_id IS NOT NULL THEN
        INSERT INTO "Contact" (first_name, last_name, title, department, role, email, phone, is_primary, notes, created_at, updated_at)
        VALUES ('Jennifer', 'Blake', 'Account Director', 'PR', 'Agency Contact', 'jennifer.b@techpr.com', '+1-415-555-0401', FALSE, 'PR agency contact, handles both CloudScale and DataFlow accounts', NOW(), NOW())
        RETURNING id INTO contact_id;
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_cloudscale_id, FALSE, NOW());
        INSERT INTO "ContactOrganization" (contact_id, organization_id, is_primary, created_at) VALUES (contact_id, org_dataflow_id, FALSE, NOW());
    END IF;

    RAISE NOTICE 'Shared multi-org contacts created successfully';
END $$;

-- ==============================
-- Return counts for seeder verification
-- ==============================
SELECT
  (SELECT COUNT(*) FROM "Contact") as total_contacts,
  (SELECT COUNT(*) FROM "ContactOrganization") as org_contact_links,
  (SELECT COUNT(*) FROM "ContactIndividual") as individual_contact_links,
  (SELECT COUNT(DISTINCT contact_id) FROM "ContactOrganization" co
   WHERE EXISTS (SELECT 1 FROM "ContactOrganization" co2 WHERE co2.contact_id = co.contact_id AND co2.organization_id != co.organization_id)) as multi_org_contacts;

-- Seeder: Sample notes for testing report builder
-- Creates notes associated with various entity types

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id BIGINT;
    v_org_id BIGINT;
    v_ind_id BIGINT;
    v_contact_id BIGINT;
    v_lead_id BIGINT;
    v_job_lead_id BIGINT;
BEGIN
    -- Get the default tenant
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping notes seed';
        RETURN;
    END IF;

    -- Get bootstrap user for audit columns
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    -- Get sample entities to attach notes to
    SELECT id INTO v_org_id FROM "Organization" LIMIT 1;
    SELECT id INTO v_ind_id FROM "Individual" LIMIT 1;
    SELECT id INTO v_contact_id FROM "Contact" LIMIT 1;
    SELECT id INTO v_lead_id FROM "Lead" WHERE type = 'Opportunity' LIMIT 1;
    SELECT id INTO v_job_lead_id FROM "Lead" WHERE type = 'Job' LIMIT 1;

    -- Insert notes for Organization
    IF v_org_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, updated_by, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Organization', v_org_id, 'Initial meeting went well. They are interested in our enterprise solution.', v_user_id, v_user_id, NOW(), NOW()),
            (v_tenant_id, 'Organization', v_org_id, 'Follow-up call scheduled for next week. Need to prepare pricing proposal.', v_user_id, v_user_id, NOW(), NOW()),
            (v_tenant_id, 'Organization', v_org_id, 'Decision maker is the VP of Engineering. Budget approved for Q1.', v_user_id, v_user_id, NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Organization ID: %', v_org_id;
    END IF;

    -- Insert notes for Individual
    IF v_ind_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, updated_by, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Individual', v_ind_id, 'Met at industry conference. Very knowledgeable about market trends.', v_user_id, v_user_id, NOW(), NOW()),
            (v_tenant_id, 'Individual', v_ind_id, 'Prefers email communication. Best time to reach is mornings.', v_user_id, v_user_id, NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Individual ID: %', v_ind_id;
    END IF;

    -- Insert notes for Contact
    IF v_contact_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, updated_by, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Contact', v_contact_id, 'Primary point of contact for technical discussions.', v_user_id, v_user_id, NOW(), NOW()),
            (v_tenant_id, 'Contact', v_contact_id, 'Has authority to sign contracts up to $50K.', v_user_id, v_user_id, NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Contact ID: %', v_contact_id;
    END IF;

    -- Insert notes for Lead (Opportunity)
    IF v_lead_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, updated_by, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Lead', v_lead_id, 'High priority opportunity. Competitor is also pitching.', v_user_id, v_user_id, NOW(), NOW()),
            (v_tenant_id, 'Lead', v_lead_id, 'Timeline: Decision expected by end of month.', v_user_id, v_user_id, NOW(), NOW()),
            (v_tenant_id, 'Lead', v_lead_id, 'Technical requirements reviewed. Our solution is a good fit.', v_user_id, v_user_id, NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Opportunity Lead ID: %', v_lead_id;
    END IF;

    -- Insert notes for Lead (Job)
    IF v_job_lead_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, updated_by, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Lead', v_job_lead_id, 'Great company culture. Remote-friendly position.', v_user_id, v_user_id, NOW(), NOW()),
            (v_tenant_id, 'Lead', v_job_lead_id, 'Technical interview scheduled for next Tuesday.', v_user_id, v_user_id, NOW(), NOW()),
            (v_tenant_id, 'Lead', v_job_lead_id, 'Salary range confirmed: $150K-$180K + equity.', v_user_id, v_user_id, NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Job Lead ID: %', v_job_lead_id;
    END IF;

    RAISE NOTICE 'Notes seed completed for tenant: %', v_tenant_id;
END $$;

-- Return counts for seeder verification
SELECT
  (SELECT COUNT(*) FROM "Note") as total_notes,
  (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Organization') as organization_notes,
  (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Individual') as individual_notes,
  (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Contact') as contact_notes,
  (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Lead') as lead_notes;

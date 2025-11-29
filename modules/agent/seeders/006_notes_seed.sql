-- Seeder: Sample notes for testing report builder
-- Creates notes associated with various entity types

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_org_id BIGINT;
    v_ind_id BIGINT;
    v_contact_id BIGINT;
    v_lead_id BIGINT;
BEGIN
    -- Get the default tenant
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping notes seed';
        RETURN;
    END IF;

    -- Get sample entities to attach notes to
    SELECT id INTO v_org_id FROM "Organization" LIMIT 1;
    SELECT id INTO v_ind_id FROM "Individual" LIMIT 1;
    SELECT id INTO v_contact_id FROM "Contact" LIMIT 1;
    SELECT id INTO v_lead_id FROM "Lead" LIMIT 1;

    -- Insert notes for Organization
    IF v_org_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Organization', v_org_id, 'Initial meeting went well. They are interested in our enterprise solution.', NOW(), NOW()),
            (v_tenant_id, 'Organization', v_org_id, 'Follow-up call scheduled for next week. Need to prepare pricing proposal.', NOW(), NOW()),
            (v_tenant_id, 'Organization', v_org_id, 'Decision maker is the VP of Engineering. Budget approved for Q1.', NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Organization ID: %', v_org_id;
    END IF;

    -- Insert notes for Individual
    IF v_ind_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Individual', v_ind_id, 'Met at industry conference. Very knowledgeable about market trends.', NOW(), NOW()),
            (v_tenant_id, 'Individual', v_ind_id, 'Prefers email communication. Best time to reach is mornings.', NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Individual ID: %', v_ind_id;
    END IF;

    -- Insert notes for Contact
    IF v_contact_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Contact', v_contact_id, 'Primary point of contact for technical discussions.', NOW(), NOW()),
            (v_tenant_id, 'Contact', v_contact_id, 'Has authority to sign contracts up to $50K.', NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Contact ID: %', v_contact_id;
    END IF;

    -- Insert notes for Lead
    IF v_lead_id IS NOT NULL THEN
        INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_at, updated_at)
        VALUES
            (v_tenant_id, 'Lead', v_lead_id, 'High priority opportunity. Competitor is also pitching.', NOW(), NOW()),
            (v_tenant_id, 'Lead', v_lead_id, 'Timeline: Decision expected by end of month.', NOW(), NOW()),
            (v_tenant_id, 'Lead', v_lead_id, 'Technical requirements reviewed. Our solution is a good fit.', NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added notes for Lead ID: %', v_lead_id;
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

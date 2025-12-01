-- ==============================
-- Document Seeds
-- ==============================
-- Seeds sample documents with tags and entity links
-- Note: Actual PDF files must be uploaded to storage separately
-- Run: docker compose exec agent python /app/scripts/seed_documents.py
-- ==============================

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id BIGINT;
    v_doc_id BIGINT;
    v_tag_id BIGINT;
    v_entity_id BIGINT;
BEGIN
    -- Get first active tenant
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';
    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'PinaColada tenant not found, skipping document seeds';
        RETURN;
    END IF;

    -- Get first user for this tenant (William)
    SELECT id INTO v_user_id FROM "User" WHERE tenant_id = v_tenant_id ORDER BY id LIMIT 1;
    IF v_user_id IS NULL THEN
        RAISE NOTICE 'No user found for tenant, skipping document seeds';
        RETURN;
    END IF;

    -- ==============================
    -- Create Tags
    -- ==============================
    INSERT INTO "Tag" (name) VALUES
        ('proposal'), ('sales'), ('2025'), ('meeting'), ('notes'), ('internal'),
        ('contract'), ('legal'), ('draft'), ('spec'), ('technical'), ('product'),
        ('invoice'), ('finance'), ('template'), ('resume'), ('candidate'), ('hiring')
    ON CONFLICT (name) DO NOTHING;

    -- ==============================
    -- Document 1: Company Proposal -> Acme Corp
    -- ==============================
    SELECT o.id INTO v_entity_id FROM "Organization" o
    JOIN "Account" a ON o.account_id = a.id
    WHERE LOWER(o.name) = LOWER('Acme Corp') AND a.tenant_id = v_tenant_id;

    INSERT INTO "Asset" (asset_type, tenant_id, user_id, filename, content_type, description)
    VALUES ('document', v_tenant_id, v_user_id, 'company_proposal.pdf', 'application/pdf',
            'Q1 2025 company proposal for Acme Corp partnership')
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_doc_id;

    IF v_doc_id IS NOT NULL THEN
        INSERT INTO "Document" (id, storage_path, file_size)
        VALUES (v_doc_id, v_tenant_id || '/seed/company_proposal.pdf', 742);

        IF v_entity_id IS NOT NULL THEN
            INSERT INTO "Entity_Asset" (asset_id, entity_type, entity_id)
            VALUES (v_doc_id, 'Organization', v_entity_id)
            ON CONFLICT DO NOTHING;
            RAISE NOTICE 'Linked company_proposal.pdf to Acme Corp';
        END IF;

        FOR v_tag_id IN SELECT id FROM "Tag" WHERE name IN ('proposal', 'sales', '2025') LOOP
            INSERT INTO "Entity_Tag" (tag_id, entity_type, entity_id)
            VALUES (v_tag_id, 'Asset', v_doc_id)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;

    -- ==============================
    -- Document 2: Meeting Notes -> Job Search 2025 Project
    -- ==============================
    SELECT id INTO v_entity_id FROM "Project" WHERE name = 'Job Search 2025' AND tenant_id = v_tenant_id;

    INSERT INTO "Asset" (asset_type, tenant_id, user_id, filename, content_type, description)
    VALUES ('document', v_tenant_id, v_user_id, 'meeting_notes.pdf', 'application/pdf',
            'Product roadmap meeting notes from January kickoff')
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_doc_id;

    IF v_doc_id IS NOT NULL THEN
        INSERT INTO "Document" (id, storage_path, file_size)
        VALUES (v_doc_id, v_tenant_id || '/seed/meeting_notes.pdf', 807);

        IF v_entity_id IS NOT NULL THEN
            INSERT INTO "Entity_Asset" (asset_id, entity_type, entity_id)
            VALUES (v_doc_id, 'Project', v_entity_id)
            ON CONFLICT DO NOTHING;
            RAISE NOTICE 'Linked meeting_notes.pdf to Job Search 2025';
        END IF;

        FOR v_tag_id IN SELECT id FROM "Tag" WHERE name IN ('meeting', 'notes', 'internal') LOOP
            INSERT INTO "Entity_Tag" (tag_id, entity_type, entity_id)
            VALUES (v_tag_id, 'Asset', v_doc_id)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;

    -- ==============================
    -- Document 3: Contract Draft -> CloudScale Systems
    -- ==============================
    SELECT o.id INTO v_entity_id FROM "Organization" o
    JOIN "Account" a ON o.account_id = a.id
    WHERE LOWER(o.name) = LOWER('CloudScale Systems') AND a.tenant_id = v_tenant_id;

    INSERT INTO "Asset" (asset_type, tenant_id, user_id, filename, content_type, description)
    VALUES ('document', v_tenant_id, v_user_id, 'contract_draft.pdf', 'application/pdf',
            'Draft service agreement pending legal review')
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_doc_id;

    IF v_doc_id IS NOT NULL THEN
        INSERT INTO "Document" (id, storage_path, file_size)
        VALUES (v_doc_id, v_tenant_id || '/seed/contract_draft.pdf', 828);

        IF v_entity_id IS NOT NULL THEN
            INSERT INTO "Entity_Asset" (asset_id, entity_type, entity_id)
            VALUES (v_doc_id, 'Organization', v_entity_id)
            ON CONFLICT DO NOTHING;
            RAISE NOTICE 'Linked contract_draft.pdf to CloudScale Systems';
        END IF;

        FOR v_tag_id IN SELECT id FROM "Tag" WHERE name IN ('contract', 'legal', 'draft') LOOP
            INSERT INTO "Entity_Tag" (tag_id, entity_type, entity_id)
            VALUES (v_tag_id, 'Asset', v_doc_id)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;

    -- ==============================
    -- Document 4: Product Spec -> Consulting Pipeline Project
    -- ==============================
    SELECT id INTO v_entity_id FROM "Project" WHERE name = 'Consulting Pipeline' AND tenant_id = v_tenant_id;

    INSERT INTO "Asset" (asset_type, tenant_id, user_id, filename, content_type, description)
    VALUES ('document', v_tenant_id, v_user_id, 'product_spec.pdf', 'application/pdf',
            'Technical specification for CRM platform v2.0')
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_doc_id;

    IF v_doc_id IS NOT NULL THEN
        INSERT INTO "Document" (id, storage_path, file_size)
        VALUES (v_doc_id, v_tenant_id || '/seed/product_spec.pdf', 834);

        IF v_entity_id IS NOT NULL THEN
            INSERT INTO "Entity_Asset" (asset_id, entity_type, entity_id)
            VALUES (v_doc_id, 'Project', v_entity_id)
            ON CONFLICT DO NOTHING;
            RAISE NOTICE 'Linked product_spec.pdf to Consulting Pipeline';
        END IF;

        FOR v_tag_id IN SELECT id FROM "Tag" WHERE name IN ('spec', 'technical', 'product') LOOP
            INSERT INTO "Entity_Tag" (tag_id, entity_type, entity_id)
            VALUES (v_tag_id, 'Asset', v_doc_id)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;

    -- ==============================
    -- Document 5: Invoice Sample -> TechVentures Inc
    -- ==============================
    SELECT o.id INTO v_entity_id FROM "Organization" o
    JOIN "Account" a ON o.account_id = a.id
    WHERE LOWER(o.name) = LOWER('TechVentures Inc') AND a.tenant_id = v_tenant_id;

    INSERT INTO "Asset" (asset_type, tenant_id, user_id, filename, content_type, description)
    VALUES ('document', v_tenant_id, v_user_id, 'invoice_sample.pdf', 'application/pdf',
            'Sample invoice for billing reference')
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_doc_id;

    IF v_doc_id IS NOT NULL THEN
        INSERT INTO "Document" (id, storage_path, file_size)
        VALUES (v_doc_id, v_tenant_id || '/seed/invoice_sample.pdf', 784);

        IF v_entity_id IS NOT NULL THEN
            INSERT INTO "Entity_Asset" (asset_id, entity_type, entity_id)
            VALUES (v_doc_id, 'Organization', v_entity_id)
            ON CONFLICT DO NOTHING;
            RAISE NOTICE 'Linked invoice_sample.pdf to TechVentures Inc';
        END IF;

        FOR v_tag_id IN SELECT id FROM "Tag" WHERE name IN ('invoice', 'finance', 'template') LOOP
            INSERT INTO "Entity_Tag" (tag_id, entity_type, entity_id)
            VALUES (v_tag_id, 'Asset', v_doc_id)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;

    -- ==============================
    -- Document 6: William Hubenschmidt Resume
    -- ==============================
    SELECT i.id INTO v_entity_id FROM "Individual" i
    JOIN "Account" a ON i.account_id = a.id
    WHERE a.tenant_id = v_tenant_id
      AND LOWER(i.first_name) = 'william'
      AND LOWER(i.last_name) = 'hubenschmidt';

    INSERT INTO "Asset" (asset_type, tenant_id, user_id, filename, content_type, description)
    VALUES ('document', v_tenant_id, v_user_id, 'individual_resume.pdf', 'application/pdf',
            'Resume for William Hubenschmidt - Senior Software Engineer')
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_doc_id;

    IF v_doc_id IS NOT NULL THEN
        INSERT INTO "Document" (id, storage_path, file_size)
        VALUES (v_doc_id, v_tenant_id || '/seed/individual_resume.pdf', 950);

        IF v_entity_id IS NOT NULL THEN
            INSERT INTO "Entity_Asset" (asset_id, entity_type, entity_id)
            VALUES (v_doc_id, 'Individual', v_entity_id)
            ON CONFLICT DO NOTHING;
            RAISE NOTICE 'Linked individual_resume.pdf to William Hubenschmidt (ID: %)', v_entity_id;
        ELSE
            RAISE NOTICE 'William Hubenschmidt not found, resume not linked';
        END IF;

        FOR v_tag_id IN SELECT id FROM "Tag" WHERE name IN ('resume', 'candidate', 'hiring') LOOP
            INSERT INTO "Entity_Tag" (tag_id, entity_type, entity_id)
            VALUES (v_tag_id, 'Asset', v_doc_id)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;

    RAISE NOTICE 'Document seeding complete for tenant %', v_tenant_id;
END $$;

-- Summary
SELECT
    (SELECT COUNT(*) FROM "Asset" WHERE asset_type = 'document') as documents,
    (SELECT COUNT(*) FROM "Entity_Asset") as entity_links,
    (SELECT COUNT(*) FROM "Entity_Tag" WHERE entity_type = 'Asset') as tag_links;

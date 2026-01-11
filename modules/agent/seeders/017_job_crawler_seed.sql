-- ==============================
-- Automation Config Seed
-- ==============================
-- Creates a test job crawler for William Hubenschmidt
-- ==============================

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id BIGINT;
    v_individual_id BIGINT;
    v_doc_id BIGINT;
BEGIN
    -- Get PinaColada tenant
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';
    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'PinaColada tenant not found, skipping automation seed';
        RETURN;
    END IF;

    -- Get William's user
    SELECT id INTO v_user_id FROM "User"
    WHERE tenant_id = v_tenant_id AND email = 'whubenschmidt@gmail.com';
    IF v_user_id IS NULL THEN
        RAISE NOTICE 'William user not found, skipping automation seed';
        RETURN;
    END IF;

    -- Get William's Individual ID
    SELECT i.id INTO v_individual_id FROM "Individual" i
    JOIN "Account" a ON i.account_id = a.id
    WHERE a.tenant_id = v_tenant_id
      AND LOWER(i.first_name) = 'william'
      AND LOWER(i.last_name) = 'hubenschmidt';

    -- Get William's resume document ID
    SELECT a.id INTO v_doc_id FROM "Asset" a
    WHERE a.filename = 'william_hubenschmidt_resume.pdf'
      AND a.tenant_id = v_tenant_id;

    -- Create test job crawler
    INSERT INTO "Automation_Config" (
        tenant_id,
        user_id,
        name,
        entity_type,
        enabled,
        interval_seconds,
        prospects_per_run,
        compilation_target,
        search_query,
        use_suggested_query,
        location,
        ats_mode,
        time_filter,
        target_type,
        target_ids,
        source_document_ids,
        use_agent,
        agent_model,
        system_prompt,
        use_suggested_prompt,
        digest_enabled,
        digest_emails,
        digest_time,
        digest_model,
        created_at,
        updated_at
    )
    VALUES (
        v_tenant_id,
        v_user_id,
        'test job crawler',
        'job',
        false,
        10,
        10,
        40,
        '(senior OR staff) software engineer -junior -intern',
        true,
        'New York',
        true,
        'week',
        'individual',
        CASE WHEN v_individual_id IS NOT NULL THEN jsonb_build_array(v_individual_id) ELSE '[]'::jsonb END,
        CASE WHEN v_doc_id IS NOT NULL THEN jsonb_build_array(v_doc_id) ELSE '[]'::jsonb END,
        true,
        'gpt-5.1',
        'Please job search for matches using the resume linked on his Individual record. Jobs should be located in New York and at startups (seed, series A, B, and C).',
        true,
        true,
        'whubenschmidt@gmail.com',
        '09:00',
        'claude-sonnet-4-5-20250929',
        NOW(),
        NOW()
    )
    ON CONFLICT (tenant_id, user_id, name) DO UPDATE SET
        interval_seconds = EXCLUDED.interval_seconds,
        prospects_per_run = EXCLUDED.prospects_per_run,
        compilation_target = EXCLUDED.compilation_target,
        search_query = EXCLUDED.search_query,
        use_suggested_query = EXCLUDED.use_suggested_query,
        location = EXCLUDED.location,
        ats_mode = EXCLUDED.ats_mode,
        time_filter = EXCLUDED.time_filter,
        target_type = EXCLUDED.target_type,
        target_ids = EXCLUDED.target_ids,
        source_document_ids = EXCLUDED.source_document_ids,
        use_agent = EXCLUDED.use_agent,
        agent_model = EXCLUDED.agent_model,
        system_prompt = EXCLUDED.system_prompt,
        use_suggested_prompt = EXCLUDED.use_suggested_prompt,
        digest_enabled = EXCLUDED.digest_enabled,
        digest_emails = EXCLUDED.digest_emails,
        digest_time = EXCLUDED.digest_time,
        digest_model = EXCLUDED.digest_model,
        updated_at = NOW();

    RAISE NOTICE 'Test job crawler created for William (tenant_id=%, user_id=%, individual_id=%, doc_id=%)',
        v_tenant_id, v_user_id, v_individual_id, v_doc_id;
END $$;

-- Verification
SELECT
    id,
    name,
    entity_type,
    enabled,
    interval_seconds,
    prospects_per_run,
    compilation_target,
    search_query,
    use_agent,
    agent_model
FROM "Automation_Config"
WHERE name = 'test job crawler';

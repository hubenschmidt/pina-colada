-- Set all users' selected_project_id to "Job Search 2025" project

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_project_id BIGINT;
    v_count INT;
BEGIN
    -- Get tenant
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;
    RAISE NOTICE 'Found tenant_id: %', v_tenant_id;

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'Tenant not found, skipping user selected project seed';
        RETURN;
    END IF;

    -- Get Job Search 2025 project
    SELECT id INTO v_project_id FROM "Project" WHERE name = 'Job Search 2025' AND tenant_id = v_tenant_id LIMIT 1;
    RAISE NOTICE 'Found project_id: %', v_project_id;

    IF v_project_id IS NULL THEN
        RAISE NOTICE 'Job Search 2025 project not found, skipping user selected project seed';
        RETURN;
    END IF;

    -- Check how many users exist
    SELECT COUNT(*) INTO v_count FROM "User" WHERE tenant_id = v_tenant_id;
    RAISE NOTICE 'Found % users in tenant', v_count;

    -- Update all users in tenant to have this as their selected project
    UPDATE "User"
    SET selected_project_id = v_project_id
    WHERE tenant_id = v_tenant_id;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Updated % users with selected_project_id = % (Job Search 2025)', v_count, v_project_id;
END $$;

-- Return count for seeder runner
SELECT COUNT(*) AS users_with_selected_project FROM "User" WHERE selected_project_id IS NOT NULL;

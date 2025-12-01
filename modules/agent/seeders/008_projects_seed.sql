-- Seeder 008: Projects sample data
-- Creates sample projects and links all leads to appropriate projects

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_job_search_id BIGINT;
    v_consulting_id BIGINT;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping project seeder';
        RETURN;
    END IF;

    -- Fix any leads without tenant_id
    UPDATE "Lead" l
    SET tenant_id = v_tenant_id
    WHERE l.tenant_id IS NULL;

    -- Create Job Search 2025 Project
    INSERT INTO "Project" (tenant_id, name, description, status, start_date, created_at, updated_at)
    VALUES (
        v_tenant_id,
        'Job Search 2025',
        'Active job search campaign for 2025',
        'Active',
        NOW(),
        NOW(),
        NOW()
    )
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_job_search_id;

    IF v_job_search_id IS NULL THEN
        SELECT id INTO v_job_search_id FROM "Project" WHERE tenant_id = v_tenant_id AND name = 'Job Search 2025';
    END IF;

    -- Create Consulting Pipeline Project
    INSERT INTO "Project" (tenant_id, name, description, status, start_date, created_at, updated_at)
    VALUES (
        v_tenant_id,
        'Consulting Pipeline',
        'Consulting opportunities and partnerships',
        'Active',
        NOW(),
        NOW(),
        NOW()
    )
    ON CONFLICT DO NOTHING
    RETURNING id INTO v_consulting_id;

    IF v_consulting_id IS NULL THEN
        SELECT id INTO v_consulting_id FROM "Project" WHERE tenant_id = v_tenant_id AND name = 'Consulting Pipeline';
    END IF;

    RAISE NOTICE 'Job Search Project ID: %, Consulting Project ID: %', v_job_search_id, v_consulting_id;

    -- Link existing Deals to Job Search project
    UPDATE "Deal"
    SET project_id = v_job_search_id
    WHERE tenant_id = v_tenant_id
      AND project_id IS NULL;

    -- Link Job-type leads to Job Search 2025
    INSERT INTO "Lead_Project" (lead_id, project_id, created_at)
    SELECT l.id, v_job_search_id, NOW()
    FROM "Lead" l
    WHERE l.type = 'Job'
      AND NOT EXISTS (
          SELECT 1 FROM "Lead_Project" lp
          WHERE lp.lead_id = l.id AND lp.project_id = v_job_search_id
      );

    -- Link Opportunity and Partnership leads to Consulting Pipeline
    INSERT INTO "Lead_Project" (lead_id, project_id, created_at)
    SELECT l.id, v_consulting_id, NOW()
    FROM "Lead" l
    WHERE l.type IN ('Opportunity', 'Partnership')
      AND NOT EXISTS (
          SELECT 1 FROM "Lead_Project" lp
          WHERE lp.lead_id = l.id AND lp.project_id = v_consulting_id
      );

    RAISE NOTICE 'Projects seeded successfully';
END $$;

-- Show counts
SELECT
    (SELECT COUNT(*) FROM "Project") as projects,
    (SELECT COUNT(*) FROM "Deal" WHERE project_id IS NOT NULL) as deals_with_project,
    (SELECT COUNT(*) FROM "Lead_Project") as lead_project_links,
    (SELECT COUNT(*) FROM "Lead" l WHERE NOT EXISTS (SELECT 1 FROM "Lead_Project" lp WHERE lp.lead_id = l.id)) as leads_without_project;

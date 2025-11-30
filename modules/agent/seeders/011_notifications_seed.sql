-- Seeder 011: Comment Notifications
-- Creates notifications from existing comments (created in 010_comments_seed.sql)
-- across different project scopes for testing scope switching
-- Includes: Project 1 (Job Search 2025), Project 2 (Consulting Pipeline), and Global (no project)

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id_william BIGINT;
    v_user_id_jennifer BIGINT;
    v_project_id_job_search BIGINT;
    v_project_id_consulting BIGINT;
    v_count INTEGER := 0;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping notifications seeder';
        RETURN;
    END IF;

    -- Get user IDs
    SELECT id INTO v_user_id_william FROM "User" WHERE email = 'whubenschmidt@gmail.com' AND tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_user_id_jennifer FROM "User" WHERE email = 'jennifervlev@gmail.com' AND tenant_id = v_tenant_id LIMIT 1;

    IF v_user_id_william IS NULL OR v_user_id_jennifer IS NULL THEN
        RAISE NOTICE 'Users not found, skipping notifications seeder';
        RETURN;
    END IF;

    -- Get project IDs
    SELECT id INTO v_project_id_job_search FROM "Project" WHERE name = 'Job Search 2025' AND tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_project_id_consulting FROM "Project" WHERE name = 'Consulting Pipeline' AND tenant_id = v_tenant_id LIMIT 1;

    RAISE NOTICE 'Found projects - Job Search: %, Consulting: %', v_project_id_job_search, v_project_id_consulting;

    -- Clear existing notifications for clean seeding
    DELETE FROM "CommentNotification" WHERE tenant_id = v_tenant_id;

    -- ===========================================
    -- NOTIFICATIONS FOR JOB SEARCH 2025 PROJECT
    -- ===========================================

    -- Notify William of Jennifer's comments on Leads in Job Search 2025
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_william,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at + INTERVAL '1 minute'
    FROM "Comment" c
    JOIN "LeadProject" lp ON c.commentable_type = 'Lead' AND c.commentable_id = lp.lead_id
    WHERE c.tenant_id = v_tenant_id
      AND c.created_by = v_user_id_jennifer
      AND lp.project_id = v_project_id_job_search
    ORDER BY c.id, c.created_at DESC
    LIMIT 3;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Created % notifications for Leads in Job Search 2025', v_count;

    -- Notify William of Jennifer's comments on Deals in Job Search 2025
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_william,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at + INTERVAL '1 minute'
    FROM "Comment" c
    JOIN "Deal" d ON c.commentable_type = 'Deal' AND c.commentable_id = d.id
    WHERE c.tenant_id = v_tenant_id
      AND c.created_by = v_user_id_jennifer
      AND d.project_id = v_project_id_job_search
    ORDER BY c.id, c.created_at DESC
    LIMIT 2;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Created % notifications for Deals in Job Search 2025', v_count;

    -- ===========================================
    -- NOTIFICATIONS FOR CONSULTING PIPELINE PROJECT
    -- ===========================================

    -- Notify William of Jennifer's comments on Leads in Consulting Pipeline
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_william,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at + INTERVAL '1 minute'
    FROM "Comment" c
    JOIN "LeadProject" lp ON c.commentable_type = 'Lead' AND c.commentable_id = lp.lead_id
    WHERE c.tenant_id = v_tenant_id
      AND c.created_by = v_user_id_jennifer
      AND lp.project_id = v_project_id_consulting
    ORDER BY c.id, c.created_at DESC
    LIMIT 3;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Created % notifications for Leads in Consulting Pipeline', v_count;

    -- Notify William of Jennifer's comments on Deals in Consulting Pipeline
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_william,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at + INTERVAL '1 minute'
    FROM "Comment" c
    JOIN "Deal" d ON c.commentable_type = 'Deal' AND c.commentable_id = d.id
    WHERE c.tenant_id = v_tenant_id
      AND c.created_by = v_user_id_jennifer
      AND d.project_id = v_project_id_consulting
    ORDER BY c.id, c.created_at DESC
    LIMIT 2;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Created % notifications for Deals in Consulting Pipeline', v_count;

    -- ===========================================
    -- NOTIFICATIONS FOR GLOBAL SCOPE (no project)
    -- ===========================================

    -- Notify William of Jennifer's comments on Tasks (Tasks are global, no project association)
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_william,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at + INTERVAL '1 minute'
    FROM "Comment" c
    JOIN "Task" t ON c.commentable_type = 'Task' AND c.commentable_id = t.id
    WHERE c.tenant_id = v_tenant_id
      AND c.created_by = v_user_id_jennifer
    ORDER BY c.id, c.created_at DESC
    LIMIT 2;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Created % notifications for Tasks', v_count;

    -- Notify William of Jennifer's comments on Individuals (global)
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_william,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at + INTERVAL '1 minute'
    FROM "Comment" c
    WHERE c.tenant_id = v_tenant_id
      AND c.commentable_type = 'Individual'
      AND c.created_by = v_user_id_jennifer
    ORDER BY c.id, c.created_at DESC
    LIMIT 2;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Created % notifications for Individuals', v_count;

    -- Notify William of Jennifer's comments on Organizations (global)
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_william,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at + INTERVAL '1 minute'
    FROM "Comment" c
    WHERE c.tenant_id = v_tenant_id
      AND c.commentable_type = 'Organization'
      AND c.created_by = v_user_id_jennifer
    ORDER BY c.id, c.created_at DESC
    LIMIT 2;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Created % notifications for Organizations', v_count;

    -- ===========================================
    -- DIRECT REPLY NOTIFICATIONS
    -- ===========================================

    -- Create direct_reply notifications for any reply comments where Jennifer replied to William
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT
        v_tenant_id,
        v_user_id_william,
        reply.id,
        'direct_reply',
        FALSE,
        reply.created_at + INTERVAL '1 minute'
    FROM "Comment" reply
    JOIN "Comment" parent ON reply.parent_comment_id = parent.id
    WHERE reply.tenant_id = v_tenant_id
      AND reply.created_by = v_user_id_jennifer
      AND parent.created_by = v_user_id_william
    ON CONFLICT (user_id, comment_id) DO NOTHING;

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RAISE NOTICE 'Created % direct_reply notifications', v_count;

    -- ===========================================
    -- MARK SOME AS READ FOR REALISTIC MIX
    -- ===========================================

    UPDATE "CommentNotification"
    SET is_read = TRUE, read_at = created_at + INTERVAL '2 hours'
    WHERE tenant_id = v_tenant_id
      AND id IN (
          SELECT id FROM "CommentNotification"
          WHERE tenant_id = v_tenant_id
          ORDER BY RANDOM()
          LIMIT 3
      );

    RAISE NOTICE 'Notifications seeder completed successfully';
END $$;

-- ===========================================
-- VERIFICATION QUERIES
-- ===========================================

-- Show notification counts by project scope
SELECT
    'Summary by Project Scope' as report,
    COALESCE(project_scope, 'GLOBAL') as scope,
    COUNT(*) as notification_count
FROM (
    SELECT
        CASE c.commentable_type
            WHEN 'Lead' THEN (SELECT p.name FROM "LeadProject" lp JOIN "Project" p ON lp.project_id = p.id WHERE lp.lead_id = c.commentable_id LIMIT 1)
            WHEN 'Deal' THEN (SELECT p.name FROM "Deal" d JOIN "Project" p ON d.project_id = p.id WHERE d.id = c.commentable_id)
            ELSE NULL
        END as project_scope
    FROM "CommentNotification" cn
    JOIN "Comment" c ON cn.comment_id = c.id
) sub
GROUP BY project_scope
ORDER BY notification_count DESC;

-- Show detailed notification list
SELECT
    u.email as notified_user,
    c.commentable_type as entity_type,
    c.commentable_id as entity_id,
    COALESCE(
        CASE c.commentable_type
            WHEN 'Lead' THEN (SELECT p.name FROM "LeadProject" lp JOIN "Project" p ON lp.project_id = p.id WHERE lp.lead_id = c.commentable_id LIMIT 1)
            WHEN 'Deal' THEN (SELECT p.name FROM "Deal" d JOIN "Project" p ON d.project_id = p.id WHERE d.id = c.commentable_id)
            ELSE NULL
        END,
        'GLOBAL'
    ) as project_scope,
    cn.notification_type,
    cn.is_read,
    LEFT(c.content, 50) || '...' as comment_preview
FROM "CommentNotification" cn
JOIN "User" u ON cn.user_id = u.id
JOIN "Comment" c ON cn.comment_id = c.id
ORDER BY cn.created_at DESC;

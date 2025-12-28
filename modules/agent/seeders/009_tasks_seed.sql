-- Seeder 009: Tasks sample data
-- Creates task statuses, priorities, and sample tasks linked to various entities

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id BIGINT;
    v_project_id BIGINT;
    v_deal_id BIGINT;
    v_lead_id BIGINT;
    v_account_id BIGINT;
    v_status_todo BIGINT;
    v_status_in_progress BIGINT;
    v_status_done BIGINT;
    v_status_blocked BIGINT;
    v_priority_low BIGINT;
    v_priority_medium BIGINT;
    v_priority_high BIGINT;
    v_priority_urgent BIGINT;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping tasks seeder';
        RETURN;
    END IF;

    -- Get bootstrap user for audit columns
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;

    -- Create Task Status entries
    INSERT INTO "Status" (name, description, category, is_terminal, created_at, updated_at)
    VALUES
        ('To Do', 'Task not yet started', 'task_status', false, NOW(), NOW()),
        ('In Progress', 'Task currently being worked on', 'task_status', false, NOW(), NOW()),
        ('Done', 'Task completed', 'task_status', true, NOW(), NOW()),
        ('Blocked', 'Task blocked by dependency', 'task_status', false, NOW(), NOW())
    ON CONFLICT (name) DO NOTHING;

    -- Create Task Priority entries
    INSERT INTO "Status" (name, description, category, is_terminal, created_at, updated_at)
    VALUES
        ('Low', 'Low priority task', 'task_priority', false, NOW(), NOW()),
        ('Medium', 'Medium priority task', 'task_priority', false, NOW(), NOW()),
        ('High', 'High priority task', 'task_priority', false, NOW(), NOW()),
        ('Urgent', 'Urgent priority task', 'task_priority', false, NOW(), NOW())
    ON CONFLICT (name) DO NOTHING;

    -- Get status IDs
    SELECT id INTO v_status_todo FROM "Status" WHERE name = 'To Do' AND category = 'task_status';
    SELECT id INTO v_status_in_progress FROM "Status" WHERE name = 'In Progress' AND category = 'task_status';
    SELECT id INTO v_status_done FROM "Status" WHERE name = 'Done' AND category = 'task_status';
    SELECT id INTO v_status_blocked FROM "Status" WHERE name = 'Blocked' AND category = 'task_status';

    -- Get priority IDs
    SELECT id INTO v_priority_low FROM "Status" WHERE name = 'Low' AND category = 'task_priority';
    SELECT id INTO v_priority_medium FROM "Status" WHERE name = 'Medium' AND category = 'task_priority';
    SELECT id INTO v_priority_high FROM "Status" WHERE name = 'High' AND category = 'task_priority';
    SELECT id INTO v_priority_urgent FROM "Status" WHERE name = 'Urgent' AND category = 'task_priority';

    -- Get sample entities
    SELECT id INTO v_project_id FROM "Project" WHERE tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_deal_id FROM "Deal" WHERE tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_lead_id FROM "Lead" WHERE tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_account_id FROM "Account" WHERE tenant_id = v_tenant_id LIMIT 1;

    -- Insert tasks for Project
    IF v_project_id IS NOT NULL THEN
        INSERT INTO "Task" (
            tenant_id, taskable_type, taskable_id, title, description,
            current_status_id, priority_id, start_date, due_date,
            estimated_hours, actual_hours, complexity, sort_order,
            created_by, updated_by, created_at, updated_at
        )
        VALUES
            (
                v_tenant_id, 'Project', v_project_id,
                'Define project milestones',
                'Break down project into key milestones and deliverables',
                v_status_todo, v_priority_high,
                CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days',
                8.0, NULL, 8, 1,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Project', v_project_id,
                'Set up project tracking',
                'Configure project management tools and dashboards',
                v_status_in_progress, v_priority_medium,
                CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '3 days',
                4.0, 2.5, 3, 2,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Project', v_project_id,
                'Kickoff meeting',
                'Schedule and conduct project kickoff meeting',
                v_status_done, v_priority_high,
                CURRENT_DATE - INTERVAL '5 days', CURRENT_DATE - INTERVAL '2 days',
                2.0, 1.5, 2, 3,
                v_user_id, v_user_id, NOW(), NOW()
            )
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added tasks for Project ID: %', v_project_id;
    END IF;

    -- Insert tasks for Deal
    IF v_deal_id IS NOT NULL THEN
        INSERT INTO "Task" (
            tenant_id, taskable_type, taskable_id, title, description,
            current_status_id, priority_id, start_date, due_date,
            estimated_hours, actual_hours, complexity, sort_order,
            created_by, updated_by, created_at, updated_at
        )
        VALUES
            (
                v_tenant_id, 'Deal', v_deal_id,
                'Send proposal',
                'Prepare and send pricing proposal to client',
                v_status_todo, v_priority_high,
                CURRENT_DATE + INTERVAL '1 day', CURRENT_DATE + INTERVAL '5 days',
                6.0, NULL, 5, 1,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Deal', v_deal_id,
                'Schedule demo',
                'Set up product demonstration call',
                v_status_in_progress, v_priority_medium,
                CURRENT_DATE, CURRENT_DATE + INTERVAL '2 days',
                1.0, 0.5, 1, 2,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Deal', v_deal_id,
                'Follow up on contract',
                'Check status of contract review',
                v_status_blocked, v_priority_urgent,
                CURRENT_DATE - INTERVAL '3 days', CURRENT_DATE + INTERVAL '10 days',
                2.0, 1.0, 2, 3,
                v_user_id, v_user_id, NOW(), NOW()
            )
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added tasks for Deal ID: %', v_deal_id;
    END IF;

    -- Insert tasks for Lead
    IF v_lead_id IS NOT NULL THEN
        INSERT INTO "Task" (
            tenant_id, taskable_type, taskable_id, title, description,
            current_status_id, priority_id, start_date, due_date,
            estimated_hours, actual_hours, complexity, sort_order,
            created_by, updated_by, created_at, updated_at
        )
        VALUES
            (
                v_tenant_id, 'Lead', v_lead_id,
                'Research company background',
                'Gather information about the company and decision makers',
                v_status_done, v_priority_medium,
                CURRENT_DATE - INTERVAL '7 days', CURRENT_DATE - INTERVAL '5 days',
                3.0, 2.5, 3, 1,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Lead', v_lead_id,
                'Prepare interview questions',
                'Draft questions for upcoming interview',
                v_status_in_progress, v_priority_high,
                CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '1 day',
                2.0, 1.0, 2, 2,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Lead', v_lead_id,
                'Update resume for role',
                'Tailor resume to match job requirements',
                v_status_todo, v_priority_medium,
                CURRENT_DATE + INTERVAL '1 day', CURRENT_DATE + INTERVAL '3 days',
                4.0, NULL, 5, 3,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Lead', v_lead_id,
                'Practice technical interview',
                'Review common technical questions and practice answers',
                v_status_todo, v_priority_high,
                CURRENT_DATE + INTERVAL '2 days', CURRENT_DATE + INTERVAL '5 days',
                8.0, NULL, 13, 4,
                v_user_id, v_user_id, NOW(), NOW()
            )
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added tasks for Lead ID: %', v_lead_id;
    END IF;

    -- Insert tasks for Account
    IF v_account_id IS NOT NULL THEN
        INSERT INTO "Task" (
            tenant_id, taskable_type, taskable_id, title, description,
            current_status_id, priority_id, start_date, due_date,
            estimated_hours, actual_hours, complexity, sort_order,
            created_by, updated_by, created_at, updated_at
        )
        VALUES
            (
                v_tenant_id, 'Account', v_account_id,
                'Update contact information',
                'Verify and update all contact details',
                v_status_todo, v_priority_low,
                CURRENT_DATE + INTERVAL '7 days', CURRENT_DATE + INTERVAL '14 days',
                1.0, NULL, 1, 1,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Account', v_account_id,
                'Quarterly business review',
                'Schedule QBR meeting with stakeholders',
                v_status_todo, v_priority_medium,
                CURRENT_DATE + INTERVAL '14 days', CURRENT_DATE + INTERVAL '30 days',
                4.0, NULL, 5, 2,
                v_user_id, v_user_id, NOW(), NOW()
            ),
            (
                v_tenant_id, 'Account', v_account_id,
                'Renewal discussion',
                'Prepare materials for contract renewal negotiation',
                v_status_todo, v_priority_high,
                CURRENT_DATE + INTERVAL '20 days', CURRENT_DATE + INTERVAL '45 days',
                12.0, NULL, 8, 3,
                v_user_id, v_user_id, NOW(), NOW()
            )
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added tasks for Account ID: %', v_account_id;
    END IF;

    -- Insert some unassigned tasks (no entity)
    INSERT INTO "Task" (
        tenant_id, taskable_type, taskable_id, title, description,
        current_status_id, priority_id, start_date, due_date,
        estimated_hours, actual_hours, complexity, sort_order,
        created_by, updated_by, created_at, updated_at
    )
    VALUES
        (
            v_tenant_id, NULL, NULL,
            'Review weekly goals',
            'Check progress on weekly objectives',
            v_status_todo, v_priority_medium,
            CURRENT_DATE, CURRENT_DATE + INTERVAL '1 day',
            0.5, NULL, 1, 1,
            v_user_id, v_user_id, NOW(), NOW()
        ),
        (
            v_tenant_id, NULL, NULL,
            'Update LinkedIn profile',
            'Refresh profile with recent achievements',
            v_status_todo, v_priority_low,
            CURRENT_DATE + INTERVAL '3 days', CURRENT_DATE + INTERVAL '7 days',
            2.0, NULL, 2, 2,
            v_user_id, v_user_id, NOW(), NOW()
        ),
        (
            v_tenant_id, NULL, NULL,
            'Plan next sprint',
            'Define goals and tasks for upcoming sprint',
            v_status_in_progress, v_priority_high,
            CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '2 days',
            3.0, 1.5, 5, 3,
            v_user_id, v_user_id, NOW(), NOW()
        ),
        (
            v_tenant_id, NULL, NULL,
            'Refactor authentication module',
            'Clean up and optimize the auth codebase',
            v_status_blocked, v_priority_medium,
            CURRENT_DATE - INTERVAL '5 days', CURRENT_DATE + INTERVAL '10 days',
            16.0, 4.0, 21, 4,
            v_user_id, v_user_id, NOW(), NOW()
        )
    ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Tasks seeder completed for tenant: %', v_tenant_id;
END $$;

-- Show counts
SELECT
    (SELECT COUNT(*) FROM "Status" WHERE category = 'task_status') as task_statuses,
    (SELECT COUNT(*) FROM "Status" WHERE category = 'task_priority') as task_priorities,
    (SELECT COUNT(*) FROM "Task") as total_tasks,
    (SELECT COUNT(*) FROM "Task" WHERE taskable_type = 'Project') as project_tasks,
    (SELECT COUNT(*) FROM "Task" WHERE taskable_type = 'Deal') as deal_tasks,
    (SELECT COUNT(*) FROM "Task" WHERE taskable_type = 'Lead') as lead_tasks,
    (SELECT COUNT(*) FROM "Task" WHERE taskable_type = 'Account') as account_tasks,
    (SELECT COUNT(*) FROM "Task" WHERE taskable_type IS NULL) as unassigned_tasks;

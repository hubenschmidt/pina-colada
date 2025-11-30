-- Seeder 009: Tasks sample data
-- Creates task statuses, priorities, and sample tasks linked to various entities

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_project_id BIGINT;
    v_deal_id BIGINT;
    v_lead_id BIGINT;
    v_account_id BIGINT;
    v_status_todo BIGINT;
    v_status_in_progress BIGINT;
    v_status_done BIGINT;
    v_priority_low BIGINT;
    v_priority_medium BIGINT;
    v_priority_high BIGINT;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada';

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping tasks seeder';
        RETURN;
    END IF;

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

    -- Get priority IDs
    SELECT id INTO v_priority_low FROM "Status" WHERE name = 'Low' AND category = 'task_priority';
    SELECT id INTO v_priority_medium FROM "Status" WHERE name = 'Medium' AND category = 'task_priority';
    SELECT id INTO v_priority_high FROM "Status" WHERE name = 'High' AND category = 'task_priority';

    -- Get sample entities
    SELECT id INTO v_project_id FROM "Project" WHERE tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_deal_id FROM "Deal" WHERE tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_lead_id FROM "Lead" WHERE tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_account_id FROM "Account" WHERE tenant_id = v_tenant_id LIMIT 1;

    -- Insert tasks for Project
    IF v_project_id IS NOT NULL THEN
        INSERT INTO "Task" (taskable_type, taskable_id, title, description, current_status_id, priority_id, due_date, created_at, updated_at)
        VALUES
            ('Project', v_project_id, 'Define project milestones', 'Break down project into key milestones and deliverables', v_status_todo, v_priority_high, CURRENT_DATE + INTERVAL '7 days', NOW(), NOW()),
            ('Project', v_project_id, 'Set up project tracking', 'Configure project management tools and dashboards', v_status_in_progress, v_priority_medium, CURRENT_DATE + INTERVAL '3 days', NOW(), NOW()),
            ('Project', v_project_id, 'Kickoff meeting', 'Schedule and conduct project kickoff meeting', v_status_done, v_priority_high, CURRENT_DATE - INTERVAL '2 days', NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added tasks for Project ID: %', v_project_id;
    END IF;

    -- Insert tasks for Deal
    IF v_deal_id IS NOT NULL THEN
        INSERT INTO "Task" (taskable_type, taskable_id, title, description, current_status_id, priority_id, due_date, created_at, updated_at)
        VALUES
            ('Deal', v_deal_id, 'Send proposal', 'Prepare and send pricing proposal to client', v_status_todo, v_priority_high, CURRENT_DATE + INTERVAL '5 days', NOW(), NOW()),
            ('Deal', v_deal_id, 'Schedule demo', 'Set up product demonstration call', v_status_in_progress, v_priority_medium, CURRENT_DATE + INTERVAL '2 days', NOW(), NOW()),
            ('Deal', v_deal_id, 'Follow up on contract', 'Check status of contract review', v_status_todo, v_priority_high, CURRENT_DATE + INTERVAL '10 days', NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added tasks for Deal ID: %', v_deal_id;
    END IF;

    -- Insert tasks for Lead
    IF v_lead_id IS NOT NULL THEN
        INSERT INTO "Task" (taskable_type, taskable_id, title, description, current_status_id, priority_id, due_date, created_at, updated_at)
        VALUES
            ('Lead', v_lead_id, 'Research company background', 'Gather information about the company and decision makers', v_status_done, v_priority_medium, CURRENT_DATE - INTERVAL '5 days', NOW(), NOW()),
            ('Lead', v_lead_id, 'Prepare interview questions', 'Draft questions for upcoming interview', v_status_in_progress, v_priority_high, CURRENT_DATE + INTERVAL '1 day', NOW(), NOW()),
            ('Lead', v_lead_id, 'Update resume for role', 'Tailor resume to match job requirements', v_status_todo, v_priority_medium, CURRENT_DATE + INTERVAL '3 days', NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added tasks for Lead ID: %', v_lead_id;
    END IF;

    -- Insert tasks for Account
    IF v_account_id IS NOT NULL THEN
        INSERT INTO "Task" (taskable_type, taskable_id, title, description, current_status_id, priority_id, due_date, created_at, updated_at)
        VALUES
            ('Account', v_account_id, 'Update contact information', 'Verify and update all contact details', v_status_todo, v_priority_low, CURRENT_DATE + INTERVAL '14 days', NOW(), NOW()),
            ('Account', v_account_id, 'Quarterly business review', 'Schedule QBR meeting with stakeholders', v_status_todo, v_priority_medium, CURRENT_DATE + INTERVAL '30 days', NOW(), NOW())
        ON CONFLICT DO NOTHING;
        RAISE NOTICE 'Added tasks for Account ID: %', v_account_id;
    END IF;

    -- Insert some unassigned tasks (no entity)
    INSERT INTO "Task" (taskable_type, taskable_id, title, description, current_status_id, priority_id, due_date, created_at, updated_at)
    VALUES
        (NULL, NULL, 'Review weekly goals', 'Check progress on weekly objectives', v_status_todo, v_priority_medium, CURRENT_DATE + INTERVAL '1 day', NOW(), NOW()),
        (NULL, NULL, 'Update LinkedIn profile', 'Refresh profile with recent achievements', v_status_todo, v_priority_low, CURRENT_DATE + INTERVAL '7 days', NOW(), NOW())
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

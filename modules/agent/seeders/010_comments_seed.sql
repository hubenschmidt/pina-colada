-- Seeder 010: Comprehensive Comments
-- Creates a large number of comments across all available entities for easy testing and discovery
-- Adds multiple comments to Accounts, Leads, Deals, and Tasks
-- Note: Comment table supports: Account, Lead, Deal, Task (per migration 037_comment_table.sql)
-- Uses multiple users to demonstrate multi-user comment attribution

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id_william BIGINT;
    v_user_id_jennifer BIGINT;
    v_comments_inserted INTEGER := 0;
    v_account_count INTEGER := 0;
    v_lead_count INTEGER := 0;
    v_deal_count INTEGER := 0;
    v_task_count INTEGER := 0;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping comments seeder';
        RETURN;
    END IF;

    -- Get William's user ID (primary user)
    SELECT id INTO v_user_id_william FROM "User" WHERE email = 'whubenschmidt@gmail.com' AND tenant_id = v_tenant_id LIMIT 1;

    -- Get Jennifer Lev's user ID (created in 001_initial_seed.sql)
    SELECT id INTO v_user_id_jennifer FROM "User" WHERE email = 'jennifervlev@gmail.com' AND tenant_id = v_tenant_id LIMIT 1;

    IF v_user_id_jennifer IS NULL THEN
        RAISE NOTICE 'Jennifer Lev user not found - she should be created in 001_initial_seed.sql';
    END IF;

    -- Fallback: if William not found, use any user
    IF v_user_id_william IS NULL THEN
        SELECT id INTO v_user_id_william FROM "User" WHERE tenant_id = v_tenant_id LIMIT 1;
    END IF;

    -- Individual comments (3 comments per individual, alternating between users)
    INSERT INTO "Comment" (tenant_id, commentable_type, commentable_id, content, created_by, created_at, updated_at)
    SELECT
        v_tenant_id,
        'Individual',
        i.id,
        CASE (s.comment_num % 5)
            WHEN 0 THEN 'Contact created. Initial research completed. Ready for outreach and engagement.'
            WHEN 1 THEN 'Contact activity: Multiple touchpoints this month. Engagement increasing steadily.'
            WHEN 2 THEN 'Contact health: Green. Responsive and engaged. Good relationship momentum building.'
            WHEN 3 THEN 'Key contact - high potential value. Prioritizing for Q1 2025 initiatives.'
            WHEN 4 THEN 'Recent interaction: Expressed strong interest. Following up next week.'
        END,
        CASE WHEN s.comment_num % 2 = 0 THEN v_user_id_william ELSE v_user_id_jennifer END,
        NOW() - (s.comment_num || ' days')::INTERVAL,
        NOW() - (s.comment_num || ' days')::INTERVAL
    FROM "Individual" i
    JOIN "Account" a ON i.account_id = a.id
    CROSS JOIN generate_series(1, 3) AS s(comment_num)
    WHERE i.id IS NOT NULL AND a.tenant_id = v_tenant_id;

    GET DIAGNOSTICS v_account_count = ROW_COUNT;
    v_comments_inserted := v_comments_inserted + v_account_count;
    RAISE NOTICE 'Added % comments for Individuals', v_account_count;

    -- Organization comments (3 comments per organization, alternating between users)
    INSERT INTO "Comment" (tenant_id, commentable_type, commentable_id, content, created_by, created_at, updated_at)
    SELECT
        v_tenant_id,
        'Organization',
        o.id,
        CASE (s.comment_num % 5)
            WHEN 0 THEN 'Company profile created. Initial research completed. Ready for outreach.'
            WHEN 1 THEN 'Company activity: Multiple touchpoints this month. Engagement increasing.'
            WHEN 2 THEN 'Company health: Green. Key contacts responsive. Good momentum building.'
            WHEN 3 THEN 'Strategic company - high potential value. Prioritizing for Q1 2025.'
            WHEN 4 THEN 'Recent interaction: Decision maker expressed interest. Following up.'
        END,
        CASE WHEN s.comment_num % 2 = 0 THEN v_user_id_william ELSE v_user_id_jennifer END,
        NOW() - (s.comment_num || ' days')::INTERVAL,
        NOW() - (s.comment_num || ' days')::INTERVAL
    FROM "Organization" o
    JOIN "Account" a ON o.account_id = a.id
    CROSS JOIN generate_series(1, 3) AS s(comment_num)
    WHERE o.id IS NOT NULL AND a.tenant_id = v_tenant_id;

    GET DIAGNOSTICS v_account_count = ROW_COUNT;
    v_comments_inserted := v_comments_inserted + v_account_count;
    RAISE NOTICE 'Added % comments for Organizations', v_account_count;

    -- Lead comments (Jobs, Opportunities, Partnerships) - 4 comments per lead, alternating between users
    INSERT INTO "Comment" (tenant_id, commentable_type, commentable_id, content, created_by, created_at, updated_at)
    SELECT
        v_tenant_id,
        'Lead',
        l.id,
        CASE
            WHEN l.type = 'Job' THEN
                CASE (s.comment_num % 4)
                    WHEN 0 THEN 'Great company culture and remote-friendly position. Team seems collaborative and supportive.'
                    WHEN 1 THEN 'Technical interview scheduled for next Tuesday at 2pm. Preparing system design questions and reviewing company tech stack.'
                    WHEN 2 THEN 'Salary range confirmed: $150K-$180K + equity. Benefits package is competitive with good health coverage.'
                    WHEN 3 THEN 'Application submitted successfully. Received automated confirmation email. Waiting for recruiter response within 3-5 business days.'
                END
            WHEN l.type = 'Opportunity' THEN
                CASE (s.comment_num % 4)
                    WHEN 0 THEN 'High priority opportunity. Competitor is also pitching. Need to emphasize our unique value proposition and faster implementation.'
                    WHEN 1 THEN 'Timeline: Decision expected by end of month. Proposal deadline is next Friday. Working on final pricing and terms.'
                    WHEN 2 THEN 'Technical requirements reviewed thoroughly. Our solution is a good fit. Custom integration needed for their legacy systems.'
                    WHEN 3 THEN 'Budget confirmed at $250K. Decision maker is VP of Sales. Final approval from CEO required but likely to proceed.'
                END
            WHEN l.type = 'Partnership' THEN
                CASE (s.comment_num % 4)
                    WHEN 0 THEN 'Strategic partnership opportunity identified. They have strong distribution channels in our target market segments.'
                    WHEN 1 THEN 'Partnership terms discussed in detail. Revenue share model: 70/30 split in our favor. Initial term: 12 months with renewal option.'
                    WHEN 2 THEN 'Legal review in progress. MSA template sent to their legal team. Expecting signed agreement by next week.'
                    WHEN 3 THEN 'Partnership benefits: Access to 50K+ customer base. Co-marketing opportunities. Joint product development potential.'
                END
            ELSE
                CASE (s.comment_num % 2)
                    WHEN 0 THEN 'Initial contact established. Following up with more detailed information and case studies.'
                    WHEN 1 THEN 'Status update: Moving forward with next steps. Scheduling discovery call to understand requirements better.'
                END
        END,
        CASE WHEN s.comment_num % 2 = 0 THEN v_user_id_william ELSE v_user_id_jennifer END,
        NOW() - (s.comment_num || ' days')::INTERVAL,
        NOW() - (s.comment_num || ' days')::INTERVAL
    FROM "Lead" l
    CROSS JOIN generate_series(1, 4) AS s(comment_num)
    WHERE l.id IS NOT NULL AND l.tenant_id = v_tenant_id;
    
    GET DIAGNOSTICS v_lead_count = ROW_COUNT;
    v_comments_inserted := v_comments_inserted + v_lead_count;
    RAISE NOTICE 'Added % comments for Leads', v_lead_count;

    -- Deal comments (3 comments per deal, alternating between users)
    INSERT INTO "Comment" (tenant_id, commentable_type, commentable_id, content, created_by, created_at, updated_at)
    SELECT
        v_tenant_id,
        'Deal',
        d.id,
        CASE (s.comment_num % 5)
            WHEN 0 THEN 'Deal created. Initial value estimated at $500K. Probability set to 25% based on early discovery stage.'
            WHEN 1 THEN 'Discovery phase completed successfully. Requirements gathered comprehensively. Moving to proposal stage next week.'
            WHEN 2 THEN 'Proposal submitted to client. They are reviewing technical specifications and pricing. Expecting feedback by Friday.'
            WHEN 3 THEN 'Negotiation in progress. Discussing payment terms (net 30 vs net 60) and implementation timeline (3-6 months).'
            WHEN 4 THEN 'Deal status update: Client requested additional references. Provided 3 case studies from similar implementations.'
        END,
        CASE WHEN s.comment_num % 2 = 0 THEN v_user_id_william ELSE v_user_id_jennifer END,
        NOW() - (s.comment_num || ' days')::INTERVAL,
        NOW() - (s.comment_num || ' days')::INTERVAL
    FROM "Deal" d
    CROSS JOIN generate_series(1, 3) AS s(comment_num)
    WHERE d.id IS NOT NULL AND d.tenant_id = v_tenant_id;
    
    GET DIAGNOSTICS v_deal_count = ROW_COUNT;
    v_comments_inserted := v_comments_inserted + v_deal_count;
    RAISE NOTICE 'Added % comments for Deals', v_deal_count;

    -- Task comments (2 comments per task, alternating between users)
    INSERT INTO "Comment" (tenant_id, commentable_type, commentable_id, content, created_by, created_at, updated_at)
    SELECT
        v_tenant_id,
        'Task',
        t.id,
        CASE (s.comment_num % 3)
            WHEN 0 THEN 'Task assigned. Reviewing requirements and gathering necessary resources to begin work.'
            WHEN 1 THEN 'Progress update: 50% complete. On track for deadline. No blockers currently.'
            WHEN 2 THEN 'Task completed successfully. Delivered ahead of schedule. Ready for review and approval.'
        END,
        CASE WHEN s.comment_num % 2 = 0 THEN v_user_id_william ELSE v_user_id_jennifer END,
        NOW() - (s.comment_num || ' days')::INTERVAL,
        NOW() - (s.comment_num || ' days')::INTERVAL
    FROM "Task" t
    CROSS JOIN generate_series(1, 2) AS s(comment_num)
    WHERE t.id IS NOT NULL;
    
    GET DIAGNOSTICS v_task_count = ROW_COUNT;
    v_comments_inserted := v_comments_inserted + v_task_count;
    RAISE NOTICE 'Added % comments for Tasks', v_task_count;

    RAISE NOTICE 'Comments seeder completed. Total comments inserted: %', v_comments_inserted;
END $$;

-- Show counts for verification
SELECT
    (SELECT COUNT(*) FROM "Comment") as total_comments,
    (SELECT COUNT(*) FROM "Comment" WHERE commentable_type = 'Individual') as individual_comments,
    (SELECT COUNT(*) FROM "Comment" WHERE commentable_type = 'Organization') as organization_comments,
    (SELECT COUNT(*) FROM "Comment" WHERE commentable_type = 'Lead') as lead_comments,
    (SELECT COUNT(*) FROM "Comment" WHERE commentable_type = 'Deal') as deal_comments,
    (SELECT COUNT(*) FROM "Comment" WHERE commentable_type = 'Task') as task_comments,
    (SELECT COUNT(DISTINCT commentable_id) FROM "Comment" WHERE commentable_type = 'Individual') as individuals_with_comments,
    (SELECT COUNT(DISTINCT commentable_id) FROM "Comment" WHERE commentable_type = 'Organization') as organizations_with_comments,
    (SELECT COUNT(DISTINCT commentable_id) FROM "Comment" WHERE commentable_type = 'Lead') as leads_with_comments,
    (SELECT COUNT(DISTINCT commentable_id) FROM "Comment" WHERE commentable_type = 'Deal') as deals_with_comments,
    (SELECT COUNT(DISTINCT commentable_id) FROM "Comment" WHERE commentable_type = 'Task') as tasks_with_comments;

-- Show comment distribution by user for verification
SELECT
    u.first_name || ' ' || u.last_name as user_name,
    u.email,
    COUNT(c.id) as comment_count
FROM "Comment" c
JOIN "User" u ON c.created_by = u.id
GROUP BY u.id, u.first_name, u.last_name, u.email
ORDER BY comment_count DESC;


-- Seeder 010: Comprehensive Notes/Comments
-- Creates a large number of notes across all available entities for easy testing and discovery
-- Adds multiple notes to Organizations, Individuals, Contacts, Leads, Deals, Projects, and Accounts

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id BIGINT;
    v_notes_inserted INTEGER := 0;
    v_org_count INTEGER := 0;
    v_ind_count INTEGER := 0;
    v_contact_count INTEGER := 0;
    v_lead_count INTEGER := 0;
    v_deal_count INTEGER := 0;
    v_project_count INTEGER := 0;
    v_account_count INTEGER := 0;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping notes seeder';
        RETURN;
    END IF;

    -- Get a user ID for created_by (use first available user)
    SELECT id INTO v_user_id FROM "User" WHERE tenant_id = v_tenant_id LIMIT 1;

    -- Sample note templates for variety
    -- Organization notes (3 notes per organization)
    INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, created_at, updated_at)
    SELECT 
        v_tenant_id,
        'Organization',
        o.id,
        CASE (s.note_num % 5)
            WHEN 0 THEN 'Initial discovery call completed. Company is actively seeking solutions in our space.'
            WHEN 1 THEN 'Follow-up meeting scheduled for next week. Need to prepare technical demo.'
            WHEN 2 THEN 'Decision maker identified: VP of Engineering. Budget approved for Q1 2025.'
            WHEN 3 THEN 'Competitive analysis: We have strong differentiation in pricing and features.'
            WHEN 4 THEN 'Contract negotiation in progress. Legal review expected by end of month.'
        END,
        v_user_id,
        NOW() - (s.note_num || ' days')::INTERVAL,
        NOW() - (s.note_num || ' days')::INTERVAL
    FROM "Organization" o
    CROSS JOIN generate_series(1, 3) AS s(note_num)
    WHERE o.id IS NOT NULL;
    
    GET DIAGNOSTICS v_org_count = ROW_COUNT;
    v_notes_inserted := v_notes_inserted + v_org_count;
    RAISE NOTICE 'Added % notes for Organizations', v_org_count;

    -- Individual notes (2 notes per individual)
    INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, created_at, updated_at)
    SELECT 
        v_tenant_id,
        'Individual',
        i.id,
        CASE (s.note_num % 4)
            WHEN 0 THEN 'Met at industry conference. Very knowledgeable about market trends and competitor landscape.'
            WHEN 1 THEN 'Prefers email communication over phone. Best time to reach is weekday mornings 9-11am.'
            WHEN 2 THEN 'Expressed interest in our enterprise solution. Wants to schedule a technical deep-dive.'
            WHEN 3 THEN 'LinkedIn profile shows strong background in our target market. Good potential advocate.'
        END,
        v_user_id,
        NOW() - (s.note_num || ' days')::INTERVAL,
        NOW() - (s.note_num || ' days')::INTERVAL
    FROM "Individual" i
    CROSS JOIN generate_series(1, 2) AS s(note_num)
    WHERE i.id IS NOT NULL;
    
    GET DIAGNOSTICS v_ind_count = ROW_COUNT;
    v_notes_inserted := v_notes_inserted + v_ind_count;
    RAISE NOTICE 'Added % notes for Individuals', v_ind_count;

    -- Contact notes (2 notes per contact)
    INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, created_at, updated_at)
    SELECT 
        v_tenant_id,
        'Contact',
        c.id,
        CASE (s.note_num % 3)
            WHEN 0 THEN 'Primary point of contact for technical discussions and implementation planning.'
            WHEN 1 THEN 'Has authority to sign contracts up to $50K. Larger deals require CFO approval.'
            WHEN 2 THEN 'Very responsive to emails. Usually replies within 2-4 hours during business days.'
        END,
        v_user_id,
        NOW() - (s.note_num || ' days')::INTERVAL,
        NOW() - (s.note_num || ' days')::INTERVAL
    FROM "Contact" c
    CROSS JOIN generate_series(1, 2) AS s(note_num)
    WHERE c.id IS NOT NULL;
    
    GET DIAGNOSTICS v_contact_count = ROW_COUNT;
    v_notes_inserted := v_notes_inserted + v_contact_count;
    RAISE NOTICE 'Added % notes for Contacts', v_contact_count;

    -- Lead notes (Jobs, Opportunities, Partnerships) - 3 notes per lead
    INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, created_at, updated_at)
    SELECT 
        v_tenant_id,
        'Lead',
        l.id,
        CASE 
            WHEN l.type = 'Job' THEN
                CASE (s.note_num % 4)
                    WHEN 0 THEN 'Great company culture and remote-friendly position. Team seems collaborative.'
                    WHEN 1 THEN 'Technical interview scheduled for next Tuesday at 2pm. Preparing system design questions.'
                    WHEN 2 THEN 'Salary range confirmed: $150K-$180K + equity. Benefits package is competitive.'
                    WHEN 3 THEN 'Application submitted. Received automated confirmation email. Waiting for recruiter response.'
                END
            WHEN l.type = 'Opportunity' THEN
                CASE (s.note_num % 4)
                    WHEN 0 THEN 'High priority opportunity. Competitor is also pitching. Need to emphasize our unique value prop.'
                    WHEN 1 THEN 'Timeline: Decision expected by end of month. Proposal deadline is next Friday.'
                    WHEN 2 THEN 'Technical requirements reviewed. Our solution is a good fit. Custom integration needed.'
                    WHEN 3 THEN 'Budget confirmed at $250K. Decision maker is VP of Sales. Final approval from CEO required.'
                END
            WHEN l.type = 'Partnership' THEN
                CASE (s.note_num % 3)
                    WHEN 0 THEN 'Strategic partnership opportunity. They have strong distribution channels in our target market.'
                    WHEN 1 THEN 'Partnership terms discussed. Revenue share model: 70/30 split. Initial term: 12 months.'
                    WHEN 2 THEN 'Legal review in progress. MSA template sent. Expecting signed agreement by next week.'
                END
            ELSE
                CASE (s.note_num % 2)
                    WHEN 0 THEN 'Initial contact established. Following up with more information.'
                    WHEN 1 THEN 'Status update: Moving forward with next steps.'
                END
        END,
        v_user_id,
        NOW() - (s.note_num || ' days')::INTERVAL,
        NOW() - (s.note_num || ' days')::INTERVAL
    FROM "Lead" l
    CROSS JOIN generate_series(1, 3) AS s(note_num)
    WHERE l.id IS NOT NULL;
    
    GET DIAGNOSTICS v_lead_count = ROW_COUNT;
    v_notes_inserted := v_notes_inserted + v_lead_count;
    RAISE NOTICE 'Added % notes for Leads', v_lead_count;

    -- Deal notes (2 notes per deal)
    INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, created_at, updated_at)
    SELECT 
        v_tenant_id,
        'Deal',
        d.id,
        CASE (s.note_num % 4)
            WHEN 0 THEN 'Deal created. Initial value estimated at $500K. Probability set to 25% based on early stage.'
            WHEN 1 THEN 'Discovery phase completed. Requirements gathered. Moving to proposal stage.'
            WHEN 2 THEN 'Proposal submitted. Client is reviewing technical specifications and pricing.'
            WHEN 3 THEN 'Negotiation in progress. Discussing payment terms and implementation timeline.'
        END,
        v_user_id,
        NOW() - (s.note_num || ' days')::INTERVAL,
        NOW() - (s.note_num || ' days')::INTERVAL
    FROM "Deal" d
    CROSS JOIN generate_series(1, 2) AS s(note_num)
    WHERE d.id IS NOT NULL AND d.tenant_id = v_tenant_id;
    
    GET DIAGNOSTICS v_deal_count = ROW_COUNT;
    v_notes_inserted := v_notes_inserted + v_deal_count;
    RAISE NOTICE 'Added % notes for Deals', v_deal_count;

    -- Project notes (2 notes per project)
    INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, created_at, updated_at)
    SELECT 
        v_tenant_id,
        'Project',
        p.id,
        CASE (s.note_num % 3)
            WHEN 0 THEN 'Project kickoff meeting completed. Team aligned on goals and timeline.'
            WHEN 1 THEN 'Milestone 1 achieved ahead of schedule. On track for Q1 delivery.'
            WHEN 2 THEN 'Weekly standup notes: Blockers resolved. Next sprint planning scheduled.'
        END,
        v_user_id,
        NOW() - (s.note_num || ' days')::INTERVAL,
        NOW() - (s.note_num || ' days')::INTERVAL
    FROM "Project" p
    CROSS JOIN generate_series(1, 2) AS s(note_num)
    WHERE p.id IS NOT NULL AND p.tenant_id = v_tenant_id;
    
    GET DIAGNOSTICS v_project_count = ROW_COUNT;
    v_notes_inserted := v_notes_inserted + v_project_count;
    RAISE NOTICE 'Added % notes for Projects', v_project_count;

    -- Account notes (polymorphic - can be Organization or Individual accounts) - 2 notes per account
    INSERT INTO "Note" (tenant_id, entity_type, entity_id, content, created_by, created_at, updated_at)
    SELECT 
        v_tenant_id,
        'Account',
        a.id,
        CASE (s.note_num % 3)
            WHEN 0 THEN 'Account created. Initial research completed. Ready for outreach.'
            WHEN 1 THEN 'Account activity: Multiple touchpoints this month. Engagement increasing.'
            WHEN 2 THEN 'Account health: Green. All key contacts responsive. Good relationship momentum.'
        END,
        v_user_id,
        NOW() - (s.note_num || ' days')::INTERVAL,
        NOW() - (s.note_num || ' days')::INTERVAL
    FROM "Account" a
    CROSS JOIN generate_series(1, 2) AS s(note_num)
    WHERE a.id IS NOT NULL AND a.tenant_id = v_tenant_id;
    
    GET DIAGNOSTICS v_account_count = ROW_COUNT;
    v_notes_inserted := v_notes_inserted + v_account_count;
    RAISE NOTICE 'Added % notes for Accounts', v_account_count;

    RAISE NOTICE 'Notes seeder completed. Total notes inserted: %', v_notes_inserted;
END $$;

-- Show counts for verification
SELECT
    (SELECT COUNT(*) FROM "Note") as total_notes,
    (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Organization') as organization_notes,
    (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Individual') as individual_notes,
    (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Contact') as contact_notes,
    (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Lead') as lead_notes,
    (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Deal') as deal_notes,
    (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Project') as project_notes,
    (SELECT COUNT(*) FROM "Note" WHERE entity_type = 'Account') as account_notes,
    (SELECT COUNT(DISTINCT entity_id) FROM "Note" WHERE entity_type = 'Organization') as organizations_with_notes,
    (SELECT COUNT(DISTINCT entity_id) FROM "Note" WHERE entity_type = 'Individual') as individuals_with_notes,
    (SELECT COUNT(DISTINCT entity_id) FROM "Note" WHERE entity_type = 'Contact') as contacts_with_notes,
    (SELECT COUNT(DISTINCT entity_id) FROM "Note" WHERE entity_type = 'Lead') as leads_with_notes,
    (SELECT COUNT(DISTINCT entity_id) FROM "Note" WHERE entity_type = 'Deal') as deals_with_notes,
    (SELECT COUNT(DISTINCT entity_id) FROM "Note" WHERE entity_type = 'Project') as projects_with_notes,
    (SELECT COUNT(DISTINCT entity_id) FROM "Note" WHERE entity_type = 'Account') as accounts_with_notes;


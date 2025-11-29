-- Seeder: Sample saved custom reports for demonstrating report builder
-- Creates various saved report definitions to showcase functionality

-- 1. Notes with Contacts
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'Notes with Contacts',
    'View all notes linked to contacts, filtered to show only notes with content',
    '{
        "primary_entity": "notes",
        "columns": ["id", "content", "entity_type", "created_at", "contact.first_name", "contact.last_name", "contact.email"],
        "joins": ["contacts"],
        "filters": [
            {"field": "entity_type", "operator": "eq", "value": "Contact"},
            {"field": "content", "operator": "is_not_null", "value": ""}
        ],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 2. Notes with Organizations
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'Notes with Organizations',
    'View all notes linked to organizations with company details',
    '{
        "primary_entity": "notes",
        "columns": ["id", "content", "created_at", "organization.name", "organization.website"],
        "joins": ["organizations"],
        "filters": [
            {"field": "entity_type", "operator": "eq", "value": "Organization"},
            {"field": "content", "operator": "is_not_null", "value": ""}
        ],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 3. Notes with Individuals
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'Notes with Individuals',
    'View all notes linked to individual accounts with person details',
    '{
        "primary_entity": "notes",
        "columns": ["id", "content", "created_at", "individual.first_name", "individual.last_name", "individual.email"],
        "joins": ["individuals"],
        "filters": [
            {"field": "entity_type", "operator": "eq", "value": "Individual"},
            {"field": "content", "operator": "is_not_null", "value": ""}
        ],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 4. Organizations Overview
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'Organizations Overview',
    'Complete list of organizations with firmographic data and funding info',
    '{
        "primary_entity": "organizations",
        "columns": ["id", "name", "website", "headquarters_city", "headquarters_state", "company_type", "founding_year", "employee_count_range.label", "funding_stage.name", "revenue_range.label"],
        "joins": [],
        "filters": [],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 5. Organizations with Notes
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'Organizations with Notes',
    'Organizations that have notes attached, showing latest note content',
    '{
        "primary_entity": "organizations",
        "columns": ["id", "name", "website", "headquarters_city", "note.content", "note.created_at"],
        "joins": ["notes"],
        "filters": [
            {"field": "note.content", "operator": "is_not_null", "value": ""}
        ],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 6. All Leads
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'All Leads',
    'Complete list of leads with source and type information',
    '{
        "primary_entity": "leads",
        "columns": ["id", "title", "description", "source", "type", "created_at"],
        "joins": [],
        "filters": [],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 7. Leads with Notes
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'Leads with Notes',
    'Leads that have notes attached for tracking activity',
    '{
        "primary_entity": "leads",
        "columns": ["id", "title", "source", "type", "note.content", "note.created_at"],
        "joins": ["notes"],
        "filters": [
            {"field": "note.content", "operator": "is_not_null", "value": ""}
        ],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 8. Decision Makers
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'Decision Makers',
    'Individuals identified as decision makers with their contact info',
    '{
        "primary_entity": "individuals",
        "columns": ["id", "first_name", "last_name", "email", "title", "department", "seniority_level", "linkedin_url"],
        "joins": [],
        "filters": [
            {"field": "is_decision_maker", "operator": "eq", "value": true}
        ],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 9. Contacts Directory
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'Contacts Directory',
    'All contacts with their associated organization',
    '{
        "primary_entity": "contacts",
        "columns": ["id", "first_name", "last_name", "email", "phone", "title", "is_primary", "organization.name"],
        "joins": [],
        "filters": [],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- 10. All Notes
INSERT INTO "SavedReport" (tenant_id, name, description, query_definition, created_at, updated_at)
SELECT
    t.id,
    'All Notes',
    'All notes across all entity types',
    '{
        "primary_entity": "notes",
        "columns": ["id", "entity_type", "entity_id", "content", "created_at"],
        "joins": [],
        "filters": [
            {"field": "content", "operator": "is_not_null", "value": ""}
        ],
        "limit": 100,
        "offset": 0
    }'::jsonb,
    NOW(),
    NOW()
FROM "Tenant" t WHERE t.slug = 'pinacolada'
ON CONFLICT DO NOTHING;

-- Return counts for seeder verification
SELECT
    (SELECT COUNT(*) FROM "SavedReport") as total_saved_reports,
    (SELECT COUNT(*) FROM "SavedReport" WHERE query_definition->>'primary_entity' = 'notes') as notes_reports,
    (SELECT COUNT(*) FROM "SavedReport" WHERE query_definition->>'primary_entity' = 'organizations') as org_reports,
    (SELECT COUNT(*) FROM "SavedReport" WHERE query_definition->>'primary_entity' = 'leads') as lead_reports,
    (SELECT COUNT(*) FROM "SavedReport" WHERE query_definition->>'primary_entity' = 'individuals') as individual_reports,
    (SELECT COUNT(*) FROM "SavedReport" WHERE query_definition->>'primary_entity' = 'contacts') as contact_reports;

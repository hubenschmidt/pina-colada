-- Seed Reasoning table with CRM schema mappings for RAG
-- This tells AI agents which tables exist for each reasoning context

INSERT INTO "Reasoning" (type, table_name, description, schema_hint) VALUES
-- Core Entities
('crm', 'Account', 'Groups of organizations and individuals', '{"key_fields": ["name", "tenant_id"]}'),
('crm', 'Organization', 'Companies and business entities', '{"key_fields": ["name", "website", "industry"]}'),
('crm', 'Individual', 'People/contacts', '{"key_fields": ["first_name", "last_name", "email", "title"]}'),
('crm', 'Contact', 'Contact points linking individuals and orgs', NULL),
-- Pipeline
('crm', 'Deal', 'Sales opportunities at tenant level', '{"key_fields": ["name", "value_amount", "probability"]}'),
('crm', 'Lead', 'Base lead entity (Job, Opportunity, Partnership)', '{"polymorphic": true}'),
('crm', 'Job', 'Job opportunity leads', '{"key_fields": ["title", "company_name", "status"]}'),
('crm', 'Opportunity', 'Sales opportunity leads', '{"key_fields": ["name", "value", "stage"]}'),
('crm', 'Partnership', 'Partnership leads', '{"key_fields": ["name", "type"]}'),
-- Work Management
('crm', 'Project', 'Project containers', '{"key_fields": ["name", "status"]}'),
('crm', 'Task', 'Work items (polymorphic attachment)', '{"polymorphic": true, "key_fields": ["title", "due_date", "status"]}'),
('crm', 'Note', 'Notes on any entity (polymorphic)', '{"polymorphic": true}'),
-- Assets
('crm', 'Asset', 'Base asset class', NULL),
('crm', 'Document', 'File attachments', '{"key_fields": ["filename", "mime_type"]}'),
-- Provenance
('crm', 'Data_Provenance', 'Field-level source tracking', '{"key_fields": ["entity_type", "entity_id", "field_name", "source", "confidence"]}')
ON CONFLICT (type, table_name) DO NOTHING;

SELECT COUNT(*) AS reasoning_rows FROM "Reasoning" WHERE type = 'crm';

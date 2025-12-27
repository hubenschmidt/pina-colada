-- Migration: 072_permission_tables.sql
-- Create Permission and Role_Permission tables for RBAC

-- 1. Create Permission table
CREATE TABLE IF NOT EXISTS "Permission" (
    id BIGSERIAL PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- 2. Create Role_Permission junction table
CREATE TABLE IF NOT EXISTS "Role_Permission" (
    role_id BIGINT REFERENCES "Role"(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES "Permission"(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- 3. Seed permissions
INSERT INTO "Permission" (resource, action, description) VALUES
-- Core entities
('individual', 'read', 'View individuals'),
('individual', 'create', 'Create individuals'),
('individual', 'update', 'Update individuals'),
('individual', 'delete', 'Delete individuals'),
('organization', 'read', 'View organizations'),
('organization', 'create', 'Create organizations'),
('organization', 'update', 'Update organizations'),
('organization', 'delete', 'Delete organizations'),
('contact', 'read', 'View contacts'),
('contact', 'create', 'Create contacts'),
('contact', 'update', 'Update contacts'),
('contact', 'delete', 'Delete contacts'),
('account', 'read', 'View accounts'),
('account', 'create', 'Create accounts'),
('account', 'update', 'Update accounts'),
('account', 'delete', 'Delete accounts'),
('job', 'read', 'View job leads'),
('job', 'create', 'Create job leads'),
('job', 'update', 'Update job leads'),
('job', 'delete', 'Delete job leads'),
('document', 'read', 'View documents'),
('document', 'create', 'Upload documents'),
('document', 'update', 'Update documents'),
('document', 'delete', 'Delete documents'),
-- Account Project (join table)
('account_project', 'create', 'Assign accounts to projects'),
('account_project', 'read', 'Read account-project assignments'),
('account_project', 'update', 'Update account-project assignments'),
('account_project', 'delete', 'Remove accounts from projects'),
-- Account Relationship
('account_relationship', 'create', 'Create account relationships'),
('account_relationship', 'read', 'Read account relationships'),
('account_relationship', 'update', 'Update account relationships'),
('account_relationship', 'delete', 'Delete account relationships'),
-- Asset
('asset', 'create', 'Create assets'),
('asset', 'read', 'Read assets'),
('asset', 'update', 'Update assets'),
('asset', 'delete', 'Delete assets'),
-- Comment
('comment', 'create', 'Create comments'),
('comment', 'read', 'Read comments'),
('comment', 'update', 'Update comments'),
('comment', 'delete', 'Delete comments'),
-- Deal
('deal', 'create', 'Create deals'),
('deal', 'read', 'Read deals'),
('deal', 'update', 'Update deals'),
('deal', 'delete', 'Delete deals'),
-- Individual Relationship
('individual_relationship', 'create', 'Create individual relationships'),
('individual_relationship', 'read', 'Read individual relationships'),
('individual_relationship', 'update', 'Update individual relationships'),
('individual_relationship', 'delete', 'Delete individual relationships'),
-- Industry
('industry', 'create', 'Create industries'),
('industry', 'read', 'Read industries'),
('industry', 'update', 'Update industries'),
('industry', 'delete', 'Delete industries'),
-- Lead
('lead', 'create', 'Create leads'),
('lead', 'read', 'Read leads'),
('lead', 'update', 'Update leads'),
('lead', 'delete', 'Delete leads'),
-- Lead Project (join table)
('lead_project', 'create', 'Assign leads to projects'),
('lead_project', 'read', 'Read lead-project assignments'),
('lead_project', 'update', 'Update lead-project assignments'),
('lead_project', 'delete', 'Remove leads from projects'),
-- Note
('note', 'create', 'Create notes'),
('note', 'read', 'Read notes'),
('note', 'update', 'Update notes'),
('note', 'delete', 'Delete notes'),
-- Opportunity
('opportunity', 'create', 'Create opportunities'),
('opportunity', 'read', 'Read opportunities'),
('opportunity', 'update', 'Update opportunities'),
('opportunity', 'delete', 'Delete opportunities'),
-- Organization Relationship
('organization_relationship', 'create', 'Create organization relationships'),
('organization_relationship', 'read', 'Read organization relationships'),
('organization_relationship', 'update', 'Update organization relationships'),
('organization_relationship', 'delete', 'Delete organization relationships'),
-- Organization Technology
('organization_technology', 'create', 'Create organization technologies'),
('organization_technology', 'read', 'Read organization technologies'),
('organization_technology', 'update', 'Update organization technologies'),
('organization_technology', 'delete', 'Delete organization technologies'),
-- Partnership
('partnership', 'create', 'Create partnerships'),
('partnership', 'read', 'Read partnerships'),
('partnership', 'update', 'Update partnerships'),
('partnership', 'delete', 'Delete partnerships'),
-- Project
('project', 'create', 'Create projects'),
('project', 'read', 'Read projects'),
('project', 'update', 'Update projects'),
('project', 'delete', 'Delete projects'),
-- Provenance
('provenance', 'create', 'Create provenance records'),
('provenance', 'read', 'Read provenance records'),
('provenance', 'update', 'Update provenance records'),
('provenance', 'delete', 'Delete provenance records'),
-- Saved Report
('saved_report', 'create', 'Create saved reports'),
('saved_report', 'read', 'Read saved reports'),
('saved_report', 'update', 'Update saved reports'),
('saved_report', 'delete', 'Delete saved reports'),
-- Saved Report Project (join table)
('saved_report_project', 'create', 'Assign saved reports to projects'),
('saved_report_project', 'read', 'Read saved report-project assignments'),
('saved_report_project', 'update', 'Update saved report-project assignments'),
('saved_report_project', 'delete', 'Remove saved reports from projects'),
-- Signal
('signal', 'create', 'Create signals'),
('signal', 'read', 'Read signals'),
('signal', 'update', 'Update signals'),
('signal', 'delete', 'Delete signals'),
-- Status
('status', 'create', 'Create statuses'),
('status', 'read', 'Read statuses'),
('status', 'update', 'Update statuses'),
('status', 'delete', 'Delete statuses'),
-- Tag
('tag', 'create', 'Create tags'),
('tag', 'read', 'Read tags'),
('tag', 'update', 'Update tags'),
('tag', 'delete', 'Delete tags'),
-- Task
('task', 'create', 'Create tasks'),
('task', 'read', 'Read tasks'),
('task', 'update', 'Update tasks'),
('task', 'delete', 'Delete tasks'),
-- Technology
('technology', 'create', 'Create technologies'),
('technology', 'read', 'Read technologies'),
('technology', 'update', 'Update technologies'),
('technology', 'delete', 'Delete technologies'),
-- Activity
('activity', 'create', 'Create activities'),
('activity', 'read', 'Read activities'),
('activity', 'update', 'Update activities'),
('activity', 'delete', 'Delete activities'),
-- Account Industry (join table)
('account_industry', 'create', 'Assign industries to accounts'),
('account_industry', 'read', 'Read account-industry assignments'),
('account_industry', 'update', 'Update account-industry assignments'),
('account_industry', 'delete', 'Remove industries from accounts'),
-- Contact Account (join table)
('contact_account', 'create', 'Link contacts to accounts'),
('contact_account', 'read', 'Read contact-account links'),
('contact_account', 'update', 'Update contact-account links'),
('contact_account', 'delete', 'Remove contact-account links'),
-- Conversation
('conversation', 'create', 'Create conversations'),
('conversation', 'read', 'Read conversations'),
('conversation', 'update', 'Update conversations'),
('conversation', 'delete', 'Delete conversations'),
-- Conversation Message
('conversation_message', 'create', 'Create conversation messages'),
('conversation_message', 'read', 'Read conversation messages'),
('conversation_message', 'update', 'Update conversation messages'),
('conversation_message', 'delete', 'Delete conversation messages'),
-- Entity Asset (polymorphic)
('entity_asset', 'create', 'Attach assets to entities'),
('entity_asset', 'read', 'Read entity-asset attachments'),
('entity_asset', 'update', 'Update entity-asset attachments'),
('entity_asset', 'delete', 'Remove assets from entities'),
-- Entity Tag (polymorphic)
('entity_tag', 'create', 'Tag entities'),
('entity_tag', 'read', 'Read entity tags'),
('entity_tag', 'update', 'Update entity tags'),
('entity_tag', 'delete', 'Remove tags from entities'),
-- Funding Round
('funding_round', 'create', 'Create funding rounds'),
('funding_round', 'read', 'Read funding rounds'),
('funding_round', 'update', 'Update funding rounds'),
('funding_round', 'delete', 'Delete funding rounds'),
-- Usage Analytics
('usage_analytics', 'create', 'Create usage analytics'),
('usage_analytics', 'read', 'Read usage analytics'),
('usage_analytics', 'update', 'Update usage analytics'),
('usage_analytics', 'delete', 'Delete usage analytics'),
-- Agent Config User Selection (for agent self-optimization)
('agent_config_user_selection', 'create', 'Create agent config selections'),
('agent_config_user_selection', 'read', 'Read agent config selections'),
('agent_config_user_selection', 'update', 'Update agent config selections'),
('agent_config_user_selection', 'delete', 'Delete agent config selections'),
-- Agent Metric (for agent self-optimization)
('agent_metric', 'create', 'Create agent metrics'),
('agent_metric', 'read', 'Read agent metrics'),
('agent_metric', 'update', 'Update agent metrics'),
('agent_metric', 'delete', 'Delete agent metrics'),
-- Agent Node Config (for agent self-optimization)
('agent_node_config', 'create', 'Create agent node configs'),
('agent_node_config', 'read', 'Read agent node configs'),
('agent_node_config', 'update', 'Update agent node configs'),
('agent_node_config', 'delete', 'Delete agent node configs')
ON CONFLICT (resource, action) DO NOTHING;

-- 4. Grant all permissions to owner roles
INSERT INTO "Role_Permission" (role_id, permission_id)
SELECT r.id, p.id
FROM "Role" r
CROSS JOIN "Permission" p
WHERE LOWER(r.name) = 'owner'
ON CONFLICT DO NOTHING;

-- 5. Grant all permissions to admin roles
INSERT INTO "Role_Permission" (role_id, permission_id)
SELECT r.id, p.id
FROM "Role" r
CROSS JOIN "Permission" p
WHERE LOWER(r.name) = 'admin'
ON CONFLICT DO NOTHING;

-- 6. Grant read permissions to member roles
INSERT INTO "Role_Permission" (role_id, permission_id)
SELECT r.id, p.id
FROM "Role" r
CROSS JOIN "Permission" p
WHERE LOWER(r.name) = 'member'
  AND p.action = 'read'
ON CONFLICT DO NOTHING;

-- 7. Grant all permissions to developer roles (global role for dev features)
INSERT INTO "Role_Permission" (role_id, permission_id)
SELECT r.id, p.id
FROM "Role" r
CROSS JOIN "Permission" p
WHERE LOWER(r.name) = 'developer'
ON CONFLICT DO NOTHING;

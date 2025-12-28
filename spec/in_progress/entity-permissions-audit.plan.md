# Entity Permissions & Approval Config Audit

> **Status**: Complete
> **Priority**: High
> **Created**: 2024-12-19
> **Updated**: 2024-12-19

## Goal

Ensure all business entities have:
1. CRUD permissions in the Permission table
2. Entry in SupportedEntities for approval config

---

## Current State

### SupportedEntities (41 entities)
```
account, account_industry, account_project, account_relationship, activity,
agent_config_user_selection, agent_metric, agent_node_config, asset, comment,
contact, contact_account, conversation, conversation_message, deal, document,
entity_asset, entity_tag, funding_round, individual, individual_relationship,
industry, job, lead, lead_project, note, opportunity, organization,
organization_relationship, organization_technology, partnership, project,
provenance, saved_report, saved_report_project, signal, status, tag, task,
technology, usage_analytics
```

### Permission table (migration 072)
All entities listed in SupportedEntities have CRUD permissions seeded in `072_permission_tables.sql`.

---

## Tables Intentionally Excluded (System/Config)

| Table | Reason |
|-------|--------|
| Agent_Config_Preset | System presets, not user-editable |
| Agent_Proposal | Managed by proposal system itself |
| Agent_Recording_Session | Debug/recording system |
| Permission, Role, Role_Permission, User_Role | RBAC system tables |
| User, User_Preferences | User management (separate concern) |
| Tenant, Tenant_Preferences | Tenant config |
| schema_migrations, schema_seeders | Database internal |
| Employee_Count_Range, Revenue_Range, Salary_Range, Funding_Stage | Lookup/enum tables |
| Research_Cache | Caching system |
| Reasoning | AI internal |
| Comment_Notification | Notification system |

---

## Action Items

### Completed
- [x] Add permissions for all 41 entities (072_permission_tables.sql)
- [x] Update SupportedEntities list (proposal_repository.go)
- [x] Consolidate migrations (removed duplicates, fixed numbering)
- [x] Run migrations

---

## Related Files

- `modules/agent-go/internal/repositories/proposal_repository.go` - SupportedEntities list
- `modules/agent/migrations/072_permission_tables.sql` - Permission inserts
- `modules/agent-go/internal/models/` - All entity models

## Migration Order (07x series)

```
070_document_summary.sql
071_system_tenant.sql
072_permission_tables.sql      # All 38 entities with CRUD permissions
073_agent_proposal_system.sql  # Proposals + system users/roles
074_add_selected_provider.sql
075_assign_admin_role.sql
```

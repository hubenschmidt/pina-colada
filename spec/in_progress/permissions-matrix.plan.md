# Complete Permissions Matrix

> **Status**: In Progress
> **Last Updated**: 2024-12-19

## Overview

- **Total DB Tables**: 61
- **Tables with Permissions**: 41 data entities
- **Total Permissions**: 164 (41 × 4 CRUD)

---

## Permissions Matrix

### Data Entities (41) - Have CRUD Permissions

| Table | Permission Resource | C | R | U | D | Notes |
|-------|---------------------|---|---|---|---|-------|
| Account | `account` | ✅ | ✅ | ✅ | ✅ | Parent entity for contacts |
| Account_Industry | `account_industry` | ✅ | ✅ | ✅ | ✅ | Join table |
| Account_Project | `account_project` | ✅ | ✅ | ✅ | ✅ | Join table |
| Account_Relationship | `account_relationship` | ✅ | ✅ | ✅ | ✅ | Account-to-account links |
| Activity | `activity` | ✅ | ✅ | ✅ | ✅ | User activity tracking |
| Agent_Config_User_Selection | `agent_config_user_selection` | ✅ | ✅ | ✅ | ✅ | Agent self-optimization |
| Agent_Metric | `agent_metric` | ✅ | ✅ | ✅ | ✅ | Agent self-optimization |
| Agent_Node_Config | `agent_node_config` | ✅ | ✅ | ✅ | ✅ | Agent self-optimization |
| Asset | `asset` | ✅ | ✅ | ✅ | ✅ | File assets |
| Comment | `comment` | ✅ | ✅ | ✅ | ✅ | Entity comments |
| Contact | `contact` | ✅ | ✅ | ✅ | ✅ | Contact methods (email, phone) |
| Contact_Account | `contact_account` | ✅ | ✅ | ✅ | ✅ | Join table with is_primary |
| Conversation | `conversation` | ✅ | ✅ | ✅ | ✅ | Chat conversations |
| Conversation_Message | `conversation_message` | ✅ | ✅ | ✅ | ✅ | Chat messages |
| Deal | `deal` | ✅ | ✅ | ✅ | ✅ | Sales deals |
| Document | `document` | ✅ | ✅ | ✅ | ✅ | Uploaded documents |
| Entity_Asset | `entity_asset` | ✅ | ✅ | ✅ | ✅ | Polymorphic asset links |
| Entity_Tag | `entity_tag` | ✅ | ✅ | ✅ | ✅ | Polymorphic tagging |
| Funding_Round | `funding_round` | ✅ | ✅ | ✅ | ✅ | Organization funding data |
| Individual | `individual` | ✅ | ✅ | ✅ | ✅ | People/persons |
| Individual_Relationship | `individual_relationship` | ✅ | ✅ | ✅ | ✅ | Person-to-person links |
| Industry | `industry` | ✅ | ✅ | ✅ | ✅ | Industry categories |
| Job | `job` | ✅ | ✅ | ✅ | ✅ | Job postings |
| Lead | `lead` | ✅ | ✅ | ✅ | ✅ | Sales leads |
| Lead_Project | `lead_project` | ✅ | ✅ | ✅ | ✅ | Join table |
| Note | `note` | ✅ | ✅ | ✅ | ✅ | Entity notes |
| Opportunity | `opportunity` | ✅ | ✅ | ✅ | ✅ | Sales opportunities |
| Organization | `organization` | ✅ | ✅ | ✅ | ✅ | Companies |
| Organization_Relationship | `organization_relationship` | ✅ | ✅ | ✅ | ✅ | Org-to-org links |
| Organization_Technology | `organization_technology` | ✅ | ✅ | ✅ | ✅ | Tech stack tracking |
| Partnership | `partnership` | ✅ | ✅ | ✅ | ✅ | Partner relationships |
| Project | `project` | ✅ | ✅ | ✅ | ✅ | Projects/campaigns |
| Data_Provenance | `provenance` | ✅ | ✅ | ✅ | ✅ | Data source tracking |
| Saved_Report | `saved_report` | ✅ | ✅ | ✅ | ✅ | Custom reports |
| Saved_Report_Project | `saved_report_project` | ✅ | ✅ | ✅ | ✅ | Join table |
| Signal | `signal` | ✅ | ✅ | ✅ | ✅ | Entity signals/alerts |
| Status | `status` | ✅ | ✅ | ✅ | ✅ | Custom statuses |
| Tag | `tag` | ✅ | ✅ | ✅ | ✅ | Tags/labels |
| Task | `task` | ✅ | ✅ | ✅ | ✅ | To-do items |
| Technology | `technology` | ✅ | ✅ | ✅ | ✅ | Technology catalog |
| Usage_Analytics | `usage_analytics` | ✅ | ✅ | ✅ | ✅ | Usage tracking |

## Excluded Tables (20)

### RBAC System Tables (4)

| Table | Reason |
|-------|--------|
| Permission | RBAC system - defines what permissions exist |
| Role | RBAC system - defines roles |
| Role_Permission | RBAC system - role-permission mappings |
| User_Role | RBAC system - user-role assignments |

### User/Tenant Management (4)

| Table | Reason |
|-------|--------|
| User | User accounts - separate auth concern |
| User_Preferences | User settings |
| Tenant | Multi-tenancy config |
| Tenant_Preferences | Tenant settings |

### Agent System Tables (3)

| Table | Reason |
|-------|--------|
| Agent_Config_Preset | System presets, not user-editable |
| Agent_Proposal | Managed by proposal system itself |
| Agent_Recording_Session | Debug/recording sessions |

### Lookup/Enum Tables (4)

| Table | Reason |
|-------|--------|
| Employee_Count_Range | Static enum values |
| Revenue_Range | Static enum values |
| Salary_Range | Static enum values |
| Funding_Stage | Static enum values |

### System Internal (5)

| Table | Reason |
|-------|--------|
| Entity_Approval_Config | Proposal system config |
| Comment_Notification | Notification system internal |
| Reasoning | AI reasoning logs |
| Research_Cache | Caching system |
| schema_migrations | Database migrations |
| schema_seeders | Database seeders |

---

## Role Assignments

### Default Role Permissions

| Role | Permissions |
|------|-------------|
| **owner** | All CRUD on all entities |
| **admin** | All CRUD on all entities |
| **developer** | All CRUD on all entities |
| **member** | Read on all entities |
| **system-agent** | All CRUD on all entities |
| **enrichment-agent** | org/individual read+update, contact read |
| **cleanup-agent** | note/contact read+delete |

---

## Migration Files

| File | Purpose |
|------|---------|
| `072_permission_tables.sql` | Creates Permission table + seeds all 167 permissions |
| `073_agent_proposal_system.sql` | Creates agent roles + assigns permissions |
| `075_assign_admin_role.sql` | Assigns admin role to specific user |

---

## SupportedEntities (for Approval Config)

Located in: `modules/agent-go/internal/repositories/proposal_repository.go`

```go
var SupportedEntities = []string{
    "account", "account_industry", "account_project", "account_relationship",
    "activity", "agent_config_user_selection", "agent_metric", "agent_node_config",
    "asset", "comment", "contact", "contact_account", "conversation",
    "conversation_message", "deal", "document", "entity_asset", "entity_tag",
    "funding_round", "individual", "individual_relationship", "industry",
    "job", "lead", "lead_project", "note", "opportunity", "organization",
    "organization_relationship", "organization_technology", "partnership",
    "project", "provenance", "saved_report", "saved_report_project",
    "signal", "status", "tag", "task", "technology", "usage_analytics",
}
```

---

## Verification Queries

### Count permissions by resource
```sql
SELECT resource, COUNT(*) as actions
FROM "Permission"
GROUP BY resource
ORDER BY resource;
```

### Check role permissions
```sql
SELECT r.name as role, COUNT(rp.permission_id) as permission_count
FROM "Role" r
LEFT JOIN "Role_Permission" rp ON r.id = rp.role_id
GROUP BY r.id, r.name
ORDER BY r.name;
```

### Find tables without permissions
```sql
SELECT tablename
FROM pg_tables
WHERE schemaname = 'public'
  AND LOWER(REPLACE(tablename, '_', '')) NOT IN (
    SELECT LOWER(REPLACE(resource, '_', '')) FROM "Permission"
  )
ORDER BY tablename;
```

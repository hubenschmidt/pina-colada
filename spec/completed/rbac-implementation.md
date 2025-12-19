# RBAC Implementation Spec

> **Status**: COMPLETED
> **Priority**: High
> **Estimated Effort**: 2-3 days
> **Code Rules**: Follow `/go-code-rules.md` and `/csr-code-rules.md`

## Goal

Implement role-based access control with granular `resource:action` permissions. Agent inherits calling user's permissions.

## Background

Currently:

- Role/UserRole tables exist but aren't fully wired
- Frontend uses roles for "developer" feature gating
- Agent tools are read-only by accident (no write tools exist), not by design
- No permission checks at tool execution level

## Database Schema

### New Tables

```sql
CREATE TABLE "Permission" (
    id BIGSERIAL PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,    -- individual, job, email, etc.
    action VARCHAR(20) NOT NULL,      -- read, create, update, delete, send, execute
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(resource, action)
);

CREATE TABLE "Role_Permission" (
    role_id BIGINT REFERENCES "Role"(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES "Permission"(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);
```

### Initial Permissions

| Resource     | Actions                      |
| ------------ | ---------------------------- |
| individual   | read, create, update, delete |
| organization | read, create, update, delete |
| contact      | read, create, update, delete |
| account      | read, create, update, delete |
| job          | read, create, update, delete |
| document     | read, create, delete         |
| email        | send                         |
| job_search   | execute                      |
| web_search   | execute                      |

## Files to Create

### 1. `internal/models/Permission.go`

```go
type Permission struct {
    ID          int64     `gorm:"primaryKey;autoIncrement"`
    Resource    string    `gorm:"not null"`
    Action      string    `gorm:"not null"`
    Description *string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 2. `internal/models/RolePermission.go`

```go
type RolePermission struct {
    RoleID       int64 `gorm:"primaryKey"`
    PermissionID int64 `gorm:"primaryKey"`
    CreatedAt    time.Time
}
```

### 3. `internal/repositories/permission_repository.go`

```go
// DTO for permission data (repositories define DTOs per CSR pattern)
type PermissionDTO struct {
    Resource string
    Action   string
}

func (r *PermissionRepository) GetUserPermissions(userID int64) ([]PermissionDTO, error)
// Returns [{Resource: "individual", Action: "read"}, ...]
// Joins User_Role -> Role_Permission -> Permission

func (r *PermissionRepository) HasPermission(userID int64, resource, action string) (bool, error)
// Returns true if user has specific permission
```

### 4. `internal/services/permission_service.go`

```go
type PermissionService struct {
    permRepo *repositories.PermissionRepository
}

func (s *PermissionService) HasPermission(userID int64, resource, action string) (bool, error)
func (s *PermissionService) GetUserPermissions(userID int64) ([]string, error)
func (s *PermissionService) CanAccess(ctx context.Context, permission string) bool
```

## Files to Modify

### 5. `internal/middleware/auth.go`

- Add `PermissionsKey` context key
- Add `PermissionLoaderMiddleware` that loads user permissions into context after UserLoader

### 6. `internal/agent/tools/crm_tools.go`

- Add `permChecker PermissionChecker` field to CRMTools struct
- Check permission before each operation:

```go
func (t *CRMTools) LookupCtx(ctx context.Context, params CRMLookupParams) (*CRMLookupResult, error) {
    requiredPerm := params.EntityType + ":read"
    if !t.permChecker.CanAccess(ctx, requiredPerm) {
        return &CRMLookupResult{Results: "Permission denied: " + requiredPerm}, nil
    }
    // ... existing logic
}
```

### 7. `internal/agent/tools/adapter.go`

- Update `BuildAgentTools` to accept and pass `permChecker` to tool constructors

### 8. `internal/agent/orchestrator.go`

- Create PermissionService instance
- Pass to tool constructors

## Tool Permission Matrix

| Tool                      | Permission Required |
| ------------------------- | ------------------- |
| crm_lookup (individual)   | individual:read     |
| crm_lookup (organization) | organization:read   |
| crm_lookup (contact)      | contact:read        |
| crm_lookup (account)      | account:read        |
| crm_lookup (job/lead)     | job:read            |
| crm_list                  | {entity}:read       |
| crm_statuses              | job:read            |
| job_search                | job_search:execute  |
| web_search                | web_search:execute  |
| search_entity_documents   | document:read       |
| read_document             | document:read       |
| send_email                | email:send          |

## Default Role Permissions

### Owner (per-tenant)

All permissions

### Admin (per-tenant)

All permissions except dangerous operations

### Member (per-tenant)

- All read permissions
- email:send
- job_search:execute
- web_search:execute

### Developer (global)

- All permissions (for dev features access)

## Implementation Order

1. Create models (Permission, RolePermission)
2. Create repository (permission_repository.go)
3. Create service (permission_service.go)
4. Add middleware (PermissionLoaderMiddleware)
5. Add permission checks to tools
6. Update orchestrator wiring
7. Database migration + seed data

## Migration Script

```sql
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

-- 2. Create Role_Permission junction
CREATE TABLE IF NOT EXISTS "Role_Permission" (
    role_id BIGINT REFERENCES "Role"(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES "Permission"(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- 3. Seed permissions
INSERT INTO "Permission" (resource, action, description) VALUES
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
('document', 'delete', 'Delete documents'),
('email', 'send', 'Send emails'),
('job_search', 'execute', 'Search for jobs on web'),
('web_search', 'execute', 'Search web for information')
ON CONFLICT (resource, action) DO NOTHING;

-- 4. Grant all permissions to Owner roles
INSERT INTO "Role_Permission" (role_id, permission_id)
SELECT r.id, p.id
FROM "Role" r
CROSS JOIN "Permission" p
WHERE r.name = 'Owner'
ON CONFLICT DO NOTHING;

-- 5. Grant read + execute permissions to Member roles
INSERT INTO "Role_Permission" (role_id, permission_id)
SELECT r.id, p.id
FROM "Role" r
CROSS JOIN "Permission" p
WHERE r.name = 'Member'
  AND (p.action = 'read' OR p.action = 'execute' OR (p.resource = 'email' AND p.action = 'send'))
ON CONFLICT DO NOTHING;
```

## Future Considerations

- Admin UI for role/permission management
- Per-tenant custom roles
- Attribute-based access control (ABAC) for row-level permissions
- Audit logging of permission checks

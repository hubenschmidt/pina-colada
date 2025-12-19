# Frontend Admin Screens -- follow ~/commands/js-code-rules.md

> **Status**: TODO
> **Priority**: Medium
> **Location**: Under /settings as tabs
> **UI Library**: Mantine v7, lucide-react icons, Tailwind CSS
> **Depends on**: Agent Proposal System backend (mostly complete), RBAC CRUD endpoints (need to add)

## Goal

Add admin screens for:

1. **Proposal Queue** - Review/approve/reject agent-proposed CUD operations
2. **Approval Config** - Toggle which entities require approval
3. **RBAC Admin** - Full role and permission management

## File Structure

```
modules/client/
  app/settings/page.jsx              # Modify: Add Tabs component

  components/
    ProposalQueue/
      ProposalQueue.jsx              # List with bulk actions
      ProposalDetailModal.jsx        # Modal with JSON editor
      hooks/useProposalQueueConfig.jsx

    ApprovalConfig/
      ApprovalConfigPanel.jsx        # Toggle switches per entity

    RbacAdmin/
      RbacAdmin.jsx                  # Main RBAC container
      RoleList.jsx                   # Roles CRUD
      RoleFormModal.jsx              # Create/Edit role
      PermissionMatrix.jsx           # Role x Permission grid
      UserRoleAssignment.jsx         # Assign roles to users
      hooks/useRbacConfig.jsx

    JsonEditor/
      JsonEditor.jsx                 # Reusable JSON editor
```

## API Functions (`/modules/client/api/index.js`)

```javascript
// Proposals (backend exists)
export const getProposals = (page, limit) =>
  apiGet(`/proposals?page=${page}&limit=${limit}`);
export const approveProposal = (id) => apiPost(`/proposals/${id}/approve`);
export const rejectProposal = (id) => apiPost(`/proposals/${id}/reject`);
export const bulkApproveProposals = (ids) =>
  apiPost("/proposals/bulk-approve", { ids });
export const bulkRejectProposals = (ids) =>
  apiPost("/proposals/bulk-reject", { ids });
export const updateProposalPayload = (id, payload) =>
  apiPut(`/proposals/${id}/payload`, { payload });

// Approval Config (backend exists)
export const getApprovalConfig = () => apiGet("/admin/approval-config");
export const updateApprovalConfig = (entityType, requiresApproval) =>
  apiPut(`/admin/approval-config/${entityType}`, {
    requires_approval: requiresApproval,
  });

// RBAC (backend endpoints need to be added)
export const getRoles = () => apiGet("/admin/roles");
export const createRole = (data) => apiPost("/admin/roles", data);
export const updateRole = (id, data) => apiPut(`/admin/roles/${id}`, data);
export const deleteRole = (id) => apiDelete(`/admin/roles/${id}`);
export const getPermissions = () => apiGet("/admin/permissions");
export const updateRolePermissions = (roleId, permissionIds) =>
  apiPut(`/admin/roles/${roleId}/permissions`, {
    permission_ids: permissionIds,
  });
export const updateUserRole = (userId, roleId) =>
  apiPut(`/admin/users/${userId}/role`, { role_id: roleId });
```

## Components

### 1. Settings Page - Add Tabs

```jsx
<Tabs defaultValue="general">
  <Tabs.List>
    <Tabs.Tab value="general" leftSection={<Settings size={16} />}>
      General
    </Tabs.Tab>
    <Tabs.Tab value="proposals" leftSection={<ListChecks size={16} />}>
      Proposal Queue
    </Tabs.Tab>
    <Tabs.Tab value="access" leftSection={<Shield size={16} />}>
      Access Control
    </Tabs.Tab>
  </Tabs.List>
  <Tabs.Panel value="general">{/* existing settings */}</Tabs.Panel>
  <Tabs.Panel value="proposals">
    <ProposalQueue />
    <ApprovalConfigPanel />
  </Tabs.Panel>
  <Tabs.Panel value="access">
    <RbacAdmin />
  </Tabs.Panel>
</Tabs>
```

### 2. ProposalQueue

- DataTable with checkbox column (new pattern for codebase)
- Columns: entity_type, operation, created_at, validation_errors badge
- Bulk action bar when items selected (Approve Selected, Reject Selected)
- Row click opens ProposalDetailModal

### 3. ProposalDetailModal

- JSON editor for payload (editable)
- Validation errors display (Alert with List)
- Save Changes button (calls PUT /proposals/{id}/payload)
- Approve/Reject buttons

### 4. ApprovalConfigPanel

- Switch toggle per entity type
- Updates immediately on toggle

### 5. PermissionMatrix

- Table grid: roles as columns, permissions as rows
- Checkbox per cell
- Groups permissions by resource (individual, organization, contact, etc.)

### 6. UserRoleAssignment

- List of tenant users
- Select dropdown for role assignment per user

## Dependencies

```bash
npm install @uiw/react-json-view
# OR for full IDE experience:
npm install @monaco-editor/react
```

## Backend: RBAC CRUD Endpoints Needed

Add to `internal/routes/router.go`:

```go
r.Route("/admin/roles", func(r chi.Router) {
    r.Get("/", c.Role.List)
    r.Post("/", c.Role.Create)
    r.Put("/{id}", c.Role.Update)
    r.Delete("/{id}", c.Role.Delete)
    r.Get("/{id}/permissions", c.Role.GetPermissions)
    r.Put("/{id}/permissions", c.Role.UpdatePermissions)
})
r.Put("/admin/users/{id}/role", c.User.UpdateRole)
```

Requires:

- `internal/controllers/role_controller.go`
- `internal/services/role_service.go`
- `internal/repositories/role_repository.go` (may exist, check for CRUD methods)

## Implementation Order

1. API functions in client
2. JsonEditor component (install dependency)
3. ProposalQueue + ProposalDetailModal
4. ApprovalConfigPanel
5. Backend RBAC CRUD endpoints
6. RBAC Admin components

## Related Specs

- `spec/todo/rbac-implementation.md` - Backend RBAC (complete)
- `spec/todo/field-level-diff-review.md` - Future enhancement for visual diffs

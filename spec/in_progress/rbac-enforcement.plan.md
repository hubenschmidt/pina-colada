# RBAC Enforcement Spec

> **Status**: Todo
> **Priority**: High
> **Created**: 2024-12-19

## Goal

Ensure the AI agent operates within RBAC constraints to prevent unauthorized data modifications.

---

## Current State

### What Exists
- Permission table with 44 resources, 167 permissions
- Role_Permission assignments (admin, member, owner, developer, system-agent roles)
- `PermissionService.HasPermission(userID, resource, action)` - checks permission
- `PermissionService.CanAccess(ctx, "resource:action")` - context-based check
- `PermissionLoaderMiddleware` - loads permissions into request context

### What's Missing
- **No enforcement** - permissions are checkable but not enforced
- Agent can currently bypass RBAC when executing operations

---

## Phase 1: Agent-Side RBAC Enforcement (Priority)

The AI agent must check permissions before proposing or executing CUD operations.

### Files to Modify

#### `internal/services/proposal_service.go`
Before creating a proposal, verify the agent user has permission:
```go
func (s *ProposalService) ProposeOperation(...) {
    // Check agent has permission for this operation
    hasPermission, err := s.permService.HasPermission(userID, entityType, operation)
    if !hasPermission {
        return ErrPermissionDenied
    }
    // ... create proposal
}
```

#### `internal/services/proposal_executor.go` (or wherever proposals are executed)
Before executing an approved proposal, verify permission still valid:
```go
func (s *ProposalService) ExecuteProposal(...) {
    // Re-check permission at execution time
    hasPermission, err := s.permService.HasPermission(proposal.ProposedByID, proposal.EntityType, proposal.Operation)
    if !hasPermission {
        return s.MarkFailed(proposalID, "permission denied")
    }
    // ... execute
}
```

### Agent Tools Integration

#### `internal/agent/tools/crm_tools.go`
CRM tools should check permissions before proposing changes:
```go
func (t *CRMTools) ProposeUpdate(ctx context.Context, params ProposeParams) error {
    userID := middleware.GetUserID(ctx)

    // Permission check
    if !t.permService.CanAccess(ctx, params.EntityType+":update") {
        return ErrPermissionDenied
    }

    // Approval check
    requiresApproval, _ := t.approvalService.RequiresApproval(tenantID, params.EntityType)
    if requiresApproval {
        return t.proposalService.ProposeOperation(...)
    }

    // Direct execution if no approval needed
    return t.executeUpdate(...)
}
```

---

## Phase 2: RequirePermission Middleware (Future)

For API endpoint protection beyond JWT auth.

### Create Middleware

#### `internal/middleware/permission.go`
```go
func RequirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            perms, ok := GetPermissions(r.Context())
            if !ok {
                writeError(w, http.StatusForbidden, "no permissions loaded")
                return
            }

            if !contains(perms, permission) {
                writeError(w, http.StatusForbidden, "permission denied")
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Apply to Routes (Example)

```go
// Admin routes - require admin role
r.Route("/admin", func(r chi.Router) {
    r.Use(RequireRole("admin"))  // or check specific permissions
    r.Route("/approval-config", ...)
    r.Route("/roles", ...)
})

// Entity routes - require entity permissions
r.Route("/organizations", func(r chi.Router) {
    r.With(RequirePermission("organization:read")).Get("/", ...)
    r.With(RequirePermission("organization:create")).Post("/", ...)
})
```

---

## Phase 3: Tenant Isolation (Future)

Ensure users can only access data within their tenant.

### Already Implemented
- `TenantIDKey` in context from JWT claims or header
- Most queries filter by tenant_id

### To Verify
- All repository methods include tenant_id filter
- Cross-tenant access is impossible even with valid permissions

---

## Implementation Order

1. **Phase 1A**: Add permission check to `ProposeOperation`
2. **Phase 1B**: Add permission check to proposal execution
3. **Phase 1C**: Add permission checks to CRM tools
4. **Phase 2**: Create `RequirePermission` middleware (when needed)
5. **Phase 3**: Audit tenant isolation

---

## Test Cases

### Agent Permission Tests
- [ ] Agent without `organization:update` cannot propose org update
- [ ] Agent without `individual:delete` cannot propose individual deletion
- [ ] Proposal execution fails if permission revoked after approval

### Role-Based Tests
- [ ] Member role can read but not delete
- [ ] Admin role has full access
- [ ] System-agent role has appropriate permissions

---

## Related Files

- `internal/services/permission_service.go` - Permission checking
- `internal/services/proposal_service.go` - Proposal creation/execution
- `internal/middleware/auth.go` - Context helpers
- `internal/agent/tools/crm_tools.go` - Agent CRM operations
- `internal/routes/router.go` - Route definitions

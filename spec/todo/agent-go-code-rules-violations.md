# Agent-Go Code Rules Violations Report

**Date:** 2025-12-12
**Scope:** `modules/agent-go/internal/`
**Rules:** go-code-rules, csr-code-rules

---

## Summary

| Category | Status | Count |
|----------|--------|-------|
| CSR Import Violations | FAIL | 3 |
| Guard Clause Violations | FAIL | 15+ |
| Import Grouping Issues | FAIL | 2 |
| Large Functions | WARN | 7+ |
| Naked Returns | PASS | 0 |
| Panic Usage | PASS | 0 |

---

## CSR Import Violations (3)

### 1. Services Importing Models

**Rule:** Services → repositories, serializers only (NEVER models)

#### `internal/services/report_service.go`
**Line 7:**
```go
"github.com/pina-colada-co/agent-go/internal/models"
```
**Fix:** Remove models import; use repository DTOs or serializer types.

#### `internal/services/project_service.go`
**Line 7:**
```go
"github.com/pina-colada-co/agent-go/internal/models"
```
**Fix:** Remove models import; use repository DTOs or serializer types.

### 2. Serializers Importing Repositories

**Rule:** Serializers → models only (NEVER repositories)

#### `internal/serializers/lead.go`
**Line 6:**
```go
"github.com/pina-colada-co/agent-go/internal/repositories"
```
**Usage:** `OpportunityToResponse(*repositories.OpportunityWithLead)`, `PartnershipToResponse(*repositories.PartnershipWithLead)`

**Fix:** Move `OpportunityWithLead` and `PartnershipWithLead` structs to models or schemas.

---

## Go Code Rules Violations

### Guard Clause Violations (15+)

**Rule:** Use guard clauses and early returns; avoid nested conditionals, `else`.

#### `internal/controllers/organization_controller.go`

Repeated nested if/else pattern (~12 instances):

**Lines 56-63, 108-113, 137-143, 168-174, 213-219, 250-256, 277-283, 301-306, 327-332, 373-378, 401-407, 460-465**
```go
result, err := c.orgService.GetOrganization(id)
if err != nil {
    if err.Error() == "organization not found" {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }
    writeError(w, http.StatusInternalServerError, err.Error())
    return
}
```

**Fix:** Use sentinel errors with early return:
```go
result, err := c.orgService.GetOrganization(id)
if errors.Is(err, services.ErrOrganizationNotFound) {
    writeError(w, http.StatusNotFound, err.Error())
    return
}
if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
}
```

#### `internal/controllers/individual_controller.go`

**Lines 56-63** (and similar patterns):
```go
if err != nil {
    if err.Error() == "individual not found" {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }
    writeError(w, http.StatusInternalServerError, err.Error())
    return
}
```

---

### Import Grouping Issues (2)

**Rule:** Imports grouped as: stdlib, external, internal (with blank lines between).

#### `internal/services/organization_service.go`
**Lines 3-10:**
```go
import (
    "errors"
    "time"

    "github.com/pina-colada-co/agent-go/internal/repositories"
    "github.com/pina-colada-co/agent-go/internal/schemas"
    "github.com/pina-colada-co/agent-go/internal/serializers"
    "github.com/shopspring/decimal"
)
```
**Issue:** External package `shopspring/decimal` mixed with internal packages.

**Fix:**
```go
import (
    "errors"
    "time"

    "github.com/shopspring/decimal"

    "github.com/pina-colada-co/agent-go/internal/repositories"
    "github.com/pina-colada-co/agent-go/internal/schemas"
    "github.com/pina-colada-co/agent-go/internal/serializers"
)
```

#### `internal/services/individual_service.go`
Same issue - `shopspring/decimal` not separated from internal imports.

---

### Large Functions (7+)

**Rule:** Prefer small, focused functions over large functions with many branches.

| File | Lines | Notes |
|------|-------|-------|
| `controllers/organization_controller.go` | 523 | Repeated error handling patterns |
| `repositories/report_repository.go` | 691 | Complex query building |
| `services/organization_service.go:455-498` | 43 | CreateOrganization - multiple nesting levels |

---

## Passing Checks

| Check | Status |
|-------|--------|
| Naked returns | PASS - None found |
| Panic for error handling | PASS - None found |
| Controllers importing repos/models | PASS |
| Repositories importing only models/gorm | PASS |

---

## Recommendations

### Priority 1: CSR Violations
1. Remove `models` import from `report_service.go` and `project_service.go`
2. Move repository DTOs in `serializers/lead.go` to models or schemas

### Priority 2: Guard Clauses
1. Define sentinel errors in services (e.g., `ErrOrganizationNotFound`)
2. Replace string comparison with `errors.Is()` checks
3. Flatten nested if/else to early returns

### Priority 3: Import Grouping
1. Run `goimports` with proper grouping configuration
2. Separate external packages from internal packages

---

## Files Reviewed

- `internal/controllers/*.go`
- `internal/services/*.go`
- `internal/repositories/*.go`
- `internal/serializers/*.go`
- `internal/schemas/*.go`

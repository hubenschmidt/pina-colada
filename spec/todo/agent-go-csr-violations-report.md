# Agent-Go CSR Pattern Violations Report

**Date:** 2025-12-08
**Scope:** `modules/agent-go/internal/services/`
**Status:** RESOLVED

---

## CSR Import Rules (Reference)

```
routes      → controllers, middleware
controllers → services, serializers, schemas
services    → repositories, serializers (NEVER models)
repositories → models
schemas     → (nothing domain-specific)
serializers → models
```

**Key principle:** Services never import models directly.

---

## Layer Summary

| Layer | Status | Notes |
|-------|--------|-------|
| Services | **PASS** | All 18 files compliant |

---

## Resolved Violations

The following 6 files were refactored to remove model imports:

| File | Fix Applied |
|------|-------------|
| `lookup_service.go` | Created `serializers/lookup.go` with response types for Industry, EmployeeCountRange, RevenueRange, FundingStage, SalaryRange |
| `note_service.go` | Created `serializers/note.go` with NoteResponse |
| `provenance_service.go` | Created `serializers/provenance.go` with ProvenanceResponse |
| `technology_service.go` | Created `serializers/technology.go` with TechnologyResponse |
| `task_service.go` | Changed `resolveEntityInfo()` to accept `*string, *int64` instead of `*models.Task` |
| `account_service.go` | Moved `getAccountTypeAndEntityID()` to repository, service now uses repository DTO |

---

## New Serializers Created

- `internal/serializers/lookup.go` - IndustryResponse, EmployeeCountRangeResponse, RevenueRangeResponse, FundingStageResponse, SalaryRangeResponse
- `internal/serializers/note.go` - NoteResponse
- `internal/serializers/provenance.go` - ProvenanceResponse
- `internal/serializers/technology.go` - TechnologyResponse

---

## Compliant Files (18)

| File | Status |
|------|--------|
| `account_service.go` | PASS |
| `auth_service.go` | PASS |
| `comment_service.go` | PASS |
| `contact_service.go` | PASS |
| `conversation_service.go` | PASS |
| `costs_service.go` | PASS |
| `document_service.go` | PASS |
| `individual_service.go` | PASS |
| `job_service.go` | PASS |
| `lead_service.go` | PASS |
| `lookup_service.go` | PASS |
| `note_service.go` | PASS |
| `organization_service.go` | PASS |
| `preferences_service.go` | PASS |
| `project_service.go` | PASS |
| `provenance_service.go` | PASS |
| `task_service.go` | PASS |
| `technology_service.go` | PASS |

---

## Metrics

- **Total services:** 18
- **Compliant services:** 18
- **Non-compliant services:** 0
- **Compliance rate:** 100%

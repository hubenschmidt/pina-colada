# Agent CSR Pattern Violations Report

**Date:** 2025-12-08
**Scope:** `modules/agent/src/`

---

## CSR Import Rules (Reference)

```
routes      → controllers, middleware
controllers → services, serializers, schemas
services    → repositories, serializers
repositories → models
schemas     → (nothing domain-specific)
serializers → models
```

**Key principle:** Services never import models directly.

---

## Layer Summary

| Layer | Status | Notes |
|-------|--------|-------|
| Routes | **PASS** | Only imports controllers |
| Controllers | **PASS** | Imports services, serializers, schemas |
| Services | **FAIL** | Imports models directly, uses session |
| Repositories | **PASS** | Imports models (correct) |
| Schemas | **PASS** | Only pydantic, stdlib |
| Serializers | **FAIL** | 1 file imports from services |

---

## Violations Detail

### Services Layer (CRITICAL)

**Rule violated:** Services should only import from repositories and serializers.

#### Model Imports in Services (48 violations across 14 files)

| File | Models Imported |
|------|-----------------|
| `report_builder.py` | Base, Organization, Individual, Contact, Lead, Account, Note, LeadProject, Job, Opportunity, Partnership, Project, Task, Asset, Document, Deal, User |
| `individual_service.py` | Account, AccountIndustry, AccountProject, IndividualRelationship |
| `organization_service.py` | Account, AccountIndustry, AccountProject, OrganizationRelationship |
| `preferences_service.py` | User, UserPreferences, TenantPreferences, Tenant, UserRole |
| `project_service.py` | Project, Deal, Lead, LeadProject |
| `contact_service.py` | Contact, ContactAccount |
| `account_service.py` | Account, AccountRelationship |
| `user_service.py` | Individual, Project, Tenant |
| `notification_service.py` | Comment |
| `tag_service.py` | Tag |
| `reasoning_service.py` | Reasoning |
| `industry_service.py` | Industry |
| `auth_service.py` | User |

#### Direct Session Usage in Services (100+ violations)

Services should call repository methods, not use session directly.

| File | Session Operations |
|------|-------------------|
| `report_builder.py` | 16 `session.execute()` calls |
| `individual_service.py` | 20 session operations |
| `organization_service.py` | 12+ session operations |
| `project_service.py` | 14 session operations |
| `contact_service.py` | 11 session operations |
| `account_service.py` | 9 session operations |
| `notification_service.py` | 9 session operations |
| `user_service.py` | 4 session operations |
| `industry_service.py` | 3 session operations |
| `reasoning_service.py` | 2 session operations |
| `tag_service.py` | 1 session operation |

### Serializers Layer (LOW)

**Rule violated:** Serializers should only import from models.

| File | Violation |
|------|-----------|
| `notification.py` | `from services.notification_service import get_entity_url, get_entity_project_id` |

---

## Compliant Layers

### Routes (PASS)
All 29 route files import only from controllers.

### Controllers (PASS)
All controller files follow correct pattern:
- Import from services
- Import from serializers
- Import from schemas
- No model imports
- No session usage

### Repositories (PASS)
All repository files correctly import from models.

### Schemas (PASS)
All schema files only use pydantic and stdlib.

---

## Remediation Priority

### P0 - Critical (services with heavy model/session usage)

| Service | Effort | Recommendation |
|---------|--------|----------------|
| `report_builder.py` | High | Create `report_repository.py` for all queries |
| `individual_service.py` | High | Move session ops to `individual_repository.py` |
| `organization_service.py` | High | Move session ops to `organization_repository.py` |

### P1 - High

| Service | Effort | Recommendation |
|---------|--------|----------------|
| `project_service.py` | Medium | Create/extend `project_repository.py` |
| `contact_service.py` | Medium | Move session ops to `contact_repository.py` |
| `preferences_service.py` | Medium | Create `preferences_repository.py` |

### P2 - Medium

| Service | Effort | Recommendation |
|---------|--------|----------------|
| `account_service.py` | Low | Create `account_repository.py` |
| `notification_service.py` | Low | Extend `notification_repository.py` |
| `user_service.py` | Low | Extend `user_repository.py` |

### P3 - Low

| Service | Effort | Recommendation |
|---------|--------|----------------|
| `industry_service.py` | Trivial | Move to `industry_repository.py` |
| `tag_service.py` | Trivial | Move to `tag_repository.py` |
| `reasoning_service.py` | Trivial | Create `reasoning_repository.py` |
| `auth_service.py` | Trivial | Move to `auth_repository.py` |

### P4 - Serializer Fix

| File | Fix |
|------|-----|
| `serializers/notification.py` | Move `get_entity_url`, `get_entity_project_id` to serializer or pass as params |

---

## Refactor Pattern

For each violating service:

1. **Create/extend repository** with methods for each `session.execute()` call
2. **Define DTOs** in repository for complex return types
3. **Define input structs** in repository for create/update operations
4. **Remove model imports** from service
5. **Replace session calls** with repository method calls

Example transformation:

```python
# BEFORE (service)
from models.Project import Project
result = await session.execute(select(Project).where(...))

# AFTER (service)
from repositories.project_repository import project_repo
result = await project_repo.find_by_tenant(tenant_id)
```

---

## Metrics

- **Total services:** 14
- **Compliant services:** 0
- **Non-compliant services:** 14
- **Estimated remediation:** 14 files to refactor

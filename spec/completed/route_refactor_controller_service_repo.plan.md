# Route Refactor: Controller/Service/Repository Pattern

## Problem
Many route files contain business logic that should be in service/controller layers. This violates separation of concerns and makes code harder to test and maintain.

## Target Architecture

```
Route (routes/*.py)
    ↓ calls
Controller (controllers/*.py)
    ↓ calls
Service (services/*.py)
    ↓ calls
Repository (repositories/*.py)
    ↓ queries
Database
```

### Layer Responsibilities

**Route Layer** (`api/routes/*.py`)
- Minimal code - just wiring
- Decorators: `@router.get()`, `@require_auth`, `@log_errors`
- Extract request params (tenant_id, user_id, query params)
- Call controller functions
- Return results directly

**Controller Layer** (`controllers/*.py`)
- Request/response formatting
- Transform ORM models to response dicts (`_to_list_response`, `_to_detail_response`)
- Pagination logic (`_to_paged_response`)
- `@handle_http_exceptions` decorator
- Calls service layer

**Service Layer** (`services/*.py`)
- Business logic and validation
- Orchestration of multiple repository calls
- Raise `HTTPException` for business rule violations
- Only imports from repositories (never routes/controllers)

**Repository Layer** (`repositories/*.py`)
- Direct database operations (CRUD)
- Query construction with SQLAlchemy
- No business logic
- No HTTP exceptions

## Current State

### Following Pattern (reference implementations)
| Route | Controller | Service | Repository |
|-------|------------|---------|------------|
| leads.py (1.2KB) | job_controller.py | job_service.py | job_repository.py |
| opportunities.py (3.4KB) | opportunity_controller.py | opportunity_service.py | - |
| partnerships.py (3.8KB) | partnership_controller.py | partnership_service.py | - |

### Needs Refactor (business logic in routes)
| Route | Size | Missing Layers |
|-------|------|----------------|
| organizations.py | 29KB | organization_controller, organization_service |
| individuals.py | 23KB | individual_controller, individual_service |
| documents.py | 15KB | document_controller, document_service |
| contacts.py | 10KB | contact_controller, contact_service |
| projects.py | 9KB | project_controller, project_service |
| notifications.py | 8KB | notification_controller (service exists) |
| preferences.py | 7KB | preferences_controller, preferences_service |
| comments.py | 5KB | comment_controller, comment_service |

## Refactor Order (by impact/size)
1. **organizations.py** - Largest, most complex, reference implementation
2. **individuals.py** - Second largest
3. **documents.py** - Medium complexity
4. **contacts.py** - Medium complexity
5. **projects.py** - Medium complexity
6. **notifications.py** - Service exists, just needs controller
7. **preferences.py** - Lower priority
8. **comments.py** - Lower priority

## Implementation Steps (per route)

1. Create `controllers/{entity}_controller.py`
   - Move `_to_dict`, `_to_list_dict` functions from route
   - Add `_to_paged_response` helper
   - Create controller functions with `@handle_http_exceptions`

2. Create `services/{entity}_service.py` (if not exists)
   - Move business logic from route
   - Move validation logic
   - Import only from repositories

3. Refactor `routes/{entity}.py`
   - Remove all `_to_dict` functions
   - Remove business logic
   - Import from controller only
   - Each route handler should be 1-3 lines

4. Update imports
   - Routes import from controllers
   - Controllers import from services
   - Services import from repositories

## Example: Minimal Route Handler

```python
@router.get("")
@log_errors
@require_auth
async def get_organizations_route(request: Request, ...params):
    tenant_id = request.state.tenant_id
    return await get_organizations(tenant_id, page, limit, order_by, order, search)
```

## Status

### Completed
- [x] organizations.py - DONE (29KB → 11KB, ~60% reduction)
- [x] individuals.py - DONE (632 → 223 lines, ~65% reduction)
- [x] documents.py - DONE (456 → 197 lines, ~57% reduction)
- [x] tasks.py - DONE (334 → 172 lines, ~48% reduction)
- [x] contacts.py - DONE (302 → 120 lines, ~60% reduction)
- [x] projects.py - DONE (277 → 105 lines, ~62% reduction)
- [x] notes.py - DONE (144 → 75 lines, ~48% reduction)

### High Priority (large files with significant DB logic)
- [ ] reports.py - 354 lines, repo imports

### Medium Priority
- [ ] notifications.py - 227 lines, 5 helper funcs
- [ ] preferences.py - 225 lines, 5 session usages
- [ ] comments.py - 161 lines, 2 helper funcs

### Low Priority
- [ ] provenance.py - 80 lines, 1 helper func
- [ ] users.py - 56 lines, 2 session usages
- [ ] accounts.py - 63 lines, 2 session usages, 2 helper funcs
- [ ] industries.py - 49 lines, 2 session usages
- [ ] technologies.py - 54 lines, repo imports
- [ ] tags.py - 24 lines, 2 session usages

### Skip (simple lookups, OK as-is)
- employee_count_ranges.py - 18 lines
- funding_stages.py - 18 lines
- revenue_ranges.py - 18 lines
- salary_ranges.py - 18 lines

### Already Following Pattern
- auth.py - uses controller
- jobs.py - uses controller
- leads.py - uses controller
- opportunities.py - uses controller
- partnerships.py - uses controller

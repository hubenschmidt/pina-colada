# Python Code Rules Audit Report

**Date**: 2025-11-30  
**Scope**: `modules/agent/src/`  
**Rules Source**: `python-code-rules.md`

---

## Executive Summary

This audit reviewed Python code in `modules/agent/src/` against the established code rules. The codebase generally follows good practices but has several areas for improvement, particularly around guard clauses, import placement, and avoiding `break`/`continue` statements.

**Overall Compliance**: ⚠️ **Moderate** - Most patterns are followed, but several violations exist.

---

## Rule-by-Rule Analysis

### ✅ 1. KISS (Keep It Simple, Stupid)

**Status**: ✅ **Mostly Compliant**

Most code follows KISS principles. Functions are reasonably sized and focused. Some exceptions:

- `modules/agent/src/services/report_builder.py` - Complex report building logic (acceptable for domain complexity)
- `modules/agent/src/services/job_service.py` - Some functions handle multiple concerns (e.g., `create_job`, `update_job`)

**Recommendations**:

- Consider splitting complex service functions into smaller, focused functions
- Extract validation logic into separate functions

---

### ✅ 2. YAGNI (You Ain't Gonna Need It)

**Status**: ✅ **Compliant**

No obvious speculative features found. Code appears focused on current requirements.

---

### ✅ 3. SOLID Principles (Functional Programming)

**Status**: ✅ **Mostly Compliant**

- **Single Responsibility**: ✅ Functions generally have single responsibilities
- **Open-Closed**: ✅ Extension points exist (e.g., worker/evaluator patterns)
- **Dependency Inversion**: ✅ Repository pattern abstracts data access

**Note**: Models use OOP (SQLAlchemy), which is acceptable and unavoidable.

---

### ✅ 4. Controller-Service-Repository Pattern

**Status**: ✅ **Compliant**

The architecture correctly follows the pattern:

- **Controllers**: `modules/agent/src/controllers/` - Handle HTTP concerns, format responses
- **Services**: `modules/agent/src/services/` - Business logic
- **Repositories**: `modules/agent/src/repositories/` - Data access

**Example**: `modules/agent/src/controllers/job_controller.py` → `modules/agent/src/services/job_service.py` → `modules/agent/src/repositories/job_repository.py`

---

### ⚠️ 5. Functional Programming Preference

**Status**: ✅ **Compliant**

Codebase is primarily functional. OOP usage is limited to:

- SQLAlchemy models (unavoidable)
- Some utility classes (e.g., `_WebSocketStreamState` in `modules/agent/src/agent/graph.py`)

**Recommendation**: Consider refactoring `_WebSocketStreamState` to use a dictionary/namespace pattern if possible.

---

### ❌ 6. GUARD CLAUSE RULE

**Status**: ❌ **Multiple Violations**

**Rule**: Use guard clauses; avoid nested conditionals, `else`, and `switch/case`.

#### Violations Found:

1. **`modules/agent/src/api/routes/organizations.py`** (Lines 136-139, 204-253)

   ```python
   # Nested if statements
   if not first_name and contact.individuals:
       first_name = contact.individuals[0].first_name
   if not last_name and contact.individuals:
       last_name = contact.individuals[0].last_name
   ```

2. **`modules/agent/src/api/routes/notifications.py`** (Lines 30-104)

   ```python
   # Multiple nested if/else blocks
   if entity_type == "Deal":
       # ...
   if entity_type == "Task":
       # ...
   if entity_type == "Lead":
       # ...
   if entity_type in ("Individual", "Organization"):
       # ...
   else:
       return None
   ```

3. **`modules/agent/src/repositories/saved_report_repository.py`** (Lines 56-59)

   ```python
   if include_global:
       base_stmt = base_stmt.where(or_(has_project, is_global))
   else:
       base_stmt = base_stmt.where(has_project)
   ```

4. **`modules/agent/src/services/job_service.py`** (Lines 131-141, 474-483)

   ```python
   if organizations:
       return organizations[0].get("name", "Unknown Company"), "Organization", None
   if individuals:
       # ...
   return "Unknown Company", "Organization", None
   ```

5. **`modules/agent/src/api/routes/organizations.py`** (Lines 365-393)
   ```python
   # Nested conditionals
   if industry_ids is not None and org.account_id:
       # ...
   if "project_ids" in data.model_dump(exclude_unset=True) and org.account_id:
       # ...
   ```

**Additional files with guard clause violations**:

- `modules/agent/src/api/routes/individuals.py` - Multiple nested conditionals
- `modules/agent/src/services/partnership_service.py` - Lines 31-34, 89
- `modules/agent/src/services/opportunity_service.py` - Lines 31-34, 89
- `modules/agent/src/repositories/task_repository.py` - Line 85
- `modules/agent/src/repositories/partnership_repository.py` - Line 87
- `modules/agent/src/repositories/opportunity_repository.py` - Line 87
- `modules/agent/src/repositories/job_repository.py` - Line 134

**Recommendations**:

- Refactor nested conditionals to use early returns
- Replace `else` clauses with guard clauses
- Use dictionary dispatch for type-based conditionals (e.g., `_get_entity_project_id`)

---

### ❌ 7. No `break`/`continue` Statements

**Status**: ❌ **Violation Found**

**Rule**: Don't use `break`/`continue`; refactor with small pure functions, early returns.

#### Violations Found:

1. **`services/notification_service.py`** (Line 53)
   ```python
   for user_id in participants:
       if user_id in notified_users:
           continue  # ❌ Should use filter or early return
   ```

**Recommendation**: Refactor to use `filter()` or list comprehension:

```python
unnotified_participants = [uid for uid in participants if uid not in notified_users]
for user_id in unnotified_participants:
    # ...
```

---

### ❌ 8. Top-Level Imports Only

**Status**: ❌ **Multiple Violations**

**Rule**: Imports should be defined at the top level of all files. Do not use local imports inside functions.

#### Violations Found: **76 instances** across multiple files

**Most Common Locations**:

1. **`modules/agent/src/api/routes/organizations.py`** - **22 instances**

   - Lines 300-304, 351-354, 430-432, 472-474, 508, 539-541, 575, 588, 606, 630, 643, 663, 696, 709, 732
   - Imports inside route handlers: `async_get_session`, models, SQLAlchemy components

2. **`modules/agent/src/api/routes/individuals.py`** - **15 instances**

   - Lines 272-274, 303-306, 366-369, 449-451, 500-502, 536, 567-569
   - Similar pattern - imports inside route handlers

3. **`modules/agent/src/repositories/contact_repository.py`** - **8 instances**

   - Lines 16, 29, 69, 83, 95, 192, 207, 301
   - Imports inside functions: `ContactIndividual`, `ContactOrganization`

4. **`modules/agent/src/repositories/job_repository.py`** - **3 instances**

   - Lines 84, 212
   - `LeadProject` imported inside functions

5. **`modules/agent/src/repositories/opportunity_repository.py`** - **3 instances**

   - Lines 45, 147
   - `LeadProject` imported inside functions

6. **`modules/agent/src/repositories/partnership_repository.py`** - **3 instances**

   - Lines 45, 147
   - `LeadProject` imported inside functions

7. **`modules/agent/src/services/report_builder.py`** (Line 643)

   - `from openpyxl import Workbook` inside function

8. **`modules/agent/src/agent/graph.py`** (Lines 180-181)

   - LangGraph imports inside `build_default_graph()`

9. **`modules/agent/src/agent/orchestrator.py`** (Lines 152-160)

   - Multiple imports inside function

10. **`modules/agent/src/lib/auth.py`** (Line 142)

    - `from services.auth_service import get_or_create_user` inside function

11. **`modules/agent/src/agent/util/langfuse_helper.py`** (Line 27)

    - `from langfuse.langchain import CallbackHandler` inside function

12. **`modules/agent/src/services/supabase_client.py`** (Line 14)

    - `from supabase import create_client, Client` inside function

13. **`modules/agent/src/repositories/company_signal_repository.py`** (Line 46)

    - `from datetime import date` inside function

14. **`modules/agent/src/repositories/funding_round_repository.py`** (Line 36)

    - `from datetime import date` inside function

**Impact**:

- Slower function execution (imports executed on each call)
- Harder to track dependencies
- Potential circular import issues

**Recommendations**:

- Move all imports to top of file
- If circular imports occur, refactor module structure
- Use lazy imports only if absolutely necessary (document why)

---

### ✅ 9. Be Concise

**Status**: ✅ **Mostly Compliant**

Code is generally concise. Some verbose areas:

- Dictionary building functions (e.g., `_org_to_dict`, `_job_to_response_dict`) - acceptable for clarity
- Some service functions have long parameter lists - consider using dataclasses/dicts

---

## Summary Statistics

| Rule                          | Status              | Violations | Files Affected |
| ----------------------------- | ------------------- | ---------- | -------------- |
| KISS                          | ✅ Mostly Compliant | 2          | 2              |
| YAGNI                         | ✅ Compliant        | 0          | 0              |
| SOLID                         | ✅ Mostly Compliant | 0          | 0              |
| Controller-Service-Repository | ✅ Compliant        | 0          | 0              |
| Functional Programming        | ✅ Compliant        | 1          | 1              |
| Guard Clauses                 | ❌ Violations       | ~15+       | 5+             |
| No break/continue             | ❌ Violation        | 1          | 1              |
| Top-Level Imports             | ❌ Violations       | 76         | 15+            |
| Be Concise                    | ✅ Mostly Compliant | 0          | 0              |

---

## Priority Recommendations

### 🔴 High Priority

1. **Move imports to top level** (76 instances)

   - **Impact**: Performance, maintainability, dependency tracking
   - **Effort**: Medium (may require refactoring to avoid circular imports)
   - **Files**: `modules/agent/src/api/routes/organizations.py`, `modules/agent/src/api/routes/individuals.py`, `modules/agent/src/repositories/contact_repository.py`, etc.

2. **Refactor guard clauses** (15+ instances)
   - **Impact**: Code readability, maintainability
   - **Effort**: Low-Medium
   - **Files**: `modules/agent/src/api/routes/notifications.py`, `modules/agent/src/api/routes/organizations.py`, `modules/agent/src/services/job_service.py`

### 🟡 Medium Priority

3. **Remove `continue` statement**

   - **Impact**: Consistency with rules
   - **Effort**: Low
   - **File**: `modules/agent/src/services/notification_service.py`

4. **Consider refactoring complex service functions**
   - **Impact**: Maintainability
   - **Effort**: Medium
   - **Files**: `modules/agent/src/services/job_service.py`, `modules/agent/src/services/report_builder.py`

---

## Example Refactorings

### Example 1: Guard Clauses

**Before** (`modules/agent/src/api/routes/notifications.py`):

```python
def _get_entity_project_id(entity_type: str, entity_id: int) -> int | None:
    if not entity_type or not entity_id:
        return None

    if entity_type == "Deal":
        # ...
    if entity_type == "Task":
        # ...
    if entity_type == "Lead":
        # ...
    if entity_type in ("Individual", "Organization"):
        # ...
    else:
        return None
```

**After**:

```python
def _get_entity_project_id(entity_type: str, entity_id: int) -> int | None:
    if not entity_type or not entity_id:
        return None

    handlers = {
        "Deal": _get_deal_project_id,
        "Task": _get_task_project_id,
        "Lead": _get_lead_project_id,
        "Individual": _get_account_project_id,
        "Organization": _get_account_project_id,
    }

    handler = handlers.get(entity_type)
    if not handler:
        return None

    return handler(entity_id)
```

### Example 2: Remove `continue`

**Before** (`modules/agent/src/services/notification_service.py`):

```python
for user_id in participants:
    if user_id in notified_users:
        continue
    await create_notification(...)
```

**After**:

```python
unnotified_participants = [uid for uid in participants if uid not in notified_users]
for user_id in unnotified_participants:
    await create_notification(...)
```

### Example 3: Top-Level Imports

**Before** (`modules/agent/src/api/routes/organizations.py`):

```python
@router.post("")
async def create_organization_route(...):
    from lib.db import async_get_session
    from models.Account import Account
    # ...
```

**After**:

```python
from lib.db import async_get_session
from models.Account import Account
# ... (at top of file)

@router.post("")
async def create_organization_route(...):
    # ...
```

---

## Conclusion

The codebase follows most Python code rules well, particularly in architecture (controller-service-repository) and functional programming style. The main areas for improvement are:

1. **Import placement** - 76 instances need to be moved to top level
2. **Guard clauses** - 15+ instances of nested conditionals/else statements
3. **Break/continue** - 1 instance of `continue` statement

These violations are fixable with moderate effort and will improve code maintainability and performance.

---

## Next Steps

1. Create tickets for high-priority refactorings
2. Set up linting rules to catch import violations
3. Add code review checklist items for guard clauses
4. Refactor incrementally, starting with most-impacted files

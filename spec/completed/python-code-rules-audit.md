# Python Code Rules Audit - modules/agent

**Date:** 2025-01-XX  
**Scope:** `modules/agent/src/`  
**Rules:** KISS, YAGNI, SOLID (functional), Controller-Service-Repository, Guard Clauses, No break/continue, Top-level imports, Conciseness

## Summary

Found violations across multiple categories:
- **Local imports inside functions:** 3 instances
- **Nested conditionals / else statements:** 4 instances  
- **OOP patterns:** 1 instance (ABC classes)

## Violations by Category

### 1. Top-Level Imports (3 violations)

**Rule:** Imports should be defined at the top level of all files. Do not use local imports inside functions.

#### `modules/agent/src/repositories/document_repository.py`

**Lines 211-215** - Local imports inside `get_document_entities()`:
```python
async def get_document_entities(document_id: int) -> List[Dict[str, Any]]:
    """Get all entities linked to a document with their names."""
    from models.Organization import Organization
    from models.Individual import Individual
    from models.Project import Project
    from models.Contact import Contact
    from models.Lead import Lead
```

**Fix:** Move imports to top of file.

**Line 414** - Local import inside `create_new_version()`:
```python
# Mark all versions as not current (use Asset table directly for inheritance)
from sqlalchemy import update
```

**Fix:** Move `update` import to top of file.

#### `modules/agent/src/lib/auth.py`

**Line 142** - Local import inside `require_auth()` decorator:
```python
from services.auth_service import get_or_create_user
```

**Fix:** Move import to top of file.

---

### 2. Guard Clause Rule (4 violations)

**Rule:** Use guard clauses; avoid nested conditionals, `else`, and `switch/case`.

#### `modules/agent/src/api/routes/documents.py`

**Lines 235-247** - If/else pattern for storage type:
```python
url = storage.get_url(document.storage_path)
if url.startswith("file://"):
    # Local storage - stream the file
    content = await storage.download(document.storage_path)
    return StreamingResponse(...)
else:
    # R2 - redirect to presigned URL
    return RedirectResponse(url=url)
```

**Fix:** Use guard clause:
```python
url = storage.get_url(document.storage_path)
if not url.startswith("file://"):
    return RedirectResponse(url=url)

# Local storage - stream the file
content = await storage.download(document.storage_path)
return StreamingResponse(...)
```

#### `modules/agent/src/repositories/document_repository.py`

**Lines 229-252** - Long elif chain in `get_document_entities()`:
```python
if entity_type == "Organization":
    name_stmt = select(Organization.name).where(Organization.id == entity_id)
    # ...
elif entity_type == "Individual":
    name_stmt = select(Individual.first_name, Individual.last_name).where(Individual.id == entity_id)
    # ...
elif entity_type == "Project":
    # ...
elif entity_type == "Contact":
    # ...
elif entity_type == "Lead":
    # ...
```

**Fix:** Use dictionary mapping or separate functions:
```python
ENTITY_NAME_QUERIES = {
    "Organization": lambda session, eid: select(Organization.name).where(Organization.id == eid),
    "Individual": lambda session, eid: select(Individual.first_name, Individual.last_name).where(Individual.id == eid),
    # ...
}
```

#### `modules/agent/src/lib/auth.py`

**Lines 106-114** - If/else for email extraction:
```python
def _get_email_from_claims(claims: dict) -> Optional[str]:
    """Extract email from standard or namespaced claim."""
    email = claims.get("email")
    if email:
        return email

    auth0_domain = os.getenv("AUTH0_DOMAIN")
    if not auth0_domain:
        return None

    return claims.get(f"https://{auth0_domain}/email")
```

**Note:** This is already using guard clauses appropriately. No violation.

#### `modules/agent/src/api/routes/notifications.py`

**Lines 38-79** - Nested if statements in `_get_entity_project_id()`:
```python
if entity_type == "Deal":
    # ...
if entity_type == "Task":
    # ...
    if result and result[0] and result[1]:
        return _get_entity_project_id(result[0], result[1])
if entity_type == "Lead":
    # ...
if entity_type in ("Individual", "Organization"):
    # ...
    if result and result[0]:
        # ...
        if proj_result:
            return proj_result[0]
```

**Fix:** Use early returns and guard clauses:
```python
if not entity_type or not entity_id:
    return None

if entity_type == "Deal":
    # ... return early

if entity_type == "Task":
    # ... return early

# etc.
```

---

### 3. OOP Patterns (1 violation)

**Rule:** Prefer functional programming; avoid OOP unless unavoidable.

#### `modules/agent/src/lib/storage.py`

**Lines 16-38** - ABC abstract base class:
```python
class StorageBackend(ABC):
    """Abstract base class for storage backends."""
    @abstractmethod
    async def upload(self, path: str, data: bytes, content_type: str) -> str:
        ...
```

**Note:** While ABC is OOP, this is a reasonable use case for polymorphism. However, could be refactored to use Protocol (structural subtyping) or dict-based dispatch for a more functional approach.

**Alternative (functional):**
```python
from typing import Protocol

class StorageBackend(Protocol):
    async def upload(self, path: str, data: bytes, content_type: str) -> str: ...
    async def download(self, path: str) -> bytes: ...
    # ...
```

---

### 4. Break/Continue Usage

**Rule:** Don't use `break`/`continue`; refactor with small pure functions, early returns.

**Status:** No violations found. The grep found:
- `modules/agent/src/agent/graph.py:207` - `return False  # not a context message; let caller continue` (comment, not code)
- `modules/agent/src/agent/evaluators/_base_evaluator.py:115` - Comment about "break retry loop" (not actual break statement)

---

## Recommendations

1. **Immediate fixes:**
   - Move all local imports to top of files
   - Refactor nested conditionals to use guard clauses
   - Replace if/else patterns with early returns

2. **Consider refactoring:**
   - `get_document_entities()` elif chain → dictionary mapping
   - `_get_entity_project_id()` nested ifs → early returns
   - Storage ABC → Protocol (optional, ABC is acceptable)

3. **Pattern consistency:**
   - Most repository functions already use guard clauses well
   - API routes generally follow patterns correctly
   - Focus on the identified violations above

---

## Files Reviewed

- `modules/agent/src/repositories/document_repository.py`
- `modules/agent/src/api/routes/documents.py`
- `modules/agent/src/lib/auth.py`
- `modules/agent/src/lib/storage.py`
- `modules/agent/src/api/routes/notifications.py`
- `modules/agent/src/api/routes/organizations.py`
- `modules/agent/src/repositories/organization_repository.py`

## Notes

- SQLAlchemy models (OOP) are acceptable as they're required by the ORM
- Pydantic models (BaseModel) are acceptable as they're required by FastAPI
- Most code already follows functional patterns well
- Violations are relatively minor and easy to fix


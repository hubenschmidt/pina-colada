# Enforce Service Layer as Only Repository Consumer

## Goal
1. Only the services layer imports from repositories
2. Pydantic schemas live in a dedicated `schemas/` directory
3. No re-export chains needed

## Architecture After Refactor
```
schemas/          ← Pydantic models (XCreate, XUpdate, etc.)
    ↑       ↑
    |       |
Routes  Repositories
    ↓       ↓
Controllers → Services → Repositories
```

- **schemas/**: All Pydantic request/response models
- **Routes**: Import schemas directly, import functions from controllers
- **Controllers**: Pass-through layer, no schema imports needed
- **Services**: Business logic, import from repositories
- **Repositories**: Data access, import schemas for type hints if needed

## Implementation Steps

### Step 1: Create schemas/ directory structure
```
src/schemas/
├── __init__.py
├── organization.py    # OrganizationCreate, OrganizationUpdate, etc.
├── individual.py      # IndividualCreate, IndividualUpdate, etc.
├── contact.py         # ContactCreate, ContactUpdate
├── job.py             # JobCreate, JobUpdate
├── opportunity.py     # OpportunityCreate, OpportunityUpdate
├── partnership.py     # PartnershipCreate, PartnershipUpdate
├── project.py         # ProjectCreate, ProjectUpdate
├── task.py            # TaskCreate, TaskUpdate
├── document.py        # DocumentUpdate, EntityLink
├── comment.py         # CommentCreate, CommentUpdate
├── note.py            # NoteCreate, NoteUpdate
├── industry.py        # IndustryCreate
├── technology.py      # TechnologyCreate
├── provenance.py      # ProvenanceCreate
├── notification.py    # MarkReadRequest, MarkEntityReadRequest
├── preferences.py     # UpdateUserPreferencesRequest, UpdateTenantPreferencesRequest
├── report.py          # ReportQueryRequest, SavedReportCreate, SavedReportUpdate
└── tenant.py          # TenantCreate
```

### Step 2: Move Pydantic models from repositories to schemas/
For each repository that has Pydantic models:
1. Move the model classes to the corresponding schema file
2. Update repository to import from schemas

### Step 3: Update routes to import schemas directly
Change from:
```python
from controllers.x_controller import XCreate, XUpdate
```
To:
```python
from schemas.x import XCreate, XUpdate
```

### Step 4: Remove re-exports from services and controllers
- Remove `__all__` re-exports of Pydantic models
- Remove Pydantic model imports that were only for re-export

### Step 5: Update repositories to import from schemas
Change from:
```python
class XCreate(BaseModel):
    ...
```
To:
```python
from schemas.x import XCreate, XUpdate
```

## Files to Create
- `schemas/__init__.py`
- `schemas/organization.py`
- `schemas/individual.py`
- `schemas/contact.py`
- `schemas/job.py`
- `schemas/opportunity.py`
- `schemas/partnership.py`
- `schemas/project.py`
- `schemas/task.py`
- `schemas/document.py`
- `schemas/comment.py`
- `schemas/note.py`
- `schemas/industry.py`
- `schemas/technology.py`
- `schemas/provenance.py`
- `schemas/notification.py`
- `schemas/preferences.py`
- `schemas/report.py`
- `schemas/tenant.py`

## Files to Modify
- All repositories (remove Pydantic model definitions, add imports from schemas)
- All routes (import schemas directly instead of from controllers)
- All services (remove `__all__` re-exports)
- All controllers (remove `__all__` re-exports)

## Verification
After refactoring:
```bash
# Controllers should NOT import from repositories:
grep -r "from repositories\." controllers/
# Should return NO results

# Routes should NOT import from repositories:
grep -r "from repositories\." api/routes/
# Should return NO results

# Routes should import schemas:
grep -r "from schemas\." api/routes/
# Should show schema imports
```

## Benefits
- Clean separation of concerns
- No re-export chains
- Schemas can be shared across layers without circular imports
- Follows FastAPI best practices

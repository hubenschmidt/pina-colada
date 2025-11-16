# Lead Parameterization Specification

## Executive Summary

This specification outlines the refactoring of the API from type-specific routes (`/jobs/*`, `/opportunities/*`, `/partnerships/*`) to a unified parameterized lead structure (`/leads/:type/*`). This change aligns the API design with the underlying database model where Job, Opportunity, and Partnership are specialized types of the Lead entity using Joined Table Inheritance (JTI).

**Key Benefits:**
- **Semantic Clarity:** API structure matches domain model (Job IS-A Lead)
- **Extensibility:** Easy to add new lead types without creating new route files
- **Code Reduction:** Unified controller/service layer reduces duplication
- **Consistency:** Single pattern for all lead operations

**Current State:** Separate `/jobs` and `/leads` endpoints with inconsistent routing patterns.

**Target State:** Unified `/leads/:type` endpoints where `type` ∈ {`job`, `opportunity`, `partnership`}.

---

## 1. Current State Analysis

### 1.1 Database Schema

The application uses **Joined Table Inheritance (JTI)** pattern:

```
┌─────────────────────────────────────────────────┐
│ Lead (Base Table)                               │
├─────────────────────────────────────────────────┤
│ id: BigInteger (PK)                             │
│ deal_id: BigInteger (FK → Deal.id)              │
│ type: Text (DISCRIMINATOR)                      │  ← 'Job' | 'Opportunity' | 'Partnership'
│ title: Text                                     │
│ description: Text                               │
│ source: Text                                    │
│ current_status_id: BigInteger (FK → Status.id)  │
│ owner_user_id: BigInteger                       │
│ created_at: DateTime                            │
│ updated_at: DateTime                            │
└─────────────────────────────────────────────────┘
         │
         ├─────── 1:1 ──────┐
         │                  │
         ▼                  ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Job              │  │ Opportunity      │  │ Partnership      │
├──────────────────┤  ├──────────────────┤  ├──────────────────┤
│ id (PK, FK)      │  │ id (PK, FK)      │  │ id (PK, FK)      │
│ organization_id  │  │ organization_id  │  │ organization_id  │
│ job_title        │  │ opportunity_name │  │ partnership_name │
│ job_url          │  │ estimated_value  │  │ partnership_type │
│ notes            │  │ probability      │  │ start_date       │
│ resume_date      │  │ expected_close   │  │ end_date         │
│ salary_range     │  │ notes            │  │ notes            │
│ created_at       │  │ created_at       │  │ created_at       │
│ updated_at       │  │ updated_at       │  │ updated_at       │
└──────────────────┘  └──────────────────┘  └──────────────────┘
```

**Key Points:**
- `Lead.type` is the discriminator column
- Child tables (`Job`, `Opportunity`, `Partnership`) share the same `id` as their parent Lead
- Deleting a Lead cascades to the child table
- Queries require JOIN between Lead and child table

### 1.2 Current API Endpoints

#### Jobs Routes (`/jobs`)
```
GET    /jobs                      → get_jobs(page, limit, orderBy, order, search)
POST   /jobs                      → create_job(job_data)
GET    /jobs/{job_id}             → get_job(job_id)
PUT    /jobs/{job_id}             → update_job(job_id, job_data)
DELETE /jobs/{job_id}             → delete_job(job_id)
GET    /jobs/recent-resume-date   → get_recent_resume_date()
```

**Request Model (JobCreate):**
```python
{
  "company": str,              # Required - mapped to Organization
  "job_title": str,            # Required
  "date": Optional[str],       # ISO date string
  "job_url": Optional[str],
  "salary_range": Optional[str],
  "notes": Optional[str],
  "resume": Optional[str],     # ISO datetime string
  "status": str = "applied",
  "source": str = "manual"
}
```

**Response Model:**
```python
{
  "id": str,
  "company": str,              # From organization.name
  "job_title": str,
  "date": str,                 # YYYY-MM-DD from lead.created_at
  "status": str,               # From lead.current_status.name
  "job_url": str | null,
  "salary_range": str | null,
  "notes": str | null,
  "resume": str | null,        # ISO datetime
  "source": str,
  "created_at": str,
  "updated_at": str
}
```

#### Leads Routes (`/leads`)
```
GET    /leads                     → get_leads(statuses: Optional[str])
GET    /leads/statuses            → get_statuses()
POST   /leads/{job_id}/apply      → mark_lead_as_applied(job_id)
POST   /leads/{job_id}/do-not-apply → mark_lead_as_do_not_apply(job_id)
```

**Issues with Current Design:**
1. `/leads` routes are hardcoded to `job_controller` - no generic lead handling
2. Parameter naming inconsistency: `{job_id}` instead of `{lead_id}` or `{id}`
3. No support for Opportunity or Partnership types despite models existing
4. Duplicate code between job_repository and lead_repository
5. Frontend has separate API functions for semantically related operations

### 1.3 Current Code Flow

```
┌─────────────┐
│   Routes    │  jobs.py, leads.py
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Controllers │  job_controller.py (no lead_controller exists)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Services   │  job_service.py (no lead_service exists)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│Repositories │  job_repository.py, lead_repository.py (limited)
└─────────────┘
```

**Current Type Discrimination:**
- **Routes:** Separate route files (`jobs.py` vs future `opportunities.py`)
- **Controllers:** Type-specific controllers (`job_controller.py`)
- **Services:** Type-specific services (`job_service.py`)
- **Repositories:** Mixed - some type-specific, some generic with type parameter
  - `job_repository.find_all_jobs()` - implicit type filtering via JOIN
  - `lead_repository.find_all_leads(lead_type)` - explicit type parameter

### 1.4 Frontend API Client

**File:** `modules/client/api/index.ts`

**Current Functions:**
```typescript
// Job-specific
getJobs(page, limit, orderBy, order, search): Promise<PageData<CreatedJob>>
createJob(job: Partial<CreatedJob>): Promise<CreatedJob>
getJob(id: string): Promise<CreatedJob>
updateJob(id: string, job: Partial<CreatedJob>): Promise<CreatedJob>
deleteJob(id: string): Promise<void>
getRecentResumeDate(): Promise<string | null>

// Lead-specific (but only for jobs currently)
getLeads(statusNames?: string[]): Promise<JobWithLeadStatus[]>
getStatuses(): Promise<LeadStatus[]>
markLeadAsApplied(jobId: string): Promise<CreatedJob>
markLeadAsDoNotApply(jobId: string): Promise<CreatedJob>
```

**Current Types:**
```typescript
interface CreatedJob {
  id: string;
  company: string;
  job_title: string;
  date: string;
  status: 'Lead' | 'Applied' | 'Interviewing' | 'Rejected' | 'Offer' | 'Accepted' | 'Do Not Apply';
  job_url: string | null;
  notes: string | null;
  resume: string | null;
  salary_range: string | null;
  source: 'manual' | 'agent';
  lead_status_id: string | null;
  created_at: string;
  updated_at: string;
}

type LeadStatus = {
  id: string;
  name: "Qualifying" | "Cold" | "Warm" | "Hot";
  description: string | null;
  created_at: string;
};
```

---

## 2. Proposed Architecture

### 2.1 New Endpoint Structure

All lead operations will use parameterized routes:

```
GET    /leads/:type                    # List leads of specific type (paginated)
POST   /leads/:type                    # Create new lead
GET    /leads/:type/:id                # Get single lead
PUT    /leads/:type/:id                # Update lead
DELETE /leads/:type/:id                # Delete lead
GET    /leads/:type/statuses           # Get available statuses for this lead type
```

Where `:type` ∈ `{'job', 'opportunity', 'partnership'}` (lowercase, validated)

**Examples:**

```bash
# Jobs
GET    /leads/job?page=1&limit=25&orderBy=date&order=DESC&search=engineer
POST   /leads/job
GET    /leads/job/123
PUT    /leads/job/123
DELETE /leads/job/123
GET    /leads/job/statuses

# Opportunities
GET    /leads/opportunity?page=1&limit=50
POST   /leads/opportunity
GET    /leads/opportunity/456
PUT    /leads/opportunity/456
DELETE /leads/opportunity/456
GET    /leads/opportunity/statuses

# Partnerships
GET    /leads/partnership
POST   /leads/partnership
GET    /leads/partnership/789
```

### 2.2 Type-Specific Operations

Some operations are specific to certain lead types. These will be namespaced under the type:

```bash
# Job-specific operations
GET    /leads/job/recent-resume-date
POST   /leads/job/:id/apply
POST   /leads/job/:id/reject

# Opportunity-specific operations (future)
PUT    /leads/opportunity/:id/close
PUT    /leads/opportunity/:id/update-probability

# Partnership-specific operations (future)
POST   /leads/partnership/:id/activate
POST   /leads/partnership/:id/renew
```

### 2.3 Unified Code Architecture

```
┌──────────────────────┐
│   Routes (leads.py)  │
│  /leads/:type/*      │
└──────────┬───────────┘
           │ validate type param
           ▼
┌──────────────────────┐
│  lead_controller.py  │
│  Generic operations  │
│  + Type dispatch     │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────────────────────────┐
│         lead_service.py                  │
│  Unified business logic                  │
│  Dispatches to type-specific services    │
└──────────┬───────────────────────────────┘
           │
           ├─────────┬─────────┬─────────┐
           ▼         ▼         ▼         ▼
      ┌─────────┐ ┌────────┐ ┌────────┐
      │  job_   │ │  opp_  │ │  ptnr_ │
      │ service │ │ service│ │ service│
      └────┬────┘ └───┬────┘ └───┬────┘
           │          │          │
           ▼          ▼          ▼
      ┌─────────┐ ┌────────┐ ┌────────┐
      │  job_   │ │  opp_  │ │  ptnr_ │
      │   repo  │ │   repo │ │   repo │
      └─────────┘ └────────┘ └────────┘
```

**Key Design Principles:**
1. **Type validation at route level** - fail fast with 400 Bad Request for invalid types
2. **Unified controller** - single `lead_controller.py` handles all types
3. **Dispatch pattern** - controller/service dispatches to type-specific implementations
4. **Type-specific details** - keep job/opportunity/partnership logic separate for maintainability

### 2.4 Type Validation

**Path Parameter Validation:**
```python
from enum import Enum
from fastapi import Path, HTTPException

class LeadType(str, Enum):
    JOB = "job"
    OPPORTUNITY = "opportunity"
    PARTNERSHIP = "partnership"

@router.get("/leads/{lead_type}")
async def get_leads_route(
    lead_type: LeadType = Path(..., description="Type of lead")
):
    # lead_type is validated automatically by FastAPI
    return await get_leads(lead_type, ...)
```

### 2.5 Request/Response Models

**Generic Lead Response:**
```python
{
  "id": str,
  "type": "job" | "opportunity" | "partnership",
  "title": str,
  "status": str,                    # From current_status.name
  "source": str,
  "created_at": str,
  "updated_at": str,

  # Type-specific fields (nested)
  "details": {
    # For jobs:
    "company": str,
    "job_title": str,
    "job_url": str | null,
    "salary_range": str | null,
    "notes": str | null,
    "resume_date": str | null,

    # For opportunities:
    # "company": str,
    # "opportunity_name": str,
    # "estimated_value": number,
    # "probability": number,
    # "expected_close_date": str,

    # For partnerships:
    # "company": str,
    # "partnership_name": str,
    # "partnership_type": str,
    # "start_date": str,
    # "end_date": str,
  }
}
```

**Alternative (Flat Structure - Easier Migration):**
Keep current flat structure, add `type` field:
```python
{
  "id": str,
  "type": "job",                    # NEW FIELD
  "company": str,
  "job_title": str,
  "date": str,
  "status": str,
  # ... rest same as current
}
```

**Recommendation:** Use flat structure for backward compatibility. Nested structure can be introduced in v2 API.

---

## 3. Migration Strategy

### Phase 1: Dual Support (Weeks 1-2)

**Goal:** Add new parameterized endpoints while keeping old endpoints functional.

**Steps:**

1. **Create new route file** `api/routes/parameterized_leads.py`:
   ```python
   from fastapi import APIRouter, Path
   from controllers.lead_controller import *

   router = APIRouter(prefix="/leads", tags=["leads"])

   @router.get("/{lead_type}")
   async def get_leads_route(lead_type: LeadType, ...):
       return await get_leads(lead_type, ...)

   @router.post("/{lead_type}")
   async def create_lead_route(lead_type: LeadType, data: dict):
       return await create_lead(lead_type, data)

   # ... etc
   ```

2. **Create unified controller** `controllers/lead_controller.py`:
   ```python
   async def get_leads(lead_type: LeadType, page, limit, ...):
       # Validate and dispatch
       return await lead_service.get_leads(lead_type.value, page, limit, ...)
   ```

3. **Create unified service** `services/lead_service.py`:
   ```python
   async def get_leads(lead_type: str, page, limit, ...):
       if lead_type == "job":
           return await job_service.get_all_jobs(page, limit, ...)
       elif lead_type == "opportunity":
           return await opportunity_service.get_all_opportunities(...)
       # ... etc
   ```

4. **Register both routers** in `server.py`:
   ```python
   from api.routes import jobs_routes, parameterized_leads_routes

   app.include_router(jobs_routes)              # OLD - keep for now
   app.include_router(parameterized_leads_routes)  # NEW
   ```

5. **Frontend dual support** - add new functions, keep old ones:
   ```typescript
   // NEW - generic
   export const getLeads = async <T>(
     type: LeadType,
     page: number,
     ...
   ): Promise<PageData<T>> => {
     return apiGet<PageData<T>>(`/leads/${type}?page=${page}...`);
   };

   // OLD - wrapper around new function (backward compatible)
   export const getJobs = async (...): Promise<PageData<CreatedJob>> => {
     return getLeads<CreatedJob>('job', ...);
   };
   ```

**Deliverables:**
- [ ] New parameterized routes working for `job` type
- [ ] Old `/jobs` routes still functional
- [ ] Frontend can call either old or new endpoints
- [ ] All tests passing for both endpoint sets

### Phase 2: Deprecation & Migration (Weeks 3-4)

**Goal:** Migrate frontend to use new endpoints, add deprecation warnings to old ones.

**Steps:**

1. **Add deprecation warnings** to old endpoints:
   ```python
   import warnings

   @router.get("/jobs")
   async def get_jobs_deprecated(...):
       warnings.warn(
           "GET /jobs is deprecated. Use GET /leads/job instead. "
           "This endpoint will be removed in v2.0",
           DeprecationWarning
       )
       # Internally call new endpoint
       return await get_leads_route(LeadType.JOB, ...)
   ```

2. **Update API documentation** - mark old endpoints as deprecated in OpenAPI schema

3. **Migrate frontend components** one by one:
   ```typescript
   // Before:
   const { data } = await getJobs(page, limit);

   // After:
   const { data } = await getLeads('job', page, limit);
   ```

4. **Update all component imports** to use new API functions

5. **Add console warnings** in frontend for old API usage:
   ```typescript
   export const getJobs = async (...) => {
     console.warn('getJobs() is deprecated. Use getLeads("job", ...) instead');
     return getLeads<CreatedJob>('job', ...);
   };
   ```

**Deliverables:**
- [ ] Frontend fully migrated to new endpoints
- [ ] Deprecation warnings visible in logs
- [ ] Updated API documentation
- [ ] Migration guide published

### Phase 3: Cleanup (Week 5)

**Goal:** Remove old endpoints and deprecated code.

**Steps:**

1. **Remove old route files:**
   - Delete `api/routes/jobs.py`
   - Remove from `api/routes/__init__.py`
   - Remove from `server.py`

2. **Remove old controller** (optional - may keep if it has type-specific logic):
   - Move any job-specific logic to `services/job_service.py`
   - Delete `controllers/job_controller.py`

3. **Clean up frontend:**
   - Remove deprecated wrapper functions (`getJobs`, etc.)
   - Update all direct imports to use `getLeads`
   - Remove feature flags

4. **Database cleanup** (optional):
   - Ensure all Leads have correct `type` value
   - Add NOT NULL constraint to `Lead.type` if not already present
   - Add CHECK constraint: `type IN ('Job', 'Opportunity', 'Partnership')`

**Deliverables:**
- [ ] All old code removed
- [ ] No deprecation warnings
- [ ] Clean codebase
- [ ] Updated tests

### Rollback Plan

If issues arise during migration:

1. **Phase 1 rollback:**
   - Remove new router from `server.py`
   - Frontend continues using old endpoints

2. **Phase 2 rollback:**
   - Revert frontend to use old API functions
   - Remove deprecation warnings
   - Continue using both endpoint sets

3. **Phase 3 rollback:**
   - Restore deleted route files from git
   - Re-register old router in `server.py`

---

## 4. Breaking Changes

### 4.1 API Endpoint Paths

| Old Endpoint | New Endpoint | Notes |
|-------------|--------------|-------|
| `GET /jobs` | `GET /leads/job` | Query params remain same |
| `POST /jobs` | `POST /leads/job` | Request body remains same |
| `GET /jobs/{id}` | `GET /leads/job/{id}` | - |
| `PUT /jobs/{id}` | `PUT /leads/job/{id}` | - |
| `DELETE /jobs/{id}` | `DELETE /leads/job/{id}` | - |
| `GET /jobs/recent-resume-date` | `GET /leads/job/recent-resume-date` | - |
| `GET /leads` | `GET /leads/job` | With status query param |
| `GET /leads/statuses` | `GET /leads/job/statuses` | - |
| `POST /leads/{job_id}/apply` | `POST /leads/job/{id}/apply` | Parameter name change |
| `POST /leads/{job_id}/do-not-apply` | `POST /leads/job/{id}/reject` | Path renamed |

### 4.2 Request Parameter Names

- `{job_id}` → `{id}` (more generic)
- No changes to query parameters
- No changes to request body structure (initially)

### 4.3 Response Schema

**Option A: No changes** (recommended for easier migration)
- Keep current flat structure
- Add `type` field to response
- All existing fields remain

**Option B: Nested structure** (cleaner but breaking)
- Separate common lead fields from type-specific fields
- Requires frontend to access `response.details.job_title` instead of `response.job_title`

### 4.4 Frontend Import Changes

**Before:**
```typescript
import { getJobs, createJob, updateJob, deleteJob } from 'api';
```

**After:**
```typescript
import { getLeads, createLead, updateLead, deleteLead } from 'api';

// Usage:
const jobs = await getLeads('job', page, limit);
const opportunities = await getLeads('opportunity', page, limit);
```

### 4.5 TypeScript Type Changes

**New Types:**
```typescript
type LeadType = 'job' | 'opportunity' | 'partnership';

interface BaseLead {
  id: string;
  type: LeadType;
  title: string;
  status: string;
  source: string;
  created_at: string;
  updated_at: string;
}

interface JobLead extends BaseLead {
  type: 'job';
  company: string;
  job_title: string;
  job_url: string | null;
  // ... job-specific fields
}

interface OpportunityLead extends BaseLead {
  type: 'opportunity';
  company: string;
  opportunity_name: string;
  estimated_value: number | null;
  // ... opportunity-specific fields
}

type Lead = JobLead | OpportunityLead | PartnershipLead;
```

---

## 5. Implementation Plan

### 5.1 Backend Implementation

#### Step 1: Models (Already Complete ✓)
- [x] Lead base model with discriminator
- [x] Job, Opportunity, Partnership models
- [x] Joined Table Inheritance configured

#### Step 2: Repositories

**Create unified lead repository functions:**

```python
# repositories/lead_repository.py

async def find_leads_by_type(
    lead_type: str,
    page: int = 1,
    page_size: int = 50,
    search: Optional[str] = None,
    order_by: str = "created_at",
    order: str = "DESC"
) -> Tuple[List[Lead], int]:
    """Generic lead finder with type discrimination."""
    async with async_get_session() as session:
        # Build base query with type filter
        stmt = select(Lead).where(Lead.type == lead_type.capitalize())

        # Eager load type-specific details
        if lead_type == "job":
            stmt = stmt.options(joinedload(Lead.job))
        elif lead_type == "opportunity":
            stmt = stmt.options(joinedload(Lead.opportunity))
        elif lead_type == "partnership":
            stmt = stmt.options(joinedload(Lead.partnership))

        # Apply search, sorting, pagination
        # ... (similar to current find_all_jobs logic)

        return leads, total_count
```

**Checklist:**
- [ ] `find_leads_by_type(lead_type, filters)` - generic query
- [ ] `create_lead_with_details(lead_type, lead_data, detail_data)` - dispatches to type-specific create
- [ ] `update_lead_with_details(lead_id, lead_type, data)` - updates both tables
- [ ] `delete_lead(lead_id)` - cascades to child tables
- [ ] Type-specific repositories: `job_repository.py`, `opportunity_repository.py`, `partnership_repository.py`

#### Step 3: Services

**Create unified lead service:**

```python
# services/lead_service.py

from services import job_service, opportunity_service, partnership_service

_SERVICE_MAP = {
    "job": job_service,
    "opportunity": opportunity_service,
    "partnership": partnership_service,
}

async def get_leads(
    lead_type: str,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None
) -> Tuple[List[Any], int]:
    """Get leads by type with pagination."""
    validate_lead_type(lead_type)

    # Dispatch to type-specific service
    service = _SERVICE_MAP[lead_type]
    return await service.get_all(page, limit, order_by, order, search)

async def create_lead(lead_type: str, data: Dict[str, Any]) -> Any:
    """Create lead of specified type."""
    validate_lead_type(lead_type)

    service = _SERVICE_MAP[lead_type]
    return await service.create(data)

def validate_lead_type(lead_type: str):
    """Validate lead type parameter."""
    if lead_type not in _SERVICE_MAP:
        raise HTTPException(
            status_code=400,
            detail=f"Invalid lead type: {lead_type}. Must be one of: {list(_SERVICE_MAP.keys())}"
        )
```

**Checklist:**
- [ ] `lead_service.py` with dispatch logic
- [ ] Keep `job_service.py` for job-specific business logic
- [ ] Create `opportunity_service.py` with opportunity logic
- [ ] Create `partnership_service.py` with partnership logic
- [ ] Type validation helper function
- [ ] Service registry/map for dynamic dispatch

#### Step 4: Controllers

**Create unified lead controller:**

```python
# controllers/lead_controller.py

from lib.decorators import handle_http_exceptions
from services import lead_service

@handle_http_exceptions
async def get_leads(
    lead_type: str,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None
) -> dict:
    """Get paginated leads of specified type."""
    leads, total_count = await lead_service.get_leads(
        lead_type, page, limit, order_by, order, search
    )

    # Transform to response format
    items = [_lead_to_response_dict(lead, lead_type) for lead in leads]
    return _to_paged_response(total_count, page, limit, items)

def _lead_to_response_dict(lead: Lead, lead_type: str) -> Dict[str, Any]:
    """Convert Lead ORM to response dictionary."""
    base_dict = {
        "id": str(lead.id),
        "type": lead_type,
        "title": lead.title,
        "status": lead.current_status.name if lead.current_status else None,
        "source": lead.source,
        "created_at": lead.created_at.isoformat(),
        "updated_at": lead.updated_at.isoformat(),
    }

    # Add type-specific fields
    if lead_type == "job" and lead.job:
        base_dict.update({
            "company": lead.job.organization.name,
            "job_title": lead.job.job_title,
            "job_url": lead.job.job_url,
            # ... other job fields
        })
    elif lead_type == "opportunity" and lead.opportunity:
        base_dict.update({
            "company": lead.opportunity.organization.name,
            "opportunity_name": lead.opportunity.opportunity_name,
            # ... other opportunity fields
        })

    return base_dict
```

**Checklist:**
- [ ] Create `lead_controller.py`
- [ ] Implement CRUD operations with type parameter
- [ ] Response transformation logic
- [ ] Error handling with `@handle_http_exceptions`
- [ ] Keep type-specific controllers if needed for special operations

#### Step 5: Routes

**Create parameterized routes:**

```python
# api/routes/leads.py (refactored)

from fastapi import APIRouter, Path, Query
from enum import Enum
from controllers import lead_controller

class LeadType(str, Enum):
    JOB = "job"
    OPPORTUNITY = "opportunity"
    PARTNERSHIP = "partnership"

router = APIRouter(prefix="/leads", tags=["leads"])

@router.get("/{lead_type}")
async def get_leads_route(
    lead_type: LeadType = Path(...),
    page: int = Query(1, ge=1),
    limit: int = Query(25, ge=1, le=100),
    orderBy: str = Query("date"),
    order: str = Query("DESC"),
    search: Optional[str] = Query(None)
):
    """Get paginated leads of specified type."""
    return await lead_controller.get_leads(
        lead_type.value, page, limit, orderBy, order, search
    )

@router.post("/{lead_type}")
async def create_lead_route(
    lead_type: LeadType = Path(...),
    data: dict  # TODO: Use proper Pydantic model based on type
):
    """Create new lead."""
    return await lead_controller.create_lead(lead_type.value, data)

@router.get("/{lead_type}/{lead_id}")
async def get_lead_route(
    lead_type: LeadType = Path(...),
    lead_id: str = Path(...)
):
    """Get single lead by ID."""
    return await lead_controller.get_lead(lead_type.value, lead_id)

@router.put("/{lead_type}/{lead_id}")
async def update_lead_route(
    lead_type: LeadType = Path(...),
    lead_id: str = Path(...),
    data: dict
):
    """Update lead."""
    return await lead_controller.update_lead(lead_type.value, lead_id, data)

@router.delete("/{lead_type}/{lead_id}")
async def delete_lead_route(
    lead_type: LeadType = Path(...),
    lead_id: str = Path(...)
):
    """Delete lead."""
    return await lead_controller.delete_lead(lead_type.value, lead_id)

# Type-specific routes
@router.get("/{lead_type}/statuses")
async def get_statuses_route(
    lead_type: LeadType = Path(...)
):
    """Get available statuses for this lead type."""
    return await lead_controller.get_statuses(lead_type.value)

# Job-specific operations
@router.get("/job/recent-resume-date")
async def get_recent_resume_date_route():
    """Get most recent resume date from job leads."""
    return await lead_controller.get_recent_resume_date()

@router.post("/job/{lead_id}/apply")
async def mark_job_as_applied_route(lead_id: str = Path(...)):
    """Mark job lead as applied."""
    return await lead_controller.mark_job_as_applied(lead_id)

@router.post("/job/{lead_id}/reject")
async def mark_job_as_rejected_route(lead_id: str = Path(...)):
    """Mark job lead as do not apply."""
    return await lead_controller.mark_job_as_rejected(lead_id)
```

**Checklist:**
- [ ] Refactor `leads.py` with parameterized routes
- [ ] Add `LeadType` enum for validation
- [ ] Implement all CRUD operations
- [ ] Add type-specific operation routes
- [ ] Update router registration in `server.py`

#### Step 6: Testing

**Unit Tests:**
```python
# tests/test_lead_controller.py

async def test_get_jobs():
    response = await get_leads("job", page=1, limit=10, ...)
    assert response["items"][0]["type"] == "job"

async def test_get_opportunities():
    response = await get_leads("opportunity", page=1, limit=10, ...)
    assert response["items"][0]["type"] == "opportunity"

async def test_invalid_lead_type():
    with pytest.raises(HTTPException) as exc:
        await get_leads("invalid_type", ...)
    assert exc.value.status_code == 400
```

**Integration Tests:**
```python
# tests/test_api_leads.py

def test_get_jobs_endpoint(client):
    response = client.get("/leads/job?page=1&limit=10")
    assert response.status_code == 200
    assert "items" in response.json()

def test_create_job(client):
    data = {"company": "Acme", "job_title": "Engineer", ...}
    response = client.post("/leads/job", json=data)
    assert response.status_code == 201
```

**Checklist:**
- [ ] Unit tests for all controller functions
- [ ] Integration tests for all endpoints
- [ ] Test type validation (invalid types return 400)
- [ ] Test backward compatibility during dual-support phase
- [ ] Performance tests for JOINs

### 5.2 Frontend Implementation

#### Step 1: Update TypeScript Types

```typescript
// types/leads.ts

export type LeadType = 'job' | 'opportunity' | 'partnership';

export interface BaseLead {
  id: string;
  type: LeadType;
  title: string;
  status: string;
  source: 'manual' | 'agent';
  created_at: string;
  updated_at: string;
}

export interface JobLead extends BaseLead {
  type: 'job';
  company: string;
  job_title: string;
  date: string;
  job_url: string | null;
  salary_range: string | null;
  notes: string | null;
  resume: string | null;
  lead_status_id: string | null;
}

export interface OpportunityLead extends BaseLead {
  type: 'opportunity';
  company: string;
  opportunity_name: string;
  estimated_value: number | null;
  probability: number | null;
  expected_close_date: string | null;
  notes: string | null;
}

export interface PartnershipLead extends BaseLead {
  type: 'partnership';
  company: string;
  partnership_name: string;
  partnership_type: string | null;
  start_date: string | null;
  end_date: string | null;
  notes: string | null;
}

export type Lead = JobLead | OpportunityLead | PartnershipLead;

// Type guard helpers
export function isJobLead(lead: Lead): lead is JobLead {
  return lead.type === 'job';
}

export function isOpportunityLead(lead: Lead): lead is OpportunityLead {
  return lead.type === 'opportunity';
}
```

**Checklist:**
- [ ] Define `LeadType` enum
- [ ] Create `BaseLead` interface
- [ ] Create type-specific interfaces extending `BaseLead`
- [ ] Create union type `Lead`
- [ ] Add type guard functions

#### Step 2: Update API Client

```typescript
// api/index.ts

import { LeadType, Lead, JobLead, OpportunityLead } from '../types/leads';

/**
 * Generic function to get leads of any type
 */
export const getLeads = async <T extends Lead>(
  type: LeadType,
  page: number = 1,
  limit: number = 25,
  orderBy: string = "date",
  order: "ASC" | "DESC" = "DESC",
  search?: string
): Promise<PageData<T>> => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  return apiGet<PageData<T>>(`/leads/${type}?${params}`);
};

/**
 * Create lead of specified type
 */
export const createLead = async <T extends Lead>(
  type: LeadType,
  data: Partial<T>
): Promise<T> => {
  return apiPost<T>(`/leads/${type}`, sanitize(data));
};

/**
 * Get single lead by ID
 */
export const getLead = async <T extends Lead>(
  type: LeadType,
  id: string
): Promise<T> => {
  return apiGet<T>(`/leads/${type}/${id}`);
};

/**
 * Update lead
 */
export const updateLead = async <T extends Lead>(
  type: LeadType,
  id: string,
  data: Partial<T>
): Promise<T> => {
  return apiPut<T>(`/leads/${type}/${id}`, sanitize(data));
};

/**
 * Delete lead
 */
export const deleteLead = async (
  type: LeadType,
  id: string
): Promise<void> => {
  await apiDelete(`/leads/${type}/${id}`);
};

/**
 * Get statuses for specific lead type
 */
export const getLeadStatuses = async (
  type: LeadType
): Promise<LeadStatus[]> => {
  return apiGet<LeadStatus[]>(`/leads/${type}/statuses`);
};

// Backward-compatible wrappers (deprecated)
export const getJobs = (page: number, limit: number, ...args) =>
  getLeads<JobLead>('job', page, limit, ...args);

export const createJob = (job: Partial<JobLead>) =>
  createLead<JobLead>('job', job);

export const updateJob = (id: string, job: Partial<JobLead>) =>
  updateLead<JobLead>('job', id, job);

export const deleteJob = (id: string) =>
  deleteLead('job', id);

// Job-specific operations
export const getRecentResumeDate = async (): Promise<string | null> => {
  const data = await apiGet<{ resume_date?: string }>("/leads/job/recent-resume-date");
  return data.resume_date || null;
};

export const markLeadAsApplied = async (jobId: string): Promise<JobLead> => {
  return apiPost<JobLead>(`/leads/job/${jobId}/apply`, {});
};

export const markLeadAsDoNotApply = async (jobId: string): Promise<JobLead> => {
  return apiPost<JobLead>(`/leads/job/${jobId}/reject`, {});
};
```

**Checklist:**
- [ ] Implement generic `getLeads<T>()` function
- [ ] Implement generic `createLead<T>()` function
- [ ] Implement generic `updateLead<T>()` function
- [ ] Implement generic `deleteLead()` function
- [ ] Add backward-compatible wrapper functions
- [ ] Update job-specific operations to use new paths
- [ ] Add JSDoc comments

#### Step 3: Update Components

**JobTracker (main table view):**
```typescript
// components/JobTracker/JobTracker.tsx

import { getLeads, createLead, updateLead, deleteLead } from '@/api';
import { JobLead } from '@/types/leads';

export function JobTracker() {
  const [jobs, setJobs] = useState<JobLead[]>([]);

  const loadJobs = async () => {
    const data = await getLeads<JobLead>('job', page, limit, orderBy, order, search);
    setJobs(data.items);
  };

  const handleCreate = async (jobData: Partial<JobLead>) => {
    const newJob = await createLead<JobLead>('job', jobData);
    setJobs([...jobs, newJob]);
  };

  // ... rest of component
}
```

**Generic LeadTracker:**
```typescript
// components/LeadTracker/LeadTracker.tsx

interface LeadTrackerProps<T extends Lead> {
  leadType: LeadType;
  columns: ColumnDef<T>[];
  // ... other props
}

export function LeadTracker<T extends Lead>({ leadType, ...props }: LeadTrackerProps<T>) {
  const [leads, setLeads] = useState<T[]>([]);

  const loadLeads = async () => {
    const data = await getLeads<T>(leadType, page, limit);
    setLeads(data.items);
  };

  // ... generic component logic
}

// Usage:
<LeadTracker<JobLead> leadType="job" columns={jobColumns} />
<LeadTracker<OpportunityLead> leadType="opportunity" columns={oppColumns} />
```

**Checklist:**
- [ ] Update `JobTracker` to use new API
- [ ] Update `LeadPanel` to use new API
- [ ] Make `LeadTracker` fully generic
- [ ] Create `OpportunityTracker` using generic component
- [ ] Create `PartnershipTracker` using generic component
- [ ] Update all forms to handle type parameter

### 5.3 Documentation

**API Documentation (OpenAPI/Swagger):**
- [ ] Document all new `/leads/:type` endpoints
- [ ] Mark old endpoints as deprecated
- [ ] Add request/response examples for each type
- [ ] Document type validation errors

**Migration Guide:**
- [ ] Create `MIGRATION.md` with step-by-step instructions
- [ ] Provide before/after code examples
- [ ] List all breaking changes
- [ ] Provide migration timeline

**Architecture Documentation:**
- [ ] Update ERD with inheritance pattern highlighted
- [ ] Create sequence diagrams for each operation
- [ ] Document type dispatch pattern
- [ ] Add code examples to README

---

## 6. Performance Considerations

### 6.1 Query Optimization

**Joined Table Inheritance requires JOINs for every query:**

```sql
-- Every job query needs this JOIN
SELECT * FROM Lead
JOIN Job ON Job.id = Lead.id
WHERE Lead.type = 'Job'
```

**Optimization strategies:**

1. **Indexes:**
   ```sql
   CREATE INDEX idx_lead_type ON Lead(type);
   CREATE INDEX idx_lead_type_status ON Lead(type, current_status_id);
   ```

2. **Eager Loading:**
   ```python
   # Always eager load type-specific table
   stmt = select(Lead).options(joinedload(Lead.job))
   ```

3. **Query Caching:**
   - Cache frequently accessed lead lists
   - Invalidate on create/update/delete

4. **Pagination:**
   - Always use LIMIT/OFFSET
   - Consider cursor-based pagination for large datasets

### 6.2 Monitoring

**Metrics to track:**
- Query execution time by lead type
- Number of JOINs per request
- Cache hit/miss ratio
- API response times by endpoint

**Tools:**
- SQLAlchemy query logging (`echo=True`)
- FastAPI middleware for timing
- APM tools (DataDog, New Relic)

---

## 7. Security Considerations

### 7.1 Type Validation

**Prevent injection via type parameter:**

```python
# GOOD - Use Enum
from enum import Enum

class LeadType(str, Enum):
    JOB = "job"
    OPPORTUNITY = "opportunity"
    PARTNERSHIP = "partnership"

@router.get("/leads/{lead_type}")
async def get_leads(lead_type: LeadType):
    # FastAPI validates automatically
    pass

# BAD - Direct string parameter
@router.get("/leads/{lead_type}")
async def get_leads(lead_type: str):
    # No validation - vulnerable
    leads = await repo.find_leads_by_type(lead_type)  # SQL injection risk
```

### 7.2 Tenant Isolation

**Ensure multi-tenant data isolation:**

```python
async def get_leads(lead_type: str, tenant_id: int, ...):
    # ALWAYS filter by tenant_id from auth context
    stmt = (
        select(Lead)
        .join(Deal)
        .where(Lead.type == lead_type.capitalize())
        .where(Deal.tenant_id == tenant_id)  # CRITICAL
    )
```

### 7.3 Authorization

**Role-based access control:**

```python
@require_auth
@require_role(["owner", "admin", "member"])
async def get_leads_route(...):
    # User must have appropriate role
    pass
```

---

## 8. Extensibility

### 8.1 Adding New Lead Types

**Steps to add a new lead type (e.g., "event"):**

1. **Create model:**
   ```python
   # models/Event.py
   class Event(Base):
       __tablename__ = "Event"
       id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
       event_name = Column(Text, nullable=False)
       event_date = Column(Date)
       # ... event-specific fields

       lead = relationship("Lead", back_populates="event")
   ```

2. **Update Lead model:**
   ```python
   # models/Lead.py
   class Lead(Base):
       # ... existing fields
       event = relationship("Event", back_populates="lead", uselist=False)
   ```

3. **Create repository:**
   ```python
   # repositories/event_repository.py
   async def create_event(data): ...
   async def find_event_by_id(event_id): ...
   ```

4. **Create service:**
   ```python
   # services/event_service.py
   async def get_all_events(...): ...
   async def create_event(...): ...
   ```

5. **Update LeadType enum:**
   ```python
   class LeadType(str, Enum):
       JOB = "job"
       OPPORTUNITY = "opportunity"
       PARTNERSHIP = "partnership"
       EVENT = "event"  # NEW
   ```

6. **Update service map:**
   ```python
   # services/lead_service.py
   _SERVICE_MAP = {
       "job": job_service,
       "opportunity": opportunity_service,
       "partnership": partnership_service,
       "event": event_service,  # NEW
   }
   ```

7. **Frontend:**
   - Add `EventLead` interface
   - Add to `Lead` union type
   - Endpoints automatically work!

**No changes needed to:**
- Routes (parameterized)
- Controller (dispatches automatically)
- Main service layer (uses service map)

### 8.2 Type-Specific Operations

**Adding custom operations for specific types:**

```python
# routes/leads.py

@router.post("/event/{lead_id}/register")
async def register_for_event(lead_id: str):
    """Event-specific operation: register attendee."""
    return await lead_controller.register_for_event(lead_id)

@router.post("/opportunity/{lead_id}/close")
async def close_opportunity(lead_id: str):
    """Opportunity-specific operation: close deal."""
    return await lead_controller.close_opportunity(lead_id)
```

---

## 9. Appendices

### Appendix A: Complete Endpoint Reference

#### Current Endpoints (Before Migration)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/jobs` | List all jobs (paginated) |
| POST | `/jobs` | Create new job |
| GET | `/jobs/{job_id}` | Get single job |
| PUT | `/jobs/{job_id}` | Update job |
| DELETE | `/jobs/{job_id}` | Delete job |
| GET | `/jobs/recent-resume-date` | Get most recent resume date |
| GET | `/leads` | List job leads (filtered by status) |
| GET | `/leads/statuses` | Get lead statuses |
| POST | `/leads/{job_id}/apply` | Mark lead as applied |
| POST | `/leads/{job_id}/do-not-apply` | Mark lead as rejected |

#### New Endpoints (After Migration)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/leads/{type}` | List leads of type (paginated) |
| POST | `/leads/{type}` | Create new lead of type |
| GET | `/leads/{type}/{id}` | Get single lead |
| PUT | `/leads/{type}/{id}` | Update lead |
| DELETE | `/leads/{type}/{id}` | Delete lead |
| GET | `/leads/{type}/statuses` | Get statuses for type |
| GET | `/leads/job/recent-resume-date` | Get most recent resume date (job-specific) |
| POST | `/leads/job/{id}/apply` | Mark job lead as applied |
| POST | `/leads/job/{id}/reject` | Mark job lead as rejected |

### Appendix B: Database Schema SQL

```sql
-- Lead base table
CREATE TABLE Lead (
    id BIGSERIAL PRIMARY KEY,
    deal_id BIGINT REFERENCES Deal(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('Job', 'Opportunity', 'Partnership')),
    title TEXT NOT NULL,
    description TEXT,
    source TEXT,
    current_status_id BIGINT REFERENCES Status(id) ON DELETE SET NULL,
    owner_user_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_lead_type ON Lead(type);
CREATE INDEX idx_lead_deal ON Lead(deal_id);
CREATE INDEX idx_lead_status ON Lead(current_status_id);

-- Job child table
CREATE TABLE Job (
    id BIGINT PRIMARY KEY REFERENCES Lead(id) ON DELETE CASCADE,
    organization_id BIGINT NOT NULL REFERENCES Organization(id) ON DELETE CASCADE,
    job_title TEXT NOT NULL,
    job_url TEXT,
    notes TEXT,
    resume_date TIMESTAMP WITH TIME ZONE,
    salary_range TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_job_organization ON Job(organization_id);

-- Opportunity child table
CREATE TABLE Opportunity (
    id BIGINT PRIMARY KEY REFERENCES Lead(id) ON DELETE CASCADE,
    organization_id BIGINT NOT NULL REFERENCES Organization(id) ON DELETE CASCADE,
    opportunity_name TEXT NOT NULL,
    estimated_value NUMERIC(18, 2),
    probability NUMERIC(5, 2),
    expected_close_date DATE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Partnership child table
CREATE TABLE Partnership (
    id BIGINT PRIMARY KEY REFERENCES Lead(id) ON DELETE CASCADE,
    organization_id BIGINT NOT NULL REFERENCES Organization(id) ON DELETE CASCADE,
    partnership_name TEXT NOT NULL,
    partnership_type TEXT,
    start_date DATE,
    end_date DATE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Appendix C: Code Examples

#### Example 1: Creating a Job Lead

**Old API:**
```bash
POST /jobs
Content-Type: application/json

{
  "company": "Acme Corp",
  "job_title": "Software Engineer",
  "job_url": "https://acme.com/jobs/123",
  "salary_range": "$100k-$150k",
  "status": "lead",
  "source": "agent"
}
```

**New API:**
```bash
POST /leads/job
Content-Type: application/json

{
  "company": "Acme Corp",
  "job_title": "Software Engineer",
  "job_url": "https://acme.com/jobs/123",
  "salary_range": "$100k-$150k",
  "status": "lead",
  "source": "agent"
}
```

#### Example 2: Listing Job Leads

**Old API:**
```bash
GET /jobs?page=1&limit=25&orderBy=date&order=DESC&search=engineer
```

**New API:**
```bash
GET /leads/job?page=1&limit=25&orderBy=date&order=DESC&search=engineer
```

#### Example 3: Frontend Usage

**Before:**
```typescript
import { getJobs, createJob } from '@/api';

const JobList = () => {
  const loadJobs = async () => {
    const data = await getJobs(1, 25, 'date', 'DESC');
    setJobs(data.items);
  };

  const handleCreate = async (jobData) => {
    await createJob(jobData);
    loadJobs();
  };
};
```

**After:**
```typescript
import { getLeads, createLead } from '@/api';
import { JobLead } from '@/types/leads';

const JobList = () => {
  const loadJobs = async () => {
    const data = await getLeads<JobLead>('job', 1, 25, 'date', 'DESC');
    setJobs(data.items);
  };

  const handleCreate = async (jobData: Partial<JobLead>) => {
    await createLead<JobLead>('job', jobData);
    loadJobs();
  };
};
```

---

## 10. Timeline & Milestones

### Week 1-2: Foundation & Dual Support
- [ ] **Day 1-2:** Create unified repositories
- [ ] **Day 3-4:** Create unified services
- [ ] **Day 5-6:** Create unified controllers
- [ ] **Day 7-8:** Create parameterized routes
- [ ] **Day 9-10:** Register new routes (dual support enabled)

### Week 3-4: Frontend Migration & Testing
- [ ] **Day 11-12:** Update TypeScript types
- [ ] **Day 13-14:** Update API client functions
- [ ] **Day 15-16:** Migrate JobTracker component
- [ ] **Day 17-18:** Migrate other components
- [ ] **Day 19-20:** Add deprecation warnings
- [ ] **Day 21-22:** Integration testing
- [ ] **Day 23-24:** User acceptance testing

### Week 5: Cleanup & Documentation
- [ ] **Day 25:** Remove old endpoints
- [ ] **Day 26:** Remove deprecated frontend code
- [ ] **Day 27:** Update documentation
- [ ] **Day 28:** Performance testing
- [ ] **Day 29:** Final QA
- [ ] **Day 30:** Deploy to production

---

## 11. Success Criteria

### Technical Metrics:
- ✅ All endpoints follow `/leads/:type` pattern
- ✅ Zero breaking changes for existing API consumers (during dual-support)
- ✅ Query performance within 10% of current baseline
- ✅ 100% test coverage for new endpoints
- ✅ All TypeScript types properly defined
- ✅ No console errors or warnings

### Business Metrics:
- ✅ Zero downtime during migration
- ✅ No customer-reported issues
- ✅ Support for all three lead types (job, opportunity, partnership)
- ✅ Clear migration path documented

### Code Quality:
- ✅ Reduced code duplication (single controller/service pattern)
- ✅ Consistent naming conventions
- ✅ Comprehensive documentation
- ✅ Easy to add new lead types

---

## 12. Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Breaking frontend during migration | High | Medium | Implement dual support, gradual rollout |
| Performance degradation from JOINs | Medium | Medium | Add indexes, optimize queries, monitor performance |
| Incomplete migration leaves orphaned code | Medium | Low | Clear checklist, code review, automated cleanup |
| Type validation bypass | High | Low | Use Enum, add tests, security review |
| Tenant data leakage | Critical | Very Low | Always filter by tenant_id, add tests |

---

## Conclusion

This specification provides a comprehensive plan to refactor the API to use parameterized leads structure (`/leads/:type`). The migration is designed to be **incremental**, **backward-compatible**, and **low-risk**, with clear rollback procedures at each phase.

**Key Takeaways:**
1. Database schema already supports this pattern (JTI)
2. Migration is primarily a routing/service refactor
3. Dual-support phase ensures zero downtime
4. Frontend changes are minimal (wrapper functions)
5. Design is extensible for future lead types

**Next Steps:**
1. Review and approve this specification
2. Create implementation tickets from checklist
3. Begin Phase 1: Dual Support implementation
4. Schedule migration timeline with stakeholders

---

**Document Version:** 1.0
**Last Updated:** 2025-01-16
**Author:** Development Team
**Status:** Draft - Awaiting Approval

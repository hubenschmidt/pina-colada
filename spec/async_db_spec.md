# Async Database Migration Specification

## Overview
Migrate job controller, service, and repository from synchronous to asynchronous architecture using SQLAlchemy async features with database-level pagination.

## Goals
1. Full async/await pattern across all layers
2. AsyncSession + create_async_engine for database operations
3. Remove module-level caching entirely
4. Add database-level pagination, filtering, and sorting
5. Apply functional programming patterns (guard clauses, no nested conditionals)

---

## Database Configuration

### File: `modules/agent/src/lib/db.py`

#### Add Async Engine and Session
```python
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker

# Create async engine
async_engine = create_async_engine(
    DATABASE_URL.replace('sqlite:///', 'sqlite+aiosqlite:///'),
    echo=False,
    future=True
)

# Create async session factory
AsyncSessionLocal = async_sessionmaker(
    async_engine,
    class_=AsyncSession,
    expire_on_commit=False
)

async def async_get_session():
    """Async context manager for database sessions."""
    async with AsyncSessionLocal() as session:
        yield session
```

#### Dependencies Required
- `sqlalchemy[asyncio]>=2.0`
- `aiosqlite` (for SQLite) or `asyncpg` (for PostgreSQL)

---

## Repository Layer

### File: `modules/agent/src/repositories/job_repository.py`

#### Changes Summary
- Convert all 11 functions to `async def`
- Replace `get_session()` → `async with async_get_session() as session:`
- Replace `session.execute()` → `await session.execute()`
- Remove `_cache` and `_details_cache` globals (lines 22-24)
- Add database-level pagination to `find_all_jobs()`

#### Key Function: `find_all_jobs()` - Database-Level Pagination

**Current (Synchronous, In-Memory):**
```python
def find_all_jobs() -> List[Job]:
    """Find all jobs with related data."""
    session = get_session()
    try:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            # ... relationships
        ).join(Lead).order_by(Lead.created_at.desc())
        return list(session.execute(stmt).unique().scalars().all())  # Loads ALL rows
    finally:
        session.close()
```

**New (Async, DB-Level Pagination):**
```python
async def find_all_jobs(
    page: int = 1,
    page_size: int = 50,
    search: Optional[str] = None,
    order_by: str = "date",
    order: str = "DESC"
) -> tuple[List[Job], int]:
    """Find jobs with pagination, filtering, and sorting at database level."""
    async with async_get_session() as session:
        # Base query with relationships
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).join(Lead)

        # Apply search filter at DB level
        if search and search.strip():
            search_lower = search.strip().lower()
            stmt = stmt.where(
                sql_func.lower(Organization.name).contains(search_lower) |
                sql_func.lower(Job.job_title).contains(search_lower)
            ).join(Organization)

        # Get total count before pagination
        count_stmt = select(sql_func.count()).select_from(stmt.subquery())
        total_result = await session.execute(count_stmt)
        total_count = total_result.scalar() or 0

        # Apply sorting at DB level
        sort_map = {
            "date": Lead.created_at,
            "application_date": Lead.created_at,
            "company": Organization.name,
            "job_title": Job.job_title,
            "resume": Job.resume_date,
        }
        sort_column = sort_map.get(order_by, Lead.created_at)
        stmt = stmt.order_by(sort_column.desc() if order.upper() == "DESC" else sort_column.asc())

        # Apply pagination at DB level
        offset = (page - 1) * page_size
        stmt = stmt.limit(page_size).offset(offset)

        # Execute query
        result = await session.execute(stmt)
        jobs = list(result.unique().scalars().all())

        return jobs, total_count
```

#### Other Functions to Convert

All functions follow the same pattern:

```python
# OLD
def find_job_by_id(job_id: int) -> Optional[Job]:
    session = get_session()
    try:
        stmt = select(Job).where(Job.id == job_id)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()

# NEW
async def find_job_by_id(job_id: int) -> Optional[Job]:
    async with async_get_session() as session:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            # ... relationships
        ).where(Job.id == job_id)
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()
```

**Functions to convert:**
1. `find_all_jobs()` - add pagination params
2. `count_jobs()`
3. `find_job_by_id()`
4. `create_job()`
5. `update_job()`
6. `delete_job()`
7. `find_job_by_company_and_title()`
8. `find_all_statuses()`
9. `find_jobs_with_status()`
10. `_load_job_with_relationships()`

---

## Service Layer

### File: `modules/agent/src/services/job_service.py`

#### Changes Summary
- Convert all 21 functions to `async def`
- Add `await` to all repository calls
- Remove caching (lines 22-24, 59-86, 182, 367, 418, 536, 574)
- Simplify `get_jobs_paginated()` to delegate to repository
- Refactor nested conditionals with guard clauses

#### Key Function: `get_jobs_paginated()`

**Current (In-Memory Pagination):**
```python
def get_jobs_paginated(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None
) -> tuple[List[Any], int]:
    # Get all jobs
    all_jobs = find_all_jobs_repo()

    # Apply search filter in Python
    if search and search.strip():
        search_lower = search.strip().lower()
        all_jobs = [
            job for job in all_jobs
            if (job.organization and search_lower in job.organization.name.lower())
            or (job.job_title and search_lower in job.job_title.lower())
        ]

    total_count = len(all_jobs)

    # Sort in Python
    sort_functions = {
        "application_date": lambda j: j.lead.created_at if j.lead else "",
        # ... more sort functions
    }
    sort_fn = sort_functions.get(order_by)
    if sort_fn:
        all_jobs.sort(key=sort_fn, reverse=reverse)

    # Paginate in Python
    offset = (page - 1) * limit
    paginated_jobs = all_jobs[offset:offset + limit]

    return paginated_jobs, total_count
```

**New (Delegate to Repository):**
```python
async def get_jobs_paginated(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None
) -> tuple[List[Any], int]:
    """Get jobs with pagination - delegates to repository for DB-level operations."""
    return await find_all_jobs_repo(
        page=page,
        page_size=limit,
        search=search,
        order_by=order_by,
        order=order
    )
```

#### Guard Clause Refactoring Examples

**Current (Nested Conditionals):**
```python
def _update_lead_status(lead: Lead, data: Dict[str, Any]) -> None:
    if "current_status_id" in data:
        lead.current_status_id = data["current_status_id"]
        return

    if "status" not in data:
        return
    if not data["status"]:
        return

    status_id = _get_status_id_from_name(data["status"])
    if status_id:
        lead.current_status_id = status_id
```

**Refactored (Guard Clauses):**
```python
async def _update_lead_status(lead: Lead, data: Dict[str, Any]) -> None:
    if "current_status_id" in data:
        lead.current_status_id = data["current_status_id"]
        return

    if "status" not in data:
        return

    if not data["status"]:
        return

    status_id = await _get_status_id_from_name(data["status"])
    if not status_id:
        return

    lead.current_status_id = status_id
```

#### Functions to Convert (21 total)

**Core CRUD:**
1. `get_all_jobs()` - remove cache
2. `get_applied_jobs_only()` - remove cache
3. `get_applied_identifiers()` - remove cache
4. `fetch_applied_jobs()` - remove cache
5. `get_jobs_details()` - remove cache
6. `is_job_applied()`
7. `filter_jobs()`
8. `add_job()` - remove `_clear_cache()`
9. `add_applied_job()`
10. `get_jobs_paginated()` - simplify
11. `create_job()` - remove `_clear_cache()`
12. `get_job()`
13. `delete_job()` - remove `_clear_cache()`
14. `get_statuses()`
15. `get_jobs_with_status()`
16. `get_recent_resume_date()`
17. `update_job()` - remove `_clear_cache()`
18. `update_job_by_company()` - remove `_clear_cache()`

**Helpers:**
19. `_map_to_dict()`
20. `_fuzzy_match_company()`
21. `_matches_job()`

**Remove entirely:**
- `_clear_cache()` (lines 59-63)
- Global `_cache`, `_details_cache` (lines 22-24)

---

## Controller Layer

### File: `modules/agent/src/controllers/job_controller.py`

#### Changes Summary
- Convert all 10 functions to `async def`
- Add `await` to all service calls

#### Example Conversion

**Current:**
```python
@handle_http_exceptions
def get_jobs(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None
) -> dict:
    """Get all jobs with pagination."""
    paginated_jobs, total_count = get_jobs_paginated(
        page, limit, order_by, order, search
    )
    items = [_job_to_response_dict(job) for job in paginated_jobs]
    return _to_paged_response(total_count, page, limit, items)
```

**New:**
```python
@handle_http_exceptions
async def get_jobs(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None
) -> dict:
    """Get all jobs with pagination."""
    paginated_jobs, total_count = await get_jobs_paginated(
        page, limit, order_by, order, search
    )
    items = [_job_to_response_dict(job) for job in paginated_jobs]
    return _to_paged_response(total_count, page, limit, items)
```

#### Functions to Convert (10)
1. `get_jobs()`
2. `create_job()`
3. `get_job()`
4. `update_job()`
5. `delete_job()`
6. `get_statuses()`
7. `get_leads()`
8. `mark_lead_as_applied()`
9. `mark_lead_as_do_not_apply()`
10. `get_recent_resume_date()`

---

## Routes Layer

### File: `modules/agent/src/api/routes/jobs.py`

#### Changes Summary
- Add `await` to all controller calls (routes already use `async def`)

#### Example Conversion

**Current:**
```python
@router.get("")
@log_errors
@require_auth
async def get_jobs_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(25, ge=1, le=100),
    order_by: str = Query("date", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
):
    """Get all jobs with pagination."""
    return get_jobs(page, limit, order_by, order, search)
```

**New:**
```python
@router.get("")
@log_errors
@require_auth
async def get_jobs_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(25, ge=1, le=100),
    order_by: str = Query("date", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
):
    """Get all jobs with pagination."""
    return await get_jobs(page, limit, order_by, order, search)
```

#### Routes to Update (5)
1. `get_jobs_route()` - add `await`
2. `create_job_route()` - add `await`
3. `get_job_route()` - add `await`
4. `update_job_route()` - add `await`
5. `delete_job_route()` - add `await`
6. `get_recent_resume_date_route()` - add `await`

---

## Code Quality Standards

### Functional Programming Patterns

1. **Guard Clauses** - Early returns, no nested if/else
2. **Pure Functions** - Helper functions with no side effects
3. **No break/continue** - Use small functions with early returns
4. **KISS** - Keep it simple
5. **YAGNI** - No speculative features

### Example: Guard Clause Pattern

```python
# BAD - Nested conditionals
if condition1:
    if condition2:
        if condition3:
            do_something()
        else:
            handle_error()
    else:
        handle_error()
else:
    handle_error()

# GOOD - Guard clauses
if not condition1:
    handle_error()
    return

if not condition2:
    handle_error()
    return

if not condition3:
    handle_error()
    return

do_something()
```

---

## Testing Strategy

### Verification Steps
1. Verify async engine connects successfully
2. Test pagination with different page sizes
3. Test search filtering at DB level
4. Test sorting by different columns
5. Verify transaction handling (commit/rollback)
6. Check relationship loading (no N+1 queries)
7. Confirm cache removal doesn't break functionality

### Performance Expectations
- Reduced memory usage (no loading all rows)
- Faster response times for large datasets
- Better database query optimization
- Proper connection pooling via async engine

---

## Migration Checklist

- [ ] Install async SQLAlchemy dependencies
- [ ] Create async engine and session factory in `lib/db.py`
- [ ] Convert repository layer (11 functions)
- [ ] Add database-level pagination to `find_all_jobs()`
- [ ] Convert service layer (21 functions)
- [ ] Remove all caching code
- [ ] Convert controller layer (10 functions)
- [ ] Update routes layer (5 routes)
- [ ] Apply guard clause refactoring
- [ ] Test pagination, search, sorting
- [ ] Verify no regressions in existing functionality

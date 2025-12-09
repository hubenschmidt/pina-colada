# API Endpoint Schema Contract Tests

## Overview
Create one test per API endpoint to verify request schema contract adherence.

## Scope

| Metric | Count |
|--------|-------|
| Total API Endpoints | 145 |
| Route Files | 29 |
| With Pydantic Request Schemas | ~110 endpoints |
| Without Schema Validation | ~35 endpoints |

### Breakdown by Complexity
- **Simple GET (list/lookup)**: ~80 endpoints - verify response structure
- **POST/PUT with schemas**: ~58 endpoints - verify request validation + response
- **DELETE**: ~7 endpoints - verify success/404 responses

## Current State
- **Existing tests**: WebSocket/agent flow only - NO API endpoint tests
- **Test infrastructure**: pytest + pytest-asyncio configured
- **Missing**: TestClient setup, test database, fixtures, mocking

## Time Estimate
- **Optimistic**: 22-25 hours
- **Realistic**: 30-35 hours (~4-5 days)
- **Conservative**: 40+ hours

## Implementation Decisions

1. **Mock vs Integration**: Hybrid - mock repository layer for most tests
2. **Auth handling**: Skip auth via FastAPI dependency override
3. **Response schemas**: Not adding - focus on request schema contracts only

## Implementation Plan

### Phase 1: Test Infrastructure Setup
1. Create `tests/conftest.py` with:
   - TestClient setup using `httpx.AsyncClient`
   - Auth bypass via `app.dependency_overrides`
   - Mock repository fixtures

2. Create `tests/utils/mocks.py` with:
   - Generic repository mock factory
   - Common test data builders

### Phase 2: Test Files (one per route file)
Create 29 test files in `tests/api/`:

| Test File | Route File | Endpoints |
|-----------|------------|-----------|
| test_organizations.py | organizations.py | 17 |
| test_individuals.py | individuals.py | 11 |
| test_leads.py | leads.py | 8 |
| test_opportunities.py | opportunities.py | 5 |
| test_partnerships.py | partnerships.py | 5 |
| test_projects.py | projects.py | 7 |
| test_tasks.py | tasks.py | 8 |
| test_documents.py | documents.py | 12 |
| test_notes.py | notes.py | 5 |
| test_comments.py | comments.py | 5 |
| test_notifications.py | notifications.py | 4 |
| test_contacts.py | contacts.py | 5 |
| test_jobs.py | jobs.py | 7 |
| test_users.py | users.py | 4 |
| test_reports.py | reports.py | 14 |
| test_preferences.py | preferences.py | 5 |
| test_conversations.py | conversations.py | 7 |
| test_accounts.py | accounts.py | 3 |
| (+ 11 smaller route files) | | ~24 |

### Test Pattern
```python
@pytest.mark.asyncio
async def test_create_organization_validates_schema(client, mock_org_repo):
    # Invalid payload - missing required field
    response = await client.post("/api/organizations", json={})
    assert response.status_code == 422

    # Valid payload
    response = await client.post("/api/organizations", json={"name": "Test"})
    assert response.status_code in (200, 201)
```

### Files to Create
- `tests/conftest.py` - shared fixtures
- `tests/api/__init__.py`
- `tests/api/test_*.py` - 29 test files

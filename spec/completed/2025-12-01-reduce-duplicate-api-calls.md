# Reduce Duplicate API Calls & Fix Connection Pool

**Completed:** 2025-12-01

## Problem
- Comments stuck on "loading" in production (database connection pool exhaustion)
- Excessive duplicate API calls in frontend (lookups, tokens, notifications)
- React Strict Mode causing double-mounts in development

## Changes

### Backend (modules/agent)

**`src/lib/db.py`** - Database connection pool settings for pgbouncer
- Added `pool_size=5`, `max_overflow=0`, `pool_timeout=30`
- Prevents connection pool exhaustion with Supabase pgbouncer in transaction mode

### Frontend (modules/client)

**`next.config.js`** - NEW - Disabled React Strict Mode
- Stops duplicate component mounts in development

**`lib/lookup-cache.ts`** - NEW - Module-level fetch deduplication
- `fetchOnce(key, fetcher)` prevents duplicate concurrent requests
- Caches results for session lifetime

**`lib/fetch-bearer-token.ts`** - Token caching
- Module-level token cache with expiry
- Deduplicates concurrent token requests

**`context/lookupsContext.tsx`** - NEW - Centralized lookups state
- Simple context exposing `lookupsState` and `dispatchLookups`
- Follows navContext/navReducer pattern

**`reducers/lookupsReducer.ts`** - NEW - Lookups state management
- Handles industries, salary ranges, projects
- Exports action type strings (SET_INDUSTRIES, SET_SALARY_RANGES, etc.)

**`components/NotificationBell/NotificationBell.tsx`** - Removed polling
- Changed from 30-second interval to fetch on route change only

**`components/AuthStateManager/AuthStateManager.tsx`** - Single Auth0 source
- Only component that calls Auth0's `useUser()`
- Other components use `useUserContext()` instead

**Updated components to use lookupsContext:**
- `components/AccountForm/IndustrySelector.tsx`
- `components/LeadTracker/hooks/useLeadFormConfig.tsx` (ProjectSelector, SalaryRangeSelector)

**Updated components to use userContext instead of Auth0 useUser:**
- Header.tsx, AuthenticatedLayout.tsx, AuthenticatedPageLayout.tsx
- login/page.tsx, tenant/select/page.tsx, RootLayoutClient.tsx

## Results
- Lookups (industries, projects, salary-ranges): reduced to 1 call each
- Token fetches: deduplicated via module-level cache
- Notifications: 1 call per route change (expected behavior)
- Development: no more Strict Mode double-mounts

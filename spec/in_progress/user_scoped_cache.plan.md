# User-Scoped Research Cache with Clear Option

## Problem
- Research_Cache is currently tenant-scoped (all users share cache)
- No way to clear cache from the frontend
- Job searches for specific individuals shouldn't be shared across users

## Solution
1. Change cache scope from tenant to user
2. Add clear cache UI in settings page and chat UI

---

## Backend Changes

### 1. Update `internal/repositories/research_cache_repository.go`

Change lookups from `tenant_id` to `created_by` (user_id):

```go
// Before
func (r *ResearchCacheRepository) Lookup(tenantID int64, cacheKey string) (*ResearchCacheDTO, error) {
    err := r.db.Where("tenant_id = ? AND cache_key = ? AND expires_at > ?", tenantID, cacheKey, time.Now())

// After
func (r *ResearchCacheRepository) Lookup(userID int64, cacheKey string) (*ResearchCacheDTO, error) {
    err := r.db.Where("created_by = ? AND cache_key = ? AND expires_at > ?", userID, cacheKey, time.Now())
```

Update all methods:
- `Lookup(userID, cacheKey)`
- `Upsert(input)` - ensure `CreatedBy` is required
- `ListRecent(userID, cacheType, limit)`
- `DeleteByUser(userID)` - new method
- `DeleteByUserAndType(userID, cacheType)` - new method

### 2. Update cache tool callers

Files using cache repository:
- `internal/agent/tools/cache_tools.go`
- `internal/agent/tools/serper_tools.go`

Change from passing `tenantID` to `userID`.

### 3. Add cache controller

Create `internal/controllers/cache_controller.go`:

```go
// DELETE /api/cache
// DELETE /api/cache?type=job_search
func (c *CacheController) ClearCache(w http.ResponseWriter, r *http.Request)
```

### 4. Add route

In `internal/routes/router.go`:

```go
r.Delete("/cache", c.Cache.ClearCache)
```

---

## Frontend Changes

### 1. Add API function

In `modules/client/api/index.js`:

```javascript
export const clearCache = async (cacheType = null) => {
  const params = cacheType ? `?type=${cacheType}` : '';
  return apiDelete(`/cache${params}`);
};
```

### 2. Settings page

Add cache section to settings with:
- "Clear All Cache" button
- Optional: dropdown to clear by type (job_search, web_fetch, company_research)
- Show cache stats (count, oldest entry)

### 3. Chat UI

Add small icon button (trash/broom icon) near chat input:
- Tooltip: "Clear research cache"
- Confirmation dialog before clearing

---

## Files to Modify

### Backend
- `internal/repositories/research_cache_repository.go`
- `internal/agent/tools/cache_tools.go`
- `internal/agent/tools/serper_tools.go`
- `internal/routes/router.go`

### Backend (New)
- `internal/controllers/cache_controller.go`

### Frontend
- `modules/client/api/index.js`
- `modules/client/app/settings/page.jsx` (or similar)
- `modules/client/components/Chat/ChatInput.jsx` (or similar)

---

## Migration

No migration needed - `created_by` column already exists on `Research_Cache` table.

Ensure existing cache entries have `created_by` populated (may need to clear old entries).

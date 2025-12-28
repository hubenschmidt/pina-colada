# Spec: Research Cache with Explicit Agent Tool

## Overview
A PostgreSQL-backed research cache that persists search/web results across requests, with an explicit agent tool for cache interaction. Supports in-memory deduplication within single requests.

## Goals
1. **Cross-request caching** - Avoid redundant API calls (Serper, web fetches)
2. **Single-request dedup** - Prevent duplicate results within one agent turn
3. **Explicit tool access** - Any specialized agent can query/store cached research
4. **Extensibility** - Support future cache types beyond job_search
5. **Per-tenant scoping** - Users in same tenant share cache for collaboration

---

## Database Schema

```sql
CREATE TABLE "Research_Cache" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    cache_key VARCHAR(255) NOT NULL,          -- SHA256 hash of normalized query
    cache_type VARCHAR(50) NOT NULL,          -- 'job_search', 'web_fetch', 'company_research'
    query_params JSONB,                       -- Original search parameters
    result_data JSONB NOT NULL,               -- Cached response data
    result_count INT NOT NULL DEFAULT 0,      -- Number of results
    created_by BIGINT REFERENCES "User"(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(tenant_id, cache_key)              -- Cache key unique per tenant
);

CREATE INDEX idx_research_cache_tenant_key ON "Research_Cache"(tenant_id, cache_key);
CREATE INDEX idx_research_cache_tenant_type ON "Research_Cache"(tenant_id, cache_type);
CREATE INDEX idx_research_cache_expires ON "Research_Cache"(expires_at);
```

---

## Agent Tool: `research_cache`

### Tool Definition
```go
type ResearchCacheParams struct {
    Action    string `json:"action" jsonschema:"required,enum=lookup|store|list_recent,description=Action to perform"`
    CacheType string `json:"cache_type" jsonschema:"description=Type of cache (job_search, web_fetch, etc)"`
    Query     string `json:"query" jsonschema:"description=Search query or URL to lookup/store"`
    Data      string `json:"data" jsonschema:"description=JSON data to store (for store action)"`
    TTLHours  int    `json:"ttl_hours" jsonschema:"description=Time-to-live in hours (default 24)"`
}

type ResearchCacheResult struct {
    Found   bool   `json:"found"`
    Results string `json:"results"`
    Count   int    `json:"count"`
    Age     string `json:"age"`  // "2h ago", "1d ago"
}
```

### Actions
| Action | Description | Use Case |
|--------|-------------|----------|
| `lookup` | Check if query exists in cache | Before making API call |
| `store` | Store results in cache | After successful API call |
| `list_recent` | List recent searches by type | "What have I already searched?" |

### Example Usage by Agent
```
Agent: "Before searching for ML Engineer jobs, let me check if I've already searched this..."
Tool call: research_cache(action="lookup", cache_type="job_search", query="ML Engineer NYC")
Result: {"found": true, "count": 15, "age": "30m ago", "results": "[...]"}
Agent: "I already have 15 results from 30 minutes ago, no need to search again."
```

---

## In-Memory Dedup (Per-Request)

```go
type CacheTools struct {
    cacheRepo    *repositories.ResearchCacheRepository

    // Per-request dedup
    seenKeys     map[string]bool
    seenKeysMu   sync.Mutex
}

func (t *CacheTools) ResetRequestState() {
    t.seenKeysMu.Lock()
    t.seenKeys = make(map[string]bool)
    t.seenKeysMu.Unlock()
}

func (t *CacheTools) MarkSeen(key string) bool {
    t.seenKeysMu.Lock()
    defer t.seenKeysMu.Unlock()
    if t.seenKeys[key] {
        return false // Already seen this request
    }
    t.seenKeys[key] = true
    return true
}
```

---

## Integration Points

### 1. Transparent Integration (Optional)
Existing tools like `job_search` can internally check cache:
```go
func (t *SerperTools) JobSearchCtx(ctx context.Context, params JobSearchParams) (*JobSearchResult, error) {
    cacheKey := t.cacheTools.MakeKey("job_search", params.Query)

    // Check cache first
    if cached, found := t.cacheTools.Lookup(cacheKey); found {
        return cached, nil
    }

    // Execute search...
    result := t.executeSearch(params)

    // Store in cache
    t.cacheTools.Store(cacheKey, result, 24*time.Hour)

    return result, nil
}
```

### 2. Explicit Tool Access
Agents can directly interact via `research_cache` tool for:
- Checking what's been searched before deciding strategy
- Storing custom research results
- Multi-turn research workflows

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `migrations/069_research_cache.sql` | Create table |
| `internal/models/ResearchCache.go` | GORM model |
| `internal/repositories/research_cache_repository.go` | DB operations |
| `internal/agent/tools/cache_tools.go` | Tool implementation |
| `internal/agent/tools/adapter.go` | Register tool |
| `internal/agent/orchestrator.go` | Initialize CacheTools, call ResetRequestState() |

---

## Cache Key Generation
```go
func MakeCacheKey(cacheType, query string) string {
    normalized := strings.ToLower(strings.TrimSpace(query))
    hash := sha256.Sum256([]byte(cacheType + ":" + normalized))
    return hex.EncodeToString(hash[:])
}
```

---

## TTL Strategy
| Cache Type | Default TTL | Rationale |
|------------|-------------|-----------|
| `job_search` | 24 hours | Job listings change daily |
| `web_fetch` | 1 hour | Web content may update frequently |
| `company_research` | 7 days | Company info relatively stable |

---

## Cleanup
Background job or on-demand cleanup of expired entries:
```sql
DELETE FROM "Research_Cache" WHERE expires_at < NOW();
```

---

## Benefits
1. **Reduced API costs** - Avoid duplicate Serper/web calls
2. **Faster responses** - Cache hits return instantly
3. **Agent awareness** - Explicit tool lets agent make informed decisions
4. **Extensible** - New cache types added without code changes
5. **Multi-turn support** - Agent can reference previous research

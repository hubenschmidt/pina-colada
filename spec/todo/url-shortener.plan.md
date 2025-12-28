# URL Shortener Service for Agent Responses

> **Code Rules**: `/go-code-rules.md`, `/csr-code-rules.md`

## Problem
Full URLs in agent responses significantly increase token usage (~100+ chars per URL, 1000+ extra tokens for 10 job results).

## Goal
Replace long URLs with short codes: `1. Company - Title [⭢](/u/abc123)`

## Architecture (CSR Pattern)

```
Route (/u/:code) → URLController → CacheTools (data layer)
```

No separate service needed - this is a simple redirect with no business logic (KISS/YAGNI).

## Implementation

### Step 1: URL Shortener Functions (`cache_tools.go`)

```go
// ShortenURL generates a short code and stores the mapping
func (t *CacheTools) ShortenURL(fullURL string) string {
    code := generateShortCode(fullURL)
    t.storeInMemory("url:"+code, fullURL, 1)
    go t.storeURLInDB(code, fullURL) // async DB persist
    return code
}

// ResolveURL looks up the full URL from a short code
func (t *CacheTools) ResolveURL(code string) string {
    entry := t.getFromMemory("url:" + code)
    if entry != nil {
        return entry.Data
    }
    // DB fallback with guard clause
    cached, err := t.cacheRepo.Lookup(t.tenantID, "url:"+code)
    if err != nil || cached == nil {
        return ""
    }
    var url string
    json.Unmarshal(cached.ResultData, &url)
    return url
}

func generateShortCode(url string) string {
    hash := sha256.Sum256([]byte(url))
    return base62Encode(hash[:6]) // 8 chars
}
```

### Step 2: Update `formatListings` (`serper_tools.go`)

Pass shortener to function, use short codes in output.

### Step 3: Controller (`url_controller.go`)

```go
type URLController struct {
    cacheTools *tools.CacheTools
}

func (c *URLController) Redirect(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")
    if code == "" {
        http.NotFound(w, r)
        return
    }

    fullURL := c.cacheTools.ResolveURL(code)
    if fullURL == "" {
        http.NotFound(w, r)
        return
    }

    http.Redirect(w, r, fullURL, http.StatusFound)
}
```

### Step 4: Route (`routes.go`)

```go
r.Get("/u/{code}", ctrls.URL.Redirect)
```

## Files to Modify

| File | Changes |
|------|---------|
| `internal/agent/tools/cache_tools.go` | Add `ShortenURL()`, `ResolveURL()`, `generateShortCode()`, `base62Encode()` |
| `internal/agent/tools/serper_tools.go` | Update `formatListings()` to accept shortener interface |
| `internal/controllers/url_controller.go` | **NEW** - Redirect handler |
| `internal/routes/routes.go` | Add `/u/{code}` route |
| `cmd/agent/main.go` | Wire URLController |
| `modules/client/hooks/useWs.jsx` | **CLEANUP** - Remove dead `applyUrlMapToLastMessage()` function and `url_map` handling |

## Token Savings

**Before**: `[⭢](https://careers.capitalonebank.com/jobs/senior-software-engineer-full-stack-python-12345)` (~100 chars)
**After**: `[⭢](/u/abc12345)` (~15 chars)
**Savings**: ~85% per URL

## Design Decisions

1. **No service layer**: Simple redirect, no business logic (YAGNI)
2. **Leverage ResearchCache**: No new migration, existing TTL/cleanup
3. **TTL**: 1 hour - URLs only matter during active session
4. **In-memory first**: Fast path, async DB persist
5. **Guard clauses**: Flat conditionals per /go-code-rules

## Checklist
- [ ] Add URL shortening functions to cache_tools.go
- [ ] Add base62 encoding helper
- [ ] Update formatListings to use shortener
- [ ] Create url_controller.go
- [ ] Add /u/{code} route
- [ ] Wire controller in main.go
- [ ] **Cleanup**: Remove dead client code for URL index mapping (---URLS--- parsing)
- [ ] Test: short URLs work, redirect correctly

# Feature: Fetch full job listing text for richer embeddings

## Context

The vectorPreFilter currently embeds only `Title + " - " + Snippet` — the short Google SERP snippet from Serper (~1-2 sentences). This limits embedding accuracy because the snippet often lacks key details (requirements, seniority, tech stack). By fetching the full job page and extracting clean text via go-readability, we get much richer input for both embedding similarity and LLM review.

## All changes in one file
**File**: `modules/agent/internal/services/automation_service.go`
**Dependency**: `github.com/go-shiori/go-readability`

---

### 1. Add `FullText` field to `jobResult` (line ~812)
```go
FullText string // extracted via go-readability; empty if fetch failed
```

### 2. New function: `fetchFullText` (after `validateURLs` ~line 2305)
- Concurrent GET requests (10 goroutines — all results fetched in parallel)
- Calls `extractReadableText` per URL
- **Never filters** — failed fetches leave `FullText` empty (graceful degradation)

### 3. New helper: `extractReadableText`
- HTTP GET, 10s timeout, 500KB body limit (matches `fetchPostingDate` pattern in `serper_tools.go:786`)
- `readability.FromReader` extracts clean text
- Caps at 8000 chars for embedding input
- New imports: `"bytes"`, `"net/url"`, `readability "github.com/go-shiori/go-readability"`

### 4. Wire into pipeline (line ~697, inside `processSearchResultsWithAgent`)
Insert between `validateURLs` and `vectorPreFilter`:
```go
filtered = s.fetchFullText(filtered)
```

### 5. Use `FullText` in `vectorPreFilter` (line ~2324)
```go
texts[i] = r.Title + " - " + r.Snippet
if r.FullText != "" {
    texts[i] = r.FullText
}
```

### 6. Include `FullText` in LLM review prompt (line ~1354, `buildAgentUserPrompt`)
Append `Job Description:` section when `FullText` is available, truncated to 3000 chars for LLM context.

---

## Verification
1. `cd modules/agent && go get github.com/go-shiori/go-readability && go build ./...`
2. Run a crawler — confirm logs show `full text extracted for X/Y results`
3. Verify vectorPreFilter still works (proposals created)
4. Check LLM review prompt includes job description text for fetched results

## Future: Deduplicate page fetches with `fetchPostingDate`

`fetchPostingDate()` (`serper_tools.go:786`) already does a separate GET request per URL during `parseSerperResults` to extract dates via regex. With `fetchFullText` now also fetching the same pages, each URL is fetched twice. A follow-up optimization could pass the already-fetched page body from `fetchFullText` into the date extraction logic, eliminating the redundant request.

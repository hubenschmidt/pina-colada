# HEAD Request URL Validation

## Problem
Crawler proposes jobs with broken/stale links. Currently no URL validation occurs - the LLM only sees Serper metadata (title, company, URL, snippet) without checking if the page still exists.

## Solution
Add HEAD request pre-filter before LLM review to catch broken links cheaply.

## Flow
```
Serper search → URLs
       ↓
HEAD request validation (filter 404/410/5xx)
       ↓
LLM review (only valid URLs)
       ↓
Proposals
```

## Benefits
- No additional Serper API cost (HEAD goes direct to job URLs)
- Saves LLM tokens by filtering broken links before review
- Fast (~50-100ms per URL, can parallelize)

## Implementation

### 1. Add `validateURLs` function in `automation_service.go`
```go
func (s *AutomationService) validateURLs(results []jobResult) []jobResult {
    valid := make([]jobResult, 0, len(results))
    client := &http.Client{Timeout: 5 * time.Second}

    for _, r := range results {
        resp, err := client.Head(r.URL)
        if err != nil {
            log.Printf("Automation: URL validation failed for %s: %v", r.URL, err)
            continue
        }
        resp.Body.Close()

        if resp.StatusCode >= 400 {
            log.Printf("Automation: skipping broken link %s (status %d)", r.URL, resp.StatusCode)
            continue
        }
        valid = append(valid, r)
    }
    return valid
}
```

### 2. Call before LLM review in `executeSearches`
Insert after dedup filtering, before `agentReviewWorker`:
```go
filtered = s.validateURLs(filtered)
if len(filtered) == 0 {
    log.Printf("Automation: all URLs failed validation")
    continue
}
```

### 3. Optional: Parallel validation
Use goroutines + sync.WaitGroup for concurrent HEAD requests (with rate limiting to avoid being blocked).

## Edge Cases
- Redirects: Follow redirects, check final status
- Timeouts: 5s timeout, treat as invalid
- Rate limiting: Some sites may block rapid HEAD requests - consider adding jitter

## Future Enhancements
- Content-based staleness detection (fetch page, look for "position filled")
- Cache validation results to avoid re-checking same URLs
- Allowlist known-good ATS domains (greenhouse.io, lever.co) that rarely have stale links

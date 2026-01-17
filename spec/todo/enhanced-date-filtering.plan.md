# Enhanced Date Filtering for Job Search

## Problem
Current date filtering relies on:
1. Serper API `tbs` parameter (Google's server-side filter)
2. Serper's optional `date` field in organic results
3. Page fetch with JSON-LD `datePosted` parsing
4. Regex patterns for "posted X days ago" text

These miss dates on many pages, leading to stale results when users request recent jobs.

## Proposed Heuristics (Priority Order)

### 1. ATS-Specific Selectors (High Value)
Target the platforms already used in `ATSMode`:
- **Lever**: `.posting-categories .sort-by-time` or JSON in page script
- **Greenhouse**: `.job-meta time` element or embedded JSON
- **Ashby**: `data-posted` attributes or API response in page

### 2. Schema.org validThrough
Parse `validThrough` from JobPosting schema—if expiration is within 30 days, job is likely recent.

### 3. Meta Tags
Check for:
- `<meta property="article:published_time">`
- `<meta property="og:updated_time">`
- `<meta name="date">`

### 4. Apply Button State
Filter out jobs where apply button is disabled or text indicates "closed", "expired", "no longer accepting".

### 5. URL Date Segments
Regex for date patterns in URL path: `/2025/01/`, `/jobs/2025-01-15/`, etc.

### 6. HTTP Last-Modified Header (Low Value)
Quick HEAD request before full fetch—skip if older than cutoff. Unreliable for dynamic pages.

## Implementation Location
`modules/agent/internal/agent/tools/serper_tools.go`

Extend `parsePostingDate()` function with additional extraction methods, ordered by reliability.

## Verification
- Run job search with `time_filter: "week"`
- Check logs for "date unknown" entries before/after
- Measure reduction in unknown-date results

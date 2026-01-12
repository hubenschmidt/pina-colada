# Task: Add Staleness-Based Query Diversification

## Problem

The crawler enters a "doom loop" where consecutive runs produce 0 proposals with similar query variations. The LLM makes small tweaks but doesn't drastically change approach, leading to repeated failures.

## Current State

- `GetRunAnalytics` already tracks `ConsecutiveZeroRuns`
- `buildAnalyticsContext` includes this in the prompt but doesn't escalate demands
- The prompt says "generate COMPLETELY DIFFERENT" but LLM still produces similar variations

## Solution

Enhance `buildQuerySuggestionPrompt` to escalate "creativity pressure" based on consecutive zero-proposal runs:

### File: `internal/services/automation_service.go`

**1. Update `buildQuerySuggestionPrompt` signature** to accept consecutive zero count:
```go
func (s *AutomationService) buildQuerySuggestionPrompt(
    currentQuery, systemPrompt, documentContext, analyticsContext string,
    location *string,
    consecutiveZeroRuns int,  // NEW
) string
```

**2. Add escalating staleness instructions** based on the count:

```go
// Add after the current requirements section
if consecutiveZeroRuns >= 3 && consecutiveZeroRuns < 6 {
    sb.WriteString(`

STALENESS WARNING: ` + strconv.Itoa(consecutiveZeroRuns) + ` consecutive runs with 0 proposals.
- Try meaningfully different keywords and job titles
- Consider broadening or narrowing scope significantly`)
}

if consecutiveZeroRuns >= 6 && consecutiveZeroRuns < 10 {
    sb.WriteString(`

HIGH STALENESS: ` + strconv.Itoa(consecutiveZeroRuns) + ` consecutive zero-proposal runs!
- Make a DRASTIC change in approach
- Try completely different job titles (e.g., if searching "engineer", try "developer", "architect", "lead")
- Consider adjacent roles or broader industry terms
- Simplify - fewer terms often means more results`)
}

if consecutiveZeroRuns >= 10 {
    sb.WriteString(`

CRITICAL STALENESS: ` + strconv.Itoa(consecutiveZeroRuns) + ` consecutive failures!
- COMPLETELY PIVOT your approach
- Try a totally different angle (different seniority, different specialty, different industry)
- Use minimal, broad terms - specificity is killing results
- Consider if the search criteria itself is too restrictive`)
}
```

**3. Update caller in `checkAndSuggestQueryImprovement`** (~line 1593):
```go
// Get analytics to extract consecutive zero count
analytics, _ := s.automationRepo.GetRunAnalytics(cfg.ID, 20)
consecutiveZeros := 0
if analytics != nil {
    consecutiveZeros = analytics.ConsecutiveZeroRuns
}

prompt := s.buildQuerySuggestionPrompt(currentQuery, systemPrompt, documentContext, analyticsContext, cfg.Location, consecutiveZeros)
```

**4. Also add recent queries to avoid** in analytics context (~line 1630):
```go
// After ConsecutiveZeroRuns line, add recent queries tried
if len(analytics.RecentQueries) > 0 {
    sb.WriteString("\nRECENT QUERIES TRIED (avoid similar):\n")
    for _, q := range analytics.RecentQueries[:min(5, len(analytics.RecentQueries))] {
        sb.WriteString(fmt.Sprintf("- %s\n", truncateString(q, 80)))
    }
}
```

**5. Add RecentQueries to RunAnalytics struct** in repository:
```go
type RunAnalytics struct {
    TotalRuns           int
    ConsecutiveZeroRuns int
    AvgConversionRate   float64
    BestQueries         []QueryPerformance
    WorstQueries        []QueryPerformance
    RecentQueries       []string  // NEW - last N unique queries
}
```

## Verification

1. Run crawler multiple times with query that produces 0 proposals
2. Check logs for escalating staleness warnings
3. Verify query suggestions become more drastically different after 6+ zero runs
4. Confirm "Recent queries tried" appears in analytics context

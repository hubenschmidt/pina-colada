# Feature: LLM-Powered Query Analytics for Crawler Optimization

## Overview

Enhance the self-healing crawler by incorporating historical run analytics into query suggestion decisions. Two architectural approaches are evaluated:

1. **Inline Analysis**: Crawler LLM analyzes historical runs at query suggestion time
2. **Background Data Analyst**: Separate LLM continuously analyzes data and provides cached insights

## Critical Consideration: Finite Job Market

**The job market in a given location is a finite, dynamic pool:**

- There are only N job postings matching criteria at any given time
- Jobs get filled, new jobs are posted (turnover rate varies by market)
- After exhausting good matches, even perfect queries may return 0 new proposals
- A 0-proposal run may indicate **market exhaustion**, not a bad query

**Implications for Analytics:**

1. **Distinguish Query Quality vs Market Saturation**
   - Track unique URLs seen over time (dedup data already exists)
   - If same URLs keep appearing, market is saturated for this query
   - High prospects + low proposals = query working, market exhausted

2. **Time-Decay Analysis**
   - Recent run performance matters more than old runs
   - A query that worked last week may be exhausted now
   - Weight analytics toward recency

3. **Market Refresh Signals**
   - Track when new unique jobs appear (market refresh)
   - Suggest revisiting previously-exhausted queries after refresh
   - Monitor "jobs seen before" vs "new jobs" ratio

4. **Smart Recommendations**
   - "This query is exhausted (98% duplicates) - try broadening scope"
   - "Market refreshed - previous successful query may work again"
   - "Low prospect count suggests narrow market, not bad query"

5. **Daily Inventory Model**
   - On any given day, there are only ~N job postings for a location/criteria
   - Once all N are processed, changing queries won't find new jobs
   - The analyst should recognize: "All available inventory processed today"
   - Strategy shifts from "find better query" to "wait for new postings"
   - Track "inventory coverage" - what % of available market has been seen

## User Story

As a crawler operator, I want the system to learn from historical run performance so that query suggestions are informed by what worked and what didn't.

---

## Approach Comparison

### Approach 1: Inline Analysis (On-Demand)

**How it works:**
- When `checkAndSuggestQueryImprovement()` is triggered (proposals = 0)
- Before calling the LLM, fetch last N runs from `Automation_Run_Log`
- Build analytics context: best performing queries, patterns in failures, conversion rates
- Include this context in the LLM prompt for query generation

**Pros:**
- Simple architecture - no new services/tables
- Always uses fresh data
- No background job overhead
- Lower infrastructure cost

**Cons:**
- Adds latency to each suggestion (~500ms-2s for analysis)
- Repeated analysis work on each zero-proposal run
- Analysis quality limited by prompt size constraints
- Not reusable for other features

**Implementation Complexity:** Low-Medium

---

### Approach 2: Background Data Analyst (Cached Insights)

**How it works:**
- New `DataAnalystService` runs on a schedule (e.g., every 5 minutes or after each run)
- Analyzes `Automation_Run_Log` and stores insights in new `Automation_Analytics` table
- Crawler LLM fetches cached analytics summary when generating query suggestions
- Generic design allows reuse for any table/entity analysis

**Pros:**
- No latency added to suggestion generation
- Deeper analysis possible (longer context windows)
- Reusable for dashboards, reports, other agents
- Can track trends over time
- Enables "proactive" suggestions (don't wait for failure)

**Cons:**
- More complex architecture (new service, table, scheduler job)
- Analytics may be slightly stale
- Higher infrastructure cost (more LLM calls)
- Need to design generic abstraction

**Implementation Complexity:** Medium-High

---

## Recommended Approach: Hybrid (Start with Inline, Evolve to Background)

### Phase 1: Inline Analysis (MVP)

Enhance `generateQueryWithLLM()` to include historical run analysis in its context.

**Data to Include:**
```
MARKET STATUS:
- Total unique jobs seen today: 47
- New jobs (not seen before): 3 (6% - market nearly exhausted)
- Duplicate rate (last run): 92%
- Estimated daily refresh: ~5-10 new postings

QUERY PERFORMANCE (last 20 runs):
- Best query: "senior software engineer" AND "LLM" - 8 proposals (80% conversion)
- Worst query: "principal engineer" AND "openai" - 0 proposals (0% conversion)
- Average conversion: 23%

MARKET INTELLIGENCE:
- Keywords in successful jobs: "LLM", "startup", "Series A"
- Keywords in rejected jobs: "observability", "platform"
- Untried related searches: "machine learning engineer startup"

RECOMMENDATION:
- If duplicate rate >80%: "Market exhausted for current queries - wait for refresh or broaden scope significantly"
- If new jobs available: "X new postings detected - retry proven queries"
```

**New Repository Method:**
```go
func (r *AutomationRepository) GetRunAnalytics(configID int64, limit int) (*RunAnalytics, error)
```

**Enhanced Prompt:**
```
You are helping optimize a job search query based on historical performance data.

HISTORICAL ANALYSIS:
{analytics_summary}

CURRENT QUERY (not working):
{current_query}

Generate a query that:
- Incorporates keywords from high-performing queries
- Avoids patterns from zero-proposal runs
- Tries untested related searches if available
```

### Phase 2: Background Analyst (Future)

If inline analysis proves valuable, extract into a background service:

**New Table: `Automation_Analytics`**
```sql
CREATE TABLE "Automation_Analytics" (
    id BIGSERIAL PRIMARY KEY,
    config_id BIGINT REFERENCES "Automation_Config"(id),
    analysis_type VARCHAR(50), -- 'query_performance', 'keyword_correlation', etc.
    summary TEXT,              -- LLM-generated summary
    raw_data JSONB,            -- Structured analytics data
    generated_at TIMESTAMP,
    expires_at TIMESTAMP
);
```

**Generic Data Analyst Service:**
```go
type DataAnalystService struct {
    db          *gorm.DB
    llmClient   LLMClient
}

func (s *DataAnalystService) AnalyzeTable(tableName string, configID int64, analysisType string) (*Analysis, error)
func (s *DataAnalystService) GetCachedAnalysis(configID int64, analysisType string) (*Analysis, error)
```

---

## Scenarios

### Scenario 1: Inline Analysis Improves Query

**Given** crawler has 15 historical runs with varying success rates
**When** a run completes with 0 proposals
**Then** system analyzes history, identifies "LLM" keyword correlates with success, suggests query including "LLM"

### Scenario 2: Avoid Repeated Failures

**Given** query pattern "X AND Y" has failed 3 times
**When** LLM generates new suggestion
**Then** suggestion avoids the failing pattern

### Scenario 3: Try Untested Related Searches

**Given** Serper returned "ml engineer startup" as related search 4 times
**And** this search was never executed
**When** LLM generates suggestion
**Then** suggestion incorporates the untested search

### Scenario 4: Market Exhaustion Detection

**Given** last 5 runs had >90% duplicate rate
**When** LLM generates suggestion
**Then** response indicates "market exhausted - recommend waiting for refresh" rather than new query

---

## Verification Checklist

### Functional Requirements
- [ ] Analytics include last N runs with query, prospects, proposals
- [ ] Analytics identify best/worst performing queries
- [ ] Analytics extract keyword correlations
- [ ] Analytics detect market saturation (duplicate rate)
- [ ] LLM prompt includes analytics context
- [ ] Generated queries reflect historical insights
- [ ] System recognizes market exhaustion vs bad queries

### Non-Functional Requirements
- [ ] Analytics query executes in <100ms
- [ ] Total suggestion time <5s (including LLM call)
- [ ] Analytics context fits within LLM token limits

### Edge Cases
- [ ] First run (no history) - graceful fallback
- [ ] All runs failed (no positive examples) - broader suggestions
- [ ] Very long query history - proper truncation/sampling
- [ ] Market fully exhausted - advise waiting vs new query

---

## Implementation Notes

### Phase 1 Files to Modify

| File | Changes |
|------|---------|
| `internal/repositories/automation_repository.go` | Add `GetRunAnalytics()` method |
| `internal/services/automation_service.go` | Enhance `generateQueryWithLLM()` to include analytics |
| `internal/models/AutomationRunLog.go` | No changes (use existing fields) |

### New Types

```go
type RunAnalytics struct {
    // Market Status (key for finite pool understanding)
    TotalUniqueJobsToday   int       // Distinct URLs seen today
    NewJobsLastRun         int       // Jobs not seen in previous runs
    DuplicateRate          float64   // % of results that were duplicates
    MarketExhausted        bool      // True if duplicate rate > 80%

    // Query Performance
    TotalRuns              int
    AvgConversionRate      float64
    BestQueries            []QueryPerformance  // Top 3 by proposals
    WorstQueries           []QueryPerformance  // Bottom 3 (with >0 prospects)

    // Intelligence
    SuccessKeywords        []string            // Keywords in approved jobs
    FailureKeywords        []string            // Keywords in rejected jobs
    UntriedSuggestions     []string            // Related searches never executed
}

type QueryPerformance struct {
    Query            string
    ProspectsFound   int
    ProposalsCreated int
    ConversionRate   float64
    DuplicateCount   int     // How many were already seen
}
```

**Key Insight:** The analytics must track the **deduplication data** to understand market saturation. The existing `dedupData` struct already loads:
- `existingJobs` - Jobs already in system
- `pendingProposals` - Proposals awaiting action
- `rejectedJobs` - Agent-rejected jobs

This data can be used to calculate duplicate rates and detect market exhaustion.

### Estimate of Scope

**Phase 1 (Inline Analysis):**
- Repository method: 2 hours
- Analytics aggregation logic: 3 hours
- Prompt enhancement: 1 hour
- Testing: 2 hours
- **Total: ~1 day**

**Phase 2 (Background Analyst):**
- New service + table: 4 hours
- Scheduler integration: 2 hours
- Generic abstraction: 4 hours
- Testing: 4 hours
- **Total: ~2 days**

### Dependencies
- Existing `Automation_Run_Log` table with `executed_query`, `prospects_found`, `proposals_created`, `related_searches`
- LLM integration (already exists)

### Out of Scope (Phase 1)
- Generic data analyst for other tables
- UI dashboard for analytics
- Long-term trend analysis
- A/B testing of query strategies

---

## Decision

**Recommendation: Start with Phase 1 (Inline Analysis)**

Reasons:
1. Lower implementation risk
2. Validates the value of historical analysis
3. Can iterate quickly on prompt engineering
4. Natural evolution path to Phase 2 if successful

---

## Next Steps

1. Implement `GetRunAnalytics()` repository method
2. Build analytics aggregation logic
3. Enhance `generateQueryWithLLM()` prompt
4. Test with real crawler data
5. Evaluate results before Phase 2

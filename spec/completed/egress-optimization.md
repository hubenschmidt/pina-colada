# Plan: Reduce Supabase Egress

## Context
Supabase free plan egress exceeded (5.94/5 GB, 119%). Grace period ends May 4, 2026. Query performance data identifies one dominant source: `SELECT * FROM "Agent_Proposal"` pulling **6.85M full rows** (with JSONB `payload` + `validation_errors`) for dedup — estimated ~5-10 GB/cycle alone. Secondary: scheduler polling at ~1.45M calls/cycle (low per-call egress).

## Part A: Egress Reduction (2 files)

### 1. Add `.Select("payload")` to proposal dedup queries *(biggest win)*
**File:** `modules/agent/internal/repositories/proposal_repository.go`
- **Line 390** (`GetPendingJobProposals`): Add `.Select("payload")` before `.Find()`
- **Line 420** (`GetRejectedJobProposals`): Same fix
- Currently fetches all 16 columns when only `payload` is parsed. Drops ~40-60% of per-row bytes across 6.85M rows.

### 2. Scheduler polling queries — no change needed
**File:** `modules/agent/internal/repositories/automation_repository.go`
- `GetPausedCrawlers` and `GetDueAutomations` pass results through `modelToDTO`, which maps nearly all 40 model columns.
- Adding `.Select()` would save <5% of egress with high maintenance cost. Skipped.

## Part B: Index Optimization (1 new migration)

### 3. Migration 118: add missing composite indexes
**File:** `modules/agent/migrations/118_add_composite_indexes.up.sql`

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_agent_proposal_tenant_status_entity
  ON "Agent_Proposal" (tenant_id, status, entity_type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_automation_run_log_config_status
  ON "Automation_Run_Log" (automation_config_id, status);
```

- First index upgrades existing 2-column `(tenant_id, status)` index to cover the full WHERE clause of dedup queries.
- Second index covers status-filtered run log lookups.
- Both use `CONCURRENTLY` to avoid table locks in production.

Existing indexes already cover: `Embedding(source_type, source_id)`, `User_Role(user_id, role_id)`.

## Verification
- `cd modules/agent && go build ./...` — passes
- Run app, trigger automation, confirm dedup still works
- Run migration 118 against dev DB
- Monitor Supabase egress dashboard next billing cycle

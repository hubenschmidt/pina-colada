# Crawler: Filter Duplicates + LLM Query Suggestions

## Status: Implemented

## Goal
Find NEW jobs, not just filter duplicates. When Serper returns same results, ask LLM to suggest different search queries.

## Problem (Before)
1. Serper returns 20 jobs (same as last run)
2. ALL 20 sent to LLM → wastes tokens
3. 0 new proposals

## Solution (Implemented)
1. Filter duplicates BEFORE LLM call (save tokens)
2. If all duplicates, ask LLM for new query suggestions (N queries for N concurrent slots)
3. Track agent-rejected jobs in separate table to prevent re-reviewing
4. Location field on config - LLM only suggests job titles, location is appended
5. Save suggested queries to run log for tracking

## Changes Made

### 1. Migration: `089_automation_rejected_jobs.up.sql`
New table `Automation_Rejected_Job` to track jobs rejected by agent review:
- `automation_config_id` - links to crawler config
- `tenant_id` - tenant
- `url` - unique per config (prevents duplicates)
- `job_title`, `company`, `reason` - metadata

### 2. Migration: `090_automation_location.up.sql`
- Added `location` column to `Automation_Config`
- Added `suggested_queries` column to `Automation_Run_Log`

### 3. Models
- `AutomationRejectedJob.go` - GORM model for rejected jobs
- `AutomationConfig.go` - added `Location` field
- `AutomationRunLog.go` - added `SuggestedQueries` field

### 4. Repository: `automation_repository.go`
- `CreateRejectedJob(input)` - records a rejected job (upsert, ignores duplicates)
- `GetRejectedJobs(tenantID)` - returns rejected jobs for dedup
- `CompleteRunLog()` - now accepts `suggestedQueries` parameter
- DTOs updated for `Location` and `SuggestedQueries`

### 5. Service: `automation_service.go`
- `filterDuplicates()` - filters duplicates BEFORE LLM call
- `suggestNewQueries()` - asks Claude Haiku for N job titles (where N = concurrent_searches), appends location
- `searchWithSuggestedQueries()` - searches with suggested queries, returns them for logging
- `recordRejectedJob()` - records rejected jobs to DB
- `executeSearches()` - now returns suggested queries
- Updated `loadDedupData()` to include rejected jobs
- Updated `executeAutomation()` to save suggested queries to run log

### 6. Serializers
- `AutomationConfigResponse` - added `Location`
- `AutomationRunLogResponse` - added `SuggestedQueries`

## Flow After Changes
1. Serper returns 20 jobs
2. Filter: check against existing jobs + pending proposals + rejected jobs
3. If 0 new → ask LLM for N query suggestions (N = concurrent_searches)
4. LLM suggests job titles only, system appends location
5. Search with new queries → get different results
6. Filter again, send new jobs to LLM review
7. Create proposals for approved, record rejected to prevent re-review
8. Save suggested queries to run log

## Critical Files
- `modules/agent/migrations/089_automation_rejected_jobs.up.sql`
- `modules/agent/migrations/090_automation_location.up.sql`
- `modules/agent/internal/models/AutomationRejectedJob.go`
- `modules/agent/internal/models/AutomationConfig.go`
- `modules/agent/internal/models/AutomationRunLog.go`
- `modules/agent/internal/repositories/automation_repository.go`
- `modules/agent/internal/services/automation_service.go`
- `modules/agent/internal/serializers/automation_serializers.go`

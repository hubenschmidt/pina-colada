# Feature: User Feedback Embeddings — Approve/Reject Centroid Learning

## Overview

When a user approves or rejects a proposal on the Data Management page, embed that proposal and store it with a feedback-specific `source_type`. Approved proposals reinforce the positive centroid; rejected proposals are removed from the positive centroid and stored as negative signal that penalizes similar future results.

## User Story

As a crawler operator, I want the system to learn from my approve/reject decisions so that the vector pre-filter and query suggestions improve over time — surfacing more of what I approve and fewer of what I reject.

---

## Problem Analysis

### Current Pain Points
- The proposal centroid treats all crawler-created proposals equally (`source_type='proposal'`) — no signal about which the user actually wanted
- Proposals created via `use_agent=false` runs are never embedded at all (embedding only happens in `embedProposalAsync` during crawler execution)
- Rejected proposals remain in the positive centroid, diluting its accuracy
- Query suggestion prompts have no awareness of what the user dislikes

### Root Causes
- Embedding only fires from `AutomationService.embedProposalAsync` during crawl — no hook on user review actions
- No `source_type` distinction between crawler-generated and user-reviewed embeddings
- No rejection centroid or penalty mechanism exists

### Business Impact
- Pre-filter threshold can't adapt to user preferences — false positives persist across runs
- Query suggestions drift toward rejected-profile content because the centroid includes it
- Manual rejection volume stays flat instead of decreasing as the system learns

---

## Proposed Solution

### Core Capabilities

1. **Embed on approve** — When a user approves a proposal, embed its title/content with `source_type='user_approved'` and include it in the positive centroid.

2. **Embed on reject** — When a user rejects a proposal, delete any existing `source_type='proposal'` embedding for it, then store a new embedding with `source_type='user_rejected'`.

3. **Rejection penalty in pre-filter** — During `vectorPreFilter`, compute cosine similarity of each candidate against the rejection centroid and penalize: `finalScore = positiveScore - (rejectionSimilarity * 0.3)`.

4. **Rejection context in query suggestions** — `buildCentroidContext` appends a `USER-REJECTED PROFILE` section to the prompt so the LLM avoids generating queries that match rejected patterns.

---

## Implementation

### Step 1: Add EmbeddingService to ProposalService

**File:** `modules/agent/internal/services/proposal_service.go`
- Add `embeddingService *EmbeddingService` field to struct
- Add `SetEmbeddingService(es *EmbeddingService)` setter method

**File:** `modules/agent/cmd/agent/main.go`
- Wire `proposalService.SetEmbeddingService(embeddingService)` inside the existing `if cfg.OpenAIAPIKey != ""` block

### Step 2: Embed on user approve

**File:** `modules/agent/internal/services/proposal_service.go`
- In `ApproveProposal()`, after `MarkExecuted`, fire `go s.embedApprovedProposal(proposal)`
- `embedApprovedProposal(proposal)`:
  - Guard: return if `embeddingService == nil` or `AutomationConfigID == nil`
  - Extract title from payload JSON (`job_title`, `title`, `note`, or `content` field)
  - Call `EmbedWithSourceType(ctx, tenantID, proposalID, configID, "user_approved", title)`

### Step 3: Remove + re-embed on user reject

**File:** `modules/agent/internal/services/proposal_service.go`
- In `RejectProposal()`, after `UpdateStatus`, fire `go s.embedRejectedProposal(proposal)`
- `embedRejectedProposal(proposal)`:
  - Guard: return if `embeddingService == nil` or `AutomationConfigID == nil`
  - Delete existing embedding: `DeleteBySource("proposal", proposalID)`
  - Extract title from payload
  - Call `EmbedWithSourceType(ctx, tenantID, proposalID, configID, "user_rejected", title)`

### Step 4: New EmbeddingService method for custom source_type

**File:** `modules/agent/internal/services/embedding_service.go`
- `EmbedWithSourceType(ctx, tenantID, sourceID, configID, sourceType, text)` — same as `EmbedProposal` but accepts `sourceType` param
- `DeleteBySource(sourceType, sourceID)` — passthrough to repo
- `GetRejectionCentroid(configID)` — passthrough to repo

**File:** `modules/agent/internal/repositories/embedding_repository.go`
- `InsertEmbeddingWithType(tenantID, sourceID, configID, sourceType, text, embedding)` — same as `InsertProposalEmbedding` but accepts `sourceType`
- `GetRejectionCentroid(configID)` — `AVG(embedding) WHERE source_type='user_rejected' AND config_id=?`

### Step 5: Update centroid query to include user approvals

**File:** `modules/agent/internal/repositories/embedding_repository.go`
- `GetProposalCentroid`: change `WHERE source_type = 'proposal'` to `WHERE source_type IN ('proposal', 'user_approved')`

### Step 6: Rejection penalty in vector pre-filter

**File:** `modules/agent/internal/services/automation_service.go` — `vectorPreFilter()`
- Before the scoring loop, load rejection centroid via `GetRejectionCentroid(cfg.ID)`
- Guard: only apply penalty if `rejCount >= 3`
- For each result: `finalScore = positiveScore - (rejectionSimilarity * 0.3)`
- Filter and sort by `finalScore` instead of raw `positiveScore`

### Step 7: Rejection context in query suggestion prompt

**File:** `modules/agent/internal/services/automation_service.go` — `buildCentroidContext()`
- After the approved profile section, load rejection centroid
- Guard: only add if `rejCount >= 3`
- Find top-3 doc chunks similar to rejection centroid
- Append: `USER-REJECTED PROFILE (the user tends to reject jobs involving): [chunk excerpts]`

---

## Files Changed

| File | Changes |
|------|---------|
| `modules/agent/internal/services/proposal_service.go` | `embeddingService` field, `SetEmbeddingService`, `embedApprovedProposal`, `embedRejectedProposal`, `extractProposalTitle` |
| `modules/agent/internal/services/embedding_service.go` | `EmbedWithSourceType`, `DeleteBySource`, `GetRejectionCentroid` |
| `modules/agent/internal/repositories/embedding_repository.go` | `InsertEmbeddingWithType`, `GetRejectionCentroid`, updated `GetProposalCentroid` WHERE clause |
| `modules/agent/internal/services/automation_service.go` | Rejection penalty in `vectorPreFilter`, rejection context in `buildCentroidContext` |
| `modules/agent/cmd/agent/main.go` | Wire `proposalService.SetEmbeddingService(embeddingService)` |

## Source Types

| `source_type` | Origin | Used in centroid |
|---------------|--------|-----------------|
| `proposal` | Crawler `embedProposalAsync` | Positive |
| `user_approved` | User approves on Data page | Positive |
| `user_rejected` | User rejects on Data page | Negative (penalty) |

## Verification

1. `go build ./...` compiles
2. Approve a proposal → logs show `📐 Embedding: stored user_approved ...`
3. Reject a proposal → logs show embedding deleted + `📐 Embedding: stored user_rejected ...`
4. After 3+ user rejections, `vectorPreFilter` logs show rejection centroid loaded and penalty applied
5. Query suggestion prompt includes `USER-REJECTED PROFILE` section after 3+ rejections
6. Centroid query includes both types: `SELECT COUNT(*) FROM "Embedding" WHERE source_type IN ('proposal','user_approved') AND config_id=?`

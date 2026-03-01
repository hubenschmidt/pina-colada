# Feature: Vector-Powered Pre-Filter and Query Generation for Agent Crawler

## Overview

Add pgvector-backed semantic search to the automation crawler pipeline. Two capabilities:
1. **Pre-filter** search results against resume embeddings before LLM review (reduce token spend)
2. **Query generation** from approved proposal embeddings (replace heuristic-based suggestions with semantic understanding)

## User Story

As a crawler operator, I want the system to automatically filter out irrelevant search results before sending them to the LLM, and to generate better search queries based on what has historically been approved, so that I spend fewer API tokens and find more relevant jobs.

---

## Problem Analysis

### Current Pain Points
- Every non-duplicate search result goes to LLM review regardless of relevance — wastes tokens on obvious mismatches
- Query suggestions rely on Serper `relatedSearches` (often tangential) or an LLM call with text-based analytics context — no semantic understanding of what "a good match" looks like
- The `filterDuplicates()` check is string-based (normalized company+title) — misses semantic near-duplicates

### Root Causes
- No vector representation of resume content or search results — all matching is lexical
- No accumulated semantic profile of approved vs rejected jobs
- Query generation has no signal about the *content* of successful matches, only conversion rates

### Business Impact
- ~40-70% of LLM review calls are wasted on obviously irrelevant results (estimated from typical crawler conversion rates below 50%)
- Query suggestions often drift into unrelated territory because they lack semantic grounding
- Each wasted LLM call costs ~$0.003-0.01 (Sonnet input+output for a review batch)

---

## Proposed Solution

### Core Capabilities

1. **Semantic Pre-Filter** — Embed source documents (resumes) and search result snippets using `text-embedding-3-large`. Compute cosine similarity between each snippet and the resume. Only pass results above a configurable threshold to LLM review.

2. **Proposal-Informed Query Generation** — Embed approved proposal snippets. Compute a centroid vector representing "what good looks like" for each crawler config. Feed this semantic profile into query suggestion generation.

### Technical Approach

#### Data Model

New table via migration 107:

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE "Embedding" (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    source_type     VARCHAR(32) NOT NULL,  -- 'document', 'proposal'
    source_id       BIGINT NOT NULL,
    config_id       BIGINT,                -- automation_config_id (nullable, for proposals)
    chunk_index     INT NOT NULL DEFAULT 0,
    chunk_text      TEXT NOT NULL,
    embedding       vector(3072) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_embedding_source ON "Embedding" (source_type, source_id);
CREATE INDEX idx_embedding_config ON "Embedding" (config_id) WHERE config_id IS NOT NULL;
CREATE INDEX idx_embedding_vector ON "Embedding" USING hnsw (embedding vector_cosine_ops);
```

`source_type` values:
- `"document"` — resume/source document chunks (linked to `Asset.id`)
- `"proposal"` — approved job snippet+title (linked to proposal ID)

#### New Files

| File | Purpose |
|------|---------|
| `migrations/107_vector_embeddings.up.sql` | pgvector extension + Embedding table |
| `migrations/107_vector_embeddings.down.sql` | Drop table + extension |
| `internal/repositories/embedding_repository.go` | CRUD + vector similarity queries |
| `internal/services/embedding_service.go` | OpenAI embedding API calls, chunking, batch operations |

#### Modified Files

| File | Change |
|------|--------|
| `internal/services/automation_service.go` | Insert pre-filter step; embed approved proposals; augment query suggestion with centroid |
| `internal/services/document_service.go` | Trigger embedding on document upload/update (extend `processDocumentAsync`) |
| `go.mod` | Add `github.com/pgvector/pgvector-go` |
| `internal/models/` | Add `Embedding` model |

#### API Changes
- No new HTTP endpoints required (all internal to automation pipeline)
- Optional: `POST /automation/crawlers/{id}/reindex` to force re-embed source documents

#### UI Changes
- Optional: Add similarity threshold slider to crawler config form
- Optional: Show "vector score" alongside agent review reason in proposal detail

---

## Architecture

### Pre-Filter Flow

```
executeSearch()
  ├── searchJobs() via Serper                          ← unchanged
  ├── filterDuplicates(results, dedup)                 ← unchanged
  ├── validateURLs(filtered)                           ← unchanged
  ├── vectorPreFilter(filtered, cfg.SourceDocumentIDs) ← NEW
  │     ├── batch embed snippet+title for each result
  │     ├── load document embeddings for source_document_ids
  │     ├── compute max cosine similarity per result vs all doc chunks
  │     ├── sort by similarity descending
  │     └── return top-K above threshold
  ├── reviewResultsWithAgent(cfg, vectorFiltered, ...)  ← receives fewer results
  └── createProposalsFromReviewed(...)                   ← unchanged
```

Insertion point: `automation_service.go:699` — between `validateURLs()` and `reviewResultsWithAgent()`.

### Proposal Embedding Flow

```
createProposalFromJob()           ← existing
  └── embedProposalAsync(job, configID)  ← NEW (goroutine)
        ├── combine title + snippet
        ├── call OpenAI text-embedding-3-large
        └── INSERT INTO "Embedding" (source_type='proposal', source_id=proposalID, config_id=configID)
```

Insertion point: `automation_service.go:794` — after successful `CreateProposal()` call.

### Query Generation Flow

```
findQuerySuggestion()                               ← existing entry point
  ├── findFirstNewRelatedSearch()                    ← unchanged (try Serper first)
  ├── computeProposalCentroid(configID)              ← NEW
  │     ├── SELECT AVG(embedding) FROM "Embedding" WHERE config_id=? AND source_type='proposal'
  │     └── return centroid vector
  ├── findNearestSkillsFromCentroid(centroid, resumeEmbeddings) ← NEW
  │     └── identify which resume skill clusters are closest to what's been approved
  └── generateQueryWithLLM(cfg, currentQuery)        ← augmented with centroid context
        └── prompt now includes: "Top matching skill areas: [X, Y, Z]"
```

Insertion point: `automation_service.go:1672` — `findQuerySuggestion()`.

### Document Embedding Flow

```
processDocumentAsync()            ← existing (runs after upload)
  ├── extract text                ← existing
  ├── summarize                   ← existing
  └── embedDocumentChunks(docID)  ← NEW
        ├── chunk text (~500 tokens, 100-token overlap)
        ├── batch call OpenAI text-embedding-3-large
        └── INSERT INTO "Embedding" (source_type='document', source_id=docID)
```

Insertion point: `document_service.go:206` — `processDocumentAsync()`.

---

## Scenarios

### Scenario 1: Pre-filter reduces LLM review batch

**Given** a crawler with a resume as source document (already embedded)
**When** Serper returns 10 results and `filterDuplicates` passes 8 through
**Then** `vectorPreFilter` embeds the 8 snippets, computes similarity against resume chunks, and passes only the 4-5 above threshold to `reviewResultsWithAgent`

### Scenario 2: First run with no embeddings yet

**Given** a crawler config with `source_document_ids` referencing documents that haven't been embedded yet
**When** the crawler runs for the first time
**Then** the pre-filter step embeds the source documents synchronously (one-time cost), caches them, then proceeds with snippet filtering. Subsequent runs reuse the cached document embeddings.

### Scenario 3: Source document updated

**Given** a crawler config referencing document ID 42
**When** document 42 is re-uploaded (new version created via `UploadDocument`)
**Then** `processDocumentAsync` deletes old embeddings for source_id=42/source_type='document' and creates new ones

### Scenario 4: Proposal centroid improves query suggestions

**Given** a crawler has 20+ approved proposals with embeddings
**When** `checkAndSuggestQueryImprovement` fires
**Then** it computes the centroid of proposal embeddings, identifies the nearest resume skill clusters, and includes this semantic context in the LLM query generation prompt — producing queries aligned with proven successful matches

### Scenario 5: No proposals yet (cold start)

**Given** a new crawler with zero approved proposals
**When** `findQuerySuggestion` is called
**Then** it falls back to the existing Serper `relatedSearches` → LLM fallback path (no centroid available)

### Scenario 6: Embedding API failure

**Given** the OpenAI embedding API returns an error
**When** `vectorPreFilter` fails to embed search results
**Then** the pre-filter step is skipped and all results pass through to LLM review (graceful degradation, same as current behavior)

---

## Verification Checklist

### Functional Requirements
- [ ] `text-embedding-3-large` embeddings generated for source documents on upload
- [ ] Search result snippets batch-embedded during crawler run
- [ ] Cosine similarity computed between snippets and document chunks
- [ ] Results below threshold filtered out before LLM review
- [ ] Approved proposals embedded after creation
- [ ] Proposal centroid computed and used in query suggestion prompt
- [ ] Existing crawlers work unchanged when `vector_prefilter` is disabled (feature flag on config)
- [ ] Document re-upload replaces old embeddings

### Non-Functional Requirements
- [ ] Embedding API call adds < 2s to crawler run (batch endpoint, ~10 snippets)
- [ ] Document embedding happens async (no blocking on upload)
- [ ] HNSW index keeps vector queries under 50ms for < 100K rows
- [ ] Graceful degradation: embedding API failure skips pre-filter, does not fail the run
- [ ] Multi-tenant isolation: all queries filter by `tenant_id`

### Edge Cases
- [ ] Source document with no extractable text (binary file) — skip embedding, log warning
- [ ] Snippet is empty string — skip that result's embedding
- [ ] Config has no source documents — skip pre-filter entirely
- [ ] Very short resume (< 100 chars) — embed as single chunk, no splitting
- [ ] pgvector extension not installed — migration fails with clear error message
- [ ] Cosine similarity threshold set to 0 — effectively disables pre-filter (all pass)

---

## Implementation Notes

### Estimate of Scope
**L** — New infrastructure (pgvector, embedding service), changes to core automation pipeline, new repository, migration.

### Chunking Strategy
- Split on sentence boundaries
- Target ~500 tokens per chunk with 100-token overlap
- For short documents (< 500 tokens): single chunk, no splitting
- Use `strings.FieldsFunc` with period/newline as delimiters, accumulate until token budget

### Embedding API Usage
- Use OpenAI `text-embedding-3-large` (3072 dimensions)
- Batch endpoint: up to 2048 inputs per call
- Reuse existing `openaiClient` from `AutomationService`
- Rate limit: 3000 RPM on tier 1 — unlikely to hit with crawler volumes

### Similarity Threshold
- Default: 0.3 (cosine similarity, tunable per config)
- Add `vector_similarity_threshold` field to `AutomationConfig`
- Values: 0.0 (disabled) to 1.0 (exact match only)
- Recommended starting range: 0.25-0.40

### Backfill
- Existing source documents: embed on first crawler run if no embeddings found (lazy backfill)
- Existing proposals: optional one-time migration script (not critical for cold start — centroid improves over time)

### Configuration Fields to Add to AutomationConfig
- `vector_prefilter_enabled` (bool, default false) — feature flag
- `vector_similarity_threshold` (float, default 0.3)
- `vector_max_results` (int, default 10) — max results to pass to LLM after vector filter

### Dependencies
- `github.com/pgvector/pgvector-go` — Go types for pgvector
- PostgreSQL with `pgvector` extension installed (>= 0.5.0 for HNSW)
- OpenAI API key (already configured)

### Out of Scope
- Embedding conversation history
- Chatbot semantic search
- Vector search UI / explorer
- Embedding CRM entity data
- Fine-tuned or self-hosted embedding models
- Re-ranking (Cohere, cross-encoder) — future optimization

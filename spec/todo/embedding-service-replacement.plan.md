# Feature: Embedding Service Replacement

## Overview

Replace OpenAI `text-embedding-3-large` with an alternative embedding provider. The OpenAI account has been deleted. This spec must be resolved before the OpenAI purge spec can proceed, since the purge deletes `embedding_service.go` and its callers.

## User Story

As a developer, I want the embedding pipeline to work without an OpenAI API key so that semantic search, document similarity, and proposal matching continue to function after the OpenAI account is removed.

---

## Blocker

This spec **blocks** the "Purge OpenAI API Key & Direct OpenAI Usage" plan. The purge plan defers embedding replacement to this spec.

---

## Decision

**Option A: Voyage AI** — selected for best retrieval accuracy, generous free tier, and OpenAI-compatible API (no new Go dependency needed).

**SDK strategy**: Voyage's REST API is OpenAI-compatible. Use the existing `openai-go` client with `BaseURL` pointed at `https://api.voyageai.com/v1` and `VOYAGE_API_KEY` as the bearer token. Zero new dependencies.

---

## Options

### Option A: Voyage AI (voyage-4-large) — SELECTED

| Attribute | Detail |
|---|---|
| Model | `voyage-4-large` (or `voyage-4` for budget) |
| Retrieval | +14% over OpenAI 3-large on RTEB benchmarks |
| Price | $0.12/M tokens (voyage-4-large), $0.06/M (voyage-4) |
| Free tier | 200M tokens |
| Dimensions | 256, 512, 1024, 2048 (Matryoshka) |
| Context | 32K tokens |
| Quantization | float32, int8, uint8, binary, ubinary |
| Go SDK | No official SDK — REST API or community pkg `github.com/austinfhunter/voyageai` |
| Hosting | Cloud API only (larger models). `voyage-4-nano` is Apache 2.0 / self-hostable |
| Availability | AWS Marketplace, Azure Marketplace, MongoDB Atlas native |
| Maturity | Acquired by MongoDB (Feb 2025, ~$220M). 250+ customers. MoE architecture |
| Specialized | `voyage-code-3` (code), `voyage-context-3` (long docs), `voyage-finance-2`, `voyage-law-2` |
| Unique | Shared embedding space across voyage-4 family — embed with large, query with lite, no re-vectorization |

**Pros:**
- Highest independently-verified retrieval accuracy
- 32K context window (4x OpenAI)
- Generous free tier for development
- Batch API with 33% discount
- Specialized models for code and domain-specific content

**Cons:**
- No official Go SDK — must wrap REST API
- Cloud-only for production models
- MongoDB acquisition may affect long-term API independence

### Option B: Cohere Embed v4 — Best Go SDK + Managed Cloud

| Attribute | Detail |
|---|---|
| Model | `embed-v4.0` |
| Retrieval | 55.1 nDCG@10 (vs OpenAI 54.6) — marginal edge |
| Price | $0.12/M tokens |
| Free tier | 1,000 API calls/month |
| Dimensions | 256, 512, 1024, 1536 (Matryoshka) |
| Context | 128K tokens |
| Quantization | float32, int8, uint8, binary, ubinary, base64 — native at inference |
| Go SDK | Official: `github.com/cohere-ai/cohere-go/v2` |
| Hosting | Cloud API, AWS Bedrock, SageMaker, Azure AI Foundry, Oracle OCI |
| Maturity | GA since April 2025. Available on 4 major cloud platforms |
| Multimodal | Text + images in single embedding call |
| Languages | 100+ languages, single model |

**Pros:**
- Official Go SDK — fastest integration path
- 128K context window (16x OpenAI)
- Available on AWS Bedrock for VPC deployment
- Native quantization at inference time (no post-processing)
- Multimodal support (text + images)
- `input_type` parameter optimizes for asymmetric retrieval

**Cons:**
- Retrieval accuracy only marginally better than OpenAI
- No self-hosting option
- Smaller free tier than Voyage

---

## Head-to-Head Comparison

| Factor | Voyage AI | Cohere |
|---|---|---|
| **Retrieval accuracy** | Significantly better (+14% over OpenAI) | Marginally better (+1% over OpenAI) |
| **Price (per 1M tokens)** | $0.12 (large) / $0.06 (standard) | $0.12 |
| **Go integration effort** | Medium — wrap REST API | Low — official SDK |
| **Context window** | 32K | 128K |
| **Free tier** | 200M tokens | 1,000 calls/month |
| **Cloud deployment** | AWS/Azure Marketplace | Bedrock, SageMaker, Azure, Oracle |
| **Quantization** | float, int8, binary | float, int8, binary, base64 |
| **Matryoshka dims** | 256–2048 | 256–1536 |
| **Self-host option** | voyage-4-nano (Apache 2.0) | None |
| **Multimodal** | voyage-multimodal-3.5 (separate model) | Native in embed-v4 |

---

## Also Evaluated (Not Recommended)

| Model | Why not |
|---|---|
| **ZeroEntropy zembed-1** | Alpha SDK, ~29 HF downloads, CC-BY-NC license, no MTEB submission, no Go SDK. Too early |
| **Jina v3** | CC-BY-NC blocks commercial self-hosting. Retrieval below OpenAI (53.4 vs 54.6). Now owned by Elastic |
| **Nomic v1.5** | Lowest retrieval accuracy (~49-50). Great for self-hosting (Apache 2.0, 137M params, Ollama) but accuracy gap too large for primary use |

---

## Scenarios

### Scenario 1: Embedding Generation

**Given** a document or query text is submitted for embedding
**When** the embedding service processes the request
**Then** it returns a vector from the selected provider with configurable dimensions

### Scenario 2: Migration from OpenAI Embeddings

**Given** existing documents have OpenAI `text-embedding-3-large` embeddings (3072-dim)
**When** the new provider is deployed
**Then** all stored embeddings are re-generated with the new model and dimension size

### Scenario 3: Provider Failover

**Given** the embedding API returns an error or timeout
**When** the service retries
**Then** it uses exponential backoff and surfaces the error to the caller after max retries

---

## Verification Checklist

### Functional Requirements
- [ ] Embedding service generates vectors for documents and queries
- [ ] Asymmetric retrieval works (query vs document input types)
- [ ] Matryoshka dimension selection is configurable
- [ ] Batch embedding works for bulk document processing
- [ ] Existing callers (`automationService`, `docService`, `proposalService`) work with new service

### Non-Functional Requirements
- [ ] Latency: embedding generation < 500ms for single document
- [ ] Cost: per-token cost <= $0.13/M tokens (OpenAI baseline)
- [ ] Reliability: retry with backoff on transient failures

### Edge Cases
- [ ] Empty input text returns error, not zero vector
- [ ] Input exceeding context window is truncated (not rejected)
- [ ] Concurrent batch requests respect rate limits

---

## Implementation Notes

### Estimate of Scope
Small-medium. Primary change is swapping the HTTP client and response parsing in `embedding_service.go`.

### Files Modified
- `modules/agent/internal/services/embedding_service.go` — Voyage AI URL, model, 1024 dims, `input_type`/`output_dimension` params
- `modules/agent/internal/config/config.go` — added `VoyageAPIKey` field
- `modules/agent/cmd/agent/main.go` — embedding init uses `cfg.VoyageAPIKey`
- `modules/agent/.env.example` — added `VOYAGE_API_KEY=`
- `modules/agent/migrations/109_voyage_embeddings.up.sql` — truncate + alter column `vector(3072)` → `vector(1024)`

### Dependencies
- Raw HTTP to `https://api.voyageai.com/v1/embeddings` (OpenAI-compatible shape, no new Go dependency)
- `VOYAGE_API_KEY` env var

### Out of Scope
- Re-ranking (separate concern)
- Multimodal embeddings (future enhancement)
- Self-hosted embedding inference
- Migration script for re-vectorizing existing documents (separate task)

### Decision
Voyage AI (Option A) selected. Use `openai-go` client with Voyage base URL — no new Go dependencies.

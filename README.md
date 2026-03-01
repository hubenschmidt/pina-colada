# PinaColada CRM Architecture

## Table of Contents

- [Description](#description)
- [System Overview](#system-overview)
  - [Key Components](#key-components)
  - [Tools](#tools)
  - [Handoff Mechanism](#handoff-mechanism)
  - [Per-User Configuration](#per-user-configuration)
- [Crawler Architecture](#crawler-architecture)
  - [Pipeline Overview](#pipeline-overview)
  - [How Embeddings Improved the Pipeline](#how-embeddings-improved-the-pipeline)
    - [Before: keyword-only pipeline](#before-keyword-only-pipeline)
    - [After: embedding-augmented pipeline](#after-embedding-augmented-pipeline)
    - [Before vs after](#before-vs-after)
  - [The Feedback Loop](#the-feedback-loop)
  - [Vector Pre-Filter Detail](#vector-pre-filter-detail)
  - [Centroid-Informed Query Generation](#centroid-informed-query-generation)
  - [Query & Prompt Suggestion Flows](#query--prompt-suggestion-flows)
  - [Document Embedding Flow](#document-embedding-flow)
  - [Data Model](#data-model)
  - [LLM Usage Summary](#llm-usage-summary)
- [License](#license)

---

## Description

PinaColada is an AI-native CRM built on the OpenAI Agents Go SDK that manages relationships and workflows through natural conversation. Core entities include Contacts, Leads, Deals, Opportunities, Organizations, and Jobs. The system uses a triage agent with handoffs to specialized workers, all driven by conversational AI rather than traditional forms and dashboards.

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         FRONTEND (pinacolada.co)                    │
│                         WebSocket Client                            │
│  • Real-time streaming responses                                    │
│  • Token usage display (current/cumulative)                         │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    GO HTTP SERVER (Chi + Gorilla WS)                │
│                    cmd/agent/main.go                                │
│  • REST endpoints (/agent/chat)                                     │
│  • WebSocket streaming (/ws)                                        │
│  • Token usage tracking                                             │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    ORCHESTRATOR                                     │
│                    internal/agent/orchestrator.go                   │
│                                                                     │
│  • Manages agent lifecycle                                          │
│  • Per-user model/settings via ConfigCache                          │
│  • Conversation state (memory + DB persistence)                     │
│  • Token usage accumulation                                         │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    TRIAGE AGENT (OpenAI Agents SDK)                 │
│                    (user-configurable model, default gpt-4o)        │
│                                                                     │
│  • Intent-based routing via native handoff mechanism                │
│  • Routes to exactly ONE worker per message                         │
│  • "search jobs" → job_search | "my leads" → crm | else → general   │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  JOB SEARCH     │ │   CRM WORKER    │ │  GENERAL WORKER │
│  WORKER         │ │                 │ │                 │
│                 │ │ • crm_lookup    │ │ • crm_lookup    │
│ • job_search    │ │ • crm_list      │ │ • read_document │
│ • send_email    │ │ • crm_propose_* │ │ • Q&A           │
│ • crm_lookup    │ │ • read_document │ │                 │
│ • read_document │ │                 │ │                 │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         └───────────────────┴───────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    EVALUATOR (Optional)                             │
│                    internal/agent/evaluator.go                      │
│                    (Claude, when ANTHROPIC_API_KEY set)             │
│                                                                     │
│  • Scores response 0-100                                            │
│  • Types: Career | CRM | General                                    │
│  • If score < 60 → retry agent with feedback                        │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
                      ┌──────────┐
                      │  DONE    │
                      │ (Stream) │
                      └──────────┘
```

### Key Components

| Component        | Path                             | Purpose                                |
| ---------------- | -------------------------------- | -------------------------------------- |
| **Orchestrator** | `internal/agent/orchestrator.go` | Agent lifecycle, config, state         |
| **Workers**      | `internal/agent/workers/`        | Specialized agents with filtered tools |
| **Tools**        | `internal/agent/tools/`          | CRM, Serper, Document, Email tools     |
| **Prompts**      | `internal/agent/prompts/`        | Centralized prompt definitions         |
| **State**        | `internal/agent/state/`          | Memory + DB persistence                |
| **Evaluator**    | `internal/agent/evaluator.go`    | Optional quality control (Claude)      |

### Tools

```
┌─────────────────────────────────────────────────────────────────────┐
│                         AVAILABLE TOOLS                             │
├─────────────────────────────────────────────────────────────────────┤
│  CRM Tools                     │  External Tools                    │
│  • crm_lookup    (read)        │  • job_search    (Serper API)      │
│  • crm_list      (list)        │  • send_email    (notifications)   │
│  • crm_propose_create          │  • read_document (resumes, PDFs)   │
│  • crm_propose_update          │                                    │
│  • crm_propose_delete          │                                    │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                    TOOL ACCESS BY WORKER                            │
├──────────────────┬──────────────────┬───────────────────────────────┤
│  job_search      │  crm_worker      │  general_worker               │
├──────────────────┼──────────────────┼───────────────────────────────┤
│  ✓ job_search    │  ✗ job_search    │  ✗ job_search                 │
│  ✓ send_email    │  ✗ send_email    │  ✗ send_email                 │
│  ✓ crm_lookup    │  ✓ crm_lookup    │  ✓ crm_lookup                 │
│  ✗ crm_list      │  ✓ crm_list      │  ✗ crm_list                   │
│  ✗ crm_propose_* │  ✓ crm_propose_* │  ✗ crm_propose_*              │
│  ✓ read_document │  ✓ read_document │  ✓ read_document              │
└──────────────────┴──────────────────┴───────────────────────────────┘
```

Workers receive filtered tool sets via `tools.FilterTools()` to enforce separation of concerns.

### Handoff Mechanism

The triage agent uses the OpenAI Agents SDK's native **handoff** feature to route requests. When the triage agent determines which worker should handle a message, it triggers a handoff—transferring full conversation context and control to that worker. The worker then executes with its own model settings and filtered tool set. This is a single-hop delegation (triage → worker), not a multi-agent orchestration loop.

### Per-User Configuration

- Model selection per node (triage, each worker, evaluator)
- LLM settings: temperature, top_p, max_tokens, penalties
- Provider: OpenAI (primary) or Anthropic

---

## Crawler Architecture

The automation crawler searches for job listings, evaluates relevance via LLM, and creates proposals. It runs on a configurable interval per crawler and supports self-healing (query/prompt suggestions, auto-pause on compilation).

### Pipeline Overview

```mermaid
flowchart TD
    SCHED[/"Scheduler (cron)"/] --> EXEC["executeAutomation()"]

    EXEC --> PAR1["Parallel Load"]
    PAR1 --> DEDUP["loadDedupData()<br/><i>4 goroutines</i><br/>Jobs · Proposals · Rejected · UserRejected"]
    PAR1 --> DOCS["loadDocumentContext()<br/><i>N goroutines</i><br/>S3 download · PDF/text extract"]

    DEDUP --> SEARCH
    DOCS --> SEARCH
    SEARCH["executeSearch()"] --> SERPER["Serper API<br/>(Google Search)"]
    SERPER -->|"jobResults[] + relatedSearches[]"| PROCESS

    subgraph PROCESS["processSearchResultsWithAgent()"]
        direction TB
        F1["1. filterDuplicates — URL exact · company|title normalized"]
        F2["2. validateURLs — concurrent HEAD requests (5x)"]
        F3["3. vectorPreFilter — embed snippets → cosine vs docs"]
        F4["4. reviewResultsWithAgent — Claude/GPT approval per result"]
        F5["5. createProposalsFromReviewed → embedProposalAsync"]
        F1 --> F2 --> F3 --> F4 --> F5
    end

    F5 --> POST["Post-run hooks"]
    POST --> COMP["checkCompilation()"]
    POST --> SUGG["triggerSuggestions()"]
    POST --> ZERO["trackZeroRuns()"]

    style F3 fill:#e8f5e9,stroke:#2e7d32,color:#1a1a1a
    style F5 fill:#e8f5e9,stroke:#2e7d32,color:#1a1a1a
    style F4 fill:#fff3e0,stroke:#e65100,color:#1a1a1a
```

### How Embeddings Improved the Pipeline

#### Before: keyword-only pipeline

```mermaid
flowchart LR
    SERPER["Serper API<br/>~10-20 results"] --> DEDUP["Dedup<br/>URL + title"]
    DEDUP --> URLS["URL validation<br/>HEAD requests"]
    URLS -->|"all survivors"| LLM["LLM review<br/>(Claude/GPT)<br/><b>$$$ per result</b>"]
    LLM -->|approved| PROP["Proposals"]
    LLM -->|rejected| REJ["Rejected jobs"]

    PROP -.->|"no feedback loop"| SERPER

    style LLM fill:#ffcdd2,stroke:#b71c1c,color:#1a1a1a
```

Every result that passed dedup and URL checks went straight to the LLM for a full review. A typical run sent 8-15 results to Claude, but only 2-4 were relevant. The rest were obvious mismatches (wrong industry, wrong seniority, unrelated role) that a human would dismiss at a glance. Each wasted review burned ~2K tokens of input + output.

Query suggestions relied entirely on keyword heuristics (Serper related searches) or an LLM generating queries with only the system prompt and document text as context. The LLM had no signal about what kinds of jobs had actually been approved in the past.

#### After: embedding-augmented pipeline

```mermaid
flowchart LR
    SERPER["Serper API<br/>~10-20 results"] --> DEDUP["Dedup"]
    DEDUP --> URLS["URL validation"]
    URLS --> VPF["Vector pre-filter<br/><b>embed + cosine</b><br/>~$0.0001/result"]
    VPF -->|"top N similar"| LLM["LLM review<br/>(Claude/GPT)"]
    LLM -->|approved| PROP["Proposals"]
    LLM -->|rejected| REJ["Rejected jobs"]

    PROP -->|"embed async"| EMB[("Embedding table<br/>proposal vectors")]
    EMB -->|"centroid → query gen"| SERPER

    style VPF fill:#e8f5e9,stroke:#2e7d32,color:#1a1a1a
    style EMB fill:#e3f2fd,stroke:#1565c0,color:#1a1a1a
    style LLM fill:#fff3e0,stroke:#e65100,color:#1a1a1a
```

The vector pre-filter embeds each result's `title + snippet` and compares against source document embeddings (resumes). Only semantically similar results pass through to the LLM.

Approved proposals are embedded and stored. Over time, their centroid vector becomes a semantic fingerprint of "what this crawler approves," which is injected into query suggestion prompts.

#### Before vs after

```mermaid
flowchart TD
    subgraph BEFORE["Before: keyword-only pipeline"]
        direction LR
        B1["10 results"] -->|"all 10"| B2["LLM review<br/>10 calls"]
        B2 -->|"~3 approved"| B3["3 proposals"]
        B2 -->|"~7 rejected"| B4["wasted tokens"]
    end

    subgraph AFTER["After: embedding-augmented pipeline"]
        direction LR
        A1["10 results"] -->|"vector filter"| A2["4 results<br/>(sim ≥ 0.3)"]
        A2 --> A3["LLM review<br/>4 calls"]
        A3 -->|"~3 approved"| A4["3 proposals"]
        A3 -->|"~1 rejected"| A5["minimal waste"]
    end

    style B4 fill:#ffcdd2,stroke:#b71c1c,color:#1a1a1a
    style A2 fill:#e8f5e9,stroke:#2e7d32,color:#1a1a1a
    style A5 fill:#e8f5e9,stroke:#2e7d32,color:#1a1a1a
```

| Metric                   | Before                        | After                               |
| ------------------------ | ----------------------------- | ----------------------------------- |
| Results sent to LLM      | All post-dedup (~8-15)        | Only semantically similar (~3-6)    |
| LLM tokens per run       | ~20K-40K                      | ~8K-16K                             |
| LLM cost per run         | ~$0.01-0.02                   | ~$0.004-0.008                       |
| Embedding cost per run   | $0                            | ~$0.0002 (batch embed)              |
| Query suggestion context | System prompt + doc text only | + centroid-informed skill profile   |
| Learning from approvals  | None                          | Centroid refines with each approval |

### The Feedback Loop

The key architectural shift is that the system now learns from its own approvals:

```mermaid
flowchart TD
    SEARCH["Search + filter"] --> REVIEW["LLM review"]
    REVIEW -->|approved| PROPOSAL["Create proposal"]
    PROPOSAL --> EMBED["Embed proposal<br/><i>async goroutine</i>"]
    EMBED --> TABLE[("Embedding table<br/>source_type = 'proposal'")]
    TABLE -->|"AVG(embedding)"| CENTROID["Centroid vector<br/><i>semantic profile of<br/>what gets approved</i>"]
    CENTROID -->|"compare to resume chunks"| CONTEXT["Centroid context<br/><i>top 5 relevant skills</i>"]
    CONTEXT -->|"injected into prompt"| QUERYGEN["Query suggestion LLM"]
    QUERYGEN -->|"better query"| SEARCH

    style CENTROID fill:#f3e5f5,stroke:#6a1b9a,color:#1a1a1a
    style CONTEXT fill:#e8f5e9,stroke:#2e7d32,color:#1a1a1a
    style TABLE fill:#e3f2fd,stroke:#1565c0,color:#1a1a1a
```

1. **Runs 1-4**: No centroid yet. Pre-filter uses document embeddings only. Query suggestions are document-context only.
2. **Runs 5+**: Centroid activates. The system knows e.g. "this crawler approves backend engineering roles with distributed systems experience." The query suggestion LLM receives this as concrete context.
3. **Runs 10+**: Centroid stabilizes. Pre-filter becomes more precise as the approval profile sharpens. Query suggestions increasingly target the niche the crawler has converged on.

### Vector Pre-Filter Detail

```mermaid
flowchart LR
    subgraph STORED["Stored at upload time"]
        DC["Doc chunk embeddings<br/>(Embedding table)"]
    end

    subgraph RUNTIME["At search time"]
        SR["Search results<br/>title + snippet"] --> OAI["OpenAI<br/>text-embedding-3-large<br/>(batch embed)"]
    end

    DC --> COS["Cosine similarity<br/>max(result vs all chunks)"]
    OAI --> COS

    COS --> FILTER["Filter: sim ≥ threshold<br/>Sort desc · top N"]
    FILTER --> OUT["Filtered results<br/>→ LLM review"]

    style DC fill:#e3f2fd,stroke:#1565c0,color:#1a1a1a
    style OAI fill:#fff3e0,stroke:#e65100,color:#1a1a1a
    style COS fill:#f3e5f5,stroke:#6a1b9a,color:#1a1a1a
```

- **Model**: `text-embedding-3-large` (3072 dimensions)
- **Config**: `vector_similarity_threshold` (default 0.3), `vector_max_results` (default 10)

### Centroid-Informed Query Generation

```mermaid
flowchart LR
    subgraph PROPOSALS["Embedding table"]
        PE["source_type = 'proposal'<br/>config_id = this crawler"]
    end

    PE -->|"AVG(embedding)"| CENT["Centroid vector<br/><i>requires ≥ 5 proposals</i>"]

    subgraph DOCUMENTS["Embedding table"]
        DE["source_type = 'document'<br/>source_id IN doc_ids"]
    end

    CENT --> SIM["Cosine similarity"]
    DE --> SIM

    SIM --> TOP5["Top 5 doc chunks<br/>most similar to centroid"]
    TOP5 --> CTX["Appended to LLM prompt:<br/><b>APPROVED PROPOSAL PROFILE</b><br/><i>skills/experience that<br/>led to approvals</i>"]

    style CENT fill:#f3e5f5,stroke:#6a1b9a,color:#1a1a1a
    style CTX fill:#e8f5e9,stroke:#2e7d32,color:#1a1a1a
```

### Query & Prompt Suggestion Flows

```mermaid
flowchart TD
    CHECK["Post-run: checkAndTriggerSuggestions()"] --> COND{"Conversion rate < threshold<br/>OR prospects < minProspects"}
    COND -->|no| SKIP["Skip"]
    COND -->|yes| QS["Query suggestion"]
    COND -->|yes| PS["Prompt suggestion"]

    QS --> SERP["Try Serper relatedSearches"]
    SERP -->|found| QSAVE["Save suggested_query"]
    SERP -->|none new| QLLM["LLM fallback<br/><i>current query + doc context +<br/>analytics + centroid context +<br/>staleness warnings</i>"]
    QLLM --> QSAVE

    PS --> COOL{"Cooldown met?"}
    COOL -->|no| WAIT["Wait for more data"]
    COOL -->|yes| PLLM["LLM: suggest improved prompt<br/><i>current prompt + rejected jobs</i>"]
    PLLM --> PSAVE["Save suggested_prompt"]

    style QLLM fill:#fff3e0,stroke:#e65100,color:#1a1a1a
    style PLLM fill:#fff3e0,stroke:#e65100,color:#1a1a1a
```

### Document Embedding Flow

```mermaid
flowchart TD
    UPLOAD["UploadDocument()"] --> ASYNC["processDocumentAsync()<br/><i>goroutine</i>"]

    ASYNC --> SUMM["Summarize<br/>(Anthropic Claude)"]
    SUMM --> DB1[("DB: summary field")]

    ASYNC --> EMBED["EmbedDocumentChunks()<br/>(OpenAI)"]

    EMBED --> CHUNK["ChunkText()<br/>sentence-boundary splitting<br/>512 tokens/chunk · 64 overlap"]
    CHUNK --> API["OpenAI API<br/>text-embedding-3-large<br/>(batch embed)"]
    API --> DB2[("DB: Embedding table<br/>UpsertDocChunks<br/>delete + insert in tx")]

    style SUMM fill:#fff3e0,stroke:#e65100,color:#1a1a1a
    style API fill:#fff3e0,stroke:#e65100,color:#1a1a1a
    style DB2 fill:#e3f2fd,stroke:#1565c0,color:#1a1a1a
```

### Data Model

```mermaid
erDiagram
    Automation_Config {
        bool vector_prefilter_enabled "default false"
        float vector_similarity_threshold "default 0.3"
        int vector_max_results "default 10"
    }

    Embedding {
        bigint id PK
        bigint tenant_id
        varchar source_type "document | proposal"
        bigint source_id "doc_id or proposal_id"
        bigint config_id FK "nullable"
        int chunk_index
        text chunk_text
        vector embedding "vector(3072)"
        timestamptz created_at
    }

    Automation_Config ||--o{ Embedding : "config_id"
```

**Indexes**: `(source_type, source_id)`, `(config_id)`

### LLM Usage Summary

| Call Site                      | Model                            | Purpose                              | Max Tokens |
| ------------------------------ | -------------------------------- | ------------------------------------ | ---------- |
| `reviewResultsWithAgent`       | Claude Sonnet 4.5 (default)      | Approve/reject search results        | 2048       |
| `generateQueryWithLLM`         | Claude Sonnet 4.5 / configurable | Suggest new search query             | 1024       |
| `generatePromptSuggestion`     | Claude Sonnet 4.5 / configurable | Suggest improved system prompt       | 1024       |
| `EmbedTexts` (pre-filter)      | OpenAI text-embedding-3-large    | Embed search snippets for similarity | N/A        |
| `EmbedDocumentChunks`          | OpenAI text-embedding-3-large    | Embed uploaded documents             | N/A        |
| `EmbedProposal`                | OpenAI text-embedding-3-large    | Embed approved proposals             | N/A        |
| `ProcessDocument` (summarizer) | Claude (Anthropic)               | Summarize uploaded documents         | varies     |

## License

MIT

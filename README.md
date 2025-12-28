# PinaColada.co - AI Agent Architecture

## Description

PinaColada is an AI-native CRM built on the OpenAI Agents Go SDK that manages relationships and workflows through natural conversation. Core entities include Contacts, Leads, Deals, Opportunities, Organizations, and Jobs. The system uses a triage agent with handoffs to specialized workers—all driven by conversational AI rather than traditional forms and dashboards.

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

| Component | Path | Purpose |
|-----------|------|---------|
| **Orchestrator** | `internal/agent/orchestrator.go` | Agent lifecycle, config, state |
| **Workers** | `internal/agent/workers/` | Specialized agents with filtered tools |
| **Tools** | `internal/agent/tools/` | CRM, Serper, Document, Email tools |
| **Prompts** | `internal/agent/prompts/` | Centralized prompt definitions |
| **State** | `internal/agent/state/` | Memory + DB persistence |
| **Evaluator** | `internal/agent/evaluator.go` | Optional quality control (Claude) |

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

## License

MIT

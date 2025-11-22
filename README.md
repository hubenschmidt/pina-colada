# PinaColada.co - AI Agent Architecture

## Description

PinaColada is an AI-native CRM built on LangGraph that manages relationships and workflows through natural conversation. Core entities include Contacts, Leads, Deals, Opportunities, Organizations, and Jobs. The system uses an orchestrator pattern with specialized workers for tasks like job searching, cover letter writing, and general assistance—all driven by conversational AI rather than traditional forms and dashboards.

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
│                    FASTAPI WEBSOCKET SERVER                         │
│                    (server.py)                                      │
│  • Handles WebSocket connections                                    │
│  • Message routing & streaming                                      │
│  • Token usage tracking                                             │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    LANGGRAPH ORCHESTRATOR                           │
│                    (orchestrator.py + graph.py)                     │
│                                                                     │
│  State Management:                                                  │
│  • Messages (conversation history with tool pair validation)        │
│  • Success criteria (auto-generated from message context)           │
│  • Resume context (concise version for workers)                     │
│  • Token usage (current call + cumulative)                          │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    LLM ROUTER (Claude Haiku 4.5)                    │
│                    (routers/agent_router.py)                        │
│                                                                     │
│  • Intent-based routing (not keyword matching)                      │
│  • Analyzes current message + recent context                        │
│  • Routes to: worker | job_search | cover_letter_writer             │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  GENERAL WORKER │ │   JOB SEARCH    │ │  COVER LETTER   │
│  (GPT-5)        │ │   (GPT-5)       │ │  WRITER (GPT-5) │
│                 │ │                 │ │                 │
│ • Q&A           │ │ • Find jobs     │ │ • 200-300 words │
│ • General tasks │ │ • Filter by     │ │ • Resume-based  │
│ • Tool calling  │ │   applied jobs  │ │ • Professional  │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         └─────────┬─────────┘                   │
                   ▼                             │
          ┌─────────────────┐                    │
          │   TOOL NODE     │                    │
          │                 │                    │
          │ • Web Search    │                    │
          │ • Job Search    │                    │
          │ • File Ops      │                    │
          │ • Email         │                    │
          │ • Notifications │                    │
          └────────┬────────┘                    │
                   │                             │
         ┌─────────┴─────────────────────────────┘
         ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    SPECIALIZED EVALUATORS                           │
│                    (Claude Haiku 4.5)                               │
│                                                                     │
│  ┌─────────────────────┐    ┌─────────────────────┐                 │
│  │  CAREER EVALUATOR   │    │  GENERAL EVALUATOR  │                 │
│  │                     │    │                     │                 │
│  │ • Job search        │    │ • General queries   │                 │
│  │ • Cover letters     │    │ • Q&A responses     │                 │
│  │ • Career advice     │    │ • Tool usage        │                 │
│  └─────────────────────┘    └─────────────────────┘                 │
│                                                                     │
│  Quality Control: Check criteria → Feedback → Route END/RETRY       │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
               ┌───────────┴───────────┐
               ▼                       ▼
         ┌──────────┐           ┌─────────────┐
         │   END    │           │   RETRY     │
         │ (Success)│           │ (w/feedback)│
         └──────────┘           └─────────────┘
```

### Key Components

- **prompts/** - Centralized prompt definitions (protected IP)
- **workers/** - Specialized workers with dependency injection via base factory
- **evaluators/** - Domain-specific evaluation with structured output
- **routers/** - LLM-based intent routing

## License

MIT

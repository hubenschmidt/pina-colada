# PinaColada.co - AI Agent Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         FRONTEND (pinacolada.co)                    │
│                         WebSocket Client                            │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    FASTAPI WEBSOCKET SERVER                         │
│                    (server.py)                                      │
│  • Handles WebSocket connections                                    │
│  • Message routing & streaming                                      │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    LANGGRAPH ORCHESTRATOR                           │
│                    (orchestrator.py + graph.py)                     │
│                                                                     │
│  State Management:                                                  │
│  • Messages (conversation history)                                  │
│  • Success criteria (context-aware)                                 │
│  • Resume context (full/concise)                                    │
│  • Routing decisions                                                │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ ROUTER NODE │
                    └──────┬──────┘
                           │
            ┌──────────────┴──────────────────┐
            ▼                                 ▼
   ┌─────────────────┐         ┌───────────────────────────┐
   │  WORKER NODE    │         │ COVER LETTER WRITER NODE  │
   │  (GPT-5)        │         │         (GPT-5)           │
   │                 │         │                           │
   │ • Q&A           │         │ • Generates 200-300       │
   │ • Job search    │         │   word letters            │
   │ • Tool calling  │         │ • Uses resume data        │
   └────────┬────────┘         │ • Professional tone       │
            │                  └─────────┬────────────────_┘
            ▼                            │
   ┌─────────────────┐                   │
   │  TOOL NODE      │                   │
   │                 │                   │
   │ • Web Search    │◄──────────────────┘
   │ • File Ops      │
   │ • Notifications │
   │ • User Recording│
   └────────┬────────┘
            │
            │         ┌────────────────────┐
            └────────►│  EVALUATOR NODE    │
                      │  (Claude Haiku 4.5)│
                      │                    │
                      │ Quality Control:   │
                      │ • Checks criteria  │
                      │ • Gives feedback   │
                      │ • Routes END/RETRY │
                      └─────────┬──────────┘
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
              ┌──────────┐           ┌─────────────┐
              │   END    │           │  RETRY      │
              │ (Success)│           │ (w/feedback)│
              └──────────┘           └─────────────┘
```

## Description

This is a production-grade LangGraph-based conversational AI agent that acts as a personal assistant for a resume website (pinacolada.co). It uses an orchestrator pattern with multiple specialized worker nodes to handle various tasks like answering questions, writing cover letters, and searching for jobs.

## License

MIT

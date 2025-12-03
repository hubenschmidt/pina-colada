# Spec: AI-Native Architecture for CRM Data Models

## Context

The LinkedIn thesis: PFM/budgeting apps are vulnerable because their moats (user financial context) will evaporate as LLMs gain permissioned access to richer data via browser automation. The same logic applies to CRMs.

**Core question**: How do we leverage existing CRM data models while architecting for an AI-native future?

---

## Current State: What We Have

### Existing Schema Strengths
- **Multi-tenant isolation** (Tenant as root boundary)
- **Rich entity graph**: Individual ↔ Contact ↔ Organization ↔ Account
- **Polymorphic flexibility**: Task/Note can attach to any entity
- **Joined table inheritance**: Lead → Job | Opportunity | Partnership
- **Full audit trail**: created_at, updated_at, created_by, updated_by on all business entities
- **Status abstraction**: Centralized status definitions across entity types

### What Traditional CRM Architecture Provides
```
[Structured DB] → [Application Logic] → [Dashboard UI]
     ↓                    ↓                    ↓
  Determinism      Business Rules        Fixed Views
```

---

## The Tension: AI-Native vs Structured Persistence

### The LinkedIn Post Implies
```
[Browser Automation / Raw Data Access] → [LLM Reasoning] → [Conversational UI]
```

### The Reality
LLMs don't "hold" data persistently. They need:
1. **Context injection** (ephemeral, per-session)
2. **RAG retrieval** (vector stores, knowledge graphs)
3. **Structured output schemas** (for determinism)

**Therefore**: The schema doesn't disappear—it shifts roles.

---

## Proposed Architecture: Schema as Infrastructure, LLM as Interface

### Layer Model

```
┌─────────────────────────────────────────────────────────┐
│  Interface Layer (AI-Native)                            │
│  - Natural language queries                             │
│  - Conversational workflows                             │
│  - Context-aware suggestions                            │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Reasoning Layer (LLM + Tools)                          │
│  - Query planning / SQL generation                      │
│  - Entity resolution                                    │
│  - Relationship inference                               │
│  - Action orchestration                                 │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Context Layer (Hybrid Storage)                         │
│  - Relational DB (existing CRM schema) → Determinism    │
│  - Vector store (embeddings) → Semantic search          │
│  - Event log (audit trail) → Temporal reasoning         │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Ingestion Layer (Future-Forward)                       │
│  - API integrations (current: manual entry)             │
│  - Browser automation (future: permissioned scraping)   │
│  - Email/calendar sync                                  │
│  - External enrichment (Clearbit, LinkedIn, etc.)       │
└─────────────────────────────────────────────────────────┘
```

---

## Key Architectural Decisions

### 1. Schema Role Shift
| Current Role | Future Role |
|--------------|-------------|
| Schema = Product IP | Schema = Commodity infrastructure |
| Application logic = Differentiator | LLM reasoning = Differentiator |
| UI = Primary interface | Conversation = Primary interface |
| Fixed queries | Dynamic query generation |

### 2. What to Preserve
- **Relational integrity**: Foreign keys, constraints, audit columns
- **Tenant isolation**: Critical for multi-tenant SaaS
- **Deterministic queries**: "How many deals closed this quarter?" must be exact
- **Status workflows**: State machines for pipeline stages

### 3. What to Add
- **Semantic layer**: Embeddings for entities (Organization descriptions, Note content, etc.)
- **Natural language → SQL**: LLM translates user intent to structured queries
- **Entity resolution**: LLM disambiguates "that company we talked to last week"
- **Context stitching**: Combine structured data + unstructured notes + external signals

---

## Implementation Considerations

### Phase 1: Expose Schema to LLM (Low Lift)
- Generate schema documentation the LLM can reference
- Build tool/function calling interface for CRUD operations
- Enable natural language queries that compile to SQL

### Phase 2: Add Semantic Layer (Medium Lift)
- Embed entity descriptions, notes, task content
- Vector store for semantic search ("find contacts similar to X")
- RAG for contextual suggestions

### Phase 3: Expand Ingestion Surface (Higher Lift)
- Email integration (extract entities, create tasks)
- Calendar sync (meeting context)
- Browser extension / automation (future)

---

## Open Questions

1. **Where does the LLM context window limitation bite us?**
   - Large tenants with thousands of contacts
   - Solution: RAG + selective context loading

2. **How do we maintain audit integrity with LLM-generated actions?**
   - Log the prompt + reasoning that led to each mutation
   - Extend created_by/updated_by to include "via_agent" flag?

3. **What's the right abstraction for "tool calling"?**
   - One tool per entity type? (create_contact, update_deal)
   - Generic CRUD tool with entity type param?
   - SQL execution tool with guardrails?

4. **How do we handle determinism requirements?**
   - Financial/count queries must be exact (SQL)
   - Qualitative queries can be semantic (embeddings + LLM)

---

## Status

**Conceptual document only** - No implementation planned at this time.

This serves as a reference for future architectural decisions when making the CRM AI-native.

---

## Related Files
- `modules/agent/migrations/` - Schema definitions
- `modules/agent/src/models/` - SQLAlchemy models

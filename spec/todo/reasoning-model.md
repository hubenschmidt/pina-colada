# Feature: AI-Assisted CRM (RAG + Semi-Automatic Updates)

## Overview

Enable AI agents to **retrieve CRM context** (RAG) and **propose updates** so users can focus on closing leads rather than manual data entry.

**Core capabilities:**
1. **RAG**: Schema registry tells agents which tables/fields to query for context
2. **Research**: Agents search web for new leads, contacts, company info
3. **Proposals**: Agents suggest CRM updates (new leads, contacts, tasks) for user approval
4. **Approval Workflow**: User reviews and approves changes before they're committed

## User Story

As a user managing job/opportunity/partnership leads, I want AI to handle research and data entry so I can focus on relationship-building and closing deals.

---

## Use Cases

### 1. Find Similar Opportunities
> "Find job leads similar to ones that reached interview stage"

**Agent flow**:
1. Query CRM for successful leads (status = "Interview")
2. Analyze patterns (company size, industry, role type)
3. Web search for similar opportunities
4. **Propose** new leads for user approval

**User benefit**: Discover opportunities without manual searching

### 2. Enrich Contacts
> "Research contacts at organizations linked to my job leads"

**Agent flow**:
1. Query CRM for leads with linked organizations
2. Web search for decision-makers at those orgs
3. **Propose** new contacts/individuals for user approval

**User benefit**: Build contact lists without LinkedIn grinding

### 3. Create Follow-up Tasks
> "Create tasks to follow up on leads that haven't been updated in 2 weeks"

**Agent flow**:
1. Query CRM for stale leads
2. **Propose** tasks with suggested due dates and notes

**User benefit**: Never forget to follow up

### 4. Data Hygiene
> "Find duplicate organizations or contacts and suggest merges"

**Agent flow**:
1. Query CRM for potential duplicates (fuzzy name matching)
2. **Propose** merge actions for user approval

**User benefit**: Clean data without manual auditing

---

## Approval Workflow

All CRM modifications go through approval:

```
Agent researches → Proposes changes → User reviews → Approves/Rejects → Committed
```

**Proposal format**:
```json
{
  "proposed_action": "create",
  "entity_type": "Individual",
  "data": { "first_name": "Jane", "last_name": "Doe", "title": "CTO" },
  "confidence": 0.85,
  "source": "LinkedIn search",
  "link_to": { "organization_id": 123 }
}
```

**UI shows**: List of proposals with approve/reject buttons, batch approve option

---

## Schema Registry (RAG Component)

The `reasoning` table tells agents which CRM tables exist and how to query them.

### Table Definition

```sql
CREATE TABLE reasoning (
    id              SERIAL PRIMARY KEY,
    type            TEXT NOT NULL,              -- 'crm', 'finance', 'project', etc.
    table_name      TEXT NOT NULL,              -- 'Account', 'Contact', 'Organization'
    description     TEXT,                       -- Human/agent-readable description
    schema_hint     JSONB,                      -- Optional: key fields, relationships
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(type, table_name)
);
```

### Example Data

```sql
INSERT INTO reasoning (type, table_name, description, schema_hint) VALUES
-- CRM Core Entities
('crm', 'Account',       'Groups of organizations and individuals', '{"key_fields": ["name", "tenant_id"]}'),
('crm', 'Organization',  'Companies and business entities',         '{"key_fields": ["name", "website", "industry"]}'),
('crm', 'Individual',    'People/contacts',                         '{"key_fields": ["first_name", "last_name", "email", "title"]}'),
('crm', 'Contact',       'Contact points linking individuals and orgs', NULL),

-- CRM Pipeline
('crm', 'Deal',          'Sales opportunities at tenant level',     '{"key_fields": ["name", "value_amount", "probability"]}'),
('crm', 'Lead',          'Base lead entity (Job, Opportunity, Partnership)', '{"polymorphic": true}'),
('crm', 'Job',           'Job opportunity leads',                   NULL),
('crm', 'Opportunity',   'Sales opportunity leads',                 NULL),
('crm', 'Partnership',   'Partnership leads',                       NULL),

-- CRM Work Management
('crm', 'Project',       'Project containers',                      NULL),
('crm', 'Task',          'Work items (polymorphic attachment)',     '{"polymorphic": true}'),
('crm', 'Note',          'Notes on any entity (polymorphic)',       '{"polymorphic": true}'),

-- CRM Assets
('crm', 'Asset',         'Base asset class',                        NULL),
('crm', 'Document',      'File attachments',                        NULL),

-- CRM Provenance
('crm', 'Data_Provenance', 'Field-level source tracking',           '{"key_fields": ["entity_type", "entity_id", "field_name", "source", "confidence"]}');
```

---

## Usage Pattern

### Agent Query Flow

```python
def get_reasoning_tables(db_session, reasoning_type: str) -> list[str]:
    """Get all tables relevant to a reasoning context."""
    result = db_session.execute(
        text("SELECT table_name FROM reasoning WHERE type = :type"),
        {"type": reasoning_type}
    )
    return [row[0] for row in result.fetchall()]

# Agent usage
tables = get_reasoning_tables(session, "crm")
# Returns: ['Account', 'Organization', 'Individual', 'Contact', 'Deal', ...]
```

### With Schema Hints

```python
def get_reasoning_schema(db_session, reasoning_type: str) -> list[dict]:
    """Get tables with schema hints for richer context."""
    result = db_session.execute(
        text("""
            SELECT table_name, description, schema_hint
            FROM reasoning
            WHERE type = :type
        """),
        {"type": reasoning_type}
    )
    return [
        {"table": r[0], "description": r[1], "hints": r[2]}
        for r in result.fetchall()
    ]
```

---

## Integration with AI-Native Architecture

This table implements the "Reasoning Layer" from `ai_native_crm_architecture.md`:

```
┌─────────────────────────────────────────────────────────┐
│  Reasoning Layer (LLM + Tools)                          │
│  - Query planning / SQL generation                      │  ← Agent queries `reasoning`
│  - Entity resolution                                    │    to discover relevant tables
│  - Relationship inference                               │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Context Layer (Hybrid Storage)                         │
│  - Relational DB (existing CRM schema) → Determinism    │  ← Agent queries discovered
│  - Vector store (embeddings) → Semantic search          │    tables for actual data
└─────────────────────────────────────────────────────────┘
```

---

## Future Extensions

### Additional Reasoning Types

```sql
-- Future contexts beyond CRM
('finance', 'Invoice', 'Customer invoices', NULL),
('finance', 'Payment', 'Payment records', NULL),
('analytics', 'Event', 'User events for analysis', NULL),
```

### Schema Hint Enhancements

```jsonc
{
  "key_fields": ["name", "email"],
  "relationships": [
    {"table": "Organization", "via": "account_id"}
  ],
  "polymorphic": true,
  "searchable_fields": ["description", "notes"],
  "embedding_fields": ["description"]  // Fields to embed for semantic search
}
```

---

## Tables Included (type="crm")

From `user-audit-columns.md`:
1. Account
2. Asset
3. Contact
4. Deal
5. Document
6. Individual
7. Job
8. Lead
9. Note
10. Opportunity
11. Organization
12. Partnership
13. Project
14. Task

Plus:
15. Data_Provenance (existing)

---

## LangGraph Integration

### Architecture

Integrates with existing orchestrator-worker-evaluator pattern in `modules/agent/`:

```
Router → CRM Worker → Tools → Evaluator → Retry or END
              ↓
    Schema context pre-loaded
    from reasoning table
```

### New CRM Worker

**Location**: `modules/agent/src/agent/workers/crm/crm_worker.py`

- Created via `create_base_worker_node` factory
- Receives pre-loaded schema context string at orchestrator startup
- Uses schema hints for both SQL generation AND prompt enrichment

### CRM Tools

**Location**: `modules/agent/src/agent/tools/crm_tools.py`

**CRM Read Tools (service layer - preferred)**:
- `lookup_account(search_query)` - via account_service
- `lookup_organization(search_query)` - via organization_service
- `lookup_individual(search_query)` - via individual_service
- `lookup_contact(search_query)` - via contact_service
- `lookup_deal(search_query)` - via deal_service
- `lookup_lead(search_query, lead_type?, status?)` - via lead_service

**SQL Fallback**:
- `execute_crm_query(query, reasoning)` - SELECT-only, validates tables against reasoning registry

**Web Search** (reuse existing from `worker_tools.py`):
- `SerperSearch` - Google search for job postings, company info
- `WikipediaQuery` - Company/industry research

**Proposal Output**:
- CRM worker returns structured proposal reports (JSON or markdown)
- No direct database writes in Phase 1
- Format: `{ proposed_action, entity_type, data, confidence, source }`

Worker prompt instructs to:
1. Prefer service tools for CRM reads
2. Use SQL for complex joins/aggregations
3. Use web search for external research
4. Return proposals as reports, not direct updates

### Router Integration

**File**: `modules/agent/src/agent/routers/agent_router.py`

Add `crm_worker` route for intents:
- Accounts, organizations, contacts, individuals
- Deals, leads, opportunities, partnerships
- "Show me", "Find", "Look up" + CRM entities

### State Changes

Add to `State` TypedDict in `orchestrator.py`:
```python
schema_context: Optional[str]  # Pre-loaded from reasoning service
tenant_id: Optional[int]       # For tenant-scoped CRM queries
```

### Vector DB

**Phase 2** - not required for initial implementation. The `schema_hint.embedding_fields` property is forward-compatible for when vector search is added.

---

## Implementation Checklist

### Phase 1: Database Layer
- [ ] Create `Reasoning` SQLAlchemy model (`models/Reasoning.py`)
- [ ] Register in `models/__init__.py`
- [ ] Create `reasoning_service.py` with `get_reasoning_schema()` and `format_schema_context()`
- [ ] Create seed script (`scripts/seed_reasoning.py`)

### Phase 2: CRM Worker
- [ ] Create `workers/crm/crm_worker.py` using base worker factory
- [ ] Add `build_crm_worker_prompt()` to `prompts/worker_prompts.py`

### Phase 3: CRM Tools
- [ ] Create `tools/crm_tools.py` with service-layer lookup tools
- [ ] Add SQL fallback tool with table validation

### Phase 4: Router
- [ ] Add `crm_worker` to `RouterDecision` literal
- [ ] Update router system prompt with CRM routing rules

### Phase 5: Orchestrator
- [ ] Add `schema_context` and `tenant_id` to State
- [ ] Load schema context at startup
- [ ] Wire CRM worker node and conditional edges

---

## Related Files

- `spec/todo/ai_native_crm_architecture.md` - Parent architecture spec
- `spec/todo/user-audit-columns.md` - CRM tables with audit columns
- `modules/agent/src/models/DataProvenance.py` - Existing provenance model

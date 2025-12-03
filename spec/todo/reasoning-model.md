# Feature: Reasoning Model (Schema Registry for AI Agents)

## Overview

A metadata registry table that maps database tables to reasoning contexts. When an AI agent needs to research/reason on a specific domain (e.g., "crm"), it queries this registry to discover which tables are relevant.

## User Story

As an AI agent, I want to know which tables contain data relevant to a specific reasoning context so that I can efficiently query the right sources without hardcoding table names.

---

## Schema

### Reasoning Table

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

## Implementation Checklist

- [ ] Create migration for `reasoning` table
- [ ] Seed with CRM table mappings
- [ ] Create `Reasoning` SQLAlchemy model
- [ ] Add utility function for agent table discovery
- [ ] Document in agent system prompt / tool definitions

---

## Related Files

- `spec/todo/ai_native_crm_architecture.md` - Parent architecture spec
- `spec/todo/user-audit-columns.md` - CRM tables with audit columns
- `modules/agent/src/models/DataProvenance.py` - Existing provenance model

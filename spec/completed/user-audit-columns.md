# Feature: User Audit Columns on Core Business Models

## Overview

Add `created_by` and `updated_by` columns (FK to User.id) to 14 core business models to track who created and last modified each record. This enables audit trails and is critical for AI user workflows (e.g., "query all records created by AI user for approval").

## User Story

As a system administrator, I want to know which user created or last modified any business record so that I can audit changes, implement approval workflows for AI-generated records, and maintain accountability.

---

## Affected Models

| Model        | File                                       | Current State                                   |
| ------------ | ------------------------------------------ | ----------------------------------------------- |
| Account      | `modules/agent/src/models/Account.py`      | No user tracking                                |
| Asset        | `modules/agent/src/models/Asset.py`        | Has `user_id` (ownership), no audit             |
| Contact      | `modules/agent/src/models/Contact.py`      | No user tracking                                |
| Deal         | `modules/agent/src/models/Deal.py`         | No user tracking                                |
| Document     | `modules/agent/src/models/Document.py`     | Inherits Asset                                  |
| Individual   | `modules/agent/src/models/Individual.py`   | No user tracking                                |
| Job          | `modules/agent/src/models/Job.py`          | No user tracking                                |
| Lead         | `modules/agent/src/models/Lead.py`         | No user tracking                                |
| Note         | `modules/agent/src/models/Note.py`         | **Has `created_by`** (extend with `updated_by`) |
| Opportunity  | `modules/agent/src/models/Opportunity.py`  | No user tracking                                |
| Organization | `modules/agent/src/models/Organization.py` | No user tracking                                |
| Partnership  | `modules/agent/src/models/Partnership.py`  | No user tracking                                |
| Project      | `modules/agent/src/models/Project.py`      | No user tracking                                |
| Task         | `modules/agent/src/models/Task.py`         | No user tracking                                |

---

## Schema Changes

### Column Definitions

```sql
created_by  INTEGER REFERENCES "user"(id),  -- nullable for legacy data
updated_by  INTEGER REFERENCES "user"(id)   -- nullable for legacy data
```

### SQLAlchemy Model Pattern

```python
created_by = Column(Integer, ForeignKey("user.id"), nullable=True)
updated_by = Column(Integer, ForeignKey("user.id"), nullable=True)
```

### Note Model (Special Case)

Note already has `created_by`. Only add `updated_by`.

---

## Migration

Single Alembic migration adding columns to all 14 tables:

```python
# migrations/versions/xxx_add_user_audit_columns.py

def upgrade():
    tables = [
        'account', 'asset', 'contact', 'deal', 'document',
        'individual', 'job', 'lead', 'opportunity',
        'organization', 'partnership', 'project', 'task'
    ]
    for table in tables:
        op.add_column(table, sa.Column('created_by', sa.Integer(), sa.ForeignKey('user.id'), nullable=True))
        op.add_column(table, sa.Column('updated_by', sa.Integer(), sa.ForeignKey('user.id'), nullable=True))

    # Note already has created_by, only add updated_by
    op.add_column('note', sa.Column('updated_by', sa.Integer(), sa.ForeignKey('user.id'), nullable=True))
```

---

## API/Service Layer Changes

### Automatic Population

These columns should be populated automatically:

- `created_by`: Set on INSERT when user context is available
- `updated_by`: Set on every UPDATE when user context is available

### Implementation Options

1. **Service layer**: Explicitly pass user_id to create/update methods
2. **SQLAlchemy event listeners**: Hook `before_insert`/`before_update` events
3. **Request context**: Use Flask/FastAPI request context to inject current user

---

## Use Cases

### Query Records by Creator

```sql
-- All leads created by AI user (user_id = 999)
SELECT * FROM lead WHERE created_by = 999;
```

### Approval Workflows

```sql
-- Pending AI-created records requiring human review
SELECT * FROM account
WHERE created_by IN (SELECT id FROM "user" WHERE is_ai_user = true)
  AND approved_at IS NULL;
```

### Audit Trail

```sql
-- Who last touched this record?
SELECT updated_by, updated_at FROM deal WHERE id = 123;
```

---

## Implementation Checklist

- [ ] Create Alembic migration for all 14 tables
- [ ] Add `created_by`, `updated_by` columns to model definitions
- [ ] Add ForeignKey relationships to User model
- [ ] Update service layer to populate columns on create/update
- [ ] Update API endpoints to pass user context
- [ ] Add Note.updated_by only (already has created_by)

---

## Files to Modify

### Models (add columns)

- `modules/agent/src/models/Account.py`
- `modules/agent/src/models/Asset.py`
- `modules/agent/src/models/Contact.py`
- `modules/agent/src/models/Deal.py`
- `modules/agent/src/models/Document.py`
- `modules/agent/src/models/Individual.py`
- `modules/agent/src/models/Job.py`
- `modules/agent/src/models/Lead.py`
- `modules/agent/src/models/Note.py` (add `updated_by` only)
- `modules/agent/src/models/Opportunity.py`
- `modules/agent/src/models/Organization.py`
- `modules/agent/src/models/Partnership.py`
- `modules/agent/src/models/Project.py`
- `modules/agent/src/models/Task.py`

### Migration

- `modules/agent/migrations/` (new Alembic migration)

### Services (populate columns)

- Service files for each model (update create/update methods)

---

## Out of Scope

- Backfilling existing records (columns nullable for legacy data)
- UI display of created_by/updated_by
- History/changelog of all modifications (see revision-history.md)

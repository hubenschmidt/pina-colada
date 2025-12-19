# Field-Level Diff Review for Agent Proposals

## Summary

Enhance the agent proposal review UI to show field-level diffs for update operations, allowing reviewers to see exactly what changes the agent proposes.

## Background

The current Agent_Proposal system stores the complete payload as JSONB. For update operations, this includes all fields that would be modified. To provide a better review experience, we should:

1. Store the "before" state alongside the proposed "after" state
2. Display a visual diff in the review UI
3. Optionally support per-field approve/reject

## Database Changes

### Option A: Add before_state column

```sql
ALTER TABLE "Agent_Proposal"
ADD COLUMN before_state JSONB;
```

Store the entity's current state when the proposal is created.

### Option B: Separate diff table

```sql
CREATE TABLE "Agent_Proposal_Diff" (
    id              BIGSERIAL PRIMARY KEY,
    proposal_id     BIGINT NOT NULL REFERENCES "Agent_Proposal"(id) ON DELETE CASCADE,
    field_path      VARCHAR(255) NOT NULL,
    before_value    JSONB,
    after_value     JSONB,
    is_approved     BOOLEAN,  -- For per-field approve/reject
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## API Changes

### GET /proposals/{id}/diff

Returns field-level diff for a proposal:

```json
{
  "proposal_id": 123,
  "entity_type": "organization",
  "entity_id": 456,
  "operation": "update",
  "diffs": [
    {
      "field": "name",
      "before": "Acme Corp",
      "after": "Acme Corporation"
    },
    {
      "field": "website",
      "before": null,
      "after": "https://acme.example.com"
    }
  ]
}
```

### POST /proposals/{id}/partial-approve (Optional)

For per-field approval:

```json
{
  "approved_fields": ["name"],
  "rejected_fields": ["website"]
}
```

## UI Component

A React component that displays:

1. Entity type and ID
2. Operation type (create/update/delete)
3. For updates: side-by-side or inline diff view
4. For creates: all proposed fields
5. For deletes: entity summary with warning
6. Approve/Reject buttons (bulk or per-field)

## Implementation Order

1. Add `before_state` column to Agent_Proposal
2. Modify ProposeOperation to capture current state for updates
3. Add GET /proposals/{id}/diff endpoint
4. Create frontend diff viewer component
5. (Optional) Implement per-field approve/reject

## Out of Scope

- History of previous proposals for the same entity
- Auto-merge conflict resolution
- Proposal dependencies/ordering

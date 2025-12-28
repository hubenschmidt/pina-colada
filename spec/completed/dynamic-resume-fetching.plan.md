# Feature: Dynamic Resume Fetching

## Problem

1. **Static resume context** is hardcoded into agent startup, causing:
   - Evaluator confusion (CRM queries evaluated against resume criteria)
   - Stale data (resume never updates)
   - Bloated context for non-career tasks

2. **Tenant scoping is optional** - agent asks "which tenant?" but should ALWAYS be restricted to current tenant

---

## Solution

### 1. Remove Static Resume Context

**Current** (`orchestrator.py`):
```python
resume_context = _build_resume_context(summary, resume_text, sample_answers, cover_letters)
```

**New**: Remove `resume_text`, `summary`, `sample_answers`, `cover_letters` from orchestrator startup. Career workers fetch dynamically when needed.

### 2. Dynamic Document Fetching

When career worker needs a resume:

**Option A: By Individual**
> "Conduct job search using the most recent resume for William Hubenschmidt"

1. Look up Individual by name
2. Query Documents linked to that Individual (via EntityAsset or direct link)
3. Sort by `created_at DESC`, take first
4. Load document content into context

**Option B: By Last Job Application**
> "Use the resume from the last job I applied to"

1. Query Jobs with `status='applied'` sorted by `created_at DESC`
2. Get linked Document (if any)
3. Load into context

### 3. Tenant Restriction (Hardcoded)

**Current**: Agent asks "which tenant or region?"

**New**:
- Pass `tenant_id` into orchestrator from auth context
- All CRM tools receive `tenant_id` as required parameter
- Remove tenant_id from tool inputs (not user-facing)
- SQL queries automatically scoped: `WHERE tenant_id = :tenant_id`

---

## Implementation

### Phase 1: Tenant Hardcoding

1. **orchestrator.py**: Accept `tenant_id` parameter
2. **CRM tools**: Remove `tenant_id` from user inputs, inject from state
3. **Service functions**: Always require `tenant_id`, no optional

### Phase 2: Remove Static Resume

1. Remove `resume_text`, `summary`, etc. from `create_orchestrator()`
2. Update worker prompts to not expect resume context
3. Career workers: prompt instructs to fetch resume via tools

### Phase 3: Document Fetching Tool

New tool: `get_resume_for_individual(individual_name: str)`
- Looks up Individual
- Finds linked Documents with tag "resume" or filename pattern
- Returns document content

### Phase 4: Evaluator Fix

- CRM worker should use `general_evaluator` with CRM-appropriate criteria
- OR create `crm_evaluator` with CRM-specific success criteria
- Remove resume references from general evaluator when route is CRM

---

## Files to Modify

| File | Change |
|------|--------|
| `orchestrator.py` | Accept tenant_id, remove static resume |
| `agent/tools/crm_tools.py` | Inject tenant_id, remove from inputs |
| `agent/workers/crm/crm_worker.py` | No resume context in prompt |
| `agent/workers/career/*.py` | Fetch resume dynamically |
| `agent/evaluators/general_evaluator.py` | Context-aware criteria |
| `services/*_service.py` | Require tenant_id |

---

## Questions

1. How is `tenant_id` passed to the agent? (from JWT? session? websocket message?)
2. Should we create a `crm_evaluator` or make `general_evaluator` smarter?
3. Resume lookup: by Individual name, or by current user?

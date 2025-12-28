# Usage Analytics Feature

## Overview
Implement a Usage Analytics feature that allows users to track their token spend and developers to analyze usage patterns for optimization decisions.

---

## Requirements

### 1. Usage Tab in Account Menu
- Add "Usage" menu item under Account dropdown in Header
- Route: `/usage` or `/account/usage`
- Available to all authenticated users

### 2. User Usage View
Display for all users:
- **Individual token spend** (user's own usage)
- **Tenant token spend** (all-time for their organization)
- **Time-filtered graph**: daily, weekly, monthly, quarterly, annual views

### 3. Developer Analytics Section
Additional section visible only to users with `developer` role:
- Same graph as user view, but with additional filters:
  - Filter by **node** (router, worker, evaluator, intent_classifier, etc.)
  - Filter by **tool** (lookup_individual, lookup_organization, execute_crm_query, etc.)
- Purpose: Identify optimization opportunities and cost drivers

### 4. Developer Role Setup
- Add new role `developer` to Role table
- Assign `developer` role to William Hubenschmidt (whubenschmidt@gmail.com, user_id=1)
- Developer analytics only visible to users with this role

### 5. Cost Estimation
Display estimated costs alongside token counts:
- Store `model_name` with each usage record (already in schema)
- Maintain pricing constants for each model (per 1M tokens):
  | Model | Input | Output |
  |-------|-------|--------|
  | gpt-4o | $2.50 | $10.00 |
  | gpt-4o-mini | $0.15 | $0.60 |
  | gpt-4-turbo | $10.00 | $30.00 |
  | claude-3-5-sonnet | $3.00 | $15.00 |
  | claude-3-5-haiku | $0.80 | $4.00 |
  | claude-3-opus | $15.00 | $75.00 |
- Calculate: `cost = (input_tokens * input_rate + output_tokens * output_rate) / 1_000_000`
- Show cost in summary cards and charts
- Color-code or badge costs as "low" / "moderate" / "high" based on thresholds

**Note:** OpenAI and Anthropic don't provide pricing APIs. Prices are from public pricing pages and must be updated manually when pricing changes.

---

## Current State Analysis

### Role Table (has potential duplicates)
```
id | tenant_id |  name  | description
---|-----------|--------|------------
 6 |         1 | member | Standard team member access
 5 |         1 | owner  | Full access to all resources
 2 |      NULL | admin  | Administrator with full access
 3 |      NULL | member | Standard member with access
 1 |      NULL | owner  | Tenant owner with full access
 4 |      NULL | viewer | Read-only access to view data
```
- Roles 1-4 are global (tenant_id=NULL)
- Roles 5-6 are tenant-scoped (tenant_id=1)
- Potential duplicate: `member` and `owner` exist both globally and for tenant 1

### Token Usage Storage
- `Conversation_Message.token_usage` (JSONB) stores per-message token data
- Currently tracks: `{input, output, total}` tokens
- **Missing**: No breakdown by node/tool for analytics

### William's Current Role
- User ID: 1
- Email: whubenschmidt@gmail.com
- Current role: `owner` (via User_Role table)

---

## Implementation Plan

### Phase 1: Database Schema

#### 1a. Add Developer Role
```sql
-- Migration: add_developer_role.sql
INSERT INTO "Role" (tenant_id, name, description)
VALUES (NULL, 'developer', 'Developer access with analytics and debugging tools');

-- Assign to William (user_id=1)
INSERT INTO "User_Role" (user_id, role_id)
SELECT 1, id FROM "Role" WHERE name = 'developer' AND tenant_id IS NULL;
```

#### 1b. Create Usage Analytics Table
```sql
-- Migration: create_usage_analytics.sql
CREATE TABLE "Usage_Analytics" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES "Tenant"(id),
    user_id BIGINT REFERENCES "User"(id),
    conversation_id BIGINT REFERENCES "Conversation"(id),
    message_id BIGINT REFERENCES "Conversation_Message"(id),

    -- Token counts
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,

    -- Breakdown for developer analytics
    node_name TEXT,  -- 'router', 'worker', 'evaluator', 'intent_classifier', 'fast_lookup'
    tool_name TEXT,  -- 'lookup_individual', 'lookup_organization', etc.
    model_name TEXT, -- 'gpt-4o-mini', 'claude-haiku-4-5', etc.

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_usage_tenant_created ON "Usage_Analytics"(tenant_id, created_at);
CREATE INDEX idx_usage_user_created ON "Usage_Analytics"(user_id, created_at);
CREATE INDEX idx_usage_node ON "Usage_Analytics"(node_name);
CREATE INDEX idx_usage_tool ON "Usage_Analytics"(tool_name);
```

### Phase 2: Backend

#### 2a. Update Token Tracking (orchestrator.py)
- Modify token accumulation to include node_name and tool_name
- Store detailed usage in Usage_Analytics table after each message

#### 2b. Create Usage Service
**File:** `services/usage_service.py`
```python
async def get_user_usage(user_id: int, period: str) -> dict
async def get_tenant_usage(tenant_id: int, period: str) -> dict
async def get_usage_by_node(tenant_id: int, period: str) -> dict  # developer only
async def get_usage_by_tool(tenant_id: int, period: str) -> dict  # developer only
```

#### 2e. Create Pricing Service
**File:** `services/pricing_service.py`
```python
# Static pricing (per 1M tokens) - update when providers change prices
MODEL_PRICING = {
    "gpt-4o": {"input": 2.50, "output": 10.00},
    "gpt-4o-mini": {"input": 0.15, "output": 0.60},
    "gpt-4-turbo": {"input": 10.00, "output": 30.00},
    "claude-3-5-sonnet-20241022": {"input": 3.00, "output": 15.00},
    "claude-3-5-haiku-20241022": {"input": 0.80, "output": 4.00},
    "claude-3-opus-20240229": {"input": 15.00, "output": 75.00},
}

def calculate_cost(model_name: str, input_tokens: int, output_tokens: int) -> float:
    """Calculate cost in USD for given token usage."""
    pricing = MODEL_PRICING.get(model_name, {"input": 0, "output": 0})
    return (input_tokens * pricing["input"] + output_tokens * pricing["output"]) / 1_000_000

def get_cost_tier(cost: float) -> str:
    """Return 'low', 'moderate', or 'high' based on cost thresholds."""
    if cost < 0.01:
        return "low"
    elif cost < 0.10:
        return "moderate"
    return "high"
```

#### 2c. Create Usage Router
**File:** `routers/usage_router.py`
- `GET /api/usage/user` - User's own usage
- `GET /api/usage/tenant` - Tenant's total usage
- `GET /api/usage/analytics` - Developer analytics (role-gated)

#### 2d. Role Checking Utility
**File:** `lib/role_check.py`
```python
async def user_has_role(user_id: int, role_name: str) -> bool
async def require_role(role_name: str)  # FastAPI dependency
```

### Phase 3: Frontend

#### 3a. Update Account Menu
**File:** `components/Header/Header.jsx`
- Add "Usage" menu item linking to `/usage`

#### 3b. Create Usage Page
**File:** `app/usage/page.jsx`
- Layout with tabs or sections
- User section (always visible)
- Developer section (conditionally rendered based on role)

#### 3c. Usage Components
**Files:**
- `components/Usage/UsageChart.jsx` - Recharts/Chart.js graph (tokens + cost overlay)
- `components/Usage/UsageSummary.jsx` - Summary cards with token counts AND estimated cost
- `components/Usage/CostBadge.jsx` - Color-coded badge (green=low, yellow=moderate, red=high)
- `components/Usage/DeveloperAnalytics.jsx` - Node/tool breakdown with cost per node/tool
- `components/Usage/PeriodSelector.jsx` - Daily/weekly/monthly/quarterly/annual filter

#### 3d. Usage API Hook
**File:** `hooks/useUsage.js`
```javascript
export const useUsage = (period) => { ... }
export const useDeveloperAnalytics = (period, filters) => { ... }
```

#### 3e. Role Context Update
**File:** `context/userContext.jsx`
- Include user roles in context
- Add `hasDeveloperRole` check

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `migrations/XXX_add_developer_role.sql` | CREATE |
| `migrations/XXX_create_usage_analytics.sql` | CREATE |
| `models/UsageAnalytics.py` | CREATE |
| `services/usage_service.py` | CREATE |
| `services/pricing_service.py` | CREATE |
| `routers/usage_router.py` | CREATE |
| `lib/role_check.py` | CREATE |
| `agent/orchestrator.py` | MODIFY - add detailed token tracking |
| `components/Header/Header.jsx` | MODIFY - add Usage menu item |
| `app/usage/page.jsx` | CREATE |
| `app/usage/layout.jsx` | CREATE |
| `components/Usage/UsageChart.jsx` | CREATE |
| `components/Usage/UsageSummary.jsx` | CREATE |
| `components/Usage/CostBadge.jsx` | CREATE |
| `components/Usage/DeveloperAnalytics.jsx` | CREATE |
| `components/Usage/PeriodSelector.jsx` | CREATE |
| `hooks/useUsage.js` | CREATE |
| `api/index.js` | MODIFY - add usage API calls |
| `context/userContext.jsx` | MODIFY - include roles |

---

## Implementation Order

1. **Database**: Add developer role, create Usage_Analytics table
2. **Backend Models**: Create UsageAnalytics model
3. **Backend Services**: Usage service with aggregation queries
4. **Backend Routes**: Usage API endpoints with role gating
5. **Backend Integration**: Update orchestrator to log detailed usage
6. **Frontend Context**: Add roles to user context
7. **Frontend Page**: Create /usage page with components
8. **Frontend Components**: Chart, summary, developer analytics
9. **Testing**: Verify role gating, data accuracy

---

## Open Questions

1. Should we backfill historical usage from Conversation_Message.token_usage?
2. ~~Cost calculation: Do we want to show estimated costs (tokens × rate)?~~ **YES - included in plan**
3. Should developer analytics be a separate route (/developer/analytics) or a tab on /usage?
4. Export functionality: CSV/JSON export of usage data?
5. How often should we verify/update model pricing? (suggest: monthly check against provider pricing pages)

---

## Enhancement: Add Request Count & Avg Tokens

### Goal
Add `request_count` and `conversation_count` columns to By Node and By Model analytics tables, enabling avg tokens per request calculation.

### Backend Changes: `usage_analytics_repository.py`

**`get_usage_by_node()`:**
```python
select(
    UsageAnalytics.node_name,
    UsageAnalytics.model_name,
    func.count().label("request_count"),
    func.count(func.distinct(UsageAnalytics.conversation_id)).label("conversation_count"),
    func.sum(UsageAnalytics.input_tokens).label("input_tokens"),
    func.sum(UsageAnalytics.output_tokens).label("output_tokens"),
    func.sum(UsageAnalytics.total_tokens).label("total_tokens"),
)
```

**`get_usage_by_model()`:**
```python
select(
    UsageAnalytics.model_name,
    func.count().label("request_count"),
    func.count(func.distinct(UsageAnalytics.conversation_id)).label("conversation_count"),
    func.sum(UsageAnalytics.input_tokens).label("input_tokens"),
    func.sum(UsageAnalytics.output_tokens).label("output_tokens"),
    func.sum(UsageAnalytics.total_tokens).label("total_tokens"),
)
```

### Frontend Changes: `app/usage/page.jsx`

Add columns to both analytics tables:
- **Requests** - raw LLM call count
- **Conversations** - unique conversation count
- **Avg Tokens** - computed: `total_tokens / request_count`

### Files to Modify

| File | Change |
|------|--------|
| `modules/agent/src/repositories/usage_analytics_repository.py` | Add count() to queries |
| `modules/client/app/usage/page.jsx` | Add Requests, Conversations, Avg Tokens columns |

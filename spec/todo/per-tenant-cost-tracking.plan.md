# Per-Tenant Cost Tracking

## Problem Statement

We need to track and bill costs per tenant. Currently, provider Cost APIs return org-wide data, making it difficult to attribute costs to individual tenants.

## Goals

1. **Accurate billing** - charge tenants based on actual usage
2. **Visibility** - understand which tenants consume the most resources
3. **Reconciliation** - match internal tracking with provider billing

---

## Option 1: Workspace per Tenant

Create a dedicated Anthropic workspace (and OpenAI project) for each tenant.

### How It Works

1. On tenant creation, call Anthropic Admin API to create workspace
2. Store `workspace_id` in tenant record
3. Use workspace-scoped API key for that tenant's requests
4. Query Cost API with `workspace_id` filter

### Anthropic API

```bash
# Create workspace
POST /v1/organizations/workspaces
{ "name": "tenant-{tenant_id}" }

# Query costs by workspace
GET /v1/organizations/cost_report?workspace_ids[]={workspace_id}
```

### OpenAI Equivalent

OpenAI uses "projects" similarly - create project per tenant, query costs with `project_ids` filter.

### Pros

- Native USD costs directly from provider
- Exact accuracy - no calculation needed
- Clean separation of concerns

### Cons

- High operational overhead
- Requires Admin API integration for workspace lifecycle
- API key management per workspace
- Rate limits may apply per workspace

---

## Option 2: API Key per Tenant

Create a dedicated API key for each tenant, use Usage API to track tokens.

### How It Works

1. On tenant creation, create API key via Admin API
2. Store `api_key_id` in tenant record
3. Route that tenant's requests through their key
4. Query Usage API with `api_key_ids[]` filter
5. Calculate costs from token counts using pricing table

### Anthropic API

```bash
# Create API key
POST /v1/organizations/api_keys
{ "name": "tenant-{tenant_id}" }

# Query usage by API key (tokens, not costs)
GET /v1/organizations/usage_report/messages?api_key_ids[]={api_key_id}
```

**Important:** Cost API does NOT support `api_key_ids` filtering - only Usage API does.

### Cost Calculation

```python
def calculate_cost(usage: dict, pricing: dict) -> float:
    input_cost = usage["input_tokens"] * pricing["input_per_million"] / 1_000_000
    output_cost = usage["output_tokens"] * pricing["output_per_million"] / 1_000_000
    return input_cost + output_cost
```

### Pros

- Easier than workspace management
- Clear attribution via API key
- Usage API supports key filtering

### Cons

- Cost API doesn't support key filtering (only Usage API)
- Must maintain pricing table
- Calculated costs may drift from actual billing
- Key rotation complexity

---

## Option 3: Internal Usage Tracking (Current Approach)

Use our existing `usage_analytics` table which already tracks `tenant_id`, `model_name`, and token counts.

### How It Works

1. Continue logging usage per request with `tenant_id`
2. Add pricing service to calculate costs from tokens
3. Use provider Cost API for org-wide spend validation only
4. Show per-tenant breakdown from internal data

### Current Schema

```sql
-- Already exists
CREATE TABLE Usage_Analytics (
    tenant_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    input_tokens INTEGER,
    output_tokens INTEGER,
    model_name TEXT,
    node_name TEXT,
    created_at TIMESTAMP
);
```

### New Pricing Service

```python
# services/pricing_service.py
PRICING = {
    "claude-haiku-4-5-20251001": {"input": 0.80, "output": 4.00},  # per million
    "claude-sonnet-4-5-20250929": {"input": 3.00, "output": 15.00},
    "gpt-4.1-mini": {"input": 0.40, "output": 1.60},
}

async def get_tenant_estimated_cost(tenant_id: int, period: str) -> float:
    usage = await usage_repo.get_usage_by_tenant_and_model(tenant_id, period)
    total = 0.0
    for row in usage:
        pricing = PRICING.get(row["model_name"], {"input": 0, "output": 0})
        total += (row["input_tokens"] * pricing["input"] / 1_000_000)
        total += (row["output_tokens"] * pricing["output"] / 1_000_000)
    return total
```

### Pros

- Already have the data
- No provider API dependency for per-tenant
- Full control over granularity
- Works across all providers uniformly

### Cons

- Requires maintaining pricing table
- Calculated costs may drift from actual billing
- Doesn't capture non-token costs (web search, code execution)

---

## Comparison

| Aspect | Workspace | API Key | Internal |
|--------|-----------|---------|----------|
| Setup complexity | High | Medium | Low |
| Cost accuracy | Exact | Calculated | Calculated |
| Maintenance | Workspace lifecycle | Key lifecycle | Pricing updates |
| OpenAI support | Projects | Projects | Internal |
| Anthropic support | Full (Cost API) | Partial (Usage API) | Internal |
| Non-token costs | Included | Not included | Not included |
| Implementation effort | 2-3 days | 1-2 days | 0.5 days |

---

## Recommendation

**Start with Option 3 (Internal)** - we already have the data, just need a pricing service.

Evolve to **Option 2 (API Key per Tenant)** if:
- We need provider-verified usage data
- We want key isolation for security/rate limiting
- Tenant count is manageable (<100)

Consider **Option 1 (Workspace)** only if:
- We need exact cost attribution including non-token costs
- We have significant multi-tenant isolation requirements
- We're willing to invest in workspace lifecycle management

---

## Next Steps

1. [ ] Create `services/pricing_service.py` with model pricing table
2. [ ] Add `get_tenant_costs()` to usage service
3. [ ] Add tenant cost breakdown to Usage page (admin only)
4. [ ] Set up pricing update process (manual or fetch from provider)

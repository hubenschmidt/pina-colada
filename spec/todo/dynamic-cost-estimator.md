# Dynamic Cost Estimator

## Status: Research Complete

Research into programmatic cost tracking using OpenAI and Anthropic Usage/Cost APIs.

## Background

We removed static cost estimation from the Usage Analytics page because:
1. Hardcoded pricing was inaccurate (estimates showed ~$0 for 2M tokens, actual spend was $1.70)
2. No programmatic way to get current pricing
3. Pricing changes frequently

## Provider APIs

### OpenAI Usage/Costs API

**Endpoint:** `https://api.openai.com/v1/organization/costs`

**Parameters:**
- `start_time` (required): Unix timestamp, inclusive
- `end_time`: Unix timestamp, exclusive
- `bucket_width`: Currently only `1d` supported
- `group_by`: `project_id`, `line_item`, or combination
- `limit`: 1-180, default 7
- `project_ids`: Filter to specific projects

**Authentication:** Requires `OPENAI_ADMIN_KEY` (organization admin API key)

**Response:** Paginated, time-bucketed cost data with breakdowns by project/line item

**Sources:**
- [OpenAI Usage API Reference](https://platform.openai.com/docs/api-reference/usage/costs)
- [OpenAI Cookbook Usage API Guide](https://cookbook.openai.com/examples/completions_usage_api)

---

### Anthropic Usage/Cost API

**Endpoints:**
- `/v1/organizations/usage_report/messages` - Token consumption by model, workspace, tier
- `/v1/organizations/cost_report` - Cost data grouped by workspace

**Authentication:** Requires Admin API key (`sk-ant-admin...`) - only org admins can provision

**Features:**
- Uncached vs cached tokens
- Prompt cache hit rates
- Daily cost in USD
- Group by model, API key, workspace, service tier
- ~5 minute data freshness

**Limitations:**
- Priority Tier costs use different billing, not in cost endpoint

**Sources:**
- [Anthropic Usage and Cost API](https://platform.claude.com/docs/en/build-with-claude/usage-cost-api)
- [Anthropic vs OpenAI Billing API Comparison](https://www.finout.io/blog/anthropic-vs-openai-billig-api)

---

## Key Considerations

### 1. Admin API Keys Required
Both providers require **organization-level admin API keys**, not regular API keys:
- OpenAI: `OPENAI_ADMIN_KEY`
- Anthropic: `sk-ant-admin...`

### 2. Organization-Level Data
Both APIs return **organization-wide** data, not per-request costs:
- Data is bucketed by day (minimum)
- Cannot correlate to individual chat messages
- Best for aggregate dashboards, not real-time per-message costs

### 3. Data Granularity
- OpenAI: Daily buckets only
- Anthropic: ~5 minute freshness, but still org-level

### 4. Multi-Provider
We use both OpenAI (GPT-5.1, GPT-5-mini) and Anthropic (Claude Haiku):
- Need to fetch from both APIs
- Aggregate costs across providers

---

## Implementation

### Environment Variables
```env
OPENAI_ADMIN_KEY=sk-admin-...
ANTHROPIC_ADMIN_KEY=sk-ant-admin-...
```

### Backend
- `services/provider_costs_service.py` - API calls to both providers
- `GET /costs/summary?period=monthly` - combined provider costs endpoint

### Frontend
- "Provider Spend" card on Usage page showing actual costs
- Per-provider breakdown (OpenAI vs Anthropic)
- "Last updated" timestamp

# OpenAI Rate Limit Mitigation - Technical Specification

## Problem Statement

The scraper worker consistently hits OpenAI API rate limits during browser automation tasks, causing:
- Task failures mid-execution
- Poor user experience (errors displayed to end users)
- Wasted API credits on incomplete executions
- Inability to scale scraper usage

### Root Causes

1. **High API call volume**: Scraper tasks generate 9+ tool calls, each requiring 2 graph iterations (scraper→tools→scraper), resulting in 18+ LLM invocations per task
2. **Expensive model usage**: All workers use `gpt-5-chat-latest` regardless of task complexity
3. **No rate limit handling**: Direct failure on 429 errors with no retry logic
4. **High iteration ceiling**: 50 recursion limit allows unbounded API consumption

### Evidence

From `scraper-integration.md`:
```
18 iterations = 9 tool calls
Each tool call = 2 iterations (scraper→tools→scraper)
9 tool calls for 401k rollover demo = 18+ OpenAI API calls
```

Current model configuration:
```python
# scraper.py:53
scraper_llm = ChatOpenAI(model="gpt-5-chat-latest", temperature=0, max_tokens=2048)

# worker.py:85
worker_llm = ChatOpenAI(model="gpt-5-chat-latest", temperature=0.7)

# job_hunter.py:73
job_hunter_llm = ChatOpenAI(model="gpt-5-chat-latest", temperature=0.7)

# cover_letter_writer.py:78
llm = ChatOpenAI(model="gpt-5-chat-latest", ...)
```

---

## Current State Analysis

### API Call Patterns

**Per scraper task:**
- Minimum: ~10-18 LLM calls (simple automation)
- Maximum: ~100 LLM calls (complex multi-page flows, 50 iteration limit)
- Average: ~20-30 calls per task

**Cost breakdown (estimated):**
- GPT-5: ~$0.10-0.30 per scraper task
- GPT-4o-mini alternative: ~$0.01-0.03 per scraper task (10x reduction)

### Rate Limit Tiers

OpenAI rate limits (approximate, tier-dependent):
- **Tier 1-2**: 3,500 RPM (requests per minute)
- **Tier 3**: 10,000 RPM
- **Tier 4**: 30,000 RPM

**Current exposure:**
- Single scraper task: 18-30 requests
- 5 concurrent users running scraper: 90-150 RPM
- 10 concurrent users: 180-300 RPM
- **Rate limit hit at ~12-20 concurrent users on Tier 1-2**

### Architectural Constraints

```
orchestrator.py:258
├─ recursion_limit: 50  # Allows up to 50 graph iterations
├─ No per-agent budgets
└─ No global rate limiting

scraper.py:99
├─ Bypasses evaluator to reduce iterations ✓
└─ Still requires ~18 iterations minimum

graph.py:254-289
├─ No retry logic on RateLimitError
└─ Immediate failure propagated to user
```

---

## Proposed Solutions

### Solution 1: Model Downgrade Strategy (RECOMMENDED - Phase 1)

**Rationale:**
Tool-calling tasks (scraper, worker) don't require GPT-5's advanced reasoning. Structured output generation works excellently with GPT-4o-mini.

**Changes:**

```python
# modules/agent/src/agent/workers/scraper.py:52-57
scraper_llm = ChatOpenAI(
    model="gpt-4o-mini",  # Changed from gpt-5-chat-latest
    temperature=0,
    max_tokens=2048,
    callbacks=callbacks,
)
```

```python
# modules/agent/src/agent/workers/worker.py:84-90
worker_llm = ChatOpenAI(
    model="gpt-4o-mini",  # Changed from gpt-5-chat-latest
    temperature=0.7,
    max_tokens=4096,
    callbacks=callbacks,
)
```

```python
# modules/agent/src/agent/workers/job_hunter.py:72-76
job_hunter_llm = ChatOpenAI(
    model="gpt-4o-mini",  # Changed from gpt-5-chat-latest
    temperature=0.7,
    max_tokens=4096,
    callbacks=callbacks,
)
```

**Keep GPT-5 for creative tasks:**
```python
# modules/agent/src/agent/workers/cover_letter_writer.py:77-78
llm = ChatOpenAI(
    model="gpt-5-chat-latest",  # KEEP - creative writing benefits from GPT-5
    # ...
)
```

**Impact:**
- **Cost**: 10-20x reduction per task
- **Rate limits**: Same RPM limits but tasks complete faster, reducing sustained load
- **Quality**: Expected 0-5% quality degradation for tool calls (minimal)
- **Latency**: 20-30% improvement (GPT-4o-mini faster than GPT-5)

---

### Solution 2: Exponential Backoff Retry Logic (Phase 1)

**Implementation:**

```python
# modules/agent/src/agent/graph.py (after line 17)
from openai import RateLimitError
from tenacity import (
    retry,
    retry_if_exception_type,
    wait_exponential,
    stop_after_attempt,
    before_sleep_log
)

# Create retry decorator
retry_on_rate_limit = retry(
    retry=retry_if_exception_type(RateLimitError),
    wait=wait_exponential(multiplier=1, min=4, max=60),
    stop=stop_after_attempt(5),
    before_sleep=before_sleep_log(logger, logging.WARNING),
    reraise=True
)
```

**Apply to graph invocation:**

```python
# modules/agent/src/agent/graph.py:254-257 (invoke_graph function)
@observe()
async def invoke_graph(
    websocket: WebSocket,
    data: Dict[str, Any] | str | list[dict[str, str]],
    user_uuid: str,
):
    # ... existing code ...

    # Wrap run_streaming with retry logic
    @retry_on_rate_limit
    async def _run_with_retry():
        return await orchestrator["run_streaming"](
            message=message,
            thread_id=user_uuid,
            success_criteria=success_criteria
        )

    try:
        await _run_with_retry()
    except RateLimitError as e:
        logger.error(f"Rate limit exceeded after retries: {e}")
        # Send graceful error to user
        await websocket.send_text(
            json.dumps({
                "on_chat_model_stream": "\n\nI'm experiencing high API usage right now. Please try again in a moment."
            })
        )
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))
    except Exception as e:
        # ... existing error handling ...
```

**Backoff schedule:**
- Attempt 1: Immediate
- Attempt 2: Wait 4s
- Attempt 3: Wait 8s
- Attempt 4: Wait 16s
- Attempt 5: Wait 32s
- Total max wait: ~60s before failure

**Impact:**
- Handles transient rate limit spikes
- Graceful degradation vs hard failures
- ~80-90% success rate improvement under moderate load

---

### Solution 3: Per-Agent Recursion Budgets (Phase 2)

**Rationale:**
Different agents have different needs. Scraper needs high iteration count; worker/job_hunter don't.

**Implementation:**

```python
# modules/agent/src/agent/orchestrator.py:249-260
async def run_streaming(
    message: str, thread_id: str, success_criteria: str = None
) -> Dict[str, Any]:
    logger.info(f"▶️  Starting graph execution for thread: {thread_id}")

    # Determine recursion limit based on detected agent type
    detected_agent = state.get("route_to_agent", "worker")
    recursion_limits = {
        "scraper": 50,          # Browser automation needs high limit
        "cover_letter_writer": 30,  # Creative writing may iterate
        "job_hunter": 25,       # Job search with tool calls
        "worker": 20,           # General Q&A, fewer iterations
    }

    config = {
        "configurable": {"thread_id": thread_id},
        "recursion_limit": recursion_limits.get(detected_agent, 20),
    }
    # ... rest of function
```

**Alternative approach - dynamic budget tracking:**

```python
# Track remaining budget in state
class State(TypedDict):
    messages: Annotated[List[Any], add_messages]
    success_criteria: str
    feedback_on_work: Optional[str]
    success_criteria_met: bool
    user_input_needed: bool
    resume_name: str
    resume_context: str
    route_to_agent: Optional[str]
    iteration_budget: int  # NEW: Track remaining iterations

# Decrement in each worker node
def worker_node(state: Dict[str, Any]) -> Dict[str, Any]:
    budget = state.get("iteration_budget", 20)
    if budget <= 0:
        logger.warning("Iteration budget exhausted, forcing completion")
        return {
            "messages": [AIMessage(content="Task requires more iterations than budgeted. Partial result returned.")],
            "success_criteria_met": False
        }

    # ... normal processing ...
    return {
        "messages": [response],
        "iteration_budget": budget - 1
    }
```

**Impact:**
- Prevents runaway iterations on simple tasks
- 30-40% reduction in average API calls
- Better cost predictability

---

### Solution 4: Response Caching Layer (Phase 3)

**Rationale:**
Browser snapshots, static page scrapes, and tool results often identical within short time windows.

**Implementation:**

```python
# modules/agent/src/agent/util/cache.py (NEW FILE)
import hashlib
import json
from datetime import datetime, timedelta
from typing import Any, Optional
import logging

logger = logging.getLogger(__name__)

# In-memory cache (consider Redis for production)
_cache: dict[str, tuple[Any, datetime]] = {}
_cache_ttl = timedelta(seconds=30)  # 30s TTL for browser state

def cache_key(tool_name: str, args: dict) -> str:
    """Generate cache key from tool name + args"""
    serialized = json.dumps({"tool": tool_name, "args": args}, sort_keys=True)
    return hashlib.sha256(serialized.encode()).hexdigest()

def get_cached(key: str) -> Optional[Any]:
    """Get cached result if not expired"""
    if key not in _cache:
        return None

    result, timestamp = _cache[key]
    if datetime.now() - timestamp > _cache_ttl:
        del _cache[key]
        return None

    logger.info(f"Cache HIT: {key[:12]}...")
    return result

def set_cached(key: str, value: Any):
    """Cache result with timestamp"""
    _cache[key] = (value, datetime.now())
    logger.info(f"Cache SET: {key[:12]}...")

def clear_cache():
    """Clear all cached results"""
    _cache.clear()
    logger.info("Cache cleared")
```

**Wrap tool execution:**

```python
# modules/agent/src/agent/tools/mcp_playwright.py (modify tool calls)
from agent.util.cache import cache_key, get_cached, set_cached

async def browser_snapshot_cached(url: str) -> str:
    """Cached browser_snapshot - snapshots valid for 30s"""
    key = cache_key("browser_snapshot", {"url": url})

    cached = get_cached(key)
    if cached:
        return cached

    result = await browser_snapshot(url)  # Original function
    set_cached(key, result)
    return result
```

**Impact:**
- 10-20% reduction in API calls (cache hit rate dependent)
- Faster response times for repeated operations
- Marginal complexity increase

---

## Implementation Plan

### Phase 1: Immediate Mitigations (Week 1)
**Priority: P0 - Critical**

1. **Model downgrade** (scraper.py, worker.py, job_hunter.py)
   - [ ] Change `gpt-5-chat-latest` → `gpt-4o-mini`
   - [ ] Keep GPT-5 for cover_letter_writer
   - [ ] Deploy to staging
   - [ ] Run 20 test scraper tasks, measure quality delta
   - [ ] Deploy to production

2. **Retry logic** (graph.py)
   - [ ] Add `tenacity` dependency to requirements.txt
   - [ ] Implement exponential backoff wrapper
   - [ ] Add Langfuse metrics for retry events
   - [ ] Deploy to staging
   - [ ] Simulate rate limit scenario, verify retries
   - [ ] Deploy to production

**Success metrics:**
- Zero rate limit errors in staging (20 test runs)
- Cost per scraper task < $0.05 (down from ~$0.15-0.30)
- Task completion rate > 95%

### Phase 2: Optimization (Week 2-3)
**Priority: P1 - High**

3. **Per-agent budgets** (orchestrator.py)
   - [ ] Add budget mapping by agent type
   - [ ] Add budget tracking to State
   - [ ] Implement budget enforcement in workers
   - [ ] Add Langfuse dashboards for iteration metrics
   - [ ] A/B test: 20% of traffic with budgets enabled
   - [ ] Full rollout if metrics positive

4. **Monitoring & alerting**
   - [ ] Add Langfuse traces for rate limit events
   - [ ] Create dashboard: API calls per agent, cost per task
   - [ ] Set up alerts for sustained rate limit errors (>5/min)

**Success metrics:**
- Average iterations per task < 15 (down from ~20-30)
- Cost reduction: 60-70% vs baseline
- Quality metrics: maintain >90% task success rate

### Phase 3: Advanced Optimization (Week 4+)
**Priority: P2 - Medium**

5. **Response caching**
   - [ ] Implement in-memory cache for browser tools
   - [ ] Add cache metrics to Langfuse
   - [ ] Measure cache hit rate, tune TTL
   - [ ] Consider Redis backend for multi-instance deployments

6. **Model routing**
   - [ ] A/B test: GPT-4o vs GPT-4o-mini for scraper
   - [ ] Implement adaptive model selection (start mini, upgrade if stuck)

**Success metrics:**
- Cache hit rate > 15%
- Total cost reduction: 70-80% vs baseline
- Maintain quality above 90%

---

## Success Criteria

### Primary Metrics
- **Rate limit errors**: 0 per day (down from current failures)
- **Cost per scraper task**: <$0.05 (GPT-4o-mini) vs ~$0.20 (GPT-5)
- **Task completion rate**: >95%
- **Quality score**: >90% (evaluated by success_criteria_met)

### Secondary Metrics
- **Average iterations per task**: <15 (down from 20-30)
- **Latency**: <30s average task completion (improved from GPT-5)
- **Cache hit rate**: >15% (Phase 3)
- **Retry success rate**: >80% (retries that succeed after backoff)

### Monitoring Dashboard (Langfuse)
1. **Cost tracking**: Daily spend by agent type
2. **Iteration distribution**: Histogram of iterations per task
3. **Error rates**: 429 errors, retry events, final failures
4. **Quality metrics**: success_criteria_met percentage by agent
5. **Cache performance**: Hit rate, TTL effectiveness

---

## Rollback Plan

If quality degrades below 85% success rate:

1. **Immediate**: Revert model changes (GPT-4o-mini → GPT-5)
2. **Keep**: Retry logic (no downside)
3. **Investigate**: Analyze failed tasks in Langfuse
4. **Alternative**: Try GPT-4-turbo as middle ground

**Rollback trigger criteria:**
- Success rate drops >10% (below 85%)
- User-reported errors increase >50%
- Cost savings <30% (not worth quality trade-off)

---

## Files to Modify

### Phase 1 Changes

```
modules/agent/src/agent/workers/scraper.py:52-57
modules/agent/src/agent/workers/worker.py:84-90
modules/agent/src/agent/workers/job_hunter.py:72-76
modules/agent/src/agent/graph.py:17 (add imports)
modules/agent/src/agent/graph.py:254-289 (add retry wrapper)
modules/agent/requirements.txt (add tenacity)
```

### Phase 2 Changes

```
modules/agent/src/agent/orchestrator.py:22-31 (update State)
modules/agent/src/agent/orchestrator.py:249-260 (add budget logic)
modules/agent/src/agent/workers/*.py (add budget tracking)
```

### Phase 3 Changes

```
modules/agent/src/agent/util/cache.py (NEW)
modules/agent/src/agent/tools/mcp_playwright.py (wrap with cache)
modules/agent/src/agent/tools/scraper_tools.py (wrap with cache)
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Quality degradation (GPT-4o-mini) | Medium | High | A/B test, keep GPT-5 as fallback |
| Retry delays frustrate users | Low | Medium | Show "thinking" indicator, max 60s total retry |
| Cache staleness issues | Low | Low | Conservative 30s TTL, cache only idempotent ops |
| Budget limits prevent task completion | Medium | Medium | Start with generous budgets, tune based on data |
| Unforeseen GPT-4o-mini limitations | Low | High | Comprehensive staging tests, gradual rollout |

---

## Alternative Approaches Considered

### 1. Switch to Anthropic Claude for tool calling
**Pros**: Claude Haiku excellent for tool calls, cheaper than GPT-4o-mini
**Cons**: Already using Claude for evaluator, mixing providers adds complexity
**Decision**: Not recommended for Phase 1, consider for Phase 3

### 2. Local LLM for tool calling (LLaMA 3, Mixtral)
**Pros**: Zero API costs, no rate limits
**Cons**: Hosting costs, latency, quality concerns
**Decision**: Not viable for production, consider for internal testing

### 3. Request OpenAI rate limit increase
**Pros**: Simple, no code changes
**Cons**: Expensive (higher tier pricing), doesn't solve cost issue
**Decision**: Fallback option if other mitigations insufficient

### 4. Queue-based rate limiting (Celery + Redis)
**Pros**: Global rate limiting across instances
**Cons**: Adds infrastructure, increases latency
**Decision**: Consider for Phase 3 if scaling beyond single instance

---

## Questions & Clarifications

1. **What is our current OpenAI tier?** (Determines exact rate limits)
2. **Do we have budget for Redis?** (Needed for Phase 3 caching at scale)
3. **What's acceptable latency increase?** (Retry logic adds up to 60s worst case)
4. **Are there quality benchmarks?** (Need baseline to measure degradation)

---

## References

- Existing spec: `spec/scraper-integration.md` (iteration counts, tool usage patterns)
- OpenAI pricing: https://openai.com/api/pricing/
- OpenAI rate limits: https://platform.openai.com/docs/guides/rate-limits
- LangChain caching: https://python.langchain.com/docs/modules/model_io/llms/llm_caching
- Tenacity docs: https://tenacity.readthedocs.io/

---

**Document Status**: Draft
**Last Updated**: 2025-11-17
**Owner**: Engineering
**Reviewers**: TBD

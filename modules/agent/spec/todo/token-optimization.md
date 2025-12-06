# Token Optimization & Chat Control

## Status: Backend Complete ✅

**Created**: 2025-12-05
**Last Updated**: 2025-12-05

---

## Problem Summary

Job search consumed **350k tokens** and **50+ iterations** due to:
1. `gpt-5-mini` (LEAN_MODEL) fails at tool-calling, causing endless retry loops
2. No way for users to stop runaway iterations
3. High token usage from worker→evaluator loop retries

---

## Implementation Status

### Phase 1: Remove LEAN_MODEL from Tool-Heavy Workers ✅ DONE

**Files Modified**:
- `agent/workers/career/job_search.py` - Removed `model=LEAN_MODEL`, now uses default gpt-5.1

**Rationale**: gpt-5-mini is not suitable for agentic tool-calling tasks.

---

### Phase 2: Add Stop/Cancel Buttons to Chat ✅ DONE

**Files Modified**:
- `server.py:200-212` - Added cancel message handling
- `orchestrator.py:22-40` - Added `_running_tasks` dict and `cancel_streaming()` function
- `orchestrator.py:322-325` - Register tasks for cancellation
- `orchestrator.py:447-455` - Handle `CancelledError` and cleanup

**WebSocket Protocol**:

Client → Server:
```json
{"type": "cancel", "uuid": "<thread_id>"}
```

Server → Client:
```json
{"type": "cancelled"}
```

---

### Phase 3: Reduce Iteration Limits ✅ DONE

**Files Modified**:
- `orchestrator.py:334` - Reduced `recursion_limit` from 50 to 20

---

### Phase 4: Token Budget Safeguards ✅ DONE

**Files Modified**:
- `orchestrator.py:26` - Added `TOKEN_BUDGET = 50000` constant
- `orchestrator.py:440-445` - Stop execution if token budget exceeded

---

### Phase 5: Tool-Level TTL Caching ✅ DONE

**Rationale**: LLM node caching is ineffective (message history changes each iteration), but tool functions often receive identical inputs within a conversation.

#### Implementation Plan

**Step 1**: Add cachetools dependency
```bash
uv add cachetools
```

**Step 2**: Add async TTL cache decorator
```python
from cachetools import TTLCache
from functools import wraps

def async_ttl_cache(cache, key_fn):
    """Decorator for async functions with TTL cache."""
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            key = key_fn(*args, **kwargs)
            if key in cache:
                logger.debug(f"Cache HIT: {func.__name__}({key})")
                return cache[key]
            result = await func(*args, **kwargs)
            cache[key] = result
            return result
        return wrapper
    return decorator
```

**Step 3**: Cache Document Tools (`agent/tools/document_tools.py`)

| Function | Cache Key | TTL | Benefit |
|----------|-----------|-----|---------|
| `get_document_content` | `(tenant_id, document_id)` | 5 min | Avoid S3 downloads, PDF parsing |
| `search_documents` | `(tenant_id, query, tags, entity_type, entity_id)` | 2 min | Avoid DB queries |

**Step 4**: Cache CRM Tools (`agent/tools/crm_tools.py`)

| Function | Cache Key | TTL | Benefit |
|----------|-----------|-----|---------|
| `lookup_individual` | `(tenant_id, query)` | 2 min | Avoid DB queries |
| `lookup_organization` | `(tenant_id, query)` | 2 min | Avoid DB queries |
| `execute_crm_query` | `(tenant_id, query_hash)` | 1 min | Avoid repeated DB queries |

**Expected Savings**:
- 50-80% reduction in tool execution time for repeated calls
- Reduced database/storage load
- Lower latency for cached responses

**Files Modified**:
- `agent/tools/document_tools.py` - Added `_document_content_cache` (5 min TTL) and `_search_cache` (2 min TTL)
- `agent/tools/crm_tools.py` - Added caches for all lookup functions and SQL queries

---

### Phase 6: Node-Level Caching Analysis ⏭️ SKIPPED

**LangGraph CachePolicy** was evaluated for node-level caching but determined to provide minimal benefit on top of tool-level caching:

| Node Type | Cache Benefit | Reason |
|-----------|---------------|--------|
| Workers (LLM) | ❌ None | Input includes message history that changes every iteration |
| Router (LLM) | ❌ None | Same as workers - message history changes |
| Evaluators (LLM) | ❌ None | Same as workers - response content changes |
| Tools Node | ⚠️ Minimal | Would cache if identical tool calls in same order, but tool-level caching already handles this |

**Conclusion**: Tool-level caching (Phase 5) provides more granular and effective caching than node-level. Node caching is not needed.

---

## Files Modified Summary

| File | Change | Status |
|------|--------|--------|
| `agent/workers/career/job_search.py` | Remove model=LEAN_MODEL | ✅ Done |
| `server.py` | Add cancel message handling | ✅ Done |
| `orchestrator.py` | Add task tracking, cancel function, token budget | ✅ Done |
| `agent/tools/document_tools.py` | Add TTL caching | ✅ Done |
| `agent/tools/crm_tools.py` | Add TTL caching | ✅ Done |

---

## Testing

1. ✅ Run existing tests: `uv run pytest src/agent/__tests__/test_ai_chat.py -vs` (7 passed in 55s)
2. 🔲 Test job search manually - should complete in <10 iterations
3. ✅ Cancel button implemented - ready for manual testing
4. 🔲 Verify token usage stays under 50k for typical queries
5. 🔲 Test cache hits by calling same tool twice in one conversation

## Status

**Implementation Complete**

---

## Frontend Changes Implemented

| File | Change |
|------|--------|
| `hooks/useWs.jsx:92-100` | Handle `{"type": "cancelled"}` message |
| `components/Chat/Chat.jsx:500-507` | Stop button (shows when `isThinking`) |
| `components/Chat/Chat.module.css:511-520` | Stop button styles (red) |

**Protocol:**
- User clicks Stop → sends `{"type": "cancel", "uuid": "<thread_id>"}`
- Server cancels task → sends `{"type": "cancelled"}`
- Frontend shows "Generation stopped." message

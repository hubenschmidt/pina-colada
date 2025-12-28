# Proposed Refactor: LangGraph → OpenAI Agents SDK

## Status: Draft / Under Evaluation

**Created**: 2025-12-06

---

## Problem Statement

A simple CRM lookup ("look up William Hubenschmidt") consumes **~2.1k tokens** across **4 LLM calls**:

| Step | Node | Model | Tokens |
|------|------|-------|--------|
| 1 | Router | Haiku | ~300-400 |
| 2 | CRM Worker (tool call) | GPT-5.1 | ~967 |
| 3 | CRM Worker (format response) | GPT-5.1 | ~500 |
| 4 | Evaluator | Haiku | ~400-500 |

**Question**: Is LangGraph adding unnecessary overhead? Would OpenAI Agents SDK be leaner?

---

## Current Architecture: LangGraph

### LangGraph Dependencies Used

| Feature | Purpose | LOC Impact |
|---------|---------|------------|
| `StateGraph` | Graph builder | ~50 lines in orchestrator.py |
| `add_messages` reducer | Auto-append messages | 1 line, but core to state |
| `conditional_edges` | Dynamic routing | ~70 lines of routing logic |
| `ToolNode` | Pre-built tool executor | 4 nodes, ~20 lines |
| `MemorySaver` | Conversation threading | ~5 lines |
| `astream/ainvoke` | Async execution | ~60 lines streaming logic |

### What LangGraph Provides

1. **Message reducer** - Auto-appends messages to history
2. **Conditional routing** - Dynamic node-to-node routing via pure functions
3. **ToolNode** - Handles tool calling, execution, response formatting
4. **Checkpointing** - Thread-based conversation persistence
5. **Streaming** - Built-in `astream()` with `stream_mode="values"`

### What's LangGraph-Agnostic (Portable)

- Worker node logic (pure functions: state → state updates)
- Evaluator logic (pure functions with Pydantic output)
- Prompt builders
- Tool definitions (LangChain `StructuredTool`)
- Message types (`SystemMessage`, `HumanMessage`, `AIMessage`)

---

## Alternative: OpenAI Agents SDK

### Core Primitives

| Primitive | Purpose | Equivalent in Current System |
|-----------|---------|------------------------------|
| **Agent** | LLM + instructions + tools | Worker node |
| **Handoff** | Delegate to another agent | Router → Worker routing |
| **Tool** | Python function as tool | LangChain StructuredTool |
| **Guardrail** | Input/output validation | Evaluator node |
| **Session** | Conversation history | MemorySaver checkpointing |

### Potential Benefits

1. **Simpler multi-agent** - Handoffs instead of conditional edges
2. **Built-in sessions** - No manual state management
3. **Provider-agnostic** - Works with OpenAI, Anthropic, 100+ LLMs
4. **Built-in tracing** - Integrates with Langfuse, AgentOps, etc.
5. **Smaller abstraction surface** - Fewer concepts to learn

### Potential Drawbacks

1. **No built-in evaluator loop** - Would need custom retry logic
2. **Less explicit control** - Graph visualization/debugging harder
3. **Newer ecosystem** - Less community knowledge
4. **Migration effort** - Need to rewrite orchestrator.py entirely

---

## Agent Building Block Checklist

From Reddit discussion on agent library requirements:

| Building Block | LangGraph | OpenAI Agents SDK |
|----------------|-----------|-------------------|
| Memory (message history) | ✅ `add_messages` reducer | ✅ Sessions |
| Tools (function calling) | ✅ `ToolNode` + `bind_tools` | ✅ Native function tools |
| Executor (state machine) | ✅ StateGraph + conditional edges | ✅ Handoffs + Python orchestration |
| Parser (structured output) | ✅ `with_structured_output` | ✅ Pydantic `output_type` |
| Retry capability | ✅ Evaluator → Worker loop | ⚠️ Manual implementation |
| Multi-LLM support | ✅ LangChain model adapters | ✅ Provider-agnostic |
| Agent-agent interaction | ✅ Via state routing | ✅ Handoffs |
| Context passing | ✅ State TypedDict | ✅ Context dependency injection |
| Observability | ⚠️ Manual Langfuse integration | ✅ Built-in tracing |
| Evaluator | ✅ Built-in via nodes | ⚠️ Guardrails (input/output only) |

---

## Migration Effort Estimate

### Files to Rewrite

| File | Lines | Effort |
|------|-------|--------|
| `orchestrator.py` | ~500 | High - Complete rewrite |
| `graph.py` | ~50 | Medium - API compatibility |
| `routers/agent_router.py` | ~165 | Medium - Convert to handoff logic |

### Files to Adapt (Minor Changes)

| File | Lines | Effort |
|------|-------|--------|
| `workers/_base_worker.py` | ~150 | Low - Convert to Agent class |
| `workers/*.py` | ~50 each | Low - Update factory calls |
| `evaluators/_base_evaluator.py` | ~240 | Medium - Convert to guardrail or custom loop |

### Files Unchanged

- `prompts/*.py` - Prompt strings are portable
- `tools/*.py` - Tool logic is portable (need schema conversion)
- `util/*.py` - Utilities are framework-agnostic

### Estimated Total Effort

- **Rewrite**: ~700 LOC
- **Adapt**: ~500 LOC
- **Timeline**: 2-3 days for core migration, 1-2 days testing

---

## Token Usage Comparison (Theoretical)

| Scenario | LangGraph Current | OpenAI SDK Projected |
|----------|-------------------|----------------------|
| Simple CRM lookup | ~2.1k tokens (4 calls) | ~1.5k tokens (2-3 calls) |
| Job search | ~5k tokens (8+ calls) | ~3k tokens (4-5 calls) |
| Cover letter | ~3k tokens (6 calls) | ~2k tokens (3-4 calls) |

**Assumptions**:
- OpenAI SDK: Skip separate router (use handoffs)
- OpenAI SDK: Guardrails replace evaluator for simple cases
- Both: Same LLMs and prompts

---

## Recommendation

### Option A: Stay with LangGraph + Optimize

**Pros**:
- No migration risk
- Known patterns
- Per-worker ToolNodes already implemented
- Can still optimize (skip evaluator for CRM, fast-path router)

**Remaining optimizations**:
1. Fast-path regex router (~300 token savings)
2. Skip evaluator for simple data queries (~400 token savings)
3. Target: ~1.4k tokens for simple CRM lookup

### Option B: Migrate to OpenAI Agents SDK

**Pros**:
- Potentially simpler codebase
- Built-in tracing
- Modern API design

**Cons**:
- 2-3 day migration effort
- Risk of bugs during transition
- Less control over execution flow

### Option C: Hybrid Approach

Keep LangGraph for complex orchestration, but:
1. Evaluate OpenAI Agents SDK for new features
2. Consider gradual migration if SDK proves beneficial

---

## Decision Matrix

| Factor | Weight | LangGraph | OpenAI SDK |
|--------|--------|-----------|------------|
| Token efficiency | 30% | 6/10 | 7/10 |
| Development speed | 20% | 8/10 | 6/10 |
| Debugging/visibility | 15% | 8/10 | 7/10 |
| Ecosystem maturity | 15% | 9/10 | 6/10 |
| Migration risk | 20% | 10/10 | 4/10 |
| **Weighted Score** | | **7.7** | **6.1** |

---

## Next Steps

1. [ ] Benchmark actual token usage with current optimizations
2. [ ] Prototype single worker in OpenAI Agents SDK
3. [ ] Compare token usage and latency
4. [ ] Make go/no-go decision based on data

---

## References

- [OpenAI Agents SDK Docs](https://openai.github.io/openai-agents-python/)
- [OpenAI Agents SDK GitHub](https://github.com/openai/openai-agents-python)
- [LangGraph Documentation](https://langchain-ai.github.io/langgraph/)

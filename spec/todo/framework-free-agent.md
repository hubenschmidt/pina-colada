# Plan: Replace LangGraph with Framework-Free Agent Flow

## Goal
Remove LangGraph/Pregel, use pure Python async state machine with aggressive token optimization. Cross-session persistence required.

## Architecture

```
Orchestrator Loop (async)
    │
    ├─► intent_classifier ──► fast_lookup/count/list ──► format ──► END
    │
    └─► router ──► worker ──► tools ──► [complexity check]
                                              │
                              simple ◄────────┼────────► complex
                                 │                          │
                                END                    evaluator
                                                           │
                                                     ┌─────┴─────┐
                                                   pass        retry
                                                    │            │
                                                   END      worker (loop)
```

**Evaluator skipped for:**
- Fast-path intents (lookup/count/list)
- Single tool call completions
- CRM worker (always direct)

**Evaluator runs for:**
- Cover letter generation
- Multi-step analysis
- Job search with synthesis

**State machine dispatch:**
```python
async def run_agent(message, ctx) -> AgentState:
    state = initialize_state(message)
    state["messages"] = await load_messages(ctx.thread_id) + state["messages"]

    while not state["done"] and iteration < 20:
        update, next_node = await dispatch_node(state["route_to"], state, ctx)
        state = merge_state(state, update)
        state["route_to"] = next_node
```

## Minimal State Schema

```python
class AgentState(TypedDict):
    messages: List[Message]      # Plain list, no reducer
    route_to: Optional[str]      # Next node
    success_criteria: str
    feedback: Optional[str]      # Evaluator feedback
    done: bool
    fast_intent: Optional[str]   # "lookup"|"count"|"list"
    entity_type: Optional[str]
    query: Optional[str]
    result: Optional[str]
    tokens: Dict[str, int]       # {"in": 0, "out": 0}

class RequestContext(TypedDict):  # Immutable, not in state
    thread_id: str
    tenant_id: Optional[int]
    schema_context: str
    tools: Dict[str, List]       # Lazy-loaded per worker
```

**Dropped fields:** `evaluator_type`, `current_node`, `success_criteria_met`, `user_input_needed` (combined into `done`)

## Token Optimization Techniques

1. **Direct SDK calls** - No LangChain wrappers
2. **Lazy tool schemas** - Generate on demand, cache
3. **JSON mode** - Replace Pydantic `with_structured_output`
4. **Aggressive trimming** - 6k token message window, drop orphaned tool pairs
5. **Conditional evaluator** - Skip for simple tasks

**Complexity check (skip evaluator if):**
```python
def needs_evaluator(state: AgentState) -> bool:
    # Skip for simple workers
    if state["route_to"] in ("crm_worker",):
        return False
    # Skip if single tool call with result
    tool_calls = count_tool_calls(state["messages"])
    if tool_calls <= 1:
        return False
    # Run for complex tasks
    return state["route_to"] in ("cover_letter", "job_search") or tool_calls > 2
```

## Files to Create

| File | Purpose |
|------|---------|
| `agent/core/state.py` | TypedDicts, trim_messages(), merge_state() |
| `agent/core/loop.py` | run_agent(), dispatch_node() |
| `agent/core/llm.py` | call_openai(), call_anthropic() wrappers |
| `agent/core/tools.py` | ToolRegistry, execute_tool(), lazy schemas |
| `agent/core/persistence.py` | load_messages(), save_message() |

## Files to Modify

| File | Change |
|------|--------|
| `orchestrator.py` | Replace graph with loop import |
| `graph.py` | Simplify to delegate to loop |
| `workers/_base_worker.py` | Direct LLM calls |
| `evaluators/_base_evaluator.py` | JSON mode, no Pydantic |
| `routers/intent_classifier.py` | Direct Anthropic |
| `routers/agent_router.py` | Direct Anthropic |

## Dependencies to Remove

```toml
# pyproject.toml - DELETE
"langgraph>=1.0,<2",
"langgraph-cli[inmem]>=0.2.8",
```

## Implementation Order

1. Create `core/state.py` + `core/llm.py` (foundation)
2. Create `core/loop.py` with dispatch skeleton
3. Adapt `intent_classifier` + `agent_router` to direct calls
4. Adapt `_base_worker` to direct calls
5. Adapt `_base_evaluator` to JSON mode
6. Create `core/tools.py` with lazy schemas
7. Create `core/persistence.py` wrapping conversation_repository
8. Wire up `orchestrator.py` to new loop
9. Remove LangGraph imports + dependencies
10. Test end-to-end

## Estimated Savings

~40-50% token reduction per request from:
- No state serialization overhead (~500 tokens/iteration)
- No Pydantic message validation (~100 tokens/message)
- Worker-specific tool schemas (~200 tokens saved)
- **Skipped evaluator for simple tasks** (~800-1500 tokens saved per skipped eval)

---

## Future Consideration: Go Backend

If migrating backend from Python to Go, the state machine pattern ports cleanly. No official OpenAI Agents SDK for Go exists, but several options:

### Go Agent SDKs

| SDK | Repo | Notes |
|-----|------|-------|
| **nlpodyssey/openai-agents-go** | github.com/nlpodyssey/openai-agents-go | Community port of OpenAI Agents SDK |
| **pontus-devoteam/agent-sdk-go** | github.com/pontus-devoteam/agent-sdk-go | Multi-provider (OpenAI, Anthropic, LM Studio), Go 1.23+ |
| **Ingenimax/agent-sdk-go** | github.com/Ingenimax/agent-sdk-go | Multi-model, MCP integration, guardrails, observability |
| **Google ADK for Go** | (official) | Code-first toolkit, released Nov 2025 |

### Manual Loop (Recommended for Token Optimization)

Use `openai/openai-go` directly + port state machine:

```go
func RunAgent(ctx context.Context, message string, reqCtx RequestContext) (*AgentState, error) {
    state := initializeState(message)
    state.Messages = append(loadMessages(reqCtx.ThreadID), state.Messages...)

    for !state.Done && state.Iteration < 20 {
        update, nextNode := dispatchNode(state.RouteTo, state, reqCtx)
        state = mergeState(state, update)
        state.RouteTo = nextNode
        state.Iteration++
    }
    return state, nil
}
```

**Go advantages:**
- Goroutines + channels for streaming/parallel tool execution
- Lower memory footprint
- Single binary deployment
- No Python runtime overhead

**Trade-offs:**
- Rewrite workers, evaluators, tools
- Lose langchain-community tool integrations (Wikipedia, etc.)
- Need Go equivalents for existing Python libs

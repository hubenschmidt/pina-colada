"""
Orchestrator - functional implementation
Uses factory function pattern to create a graph with closed-over state
"""

import json
import logging
from typing import List, Any, Optional, Dict, Annotated

from typing_extensions import TypedDict
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import StateGraph, START, END
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode
from langchain_core.messages import AIMessage, HumanMessage, ToolMessage, trim_messages
from dotenv import load_dotenv

logger = logging.getLogger(__name__)
load_dotenv(override=True)


# State type definition
class State(TypedDict):
    messages: Annotated[List[Any], add_messages]
    success_criteria: str
    feedback_on_work: Optional[str]
    success_criteria_met: bool
    user_input_needed: bool
    score: Optional[int]  # Numeric score from 0-100 rating response quality
    resume_name: str
    route_to_agent: Optional[str]
    evaluator_type: Optional[str]  # career, scraper, or general
    token_usage: Optional[Dict[str, int]]  # current call: input, output, total
    token_usage_cumulative: Optional[Dict[str, int]]  # cumulative for entire request
    schema_context: Optional[str]  # CRM schema context from reasoning service
    tenant_id: Optional[int]  # Tenant ID for CRM queries


def _get_tool_call_ids(msg: Any) -> set:
    """Extract tool call IDs from a message."""
    if not isinstance(msg, AIMessage):
        return set()
    if not hasattr(msg, 'tool_calls') or not msg.tool_calls:
        return set()
    return {tc.get('id') for tc in msg.tool_calls}


def _find_pending_tool_calls(messages: List[Any]) -> set:
    """Find tool_call_ids that don't have responses."""
    pending = set()
    for msg in messages:
        pending.update(_get_tool_call_ids(msg))
        if isinstance(msg, ToolMessage):
            pending.discard(msg.tool_call_id)
    return pending


def _is_orphaned_tool_call(msg: Any, pending_ids: set) -> bool:
    """Check if a message contains an orphaned tool call."""
    call_ids = _get_tool_call_ids(msg)
    return bool(call_ids & pending_ids)


def _ensure_tool_pairs_intact(messages: List[Any]) -> List[Any]:
    """Ensure tool_call messages are followed by their tool responses."""
    if not messages:
        return messages

    pending = _find_pending_tool_calls(messages)
    if not pending:
        return messages

    logger.warning(f"Removing {len(pending)} orphaned tool calls")
    return [msg for msg in messages if not _is_orphaned_tool_call(msg, pending)]


def _trim_messages(messages: List[Any], max_tokens: int = 8000) -> List[Any]:
    """Pure function to trim message history"""
    try:
        trimmed = trim_messages(
            messages,
            max_tokens=max_tokens,
            strategy="last",
            token_counter=len,
            include_system=True,
            allow_partial=False,
        )

        if len(trimmed) < len(messages):
            logger.info(f"Trimmed messages from {len(messages)} to {len(trimmed)}")

        # Ensure tool call/response pairs are intact
        trimmed = _ensure_tool_pairs_intact(trimmed)

        return trimmed
    except Exception as e:
        logger.warning(f"Message trimming failed: {e}, using last 10 messages")
        return messages[-10:]


async def create_orchestrator():
    """
    Factory function that creates an orchestrator with closed-over state.

    Resume/document context is fetched dynamically by workers using document tools.

    Returns:
        - run_streaming: async function to execute the graph
        - set_websocket_sender: function to set WS sender
    """
    from agent.tools.worker_tools import get_worker_tools
    from agent.tools.crm_tools import get_crm_tools
    from agent.tools.document_tools import get_document_tools
    from agent.workers.general_worker import create_worker_node, route_from_worker
    from agent.workers.career import (
        create_job_search_node,
        route_from_job_search,
        create_cover_letter_writer_node,
        route_from_cover_letter_writer,
    )
    from agent.workers.crm import create_crm_worker_node, route_from_crm_worker
    from agent.routers.agent_router import create_router_node, route_from_router_edge
    from agent.evaluators import (
        route_from_evaluator,
        create_career_evaluator_node,
        create_general_evaluator_node,
        create_crm_evaluator_node,
    )
    from services.reasoning_service import get_reasoning_schema, format_schema_context
    from lib.tenant_context import set_tenant_id

    logger.info("=== AGENT SETUP ===")

    # Load tools
    base_tools = await get_worker_tools()
    document_tools = await get_document_tools()
    tools = base_tools + document_tools  # All workers get document tools
    logger.info(f"✓ Loaded {len(base_tools)} base tools + {len(document_tools)} document tools")

    # Load CRM schema context for RAG
    try:
        crm_schema = await get_reasoning_schema("crm")
        schema_context = format_schema_context(crm_schema)
        logger.info(f"✓ Loaded CRM schema context ({len(crm_schema)} tables)")
    except Exception as e:
        logger.warning(f"Could not load CRM schema: {e}")
        schema_context = "CRM schema not available."

    # Load CRM tools
    crm_tools = await get_crm_tools()
    all_crm_tools = tools + crm_tools  # CRM worker gets all tools including document + CRM
    logger.info(f"✓ Loaded {len(crm_tools)} CRM tools")

    # Create nodes (each returns a pure function with closed-over LLMs)
    # All workers now have access to document tools for dynamic resume/document fetching
    worker = await create_worker_node(tools, _trim_messages)
    job_search = await create_job_search_node(tools, _trim_messages)
    cover_letter_writer = await create_cover_letter_writer_node(tools, _trim_messages)
    crm_worker = await create_crm_worker_node(all_crm_tools, schema_context, _trim_messages)
    logger.info("✓ Created workers with document tool access")

    # Create specialized evaluators
    career_evaluator = await create_career_evaluator_node()
    general_evaluator = await create_general_evaluator_node()
    crm_evaluator = await create_crm_evaluator_node()
    logger.info("✓ Created specialized evaluators (career, general, crm)")

    # Create LLM-based router
    router = await create_router_node()
    logger.info("✓ Created LLM-based router")

    # Build graph
    logger.info("Building LangGraph workflow...")
    graph_builder = StateGraph(State)

    # Add nodes - now they're all pure functions
    logger.info("Adding nodes to graph...")
    graph_builder.add_node("router", router)
    graph_builder.add_node("worker", worker)
    graph_builder.add_node("job_search", job_search)
    graph_builder.add_node("cover_letter_writer", cover_letter_writer)
    graph_builder.add_node("crm_worker", crm_worker)
    graph_builder.add_node("tools", ToolNode(tools=all_crm_tools))  # Use all tools including CRM
    # Add specialized evaluators
    graph_builder.add_node("career_evaluator", career_evaluator)
    graph_builder.add_node("general_evaluator", general_evaluator)
    graph_builder.add_node("crm_evaluator", crm_evaluator)
    logger.info("✓ Nodes added")

    # Add edges
    logger.info("Adding edges to graph...")
    graph_builder.add_edge(START, "router")

    graph_builder.add_conditional_edges(
        "router",
        route_from_router_edge,
        {
            "worker": "worker",
            "job_search": "job_search",
            "cover_letter_writer": "cover_letter_writer",
            "crm_worker": "crm_worker",
        },
    )

    graph_builder.add_conditional_edges(
        "worker",
        route_from_worker,
        {"tools": "tools", "evaluator": "general_evaluator"},
    )

    graph_builder.add_conditional_edges(
        "job_search",
        route_from_job_search,
        {"tools": "tools", "evaluator": "career_evaluator"},
    )

    graph_builder.add_conditional_edges(
        "crm_worker",
        route_from_crm_worker,
        {"tools": "tools", "evaluator": "crm_evaluator"},
    )

    def route_from_tools(state: Dict[str, Any]) -> str:
        """Route back to the node that called tools."""
        route_to = state.get("route_to_agent", "worker")
        logger.info(f"→ Routing from tools back to {route_to}")
        return route_to

    graph_builder.add_conditional_edges(
        "tools",
        route_from_tools,
        {
            "worker": "worker",
            "job_search": "job_search",
            "cover_letter_writer": "cover_letter_writer",
            "crm_worker": "crm_worker",
        },
    )

    # Cover letter writer now has tools, so needs conditional routing
    graph_builder.add_conditional_edges(
        "cover_letter_writer",
        route_from_cover_letter_writer,
        {"tools": "tools", "evaluator": "career_evaluator"},
    )

    # Route from each evaluator back to worker or END
    evaluator_routing = {
        "worker": "worker",
        "job_search": "job_search",
        "cover_letter_writer": "cover_letter_writer",
        "crm_worker": "crm_worker",
        "END": END,
    }

    graph_builder.add_conditional_edges(
        "career_evaluator",
        route_from_evaluator,
        evaluator_routing,
    )

    graph_builder.add_conditional_edges(
        "general_evaluator",
        route_from_evaluator,
        evaluator_routing,
    )

    graph_builder.add_conditional_edges(
        "crm_evaluator",
        route_from_evaluator,
        evaluator_routing,
    )

    logger.info("✓ Edges added")

    # Compile graph
    logger.info("Compiling graph...")
    memory = MemorySaver()
    compiled_graph = graph_builder.compile(checkpointer=memory)

    if compiled_graph is None:
        raise RuntimeError("Graph compilation returned None")

    logger.info("✓ Graph compiled successfully")
    logger.info("===================")

    # Closed-over mutable state for WebSocket sender (only stateful thing we need)
    ws_sender_ref = {"send_ws": None}

    def set_websocket_sender(send_ws):
        """Set the WebSocket sender function"""
        ws_sender_ref["send_ws"] = send_ws

    async def run_streaming(
        message: str, thread_id: str, success_criteria: str = None, tenant_id: int = None
    ) -> Dict[str, Any]:
        """Run the graph with streaming support"""
        logger.info(f"▶️  Starting graph execution for thread: {thread_id}")
        logger.info(f"   User message: {message[:100]}...")

        # Set tenant context for CRM tools (request-scoped)
        set_tenant_id(tenant_id)
        if tenant_id:
            logger.info(f"   Tenant ID: {tenant_id}")

        config = {
            "configurable": {"thread_id": thread_id},
            "recursion_limit": 50,  # Higher limit for browser automation tasks
        }

        sc = (
            success_criteria
            or "Provide a clear and accurate response to the user's question"
        )

        state = {
            "messages": [HumanMessage(content=message)],
            "success_criteria": sc,
            "feedback_on_work": None,
            "success_criteria_met": False,
            "user_input_needed": False,
            "score": None,
            "resume_name": "William Hubenschmidt",
            "route_to_agent": None,
            "evaluator_type": None,
            "token_usage": None,
            "token_usage_cumulative": {"input": 0, "output": 0, "total": 0},
            "schema_context": schema_context,
            "tenant_id": tenant_id,
        }

        send_ws = ws_sender_ref["send_ws"]

        # Non-streaming path
        if not send_ws:
            result = await compiled_graph.ainvoke(state, config=config)
            logger.info("✅ Graph execution completed")
            return result

        # Streaming path
        await send_ws(json.dumps({"type": "start"}))

        last_content = ""
        iteration = 0
        cumulative = {"input": 0, "output": 0, "total": 0}

        async for event in compiled_graph.astream(
            state, config=config, stream_mode="values"
        ):
            iteration += 1
            logger.info(f"   Graph iteration {iteration}")

            messages = event.get("messages")
            last_msg = None
            try:
                last_msg = (messages or [])[-1]
            except Exception:
                last_msg = None

            current_content = getattr(last_msg, "content", None)
            is_ai = isinstance(last_msg, AIMessage)
            changed = (
                isinstance(current_content, str)
                and current_content
                and current_content != last_content
            )

            # Detect if this is a NEW response (content got shorter or replaced, not just appended)
            is_new_response = (
                changed
                and last_content
                and not current_content.startswith(last_content)
            )

            should_emit = is_ai and changed

            if should_emit:
                # Send start event for new response to reset frontend streaming state
                if is_new_response:
                    logger.info(
                        f"New response detected (content replaced), sending start event"
                    )
                    await send_ws(json.dumps({"type": "start"}))

                last_content = current_content
                await send_ws(
                    json.dumps(
                        {
                            "type": "content",
                            "content": current_content,
                            "is_final": False,
                        }
                    )
                )

            # Send token usage if available
            token_usage = event.get("token_usage")
            if token_usage and token_usage.get("total", 0) > 0:
                # Accumulate tokens
                cumulative["input"] += token_usage.get("input", 0)
                cumulative["output"] += token_usage.get("output", 0)
                cumulative["total"] += token_usage.get("total", 0)

                await send_ws(
                    json.dumps({
                        "type": "token_usage",
                        "token_usage": token_usage,
                        "cumulative": cumulative.copy()
                    })
                )

        await send_ws(
            json.dumps({"type": "content", "content": last_content, "is_final": True})
        )
        await send_ws(json.dumps({"type": "end"}))
        logger.info(f"✅ Graph execution completed in {iteration} iterations")

        return state

    # Return functions that use the closed-over state
    return {
        "run_streaming": run_streaming,
        "set_websocket_sender": set_websocket_sender,
    }

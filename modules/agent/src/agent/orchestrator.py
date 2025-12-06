"""
Orchestrator - functional implementation
Uses factory function pattern to create a graph with closed-over state
"""

import asyncio
import json
import logging
from typing import List, Any, Optional, Dict, Annotated
from uuid import UUID

from typing_extensions import TypedDict
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import StateGraph, START, END
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode
from langchain_core.messages import AIMessage, HumanMessage, ToolMessage, trim_messages
from dotenv import load_dotenv

from agent.tools.worker_tools import get_general_tools, get_job_search_tools
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
from services import conversation_service
from lib.tenant_context import set_tenant_id

logger = logging.getLogger(__name__)
load_dotenv(override=True)

# Track running tasks by thread_id for cancellation
_running_tasks: Dict[str, asyncio.Task] = {}

# Token budget limit (50k tokens max per request)
TOKEN_BUDGET = 50000


async def cancel_streaming(thread_id: str) -> bool:
    """Cancel a running streaming task by thread_id.

    Returns True if a task was cancelled, False otherwise.
    """
    task = _running_tasks.get(thread_id)
    if task and not task.done():
        logger.info(f"Cancelling streaming task for thread: {thread_id}")
        task.cancel()
        return True
    logger.warning(f"No running task found for thread: {thread_id}")
    return False


# State type definition
class State(TypedDict):
    messages: Annotated[List[Any], add_messages]
    success_criteria: str
    feedback_on_work: Optional[str]
    success_criteria_met: bool
    user_input_needed: bool
    score: Optional[int]  # Numeric score from 0-100 rating response quality
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


def _extract_content_info(last_msg: Any, last_content: str) -> Dict[str, Any]:
    """Extract content emission info from message. Returns dict with should_emit, content, is_new."""
    current_content = getattr(last_msg, "content", None)
    is_ai = isinstance(last_msg, AIMessage)

    if not isinstance(current_content, str):
        return {"should_emit": False, "content": last_content, "is_new": False}
    if not current_content:
        return {"should_emit": False, "content": last_content, "is_new": False}
    if current_content == last_content:
        return {"should_emit": False, "content": last_content, "is_new": False}
    if not is_ai:
        return {"should_emit": False, "content": current_content, "is_new": False}

    is_new = bool(last_content and not current_content.startswith(last_content))
    return {"should_emit": True, "content": current_content, "is_new": is_new}


def _should_accumulate_tokens(token_usage: Optional[Dict], last_id: Optional[tuple]) -> tuple:
    """Check if token usage should be accumulated. Returns (should_accumulate, token_id)."""
    if not token_usage:
        return (False, last_id)
    if token_usage.get("total", 0) <= 0:
        return (False, last_id)

    token_id = (
        token_usage.get("input", 0),
        token_usage.get("output", 0),
        token_usage.get("total", 0),
    )
    if token_id == last_id:
        return (False, last_id)

    return (True, token_id)


async def _stream_with_budget(stream, cumulative: Dict, budget: int):
    """Async generator that stops when budget exceeded. Avoids break in main loop."""
    async for event in stream:
        yield event
        if cumulative["total"] > budget:
            logger.warning(f"Token budget exceeded ({cumulative['total']} > {budget}), stopping")
            return


async def create_orchestrator():
    """
    Factory function that creates an orchestrator with closed-over state.

    Resume/document context is fetched dynamically by workers using document tools.

    Returns:
        - run_streaming: async function to execute the graph
        - set_websocket_sender: function to set WS sender
    """
    logger.info("=== AGENT SETUP ===")

    # Load tools - separated by purpose
    general_tools = await get_general_tools()  # web search, wikipedia
    document_tools = await get_document_tools()  # search_documents, get_document_content
    job_search_tools = await get_job_search_tools()  # job_search, check_applied_jobs, etc.
    crm_tools = await get_crm_tools()  # lookup_individual, lookup_organization, etc.

    logger.info(f"✓ Loaded tools: {len(general_tools)} general, {len(document_tools)} document, {len(job_search_tools)} job search, {len(crm_tools)} CRM")

    # Load CRM schema context for RAG
    try:
        crm_schema = await get_reasoning_schema("crm")
        schema_context = format_schema_context(crm_schema)
        logger.info(f"✓ Loaded CRM schema context ({len(crm_schema)} tables)")
    except Exception as e:
        logger.warning(f"Could not load CRM schema: {e}")
        schema_context = "CRM schema not available."

    # Build tool sets for each worker (no longer need all_tools combined)
    worker_tools = document_tools + general_tools  # General worker: docs + web search (NO job search)
    job_worker_tools = document_tools + job_search_tools  # Job search: docs + job tools
    crm_worker_tools = document_tools + crm_tools  # CRM: docs + CRM only (removed general_tools)
    cover_letter_tools = document_tools  # Cover letter: just docs

    # Create nodes - each worker gets appropriate tools
    worker = await create_worker_node(worker_tools, _trim_messages)
    job_search = await create_job_search_node(job_worker_tools, _trim_messages)
    cover_letter_writer = await create_cover_letter_writer_node(document_tools, _trim_messages)  # Only docs
    crm_worker = await create_crm_worker_node(crm_worker_tools, schema_context, _trim_messages)
    logger.info("✓ Created workers with appropriate tool access")

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
    # Per-worker ToolNodes to reduce token usage (only bind relevant tools)
    graph_builder.add_node("worker_tools", ToolNode(tools=worker_tools))
    graph_builder.add_node("job_tools", ToolNode(tools=job_worker_tools))
    graph_builder.add_node("crm_tools", ToolNode(tools=crm_worker_tools))
    graph_builder.add_node("cover_letter_tools", ToolNode(tools=cover_letter_tools))
    logger.info(f"✓ Created per-worker ToolNodes: worker({len(worker_tools)}), job({len(job_worker_tools)}), crm({len(crm_worker_tools)}), cover_letter({len(cover_letter_tools)})")
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

    # Route each worker to its dedicated ToolNode
    graph_builder.add_conditional_edges(
        "worker",
        route_from_worker,
        {"tools": "worker_tools", "evaluator": "general_evaluator"},
    )

    graph_builder.add_conditional_edges(
        "job_search",
        route_from_job_search,
        {"tools": "job_tools", "evaluator": "career_evaluator"},
    )

    graph_builder.add_conditional_edges(
        "crm_worker",
        route_from_crm_worker,
        {"tools": "crm_tools", "evaluator": "crm_evaluator"},
    )

    graph_builder.add_conditional_edges(
        "cover_letter_writer",
        route_from_cover_letter_writer,
        {"tools": "cover_letter_tools", "evaluator": "career_evaluator"},
    )

    # Route from each ToolNode back to its worker
    graph_builder.add_edge("worker_tools", "worker")
    graph_builder.add_edge("job_tools", "job_search")
    graph_builder.add_edge("crm_tools", "crm_worker")
    graph_builder.add_edge("cover_letter_tools", "cover_letter_writer")

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
        message: str,
        thread_id: str,
        success_criteria: str = None,
        tenant_id: int = None,
        user_id: int = None,
    ) -> Dict[str, Any]:
        """Run the graph with streaming support"""
        logger.info(f"▶️  Starting graph execution for thread: {thread_id}")
        logger.info(f"   User message: {message[:100]}...")

        # Register this task for cancellation support
        current_task = asyncio.current_task()
        if current_task:
            _running_tasks[thread_id] = current_task

        # Set tenant context for CRM tools (request-scoped)
        set_tenant_id(tenant_id)
        if tenant_id:
            logger.info(f"   Tenant ID: {tenant_id}")

        # Check if this is a new conversation (for title generation)
        is_new_conversation = False
        thread_uuid = None
        if user_id and tenant_id:
            try:
                thread_uuid = UUID(thread_id)
                existing = await conversation_service.get_conversation(thread_uuid)
            except Exception:
                existing = None
                is_new_conversation = True
            if not existing:
                is_new_conversation = True

            # Save user message
            try:
                await conversation_service.add_message(
                    thread_id=thread_uuid,
                    user_id=user_id,
                    tenant_id=tenant_id,
                    role="user",
                    content=message,
                )
            except Exception as e:
                logger.warning(f"Failed to save user message: {e}")

        config = {
            "configurable": {"thread_id": thread_id},
            "recursion_limit": 20,  # Reduced from 50 to prevent runaway iterations
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
        last_token_usage_id = None
        cancelled = False

        raw_stream = compiled_graph.astream(state, config=config, stream_mode="values")
        budget_stream = _stream_with_budget(raw_stream, cumulative, TOKEN_BUDGET)

        try:
            async for event in budget_stream:
                iteration += 1
                logger.info(f"   Graph iteration {iteration}")

                messages = event.get("messages")
                last_msg = (messages or [None])[-1]

                content_info = _extract_content_info(last_msg, last_content)
                last_content = content_info["content"]

                if content_info["should_emit"] and content_info["is_new"]:
                    logger.info("New response detected, sending start event")
                    await send_ws(json.dumps({"type": "start"}))

                if content_info["should_emit"]:
                    await send_ws(json.dumps({
                        "type": "content",
                        "content": content_info["content"],
                        "is_final": False,
                    }))

                token_usage = event.get("token_usage")
                should_accumulate, last_token_usage_id = _should_accumulate_tokens(
                    token_usage, last_token_usage_id
                )

                if should_accumulate:
                    cumulative["input"] += token_usage.get("input", 0)
                    cumulative["output"] += token_usage.get("output", 0)
                    cumulative["total"] += token_usage.get("total", 0)

                    await send_ws(json.dumps({
                        "type": "token_usage",
                        "token_usage": token_usage,
                        "cumulative": cumulative.copy()
                    }))

        except asyncio.CancelledError:
            cancelled = True
            logger.info(f"⚠️  Graph execution cancelled for thread: {thread_id}")
            await send_ws(json.dumps({"type": "cancelled"}))
            raise  # Re-raise to properly cancel the task

        finally:
            # Cleanup task tracking
            _running_tasks.pop(thread_id, None)

            if not cancelled:
                await send_ws(
                    json.dumps({"type": "content", "content": last_content, "is_final": True})
                )
                await send_ws(json.dumps({"type": "end"}))
                logger.info(f"✅ Graph execution completed in {iteration} iterations")

                # Save assistant message
                if user_id and tenant_id and last_content and thread_uuid:
                    try:
                        await conversation_service.add_message(
                            thread_id=thread_uuid,
                            user_id=user_id,
                            tenant_id=tenant_id,
                            role="assistant",
                            content=last_content,
                            token_usage=cumulative if cumulative.get("total", 0) > 0 else None,
                        )

                        # Generate title for new conversations
                        if is_new_conversation:
                            conversation_service.schedule_title_generation(
                                thread_uuid, message, last_content
                            )
                    except Exception as e:
                        logger.warning(f"Failed to save assistant message: {e}")

        return state

    # Return functions that use the closed-over state
    return {
        "run_streaming": run_streaming,
        "set_websocket_sender": set_websocket_sender,
    }

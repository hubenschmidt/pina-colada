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
from langchain_core.messages import AIMessage, HumanMessage, trim_messages
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
    resume_name: str
    resume_context: str
    route_to_agent: Optional[str]


def _build_resume_context(
    summary: str, resume_text: str, sample_answers: str, cover_letters: list
) -> str:
    """
    Pure function to build full resume context.
    Uses single space separator to avoid newline rendering issues in LangSmith.
    """
    context_parts = []

    if summary:
        context_parts.append(f"SUMMARY: {summary}")
    if resume_text:
        context_parts.append(f"RESUME: {resume_text}")
    if sample_answers:
        context_parts.append(f"SAMPLE_ANSWERS: {sample_answers}")
    if cover_letters:
        context_parts.append(f"COVER_LETTERS: {cover_letters}")

    # Join with space instead of newlines for cleaner rendering
    return " | ".join(context_parts)


def _build_resume_context_concise(resume_text: str, summary: str) -> str:
    """
    Pure function to build concise resume context.
    Uses single space separator for clean rendering.
    """
    context_parts = []

    if summary:
        context_parts.append(f"SUMMARY: {summary}")
    if resume_text:
        preview = resume_text[:500] + "..." if len(resume_text) > 500 else resume_text
        context_parts.append(f"RESUME: {preview}")

    # Join with space for cleaner rendering
    return " | ".join(context_parts)


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

        return trimmed
    except Exception as e:
        logger.warning(f"Message trimming failed: {e}, using last 10 messages")
        return messages[-10:]


async def create_orchestrator(
    resume_text: str = "",
    summary: str = "",
    sample_answers: str = "",
    cover_letters: list = None,
):
    """
    Factory function that creates an orchestrator with closed-over state.

    Returns:
        - run_streaming: async function to execute the graph
        - set_websocket_sender: function to set WS sender
    """
    from agent.tools.worker_tools import get_worker_tools
    from agent.tools.mcp_playwright import PLAYWRIGHT_MCP_TOOLS, init_playwright_mcp
    from agent.tools.scraper_tools import SCRAPER_TOOLS
    from agent.workers.worker import create_worker_node, route_from_worker
    from agent.workers.job_hunter import create_job_hunter_node, route_from_job_hunter
    from agent.workers.evaluator import create_evaluator_node, route_from_evaluator
    from agent.workers.cover_letter_writer import create_cover_letter_writer_node
    from agent.workers.scraper import create_scraper_node, route_from_scraper
    from agent.routers.agent_router import route_to_agent, route_from_router_edge

    logger.info("=== AGENT SETUP ===")

    # Build context
    cover_letters = cover_letters or []
    resume_context = _build_resume_context(
        summary, resume_text, sample_answers, cover_letters
    )
    resume_context_concise = _build_resume_context_concise(resume_text, summary)

    # Load tools
    tools = await get_worker_tools()
    logger.info(f"✓ Loaded {len(tools)} tools")

    # Initialize Playwright MCP
    logger.info("Initializing Playwright MCP...")
    await init_playwright_mcp()
    logger.info("✓ Playwright MCP initialized")

    # Combine regular tools with Playwright MCP tools
    all_tools = tools + PLAYWRIGHT_MCP_TOOLS + SCRAPER_TOOLS
    logger.info(f"✓ Total tools (including Playwright MCP + static scraping): {len(all_tools)}")

    # Scraper gets both Playwright (for browser automation) and static scraping tools
    scraper_tools = PLAYWRIGHT_MCP_TOOLS + SCRAPER_TOOLS
    logger.info(f"✓ Scraper tools: {len(scraper_tools)} (Playwright MCP + static scraping)")

    # Create nodes (each returns a pure function with closed-over LLMs)
    worker = await create_worker_node(tools, resume_context_concise, _trim_messages)
    job_hunter = await create_job_hunter_node(
        tools, resume_context_concise, _trim_messages
    )
    scraper = await create_scraper_node(scraper_tools, _trim_messages)
    evaluator = await create_evaluator_node()
    cover_letter_writer = await create_cover_letter_writer_node(_trim_messages)

    # Build graph
    logger.info("Building LangGraph workflow...")
    graph_builder = StateGraph(State)

    # Add nodes - now they're all pure functions
    logger.info("Adding nodes to graph...")
    graph_builder.add_node("router", route_to_agent)
    graph_builder.add_node("worker", worker)
    graph_builder.add_node("job_hunter", job_hunter)
    graph_builder.add_node("scraper", scraper)
    graph_builder.add_node("cover_letter_writer", cover_letter_writer)
    graph_builder.add_node("tools", ToolNode(tools=all_tools))
    graph_builder.add_node("evaluator", evaluator)
    logger.info("✓ Nodes added")

    # Add edges
    logger.info("Adding edges to graph...")
    graph_builder.add_edge(START, "router")

    graph_builder.add_conditional_edges(
        "router",
        route_from_router_edge,
        {
            "worker": "worker",
            "job_hunter": "job_hunter",
            "scraper": "scraper",
            "cover_letter_writer": "cover_letter_writer",
        },
    )

    graph_builder.add_conditional_edges(
        "worker",
        route_from_worker,
        {"tools": "tools", "evaluator": "evaluator"},
    )

    graph_builder.add_conditional_edges(
        "job_hunter",
        route_from_job_hunter,
        {"tools": "tools", "evaluator": "evaluator"},
    )

    graph_builder.add_conditional_edges(
        "scraper",
        route_from_scraper,
        {"tools": "tools", "END": END},
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
            "job_hunter": "job_hunter",
            "scraper": "scraper",
            "cover_letter_writer": "cover_letter_writer",
        },
    )
    graph_builder.add_edge("cover_letter_writer", "evaluator")

    graph_builder.add_conditional_edges(
        "evaluator",
        route_from_evaluator,
        {
            "worker": "worker",
            "job_hunter": "job_hunter",
            "scraper": "scraper",
            "cover_letter_writer": "cover_letter_writer",
            "END": END,
        },
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
        message: str, thread_id: str, success_criteria: str = None
    ) -> Dict[str, Any]:
        """Run the graph with streaming support"""
        logger.info(f"▶️  Starting graph execution for thread: {thread_id}")
        logger.info(f"   User message: {message[:100]}...")

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
            "resume_name": "William Hubenschmidt",
            "resume_context": resume_context,
            "route_to_agent": None,
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

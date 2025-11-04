"""
LangGraph Chat Application - Functional Implementation
=====================================================
Pure functional approach using factory functions and closures.

Graph Flow:
    START → router → [worker → tools] → evaluator → [retry or END]
"""

__all__ = ["graph", "invoke_graph", "build_default_graph"]

import json
import logging
import os
from typing import Dict, Any, Optional, Union, List
from fastapi import WebSocket
from langfuse import observe
from dotenv import load_dotenv
from agent.orchestrator import create_orchestrator
from agent.util.document_scanner import load_documents
from agent.util.get_success_criteria import get_success_criteria
from agent.tools.static_tools import record_user_context

logger = logging.getLogger("app.graph")
load_dotenv(override=True)

# Load resume documents
RESUME_NAME = "William Hubenschmidt"
logger.info("Loading documents from me/ directory...")
resume_text, summary, cover_letters, sample_answers = load_documents("me")
logger.info(
    f"Documents loaded - Resume: {len(resume_text)} chars, "
    f"Summary: {len(summary)} chars, "
    f"Sample answers: {len(sample_answers)} chars, "
    f"Cover letters: {len(cover_letters)}"
)

# Orchestrator instance (will be lazily initialized)
_orchestrator_ref = {"instance": None}


async def get_orchestrator():
    """Get or create the orchestrator instance"""
    if _orchestrator_ref["instance"] is None:
        logger.info("Initializing Orchestrator...")
        _orchestrator_ref["instance"] = await create_orchestrator(
            resume_text=resume_text,
            summary=summary,
            sample_answers=sample_answers,
            cover_letters=cover_letters,
        )
        logger.info("Orchestrator initialized successfully")

    return _orchestrator_ref["instance"]


import json
import logging
from typing import Callable, Awaitable

logger = logging.getLogger(__name__)


def make_websocket_stream_adapter(websocket) -> Callable[[str], Awaitable[None]]:
    """
    Returns an async function `send(payload_str: str)` that adapts orchestrator
    messages to the frontend format, maintaining internal streaming state.
    """
    last_sent_length = 0  # captured by the closure

    async def handle_start(_):
        return  # ignore

    async def handle_end(_):
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))

    async def handle_content(payload):
        nonlocal last_sent_length
        content = payload.get("content", "")
        delta = content[last_sent_length:]
        if not delta:  # guard
            return
        last_sent_length = len(content)
        await websocket.send_text(json.dumps({"on_chat_model_stream": delta}))

    handlers = {
        "start": handle_start,
        "end": handle_end,
        "content": handle_content,
    }

    async def send(payload_str: str):
        try:
            payload = json.loads(payload_str)
        except json.JSONDecodeError:
            logger.warning("Could not parse payload: %s", payload_str)
            return

        msg_type = payload.get("type", "")
        handler = handlers.get(msg_type)

        if not handler:
            logger.warning("Unknown message type: %s", msg_type)
            return

        await handler(payload)

    return send


def build_default_graph():
    """
    Build a placeholder graph for LangGraph API compatibility.
    The actual work is done by the orchestrator.
    """
    from langgraph.graph import StateGraph, START, END
    from typing import TypedDict, List, Dict, Any

    class SimpleState(TypedDict):
        messages: List[Dict[str, Any]]

    def passthrough(state: SimpleState) -> Dict[str, Any]:
        return {}

    graph_builder = StateGraph(SimpleState)
    graph_builder.add_node("passthrough", passthrough)
    graph_builder.add_edge(START, "passthrough")
    graph_builder.add_edge("passthrough", END)

    logger.info("Built placeholder graph for LangGraph API")
    return graph_builder.compile()


# Graph instance for LangGraph compatibility
graph = build_default_graph()


@observe()
async def invoke_graph(
    websocket: WebSocket,
    data: Union[str, Dict[str, Any], List[Dict[str, str]]],
    user_uuid: str,
):
    """Main entrypoint for invoking the graph via WebSocket"""

    # Normalize input
    payload: Optional[Dict[str, Any]] = None

    if isinstance(data, dict):
        payload = data
    elif isinstance(data, str):
        try:
            parsed = json.loads(data)
            if isinstance(parsed, dict):
                payload = parsed
        except Exception:
            pass

    # Handle control messages
    if isinstance(payload, dict) and payload.get("init") is True:
        return

    if isinstance(payload, dict) and payload.get("type") in (
        "user_context",
        "user_context_update",
    ):
        if not isinstance((payload.get("ctx") or {}), dict):
            return

        ctx = payload.get("ctx") or {}
        headers = {k.decode().lower(): v.decode() for k, v in websocket.headers.raw}

        server_block = ctx.get("server") or {}
        server_block.update(
            {
                "user_uuid": user_uuid,
                "ua": headers.get("user-agent"),
                "origin": headers.get("origin"),
            }
        )
        ctx["server"] = server_block

        logger.info(os.getenv("ENVIRONMENT"))
        if os.getenv("ENVIRONMENT") != "development":
            record_user_context(ctx)
        return

    # Extract message
    message = str(data)
    if isinstance(payload, dict) and "message" in payload:
        message = str(payload["message"])

    # Get orchestrator
    try:
        orchestrator = await get_orchestrator()
    except Exception as e:
        logger.error(f"Failed to get orchestrator: {e}", exc_info=True)
        await websocket.send_text(
            json.dumps(
                {
                    "on_chat_model_stream": "Sorry, there was an error initializing the assistant."
                }
            )
        )
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))
        return

    # Set up WebSocket adapter
    send = make_websocket_stream_adapter(websocket)
    orchestrator["set_websocket_sender"](send)

    # Generate success criteria
    success_criteria = get_success_criteria(message)
    logger.info(f"Generated success criteria: {success_criteria[:60]}...")

    # Run orchestrator
    try:
        logger.info("Invoking orchestrator...")
        await orchestrator["run_streaming"](
            message=message,
            thread_id=user_uuid,
            success_criteria=success_criteria,
        )
        logger.info("Orchestrator completed successfully")
    except Exception as e:
        logger.error(f"Error invoking orchestrator: {e}", exc_info=True)
        await websocket.send_text(
            json.dumps(
                {
                    "on_chat_model_stream": "Sorry, there was an error generating the response."
                }
            )
        )
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))

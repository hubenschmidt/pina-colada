"""
LangGraph Chat Application with Orchestrator Agent
===============================================
This module creates an intelligent chatbot using the Orchestrator worker/evaluator pattern.

Graph Flow (via Orchestrator):
    START → worker → [tools] → evaluator → [worker or END]

Key Components:
    - Orchestrator: Worker/evaluator agent pattern
    - Document loading from me/ directory
    - WebSocket streaming support
"""

__all__ = ["graph", "invoke_our_graph", "build_default_graph"]

import json
import logging
import os
from typing import Dict, Any, Optional, Union, List
from fastapi import WebSocket
from langfuse import observe
from dotenv import load_dotenv
from agent.orchestrator import Orchestrator
from agent.util.document_scanner import load_documents
from agent.util.get_success_criteria import get_success_criteria
from agent.tools.static_tools import record_user_context

# =============================================================================
# CONFIGURATION
# =============================================================================

logger = logging.getLogger("app.graph")
load_dotenv(override=True)

# Load resume documents
RESUME_NAME = "William Hubenschmidt"
logger.info("Loading documents from me/ directory...")
resume_text, summary, cover_letters, sample_answers = load_documents("me")
logger.info(
    f"Documents loaded - Resume: {len(resume_text)} chars, Summary: {len(summary)} chars, Sample answers: {len(sample_answers)} chars, Cover letters: {len(cover_letters)}"
)

# =============================================================================
# ORCHESTRATOR INSTANCE (Singleton)
# =============================================================================

_orchestrator_instance: Optional[Orchestrator] = None


async def get_orchestrator() -> Orchestrator:
    """Get or create the Orchestrator instance."""
    global _orchestrator_instance

    if _orchestrator_instance is None:
        logger.info("Initializing Orchestrator...")
        _orchestrator_instance = Orchestrator(
            resume_text=resume_text,
            summary=summary,
            sample_answers=sample_answers,
            cover_letters=cover_letters,
        )
        await _orchestrator_instance.setup()
        logger.info("Orchestrator initialized successfully")

    return _orchestrator_instance


class WebSocketStreamAdapter:
    """
    Adapts Orchestrator's streaming format to match the frontend's expectations.

    Frontend expects:
        - {"on_chat_model_stream": "content"} for each chunk
        - {"on_chat_model_end": true} when done

    Orchestrator sends:
        - {"type": "start"} at beginning
        - {"type": "content", "content": "...", "is_final": false} for chunks
        - {"type": "end"} at end
    """

    def __init__(self, websocket: WebSocket):
        self.websocket = websocket
        self.current_content = ""
        self.last_sent_length = 0

    async def send(self, payload_str: str):
        """Receive from Orchestrator and translate to frontend format."""
        try:
            payload = json.loads(payload_str)
        except json.JSONDecodeError:
            logger.warning("Could not parse payload: %s", payload_str)
            return

        msg_type = payload.get("type", "")

        async def handle_start(_: dict):
            return  # intentionally ignore

        async def handle_end(_: dict):
            await self.websocket.send_text(json.dumps({"on_chat_model_end": True}))

        async def handle_content(p: dict):
            content = p.get("content", "")
            delta = content[self.last_sent_length :]
            if not delta:
                return
            self.last_sent_length = len(content)
            await self.websocket.send_text(json.dumps({"on_chat_model_stream": delta}))

        async def handle_unknown(_: dict):
            logger.warning("Unknown message type: %s", msg_type)

        handlers = {
            "start": handle_start,
            "end": handle_end,
            "content": handle_content,
        }

        await handlers.get(msg_type, handle_unknown)(payload)


# =============================================================================
# DEFAULT EXPORT (Graph factory for LangGraph compatibility)
# =============================================================================


def build_default_graph():
    """
    Build a simple graph for LangGraph API compatibility.
    This returns a minimal graph structure that LangGraph can load.
    The actual work is done by Orchestrator in invoke_our_graph.
    """
    from langgraph.graph import StateGraph, START, END
    from typing import TypedDict, List, Dict, Any

    class SimpleState(TypedDict):
        messages: List[Dict[str, Any]]

    def passthrough(state: SimpleState) -> Dict[str, Any]:
        """Simple passthrough node"""
        return {}

    graph_builder = StateGraph(SimpleState)
    graph_builder.add_node("passthrough", passthrough)
    graph_builder.add_edge(START, "passthrough")
    graph_builder.add_edge("passthrough", END)

    logger.info("Built placeholder graph for LangGraph API")
    return graph_builder.compile()


# Graph instance - this is what LangGraph loads
graph = build_default_graph()


@observe()
async def invoke_our_graph(
    websocket: WebSocket,
    data: Union[str, Dict[str, Any], List[Dict[str, str]]],
    user_uuid: str,
):
    """
    Main entrypoint for invoking the graph via WebSocket.
    """

    # =========================================================================
    # NORMALIZE INPUT
    # =========================================================================
    payload: Optional[Dict[str, Any]] = None

    if isinstance(data, dict):
        payload = data

    parsed = None
    if isinstance(data, str):
        try:
            parsed = json.loads(data)
        except Exception:
            parsed = None
    if isinstance(parsed, dict):
        payload = parsed

    # =========================================================================
    # HANDLE CONTROL MESSAGES
    # =========================================================================

    # init/handshake
    if isinstance(payload, dict) and payload.get("init") is True:
        return

    # invalid ctx when it *is* a user_context control
    if (
        isinstance(payload, dict)
        and payload.get("type") in ("user_context", "user_context_update")
        and not isinstance((payload.get("ctx") or {}), dict)
    ):
        return

    # user_context control — process and return
    if isinstance(payload, dict) and payload.get("type") in (
        "user_context",
        "user_context_update",
    ):
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

    # =========================================================================
    # EXTRACT MESSAGE AND INVOKE ORCHESTRATOR
    # =========================================================================

    message = str(data)
    if isinstance(payload, dict) and "message" in payload:
        message = str(payload["message"])

    try:
        orchestrator = await get_orchestrator()
    except Exception as e:
        logger.error(f"Failed to get Orchestrator: {e}", exc_info=True)
        await websocket.send_text(
            json.dumps(
                {
                    "on_chat_model_stream": "Sorry, there was an error initializing the assistant."
                }
            )
        )
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))
        return

    adapter = WebSocketStreamAdapter(websocket)
    orchestrator.set_websocket_sender(adapter.send)

    # Generate context-aware success criteria based on the user's message
    success_criteria = get_success_criteria(message)
    logger.info(
        f"Generated success criteria for message type: {success_criteria[:60]}..."
    )

    try:
        logger.info("Invoking Orchestrator...")
        await orchestrator.run_streaming(
            message=message,
            thread_id=user_uuid,
            success_criteria=success_criteria,
        )
        logger.info("Orchestrator completed successfully")
    except Exception as e:
        logger.error(f"Error invoking Orchestrator: {e}", exc_info=True)
        await websocket.send_text(
            json.dumps(
                {
                    "on_chat_model_stream": "Sorry, there was an error generating the response."
                }
            )
        )
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))
        return

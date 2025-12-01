"""
PinaColada.co | LangGraph
=====================================================

Graph Flow:
    START → router → [worker → tools] → evaluator → [retry or END]
"""

__all__ = ["graph", "invoke_graph", "build_default_graph"]

import json
import logging
from typing import Dict, Any, Optional, Callable, Awaitable
from fastapi import WebSocket
from dotenv import load_dotenv
from agent.orchestrator import create_orchestrator
from agent.util.document_scanner import load_documents
from agent.util.get_success_criteria import get_success_criteria
from agent.tools.static_tools import record_user_context

logger = logging.getLogger(__name__)
load_dotenv(override=True)

from langfuse import observe


def _is_disconnect_error(err: Exception) -> bool:
    """Check if exception indicates websocket disconnect."""
    error_name = type(err).__name__.lower()
    error_msg = str(err).lower()
    return (
        "disconnect" in error_name
        or "close" in error_name
        or "disconnect" in error_msg
        or "close" in error_msg
    )

# Load resume documents
RESUME_NAME = "William Hubenschmidt"
logger.info("Loading documents from me/ directory...")
resume_text, summary, sample_answers, cover_letters = load_documents("me")
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


class _WebSocketStreamState:
    """Encapsulates websocket streaming state."""

    def __init__(self, websocket):
        self.websocket = websocket
        self.last_sent_length = 0
        self.disconnected = False

    async def safe_send(self, data: dict) -> bool:
        """Safely send JSON to websocket, return False if disconnected."""
        if self.disconnected:
            return False
        try:
            await self.websocket.send_text(json.dumps(data))
            return True
        except Exception as e:
            if _is_disconnect_error(e):
                self.disconnected = True
                logger.debug(f"WebSocket disconnected: {type(e).__name__}")
                return False
            raise


async def _handle_stream_start(state: _WebSocketStreamState, _payload: dict) -> None:
    """Handle stream start event."""
    logger.info(f"Stream start - resetting length from {state.last_sent_length} to 0")
    if state.last_sent_length > 0:
        await state.safe_send({"on_chat_model_start": True})
    state.last_sent_length = 0


async def _handle_stream_end(state: _WebSocketStreamState, _payload: dict) -> None:
    """Handle stream end event."""
    await state.safe_send({"on_chat_model_end": True})


async def _handle_stream_content(state: _WebSocketStreamState, payload: dict) -> None:
    """Handle stream content event."""
    content = payload.get("content", "")
    delta = content[state.last_sent_length:]
    if delta:
        state.last_sent_length = len(content)
        await state.safe_send({"on_chat_model_stream": delta})


async def _handle_token_usage(state: _WebSocketStreamState, payload: dict) -> None:
    """Handle token usage event."""
    token_usage = payload.get("token_usage", {})
    if token_usage:
        await state.safe_send({
            "on_token_usage": token_usage,
            "on_token_cumulative": payload.get("cumulative", {})
        })


_STREAM_HANDLERS = {
    "start": _handle_stream_start,
    "end": _handle_stream_end,
    "content": _handle_stream_content,
    "token_usage": _handle_token_usage,
}


def make_websocket_stream_adapter(websocket) -> Callable[[str], Awaitable[None]]:
    """Returns an async send function that adapts orchestrator messages to frontend format."""
    state = _WebSocketStreamState(websocket)

    async def send(payload_str: str):
        if state.disconnected:
            return

        try:
            payload = json.loads(payload_str)
        except json.JSONDecodeError:
            logger.warning("Could not parse payload: %s", payload_str)
            return

        msg_type = payload.get("type", "")
        handler = _STREAM_HANDLERS.get(msg_type)
        if not handler:
            logger.warning("Unknown message type: %s", msg_type)
            return

        try:
            await handler(state, payload)
        except Exception as e:
            if _is_disconnect_error(e):
                logger.debug(f"WebSocket disconnected during handler: {type(e).__name__}")

    return send


async def _send_error_to_websocket(websocket) -> None:
    """Send error message to websocket, handling disconnection gracefully."""
    try:
        await websocket.send_text(json.dumps({
            "on_chat_model_stream": "\n\nSorry, there was an error generating the response."
        }))
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))
    except Exception as send_err:
        err_type = type(send_err).__name__
        if _is_disconnect_error(send_err):
            logger.debug("Could not send error message, client already disconnected")
            return
        logger.debug(f"Could not send error message: {err_type}")


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


# --- user context handler: guard clauses, no nesting -------------------------
async def _handle_user_context(
    websocket: WebSocket, payload: Dict[str, Any], user_uuid: str
) -> bool:
    if payload.get("type") not in {"user_context", "user_context_update"}:
        return False  # not a context message; let caller continue

    ctx = payload.get("ctx")
    if not isinstance(ctx, dict):
        return True  # we handled (and chose to ignore) invalid ctx

    headers = {k.decode().lower(): v.decode() for k, v in websocket.headers.raw}

    server_block = dict(ctx.get("server") or {})
    server_block.update(
        {
            "user_uuid": user_uuid,
            "ua": headers.get("user-agent"),
            "origin": headers.get("origin"),
        }
    )
    ctx["server"] = server_block

    try:
        record_user_context(ctx)
    except Exception:
        logger.error("record_user_context failed", exc_info=True)

    return True  # consumed


def _to_dict(d) -> Optional[Dict[str, Any]]:
    if isinstance(d, dict):
        return d
    if not isinstance(d, str):
        return None
    try:
        parsed = json.loads(d)
    except Exception:
        return None
    if isinstance(parsed, dict):
        return parsed
    return None


@observe()
async def invoke_graph(
    websocket: WebSocket,
    data: Dict[str, Any] | str | list[dict[str, str]],
    user_uuid: str,
):
    # Normalize once; from here on, payload is a dict
    payload: Dict[str, Any] = _to_dict(data) or {}

    # init control
    if payload.get("init") is True:
        return

    # user context
    handled = await _handle_user_context(websocket, payload, user_uuid)
    if handled:
        return

    # message
    message = str(payload.get("message", data))

    # Test triggers (disabled by default - uncomment to enable)
    # from agent.util.test_triggers import handle_test_error_trigger
    # if await handle_test_error_trigger(websocket, message):
    #     return

    # orchestrator
    orchestrator = await get_orchestrator()

    send = make_websocket_stream_adapter(websocket)
    set_sender = orchestrator.get("set_websocket_sender")
    set_sender(send)

    # run
    success_criteria = get_success_criteria(message)
    logger.info(f"Generated success criteria: {success_criteria[:60]}...")
    try:
        await orchestrator["run_streaming"](
            message=message, thread_id=user_uuid, success_criteria=success_criteria
        )
    except Exception as e:
        if _is_disconnect_error(e):
            logger.info(f"Client disconnected during processing: {type(e).__name__}")
            return

        logger.error(f"Error invoking orchestrator: {e}", exc_info=True)
        await _send_error_to_websocket(websocket)

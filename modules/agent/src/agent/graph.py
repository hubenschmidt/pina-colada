"""
LangGraph Chat Application with Sidekick Agent
===============================================
This module creates an intelligent chatbot using the Sidekick worker/evaluator pattern.

Graph Flow (via Sidekick):
    START → worker → [tools] → evaluator → [worker or END]

Key Components:
    - Sidekick: Worker/evaluator agent pattern
    - Document loading from me/ directory
    - WebSocket streaming support
"""

__all__ = ["graph", "invoke_our_graph", "build_default_graph", "build_streaming_graph"]

import json
import logging
import uuid
from typing import Dict, Any, Optional, Callable, Awaitable, Union, List
from fastapi import WebSocket
from langfuse import observe
from dotenv import load_dotenv

from agent.sidekick import Sidekick
from agent.document_scanner import load_documents

# =============================================================================
# CONFIGURATION
# =============================================================================

logger = logging.getLogger("app.graph")
load_dotenv(override=True)

# Load resume documents
RESUME_NAME = "William Hubenschmidt"
logger.info("Loading documents from me/ directory...")
resume_text, summary, cover_letters = load_documents("me")
logger.info(
    f"Documents loaded - Resume: {len(resume_text)} chars, Summary: {len(summary)} chars, Cover letters: {len(cover_letters)}"
)

# =============================================================================
# SIDEKICK INSTANCE (Singleton)
# =============================================================================

_sidekick_instance: Optional[Sidekick] = None


async def get_sidekick() -> Sidekick:
    """Get or create the Sidekick instance."""
    global _sidekick_instance

    if _sidekick_instance is None:
        logger.info("Initializing Sidekick...")
        _sidekick_instance = Sidekick(
            resume_text=resume_text, summary=summary, cover_letters=cover_letters
        )
        await _sidekick_instance.setup()
        logger.info("Sidekick initialized successfully")

    return _sidekick_instance


# =============================================================================
# USER CONTEXT RECORDING (for analytics)
# =============================================================================


def record_user_context(ctx: Dict[str, Any]) -> Dict[str, Any]:
    """
    Record user context (browser info, device info, etc) for analytics.
    This is a simplified version - you can expand it as needed.
    """
    try:
        ip = ((ctx.get("server") or {}).get("ip")) or None
        logger.info(f"User context received from IP: {ip}")
        # You can add pushover notifications here if needed
        return {"ok": True}
    except Exception as e:
        logger.error(f"record_user_context failed: {e}")
        return {"ok": False, "error": str(e)}


def extract_client_ip(headers: Dict[str, str], peer_host: str) -> str:
    """Extract the real client IP from headers (handling proxies/CDNs)."""
    xff = headers.get("x-forwarded-for")

    if not xff:
        return peer_host

    parts = [p.strip() for p in xff.split(",") if p.strip()]
    if not parts:
        return peer_host

    return parts[-1]


# =============================================================================
# WEBSOCKET STREAMING ADAPTER
# =============================================================================


class WebSocketStreamAdapter:
    """
    Adapts Sidekick's streaming format to match the frontend's expectations.

    Frontend expects:
        - {"on_chat_model_stream": "content"} for each chunk
        - {"on_chat_model_end": true} when done

    Sidekick sends:
        - {"type": "start"} at beginning
        - {"type": "content", "content": "...", "is_final": false} for chunks
        - {"type": "end"} at end
    """

    def __init__(self, websocket: WebSocket):
        self.websocket = websocket
        self.current_content = ""
        self.last_sent_length = 0

    async def send(self, payload_str: str):
        """Receive from Sidekick and translate to frontend format."""
        try:
            payload = json.loads(payload_str)
        except json.JSONDecodeError:
            logger.warning(f"Could not parse payload: {payload_str}")
            return

        msg_type = payload.get("type")

        # Guard: Start message - ignore
        if msg_type == "start":
            logger.info("Stream starting...")
            return

        # Guard: End message - send end signal
        if msg_type == "end":
            logger.info("Stream ending...")
            await self.websocket.send_text(json.dumps({"on_chat_model_end": True}))
            return

        # Guard: Content message - stream the delta
        if msg_type == "content":
            content = payload.get("content", "")

            # Calculate what's new since last send
            if len(content) > self.last_sent_length:
                delta = content[self.last_sent_length :]
                self.last_sent_length = len(content)

                # Send in frontend format
                await self.websocket.send_text(
                    json.dumps({"on_chat_model_stream": delta})
                )
                logger.info(f"Streamed {len(delta)} chars")

            return

        # Unknown message type
        logger.warning(f"Unknown message type: {msg_type}")


# =============================================================================
# DEFAULT EXPORT (Graph factory for LangGraph compatibility)
# =============================================================================


def build_default_graph():
    """
    Build a simple graph for LangGraph API compatibility.
    This returns a minimal graph structure that LangGraph can load.
    The actual work is done by Sidekick in invoke_our_graph.
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


# =============================================================================
# WEBSOCKET ENTRYPOINT
# =============================================================================


def build_streaming_graph(send_ws: Callable[[str], Awaitable[None]]):
    """
    Build streaming graph - actually just returns the send_ws function.
    The real work is done by Sidekick in invoke_our_graph.
    """
    logger.info("build_streaming_graph called (returns send_ws wrapper)")
    return send_ws


@observe()
async def invoke_our_graph(
    websocket: WebSocket,
    data: Union[str, Dict[str, Any], List[Dict[str, str]]],
    user_uuid: str,
):
    """
    Main entrypoint for invoking the graph via WebSocket.

    This function:
    1. Handles control messages (init, user_context)
    2. Gets the Sidekick instance
    3. Sets up WebSocket streaming adapter
    4. Invokes Sidekick with the user's message

    Args:
        websocket: FastAPI WebSocket connection
        data: Message data (string or dict or list)
        user_uuid: Unique user identifier for this conversation
    """
    logger.info(
        f"invoke_our_graph recv: type={type(data).__name__} "
        f"keys={list(data.keys()) if isinstance(data, dict) else 'n/a'}"
    )

    # =========================================================================
    # NORMALIZE INPUT
    # =========================================================================

    payload: Optional[Dict[str, Any]] = None

    # Parse string to dict
    if isinstance(data, str):
        try:
            payload = json.loads(data)
        except Exception:
            payload = None
    elif isinstance(data, dict):
        payload = data

    # =========================================================================
    # HANDLE CONTROL MESSAGES
    # =========================================================================

    # Guard: Init/handshake message
    if isinstance(payload, dict) and payload.get("init") is True:
        logger.info("Init message received")
        return

    # Guard: User context message
    if isinstance(payload, dict) and payload.get("type") in (
        "user_context",
        "user_context_update",
    ):
        ctx = payload.get("ctx") or {}

        # Attach server-known info
        headers = {k.decode().lower(): v.decode() for k, v in websocket.headers.raw}
        client_ip = extract_client_ip(
            headers, websocket.client.host if websocket.client else ""
        )

        ctx.setdefault("server", {})
        ctx["server"].update(
            {
                "ip": client_ip,
                "user_uuid": user_uuid,
                "ua": headers.get("user-agent"),
                "origin": headers.get("origin"),
            }
        )

        record_user_context(ctx)
        return

    # =========================================================================
    # EXTRACT MESSAGE AND INVOKE SIDEKICK
    # =========================================================================

    # Extract message content
    if isinstance(payload, dict) and "message" in payload:
        message = str(payload["message"])
    else:
        message = str(data)

    logger.info(f"Processing message: {message}")
    logger.info(f"Thread ID: {user_uuid}")

    # Get Sidekick instance
    try:
        sidekick = await get_sidekick()
    except Exception as e:
        logger.error(f"Failed to get Sidekick: {e}", exc_info=True)
        await websocket.send_text(
            json.dumps(
                {
                    "on_chat_model_stream": "Sorry, there was an error initializing the assistant."
                }
            )
        )
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))
        return

    # Set up streaming adapter
    adapter = WebSocketStreamAdapter(websocket)
    sidekick.set_websocket_sender(adapter.send)

    # Invoke Sidekick
    try:
        logger.info("Invoking Sidekick...")
        result = await sidekick.run_streaming(
            message=message,
            thread_id=user_uuid,
            success_criteria="Provide a clear, accurate, and helpful response to the user's question",
        )
        logger.info(f"Sidekick completed successfully")
    except Exception as e:
        logger.error(f"Error invoking Sidekick: {e}", exc_info=True)
        await websocket.send_text(
            json.dumps(
                {
                    "on_chat_model_stream": "Sorry, there was an error generating the response."
                }
            )
        )
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))

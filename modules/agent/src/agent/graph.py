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

__all__ = ["graph", "invoke_our_graph", "build_default_graph"]

import json
import logging
import os
from typing import Dict, Any, Optional, Union, List
from fastapi import WebSocket
from langfuse import observe
from dotenv import load_dotenv
from agent.sidekick import Sidekick
from agent.util.document_scanner import load_documents
from agent.tools.sidekick_tools import push

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
# SIDEKICK INSTANCE (Singleton)
# =============================================================================

_sidekick_instance: Optional[Sidekick] = None


async def get_sidekick() -> Sidekick:
    """Get or create the Sidekick instance."""
    global _sidekick_instance

    if _sidekick_instance is None:
        logger.info("Initializing Sidekick...")
        _sidekick_instance = Sidekick(
            resume_text=resume_text,
            summary=summary,
            sample_answers=sample_answers,
            cover_letters=cover_letters,
        )
        await _sidekick_instance.setup()
        logger.info("Sidekick initialized successfully")

    return _sidekick_instance


# =============================================================================
# USER CONTEXT RECORDING (for analytics)
# =============================================================================


# --- Pretty push with blacklist check ---
def record_user_context(ctx: Dict[str, Any]):
    """Persist raw user context (testing: push as a notification)"""

    try:
        # Always log full JSON for debugging
        pretty_full = json.dumps(ctx, indent=2, ensure_ascii=False, sort_keys=True)

        # Compose body within practical push size (e.g., Pushover ~1024 chars)
        max_len = 1024
        body = pretty_full[:max_len]
        # Add ellipsis if we actually truncated
        body += "…" * (len(pretty_full) > max_len)

        push(body)
        return {"ok": True, "pushed": True}
    except Exception as e:
        logger.error(f"record_user_context failed: {e}")
        return {"ok": False, "error": str(e)}


# =============================================================================
# SUCCESS CRITERIA GENERATION
# =============================================================================


def generate_success_criteria(message: str) -> str:
    """
    Generate context-aware success criteria based on the user's message.
    This gives the evaluator clear, measurable criteria to check.
    """
    message_lower = message.lower()

    # 1. CONTACT/EMAIL QUERIES
    if any(word in message_lower for word in ["email", "contact", "reach", "touch"]):
        return """Success criteria:
- Provide William's email address or contact method
- Response is brief and direct
- No unnecessary elaboration"""

    # 2. COVER LETTER REQUESTS
    if any(
        phrase in message_lower
        for phrase in ["cover letter", "write a letter", "application letter"]
    ):
        return """Success criteria:
- Cover letter is properly formatted (greeting, body paragraphs, closing)
- References specific job details from the posting
- Uses William's actual experience from resume data
- Matches professional tone from sample cover letters
- Between 250-400 words
- No markdown formatting (plain text only)"""

    # 3. JOB SEARCH REQUESTS
    if (
        any(
            phrase in message_lower
            for phrase in ["job search", "find jobs", "job postings", "job openings"]
        )
        or ("find" in message_lower and "jobs" in message_lower)
        or ("search" in message_lower and "jobs" in message_lower)
        or ("looking for" in message_lower and "jobs" in message_lower)
    ):
        return """Success criteria:
- Provides at least 3-5 specific job postings
- Each posting includes: Company name, Job title, and DIRECT URL to application page
- URLs are actual job posting links (not general career pages)
- Jobs match William's skills and experience level
- Jobs are recent (posted within last 30 days if possible)
- Jobs are in requested location (default: NYC)"""

    # 4. "TELL ME ABOUT" QUESTIONS (check this before background questions)
    if "tell me about" in message_lower:
        return """Success criteria:
- Directly answers what was asked about
- Provides specific details from resume data
- Structured response (3-5 sentences)
- Stays on topic
- Professional and conversational tone"""

    # 5. TECHNICAL/EXPERIENCE QUESTIONS
    if any(
        word in message_lower
        for word in [
            "experience",
            "skills",
            "worked with",
            "know about",
            "projects",
            "built",
        ]
    ):
        return """Success criteria:
- Answer directly addresses the specific skill/experience asked about
- Cites specific examples or projects from William's resume
- Response is concise (2-4 sentences typically)
- Accurate information from resume data only
- No invented or assumed information"""

    # 5. TECHNICAL/EXPERIENCE QUESTIONS
    if any(
        word in message_lower
        for word in [
            "experience",
            "skills",
            "worked with",
            "know about",
            "projects",
            "built",
        ]
    ):
        return """Success criteria:
- Answer directly addresses the specific skill/experience asked about
- Cites specific examples or projects from William's resume
- Response is concise (2-4 sentences typically)
- Accurate information from resume data only
- No invented or assumed information"""

    # 6. BACKGROUND/CAREER QUESTIONS
    if any(
        word in message_lower
        for word in [
            "background",
            "career",
            "education",
            "worked at",
            "previous",
            "history",
            "about you",
        ]
    ):
        return """Success criteria:
- Provides accurate information from resume/summary
- Organized and easy to understand
- Focuses on most relevant information
- 3-5 sentences maximum unless more detail explicitly requested
- Professional tone"""

    # 7. COMPARISON QUESTIONS (X vs Y, better at, prefer)
    if any(
        word in message_lower
        for word in [
            " vs ",
            " versus ",
            "compared to",
            "better at",
            "prefer",
            "difference between",
        ]
    ):
        return """Success criteria:
- Addresses both items being compared
- Provides specific examples or experience with each
- Clear statement of preference/strength if applicable
- Based on actual resume information
- Concise comparison (2-4 sentences)"""

    # 8. GREETING/CASUAL
    if any(
        word in message_lower
        for word in [
            "hi",
            "hello",
            "hey",
            "greetings",
            "good morning",
            "good afternoon",
        ]
    ):
        return """Success criteria:
- Warm, professional greeting response
- Brief introduction of who William is
- Offers to help with questions
- 2-3 sentences maximum
- No repeated greeting if already established in conversation"""

    # 9. AVAILABILITY/HIRING QUESTIONS
    if any(
        phrase in message_lower
        for phrase in [
            "available",
            "looking for work",
            "open to",
            "hiring",
            "can you start",
            "when can you",
        ]
    ):
        return """Success criteria:
- Clear statement about current availability status
- Information about ideal role type if applicable
- Brief and direct (1-3 sentences)
- Professional and positive tone"""

    # 10. SALARY/COMPENSATION QUESTIONS
    if any(
        word in message_lower
        for word in ["salary", "compensation", "pay", "rate", "wage"]
    ):
        return """Success criteria:
- Professional deflection to discuss during interview process
- OR provide salary range if explicitly available in resume data
- Brief response (1-2 sentences)
- Maintains negotiating position"""

    # DEFAULT: Generic but more specific than before
    return """Success criteria:
- Directly answers the user's question
- Uses accurate information from resume data only
- Response is appropriately concise (typically 2-5 sentences)
- Professional and conversational tone
- Plain text format (no markdown)
- If information isn't in resume data, uses record_unknown_question tool"""


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

        # Start message - ignore
        if msg_type == "start":
            return

        # End message - send end signal
        if msg_type == "end":
            await self.websocket.send_text(json.dumps({"on_chat_model_end": True}))
            return

        # Content message - stream the delta
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
    # EXTRACT MESSAGE AND INVOKE SIDEKICK
    # =========================================================================

    message = str(data)
    if isinstance(payload, dict) and "message" in payload:
        message = str(payload["message"])

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

    adapter = WebSocketStreamAdapter(websocket)
    sidekick.set_websocket_sender(adapter.send)

    # Generate context-aware success criteria based on the user's message
    success_criteria = generate_success_criteria(message)
    logger.info(
        f"Generated success criteria for message type: {success_criteria[:60]}..."
    )

    try:
        logger.info("Invoking Sidekick...")
        await sidekick.run_streaming(
            message=message,
            thread_id=user_uuid,
            success_criteria=success_criteria,
        )
        logger.info("Sidekick completed successfully")
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
        return

"""
LangGraph Chat Application with OpenAI
=======================================
This module creates a chatbot that answers questions about a resume using LangGraph and OpenAI.

Graph Flow:
    START → conditional → model → END

Key Components:
    - ChatState: Defines the conversation state structure
    - SimpleChatGraph: Main class that builds and runs the graph
    - Tool Functions: Functions the AI can call (record_user_details, record_unknown_question)
"""

# Explicitly export the graph variable
__all__ = ["graph", "invoke_our_graph", "build_default_graph", "build_streaming_graph"]

import os
import sys
import json
import logging
import uuid
import requests
import re
from typing import TypedDict, List, Dict, Any, Optional, Callable, Awaitable, Union

from dotenv import load_dotenv
from openai import AsyncOpenAI
from pypdf import PdfReader
from ipaddress import (
    ip_address,
    ip_network,
    IPv4Address,
    IPv6Address,
    IPv4Network,
    IPv6Network,
)
from langgraph.graph import START, END, StateGraph
from langgraph.checkpoint.memory import MemorySaver
from fastapi import WebSocket
from langfuse import observe


# =============================================================================
# CONFIGURATION
# =============================================================================

logger = logging.getLogger("app.graph")
load_dotenv(override=True)

# OpenAI Configuration
API_KEY = os.getenv("OPENAI_API_KEY")
if not API_KEY:
    logger.fatal("Fatal Error: OPENAI_API_KEY is missing.")
    sys.exit(1)

MODEL_NAME = os.getenv("OPENAI_MODEL", "gpt-5-chat-latest")
TEMPERATURE = float(os.getenv("OPENAI_TEMPERATURE", "0"))
MAX_TOKENS_ENV = os.getenv("OPENAI_MAX_TOKENS")
MAX_TOKENS = None if not MAX_TOKENS_ENV else int(MAX_TOKENS_ENV)

# Resume Configuration
RESUME_NAME = "William Hubenschmidt"

# Pushover Configuration (for notifications)
PUSHOVER_USER = os.getenv("PUSHOVER_USER")
PUSHOVER_TOKEN = os.getenv("PUSHOVER_TOKEN")
PUSHOVER_URL = "https://api.pushover.net/1/messages.json"

# IP Blacklist Configuration
IP_BLACKLIST_STR = os.getenv("IP_BLACKLIST", "127.0.0.1, 10.0.0.0/8, 172.18.0.1")


# =============================================================================
# LOAD RESUME AND SUMMARY
# =============================================================================


def load_resume() -> str:
    """Load and clean resume text from PDF file."""
    try:
        logger.info(f"Loading resume from me/resume.pdf")
        reader = PdfReader("me/resume.pdf")

        resume_text = ""
        for page in reader.pages:
            text = page.extract_text()
            if text:
                resume_text += text

        # Clean up excessive whitespace
        resume_text = re.sub(r"\n\s*\n", "\n\n", resume_text)
        resume_text = re.sub(r" +", " ", resume_text)
        resume_text = resume_text.strip()

        logger.info(f"Resume loaded successfully: {len(resume_text)} characters")
        return resume_text

    except FileNotFoundError:
        logger.warning("Resume PDF not found at me/resume.pdf")
        return "[Resume not available]"

    except Exception as e:
        logger.error(f"Could not load resume PDF: {e}")
        return "[Resume not available]"


def load_summary() -> str:
    """Load summary text from file."""
    try:
        with open("me/summary.txt", "r", encoding="utf-8") as f:
            summary = f.read()
        logger.info(f"Summary loaded successfully: {len(summary)} characters")
        return summary

    except FileNotFoundError:
        logger.warning("Summary file not found at me/summary.txt")
        return "[Summary not available]"

    except Exception as e:
        logger.error(f"Could not load summary: {e}")
        return "[Summary not available]"


resume_text = load_resume()
summary = load_summary()

logger.info(f"MODEL_NAME: {MODEL_NAME}")
logger.info(f"Resume text length: {len(resume_text)} characters")
logger.info(f"Summary length: {len(summary)} characters")


# =============================================================================
# SYSTEM PROMPT
# =============================================================================

SYSTEM_PROMPT = f"""You are acting as {RESUME_NAME}. You are answering questions on {RESUME_NAME}'s website,
particularly questions related to {RESUME_NAME}'s career, background, skills and experience. Your responsibility
is to represent {RESUME_NAME} for interactions on the website as faithfully as possible. You are given a summary
of {RESUME_NAME}'s background and a copy of his resume which you can use to answer questions. Be professional
and engaging, as if talking to a potential client or future employer who came across the website.

IMPORTANT KNOWLEDGE:
You have direct access to {RESUME_NAME}'s complete resume and summary below. Use this information to answer all
questions about his career, work history, skills, and experience. Do not claim you don't have access to this information.

STYLE RULES (MUST FOLLOW):
- Output MUST be plain text only. Do NOT use Markdown or any formatting.
- Do NOT use asterisks, underscores, backticks, tildes, hashes, brackets, angle brackets, emojis, or any special characters for styling.
- Do NOT bold, italicize, add headings, bullet points, numbered lists, tables, code fences, links, or inline formatting of any kind.
- If the user sends Markdown or HTML, reply in plain text without reproducing the formatting.
- Keep paragraphs short and readable. Use newlines only; no decorative characters.

BEHAVIOR:
Your responses should always be as concise as possible. 
If you don't know the answer to any question, use your record_unknown_question tool to record the question that you couldn't answer,
even if it's about something trivial or unrelated to career. Ask for their name and email address so you can follow up, and use the
record_user_details tool to store it if they provide it.

If the user asks questions that do not directly pertain to {RESUME_NAME}'s career, background, skills, and experience,
do not answer them; briefly steer the conversation back to those topics.

SUMMARY:
{summary}

RESUME:
{resume_text}

With this context, chat with the user, always staying in character as {RESUME_NAME}."""


# =============================================================================
# NOTIFICATION HELPER
# =============================================================================


def send_push_notification(message: str) -> None:
    """Send a push notification via Pushover service."""
    print(f"Push: {message}")

    if not PUSHOVER_USER or not PUSHOVER_TOKEN:
        logger.warning("Pushover credentials not configured")
        return

    try:
        payload = {"user": PUSHOVER_USER, "token": PUSHOVER_TOKEN, "message": message}
        response = requests.post(PUSHOVER_URL, data=payload, timeout=5)
        logger.info(f"Push notification sent: {response.status_code}")

    except Exception as e:
        logger.error(f"Failed to send push notification: {e}")


# =============================================================================
# TOOL FUNCTIONS (AI can call these)
# =============================================================================


def record_user_details(
    email: str, name: str = "Name not provided", notes: str = "not provided"
) -> Dict[str, Any]:
    """
    Record user contact details when they express interest.

    Args:
        email: User's email address
        name: User's name (optional)
        notes: Additional context about the conversation (optional)

    Returns:
        Dictionary confirming the details were recorded
    """
    send_push_notification(
        f"Recording interest from {name} with email {email} and notes {notes}"
    )
    return {"recorded": "ok", "email": email, "name": name}


def record_unknown_question(question: str) -> Dict[str, Any]:
    """
    Record questions that couldn't be answered.

    Args:
        question: The question that couldn't be answered

    Returns:
        Dictionary confirming the question was recorded
    """
    send_push_notification(f"Recording question that couldn't be answered: {question}")
    return {"recorded": "ok", "question": question}


# =============================================================================
# TOOL DEFINITIONS FOR OPENAI
# =============================================================================

TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "record_user_details",
            "description": "Use this tool to record that a user is interested in being in touch and provided an email address",
            "parameters": {
                "type": "object",
                "properties": {
                    "email": {
                        "type": "string",
                        "description": "The email address of this user",
                    },
                    "name": {
                        "type": "string",
                        "description": "The user's name, if they provided it",
                    },
                    "notes": {
                        "type": "string",
                        "description": "Any additional information about the conversation that's worth recording to give context",
                    },
                },
                "required": ["email"],
                "additionalProperties": False,
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "record_unknown_question",
            "description": "Always use this tool to record any question that couldn't be answered as you didn't know the answer",
            "parameters": {
                "type": "object",
                "properties": {
                    "question": {
                        "type": "string",
                        "description": "The question that couldn't be answered",
                    }
                },
                "required": ["question"],
                "additionalProperties": False,
            },
        },
    },
]

# Map tool names to their functions
TOOL_MAP = {
    "record_user_details": record_user_details,
    "record_unknown_question": record_unknown_question,
}


# =============================================================================
# IP BLACKLIST UTILITIES
# =============================================================================


def parse_ip_blacklist(raw: str) -> tuple:
    """
    Parse IP blacklist string into a tuple of IP addresses and networks.

    Args:
        raw: Comma-separated string of IPs and CIDR ranges

    Returns:
        Tuple of IP addresses and networks
    """
    items = []

    for token in (t.strip() for t in raw.split(",") if t.strip()):
        # Try parsing as CIDR network
        try:
            items.append(ip_network(token, strict=False))
            continue
        except Exception:
            pass

        # Try parsing as single IP
        try:
            items.append(ip_address(token))
        except Exception:
            pass

    return tuple(items)


def is_ip_blacklisted(ip_str: Optional[str], blacklist: tuple) -> bool:
    """
    Check if an IP address is in the blacklist.

    Args:
        ip_str: IP address string to check
        blacklist: Tuple of blacklisted IPs and networks

    Returns:
        True if IP is blacklisted, False otherwise
    """
    if not ip_str:
        return False

    try:
        ip = ip_address(ip_str)
    except Exception:
        return False

    for item in blacklist:
        if isinstance(item, (IPv4Address, IPv6Address)):
            if ip == item:
                return True
        elif isinstance(item, (IPv4Network, IPv6Network)):
            if ip in item:
                return True

    return False


IP_BLACKLIST = parse_ip_blacklist(IP_BLACKLIST_STR)


# =============================================================================
# USER CONTEXT RECORDING
# =============================================================================


def record_user_context(ctx: Dict[str, Any]) -> Dict[str, Any]:
    """
    Record user context (browser info, device info, etc) for analytics.

    Args:
        ctx: Dictionary containing user context information

    Returns:
        Dictionary with success status
    """
    try:
        ip = ((ctx.get("server") or {}).get("ip")) or None

        # Always log full JSON for debugging
        logger.info(f"User context received: {json.dumps(ctx, indent=2)}")

        # Check if IP is blacklisted
        if is_ip_blacklisted(ip, IP_BLACKLIST):
            logger.info(f"Skipping push for blacklisted IP: {ip}")
            return {"ok": True, "skipped": "blacklisted_ip"}

        # Send notification about new visitor
        schema = ctx.get("schema", "unknown")
        ua = ctx.get("ua", "unknown")
        url = ctx.get("url", "unknown")
        message = f"New visitor [{schema}]\nUA: {ua}\nURL: {url}\nIP: {ip}"
        send_push_notification(message)

        return {"ok": True}

    except Exception as e:
        logger.error(f"record_user_context failed: {e}")
        return {"ok": False, "error": str(e)}


def extract_client_ip(headers: Dict[str, str], peer_host: str) -> str:
    """
    Extract the real client IP from headers (handling proxies/CDNs).

    Args:
        headers: HTTP headers dictionary
        peer_host: Direct peer host address

    Returns:
        Client IP address string
    """
    # Trust rightmost hop from X-Forwarded-For if proxy/CDN sets it
    xff = headers.get("x-forwarded-for")

    if not xff:
        return peer_host

    parts = [p.strip() for p in xff.split(",") if p.strip()]
    if not parts:
        return peer_host

    # Last element is closest to server
    candidate = parts[-1]

    try:
        ip_address(candidate)
        return candidate
    except Exception:
        return peer_host


# =============================================================================
# STATE DEFINITION
# =============================================================================


class ChatState(TypedDict):
    """
    Defines the structure of the conversation state.

    Attributes:
        messages: List of conversation messages (user, assistant, tool)
    """

    messages: List[Dict[str, str]]


# =============================================================================
# MESSAGE HELPERS
# =============================================================================


def normalize_messages_to_openai_format(messages: List[Any]) -> List[Dict[str, Any]]:
    """
    Convert various message formats to OpenAI's expected format.

    Args:
        messages: List of messages in various formats

    Returns:
        List of messages in OpenAI format
    """
    normalized = []

    for msg in messages:
        # Already in correct format
        if isinstance(msg, dict) and "role" in msg and "content" in msg:
            normalized.append(msg)
            continue

        # Tool call or tool response format
        if isinstance(msg, dict) and ("tool_calls" in msg or "tool_call_id" in msg):
            normalized.append(msg)
            continue

        # Convert object to dict format
        role = getattr(msg, "role", None) or "user"
        content = getattr(msg, "content", None) or str(msg)
        normalized.append({"role": role, "content": content})

    return normalized


def ensure_system_prompt(
    messages: List[Dict[str, Any]], system_text: str
) -> List[Dict[str, Any]]:
    """
    Ensure system message is always first and up-to-date.

    Args:
        messages: List of conversation messages
        system_text: System prompt text to use

    Returns:
        Messages list with system prompt at the beginning
    """
    # Remove any existing system messages
    non_system_msgs = [m for m in messages if m.get("role") != "system"]

    # Prepend fresh system prompt
    return [{"role": "system", "content": system_text}] + non_system_msgs


# =============================================================================
# GRAPH CLASS
# =============================================================================


class SimpleChatGraph:
    """
    Main LangGraph class that handles the conversation flow.

    This class:
    1. Manages the OpenAI client
    2. Defines graph nodes (conditional, model)
    3. Builds the graph structure
    4. Handles tool execution
    """

    def __init__(self, use_memory: bool):
        """
        Initialize the graph.

        Args:
            use_memory: Whether to use MemorySaver for conversation persistence
        """
        self.id = str(uuid.uuid4())
        self._client: Optional[AsyncOpenAI] = None
        self._memory = MemorySaver() if use_memory else None

    def _get_or_create_client(self) -> AsyncOpenAI:
        """Get existing OpenAI client or create a new one."""
        if self._client is None:
            self._client = AsyncOpenAI()
        return self._client

    # =========================================================================
    # NODE: Conditional (Easter Egg)
    # =========================================================================

    async def node_conditional(
        self, state: ChatState, send_ws: Optional[Callable[[str], Awaitable[None]]]
    ) -> Dict[str, Any]:
        """
        Check for easter egg keywords in the message.

        This node looks for specific keywords (LangChain, LangGraph) and
        sends a special easter egg notification if found.

        Args:
            state: Current conversation state
            send_ws: WebSocket send function (optional)

        Returns:
            Empty dict (no state changes)
        """
        logger.info(
            f"node_conditional called - state has {len(state.get('messages', []))} messages"
        )

        # Guard: No messages to check
        if not state.get("messages"):
            logger.info("node_conditional: No messages, returning")
            return {}

        # Guard: No WebSocket connection
        if not send_ws:
            return {}

        # Get the last message
        last = state["messages"][-1]
        text = last.get("content", "")

        # Guard: No text content
        if not text:
            return {}

        # Check for easter egg keywords
        keywords = ("LangChain", "langchain", "LangGraph", "langgraph")
        has_keyword = any(k in text for k in keywords)

        # Guard: No keywords found
        if not has_keyword:
            return {}

        # Send easter egg notification
        await send_ws(json.dumps({"on_easter_egg": True}))
        return {}

    # =========================================================================
    # NODE: Model (OpenAI Chat with Tool Support)
    # =========================================================================

    async def node_model(
        self,
        state: ChatState,
        send_ws: Optional[Callable[[str], Awaitable[None]]],
        stream: bool,
    ) -> Dict[str, Any]:
        """
        Call OpenAI model with tool support and handle the response.

        This node:
        1. Prepares messages for OpenAI
        2. Calls the model (streaming or non-streaming)
        3. Executes any tool calls the model requests
        4. Loops until no more tool calls are needed

        Args:
            state: Current conversation state
            send_ws: WebSocket send function (optional)
            stream: Whether to stream the response

        Returns:
            Dictionary with new messages to add to state
        """
        logger.info(
            f"node_model called - stream={stream}, state has {len(state.get('messages', []))} messages"
        )

        client = self._get_or_create_client()

        # Prepare messages
        messages = normalize_messages_to_openai_format(state.get("messages", []))
        messages = ensure_system_prompt(messages, SYSTEM_PROMPT)

        logger.info(f"Prepared {len(messages)} messages for OpenAI")

        new_messages = []

        # Tool calling loop - continue until no more tools are called
        while True:
            try:
                logger.info(f"Starting OpenAI call - stream={stream}")

                # ============================================================
                # STREAMING MODE
                # ============================================================
                if stream:
                    assistant_msg = {"role": "assistant", "content": ""}
                    finish_reason = None

                    # Create streaming request
                    stream_response = await client.chat.completions.create(
                        model=MODEL_NAME,
                        messages=messages,
                        temperature=TEMPERATURE,
                        max_tokens=MAX_TOKENS,
                        tools=TOOLS,
                        stream=True,
                    )

                    # Process stream chunks
                    async for chunk in stream_response:
                        # Guard: No choices in chunk
                        if not chunk.choices:
                            continue

                        delta = chunk.choices[0].delta
                        finish_reason = chunk.choices[0].finish_reason

                        # Handle content delta
                        if hasattr(delta, "content") and delta.content:
                            assistant_msg["content"] += delta.content

                            # Send to WebSocket if available
                            if send_ws:
                                # Frontend expects: {"on_chat_model_stream": "content here"}
                                payload = json.dumps(
                                    {"on_chat_model_stream": delta.content}
                                )
                                logger.info(f"Sending to WebSocket: {payload[:100]}...")
                                await send_ws(payload)
                            else:
                                logger.warning(
                                    "send_ws is None, cannot send stream chunk"
                                )

                        # Handle tool calls delta
                        if hasattr(delta, "tool_calls") and delta.tool_calls:
                            if "tool_calls" not in assistant_msg:
                                assistant_msg["tool_calls"] = []

                            for tc_delta in delta.tool_calls:
                                # Extend tool_calls array if needed
                                while (
                                    len(assistant_msg["tool_calls"]) <= tc_delta.index
                                ):
                                    assistant_msg["tool_calls"].append(
                                        {
                                            "id": "",
                                            "type": "function",
                                            "function": {"name": "", "arguments": ""},
                                        }
                                    )

                                tc = assistant_msg["tool_calls"][tc_delta.index]

                                if tc_delta.id:
                                    tc["id"] = tc_delta.id

                                if hasattr(tc_delta, "function"):
                                    if tc_delta.function.name:
                                        tc["function"]["name"] = tc_delta.function.name
                                    if tc_delta.function.arguments:
                                        tc["function"][
                                            "arguments"
                                        ] += tc_delta.function.arguments

                    # Notify end of stream
                    if send_ws:
                        end_payload = json.dumps({"on_chat_model_end": True})
                        logger.info(f"Sending end of stream: {end_payload}")
                        await send_ws(end_payload)
                    else:
                        logger.warning("send_ws is None, cannot send end notification")

                # ============================================================
                # NON-STREAMING MODE
                # ============================================================
                else:
                    response = await client.chat.completions.create(
                        model=MODEL_NAME,
                        messages=messages,
                        temperature=TEMPERATURE,
                        max_tokens=MAX_TOKENS,
                        tools=TOOLS,
                        stream=False,
                    )

                    assistant_msg = response.choices[0].message.model_dump()
                    finish_reason = response.choices[0].finish_reason

                # Add assistant message to history
                messages.append(assistant_msg)
                new_messages.append(assistant_msg)

                # Guard: No tool calls - we're done
                if finish_reason != "tool_calls":
                    break

                if "tool_calls" not in assistant_msg:
                    break

                # ============================================================
                # EXECUTE TOOL CALLS
                # ============================================================
                tool_responses = []

                for tool_call in assistant_msg.get("tool_calls", []):
                    # Extract tool call details
                    tool_name = (
                        tool_call.function.name
                        if hasattr(tool_call, "function")
                        else tool_call["function"]["name"]
                    )
                    arguments_str = (
                        tool_call.function.arguments
                        if hasattr(tool_call, "function")
                        else tool_call["function"]["arguments"]
                    )
                    tool_id = (
                        tool_call.id if hasattr(tool_call, "id") else tool_call["id"]
                    )

                    logger.info(f"Executing tool: {tool_name}")

                    try:
                        # Parse arguments and get tool function
                        arguments = json.loads(arguments_str)
                        tool_func = TOOL_MAP.get(tool_name)

                        # Execute tool
                        if tool_func:
                            result = tool_func(**arguments)
                            tool_response = {
                                "role": "tool",
                                "content": json.dumps(result),
                                "tool_call_id": tool_id,
                            }
                        else:
                            tool_response = {
                                "role": "tool",
                                "content": json.dumps(
                                    {"error": f"Unknown tool: {tool_name}"}
                                ),
                                "tool_call_id": tool_id,
                            }

                        tool_responses.append(tool_response)

                    except Exception as e:
                        logger.error(f"Tool execution error: {e}")
                        tool_responses.append(
                            {
                                "role": "tool",
                                "content": json.dumps({"error": str(e)}),
                                "tool_call_id": tool_id,
                            }
                        )

                # Add tool responses to conversation
                messages.extend(tool_responses)
                new_messages.extend(tool_responses)

                # Continue loop to get next assistant response

            except Exception as e:
                logger.exception(f"OpenAI call failed: {e}")
                error_msg = {
                    "role": "assistant",
                    "content": "Sorry—there was an error generating the response.",
                }
                new_messages.append(error_msg)

                if stream and send_ws:
                    await send_ws(json.dumps({"on_chat_model_end": True}))

                break

        return {"messages": new_messages}

    # =========================================================================
    # BUILD GRAPH
    # =========================================================================

    def build(self, send_ws: Optional[Callable[[str], Awaitable[None]]], stream: bool):
        """
        Build and compile the LangGraph.

        Graph structure:
            START → conditional → model → END

        Args:
            send_ws: WebSocket send function (optional)
            stream: Whether to stream responses

        Returns:
            Compiled LangGraph
        """
        graph = StateGraph(ChatState)

        # Define node wrappers (capture send_ws and stream)
        async def _conditional(s: ChatState):
            return await self.node_conditional(s, send_ws)

        async def _model(s: ChatState):
            return await self.node_model(s, send_ws=send_ws, stream=stream)

        # Add nodes
        graph.add_node("conditional", _conditional)
        graph.add_node("model", _model)

        # Add edges (define flow)
        graph.add_edge(START, "conditional")
        graph.add_edge("conditional", "model")
        graph.add_edge("model", END)

        # Compile with optional memory
        return graph.compile(checkpointer=self._memory)


# =============================================================================
# DEFAULT EXPORT (Non-streaming, no memory)
# =============================================================================


def build_default_graph():
    """Build a non-streaming graph without memory for auto-import."""
    return SimpleChatGraph(use_memory=False).build(send_ws=None, stream=False)


# Default graph instance
try:
    graph = build_default_graph()
    logger.info("Default graph built successfully")
except Exception as e:
    logger.error(f"Failed to build default graph: {e}")
    # Create a placeholder to prevent import errors
    graph = None


# =============================================================================
# WEBSOCKET ENTRYPOINT
# =============================================================================


def build_streaming_graph(send_ws: Callable[[str], Awaitable[None]]):
    """Build a streaming graph with memory for WebSocket connections."""
    logger.info("Building streaming graph...")
    try:
        graph = SimpleChatGraph(use_memory=True).build(send_ws=send_ws, stream=True)
        logger.info("Streaming graph built successfully")
        return graph
    except Exception as e:
        logger.error(f"Failed to build streaming graph: {e}", exc_info=True)
        raise


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
    2. Builds the streaming graph
    3. Invokes the graph with the user's message

    Args:
        websocket: FastAPI WebSocket connection
        data: Message data (string or dict or list)
        user_uuid: Unique user identifier for this conversation
    """
    logger.info(
        f"invoke_our_graph recv: type={type(data).__name__} "
        f"keys={list(data.keys()) if isinstance(data, dict) else 'n/a'} "
        f"data={data}"
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
        return  # Don't run the chat graph

    # =========================================================================
    # BUILD AND INVOKE GRAPH
    # =========================================================================

    async def _send_ws(payload_text: str):
        """Helper to send messages via WebSocket."""
        await websocket.send_text(payload_text)

    # Build the graph with WebSocket support
    graph_runnable = build_streaming_graph(_send_ws)
    config = {"configurable": {"thread_id": user_uuid}}

    # Handle batch mode (list of messages)
    if isinstance(data, list):
        state: ChatState = {"messages": data}
        await graph_runnable.ainvoke(state, config)
        return

    # Handle single message
    if isinstance(payload, dict) and "message" in payload:
        content = str(payload["message"])
    else:
        content = str(data)

    logger.info(f"About to invoke graph with content: {content}")
    logger.info(f"Thread ID: {user_uuid}")

    state: ChatState = {"messages": [{"role": "user", "content": content}]}

    try:
        logger.info("Calling graph_runnable.ainvoke...")
        result = await graph_runnable.ainvoke(state, config)
        logger.info(f"Graph invocation complete. Result: {result}")
    except Exception as e:
        logger.error(f"Error invoking graph: {e}", exc_info=True)

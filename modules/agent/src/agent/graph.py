# graph.py — LangGraph + OpenAI with integrated resume tooling
# - Streaming graph for WebSocket with memory
# - Tool calling integrated into the graph flow
# - Per-conversation memory via MemorySaver()

import os, sys, json, logging, uuid, requests
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

# -----------------------------------------------------------------------------
# Setup
# -----------------------------------------------------------------------------
logger = logging.getLogger("app.graph")
load_dotenv(override=True)

API_KEY = os.getenv("OPENAI_API_KEY")
if not API_KEY:
    logger.fatal("Fatal Error: OPENAI_API_KEY is missing.")
    sys.exit(1)

MODEL_NAME = os.getenv("OPENAI_MODEL", "gpt-5-chat-latest")
TEMPERATURE = float(os.getenv("OPENAI_TEMPERATURE", "0"))
MAX_TOKENS_ENV = os.getenv("OPENAI_MAX_TOKENS")
MAX_TOKENS = None if not MAX_TOKENS_ENV else int(MAX_TOKENS_ENV)

# Resume tooling setup
RESUME_NAME = "William Hubenschmidt"

from agent.document_scanner import load_documents

resume_text, summary, cover_letters = load_documents("me")


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
Your responses should always be as concise as possible. Do not ask multiple questions in the same response. Always answer the question.
If you don't know the answer to any question, use your record_unknown_question tool to record the question that you couldn't answer,
even if it's about something trivial or unrelated to career. 

CONTACT CAPTURE (ASK ONCE, RECORD ONCE)
- Early in the conversation (after your third substantive answer), briefly ask for the user’s name and email so you can follow up.
- If the user provides an email (with or without a name), immediately call record_user_details with the fields provided. Email is required; name is optional.
- If the user provides a name but no email, ask once for the email. Do not keep asking in later turns.
- If the user declines to share contact info, acknowledge and continue. Never block answering.
- After record_user_details has run successfully, do not ask for contact info again.

If the user asks questions that do not directly pertain to {RESUME_NAME}'s career, background, skills, and experience,
do not answer them; briefly steer the conversation back to those topics.

If the user prompts you to write a cover letter for a job on your behalf, please ask for a URL link to the job posting. Read the job posting, \
and then use the examples in {cover_letters}.  Do not ask the user to confirm, just read the job posting and write the cover letter.

SUMMARY:
{summary}

RESUME:
{resume_text}

COVER_LETTERS:
{cover_letters}

With this context, chat with the user, always staying in character as {RESUME_NAME}.


"""


logger.info(f"System prompt length: {len(SYSTEM_PROMPT)} characters")
logger.info(f"Resume text length: {len(resume_text)} characters")
logger.info(f"Summary length: {len(summary)} characters")

# Pushover configuration
pushover_user = os.getenv("PUSHOVER_USER")
pushover_token = os.getenv("PUSHOVER_TOKEN")
pushover_url = "https://api.pushover.net/1/messages.json"


# -----------------------------------------------------------------------------
# Tool Functions
# -----------------------------------------------------------------------------
def push(message):
    """Send notification via Pushover"""
    print(f"Push: {message}")
    if not pushover_user or not pushover_token:
        logger.warning("Pushover credentials not configured")
        return
    try:
        payload = {"user": pushover_user, "token": pushover_token, "message": message}
        response = requests.post(pushover_url, data=payload, timeout=5)
        logger.info(f"Push notification sent: {response.status_code}")
    except Exception as e:
        logger.error(f"Failed to send push notification: {e}")


def record_user_details(
    email: str, name: str = "Name not provided", notes: str = "not provided"
):
    """Record user contact details"""
    push(f"Recording interest from {name} with email {email} and notes {notes}")
    return {"recorded": "ok", "email": email, "name": name}


def record_unknown_question(question: str):
    """Record questions that couldn't be answered"""
    push(f"Recording question that couldn't be answered: {question}")
    return {"recorded": "ok", "question": question}


# -----------------------------------------------------------------------------
# Tool-Free server-side recorder for client metadata
# -----------------------------------------------------------------------------

# keep the hardcoded string, or allow env override:
IP_BLACKLIST_STR = os.getenv("IP_BLACKLIST", "127.0.0.1, 10.0.0.0/8, 172.18.0.1")


def _parse_ip_blacklist(raw: str):
    items = []
    for token in (t.strip() for t in raw.split(",") if t.strip()):
        # try CIDR first, then single IP
        try:
            items.append(ip_network(token, strict=False))
            continue
        except Exception:
            pass
        try:
            items.append(ip_address(token))
        except Exception:
            pass
    return tuple(items)


IP_BLACKLIST = _parse_ip_blacklist(IP_BLACKLIST_STR)


def _is_ip_blacklisted(ip_str: str | None) -> bool:
    if not ip_str:
        return False
    try:
        ip = ip_address(ip_str)
    except Exception:
        return False
    for item in IP_BLACKLIST:
        if isinstance(item, (IPv4Address, IPv6Address)):
            if ip == item:
                return True
        elif isinstance(item, (IPv4Network, IPv6Network)):
            if ip in item:
                return True
    return False


# --- Pretty push with blacklist check ---
def record_user_context(ctx: Dict[str, Any]):
    """Persist raw user context (testing: push as a notification)"""
    try:
        ip = ((ctx.get("server") or {}).get("ip")) or None

        # Always log full JSON for debugging
        pretty_full = json.dumps(ctx, indent=2, ensure_ascii=False, sort_keys=True)
        logger.info("[user_context_full] %s", pretty_full)

        # Skip push for blacklisted IPs
        if _is_ip_blacklisted(ip):
            logger.info("user_context suppressed (blacklisted ip=%s)", ip)
            return {"ok": True, "pushed": False, "reason": "ip_blacklisted"}

        # Pushover practical limit ~1024 chars
        max_len = 1024
        body = (
            pretty_full
            if len(pretty_full) <= max_len
            else (pretty_full[: max_len - 1] + "…")
        )

        push(body)
        return {"ok": True, "pushed": True}
    except Exception as e:
        logger.error(f"record_user_context failed: {e}")
        return {"ok": False, "error": str(e)}


def extract_client_ip(headers: Dict[str, str], peer_host: str) -> str:
    # Trust right-most hop from X-Forwarded-For if your proxy/CDN sets it
    xff = headers.get("x-forwarded-for")
    if xff:
        parts = [p.strip() for p in xff.split(",") if p.strip()]
        # last element is closest to server; adjust if you trust multiple hops
        candidate = parts[-1]
        try:
            ip_address(candidate)
            return candidate
        except Exception:
            pass
    return peer_host


# Tool definitions for OpenAI
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

TOOL_MAP = {
    "record_user_details": record_user_details,
    "record_unknown_question": record_unknown_question,
}


# -----------------------------------------------------------------------------
# State type
# -----------------------------------------------------------------------------
class ChatState(TypedDict):
    messages: List[Dict[str, str]]


# -----------------------------------------------------------------------------
# Helpers
# -----------------------------------------------------------------------------
def to_openai_messages(messages: List[Any]) -> List[Dict[str, Any]]:
    """Normalize messages to OpenAI format, preserving tool calls and tool responses"""
    out: List[Dict[str, Any]] = []
    for m in messages:
        # If it's already a proper dict, use it as-is (preserves tool_calls, tool_call_id, etc.)
        if isinstance(m, dict) and "role" in m:
            out.append(m)
            continue

        # Otherwise normalize
        role = getattr(m, "role", None) or "user"
        content = getattr(m, "content", None) or str(m)
        out.append({"role": role, "content": content})

    return out


def ensure_system_prompt(
    msgs: List[Dict[str, Any]], system_text: str
) -> List[Dict[str, Any]]:
    """Ensure system message is always first and up-to-date"""
    # Remove any existing system messages
    non_system_msgs = [m for m in msgs if m.get("role") != "system"]
    # Always prepend fresh system prompt
    return [{"role": "system", "content": system_text}] + non_system_msgs


# -----------------------------------------------------------------------------
# Graph Builder Class
# -----------------------------------------------------------------------------
class SimpleChatGraph:
    def __init__(self, *, use_memory: bool):
        self.id = str(uuid.uuid4())
        self._client: Optional[AsyncOpenAI] = None
        self._memory = MemorySaver() if use_memory else None

    def _client_or_create(self) -> AsyncOpenAI:
        if self._client is None:
            self._client = AsyncOpenAI()
        return self._client

    # Node 1: Optional easter egg
    async def node_conditional(
        self, state: ChatState, send_ws: Optional[Callable[[str], Awaitable[None]]]
    ):
        if not state.get("messages"):
            return

        last = state["messages"][-1]
        text = last.get("content", "")
        if not text:
            return

        if send_ws and any(
            k in text for k in ("LangChain", "langchain", "LangGraph", "langgraph")
        ):
            await send_ws(json.dumps({"on_easter_egg": True}))

    # Node 2: Call OpenAI with tool support
    async def node_model(
        self,
        state: ChatState,
        *,
        send_ws: Optional[Callable[[str], Awaitable[None]]],
        stream: bool,
    ) -> Dict[str, Any]:
        client = self._client_or_create()

        # Normalize and ensure system prompt is always fresh
        msgs = to_openai_messages(state["messages"])
        msgs = ensure_system_prompt(msgs, SYSTEM_PROMPT)

        new_messages = []
        max_iterations = 5  # Prevent infinite tool calling loops

        for iteration in range(max_iterations):
            try:
                # Call OpenAI with tools
                if stream and send_ws:
                    # Streaming mode
                    resp_stream = await client.chat.completions.create(
                        model=MODEL_NAME,
                        messages=msgs,
                        temperature=TEMPERATURE,
                        max_tokens=MAX_TOKENS,
                        tools=TOOLS,
                        stream=True,
                    )

                    pieces: List[str] = []
                    tool_calls = []
                    current_tool_call = None

                    async for chunk in resp_stream:
                        delta = chunk.choices[0].delta

                        # Handle content streaming
                        if delta.content:
                            pieces.append(delta.content)
                            await send_ws(
                                json.dumps({"on_chat_model_stream": delta.content})
                            )

                        # Handle tool calls (accumulated across chunks)
                        if delta.tool_calls:
                            for tc_delta in delta.tool_calls:
                                if tc_delta.index is not None:
                                    # Start new tool call or update existing
                                    while len(tool_calls) <= tc_delta.index:
                                        tool_calls.append(
                                            {
                                                "id": None,
                                                "type": "function",
                                                "function": {
                                                    "name": "",
                                                    "arguments": "",
                                                },
                                            }
                                        )

                                    if tc_delta.id:
                                        tool_calls[tc_delta.index]["id"] = tc_delta.id
                                    if tc_delta.function:
                                        if tc_delta.function.name:
                                            tool_calls[tc_delta.index]["function"][
                                                "name"
                                            ] = tc_delta.function.name
                                        if tc_delta.function.arguments:
                                            tool_calls[tc_delta.index]["function"][
                                                "arguments"
                                            ] += tc_delta.function.arguments

                    finish_reason = (
                        chunk.choices[0].finish_reason if chunk.choices else None
                    )

                    # Build assistant message
                    assistant_msg = {"role": "assistant"}
                    if pieces:
                        assistant_msg["content"] = "".join(pieces)
                        await send_ws(json.dumps({"on_chat_model_end": True}))
                    else:
                        assistant_msg["content"] = ""

                    if tool_calls and any(tc["id"] for tc in tool_calls):
                        # Convert to proper format
                        from openai.types.chat import ChatCompletionMessageToolCall
                        from openai.types.chat.chat_completion_message_tool_call import (
                            Function,
                        )

                        formatted_tool_calls = []
                        for tc in tool_calls:
                            if tc["id"]:
                                formatted_tool_calls.append(
                                    ChatCompletionMessageToolCall(
                                        id=tc["id"],
                                        type="function",
                                        function=Function(
                                            name=tc["function"]["name"],
                                            arguments=tc["function"]["arguments"],
                                        ),
                                    )
                                )
                        assistant_msg["tool_calls"] = formatted_tool_calls

                else:
                    # Non-streaming mode
                    resp = await client.chat.completions.create(
                        model=MODEL_NAME,
                        messages=msgs,
                        temperature=TEMPERATURE,
                        max_tokens=MAX_TOKENS,
                        tools=TOOLS,
                        stream=False,
                    )

                    assistant_msg = resp.choices[0].message.model_dump()
                    finish_reason = resp.choices[0].finish_reason

                # Add assistant message to history
                msgs.append(assistant_msg)
                new_messages.append(assistant_msg)

                # If no tool calls, we're done
                if finish_reason != "tool_calls" or "tool_calls" not in assistant_msg:
                    break

                # Execute tool calls
                tool_responses = []
                for tool_call in assistant_msg.get("tool_calls", []):
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
                        arguments = json.loads(arguments_str)
                        tool_func = TOOL_MAP.get(tool_name)

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
                msgs.extend(tool_responses)
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

    # Build the graph
    def build(
        self, *, send_ws: Optional[Callable[[str], Awaitable[None]]], stream: bool
    ):
        """Build and compile: START -> conditional -> model -> END"""
        g = StateGraph(ChatState)

        async def _conditional(s: ChatState):
            return await self.node_conditional(s, send_ws)

        async def _model(s: ChatState):
            return await self.node_model(s, send_ws=send_ws, stream=stream)

        g.add_node("conditional", _conditional)
        g.add_node("model", _model)

        g.add_edge(START, "conditional")
        g.add_edge("conditional", "model")
        g.add_edge("model", END)

        return g.compile(checkpointer=self._memory)


# =============================================================================
# DEFAULT EXPORT for loaders (non-streaming, NO memory)
# =============================================================================
def build_default_graph():
    """Non-streaming graph for auto-import"""
    return SimpleChatGraph(use_memory=False).build(send_ws=None, stream=False)


graph = build_default_graph()

# =============================================================================
# WEBSOCKET ENTRYPOINT
# =============================================================================
from fastapi import WebSocket
from langfuse import observe


def build_streaming_graph(send_ws: Callable[[str], Awaitable[None]]):
    """Build streaming graph with MemorySaver"""
    return SimpleChatGraph(use_memory=True).build(send_ws=send_ws, stream=True)


@observe()
async def invoke_our_graph(
    websocket: WebSocket,
    data: Union[str, Dict[str, Any], List[Dict[str, str]]],
    user_uuid: str,
):
    """WebSocket entrypoint - handles control envelopes silently, runs chat graph otherwise."""
    logger.info(
        f"invoke_our_graph recv: type={type(data).__name__} keys={list(data.keys()) if isinstance(data, dict) else 'n/a'}"
    )
    # ---------- 0) Normalize the incoming frame ----------
    payload: Optional[Dict[str, Any]] = None
    if isinstance(data, str):
        try:
            payload = json.loads(data)
        except Exception:
            payload = None
    elif isinstance(data, dict):
        payload = data  # server already parsed JSON

    # ---------- 1) Control envelopes (silent) ----------
    if isinstance(payload, dict):
        # Ignore handshake frames like {"uuid": "...", "init": true}
        if payload.get("init") is True:
            return

        # Telemetry/control frames
        if payload.get("type") in ("user_context", "user_context_update"):
            ctx = payload.get("ctx") or {}

            # Attach server-known info
            hdrs = {k.decode().lower(): v.decode() for k, v in websocket.headers.raw}
            client_ip = extract_client_ip(
                hdrs, websocket.client.host if websocket.client else ""
            )
            ctx.setdefault("server", {})
            ctx["server"].update(
                {
                    "ip": client_ip,
                    "user_uuid": user_uuid,
                    "ua": hdrs.get("user-agent"),
                    "origin": hdrs.get("origin"),
                }
            )

            record_user_context(ctx)
            return  # ✅ do not run the chat graph or echo anything

    # ---------- 2) Build graph only for chat turns ----------
    async def _send_ws(payload_text: str):
        await websocket.send_text(payload_text)

    graph_runnable = build_streaming_graph(_send_ws)
    config = {"configurable": {"thread_id": user_uuid}}

    # Batch path: list of {role, content}
    if isinstance(data, list):
        state: ChatState = {"messages": data}
        await graph_runnable.ainvoke(state, config)
        return

    # Single-turn chat path: prefer payload["message"] if present
    if isinstance(payload, dict) and "message" in payload:
        content = str(payload["message"])
    else:
        content = str(data)

    state: ChatState = {"messages": [{"role": "user", "content": content}]}
    await graph_runnable.ainvoke(state, config)

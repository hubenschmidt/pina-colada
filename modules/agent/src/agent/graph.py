# graph.py — LangGraph + OpenAI with integrated resume tooling
# - Streaming graph for WebSocket with memory
# - Tool calling integrated into the graph flow
# - Per-conversation memory via MemorySaver()

import os, sys, json, logging, uuid, requests
from typing import TypedDict, List, Dict, Any, Optional, Callable, Awaitable, Union
from dotenv import load_dotenv
from openai import AsyncOpenAI
from pypdf import PdfReader
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

MODEL_NAME = os.getenv("OPENAI_MODEL", "gpt-4o")
TEMPERATURE = float(os.getenv("OPENAI_TEMPERATURE", "0"))
MAX_TOKENS_ENV = os.getenv("OPENAI_MAX_TOKENS")
MAX_TOKENS = None if not MAX_TOKENS_ENV else int(MAX_TOKENS_ENV)

# Resume tooling setup
RESUME_NAME = "William Hubenschmidt"

# Load resume and summary
resume_text = ""
summary = ""

try:
    reader = PdfReader("me/resume.pdf")
    for page in reader.pages:
        text = page.extract_text()
        if text:
            resume_text += text
    
    # Clean up excessive whitespace from PDF extraction
    import re
    resume_text = re.sub(r'\n\s*\n', '\n\n', resume_text)  # Multiple newlines to double
    resume_text = re.sub(r' +', ' ', resume_text)  # Multiple spaces to single
    resume_text = resume_text.strip()
    
    logger.info(f"Resume loaded successfully: {len(resume_text)} characters")
except FileNotFoundError:
    logger.warning("Resume PDF not found at me/resume.pdf")
    resume_text = "[Resume not available]"
except Exception as e:
    logger.error(f"Could not load resume PDF: {e}")
    resume_text = "[Resume not available]"

try:
    with open("me/summary.txt", "r", encoding="utf-8") as f:
        summary = f.read()
    logger.info(f"Summary loaded successfully: {len(summary)} characters")
except FileNotFoundError:
    logger.warning("Summary file not found at me/summary.txt")
    summary = "[Summary not available]"
except Exception as e:
    logger.error(f"Could not load summary: {e}")
    summary = "[Summary not available]"

# Build system prompt
SYSTEM_PROMPT = f"""You are acting as {RESUME_NAME}. You are answering questions on {RESUME_NAME}'s website, \
particularly questions related to {RESUME_NAME}'s career, background, skills and experience. \
Your responsibility is to represent {RESUME_NAME} for interactions on the website as faithfully as possible. \
You are given a summary of {RESUME_NAME}'s background and a copy of his resume which you can use to answer questions. \
Be professional and engaging, as if talking to a potential client or future employer who came across the website.

IMPORTANT: You have direct access to {RESUME_NAME}'s complete resume and summary below. Use this information to answer all questions about his career, work history, skills, and experience. Do not claim you don't have access to this information.

If you don't know the answer to any question, use your record_unknown_question tool to record the question that you couldn't answer, \
even if it's about something trivial or unrelated to career.

If the user is engaging in discussion, try to steer them towards getting in touch via email; \
ask for their email and record it using your record_user_details tool.

## Summary:
{summary}

## Resume:
{resume_text}

With this context, please chat with the user, always staying in character as {RESUME_NAME}."""

logger.info(f"System prompt length: {len(SYSTEM_PROMPT)} characters")
logger.info(f"Resume text length: {len(resume_text)} characters")
logger.info(f"Summary length: {len(summary)} characters")

# Pushover configuration
pushover_user = os.getenv("PUSHOVER_USER")
pushover_token = os.getenv("PUSHOVER_TOKEN")
pushover_url = 'https://api.pushover.net/1/messages.json'

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

def record_user_details(email: str, name: str = "Name not provided", notes: str = "not provided"):
    """Record user contact details"""
    push(f"Recording interest from {name} with email {email} and notes {notes}")
    return {"recorded": "ok", "email": email, "name": name}

def record_unknown_question(question: str):
    """Record questions that couldn't be answered"""
    push(f"Recording question that couldn't be answered: {question}")
    return {"recorded": "ok", "question": question}

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
                        "description": "The email address of this user"
                    },
                    "name": {
                        "type": "string",
                        "description": "The user's name, if they provided it"
                    },
                    "notes": {
                        "type": "string",
                        "description": "Any additional information about the conversation that's worth recording to give context"
                    }
                },
                "required": ["email"],
                "additionalProperties": False
            }
        }
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
                        "description": "The question that couldn't be answered"
                    }
                },
                "required": ["question"],
                "additionalProperties": False
            }
        }
    }
]

TOOL_MAP = {
    "record_user_details": record_user_details,
    "record_unknown_question": record_unknown_question
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

def ensure_system_prompt(msgs: List[Dict[str, Any]], system_text: str) -> List[Dict[str, Any]]:
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
    async def node_conditional(self, state: ChatState, send_ws: Optional[Callable[[str], Awaitable[None]]]):
        if not state.get("messages"):
            return
        
        last = state["messages"][-1]
        text = last.get("content", "")
        if not text:
            return
        
        if send_ws and any(k in text for k in ("LangChain", "langchain", "LangGraph", "langgraph")):
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
        
        # Debug: log to verify system prompt and total message count
        logger.info(f"Processing {len(msgs)} messages, system prompt: {len(msgs[0].get('content', ''))} chars")
        if msgs and msgs[0].get("role") == "system":
            # Log a snippet to verify resume content is included
            system_content = msgs[0].get('content', '')

            # Check for Resume
            if "## Resume:" in system_content:
                resume_start = system_content.find("## Resume:")
                logger.info(f"✓ Resume section found at position {resume_start}")
                # Log first 200 chars of resume to verify
                resume_snippet = system_content[resume_start:resume_start+200]
                logger.info(f"Resume snippet: {resume_snippet}...")
            else:
                logger.error("✗ Resume section NOT found in system prompt!")
            
            # Check for Summary
            if "## Summary:" in system_content:
                summary_start = system_content.find("## Summary:")
                logger.info(f"✓ Summary section found at position {summary_start}")
                # Log first 200 chars of summary to verify
                summary_snippet = system_content[summary_start:summary_start+200]
                logger.info(f"Summary snippet: {summary_snippet}...")
            else:
                logger.error("✗ Summary section NOT found in system prompt!")
        else:
            logger.error("✗ System prompt not found in messages!")
        
        # Log all message roles to debug conversation flow
        logger.info(f"Message roles: {[m.get('role') for m in msgs]}")
        
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
                            await send_ws(json.dumps({"on_chat_model_stream": delta.content}))
                        
                        # Handle tool calls (accumulated across chunks)
                        if delta.tool_calls:
                            for tc_delta in delta.tool_calls:
                                if tc_delta.index is not None:
                                    # Start new tool call or update existing
                                    while len(tool_calls) <= tc_delta.index:
                                        tool_calls.append({
                                            "id": None,
                                            "type": "function",
                                            "function": {"name": "", "arguments": ""}
                                        })
                                    
                                    if tc_delta.id:
                                        tool_calls[tc_delta.index]["id"] = tc_delta.id
                                    if tc_delta.function:
                                        if tc_delta.function.name:
                                            tool_calls[tc_delta.index]["function"]["name"] = tc_delta.function.name
                                        if tc_delta.function.arguments:
                                            tool_calls[tc_delta.index]["function"]["arguments"] += tc_delta.function.arguments
                    
                    finish_reason = chunk.choices[0].finish_reason if chunk.choices else None
                    
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
                        from openai.types.chat.chat_completion_message_tool_call import Function
                        
                        formatted_tool_calls = []
                        for tc in tool_calls:
                            if tc["id"]:
                                formatted_tool_calls.append(
                                    ChatCompletionMessageToolCall(
                                        id=tc["id"],
                                        type="function",
                                        function=Function(
                                            name=tc["function"]["name"],
                                            arguments=tc["function"]["arguments"]
                                        )
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
                    tool_name = tool_call.function.name if hasattr(tool_call, 'function') else tool_call["function"]["name"]
                    arguments_str = tool_call.function.arguments if hasattr(tool_call, 'function') else tool_call["function"]["arguments"]
                    tool_id = tool_call.id if hasattr(tool_call, 'id') else tool_call["id"]
                    
                    logger.info(f"Executing tool: {tool_name}")
                    
                    try:
                        arguments = json.loads(arguments_str)
                        tool_func = TOOL_MAP.get(tool_name)
                        
                        if tool_func:
                            result = tool_func(**arguments)
                            tool_response = {
                                "role": "tool",
                                "content": json.dumps(result),
                                "tool_call_id": tool_id
                            }
                        else:
                            tool_response = {
                                "role": "tool",
                                "content": json.dumps({"error": f"Unknown tool: {tool_name}"}),
                                "tool_call_id": tool_id
                            }
                        
                        tool_responses.append(tool_response)
                        
                    except Exception as e:
                        logger.error(f"Tool execution error: {e}")
                        tool_responses.append({
                            "role": "tool",
                            "content": json.dumps({"error": str(e)}),
                            "tool_call_id": tool_id
                        })
                
                # Add tool responses to conversation
                msgs.extend(tool_responses)
                new_messages.extend(tool_responses)
                
                # Continue loop to get next assistant response
                
            except Exception as e:
                logger.exception(f"OpenAI call failed: {e}")
                error_msg = {"role": "assistant", "content": "Sorry—there was an error generating the response."}
                new_messages.append(error_msg)
                if stream and send_ws:
                    await send_ws(json.dumps({"on_chat_model_end": True}))
                break
        
        return {"messages": new_messages}

    # Build the graph
    def build(self, *, send_ws: Optional[Callable[[str], Awaitable[None]]], stream: bool):
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
async def invoke_our_graph(websocket: WebSocket, data: Union[str, List[Dict[str, str]]], user_uuid: str):
    """WebSocket entrypoint - builds streaming graph and invokes it"""
    async def _send_ws(payload: str):
        await websocket.send_text(payload)
    
    graph_runnable = build_streaming_graph(_send_ws)
    config = {"configurable": {"thread_id": user_uuid}}
    
    if isinstance(data, list):
        state: ChatState = {"messages": data}
        await graph_runnable.ainvoke(state, config)
        return
    
    # Fallback: treat as single user message
    state: ChatState = {"messages": [{"role": "user", "content": str(data)}]}
    await graph_runnable.ainvoke(state, config)
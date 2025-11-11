"""
Worker node - functional implementation with closure
"""

import logging
import os
from typing import Dict, Any, Callable
from datetime import datetime
from langchain_core.messages import SystemMessage, HumanMessage
from langchain_openai import ChatOpenAI

logger = logging.getLogger(__name__)

# Check if Langfuse should be enabled (only in development)
def _is_langfuse_enabled() -> bool:
    """Check if Langfuse should be enabled - only in development"""
    node_env = os.getenv("NODE_ENV", "").lower()
    environment = os.getenv("ENVIRONMENT", "").lower()
    langfuse_host = os.getenv("LANGFUSE_HOST", "")
    
    # Disable in production
    if node_env == "production" or environment == "production":
        return False
    
    # Only enable if LANGFUSE_HOST is set and points to local development
    if langfuse_host and ("langfuse:" in langfuse_host or "localhost" in langfuse_host or "127.0.0.1" in langfuse_host):
        return True
    
    return False

# Conditionally import and use Langfuse
if _is_langfuse_enabled():
    try:
        from langfuse.langchain import CallbackHandler
        _get_langfuse_handler = lambda: CallbackHandler()
    except ImportError:
        logger.warning("Langfuse import failed, disabling")
        _get_langfuse_handler = lambda: None
else:
    _get_langfuse_handler = lambda: None


def _should_use_full_context(message: str) -> bool:
    """Pure function to determine if full context is needed"""
    detailed_keywords = [
        "cover letter",
        "write a letter",
        "detailed",
        "comprehensive",
        "all experience",
        "job search",
        "apply",
    ]
    message_lower = message.lower()
    return any(keyword in message_lower for keyword in detailed_keywords)


def _build_system_prompt(
    state: Dict[str, Any], use_full_context: bool, resume_context_concise: str
) -> str:
    """Pure function to build system prompt"""
    context = state["resume_context"] if use_full_context else resume_context_concise

    return f"""ROLE: You are {state['resume_name']} on his website. Answer questions professionally and concisely.

DATA ACCESS:
{context}

TASK: {state['success_criteria']}

STYLE:
- Plain text only (no markdown/formatting)
- Concise responses
- No repeated greetings

INSTRUCTIONS:
- Answer using resume data above
- Use record_user_details for contact info
- Use record_unknown_question if you can't answer
- For job searches, use the job_search tool (NOT web_search) to automatically filter out already-applied positions. Search in NYC for jobs posted in the last 7 days, and always return direct posting URLs
- For questions about job applications (e.g., "how many jobs have I applied to?", "did I apply to X?", "what jobs have I applied to?"), ALWAYS use the check_applied_jobs tool to get accurate information from Supabase. When listing applied jobs, include: company name, job title, and direct link to the job posting. Do NOT include job board names or sources - only include direct links to the actual job postings.
- Ask for email after 3rd answer (once only)

Date: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
"""


def _get_last_user_message(messages) -> str:
    """Pure function to extract last user message"""
    for msg in reversed(messages):
        if isinstance(msg, HumanMessage):
            return msg.content
    return ""


async def create_worker_node(
    tools: list,
    resume_context_concise: str,  # takes a shorter context as a param instead of using the full resume_context in state
    trim_messages_fn: Callable,
):
    """
    Factory function that creates a worker node with closed-over state

    Returns a pure function that takes state and returns updated state
    """
    logger.info("Setting up Worker LLM: OpenAI GPT-5 (temperature=0.7)")
    langfuse_handler = _get_langfuse_handler()
    
    callbacks = [langfuse_handler] if langfuse_handler else []

    worker_llm = ChatOpenAI(
        model="gpt-5-chat-latest",
        temperature=0.7,
        max_completion_tokens=512,
        max_retries=3,
        callbacks=callbacks,
    )

    llm_with_tools = worker_llm.bind_tools(tools)
    logger.info(f"✓ Worker LLM configured with {len(tools)} tools")

    # Return the actual node function with closed-over state
    def worker_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Pure function: state in -> state updates out"""
        logger.info("🤖 WORKER NODE: Generating response with GPT-5...")

        # Determine if we need full context
        last_user_message = _get_last_user_message(state["messages"])
        use_full = _should_use_full_context(last_user_message)
        logger.info(f"Using {'FULL' if use_full else 'CONCISE'} context")

        # Build system prompt
        system_prompt = _build_system_prompt(state, use_full, resume_context_concise)

        if state.get("feedback_on_work"):
            logger.info("⚠️  Worker received feedback from evaluator, retrying...")
            system_prompt += f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\nPlease improve your response based on this feedback."

        # Trim message history
        trimmed_messages = trim_messages_fn(state["messages"], max_tokens=6000)
        messages = [SystemMessage(content=system_prompt)] + trimmed_messages

        logger.info(
            f"   Message count: {len(messages)} (trimmed from {len(state['messages'])})"
        )
        logger.info(f"   System prompt length: ~{len(system_prompt)} chars")

        # Get response
        response = llm_with_tools.invoke(messages)

        logger.info("✓ Worker response generated")
        logger.info(f"   Response length: {len(response.content)} chars")
        logger.info(f"   Has tool calls: {bool(response.tool_calls)}")

        return {"messages": [response]}

    return worker_node


def route_from_worker(state: Dict[str, Any]) -> str:
    """Pure routing function - examines state and returns next node"""
    last_message = state["messages"][-1]

    if hasattr(last_message, "tool_calls") and last_message.tool_calls:
        logger.info("→ Routing to TOOLS")
        return "tools"

    logger.info("→ Routing to EVALUATOR")
    return "evaluator"

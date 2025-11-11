"""
Job Hunter node - specialized worker for job search tasks.
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


def _build_job_hunter_prompt(
    state: Dict[str, Any], resume_context_concise: str
) -> str:
    """Pure function to build job hunter system prompt."""
    return f"""ROLE: You are {state['resume_name']} on his website, specialized in finding and presenting job opportunities.

DATA ACCESS:
{resume_context_concise}

TASK: {state['success_criteria']}

STYLE:
- Plain text only (no markdown/formatting)
- Concise, focused responses
- Always include direct links to job postings
- List jobs clearly with: Company - Job Title - Direct Link

INSTRUCTIONS:
- Use the job_search tool to find jobs matching the user's criteria
- Search in NYC for jobs posted in the last 7 days
- Search first for startups (series A, B, or C, etc) and then for larger companies
- ALWAYS return direct posting URLs - never job board links or sources
- Filter out jobs already applied to using check_applied_jobs tool
- Present results as: Company Name - Job Title - Direct Link
- Include only relevant jobs that match the user's skills and interests
- When listing jobs, provide company name, job title, and direct link only
- Do NOT include job board names, sources, or intermediate links
- Do NOT use any special characters in the job title or company name

Date: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
"""


def _get_last_user_message(messages) -> str:
    """Pure function to extract last user message."""
    for msg in reversed(messages):
        if isinstance(msg, HumanMessage):
            return msg.content
    return ""


async def create_job_hunter_node(
    tools: list,
    resume_context_concise: str,
    trim_messages_fn: Callable,
):
    """
    Factory function that creates a job hunter node with closed-over state.

    Returns a pure function that takes state and returns updated state.
    """
    logger.info("Setting up Job Hunter LLM: OpenAI GPT-5 (temperature=0.7)")
    langfuse_handler = _get_langfuse_handler()
    
    callbacks = [langfuse_handler] if langfuse_handler else []

    job_hunter_llm = ChatOpenAI(
        model="gpt-5-chat-latest",
        temperature=0.7,
        max_completion_tokens=1024,
        max_retries=3,
        callbacks=callbacks,
    )

    llm_with_tools = job_hunter_llm.bind_tools(tools)
    logger.info(f"✓ Job Hunter LLM configured with {len(tools)} tools")

    def job_hunter_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Pure function: state in -> state updates out"""
        logger.info("🔍 JOB HUNTER NODE: Searching for jobs...")

        system_prompt = _build_job_hunter_prompt(state, resume_context_concise)

        if state.get("feedback_on_work"):
            logger.info("⚠️  Job Hunter received feedback from evaluator, retrying...")
            system_prompt += f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\nPlease improve your response based on this feedback."

        trimmed_messages = trim_messages_fn(state["messages"], max_tokens=6000)
        messages = [SystemMessage(content=system_prompt)] + trimmed_messages

        logger.info(
            f"   Message count: {len(messages)} (trimmed from {len(state['messages'])})"
        )
        logger.info(f"   System prompt length: ~{len(system_prompt)} chars")

        response = llm_with_tools.invoke(messages)

        logger.info("✓ Job Hunter response generated")
        logger.info(f"   Response length: {len(response.content)} chars")
        logger.info(f"   Has tool calls: {bool(response.tool_calls)}")

        return {"messages": [response]}

    return job_hunter_node


def route_from_job_hunter(state: Dict[str, Any]) -> str:
    """Pure routing function - examines state and returns next node"""
    last_message = state["messages"][-1]

    if hasattr(last_message, "tool_calls") and last_message.tool_calls:
        logger.info("→ Routing to TOOLS")
        return "tools"

    logger.info("→ Routing to EVALUATOR")
    return "evaluator"


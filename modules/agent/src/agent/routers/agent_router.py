"""
LLM-based router - uses Claude Haiku to decide which worker should handle the request
"""

import logging
from typing import Dict, Any, Literal

from langchain_core.messages import HumanMessage, AIMessage, SystemMessage
from langchain_anthropic import ChatAnthropic
from pydantic import BaseModel, Field

from agent.util.langfuse_helper import get_langfuse_handler

logger = logging.getLogger(__name__)


class RouterDecision(BaseModel):
    """Structured output for routing decision"""
    route: Literal["worker", "job_search", "cover_letter_writer"] = Field(
        description="The worker to route to based on user intent"
    )
    reasoning: str = Field(
        description="Brief explanation of why this route was chosen"
    )


ROUTER_SYSTEM_PROMPT = """You are a routing assistant that decides which specialized worker should handle a user request.

AVAILABLE WORKERS:
1. worker - General-purpose assistant for questions, conversation, and non-specialized tasks
2. job_search - Specialized for finding job opportunities, searching job boards, filtering jobs
3. cover_letter_writer - Specialized for writing cover letters tailored to specific job postings

ROUTING RULES:
- Route to "job_search" when the user wants to FIND or SEARCH for jobs, see job listings, or filter job results
- Route to "cover_letter_writer" when the user wants to WRITE or CREATE a cover letter for a specific role
- Route to "worker" for ALL other requests including:
  - General questions (technical, knowledge, etc.)
  - Conversation and chitchat
  - Resume questions
  - Career advice (that isn't job search or cover letter writing)
  - Any ambiguous requests

IMPORTANT:
- Focus on the CURRENT message's intent, not just keywords from context
- If the current message is a general question, route to "worker" even if previous messages were about jobs/cover letters
- When in doubt, route to "worker"

Respond with your routing decision."""


def _get_last_user_message(messages) -> str:
    """Extract the last user message from message history"""
    for msg in reversed(messages):
        if isinstance(msg, HumanMessage):
            content = msg.content if isinstance(msg.content, str) else ""
            return content
    return ""


def _get_recent_context(messages, limit: int = 4) -> str:
    """Get recent conversation context for routing decision"""
    recent = []
    count = 0

    def should_process():
        return count < limit

    def process_human_message(msg):
        nonlocal count
        if not should_process():
            return
        content = msg.content if isinstance(msg.content, str) else ""
        recent.append(f"User: {content}")
        count += 1

    def process_ai_message(msg):
        nonlocal count
        if not should_process():
            return
        content = msg.content if isinstance(msg.content, str) else ""
        # Truncate long AI responses
        truncated = content[:200] + "..." if len(content) > 200 else content
        recent.append(f"Assistant: {truncated}")
        count += 1

    for msg in reversed(messages):
        if not should_process():
            return "\n".join(reversed(recent))
        
        if isinstance(msg, HumanMessage):
            process_human_message(msg)
        if isinstance(msg, AIMessage) and msg.content:
            process_ai_message(msg)

    return "\n".join(reversed(recent))


async def create_router_node():
    """
    Factory function that creates an LLM-based router node.

    Returns:
        Pure function that takes state and returns routing decision
    """
    logger.info("Setting up Router LLM: Claude Haiku 4.5")
    langfuse_handler = get_langfuse_handler()
    callbacks = [langfuse_handler] if langfuse_handler else []

    router_llm = ChatAnthropic(
        model="claude-haiku-4-5-20251001",
        temperature=0,
        max_tokens=200,
        callbacks=callbacks,
    )
    llm_with_output = router_llm.with_structured_output(RouterDecision)
    logger.info("✓ Router LLM configured")

    async def router_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Route to appropriate worker based on user intent"""
        logger.info("🔀 ROUTER NODE: Deciding which agent to use...")

        current_message = _get_last_user_message(state["messages"])
        recent_context = _get_recent_context(state["messages"])

        logger.info(f"   Current message: {current_message[:100]}...")

        user_prompt = f"""Recent conversation:
{recent_context}

Current user message to route: "{current_message}"

Which worker should handle this request?"""

        messages = [
            SystemMessage(content=ROUTER_SYSTEM_PROMPT),
            HumanMessage(content=user_prompt),
        ]

        try:
            decision = await llm_with_output.ainvoke(messages)
            logger.info(f"✓ Router decision: {decision.route}")
            logger.info(f"   Reasoning: {decision.reasoning}")
            return {"route_to_agent": decision.route}
        except Exception as e:
            logger.error(f"⚠️ Router failed: {e}, defaulting to worker")
            return {"route_to_agent": "worker"}

    return router_node


def route_from_router_edge(state: Dict[str, Any]) -> str:
    """Read which agent the router decided to use"""
    route = state.get("route_to_agent", "worker")
    logger.info(f"🔀 Router decided: {route}")

    valid_routes = {"worker", "cover_letter_writer", "job_search"}
    if route not in valid_routes:
        logger.warning(f"Invalid route '{route}', defaulting to 'worker'")
        return "worker"

    return route

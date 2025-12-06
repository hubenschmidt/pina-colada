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
    route: Literal["worker", "job_search", "cover_letter_writer", "crm_worker"] = Field(
        description="The worker to route to based on user intent"
    )
    reasoning: str = Field(
        description="Brief explanation of why this route was chosen"
    )


ROUTER_SYSTEM_PROMPT = """Route user request to correct worker.

WORKERS:
- job_search: find/search jobs, job listings
- cover_letter_writer: write/create cover letter
- crm_worker: accounts, orgs, contacts, leads, deals, tasks, CRM data research
- worker: everything else (general questions, conversation, resume questions, ambiguous)

RULES:
- Route based on CURRENT message intent
- When in doubt → worker"""


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

    valid_routes = {"worker", "cover_letter_writer", "job_search", "crm_worker"}
    if route not in valid_routes:
        logger.warning(f"Invalid route '{route}', defaulting to 'worker'")
        return "worker"

    return route

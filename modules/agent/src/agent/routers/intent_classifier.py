"""
Intent classifier - lightweight Haiku classifier for fast-path routing.
"""

import logging
from typing import Dict, Any, Literal

from langchain_core.messages import HumanMessage, SystemMessage
from langchain_anthropic import ChatAnthropic
from pydantic import BaseModel, Field

from agent.util.langfuse_helper import get_langfuse_handler

logger = logging.getLogger(__name__)


class FastPathIntent(BaseModel):
    """Structured output for intent classification."""
    intent_type: Literal["lookup", "count", "list", "other"] = Field(
        description="Type of intent: lookup (single entity), count (how many), list (show all), other (complex)"
    )
    entity_type: Literal["individual", "account", "organization", "contact", "document"] | None = Field(
        default=None, description="Type of entity"
    )
    query: str | None = Field(default=None, description="Search query if lookup, else null")


CLASSIFIER_PROMPT = """Classify user intent for CRM fast-path routing.

INTENT TYPES:
- lookup: Single entity search by name (e.g., "look up John Smith", "find Acme Corp")
- count: How many entities exist (e.g., "how many contacts", "count individuals")
- list: Show all entities of a type (e.g., "list all accounts", "show organizations")
- other: Complex queries, comparisons, analysis, document operations

Output intent_type, entity_type (individual/account/organization/contact), query (name if lookup, else null)."""


def _get_last_user_message(messages) -> str:
    """Extract the last user message."""
    for msg in reversed(messages):
        if isinstance(msg, HumanMessage):
            return msg.content if isinstance(msg.content, str) else ""
    return ""


async def create_intent_classifier_node():
    """Create a lightweight intent classifier node."""
    logger.info("Setting up Intent Classifier: Claude Haiku 4.5")
    langfuse_handler = get_langfuse_handler()
    callbacks = [langfuse_handler] if langfuse_handler else []

    classifier_llm = ChatAnthropic(
        model="claude-haiku-4-5-20251001",
        temperature=0,
        max_tokens=100,
        callbacks=callbacks,
    )
    llm_with_output = classifier_llm.with_structured_output(FastPathIntent)
    logger.info("✓ Intent Classifier configured")

    async def classifier_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Classify user intent for fast-path routing."""
        logger.info("🎯 INTENT CLASSIFIER: Checking for fast-path...")

        user_message = _get_last_user_message(state["messages"])
        logger.info(f"   Message: {user_message[:80]}...")

        messages = [
            SystemMessage(content=CLASSIFIER_PROMPT),
            HumanMessage(content=user_message),
        ]

        input_chars = sum(len(m.content) for m in messages)
        estimated_input = input_chars // 4

        try:
            intent = await llm_with_output.ainvoke(messages)
            output_chars = len(intent.query or "") + 40
            estimated_output = output_chars // 4
            token_usage = {
                "input": estimated_input,
                "output": estimated_output,
                "total": estimated_input + estimated_output,
            }

            logger.info(f"✓ Intent: type={intent.intent_type}, entity={intent.entity_type}, query={intent.query}")
            logger.info(f"   Classifier tokens (est): {token_usage['total']}")

            return {
                "fast_path_intent": intent.intent_type,
                "lookup_entity_type": intent.entity_type,
                "lookup_query": intent.query,
                "token_usage": token_usage,
                "model_name": "claude-haiku-4-5-20251001",
                "current_node": "intent_classifier",
            }
        except Exception as e:
            logger.error(f"⚠️ Classifier failed: {e}, falling back to full flow")
            return {
                "fast_path_intent": "other",
                "lookup_entity_type": None,
                "lookup_query": None,
                "token_usage": {"input": 0, "output": 0, "total": 0},
            }

    return classifier_node


def route_from_classifier(state: Dict[str, Any]) -> str:
    """Route based on classifier result."""
    intent = state.get("fast_path_intent")
    entity_type = state.get("lookup_entity_type")

    # Document operations always go to full flow (need document tools)
    if entity_type == "document":
        logger.info("→ FULL FLOW: Document operation, using router")
        return "router"

    if intent == "lookup" and entity_type and state.get("lookup_query"):
        logger.info("→ FAST PATH: Simple lookup")
        return "fast_lookup"
    if intent == "count" and entity_type:
        logger.info("→ FAST PATH: Count query")
        return "fast_count"
    if intent == "list" and entity_type:
        logger.info("→ FAST PATH: List query")
        return "fast_list"

    logger.info("→ FULL FLOW: Complex query, using router")
    return "router"

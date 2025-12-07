"""
Response formatter - lightweight Haiku formatter for fast-path responses.
"""

import logging
from typing import Dict, Any

from langchain_core.messages import HumanMessage, SystemMessage, AIMessage
from langchain_anthropic import ChatAnthropic

from agent.util.langfuse_helper import get_langfuse_handler

logger = logging.getLogger(__name__)


FORMATTER_PROMPT = """Format this CRM lookup result as a helpful, conversational response.
Be concise but include all relevant details (name, email, title, etc).
Plain text only, no markdown."""


async def create_format_response_node():
    """Create a lightweight response formatter node."""
    logger.info("Setting up Response Formatter: Claude Haiku 4.5")
    langfuse_handler = get_langfuse_handler()
    callbacks = [langfuse_handler] if langfuse_handler else []

    formatter_llm = ChatAnthropic(
        model="claude-haiku-4-5-20251001",
        temperature=0.3,
        max_tokens=500,
        callbacks=callbacks,
    )
    logger.info("✓ Response Formatter configured")

    async def format_response_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Format lookup result into natural language response."""
        lookup_result = state.get("lookup_result", "")
        user_message = ""
        for msg in reversed(state.get("messages", [])):
            if isinstance(msg, HumanMessage):
                user_message = msg.content if isinstance(msg.content, str) else ""
                break

        logger.info(f"📝 FORMAT RESPONSE: Formatting {len(lookup_result)} chars")

        messages = [
            SystemMessage(content=FORMATTER_PROMPT),
            HumanMessage(content=f"User asked: {user_message}\n\nLookup result:\n{lookup_result}"),
        ]

        input_chars = sum(len(m.content) for m in messages)
        estimated_input = input_chars // 4

        try:
            response = await formatter_llm.ainvoke(messages)
            output_chars = len(response.content)
            estimated_output = output_chars // 4
            token_usage = {
                "input": estimated_input,
                "output": estimated_output,
                "total": estimated_input + estimated_output,
            }

            logger.info(f"✓ Formatted response: {len(response.content)} chars")
            logger.info(f"   Formatter tokens (est): {token_usage['total']}")

            return {
                "messages": [AIMessage(content=response.content)],
                "token_usage": token_usage,
                "model_name": "claude-haiku-4-5-20251001",
                "current_node": "format_response",
            }
        except Exception as e:
            logger.error(f"Format response failed: {e}")
            # Fallback: return raw result
            return {
                "messages": [AIMessage(content=lookup_result)],
                "token_usage": {"input": 0, "output": 0, "total": 0},
            }

    return format_response_node

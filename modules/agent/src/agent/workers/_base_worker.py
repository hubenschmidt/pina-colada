"""
Base worker - shared factory function for all worker types
"""

import logging
from typing import Dict, Any, Callable, Optional

from langchain_core.messages import SystemMessage
from langchain_openai import ChatOpenAI

from agent.util.langfuse_helper import get_langfuse_handler

logger = logging.getLogger(__name__)


async def create_base_worker_node(
    worker_name: str,
    build_prompt: Callable[[Dict[str, Any]], str],
    tools: Optional[list] = None,
    trim_messages_fn: Optional[Callable] = None,
    max_tokens: int = 512,
    temperature: float = 0.7,
):
    """
    Factory function that creates a worker node with dependency injection.

    Args:
        worker_name: Name for logging purposes
        build_prompt: Function(state) -> system_prompt
        tools: List of tools to bind to LLM (None for no tools)
        trim_messages_fn: Function to trim message history
        max_tokens: Max completion tokens for LLM
        temperature: LLM temperature

    Returns:
        Pure function that takes state and returns updated state
    """
    logger.info(f"Setting up {worker_name} LLM: OpenAI GPT-5 (temperature={temperature})")
    langfuse_handler = get_langfuse_handler()

    callbacks = [langfuse_handler] if langfuse_handler else []

    worker_llm = ChatOpenAI(
        model="gpt-5-chat-latest",
        temperature=temperature,
        max_completion_tokens=max_tokens,
        max_retries=3,
        callbacks=callbacks,
    )

    llm = worker_llm
    if tools:
        llm = worker_llm.bind_tools(tools)
        logger.info(f"✓ {worker_name} LLM configured with {len(tools)} tools")

    if not tools:
        logger.info(f"✓ {worker_name} LLM configured (no tools)")

    def worker_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Pure function: state in -> state updates out"""
        logger.info(f"🤖 {worker_name.upper()} NODE: Processing...")

        # Build system prompt
        system_prompt = build_prompt(state)

        # Handle evaluator feedback
        if state.get("feedback_on_work"):
            logger.info(f"⚠️  {worker_name} received feedback from evaluator, retrying...")
            system_prompt += f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\nPlease improve your response based on this feedback."

        # Trim message history
        trimmed_messages = state["messages"]
        if trim_messages_fn:
            trimmed_messages = trim_messages_fn(state["messages"], max_tokens=6000)

        messages = [SystemMessage(content=system_prompt)] + trimmed_messages

        logger.info(f"   Message count: {len(messages)} (trimmed from {len(state['messages'])})")
        logger.info(f"   System prompt length: ~{len(system_prompt)} chars")

        # Get response
        response = llm.invoke(messages)

        logger.info(f"✓ {worker_name} response generated")
        logger.info(f"   Response length: {len(response.content)} chars")
        if tools:
            logger.info(f"   Has tool calls: {bool(response.tool_calls)}")

        return {"messages": [response]}

    return worker_node


def route_from_worker_with_tools(state: Dict[str, Any]) -> str:
    """Standard routing for workers with tools - route to tools or evaluator"""
    last_message = state["messages"][-1]

    if hasattr(last_message, "tool_calls") and last_message.tool_calls:
        logger.info("→ Routing to TOOLS")
        return "tools"

    logger.info("→ Routing to EVALUATOR")
    return "evaluator"

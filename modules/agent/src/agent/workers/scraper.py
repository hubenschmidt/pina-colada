"""
Scraper node - specialized worker for web scraping and automation tasks.
"""

import logging
from typing import Dict, Any, Callable
from datetime import datetime
from langchain_core.messages import SystemMessage, HumanMessage
from langchain_openai import ChatOpenAI

logger = logging.getLogger(__name__)

from agent.util.langfuse_helper import get_langfuse_handler


def _build_scraper_prompt(state: Dict[str, Any]) -> str:
    """Pure function to build scraper system prompt."""
    return f"""You are a browser automation agent. Your job is to USE TOOLS, not describe them.

TASK: {state['success_criteria']}

YOU MUST CALL THESE TOOLS TO COMPLETE THE TASK:
- browser_navigate(url)
- browser_snapshot()
- browser_type(element, ref, text, submit)
- browser_click(element, ref)
- browser_wait_for(time)
- browser_take_screenshot(filename)

DO NOT respond with text explanations. DO NOT say what you "will" do.
IMMEDIATELY CALL browser_navigate to start. Then call browser_snapshot.
Use refs from snapshot output for browser_click and browser_type.

Date: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
"""


async def create_scraper_node(
    tools: list,
    trim_messages_fn: Callable,
):
    """
    Factory function that creates a scraper node with closed-over state.

    Returns a pure function that takes state and returns updated state.
    """
    logger.info("Setting up Scraper LLM: GPT-5 (temperature=0)")
    langfuse_handler = get_langfuse_handler()

    callbacks = [langfuse_handler] if langfuse_handler else []

    scraper_llm = ChatOpenAI(
        model="gpt-5-chat-latest",
        temperature=0,  # Deterministic for reliable tool calling
        max_tokens=2048,
        callbacks=callbacks,
    )

    llm_with_tools = scraper_llm.bind_tools(tools)
    logger.info(f"✓ Scraper LLM configured with {len(tools)} tools")

    async def scraper_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Pure function: state in -> state updates out"""
        logger.info("🌐 SCRAPER NODE: Processing web scraping request...")

        system_prompt = _build_scraper_prompt(state)

        if state.get("feedback_on_work"):
            logger.info("⚠️  Scraper received feedback from evaluator, retrying...")
            system_prompt += f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\nPlease improve your scraping approach based on this feedback."

        trimmed_messages = trim_messages_fn(state["messages"], max_tokens=1000)
        messages = [SystemMessage(content=system_prompt)] + trimmed_messages

        logger.info(
            f"   Message count: {len(messages)} (trimmed from {len(state['messages'])})"
        )
        logger.info(f"   System prompt length: ~{len(system_prompt)} chars")

        response = llm_with_tools.invoke(messages)

        logger.info("✓ Scraper response generated")
        logger.info(f"   Response length: {len(response.content)} chars")
        logger.info(f"   Has tool calls: {bool(response.tool_calls)}")

        return {"messages": [response]}

    return scraper_node


def route_from_scraper(state: Dict[str, Any]) -> str:
    """Pure routing function - examines state and returns next node"""
    last_message = state["messages"][-1]

    if hasattr(last_message, "tool_calls") and last_message.tool_calls:
        logger.info("→ Routing to TOOLS")
        return "tools"

    logger.info("→ Routing to END (bypassing evaluator to reduce iterations)")
    return "END"

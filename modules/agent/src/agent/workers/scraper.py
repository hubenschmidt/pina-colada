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


def _build_scraper_prompt(
    state: Dict[str, Any], resume_context_concise: str
) -> str:
    """Pure function to build scraper system prompt."""
    return f"""ROLE: You are {state['resume_name']} on his website, specialized in web scraping and browser automation.

DATA ACCESS:
{resume_context_concise}

TASK: {state['success_criteria']}

CAPABILITIES:
You have access to 4 powerful scraping tools:

1. analyze_page_structure - Analyzes a page to determine optimal scraping strategy
   Use this FIRST to understand if you need static scraping or browser automation

2. scrape_static_page - Fast scraping for static HTML pages
   Use for simple pages without JavaScript (requests + BeautifulSoup)

3. automate_browser - Full browser automation with Playwright
   Use for JavaScript-heavy sites, forms, multi-step flows
   Supports: click, fill, wait, extract, screenshot, navigate

4. fill_form - Intelligent form filling with auto-detection
   Use for login forms, signup forms, data entry
   Automatically matches fields by name/id/label

STRATEGY:
1. ALWAYS start by analyzing the page structure
2. Choose static scraping when possible (faster, more reliable)
3. Use browser automation for:
   - JavaScript-rendered content
   - Multi-step forms (like 401k rollover)
   - Sites requiring interaction (clicks, waits)
4. For 401k rollover automation:
   - Login → Navigate accounts → Fill rollover form → Confirm

STYLE:
- Plain text only (no markdown/formatting)
- Concise, focused responses
- Explain which tool you're using and why
- Report extracted data clearly
- Handle errors gracefully

Date: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
"""


def _get_last_user_message(messages) -> str:
    """Pure function to extract last user message."""
    for msg in reversed(messages):
        if isinstance(msg, HumanMessage):
            return msg.content
    return ""


async def create_scraper_node(
    tools: list,
    resume_context_concise: str,
    trim_messages_fn: Callable,
):
    """
    Factory function that creates a scraper node with closed-over state.

    Returns a pure function that takes state and returns updated state.
    """
    logger.info("Setting up Scraper LLM: OpenAI GPT-5 (temperature=0.7)")
    langfuse_handler = get_langfuse_handler()

    callbacks = [langfuse_handler] if langfuse_handler else []

    scraper_llm = ChatOpenAI(
        model="gpt-5-chat-latest",
        temperature=0.7,
        max_completion_tokens=2048,  # Higher limit for scraping tasks
        max_retries=3,
        callbacks=callbacks,
    )

    llm_with_tools = scraper_llm.bind_tools(tools)
    logger.info(f"✓ Scraper LLM configured with {len(tools)} tools")

    def scraper_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Pure function: state in -> state updates out"""
        logger.info("🌐 SCRAPER NODE: Processing web scraping request...")

        system_prompt = _build_scraper_prompt(state, resume_context_concise)

        if state.get("feedback_on_work"):
            logger.info("⚠️  Scraper received feedback from evaluator, retrying...")
            system_prompt += f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\nPlease improve your scraping approach based on this feedback."

        trimmed_messages = trim_messages_fn(state["messages"], max_tokens=6000)
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

    logger.info("→ Routing to EVALUATOR")
    return "evaluator"

"""
Scraper node - specialized worker for web scraping and automation tasks.

TODO: Implement browser automation with Playwright when needed.
"""

import logging
from typing import Dict, Any, Callable

from langchain_core.messages import AIMessage

logger = logging.getLogger(__name__)


async def create_scraper_node(
    tools: list,
    trim_messages_fn: Callable,
):
    """
    Factory function that creates a scraper node.

    TODO: Implement actual scraping logic with Playwright.
    """
    logger.info("Setting up Scraper node (stub)")

    async def scraper_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Stub scraper node - returns TODO message"""
        logger.info("🌐 SCRAPER NODE: Stub implementation")

        response = AIMessage(
            content="Scraper functionality is not yet implemented. "
            "This feature will support browser automation for web scraping tasks."
        )

        return {"messages": [response]}

    return scraper_node


def route_from_scraper(state: Dict[str, Any]) -> str:
    """Route to evaluator (no tool calls in stub)"""
    logger.info("→ Routing to EVALUATOR")
    return "evaluator"

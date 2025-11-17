import logging
from typing import Dict, Any
from langchain_core.messages import HumanMessage, AIMessage

logger = logging.getLogger(__name__)


def _find_last_user_message(messages) -> str:
    """Extract the last user message from message history"""
    for msg in reversed(messages):
        if isinstance(msg, HumanMessage):
            return msg.content
    return ""


def _collect_recent_messages(messages, message_type, limit: int) -> list[str]:
    """Collect recent messages of a specific type up to limit"""
    collected = []
    for msg in reversed(messages):
        if isinstance(msg, message_type):
            # Handle both str and list content (multimodal)
            content = msg.content if isinstance(msg.content, str) else ""
            collected.append(content)
        if len(collected) >= limit:
            return collected
    return collected


def route_to_agent(state: Dict[str, Any]) -> Dict[str, Any]:
    """Initial router that decides which agent should handle the request"""
    logger.info("🔀 ROUTER NODE: Deciding which agent to use...")

    last_user_message = _find_last_user_message(state["messages"])
    logger.info(f"   Last user message: {last_user_message[:100]}...")

    # Check if ANY recent user message mentions cover letter (not just the last one)
    # This handles cases where user asks for cover letter, then provides job details
    recent_user_messages = _collect_recent_messages(state["messages"], HumanMessage, 3)
    recent_assistant_messages = _collect_recent_messages(state["messages"], AIMessage, 2)

    # Combine recent context
    recent_context = " ".join(recent_user_messages + recent_assistant_messages).lower()

    cover_letter_phrases = [
        "cover letter",
        "write a letter",
        "write me a letter",
        "draft a letter",
        "draft me a letter",
        "create a cover letter",
        "application letter",
    ]

    job_search_phrases = [
        "job search",
        "find jobs",
        "job leads",
        "job opportunities",
        "search for jobs",
        "jobs in",
        "job postings",
        "job openings",
        "hiring",
        "positions available",
    ]

    scraper_phrases = [
        "scrape",
        "scraping",
        "automate",
        "browser automation",
        "extract data",
        "fill form",
        "401k",
        "rollover",
        "web automation",
        "playwright",
    ]

    is_cover_letter_request = any(
        phrase in recent_context for phrase in cover_letter_phrases
    )

    is_job_search_request = any(
        phrase in recent_context for phrase in job_search_phrases
    )

    is_scraper_request = any(
        phrase in recent_context for phrase in scraper_phrases
    )

    if is_cover_letter_request:
        logger.info("✅ Detected cover letter request in conversation context!")
        logger.info("→ Routing to COVER_LETTER_WRITER")
        return {"route_to_agent": "cover_letter_writer"}

    if is_job_search_request:
        logger.info("✅ Detected job search request in conversation context!")
        logger.info("→ Routing to JOB_HUNTER")
        return {"route_to_agent": "job_hunter"}

    if is_scraper_request:
        logger.info("✅ Detected scraper request in conversation context!")
        logger.info("→ Routing to SCRAPER")
        return {"route_to_agent": "scraper"}

    logger.info("→ Routing to WORKER")
    return {"route_to_agent": "worker"}


def route_from_router_edge(state: Dict[str, Any]) -> str:
    """Read which agent the router decided to use"""
    route = state.get("route_to_agent", "worker")
    logger.info(f"🔀 Router decided: {route}")

    valid_routes = {"worker", "cover_letter_writer", "job_hunter", "scraper"}
    if route not in valid_routes:
        logger.warning(f"Invalid route '{route}', defaulting to 'worker'")
        return "worker"

    return route

import logging
from typing import Dict, Any
from langchain_core.messages import HumanMessage, AIMessage

logger = logging.getLogger(__name__)


def route_to_agent(state: Dict[str, Any]) -> Dict[str, Any]:
    """Initial router that decides which agent should handle the request"""
    logger.info("🔀 ROUTER NODE: Deciding which agent to use...")

    # Get the last user message (the most recent one)
    last_user_message = ""
    for msg in reversed(state["messages"]):
        if isinstance(msg, HumanMessage):
            last_user_message = msg.content
            break

    logger.info(f"   Last user message: {last_user_message[:100]}...")

    # Check if ANY recent user message mentions cover letter (not just the last one)
    # This handles cases where user asks for cover letter, then provides job details
    recent_user_messages = []
    for msg in reversed(state["messages"]):
        if isinstance(msg, HumanMessage):
            recent_user_messages.append(msg.content)
        if len(recent_user_messages) >= 3:  # Check last 3 user messages
            break

    # Also check if assistant already responded about cover letters
    recent_assistant_messages = []
    for msg in reversed(state["messages"]):
        if isinstance(msg, AIMessage):
            recent_assistant_messages.append(msg.content)
        if len(recent_assistant_messages) >= 2:
            break

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

    is_cover_letter_request = any(
        phrase in recent_context for phrase in cover_letter_phrases
    )

    if is_cover_letter_request:
        logger.info("✅ Detected cover letter request in conversation context!")
        logger.info("→ Routing to COVER_LETTER_WRITER")
        return {"route_to_agent": "cover_letter_writer"}
    else:
        logger.info("→ Routing to WORKER")
        return {"route_to_agent": "worker"}


def route_from_router_edge(state: Dict[str, Any]) -> str:
    """Read which agent the router decided to use"""
    route = state.get("route_to_agent", "worker")
    logger.info(f"🔀 Router decided: {route}")
    return route

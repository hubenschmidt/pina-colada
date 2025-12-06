"""
Fast lookup/count/list nodes - direct tool execution without LLM overhead.
"""

import logging
from typing import Dict, Any

from agent.tools.crm_tools import (
    lookup_individual,
    lookup_account,
    lookup_organization,
    count_entities,
    list_entities,
)

logger = logging.getLogger(__name__)


LOOKUP_FUNCTIONS = {
    "individual": lookup_individual,
    "account": lookup_account,
    "organization": lookup_organization,
}


async def fast_lookup_node(state: Dict[str, Any]) -> Dict[str, Any]:
    """Execute CRM lookup directly without LLM involvement."""
    entity_type = state.get("lookup_entity_type")
    query = state.get("lookup_query")

    logger.info(f"⚡ FAST LOOKUP: {entity_type} '{query}'")

    lookup_fn = LOOKUP_FUNCTIONS.get(entity_type)
    if not lookup_fn:
        logger.error(f"Unknown entity type: {entity_type}")
        return {"lookup_result": f"Unknown entity type: {entity_type}"}

    try:
        result = await lookup_fn(query)
        logger.info(f"✓ Fast lookup complete: {len(result)} chars")
        return {"lookup_result": result, "token_usage": {"input": 0, "output": 0, "total": 0}}
    except Exception as e:
        logger.error(f"Fast lookup failed: {e}")
        return {"lookup_result": f"Lookup failed: {e}", "token_usage": {"input": 0, "output": 0, "total": 0}}


async def fast_count_node(state: Dict[str, Any]) -> Dict[str, Any]:
    """Execute count query directly without LLM involvement."""
    entity_type = state.get("lookup_entity_type")

    logger.info(f"⚡ FAST COUNT: {entity_type}")

    try:
        result = await count_entities(entity_type)
        logger.info(f"✓ Fast count complete: {result}")
        return {"lookup_result": result, "token_usage": {"input": 0, "output": 0, "total": 0}}
    except Exception as e:
        logger.error(f"Fast count failed: {e}")
        return {"lookup_result": f"Count failed: {e}", "token_usage": {"input": 0, "output": 0, "total": 0}}


async def fast_list_node(state: Dict[str, Any]) -> Dict[str, Any]:
    """Execute list query directly without LLM involvement."""
    entity_type = state.get("lookup_entity_type")

    logger.info(f"⚡ FAST LIST: {entity_type}")

    try:
        result = await list_entities(entity_type)
        logger.info(f"✓ Fast list complete: {len(result)} chars")
        return {"lookup_result": result, "token_usage": {"input": 0, "output": 0, "total": 0}}
    except Exception as e:
        logger.error(f"Fast list failed: {e}")
        return {"lookup_result": f"List failed: {e}", "token_usage": {"input": 0, "output": 0, "total": 0}}

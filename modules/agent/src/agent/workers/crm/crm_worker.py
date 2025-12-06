"""
CRM Worker node - AI-assisted CRM with RAG and semi-automatic updates.

Capabilities:
- Query CRM data using schema registry (RAG)
- Web search for new leads, contacts, company info
- Propose updates for user approval (no direct writes)
"""

from typing import Dict, Any, Callable

from agent.prompts.worker_prompts import build_crm_worker_prompt
from agent.workers._base_worker import create_base_worker_node, route_from_worker_with_tools


async def create_crm_worker_node(
    tools: list,
    schema_context: str,
    trim_messages_fn: Callable,
):
    """Create a CRM-specialized worker node.

    Args:
        tools: List of tools (CRM lookups + web search)
        schema_context: Pre-loaded schema context from reasoning service
        trim_messages_fn: Function to trim message history
    """

    def build_prompt(state: Dict[str, Any]) -> str:
        """Build CRM worker prompt with schema context."""
        return build_crm_worker_prompt(
            schema_context=schema_context,
            success_criteria=state["success_criteria"],
        )

    return await create_base_worker_node(
        worker_name="CRM Worker",
        build_prompt=build_prompt,
        tools=tools,
        trim_messages_fn=trim_messages_fn,
        max_tokens=2048,  # Higher for proposal output
        temperature=0.3,  # Lower for more deterministic CRM queries
        # Keep GPT-5 for reliable tool calling
    )


def route_from_crm_worker(state: Dict[str, Any]) -> str:
    """Route from CRM worker to tools or evaluator."""
    return route_from_worker_with_tools(state)

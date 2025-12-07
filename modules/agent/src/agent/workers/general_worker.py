"""
Worker node - general-purpose assistant
"""

from typing import Dict, Any, Callable

from agent.prompts.worker_prompts import build_worker_prompt
from agent.workers._base_worker import create_base_worker_node, route_from_worker_with_tools


async def create_worker_node(
    tools: list,
    trim_messages_fn: Callable,
):
    """Create a general-purpose worker node"""

    def build_prompt(state: Dict[str, Any]) -> str:
        return build_worker_prompt(state["success_criteria"])

    return await create_base_worker_node(
        worker_name="general_worker",
        build_prompt=build_prompt,
        tools=tools,
        trim_messages_fn=trim_messages_fn,
        max_tokens=2048,  # Higher for document content
    )


def route_from_worker(state: Dict[str, Any]) -> str:
    """Route from worker to tools or evaluator"""
    return route_from_worker_with_tools(state)

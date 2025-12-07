"""
Job Search node - specialized for finding job opportunities
"""

from typing import Dict, Any, Callable

from agent.prompts.worker_prompts import build_job_search_prompt
from agent.workers._base_worker import create_base_worker_node, route_from_worker_with_tools


async def create_job_search_node(
    tools: list,
    trim_messages_fn: Callable,
):
    """Create a job search specialized node"""

    def build_prompt(state: Dict[str, Any]) -> str:
        """Build job search prompt"""
        return build_job_search_prompt(state["success_criteria"])

    return await create_base_worker_node(
        worker_name="job_search",
        build_prompt=build_prompt,
        tools=tools,
        trim_messages_fn=trim_messages_fn,
        max_tokens=1024,
        # Using default gpt-5.1 - gpt-5-mini fails at tool calling
    )


def route_from_job_search(state: Dict[str, Any]) -> str:
    """Route from job search to tools or evaluator"""
    return route_from_worker_with_tools(state)

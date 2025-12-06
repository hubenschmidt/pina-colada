"""
Cover Letter Writer node - specialized for writing cover letters
"""

from typing import Dict, Any, Callable

from agent.prompts.worker_prompts import build_cover_letter_writer_prompt
from agent.workers._base_worker import create_base_worker_node, route_from_worker_with_tools


async def create_cover_letter_writer_node(
    tools: list,
    trim_messages_fn: Callable,
):
    """Create a cover letter writer node with document tool access"""

    def build_prompt(state: Dict[str, Any]) -> str:
        """Build cover letter writer prompt"""
        return build_cover_letter_writer_prompt(state["resume_name"])

    return await create_base_worker_node(
        worker_name="Cover Letter Writer",
        build_prompt=build_prompt,
        tools=tools,  # Document tools for fetching resume/cover letters
        trim_messages_fn=trim_messages_fn,
        max_tokens=1500,
    )


def route_from_cover_letter_writer(state: Dict[str, Any]) -> str:
    """Route from cover letter writer to tools or evaluator"""
    return route_from_worker_with_tools(state)

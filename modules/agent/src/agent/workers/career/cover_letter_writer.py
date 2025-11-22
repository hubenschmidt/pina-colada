"""
Cover Letter Writer node - specialized for writing cover letters
"""

from typing import Dict, Any, Callable

from agent.prompts.worker_prompts import build_cover_letter_writer_prompt
from agent.workers._base_worker import create_base_worker_node


async def create_cover_letter_writer_node(trim_messages_fn: Callable):
    """Create a cover letter writer node"""

    def build_prompt(state: Dict[str, Any]) -> str:
        """Build cover letter writer prompt"""
        return build_cover_letter_writer_prompt(
            state["resume_name"],
            state["resume_context"],
        )

    return await create_base_worker_node(
        worker_name="Cover Letter Writer",
        build_prompt=build_prompt,
        tools=None,  # No tools - always routes to evaluator
        trim_messages_fn=trim_messages_fn,
        max_tokens=1500,
    )

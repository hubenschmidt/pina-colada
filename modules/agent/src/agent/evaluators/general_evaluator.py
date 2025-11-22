"""
General evaluator - default evaluator for general-purpose tasks
"""

from agent.prompts.evaluator_prompts import build_general_evaluator_prompt
from agent.evaluators._base_evaluator import create_base_evaluator_node


async def create_general_evaluator_node():
    """Create a general-purpose evaluator node"""
    return await create_base_evaluator_node(
        evaluator_name="General",
        build_system_prompt=build_general_evaluator_prompt,
    )

"""
Career evaluator - specialized for job search and career-related tasks
"""

from agent.prompts.evaluator_prompts import build_career_evaluator_prompt
from agent.evaluators._base_evaluator import create_base_evaluator_node


async def create_career_evaluator_node():
    """Create a career-specialized evaluator node"""
    return await create_base_evaluator_node(
        evaluator_name="Career",
        build_system_prompt=build_career_evaluator_prompt,
    )

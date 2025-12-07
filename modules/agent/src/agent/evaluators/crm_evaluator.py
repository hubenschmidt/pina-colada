"""
CRM evaluator - specialized evaluator for CRM tasks
"""

from agent.prompts.evaluator_prompts import build_crm_evaluator_prompt
from agent.evaluators._base_evaluator import create_base_evaluator_node


async def create_crm_evaluator_node():
    """Create a CRM-specialized evaluator node"""
    return await create_base_evaluator_node(
        evaluator_name="crm",
        build_system_prompt=build_crm_evaluator_prompt,
    )

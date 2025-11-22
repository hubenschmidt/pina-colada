"""
General evaluator - default evaluator for general-purpose tasks
"""

from agent.workers.evaluators.base_evaluator import create_base_evaluator_node


def _build_general_system_prompt(resume_context: str) -> str:
    """Build system prompt for general evaluation"""
    base_prompt = (
        "You evaluate whether the Assistant's response meets the success criteria.\n\n"
    )

    if resume_context:
        base_prompt += (
            "RESUME_CONTEXT\n"
            f"{resume_context}\n\n"
        )

    base_prompt += (
        "EVALUATION CRITERIA:\n"
        "- Accuracy: Is the response factually correct?\n"
        "- Completeness: Does it fully address the user's request?\n"
        "- Clarity: Is the response clear and well-structured?\n"
        "- Relevance: Does it stay on topic and avoid unnecessary information?\n\n"
        "Respond with a strict JSON object matching the schema (handled by the tool):\n"
        "- feedback: Brief assessment\n"
        "- success_criteria_met: true/false\n"
        "- user_input_needed: true if user must respond\n\n"
        "Be slightly lenient—approve unless clearly inadequate."
    )

    return base_prompt


async def create_general_evaluator_node():
    """Create a general-purpose evaluator node"""
    return await create_base_evaluator_node(
        evaluator_name="General",
        build_system_prompt=_build_general_system_prompt,
    )

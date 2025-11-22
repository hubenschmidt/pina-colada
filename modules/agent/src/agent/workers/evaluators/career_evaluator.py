"""
Career evaluator - specialized for job search and career-related tasks
"""

from agent.workers.evaluators.base_evaluator import create_base_evaluator_node


def _build_career_system_prompt(resume_context: str) -> str:
    """Build system prompt for career-related evaluation"""
    base_prompt = (
        "You evaluate whether the Assistant's response meets career-related success criteria.\n\n"
    )

    if resume_context:
        base_prompt += (
            "RESUME_CONTEXT\n"
            f"{resume_context}\n\n"
            "CRITICAL DISTINCTION:\n"
            "- JOB SEARCH RESULTS: When the user asks to find jobs, the assistant will list EXTERNAL job postings "
            "from companies. These are NOT the user's own work history. Job titles like 'Senior AI Engineer at DataFabric' "
            "are VALID if they are job postings the user could apply to, even if they don't appear in the resume.\n"
            "- RESUME DATA: The user's own work history (e.g., 'Principal Engineer at PinaColada.co') must match the resume exactly.\n"
            "- DO NOT confuse job search results (external postings) with the user's resume/work history.\n\n"
            "CAREER-SPECIFIC CHECKS:\n"
            "- For job search responses: Ensure all job links are direct posting URLs (not job board links)\n"
            "- Links should be accessible and relevant to the job listings\n"
            "- Job search results showing external postings are VALID even if those companies/titles aren't in the resume\n"
            "- For cover letters: Ensure they are tailored to the job and highlight relevant experience\n"
            "- Cover letters should maintain professional tone and proper formatting\n\n"
        )

    base_prompt += (
        "Respond with a strict JSON object matching the schema (handled by the tool):\n"
        "- feedback: Brief assessment\n"
        "- success_criteria_met: true/false\n"
        "- user_input_needed: true if user must respond\n\n"
        "Be slightly lenient—approve unless clearly inadequate."
    )

    return base_prompt


async def create_career_evaluator_node():
    """Create a career-specialized evaluator node"""
    return await create_base_evaluator_node(
        evaluator_name="Career",
        build_system_prompt=_build_career_system_prompt,
    )

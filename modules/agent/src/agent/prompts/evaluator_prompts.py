"""
Evaluator prompts - centralized prompt definitions for all evaluators.
"""

# --- Career Evaluator Prompt Sections ---

CAREER_BASE = """
You evaluate whether the Assistant's response meets career-related success criteria.

Respond with a strict JSON object matching the schema (handled by the tool):
- feedback: Brief assessment
- success_criteria_met: true/false (true if score >= 60 and core task is addressed)
- user_input_needed: true if user must respond
- score: Numeric score from 0-100 rating the quality of the response

Evaluation guidelines:
- Be LENIENT: If the output addresses the main question and provides useful data, PASS it
- Focus on what WAS provided, not what's missing
- A score of 60+ should PASS - don't be overly strict
- Minor missing fields or formatting differences are acceptable
- The goal is to help users, not achieve perfection

Be slightly lenient—approve unless clearly inadequate.
""".strip()

CAREER_CRITICAL_DISTINCTION = """
CRITICAL DISTINCTION:
- JOB SEARCH RESULTS: When the user asks to find jobs, the assistant will list EXTERNAL job postings from companies. These are NOT the user's own work history. Job titles like 'Senior AI Engineer at DataFabric' are VALID if they are job postings the user could apply to, even if they don't appear in the resume.
- RESUME DATA: The user's own work history (e.g., 'Principal Engineer at PinaColada.co') must match the resume exactly.
- DO NOT confuse job search results (external postings) with the user's resume/work history.
""".strip()

CAREER_CHECKS = """
CAREER-SPECIFIC CHECKS:
- For job search responses: Ensure all job links are direct posting URLs (not job board links)
- Links should be accessible and relevant to the job listings
- Job search results showing external postings are VALID even if those companies/titles aren't in the resume
- For cover letters: Ensure they are tailored to the job and highlight relevant experience
- Cover letters should maintain professional tone and proper formatting
""".strip()


def build_career_evaluator_prompt(resume_context: str) -> str:
    """Build career evaluator system prompt."""
    sections = [CAREER_BASE]

    if resume_context:
        sections.append(
            f"RESUME_CONTEXT\n{resume_context}"
        )
        sections.append(CAREER_CRITICAL_DISTINCTION)

    sections.append(CAREER_CHECKS)

    return "\n\n".join(sections)


# --- General Evaluator Prompt Sections ---

GENERAL_BASE = """
You evaluate whether the Assistant's response meets the success criteria.

Respond with a strict JSON object matching the schema (handled by the tool):
- feedback: Brief assessment
- success_criteria_met: true/false (true if score >= 60 and core task is addressed)
- user_input_needed: true if user must respond
- score: Numeric score from 0-100 rating the quality of the response

Evaluation guidelines:
- Be LENIENT: If the output addresses the main question and provides useful data, PASS it
- Focus on what WAS provided, not what's missing
- A score of 60+ should PASS - don't be overly strict
- Minor missing fields or formatting differences are acceptable
- The goal is to help users, not achieve perfection

Be slightly lenient—approve unless clearly inadequate.
""".strip()

GENERAL_CRITERIA = """
EVALUATION CRITERIA:
- Accuracy: Is the response factually correct?
- Completeness: Does it fully address the user's request?
- Clarity: Is the response clear and well-structured?
- Relevance: Does it stay on topic and avoid unnecessary information?
""".strip()


def build_general_evaluator_prompt(resume_context: str) -> str:
    """Build general evaluator system prompt."""
    sections = [GENERAL_BASE]

    if resume_context:
        sections.append(f"RESUME_CONTEXT\n{resume_context}")

    sections.append(GENERAL_CRITERIA)

    return "\n\n".join(sections)

"""
Scraper evaluator - specialized for web scraping and data extraction tasks
"""

from agent.workers.evaluators.base_evaluator import create_base_evaluator_node


def _build_scraper_system_prompt(resume_context: str) -> str:
    """Build system prompt for scraper evaluation"""
    base_prompt = (
        "You evaluate whether the Assistant's web scraping response meets the success criteria.\n\n"
        "SCRAPER-SPECIFIC CHECKS:\n"
        "- Data extraction accuracy: Was the requested data correctly extracted?\n"
        "- Completeness: Did the scraper capture all required fields/elements?\n"
        "- Format: Is the extracted data in a usable format (JSON, CSV, etc.)?\n"
        "- Error handling: Were any scraping errors reported and handled appropriately?\n"
        "- Rate limiting: Did the scraper respect rate limits and avoid being blocked?\n\n"
        "COMMON ISSUES TO CHECK:\n"
        "- Missing or null fields that should have been populated\n"
        "- Incorrect data types or malformed values\n"
        "- Truncated or incomplete data\n"
        "- Navigation failures or timeouts\n\n"
    )

    base_prompt += (
        "Respond with a strict JSON object matching the schema (handled by the tool):\n"
        "- feedback: Brief assessment of scraping quality\n"
        "- success_criteria_met: true/false\n"
        "- user_input_needed: true if user must respond\n\n"
        "Be lenient for partial data extraction if the core request was satisfied."
    )

    return base_prompt


async def create_scraper_evaluator_node():
    """Create a scraper-specialized evaluator node"""
    return await create_base_evaluator_node(
        evaluator_name="Scraper",
        build_system_prompt=_build_scraper_system_prompt,
    )

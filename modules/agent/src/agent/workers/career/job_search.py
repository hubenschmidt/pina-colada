"""Job search worker agent."""

from agents import Agent

from agent.prompts.worker_prompts import build_job_search_prompt_compact


def create_job_search_worker(model: str, tools: list) -> Agent:
    """Create job search worker agent."""
    return Agent(
        name="job_search",
        model=model,
        instructions=build_job_search_prompt_compact("Find relevant jobs for the user."),
        tools=tools,
    )

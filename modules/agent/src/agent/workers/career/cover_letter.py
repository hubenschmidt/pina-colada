"""Cover letter writer agent."""

from agents import Agent

from agent.prompts.worker_prompts import build_cover_letter_writer_prompt


def create_cover_letter_worker(model: str, tools: list) -> Agent:
    """Create cover letter writer agent."""
    return Agent(
        name="cover_letter_writer",
        model=model,
        instructions=build_cover_letter_writer_prompt(),
        tools=tools,
    )

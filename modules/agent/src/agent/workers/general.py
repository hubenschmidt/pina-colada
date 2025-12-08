"""General worker agent for miscellaneous requests."""

from agents import Agent

from agent.prompts.worker_prompts import build_worker_prompt


def create_general_worker(model: str, tools: list) -> Agent:
    """Create general worker agent."""
    return Agent(
        name="worker",
        model=model,
        instructions=build_worker_prompt("Help the user with their request."),
        tools=tools,
    )

"""CRM worker agent."""

from agents import Agent

from agent.prompts.worker_prompts import build_crm_worker_prompt_compact


def create_crm_worker(model: str, tools: list, schema_context: str = "") -> Agent:
    """Create CRM worker agent."""
    return Agent(
        name="crm_worker",
        model=model,
        instructions=build_crm_worker_prompt_compact(schema_context, "Help with CRM data."),
        tools=tools,
    )

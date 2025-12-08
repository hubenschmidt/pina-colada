"""Triage router agent - routes requests to specialist workers."""

from agents import Agent


TRIAGE_INSTRUCTIONS = """Route requests to the appropriate specialist.

Specialists:
- job_search: job hunting, applications, job listings, emailing job results
- cover_letter_writer: writing cover letters
- crm_worker: CRM data (contacts, accounts, leads, deals)
- worker: general questions, conversation, resume analysis, everything else

Hand off to the appropriate specialist based on the user's request."""


def create_triage_router(model, workers: list) -> Agent:
    """Create triage router that hands off to specialist workers."""
    return Agent(
        name="triage",
        model=model,
        instructions=TRIAGE_INSTRUCTIONS,
        handoffs=workers,
    )

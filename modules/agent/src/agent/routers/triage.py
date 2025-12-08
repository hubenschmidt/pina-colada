"""Triage router agent - routes requests to specialist workers."""

from agents import Agent


TRIAGE_INSTRUCTIONS = """Route requests to the appropriate specialist.

Specialists:
- job_search: job hunting, applications, job listings, emailing job results. Also handles CRM lookups when needed for job matching (has access to CRM tools).
- cover_letter_writer: writing cover letters
- crm_worker: CRM data only (contacts, accounts, leads, deals) - use when NO job search is involved
- worker: general questions, conversation, resume analysis, everything else

IMPORTANT: If request mentions BOTH job search AND CRM/resume lookup, route to job_search (not crm_worker).

Hand off to the appropriate specialist based on the user's request."""


def create_triage_router(model, workers: list) -> Agent:
    """Create triage router that hands off to specialist workers."""
    return Agent(
        name="triage",
        model=model,
        instructions=TRIAGE_INSTRUCTIONS,
        handoffs=workers,
    )

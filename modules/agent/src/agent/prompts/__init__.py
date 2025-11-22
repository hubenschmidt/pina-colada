"""
Centralized prompt definitions for all agents and evaluators.
"""

from agent.prompts.evaluator_prompts import (
    build_career_evaluator_prompt,
    build_general_evaluator_prompt,
)

from agent.prompts.worker_prompts import (
    build_worker_prompt,
    build_job_search_prompt,
    build_cover_letter_writer_prompt,
)

__all__ = [
    # Evaluator prompts
    "build_career_evaluator_prompt",
    "build_general_evaluator_prompt",
    # Worker prompts
    "build_worker_prompt",
    "build_job_search_prompt",
    "build_cover_letter_writer_prompt",
]

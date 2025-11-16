"""Services module for business logic."""

from services.job_service import (
    get_all_jobs,
    get_applied_jobs_only,
    get_applied_identifiers,
    fetch_applied_jobs,
    get_jobs_details,
    is_job_applied,
    filter_jobs,
    add_job,
    add_applied_job,
    update_job_by_company,
)

__all__ = [
    "get_all_jobs",
    "get_applied_jobs_only",
    "get_applied_identifiers",
    "fetch_applied_jobs",
    "get_jobs_details",
    "is_job_applied",
    "filter_jobs",
    "add_job",
    "add_applied_job",
    "update_job_by_company",
]

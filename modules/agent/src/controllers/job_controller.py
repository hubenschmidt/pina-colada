"""Controller layer for job routing to services"""

import logging
from typing import List, Optional, Dict, Any
from datetime import datetime
from lib.serialization import model_to_dict
from lib.decorators import handle_http_exceptions
from services.job_service import (
    get_jobs_paginated,
    create_job as create_job_service,
    get_job as get_job_service,
    update_job as update_job_service,
    delete_job as delete_job_service,
    get_statuses as get_statuses_service,
    get_jobs_with_status as get_jobs_with_status_service,
    get_recent_resume_date as get_recent_resume_date_service,
)

logger = logging.getLogger(__name__)


def _to_paged_response(count: int, page: int, limit: int, items: List) -> dict:
    """Convert to paged response format."""
    return {
        "items": items,
        "currentPage": page,
        "totalPages": max(1, (count + limit - 1) // limit),
        "total": count,
        "pageSize": limit,
    }


def _extract_company_info(organizations: list, individuals: list) -> tuple[str, str]:
    """Extract company name and type from account data."""
    if organizations:
        return organizations[0].get("name", ""), "Organization"
    if individuals:
        ind = individuals[0]
        first_name = ind.get("first_name", "")
        last_name = ind.get("last_name", "")
        return f"{last_name}, {first_name}".strip(", "), "Individual"
    return "", "Organization"


def _format_date(value) -> str:
    """Format datetime to ISO string, truncated to date."""
    if not value:
        return ""
    date_str = value.isoformat() if isinstance(value, datetime) else str(value)
    return date_str[:10] if date_str else ""


def _format_datetime(value) -> str:
    """Format datetime to full ISO string."""
    if not value:
        return ""
    return value.isoformat() if isinstance(value, datetime) else str(value)


def _get_account_contacts(job) -> list:
    """Get contacts from job's account (org or individual)."""
    if not job.lead or not job.lead.account:
        return []

    if job.lead.account.organizations:
        return job.lead.account.organizations[0].contacts or []

    if job.lead.account.individuals:
        return job.lead.account.individuals[0].contacts or []

    return []


def _build_contact_dict(contact) -> dict:
    """Build contact dictionary from ORM contact."""
    first_name = contact.first_name or ""
    last_name = contact.last_name or ""

    # Fallback to first linked individual if contact has no name
    if not first_name and contact.individuals:
        first_name = contact.individuals[0].first_name or ""
    if not last_name and contact.individuals:
        last_name = contact.individuals[0].last_name or ""

    return {
        "id": contact.id,
        "first_name": first_name,
        "last_name": last_name,
        "email": contact.email or "",
        "phone": contact.phone or "",
        "title": contact.title,
        "is_primary": contact.is_primary,
    }


def _get_salary_info(job, job_dict: dict) -> tuple[Optional[str], Optional[int]]:
    """Get salary range and ID from job."""
    if job.salary_range_ref:
        return job.salary_range_ref.label, job.salary_range_ref.id
    return job_dict.get("salary_range"), None


def _get_industries(job) -> list:
    """Get industry names from job's account."""
    if not job.lead or not job.lead.account or not job.lead.account.industries:
        return []
    return [ind.name for ind in job.lead.account.industries]


def _job_to_response_dict(job) -> Dict[str, Any]:
    """Convert job ORM to response dictionary."""
    job_dict = model_to_dict(job, include_relationships=True)

    lead = job_dict.get("lead") or {}
    account = lead.get("account") or {}
    company, company_type = _extract_company_info(
        account.get("organizations") or [],
        account.get("individuals") or [],
    )

    contacts = [_build_contact_dict(c) for c in _get_account_contacts(job)]

    status = "Applied"
    if job.lead and job.lead.current_status:
        status = job.lead.current_status.name

    created_at = job_dict.get("created_at", "")
    salary_range, salary_range_id = _get_salary_info(job, job_dict)

    return {
        "id": str(job_dict.get("id", "")),
        "account": company,
        "account_type": company_type,
        "job_title": job_dict.get("job_title", ""),
        "date": _format_date(created_at),
        "status": status,
        "job_url": job_dict.get("job_url"),
        "salary_range": salary_range,
        "salary_range_id": salary_range_id,
        "notes": job_dict.get("notes"),
        "resume": _format_datetime(job_dict.get("resume_date")),
        "source": job_dict.get("source", "manual"),
        "created_at": _format_datetime(created_at),
        "updated_at": job_dict.get("updated_at", ""),
        "contacts": contacts,
        "industry": _get_industries(job),
    }


@handle_http_exceptions
async def get_jobs(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None, tenant_id: Optional[int] = None
) -> dict:
    """Get all jobs with pagination."""
    paginated_jobs, total_count = await get_jobs_paginated(
        page, limit, order_by, order, search, tenant_id
    )
    items = [_job_to_response_dict(job) for job in paginated_jobs]
    return _to_paged_response(total_count, page, limit, items)


@handle_http_exceptions
async def create_job(job_data: Dict[str, Any]) -> Dict[str, Any]:
    """Create a new job."""
    created = await create_job_service(job_data)
    return _job_to_response_dict(created)


@handle_http_exceptions
async def get_job(job_id: str) -> Dict[str, Any]:
    """Get a job by ID."""
    job = await get_job_service(job_id)
    return _job_to_response_dict(job)


@handle_http_exceptions
async def update_job(job_id: str, job_data: Dict[str, Any]) -> Dict[str, Any]:
    """Update a job."""
    updated = await update_job_service(job_id, job_data)
    return _job_to_response_dict(updated)


@handle_http_exceptions
async def delete_job(job_id: str) -> dict:
    """Delete a job."""
    await delete_job_service(job_id)
    return {"success": True}


@handle_http_exceptions
async def get_statuses() -> List[Dict[str, Any]]:
    """Get all job statuses."""
    statuses = await get_statuses_service()
    return [model_to_dict(s, include_relationships=False) for s in statuses]


@handle_http_exceptions
async def get_leads(statuses: Optional[str] = None) -> List[Dict[str, Any]]:
    """Get all job leads, optionally filtered by status names."""
    status_names = None
    if statuses:
        status_names = [s.strip() for s in statuses.split(",")]

    jobs = await get_jobs_with_status_service(status_names)

    items = []
    for job in jobs:
        job_dict = _job_to_response_dict(job)
        if job.lead and job.lead.current_status:
            job_dict["lead_status"] = model_to_dict(
                job.lead.current_status, include_relationships=False
            )
        items.append(job_dict)

    return items


@handle_http_exceptions
async def mark_lead_as_applied(job_id: str) -> Dict[str, Any]:
    """Mark a lead as applied."""
    update_data: Dict[str, Any] = {"status": "applied"}
    updated = await update_job_service(job_id, update_data)
    return _job_to_response_dict(updated)


@handle_http_exceptions
async def mark_lead_as_do_not_apply(job_id: str) -> Dict[str, Any]:
    """Mark a lead as do not apply."""
    update_data: Dict[str, Any] = {"status": "do_not_apply"}
    updated = await update_job_service(job_id, update_data)
    return _job_to_response_dict(updated)


@handle_http_exceptions
async def get_recent_resume_date() -> Dict[str, Optional[str]]:
    """Get the most recent resume date."""
    resume_date = await get_recent_resume_date_service()
    return {"resume_date": resume_date}

"""Controller layer for job routing to services"""

import logging
from typing import List, Optional, Dict, Any

from fastapi import Request

from lib.serialization import model_to_dict
from lib.decorators import handle_http_exceptions
from lib.date_utils import format_date, format_datetime, format_display_date
from repositories.job_repository import JobCreate, JobUpdate
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

# Re-export for routes
__all__ = ["JobCreate", "JobUpdate"]


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
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def _get_salary_info(job, job_dict: dict) -> tuple[Optional[str], Optional[int]]:
    """Get salary range and ID from job."""
    if job.salary_range_ref:
        return job.salary_range_ref.label, job.salary_range_ref.id
    if job_dict:
        return job_dict.get("salary_range"), None
    return job.salary_range, None


def _get_industries(job) -> list:
    """Get industry names from job's account."""
    if not job.lead or not job.lead.account or not job.lead.account.industries:
        return []
    return [ind.name for ind in job.lead.account.industries]


def _extract_company_from_orm(job) -> tuple[str, str]:
    """Extract company name and type directly from ORM."""
    if not job.lead or not job.lead.account:
        return "", "Organization"
    
    if job.lead.account.organizations:
        return job.lead.account.organizations[0].name, "Organization"
    
    if job.lead.account.individuals:
        ind = job.lead.account.individuals[0]
        first_name = ind.first_name or ""
        last_name = ind.last_name or ""
        company = f"{last_name}, {first_name}".strip(", ")
        return company, "Individual"
    
    return "", "Organization"


def _get_status_name(job) -> str:
    """Get status name from job lead."""
    if not job.lead or not job.lead.current_status:
        return "Applied"
    return job.lead.current_status.name


def _get_project_ids(job) -> list:
    """Get project IDs from job lead."""
    if not job.lead or not job.lead.projects:
        return []
    return [p.id for p in job.lead.projects]


def _job_to_list_response(job) -> Dict[str, Any]:
    """Convert job ORM to response dictionary - optimized for list/table view.

    Only returns fields needed for table columns:
    Account, Job Title, Status, Description, Resume, URL, Created, Updated
    """
    company, _ = _extract_company_from_orm(job)
    status = _get_status_name(job)
    created_at = job.lead.created_at if job.lead else None

    return {
        "id": str(job.id),
        "account": company,
        "job_title": job.job_title or "",
        "status": status,
        "description": job.description,
        "resume": format_datetime(job.resume_date),
        "formatted_resume_date": format_display_date(job.resume_date),
        "job_url": job.job_url,
        "created_at": format_datetime(created_at),
        "formatted_created_at": format_display_date(created_at),
        "updated_at": format_datetime(job.updated_at),
        "formatted_updated_at": format_display_date(job.updated_at),
    }


def _job_to_detail_response(job) -> Dict[str, Any]:
    """Convert job ORM to full response dictionary - for detail/edit views."""
    company, company_type = _extract_company_from_orm(job)
    status = _get_status_name(job)
    created_at = job.lead.created_at if job.lead else None
    salary_range, salary_range_id = _get_salary_info(job, {})
    project_ids = _get_project_ids(job)
    contacts = [_contact_to_dict(c) for c in _get_account_contacts(job)]
    industry = _get_industries(job)

    return {
        "id": str(job.id),
        "account": company,
        "account_type": company_type,
        "job_title": job.job_title or "",
        "date": format_date(created_at),
        "formatted_date": format_display_date(created_at),
        "status": status,
        "job_url": job.job_url,
        "salary_range": salary_range,
        "salary_range_id": salary_range_id,
        "description": job.description,
        "resume": format_datetime(job.resume_date),
        "formatted_resume_date": format_display_date(job.resume_date),
        "source": job.lead.source if job.lead else "manual",
        "created_at": format_datetime(created_at),
        "updated_at": format_datetime(job.updated_at),
        "contacts": contacts,
        "industry": industry,
        "project_ids": project_ids,
    }


@handle_http_exceptions
async def get_jobs(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
    project_id: Optional[int] = None,
) -> dict:
    """Get all jobs with pagination."""
    tenant_id = getattr(request.state, "tenant_id", None)
    paginated_jobs, total_count = await get_jobs_paginated(
        page, limit, order_by, order, search, tenant_id, project_id
    )
    items = [_job_to_list_response(job) for job in paginated_jobs]
    return _to_paged_response(total_count, page, limit, items)


@handle_http_exceptions
async def create_job(request: Request, data: JobCreate) -> Dict[str, Any]:
    """Create a new job."""
    job_data = data.dict()
    job_data["tenant_id"] = getattr(request.state, "tenant_id", None)
    job_data["user_id"] = getattr(request.state, "user_id", None)
    created = await create_job_service(job_data)
    return _job_to_detail_response(created)


@handle_http_exceptions
async def get_job(job_id: str) -> Dict[str, Any]:
    """Get a job by ID."""
    job = await get_job_service(job_id)
    return _job_to_detail_response(job)


@handle_http_exceptions
async def update_job(request: Request, job_id: str, data: JobUpdate) -> Dict[str, Any]:
    """Update a job."""
    job_data = data.dict(exclude_unset=True)
    job_data["user_id"] = getattr(request.state, "user_id", None)
    updated = await update_job_service(job_id, job_data)
    return _job_to_detail_response(updated)


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
        job_dict = _job_to_detail_response(job)
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
    return _job_to_detail_response(updated)


@handle_http_exceptions
async def mark_lead_as_do_not_apply(job_id: str) -> Dict[str, Any]:
    """Mark a lead as do not apply."""
    update_data: Dict[str, Any] = {"status": "do_not_apply"}
    updated = await update_job_service(job_id, update_data)
    return _job_to_detail_response(updated)


@handle_http_exceptions
async def get_recent_resume_date() -> Dict[str, Optional[str]]:
    """Get the most recent resume date."""
    resume_date = await get_recent_resume_date_service()
    return {"resume_date": resume_date}

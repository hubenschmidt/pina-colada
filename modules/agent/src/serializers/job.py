"""Serializers for job-related models."""

from typing import Dict, Any, Optional

from lib.date_utils import format_date, format_datetime, format_display_date
from serializers.common import contact_to_dict


def extract_company_from_orm(job) -> tuple[str, str]:
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


def get_account_contacts(job) -> list:
    """Get contacts from job's account (org or individual)."""
    if not job.lead or not job.lead.account:
        return []

    if job.lead.account.organizations:
        return job.lead.account.organizations[0].contacts or []

    if job.lead.account.individuals:
        return job.lead.account.individuals[0].contacts or []

    return []


def get_salary_info(job, job_dict: dict) -> tuple[Optional[str], Optional[int]]:
    """Get salary range and ID from job."""
    if job.salary_range_ref:
        return job.salary_range_ref.label, job.salary_range_ref.id
    if job_dict:
        return job_dict.get("salary_range"), None
    return job.salary_range, None


def get_industries(job) -> list:
    """Get industry names from job's account."""
    if not job.lead or not job.lead.account or not job.lead.account.industries:
        return []
    return [ind.name for ind in job.lead.account.industries]


def get_status_name(job) -> str:
    """Get status name from job lead."""
    if not job.lead or not job.lead.current_status:
        return "Applied"
    return job.lead.current_status.name


def get_project_ids(job) -> list:
    """Get project IDs from job lead."""
    if not job.lead or not job.lead.projects:
        return []
    return [p.id for p in job.lead.projects]


def job_to_list_response(job) -> Dict[str, Any]:
    """Convert job ORM to response dictionary - optimized for list/table view."""
    company, _ = extract_company_from_orm(job)
    status = get_status_name(job)
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
        "updated_at": format_datetime(job.lead.updated_at if job.lead else None),
        "formatted_updated_at": format_display_date(job.lead.updated_at if job.lead else None),
    }


def job_to_detail_response(job) -> Dict[str, Any]:
    """Convert job ORM to full response dictionary - for detail/edit views."""
    company, company_type = extract_company_from_orm(job)
    status = get_status_name(job)
    created_at = job.lead.created_at if job.lead else None
    salary_range, salary_range_id = get_salary_info(job, {})
    project_ids = get_project_ids(job)
    contacts = [contact_to_dict(c) for c in get_account_contacts(job)]
    industry = get_industries(job)

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
        "updated_at": format_datetime(job.lead.updated_at if job.lead else None),
        "contacts": contacts,
        "industry": industry,
        "project_ids": project_ids,
    }
